package database

import (
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
