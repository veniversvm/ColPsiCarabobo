// api/internal/service/emergency_contact.go

// Package service implementa la lógica de negocio central de la aplicación.
//
// Este archivo gestiona el submódulo de Persona de Contacto para Emergencias.
// El agremiado declara hasta MaxEmergencyContacts personas a las que el Colegio
// puede avisar cuando el profesional no puede ser localizado o cuando ocurre un
// accidente durante el ejercicio profesional.
//
// Consideraciones de privacidad (LOPDP):
// Estos registros son datos personales de un TERCERO. Nunca se exponen en los
// endpoints públicos (el directorio y la ficha pública usan DTOs propios que no
// incluyen la relación). Solo los consultan el propio agremiado (auto-gestión)
// y el personal autorizado del Colegio a través de la ficha administrativa.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
)

// MaxEmergencyContacts define el límite de personas de contacto por agremiado.
//
// Mitigación de Agotamiento de Recursos: evita que un script automatizado
// inunde la tabla de contactos (que es un canal de contacto para terceros) y que
// la interfaz quede inutilizable. Tres es suficiente para un núcleo familiar.
const MaxEmergencyContacts = 3

// =========================================================================
// VALIDACIÓN Y NORMALIZACIÓN
// =========================================================================

// emergencyContactInput agrupa los valores ya normalizados a persistir.
type emergencyContactInput struct {
	Name         string
	Relationship string
	Phone        string
	Email        string
}

// normalizeEmergencyContact aplica limpieza y valida la regla de negocio del
// submódulo: nombre y parentesco obligatorios, más al menos un canal de contacto
// (teléfono o correo). Centralizarlo aquí garantiza que el alta, el PATCH de la
// autogestión y el del panel admin apliquen EXACTAMENTE la misma regla.
func normalizeEmergencyContact(name, relationship, phone, email string) (emergencyContactInput, error) {
	out := emergencyContactInput{
		Name:         strings.TrimSpace(collapseSpaces(name)),
		Relationship: strings.TrimSpace(collapseSpaces(relationship)),
		Phone:        normalizeEmergencyPhone(phone),
		Email:        strings.ToLower(strings.TrimSpace(email)),
	}

	if out.Name == "" {
		return emergencyContactInput{}, domain.ErrEmergencyContactIncomplete
	}
	if out.Relationship == "" {
		return emergencyContactInput{}, domain.ErrEmergencyContactIncomplete
	}

	// Canal de contacto: al menos UNO de los dos (regla declarada por el Colegio).
	if out.Phone == "" && out.Email == "" {
		return emergencyContactInput{}, domain.ErrEmergencyContactIncomplete
	}

	// El correo, si viene, se valida con el parser nativo (RFC 5322).
	if out.Email != "" {
		clean, err := utils.ParseAndValidateEmail(out.Email)
		if err != nil {
			return emergencyContactInput{}, errors.New("formato de correo electrónico inválido")
		}
		out.Email = clean
	}

	return out, nil
}

// normalizeEmergencyPhone sanea el teléfono a un formato estable: solo dígitos y
// un '+' inicial opcional. Eliminar separadores ("0412-123.45.67", "(0412) 1234567")
// evita duplicados del mismo número y unifica fijo/celular en un solo campo.
//
// El rango 7..20 dígitos cubre tanto el formato local (10-11 dígitos) como el
// internacional E.164, y a la vez descarta basura como "abc" o "123".
func normalizeEmergencyPhone(raw string) string {
	var b strings.Builder
	hasPlus := false
	digits := 0
	for _, r := range raw {
		switch {
		case unicode.IsDigit(r):
			digits++
			b.WriteRune(r)
		case (r == '+') && b.Len() == 0 && !hasPlus:
			hasPlus = true
			b.WriteRune(r)
		}
	}
	if digits < 7 || digits > 20 {
		return ""
	}
	return b.String()
}

// collapseSpaces reduce secuencias de espacios a uno solo y recorta los extremos.
func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// =========================================================================
// GESTIÓN DE CONTACTO DE EMERGENCIA (AUTOGESTIÓN)
// =========================================================================

// AddEmergencyContact registra una persona de contacto para emergencias del propio
// agremiado. Aplica control de cuota y normalización de datos.
func (s *PsiService) AddEmergencyContact(ctx context.Context, psi *domain.PsiUserModel, req request_structs.CreateEmergencyContactRequest) error {
	// 1. CONTROL DE CUOTA
	currentCount, err := s.repo.CountEmergencyContactsByPsiID(ctx, psi.ID)
	if err != nil {
		return fmt.Errorf("error al verificar límite de contactos de emergencia: %w", err)
	}
	if currentCount >= MaxEmergencyContacts {
		return domain.ErrMaxEmergencyContacts
	}

	// 2. NORMALIZACIÓN Y REGLA DE INTEGRIDAD
	in, err := normalizeEmergencyContact(req.Name, req.Relationship, req.Phone, req.Email)
	if err != nil {
		return err
	}

	// 3. PERSISTENCIA
	contact := &domain.PsiUserEmergencyContact{
		ID: uuid.Must(uuid.NewV7()),
		AuditModel: domain.AuditModel{
			CreateBy:   psi.Username,
			CreateById: &psi.ID,
			UpdateBy:   psi.Username,
			UpdateById: &psi.ID,
		},
		PsiUserID:    psi.ID,
		Name:         in.Name,
		Relationship: in.Relationship,
		Phone:        in.Phone,
		Email:        in.Email,
	}

	if err := s.repo.CreateEmergencyContact(ctx, contact); err != nil {
		return err
	}

	// Bitácora de cambios: alta de contacto de emergencia (auto-gestión).
	evt := auditPsiSelfEvent(psi, domain.AuditActionCreate)
	evt.Changes = emergencyContactCreateChanges(contact)
	RecordAudit(ctx, evt)
	return nil
}

// UpdateEmergencyContact permite la edición parcial (PATCH) de una persona de
// contacto de emergencia.
//
// Prevención de Vulnerabilidad IDOR: valida que el UUID del contacto pertenezca
// realmente al agremiado autenticado, bloqueando la inyección del ID de otra
// ficha en la URL.
func (s *PsiService) UpdateEmergencyContact(ctx context.Context, psi *domain.PsiUserModel, contactID uuid.UUID, req request_structs.UpdateEmergencyContactRequest) error {
	contact, err := s.repo.GetEmergencyContactByID(ctx, contactID)
	if err != nil {
		return domain.ErrEmergencyContactNotFound
	}

	// SEGURIDAD (Ownership Check)
	if contact.PsiUserID != psi.ID {
		return domain.ErrEmergencyContactPermDenied
	}

	// Snapshot previo para el diff por campo de la bitácora.
	beforeSnapshot := emergencyContactSnapshot(contact)

	// Aplicación de cambios parciales (semántica PATCH real).
	if req.Name != nil {
		contact.Name = *req.Name
	}
	if req.Relationship != nil {
		contact.Relationship = *req.Relationship
	}
	if req.Phone != nil {
		contact.Phone = *req.Phone
	}
	if req.Email != nil {
		contact.Email = *req.Email
	}

	// La regla de integridad se re-evalúa sobre el resultado combinado: un PATCH
	// nunca puede dejar el contacto sin nombre, sin parentesco o sin canal.
	in, err := normalizeEmergencyContact(contact.Name, contact.Relationship, contact.Phone, contact.Email)
	if err != nil {
		return err
	}
	contact.Name, contact.Relationship = in.Name, in.Relationship
	contact.Phone, contact.Email = in.Phone, in.Email

	// Rastro forense del actor.
	contact.UpdateBy = psi.Username
	contact.UpdateById = &psi.ID

	if err := s.repo.UpdateEmergencyContact(ctx, contact); err != nil {
		return err
	}

	// Bitácora de cambios: edición de contacto de emergencia (auto-gestión).
	evt := auditPsiSelfEvent(psi, domain.AuditActionUpdate)
	evt.Changes = BuildDiff(beforeSnapshot, emergencyContactSnapshot(contact))
	RecordAudit(ctx, evt)
	return nil
}

// DeleteEmergencyContact elimina (soft delete) una persona de contacto.
//
// Control de acceso polimórfico: el mismo método lo invocan la autogestión (solo
// sus propios registros) y la moderación admin (autoridad global, ya validada en
// la capa de middleware).
func (s *PsiService) DeleteEmergencyContact(ctx context.Context, executorRole string, executorID uuid.UUID, contactID uuid.UUID) error {
	contact, err := s.repo.GetEmergencyContactByID(ctx, contactID)
	if err != nil {
		return domain.ErrEmergencyContactNotFound
	}

	// VALIDACIÓN DE JERARQUÍA Y PERMISOS (RBAC Dinámico)
	if executorRole == "psi" {
		if contact.PsiUserID != executorID {
			return domain.ErrEmergencyContactOwnDenied
		}
	} else if executorRole == "admin" {
		// La autoridad del admin sobre la ficha ya fue validada por el middleware.
	} else {
		return domain.ErrInsufficientPerms
	}

	// Etiqueta del expediente para la bitácora (best-effort).
	label := contact.PsiUserID.String()
	actorUsername := ""
	if owner, gErr := s.repo.GetByID(ctx, contact.PsiUserID); gErr == nil && owner != nil {
		label = psiAuditLabel(owner)
		if executorRole == "psi" {
			actorUsername = owner.Username
		}
	}

	if err := s.repo.DeleteEmergencyContact(ctx, contactID); err != nil {
		return err
	}

	// Bitácora de cambios: baja de contacto de emergencia.
	evt := AuditEvent{
		Entity:        domain.AuditEntityPsi,
		EntityID:      contact.PsiUserID.String(),
		EntityLabel:   label,
		Action:        domain.AuditActionDelete,
		ActorID:       executorID,
		ActorRole:     executorRole,
		ActorUsername: actorUsername,
		Changes:       emergencyContactRemovalChanges(contact),
	}
	RecordAudit(ctx, evt)
	return nil
}

// =========================================================================
// GESTIÓN DE CONTACTO DE EMERGENCIA (MODERACIÓN ADMIN)
// =========================================================================

// AddEmergencyContactByAdmin registra un contacto de emergencia en la ficha de un
// psicólogo desde el panel administrativo (p. ej. cuando el Colegio lo obtiene por
// teléfono). Aplica la misma cuota y normalización que la autogestión, pero la
// auditoría refleja al operador administrativo.
func (s *PsiService) AddEmergencyContactByAdmin(ctx context.Context, admin *domain.UserAdmin, psiID uuid.UUID, req request_structs.CreateEmergencyContactRequest) error {
	// 1. GATEKEEPING: visibilidad de la ficha (mismo criterio que GetPsiByIDAdmin).
	if !admin.Sudo && !admin.CanUpdatePsi && !admin.CanCreatePsi && !admin.CanReadPsi {
		return domain.ErrInsufficientPerms
	}

	// 2. INTEGRIDAD REFERENCIAL
	target, err := s.repo.GetByID(ctx, psiID)
	if err != nil {
		return domain.ErrPsiNotFound
	}

	// 3. CONTROL DE CUOTA
	currentCount, err := s.repo.CountEmergencyContactsByPsiID(ctx, psiID)
	if err != nil {
		return fmt.Errorf("error al verificar límite de contactos de emergencia: %w", err)
	}
	if currentCount >= MaxEmergencyContacts {
		return domain.ErrMaxEmergencyContacts
	}

	// 4. NORMALIZACIÓN Y REGLA DE INTEGRIDAD
	in, err := normalizeEmergencyContact(req.Name, req.Relationship, req.Phone, req.Email)
	if err != nil {
		return err
	}

	// 5. PERSISTENCIA
	contact := &domain.PsiUserEmergencyContact{
		ID: uuid.Must(uuid.NewV7()),
		AuditModel: domain.AuditModel{
			CreateBy:   admin.Username,
			CreateById: &admin.ID,
			UpdateBy:   admin.Username,
			UpdateById: &admin.ID,
		},
		PsiUserID:    psiID,
		Name:         in.Name,
		Relationship: in.Relationship,
		Phone:        in.Phone,
		Email:        in.Email,
	}

	if err := s.repo.CreateEmergencyContact(ctx, contact); err != nil {
		return err
	}

	// Bitácora de cambios: alta de contacto de emergencia (moderación admin).
	evt := auditAdminEvent(admin, domain.AuditEntityPsi, psiID.String(), psiAuditLabel(target), domain.AuditActionCreate)
	evt.Changes = emergencyContactCreateChanges(contact)
	RecordAudit(ctx, evt)
	return nil
}

// UpdateEmergencyContactByAdmin edita un contacto de emergencia de un psicólogo
// desde el panel administrativo.
//
// Prevención de IDOR: el contacto debe pertenecer al psicólogo indicado en la
// ruta; de lo contrario un admin podría editar (y ver en la respuesta de error)
// el contacto de otra ficha.
func (s *PsiService) UpdateEmergencyContactByAdmin(ctx context.Context, admin *domain.UserAdmin, psiID, contactID uuid.UUID, req request_structs.UpdateEmergencyContactRequest) error {
	// 1. GATEKEEPING
	if !admin.Sudo && !admin.CanUpdatePsi && !admin.CanCreatePsi && !admin.CanReadPsi {
		return domain.ErrInsufficientPerms
	}

	// 2. EXISTENCIA
	contact, err := s.repo.GetEmergencyContactByID(ctx, contactID)
	if err != nil {
		return domain.ErrEmergencyContactNotFound
	}

	// 3. OWNERSHIP CHECK: la fila debe pertenecer al psicólogo de la ruta.
	if contact.PsiUserID != psiID {
		return domain.ErrEmergencyContactOwnDenied
	}

	// Etiqueta del expediente para la bitácora (best-effort).
	label := psiID.String()
	if target, gErr := s.repo.GetByID(ctx, psiID); gErr == nil && target != nil {
		label = psiAuditLabel(target)
	}

	// Snapshot previo para el diff por campo.
	beforeSnapshot := emergencyContactSnapshot(contact)

	// 4. Cambios parciales
	if req.Name != nil {
		contact.Name = *req.Name
	}
	if req.Relationship != nil {
		contact.Relationship = *req.Relationship
	}
	if req.Phone != nil {
		contact.Phone = *req.Phone
	}
	if req.Email != nil {
		contact.Email = *req.Email
	}

	// 5. Re-validación de la regla de integridad sobre el resultado combinado.
	in, err := normalizeEmergencyContact(contact.Name, contact.Relationship, contact.Phone, contact.Email)
	if err != nil {
		return err
	}
	contact.Name, contact.Relationship = in.Name, in.Relationship
	contact.Phone, contact.Email = in.Phone, in.Email

	// Rastro forense del operador.
	contact.UpdateBy = admin.Username
	contact.UpdateById = &admin.ID

	if err := s.repo.UpdateEmergencyContact(ctx, contact); err != nil {
		return err
	}

	// Bitácora de cambios: edición de contacto de emergencia (moderación admin).
	evt := auditAdminEvent(admin, domain.AuditEntityPsi, psiID.String(), label, domain.AuditActionUpdate)
	evt.Changes = BuildDiff(beforeSnapshot, emergencyContactSnapshot(contact))
	RecordAudit(ctx, evt)
	return nil
}

// DeleteEmergencyContactByAdmin elimina un contacto de emergencia de un
// psicólogo desde el panel administrativo (verifica pertenencia al psicólogo de
// la ruta para prevenir IDOR).
func (s *PsiService) DeleteEmergencyContactByAdmin(ctx context.Context, admin *domain.UserAdmin, psiID, contactID uuid.UUID) error {
	// 1. GATEKEEPING (el borrado acepta además CanDeletePsi).
	if !admin.Sudo && !admin.CanUpdatePsi && !admin.CanCreatePsi && !admin.CanReadPsi && !admin.CanDeletePsi {
		return domain.ErrInsufficientPerms
	}

	// 2. EXISTENCIA
	contact, err := s.repo.GetEmergencyContactByID(ctx, contactID)
	if err != nil {
		return domain.ErrEmergencyContactNotFound
	}

	// 3. OWNERSHIP CHECK
	if contact.PsiUserID != psiID {
		return domain.ErrEmergencyContactOwnDenied
	}

	// Etiqueta del expediente para la bitácora (best-effort).
	label := psiID.String()
	if target, gErr := s.repo.GetByID(ctx, psiID); gErr == nil && target != nil {
		label = psiAuditLabel(target)
	}

	if err := s.repo.DeleteEmergencyContact(ctx, contactID); err != nil {
		return err
	}

	// Bitácora de cambios: baja de contacto de emergencia (moderación admin).
	evt := auditAdminEvent(admin, domain.AuditEntityPsi, psiID.String(), label, domain.AuditActionDelete)
	evt.Changes = emergencyContactRemovalChanges(contact)
	RecordAudit(ctx, evt)
	return nil
}
