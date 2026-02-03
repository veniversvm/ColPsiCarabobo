// Package database centraliza la lógica de bajo nivel para la interacción con
// el motor de base de datos relacional.
package database

import (
	"fmt"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB inicializa una nueva sesión de GORM utilizando el driver de PostgreSQL.
// Utiliza la configuración global cargada en config.Envs para establecer la conexión.
//
// Retorna:
//   - *gorm.DB: El objeto de conexión que permite realizar operaciones ORM.
//   - error: Cualquier fallo durante el proceso de apertura de la conexión.
func ConnectDB() (*gorm.DB, error) {
	// DSN (Data Source Name): Cadena de conexión estandarizada para PostgreSQL.
	// sslmode=disable se utiliza para facilitar el desarrollo local con Docker.
	// Para producción, se recomienda cambiar a sslmode=require o verify-full.
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.Envs.DBHost,
		config.Envs.DBUser,
		config.Envs.DBPass,
		config.Envs.DBName,
		config.Envs.DBPort,
	)

	// gorm.Open inicializa el pool de conexiones.
	// Aunque el primer argumento es postgres.Open, GORM maneja internamente
	// la apertura y el mantenimiento de las conexiones inactivas (idle connections).
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Aquí se podrían añadir configuraciones globales de GORM,
		// como el Logger o estrategias de nombres de tablas.
	})

	if err != nil {
		return nil, fmt.Errorf("falló la conexión a la base de datos: %w", err)
	}

	return db, nil
}
