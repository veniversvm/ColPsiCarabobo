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
		"service_modality_presencial":      psi.ServiceModalityPresencial,
		"service_modality_distance":        psi.ServiceModalityDistance,
		"service_modality_telephone":       psi.ServiceModalityTelephone,
		"show_service_modality":            psi.ShowServiceModality,
	}
}