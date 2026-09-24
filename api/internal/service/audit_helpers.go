// api/internal/service/audit_helpers.go

// Helpers para construir eventos de auditoría y snapshots de diff
// (usados por la instrumentación de los servicios).
package service

import (
	"fmt"
	"strings"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// psiAuditLabel arma la etiqueta legible de un psicólogo para la bitácora.
func psiAuditLabel(psi *domain.PsiUserModel) string {
	if psi == nil {
		return ""
	}
	name := strings.TrimSpace(strings.TrimSpace(psi.FirstName) + " " + strings.TrimSpace(psi.LastName))
	return fmt.Sprintf("%s (FPV %d)", name, psi.FPV)
}

// auditPsiAuthEvent construye un evento de login/logout con el psicólogo como actor.
func auditPsiAuthEvent(psi *domain.PsiUserModel, action string) AuditEvent {
	return AuditEvent{
		Entity:        domain.AuditEntityPsi,
		EntityID:      psi.ID.String(),
		EntityLabel:   psiAuditLabel(psi),
		Action:        action,
		ActorID:       psi.ID,
		ActorRole:     "psi",
		ActorUsername: psi.Username,
	}
}

// auditAdminEvent construye un evento base con el admin como actor real.
func auditAdminEvent(admin *domain.UserAdmin, entity, entityID, entityLabel, action string) AuditEvent {
	return AuditEvent{
		Entity:        entity,
		EntityID:      entityID,
		EntityLabel:   entityLabel,
		Action:        action,
		ActorID:       admin.ID,
		ActorRole:     auditRoleOfAdmin(admin),
		ActorUsername: admin.Username,
	}
}

// permissionSetToMap serializa la matriz de permisos a un mapa plano para los
// diffs de la bitácora (BuildDiff) al actualizar el rol/perfil de un staff.
func permissionSetToMap(p PermissionSet) map[string]any {
	return map[string]any{
		"can_read_psi":             p.CanReadPsi,
		"can_create_psi":           p.CanCreatePsi,
		"can_update_psi":           p.CanUpdatePsi,
		"can_delete_psi":           p.CanDeletePsi,
		"can_create_admin":         p.CanCreateAdmin,
		"can_update_admin":         p.CanUpdateAdmin,
		"can_delete_admin":         p.CanDeleteAdmin,
		"can_publish":              p.CanPublish,
		"can_update_publish":       p.CanUpdatePublish,
		"can_delete_publish":       p.CanDeletePublish,
		"can_send_notifications":   p.CanSendNotifications,
		"can_manage_notifications": p.CanManageNotifications,
		"can_read_notifications":   p.CanReadNotifications,
		"can_create_tags":          p.CanCreateTags,
		"can_edit_tags":            p.CanEditTags,
		"can_delete_tags":          p.CanDeleteTags,
		"can_manage_projects":      p.CanManageProjects,
		"can_manage_tickets":       p.CanManageTickets,
		"can_view_logs":            p.CanViewLogs,
		"can_export_logs":          p.CanExportLogs,
	}
}

// auditPsiSelfEvent construye un evento con el psicólogo como actor y entidad
// (auto-gestión: perfil, títulos académicos y redes sociales).
func auditPsiSelfEvent(psi *domain.PsiUserModel, action string) AuditEvent {
	return AuditEvent{
		Entity:        domain.AuditEntityPsi,
		EntityID:      psi.ID.String(),
		EntityLabel:   psiAuditLabel(psi),
		Action:        action,
		ActorID:       psi.ID,
		ActorRole:     "psi",
		ActorUsername: psi.Username,
	}
}

// psiCoreSnapshot captura los campos editables del expediente de un psicólogo
// (identidad, contactos, estado gremial y ubicación) para calcular el diff por
// campo al actualizarlo desde el panel admin. Se excluyen credenciales, textos
// HTML y vectores para no ensuciar la bitácora (se registran en Metadata).
func psiCoreSnapshot(psi *domain.PsiUserModel) map[string]any {
	if psi == nil {
		return map[string]any{}
	}
	return map[string]any{
		"first_name":        psi.FirstName,
		"second_name":       psi.SecondName,
		"last_name":         psi.LastName,
		"second_last_name":  psi.SecondLastName,
		"fpv":               psi.FPV,
		"ci":                psi.CI,
		"nationality":       psi.Nationality,
		"control_number":    psi.ControlNumber,
		"genre":             psi.Genre,
		"solvent":           psi.Solvent,
		"proof_of_life":     psi.ProofOfLife,
		"is_active":         psi.IsActive,
		"username":          psi.Username,
		"email":             psi.Email,
		"contact_phone":     psi.ContactPhone,
		"contact_cell_phone": psi.ContactCellPhone,
		"contact_email":     psi.ContactEmail,
		"service_address":   psi.ServiceAddress,
		"municipality_carabobo":            psi.MunicipalityCarabobo,
		"state_outside":                    psi.StateOutside,
		"municipality_outside_carabobo":    psi.MunicipalityOutSideCarabobo,
		"country":                          psi.Country,
		"primary_work_area":                psi.PrimaryWorkArea,
		"secondary_work_area":              psi.SecondaryWorkArea,
		"primary_specialty_id":             psi.PrimarySpecialtyID,
		"secondary_specialty_id":           psi.SecondarySpecialtyID,
		"show_service_modality":            psi.ShowServiceModality,
	}
}

// psiSelfSnapshot captura los campos que el psicólogo puede auto-gestionar en
// su portal (contactos, privacidad, modalidad y mini bio) para el diff por
// campo de la bitácora. Reutiliza psiCoreSnapshot y añade los switches de
// visibilidad y canales por ubicación. Se excluyen credenciales, hashes, S3
// keys y campos de auditoría.
func psiSelfSnapshot(psi *domain.PsiUserModel) map[string]any {
	if psi == nil {
		return map[string]any{}
	}
	m := psiCoreSnapshot(psi)
	extra := map[string]any{
		"show_contact_email":                  psi.ShowContactEmail,
		"show_public_service_address":         psi.ShowPublicServiceAddress,
		"show_municipality_carabobo":          psi.ShowMunicipalityCarabobo,
		"phone_carabobo":                      psi.PhoneCarabobo,
		"show_phone_carabobo":                 psi.ShowPhoneCarabobo,
		"cel_phone_carabobo":                  psi.CelPhoneCarabobo,
		"show_cel_phone_carabobo":             psi.ShowCelPhoneCarabobo,
		"show_state_outside":                  psi.ShowStateOutside,
		"show_municipality_outside_carabobo":  psi.ShowMunicipalityOutSideCarabobo,
		"phone_outside_carabobo":              psi.PhoneOutSideCarabobo,
		"show_phone_outside_carabobo":         psi.ShowPhoneOutSideCarabobo,
		"cel_phone_outside_carabobo":          psi.CelPhoneOutSideCarabobo,
		"show_cel_phone_outside_carabobo":     psi.ShowCellPhoneOutSideCarabobo,
		"service_address_outside_carabobo":    psi.ServiceAddressOutSideCarabobo,
		"show_public_service_address_outside_carabobo": psi.ShowPublicServiceAddressOutSideCarabobo,
		"phone_outside_venezuela":             psi.PhoneOutSideVenezuela,
		"show_phone_outside_venezuela":        psi.ShowPhoneOutSideVenezuela,
		"cell_phone_outside_venezuela":        psi.CellPhoneOutSideVenezuela,
		"show_cell_phone_outside_venezuela":   psi.ShowCellPhoneOutSideVenezuela,
		"service_address_outside_venezuela":   psi.ServiceAddressOutSideVenezuela,
		"show_public_service_address_outside_venezuela": psi.ShowPublicServiceAddressOutSideVenezuela,
		"mini_bio":                            psi.MiniBio,
	}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

// colDataPrivacySnapshot captura los switches de privacidad académica que viven
// en psi_user_col_data (auto-gestión) para el diff por campo de la bitácora.
func colDataPrivacySnapshot(c *domain.PsiUserColData) map[string]any {
	if c == nil {
		return map[string]any{}
	}
	return map[string]any{
		"show_university_undergraduate": c.ShowUniversityUndergraduate,
		"show_graduate_date":            c.ShowGraduateDate,
		"show_mention_undergraduate":    c.ShowMentionUndergraduate,
		"birthday_notification":         c.BirthdayNotification,
	}
}

// postGradeSnapshot captura los campos de un título académico para el diff.
func postGradeSnapshot(pg *domain.PsiUserPostGrade) map[string]any {
	if pg == nil {
		return map[string]any{}
	}
	return map[string]any{
		"post_grade_title":           pg.Title,
		"post_grade_university":      pg.University,
		"post_grade_graduation_year": pg.GraduationYear,
		"post_grade_description":     pg.Description,
	}
}

// postGradeRemovalChanges convierte el snapshot de un título académico en diff
// de eliminación (solo `from`: el registro ya no existe tras la baja).
func postGradeRemovalChanges(pg *domain.PsiUserPostGrade) map[string]domain.AuditChange {
	removal := map[string]domain.AuditChange{}
	for k, v := range postGradeSnapshot(pg) {
		removal[k] = domain.AuditChange{From: v}
	}
	return removal
}

// socialSnapshot captura los campos de una red social para el diff.
func socialSnapshot(sn *domain.PsiUserSocialNetwork) map[string]any {
	if sn == nil {
		return map[string]any{}
	}
	return map[string]any{
		"social_name":   sn.Name,
		"social_url":    sn.URL,
		"social_active": sn.IsActive,
	}
}

// socialCreateChanges arma el diff de creación de una red social (solo `to`).
func socialCreateChanges(sn *domain.PsiUserSocialNetwork) map[string]domain.AuditChange {
	return map[string]domain.AuditChange{
		"social_name":   {To: sn.Name},
		"social_url":    {To: sn.URL},
		"social_active": {To: sn.IsActive},
	}
}

// socialRemovalChanges arma el diff de eliminación de una red social (solo `from`).
func socialRemovalChanges(sn *domain.PsiUserSocialNetwork) map[string]domain.AuditChange {
	return map[string]domain.AuditChange{
		"social_name": {From: sn.Name},
		"social_url":  {From: sn.URL},
	}
}