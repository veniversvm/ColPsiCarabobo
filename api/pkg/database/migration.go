// Package database centraliza las operaciones de bajo nivel con el motor de base de datos.
package database

import (
	"github.com/rs/zerolog/log"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

// RunMigrations sincroniza los modelos de dominio con PostgreSQL e inyecta reglas de integridad.
// Esta función garantiza que la base de datos sea "self-healing" al arrancar.
func RunMigrations(db *gorm.DB) error {
	log.Info().Str("component", "migrate").Msg("Iniciando proceso de sincronización de esquema...")

	// 1. EXTENSIONES DE POSTGRES
	// Habilitamos 'pgcrypto' para generación nativa de UUIDs (gen_random_uuid).
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";").Error; err != nil {
		log.Error().Err(err).Str("component", "migrate").Msg("Error crítico: No se pudo habilitar la extensión pgcrypto")
		return err
	}

	// 2. AUTOMIGRATE (GORM)
	// Sincroniza los structs con las tablas físicas. El orden previene conflictos de FK.
	err := db.AutoMigrate(
		&domain.TextModel{},         // Contenido extenso
		&domain.UserAdmin{},         // Personal administrativo
		&domain.PsiUserModel{},      // Miembros (Psicólogos)
		&domain.PsiUserColData{},    // Datos gremiales (1:1 con PsiUserModel)
		&domain.PsiUserPostGrade{},  // Postgrados (1:N con PsiUserModel)
		&domain.Post{},              // Noticias y publicaciones
		&domain.PsiSpecialtyModel{}, // Especialidades (Nuevo modelo)
		&domain.KanbanProject{},     // Proyectos (tableros Kanban)
		&domain.KanbanMember{},      // Miembros de proyectos Kanban
		&domain.KanbanColumn{},      // Columnas del tablero Kanban
		&domain.KanbanCard{},        // Tarjetas Kanban
		&domain.KanbanNote{},        // Notas de tarjetas Kanban (máx 10 × 500 chars)
		&domain.AdminPermissionLog{}, // Auditoría forense de cambios de permisos
		&domain.AppSetting{},         // KV global de configuración (interruptores)
		&domain.SettingsAuditLog{},   // Auditoría de cambios de configuración global
	)

	if err != nil {
		log.Error().Err(err).Str("component", "migrate").Msg("Error crítico: Falló la ejecución de AutoMigrate")
		return err
	}

	// 3. REGLA DE NEGOCIO: SUDO ÚNICO (Restricción de Integridad)
	// Creamos un índice parcial único.
	// Lógica:
	// - Permite infinitas filas con 'sudo = false'.
	// - Permite solo UNA fila con 'sudo = true'.
	// - Ignora filas borradas (Soft Delete) para permitir nombrar un nuevo SUDO si el anterior fue eliminado.
	const sudoIndexSQL = `
		CREATE UNIQUE INDEX IF NOT EXISTS "idx_user_admins_unique_sudo" 
		ON "user_admins" ("sudo") 
		WHERE (sudo IS TRUE AND deleted_at IS NULL);
	`
	if err := db.Exec(sudoIndexSQL).Error; err != nil {
		log.Error().Err(err).Str("component", "migrate").Msg("Error al crear restricción de SUDO único")
		return err
	}

	log.Info().Str("component", "migrate").Msg("Esquema y reglas de integridad sincronizados exitosamente")
	return nil
}
