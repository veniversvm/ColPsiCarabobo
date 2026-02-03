// Package database centraliza las operaciones de bajo nivel con el motor de base de datos,
// incluyendo la conexión, configuración del pool y sincronización de esquemas.
package database

import (
	"log"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/gorm"
)

// RunMigrations sincroniza los modelos definidos en el dominio con el esquema de PostgreSQL.
// Utiliza la funcionalidad AutoMigrate de GORM para realizar cambios no destructivos
// (creación de tablas, nuevas columnas e índices).
//
// Nota de Arquitectura: Aunque usamos Atlas para migraciones versionadas en producción,
// esta función garantiza que el entorno de desarrollo local sea "auto-congelable" y
// que las extensiones críticas del motor estén habilitadas.
func RunMigrations(db *gorm.DB) error {
	log.Println("⏳ Iniciando proceso de sincronización de esquema...")

	// 1. EXTENSIONES DE POSTGRES
	// Habilitamos 'pgcrypto' para permitir que Postgres genere UUIDs v4 de forma nativa
	// mediante la función gen_random_uuid(), que es la que definimos en AuditModel.
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";").Error; err != nil {
		log.Printf("❌ Error crítico: No se pudo habilitar la extensión pgcrypto: %v", err)
		return err
	}

	// 2. AUTOMIGRATE (GORM)
	// Sincroniza los structs de Go con las tablas físicas.
	// El orden de los parámetros es importante para que GORM pueda resolver las
	// llaves foráneas y relaciones complejas en una sola pasada.
	err := db.AutoMigrate(
		&domain.TextModel{},        // Tabla independiente (almacena textos largos)
		&domain.UserAdmin{},        // Tabla de administración (sin dependencias directas)
		&domain.PsiUserModel{},     // Entidad principal de psicólogos
		&domain.PsiUserColData{},   // Dependencia 1:1 de PsiUserModel
		&domain.PsiUserPostGrade{}, // Dependencia 1:N de PsiUserModel
		&domain.Post{},             // Dependencia de TextModel (post -> contenido)
	)

	if err != nil {
		log.Printf("❌ Error crítico: Falló la ejecución de AutoMigrate: %v", err)
		return err
	}

	log.Println("✅ Esquema de base de datos sincronizado exitosamente")
	return nil
}
