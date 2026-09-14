// api/internal/service/psi_service_abs_sync.go
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// absSyncPace espacia las llamadas de creación a ABS en la pasada masiva para
// no saturar los rate limiters del endpoint /login de Audiobookshelf.
const absSyncPace = 100 * time.Millisecond

// ABSSyncReport resume el resultado de una sincronización de cuentas ABS.
type ABSSyncReport struct {
	Created     int      `json:"created"`     // cuentas ABS creadas
	Reactivated int      `json:"reactivated"` // cuentas ABS reactivadas (volvieron a ser solventes)
	Deactivated int      `json:"deactivated"` // cuentas ABS desactivadas
	Skipped     int      `json:"skipped"`     // solventes ya existentes (o sin trabajo)
	Errors      []string `json:"errors,omitempty"`
}

// AbsUsernameFor devuelve el nombre de usuario canónico de un agremiado en ABS:
// su correo normalizado (minúsculas, sin espacios). Si el agremiado no tiene
// correo, se usa el legado psi_<ci> para no colisionar.
func AbsUsernameFor(psi *domain.PsiUserModel) string {
	email := strings.ToLower(strings.TrimSpace(psi.Email))
	if email != "" {
		return email
	}
	return fmt.Sprintf("psi_%d", psi.CI)
}

// EnsureAudiobookshelf garantiza la cuenta ABS del agremiado (la crea si hace
// falta y la reactiva si quedó desactivada) y persiste el id en el expediente.
// Usado en el alta manual de un psicólogo solvente. Fallo no bloqueante: solo
// se registra en logs.
func (s *PsiService) EnsureAudiobookshelf(ctx context.Context, psi *domain.PsiUserModel) {
	if s.absSvc == nil || psi == nil {
		return
	}
	username := AbsUsernameFor(psi)
	access, err := s.absSvc.GetAccess(ctx, username)
	if err != nil {
		log.Warn().Err(err).Str("component", "psi-abs-sync").
			Str("id", psi.ID.String()).Str("username", username).
			Msg("Error aprovisionando cuenta ABS en alta manual")
		return
	}
	if err := s.SetAudiobookshelfID(ctx, psi, access.UserID); err != nil {
		log.Error().Err(err).Str("component", "psi-abs-sync").
			Str("id", psi.ID.String()).Msg("Error persistiendo AudioBookShellId")
	}
}

// SyncAudiobookshelfAccounts reconcilia las cuentas ABS con la base de datos:
//   - Crea la cuenta (username = correo del agremiado) para todo solvente
//     activo que no la tenga y reactiva la de quien recuperó la solvencia.
//   - Desactiva en ABS toda cuenta tipo "user" cuyo agremiado ya no es
//     solvente/activo o fue eliminado (soft-delete), revocando el acceso.
//   - Desactiva cuentas tipo "user" huérfanas (p.ej. los psi_<ci> legacy tras
//     migrar al correo). Las cuentas admin/root y de otros tipos se ignoran.
//
// La desactivación está SIEMPRE activa (decisión de producto); quien deja de
// ser solvente pierde el acceso al sincronizar.
func (s *PsiService) SyncAudiobookshelfAccounts(ctx context.Context) (*ABSSyncReport, error) {
	report := &ABSSyncReport{}
	if s.absSvc == nil {
		return report, nil
	}

	// 1) Estado deseado desde la base de datos (incluye soft-deleted).
	psis, err := s.repo.GetAllForABSSync(ctx)
	if err != nil {
		return nil, fmt.Errorf("obteniendo agremiados para sync ABS: %w", err)
	}

	// 2) Estado real en ABS. ListUsersWithToken reutiliza el token admin en
	//    toda la pasada: hacer un login por cuenta agota el rate limiter
	//    de /login de ABS y deja cuentas sin crear (error silencioso visto
	//    en producción).
	adminToken, absUsers, err := s.absSvc.ListUsersWithToken(ctx)
	if err != nil {
		return nil, err
	}
	existing := make(map[string]AbsUser, len(absUsers))
	for _, u := range absUsers {
		existing[u.Username] = u
	}

	// 3) Reconciliar la cuenta de cada agremiado (crear si falta, reactivar si
	//    volvió a ser solvente, desactivar si perdió el derecho).
	for i := range psis {
		psi := &psis[i]
		partial := s.reconcileOneAudiobookshelfAccount(ctx, adminToken, psi, existing)
		mergeReports(report, partial)
		if partial.Created > 0 || partial.Reactivated > 0 {
			time.Sleep(absSyncPace)
		}
	}

	// 4) Cuentas huérfanas: tipo "user" que ya no corresponden a ningún
	//    agremiado (por ejemplo, los username psi_<ci> del esquema legacy tras
	//    migrar a correo) se desactivan.
	desired := make(map[string]struct{}, len(psis))
	for i := range psis {
		desired[AbsUsernameFor(&psis[i])] = struct{}{}
	}
	for username, u := range existing {
		if u.Type != "user" {
			continue // admin/root y otras cuentas quedan intactas
		}
		if _, ok := desired[username]; ok {
			continue
		}
		if !u.IsActive {
			continue // ya está desactivada
		}
		if err := s.absSvc.DeactivateUserWithToken(ctx, adminToken, u.ID); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", username, err))
			continue
		}
		report.Deactivated++
	}

	// Registrar los errores individuales de la pasada (no bloquearon el resto).
	for _, e := range report.Errors {
		log.Warn().Str("component", "psi-abs-sync").Msg("Error en sincronización ABS: " + e)
	}

	log.Info().Str("component", "psi-abs-sync").
		Int("created", report.Created).
		Int("reactivated", report.Reactivated).
		Int("deactivated", report.Deactivated).
		Int("skipped", report.Skipped).
		Int("errors", len(report.Errors)).
		Msg("Sincronización de cuentas Audiobookshelf completada")

	return report, nil
}

// SyncAudiobookshelfForPsi sincroniza la cuenta ABS de un solo agremiado con la
// misma lógica de la pasada masiva. Alimenta el botón "Sincronizar biblioteca"
// de la ficha del psicólogo en el panel de administración.
func (s *PsiService) SyncAudiobookshelfForPsi(ctx context.Context, admin *domain.UserAdmin, targetID uuid.UUID) (*ABSSyncReport, error) {
	// RBAC: sincronizar la cuenta toca el estado del agremiado → edición.
	if !admin.Sudo && !admin.CanUpdatePsi {
		return nil, errors.New("no tienes permiso para editar registros de psicólogos")
	}
	psi, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return nil, domain.ErrPsiNotFound
	}
	if s.absSvc == nil {
		return nil, errors.New("biblioteca digital no configurada")
	}

	adminToken, absUsers, err := s.absSvc.ListUsersWithToken(ctx)
	if err != nil {
		return nil, err
	}
	existing := make(map[string]AbsUser, len(absUsers))
	for _, u := range absUsers {
		existing[u.Username] = u
	}

	return s.reconcileOneAudiobookshelfAccount(ctx, adminToken, psi, existing), nil
}

// reconcileOneAudiobookshelfAccount aplica la lógica de sync ABS a un solo
// agremiado usando un token admin ya obtenido y el mapa de cuentas existentes.
// Devuelve un reporte parcial (mergeReports lo combina en el masivo).
func (s *PsiService) reconcileOneAudiobookshelfAccount(ctx context.Context, adminToken string, psi *domain.PsiUserModel, existing map[string]AbsUser) *ABSSyncReport {
	report := &ABSSyncReport{}
	if s.absSvc == nil || psi == nil {
		report.Skipped++
		return report
	}

	username := AbsUsernameFor(psi)
	u, exists := existing[username]
	shouldHave := psi.Solvent && psi.IsActive && !psi.DeletedAt.Valid

	// Las cuentas que no son de tipo "user" (admin/root) nunca se gestionan:
	// la protección se basa en el tipo que devuelve ABS, no en el prefijo del
	// nombre.
	if exists && u.Type != "user" {
		report.Skipped++
		return report
	}

	// Sin cuenta ABS: se crea solo si tiene derecho. Si no tiene derecho, no
	// hay nada que desactivar.
	if !exists {
		if !shouldHave {
			report.Skipped++
			return report
		}
		userID, created, err := s.absSvc.createUserIfMissing(ctx, adminToken, username)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", username, err))
			return report
		}
		if created {
			report.Created++
		} else {
			report.Skipped++
		}
		// Persistir el id ABS (no bloquea la sincronización).
		if err := s.SetAudiobookshelfID(ctx, psi, userID); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s (persistencia de id): %v", username, err))
		}
		return report
	}

	// Cuenta existente con derecho: reactivar si estaba desactivada.
	if shouldHave {
		if !u.IsActive {
			if err := s.absSvc.ReactivateUserWithToken(ctx, adminToken, u.ID); err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", username, err))
				return report
			}
			report.Reactivated++
		} else {
			report.Skipped++ // ya existente y activa
		}
		return report
	}

	// Cuenta existente sin derecho: desactivar si seguía activa.
	if u.IsActive {
		if err := s.absSvc.DeactivateUserWithToken(ctx, adminToken, u.ID); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", username, err))
			return report
		}
		report.Deactivated++
	}
	return report
}

// mergeReports suma un reporte parcial al acumulado de una pasada masiva.
func mergeReports(dst, src *ABSSyncReport) {
	dst.Created += src.Created
	dst.Reactivated += src.Reactivated
	dst.Deactivated += src.Deactivated
	dst.Skipped += src.Skipped
	dst.Errors = append(dst.Errors, src.Errors...)
}
