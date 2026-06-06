// Package postgres contiene la suite de pruebas de integración para el repositorio
// de administradores (AdminRepository).
//
// Garantiza que las operaciones críticas de seguridad (como el login dual,
// la protección del último superusuario y la paginación del panel de control)
// funcionen de manera predecible sobre una instancia real de PostgreSQL.
package postgres

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupAdminTestDB gestiona el ciclo de vida de la base de datos de pruebas para administradores.
//
// Realiza tres tareas fundamentales:
//  1. Asegura la existencia física de la base de datos (colpsi_test).
//  2. Activa extensiones criptográficas nativas (pgcrypto) para manejar UUIDs.
//  3. Sincroniza el esquema del modelo UserAdmin.
func setupAdminTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
	}

	// Paso 1: Conexión administrativa temporal
	tmpDb, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	tmpDb.Exec("CREATE DATABASE colpsi_test")
	sqlTmp, _ := tmpDb.DB()
	sqlTmp.Close()

	// Paso 2: Conexión definitiva al entorno de test
	testDSN := strings.Replace(dsn, "dbname=postgres", "dbname=colpsi_test", 1)
	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{})
	require.NoError(t, err)

	// Paso 3: Infraestructura y Esquemas
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";")
	err = db.AutoMigrate(&domain.UserAdmin{})
	require.NoError(t, err)

	return db
}

// pBool es un helper técnico que convierte un booleano primitivo en un puntero.
// Esencial para probar filtros tri-estatales (true/false/nil) en GORM.
func pBool(b bool) *bool {
	return &b
}

// TestAdminRepo_ComprehensiveSuite agrupa las validaciones del módulo de staff.
// Utiliza aislamiento transaccional (tx.Begin / tx.Rollback) para que los tests
// puedan ejecutarse concurrentemente o en cualquier orden sin colisión de datos.
func TestAdminRepo_ComprehensiveSuite(t *testing.T) {
	mainDB := setupAdminTestDB(t)
	ctx := context.Background()

	// # Escenario 1: Autenticación Dual (Login)
	// Valida que el sistema permita iniciar sesión indistintamente con el username
	// o el correo electrónico, y que los errores de GORM se mapeen correctamente
	// para no exponer detalles técnicos al cliente.
	t.Run("Create and GetByIdentifier (Login Logic)", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAdminRepository(tx)

		admin := &domain.UserAdmin{
			ID:       uuid.New(),
			Username: "admin_test",
			Email:    "test@admin.com",
			Password: "hashed_password",
		}
		err := r.Create(ctx, admin)
		require.NoError(t, err)

		// A. Identificación por Username
		foundUser, err := r.GetByIdentifier(ctx, "admin_test")
		require.NoError(t, err)
		require.Equal(t, admin.ID, foundUser.ID)

		// B. Identificación por Email
		foundEmail, err := r.GetByIdentifier(ctx, "test@admin.com")
		require.NoError(t, err)
		require.Equal(t, admin.ID, foundEmail.ID)

		// C. Prevención de Fugas de Información
		// El repositorio debe interceptar gorm.ErrRecordNotFound y devolver un error genérico de dominio.
		_, err = r.GetByIdentifier(ctx, "no_existo")
		require.Error(t, err)
		require.Equal(t, "administrador no encontrado", err.Error())
	})

	// # Escenario 2: Prevención de Bloqueo del Sistema (SUDO Soft Delete)
	// Evalúa la regla de negocio crítica que cuenta cuántos Super Usuarios quedan activos.
	// Garantiza que los SUDOs eliminados lógicamente (Soft Delete) NO sean contabilizados,
	// lo cual previene que el sistema permita borrar al último administrador real.
	t.Run("CountSudos ignores Soft Deleted records", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAdminRepository(tx)

		// 1. Staff estándar (No debe ser contado)
		tx.Create(&domain.UserAdmin{ID: uuid.New(), Username: "staff1", Email: "1@t.com", Sudo: false})

		// 2. SUDO Activo (Este es el único que debe contar)
		sudoActive := domain.UserAdmin{ID: uuid.New(), Username: "sudo_real", Email: "2@t.com", Sudo: true}
		tx.Create(&sudoActive)

		// 3. SUDO Eliminado (Soft Delete)
		sudoDeleted := domain.UserAdmin{ID: uuid.New(), Username: "sudo_ghost", Email: "3@t.com", Sudo: true}
		tx.Create(&sudoDeleted)

		err := r.Delete(ctx, sudoDeleted.ID) // Aplicamos el Soft Delete
		require.NoError(t, err)

		// Validación
		count, err := r.CountSudos(ctx)
		require.NoError(t, err)
		require.Equal(t, int64(1), count, "El conteo debe ignorar a los usuarios regulares y a los SUDOs borrados lógicamente")
	})

	// # Escenario 3: Mutación y Persistencia
	// Verifica la actualización de campos básicos y el correcto funcionamiento de GetByID.
	t.Run("Update Admin fields", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAdminRepository(tx)

		admin := &domain.UserAdmin{ID: uuid.New(), Username: "upd_test", Email: "upd@t.com", IsActive: true}
		tx.Create(admin)

		// Modificación en memoria
		admin.IsActive = false
		admin.Username = "upd_changed"

		err := r.Update(ctx, admin)
		require.NoError(t, err)

		// Verificación de integridad en la base de datos
		check, err := r.GetByID(ctx, admin.ID)
		require.NoError(t, err)
		require.False(t, check.IsActive)
		require.Equal(t, "upd_changed", check.Username)
	})

	// # Escenario 4: Búsqueda, Filtros y Paginación
	// Comprueba el motor de consultas para el listado del panel de control, asegurando
	// que los filtros booleanos, el buscador de texto parcial (ILIKE) y los límites
	// de paginación trabajen en conjunto de manera armónica.
	t.Run("List with Filters and Pagination", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAdminRepository(tx)

		// Limpieza agresiva exclusiva para este test.
		// Al hacer aserciones exactas de 'total', requerimos una tabla vacía.
		tx.Exec("DELETE FROM user_admins")

		// Inserción de prueba
		a1 := domain.UserAdmin{ID: uuid.New(), Username: "alpha", Email: "al@mail.com"}
		tx.Create(&a1)

		a2 := domain.UserAdmin{ID: uuid.New(), Username: "bravo", Email: "br@mail.com"}
		tx.Create(&a2)

		a3 := domain.UserAdmin{ID: uuid.New(), Username: "charlie", Email: "ch@alpha.com"}
		tx.Create(&a3)

		// NOTA TÉCNICA: Usamos UpdateColumn para evadir los hooks de GORM.
		// Si el modelo define 'is_active' con un valor por defecto (default:true) en la base de datos,
		// insertarlo en false requiere esta técnica para garantizar la persistencia del estado inactivo.
		tx.Model(&a3).UpdateColumn("is_active", false)

		// Prueba A: Filtro Tri-Estatal (Solo Activos)
		res, total, err := r.List(ctx, pBool(true), "", 1, 10)
		require.NoError(t, err)
		require.Equal(t, int64(2), total)

		// Prueba B: Filtro Tri-Estatal (Solo Inactivos)
		res, total, err = r.List(ctx, pBool(false), "", 1, 10)
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Equal(t, "charlie", res[0].Username)

		// Prueba C: Búsqueda de coincidencia parcial (ILIKE) cruzada
		// El término "alph" debe atrapar el Username de a1 ("alpha") y el Email de a3 ("ch@alpha.com")
		res, total, err = r.List(ctx, nil, "alph", 1, 10)
		require.NoError(t, err)
		require.Equal(t, int64(2), total, "La búsqueda parcial ILIKE no resolvió correctamente los campos cruzados")

		// Prueba D: Lógica Matemática de Paginación
		// Solicitamos la página 1 con un límite visual de 2 registros.
		res, total, err = r.List(ctx, nil, "", 1, 2)
		require.NoError(t, err)
		require.Equal(t, int64(3), total, "El total debe reflejar el universo completo independientemente del límite visual")
		require.Len(t, res, 2, "La consulta solo debe devolver el límite solicitado (2 registros)")

		// Verificación de Orden (DESC)
		require.Equal(t, "charlie", res[0].Username, "El ordenamiento debe priorizar a los registros más recientes (DESC)")
		require.Equal(t, "bravo", res[1].Username)
	})
}
