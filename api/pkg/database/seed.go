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

// SeedWorkAreas siembra el catálogo inicial de áreas de ejercicio profesional
// (y el modelo teórico de Psicoanálisis) tanto en desarrollo como en
// producción. Es idempotente: inserta solo las entradas cuyo nombre no exista,
// de modo que nunca sobrescribe áreas renombradas o editadas por el staff.
func SeedWorkAreas(db *gorm.DB) {
	areas := []domain.PsiSpecialtyModel{
		{
			Name: "Clínica",
			Description: "Se centra en evaluar, diagnosticar y tratar trastornos mentales, " +
				"emocionales y de conducta. Los profesionales intervienen en patologías como " +
				"depresión, ansiedad, adicciones o traumas para mejorar el bienestar emocional " +
				"del paciente mediante diversos enfoques psicoterapéuticos. Trabajan principalmente " +
				"en hospitales, centros de salud mental, consultorios privados y unidades de " +
				"rehabilitación física o psicológica.",
		},
		{
			Name: "Industrial y Organizacional",
			Description: "Aplica los principios psicológicos al ámbito laboral para optimizar el " +
				"rendimiento y bienestar de los trabajadores. Sus funciones incluyen selección de " +
				"personal, evaluación del desempeño, prevención de riesgos psicosociales (burnout), " +
				"capacitación, liderazgo y gestión del clima organizacional. Se desempeñan en " +
				"departamentos de Recursos Humanos, consultoras de talento y empresas públicas o " +
				"privadas.",
		},
		{
			Name: "Educativa",
			Description: "Analiza cómo aprenden los seres humanos en entornos educativos para " +
				"optimizar el proceso de enseñanza. Evalúa y atiende dificultades de aprendizaje, " +
				"necesidades educativas especiales, problemas de conducta escolar y orientación " +
				"vocacional. Colaboran estrechamente con docentes, estudiantes y familias dentro de " +
				"colegios, universidades, centros de orientación e instituciones de investigación " +
				"pedagógica.",
		},
		{
			Name: "Social y Comunitaria",
			Description: "Estudia cómo los contextos sociales, grupales y culturales influyen en el " +
				"comportamiento, pensamientos y emociones de las personas. Se enfoca en desarrollar " +
				"programas de intervención para resolver problemáticas colectivas (violencia, " +
				"marginación, inclusión) y empoderar a comunidades vulnerables. Trabajan en " +
				"organizaciones no gubernamentales (ONG), entidades gubernamentales y centros " +
				"comunitarios.",
		},
		{
			Name: "Forense y Jurídica",
			Description: "Aplica los conocimientos psicológicos en el sistema legal y la " +
				"administración de justicia. Realiza peritajes psicológicos sobre la imputabilidad " +
				"de acusados, custodia de menores, evaluación de secuelas en víctimas y veracidad de " +
				"testimonios. Desempeñan su labor en juzgados, instituciones penitenciarias, " +
				"fiscalías y despachos de asesoría jurídica.",
		},
		{
			Name: "Neuropsicología",
			Description: "Examina la relación entre las estructuras del cerebro y las funciones " +
				"cognitivas, emocionales y conductuales. Se especializa en diagnosticar y rehabilitar " +
				"secuelas derivadas de daño cerebral adquirido (traumatismos, ictus) o enfermedades " +
				"neurodegenerativas (Alzheimer, Parkinson). Trabajan en unidades de neurología, " +
				"centros de neurorrehabilitación y laboratorios de investigación.",
		},
		{
			Name: "Psicoanálisis",
			Description: "Modelo teórico y método terapéutico fundado por Sigmund Freud que postula " +
				"que la conducta humana está motivada por conflictos, deseos y traumas reprimidos en " +
				"el inconsciente. A través de la asociación libre y la interpretación de los sueños, " +
				"busca hacer consciente lo inconsciente para resolver bloqueos emocionales. Se aplica " +
				"mayoritariamente en el ámbito de la psicología clínica en consulta privada.",
		},
	}

	for _, area := range areas {
		var count int64
		db.Model(&domain.PsiSpecialtyModel{}).
			Where("name = ?", area.Name).
			Count(&count)
		if count > 0 {
			continue
		}
		if err := db.Create(&area).Error; err != nil {
			log.Error().Err(err).Str("component", "seed").Str("area", area.Name).Msg("Error al sembrar área de trabajo")
		} else {
			log.Info().Str("component", "seed").Str("area", area.Name).Msg("Área de trabajo sembrada")
		}
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
