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
	// options=-c statement_timeout=30000: red de seguridad server-side. Si una
	// query se bloquea (lock/deadlock) el server la aborta a los 30s, liberando
	// la conexión y el backend pined en PgBouncer. Evita conexiones colgadas.
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable connect_timeout=5 options='-c statement_timeout=30000'",
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

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
