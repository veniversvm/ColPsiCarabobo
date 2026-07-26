package database

import (
	"log"

	"github.com/google/uuid"
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
		log.Println("[INFO] No se encontraron administradores. Creando Super Admin por defecto...")

		var defaultPass string
		if config.Envs.Environment == "development" {
			defaultPass = "admin123"
		} else {
			defaultPass = utils.GenerateSecureRandomString(16)
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPass), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("[ERROR] Error al hashear contraseña de seed: %v", err)
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
				Username: "admin",
				Email:    "admin@colpsicarabobo.com",
				Password: string(hashedPassword),
				IsActive: true,
			},
			Sudo:     true, // Acceso total

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
		}

		if err := db.Create(admin).Error; err != nil {
			log.Printf("[ERROR] Error al crear el Super Admin: %v", err)
		} else {
			log.Println("[OK] Super Admin creado exitosamente.")
			if config.Envs.Environment == "development" {
				log.Printf("[INFO] [DEV] User: %s | Pass: %s | ID: %s", admin.Username, defaultPass, admin.ID)
			} else {
				log.Printf("[INFO] Super Admin creado — User: %s | ID: %s", admin.Username, admin.ID)
				log.Println("[WARN] La contraseña fue generada automáticamente. Cámbiela al iniciar sesión.")
			}
		}
	}
}
