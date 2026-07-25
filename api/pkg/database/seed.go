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
		log.Println("🌱 No se encontraron administradores. Creando Super Admin por defecto...")

		var defaultPass string
		if config.Envs.Environment == "development" {
			defaultPass = "admin123"
		} else {
			defaultPass = utils.GenerateSecureRandomString(16)
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPass), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("❌ Error al hashear contraseña de seed: %v", err)
			return
		}

		adminID := uuid.New()

		admin := &domain.UserAdmin{
			ID: adminID,
			AuditModel: domain.AuditModel{

				CreateBy:   "SYSTEM",
				CreateById: &adminID, // Se auto-referencia como creador inicial
			},
			Username: "admin",
			Email:    "admin@colpsicarabobo.com",
			Password: string(hashedPassword),
			IsActive: true,
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
			log.Printf("❌ Error al crear el Super Admin: %v", err)
		} else {
			log.Println("✅ Super Admin creado exitosamente.")
			if config.Envs.Environment == "development" {
				log.Printf("ℹ️  [DEV] User: %s | Pass: %s | ID: %s", admin.Username, defaultPass, admin.ID)
			} else {
				log.Printf("ℹ️  Super Admin creado — User: %s | ID: %s", admin.Username, admin.ID)
				log.Println("⚠️  La contraseña fue generada automáticamente. Cámbiela al iniciar sesión.")
			}
		}
	}
}
