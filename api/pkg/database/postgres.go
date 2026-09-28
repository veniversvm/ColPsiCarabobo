// Package database centraliza la lógica de bajo nivel para la interacción con
// el motor de base de datos relacional.
package database

import (
	"fmt"
	"time"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectDB inicializa una nueva sesión de GORM utilizando el driver de PostgreSQL.
// Se conecta a través de PgBouncer (puerto 6432 por defecto) para optimizar el pooling.
//
// Retorna:
//   - *gorm.DB: El objeto de conexión que permite realizar operaciones ORM.
//   - error: Cualquier fallo durante el proceso de apertura de la conexión.
func ConnectDB() (*gorm.DB, error) {
	// DSN (Data Source Name): Cadena de conexión estandarizada.
	// Nota: config.Envs.DBPort debe ser 6432 para pasar por PgBouncer.
	// statement_timeout NO se manda por DSN: PgBouncer rechaza el startup param
	// `options` (08P01). El timeout se fija server-side en init-db/postgresql.conf
	// (30s), que aplica a todas las conexiones incluida la de PgBouncer.
	//
	// prefer_simple_protocol=true es un param LOCAL de pgx (no viaja como startup
	// param): desactiva el statement cache / protocolo extendido del driver.
	// PgBouncer en modo transacción reparte cada consulta entre conexiones de
	// backend distintas, así que un prepared statement cacheado por pgx puede
	// ejecutarse con un plan viejo tras un cambio de esquema o un reinicio del
	// pooler → 500 SQLSTATE 0A000 ("cached plan must not change result type") /
	// 08P01 ("prepared statement name is already in use"). Con protocolo simple
	// el driver no prepara statements y el pooler deja de chocar.
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable connect_timeout=5 prefer_simple_protocol=true",
		config.Envs.DBHost,
		config.Envs.DBUser,
		config.Envs.DBPass,
		config.Envs.DBName,
		config.Envs.DBPort,
	)

	// Inicialización de GORM con configuración optimizada para PgBouncer
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// REQUERIDO PARA PGBOUNCER (Transaction Mode):
		// PgBouncer en modo transacción no garantiza que la siguiente consulta
		// use la misma conexión de backend, por lo tanto, las consultas preparadas
		// (prepared statements) deben desactivarse a nivel de ORM.
		PrepareStmt: false,

		// Opcional: Se puede habilitar SkipDefaultTransaction si se busca
		// un rendimiento extremo y se manejan las transacciones manualmente,
		// pero para este proyecto lo dejaremos en false para mayor seguridad.
		SkipDefaultTransaction: false,

		// Logs: Info solo en desarrollo, Warn en producción para no exponer datos sensibles
		Logger: logger.Default.LogMode(func() logger.LogLevel {
			if config.Envs.Environment == "development" {
				return logger.Info
			}
			return logger.Warn
		}()),
	})

	if err != nil {
		return nil, fmt.Errorf("falló la conexión a la base de datos: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("error al obtener pool de conexiones: %w", err)
	}

	// Pool 1:1 con PgBouncer (DEFAULT_POOL_SIZE=20): sin colas de clientes en
	// el pooler. Tamaños acotados para un server de 4 vCPU / 8 GB.
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(8)
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
