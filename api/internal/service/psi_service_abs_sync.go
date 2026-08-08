// api/internal/service/psi_service_abs_sync.go
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// absSyncPace espacia las llamadas de creación a ABS en la pasada masiva para
// no saturar los rate limiters del endpoint /login de Audiobookshelf.
const absSyncPace = 100 * time.Millisecond

// ABSSyncReport resume el resultado de una sincronización de cuentas ABS.
type ABSSyncReport struct {
	Created     int      `json:"created"`     // cuentas ABS creadas
	Deactivated int      `json:"deactivated"` // cuentas ABS desactivadas
	Skipped     int      `json:"skipped"`     // solventes ya existentes
	Errors      []string `json:"errors,omitempty"`
}

// EnsureAudiobookshelf garantiza la cuenta ABS del agremiado (la crea si hace
// falta) y persiste el id en el expediente. Usado en el alta manual de un
// psicólogo solvente. Fallo no bloqueante: solo se registra en logs.
func (s *PsiService) EnsureAudiobookshelf(ctx context.Context, psi *domain.PsiUserModel) {
	if s.absSvc == nil || psi == nil {
		return
	}
	username := fmt.Sprintf("psi_%d", psi.CI)
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
//   - Crea la cuenta psi_<ci> para todo agremiado solvente y activo que no la tenga.
//   - Desactiva en ABS toda cuenta psi_* cuyo agremiado ya no es solvente/activo
//     o fue eliminado (soft-delete), revocando su acceso a la biblioteca.
//
// La desactivación está SIEMPRE activa (decisión de producto); quien deja de ser
// solvente pierde el acceso al sincronizar.
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
	desired := make(map[string]*domain.PsiUserModel) // username -> psi
	for i := range psis {
		psi := &psis[i]
		desired[fmt.Sprintf("psi_%d", psi.CI)] = psi
	}

	// 2) Estado real en ABS. ListUsersWithToken reutiliza el token admin en toda
	//    la pasada: hacer un login por cuenta agota el rate limiter de /login de
	//    ABS y deja cuentas sin crear (error silencioso visto en producción).
	adminToken, absUsers, err := s.absSvc.ListUsersWithToken(ctx)
	if err != nil {
		return nil, err
	}
	existing := make(map[string]AbsUser)
	for _, u := range absUsers {
		existing[u.Username] = u
	}

	// 3) Crear cuentas faltantes para solventes activos y no borrados.
	for username, psi := range desired {
		if !psi.Solvent || !psi.IsActive || psi.DeletedAt.Valid {
			continue
		}
		if _, ok := existing[username]; ok {
			report.Skipped++
			continue
		}
		userID, created, err := s.absSvc.createUserIfMissing(ctx, adminToken, username)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", username, err))
			continue
		}
		if created {
			report.Created++
		} else {
			report.Skipped++
		}
		// Persistir el id ABS (no bloquea la sync).
		if err := s.SetAudiobookshelfID(ctx, psi, userID); err != nil {
			log.Error().Err(err).Str("component", "psi-abs-sync").
				Str("id", psi.ID.String()).Msg("Error persistiendo AudioBookShellId en sync")
		}
		time.Sleep(absSyncPace)
	}

	// 4) Desactivar cuentas psi_* que ya no tienen derecho (insolventes,
	//    inactivos, eliminados, o que ni existen en la DB).
	for username, u := range existing {
		if !strings.HasPrefix(username, "psi_") {
			continue // no tocar cuentas que no son de agremiados
		}
		psi, ok := desired[username]
		shouldHave := ok && psi.Solvent && psi.IsActive && !psi.DeletedAt.Valid
		if shouldHave {
			continue
		}
		if !u.IsActive {
			continue // ya está desactivada
		}
		if err := s.absSvc.DeactivateUser(ctx, u.ID); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", username, err))
			continue
		}
		report.Deactivated++
	}

	log.Info().Str("component", "psi-abs-sync").
		Int("created", report.Created).
		Int("deactivated", report.Deactivated).
		Int("skipped", report.Skipped).
		Int("errors", len(report.Errors)).
		Msg("Sincronización de cuentas Audiobookshelf completada")

	return report, nil
}
