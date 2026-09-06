package database

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedAdmin verifica si existe algún administrador; si no, crea uno por defecto.
func SeedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&domain.UserAdmin{}).Count(&count)

	if count == 0 {
		log.Info().Str("component", "seed").Msg("No se encontraron administradores. Creando Super Admin por defecto...")

		var defaultPass string
		if config.Envs.AdminPassword != "" {
			defaultPass = config.Envs.AdminPassword
		} else if config.Envs.Environment == "development" {
			defaultPass = "admin123"
		} else {
			defaultPass = utils.GenerateSecureRandomString(16)
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPass), bcrypt.DefaultCost)
		if err != nil {
			log.Error().Err(err).Str("component", "seed").Msg("Error al hashear contraseña de seed")
			return
		}

		adminID := uuid.Must(uuid.NewV7())

		admin := &domain.UserAdmin{
			ID: adminID,
			AuditModel: domain.AuditModel{

				CreateBy:   "SYSTEM",
				CreateById: &adminID, // Se auto-referencia como creador inicial
			},
			Credentials: domain.Credentials{
				Username: config.Envs.AdminUsername,
				Email:    config.Envs.AdminEmail,
				Password: string(hashedPassword),
				IsActive: true,
			},
			Sudo: true, // Acceso total

			// Activamos todos los permisos granulares
			CanReadPsi:             true,
			CanCreatePsi:           true,
			CanUpdatePsi:           true,
			CanDeletePsi:           true,
			CanCreateAdmin:         true,
			CanUpdateAdmin:         true,
			CanDeleteAdmin:         true,
			CanPublish:             true,
			CanUpdatePublish:       true,
			CanDeletePublish:       true,
			CanSendNotifications:   true,
			CanManageNotifications: true,
			CanReadNotifications:   true,
			CanCreateTags:          true,
			CanEditTags:            true,
			CanDeleteTags:          true,
			CanManageProjects:      true,
			CanManageTickets:       true,
		}

		if err := db.Create(admin).Error; err != nil {
			log.Error().Err(err).Str("component", "seed").Msg("Error al crear el Super Admin")
		} else {
			log.Info().Str("component", "seed").Msg("Super Admin creado exitosamente.")
			if config.Envs.Environment == "development" {
				log.Info().Str("component", "seed").Str("user", admin.Username).Str("pass", defaultPass).Str("id", admin.ID.String()).Msg("Super Admin dev credentials")
			} else if config.Envs.AdminPassword != "" {
				log.Info().Str("component", "seed").Str("user", admin.Username).Str("email", admin.Email).Str("id", admin.ID.String()).Msg("Super Admin creado con credenciales de ADMIN_*")
			} else {
				log.Info().Str("component", "seed").Str("user", admin.Username).Str("id", admin.ID.String()).Msg("Super Admin creado")
				log.Warn().Str("component", "seed").Msg("La contraseña fue generada automáticamente. Cámbiela al iniciar sesión.")
			}
		}
	}
}

// SeedSudoPermissions garantiza que TODO administrador con rol SUDO tenga la
// matriz de permisos completa en `true`. Es idempotente y se ejecuta en cada
// arranque: cubre a los SUDO pre-existentes creados antes de que existieran
// flags como can_read_psi o can_manage_tickets (sus columnas nacieron con
// DEFAULT false). La autorización real ya bypassa con `sudo`, pero esta
// corrección hace que el estado sea coherente también en la UI.
func SeedSudoPermissions(db *gorm.DB) {
	res := db.Model(&domain.UserAdmin{}).
		Where("sudo = ?", true).
		Updates(map[string]interface{}{
			"can_read_psi":             true,
			"can_create_psi":           true,
			"can_update_psi":           true,
			"can_delete_psi":           true,
			"can_create_admin":         true,
			"can_update_admin":         true,
			"can_delete_admin":         true,
			"can_publish":              true,
			"can_update_publish":       true,
			"can_delete_publish":       true,
			"can_send_notifications":   true,
			"can_manage_notifications": true,
			"can_read_notifications":   true,
			"can_create_tags":          true,
			"can_edit_tags":            true,
			"can_delete_tags":          true,
			"can_manage_projects":      true,
			"can_manage_tickets":       true,
		})
	if res.Error != nil {
		log.Error().Err(res.Error).Str("component", "seed").Msg("Error al sincronizar permisos de SUDO")
		return
	}
	if res.RowsAffected > 0 {
		log.Info().Int64("count", res.RowsAffected).Str("component", "seed").Msg("Permisos del SUDO normalizados a TODOS=true")
	}
}

// SeedAppSettings siembra los defaults de la configuración global (KV) si las
// claves no existen aún. No sobrescribe valores ya configurados.
func SeedAppSettings(db *gorm.DB) {
	defaults := map[string]domain.ReceptionSetting{
		domain.SettingsKeyTicketsReception:      {Enabled: true},
		domain.SettingsKeyInscriptionsReception: {Enabled: true},
	}
	for key, setting := range defaults {
		var count int64
		db.Model(&domain.AppSetting{}).Where("key = ?", key).Count(&count)
		if count > 0 {
			continue
		}
		value, err := json.Marshal(setting)
		if err != nil {
			log.Error().Err(err).Str("component", "seed").Str("key", key).Msg("Error al serializar default de app_settings")
			continue
		}
		record := &domain.AppSetting{
			Key:       key,
			Value:     value,
			UpdatedAt: time.Now(),
		}
		if err := db.Create(record).Error; err != nil {
			log.Error().Err(err).Str("component", "seed").Str("key", key).Msg("Error al sembrar default de app_settings")
		} else {
			log.Info().Str("component", "seed").Str("key", key).Msg("app_settings sembrado con recepción habilitada")
		}
	}
}
