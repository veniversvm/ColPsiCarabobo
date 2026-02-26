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

// setupAdminTestDB prepara el entorno asilado para las pruebas de administración.
func setupAdminTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
	}

	// 1. Crear BD de test si no existe
	tmpDb, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	tmpDb.Exec("CREATE DATABASE colpsi_test")
	sqlTmp, _ := tmpDb.DB()
	sqlTmp.Close()

	// 2. Conectar a la DB de test
	testDSN := strings.Replace(dsn, "dbname=postgres", "dbname=colpsi_test", 1)
	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{})
	require.NoError(t, err)

	// 3. Extensión UUID y Migración
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";")
	err = db.AutoMigrate(&domain.UserAdmin{})
	require.NoError(t, err)

	return db
}

// helper para punteros booleanos
func pBool(b bool) *bool {
	return &b
}

func TestAdminRepo_ComprehensiveSuite(t *testing.T) {
	mainDB := setupAdminTestDB(t)
	ctx := context.Background()

	// --- 1. TEST DE CREACIÓN Y BÚSQUEDA DUAL (LOGIN) ---
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

		// A. Buscar por Username
		foundUser, err := r.GetByIdentifier(ctx, "admin_test")
		require.NoError(t, err)
		require.Equal(t, admin.ID, foundUser.ID)

		// B. Buscar por Email
		foundEmail, err := r.GetByIdentifier(ctx, "test@admin.com")
		require.NoError(t, err)
		require.Equal(t, admin.ID, foundEmail.ID)

		// C. Buscar Inexistente (Verificar error mapeado)
		_, err = r.GetByIdentifier(ctx, "no_existo")
		require.Error(t, err)
		require.Equal(t, "administrador no encontrado", err.Error())
	})

	// --- 2. TEST DE REGLA SUDO Y SOFT DELETE ---
	t.Run("CountSudos ignores Soft Deleted records", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAdminRepository(tx)

		// Crear Admin Normal
		tx.Create(&domain.UserAdmin{ID: uuid.New(), Username: "staff1", Email: "1@t.com", Sudo: false})

		// Crear SUDO Activo
		sudoActive := domain.UserAdmin{ID: uuid.New(), Username: "sudo_real", Email: "2@t.com", Sudo: true}
		tx.Create(&sudoActive)

		// Crear SUDO pero eliminarlo (Soft Delete)
		sudoDeleted := domain.UserAdmin{ID: uuid.New(), Username: "sudo_ghost", Email: "3@t.com", Sudo: true}
		tx.Create(&sudoDeleted)
		err := r.Delete(ctx, sudoDeleted.ID) // Aplicamos el borrado
		require.NoError(t, err)

		// Validación: Solo debe contar 1 SUDO (el activo)
		count, err := r.CountSudos(ctx)
		require.NoError(t, err)
		require.Equal(t, int64(1), count, "Debe ignorar a los SUDOs borrados lógicamente")
	})

	// --- 3. TEST DE ACTUALIZACIÓN Y OBTENCIÓN POR ID ---
	t.Run("Update Admin fields", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAdminRepository(tx)

		admin := &domain.UserAdmin{ID: uuid.New(), Username: "upd_test", Email: "upd@t.com", IsActive: true}
		tx.Create(admin)

		// Modificar y guardar
		admin.IsActive = false
		admin.Username = "upd_changed"
		err := r.Update(ctx, admin)
		require.NoError(t, err)

		// Recuperar y verificar
		check, err := r.GetByID(ctx, admin.ID)
		require.NoError(t, err)
		require.False(t, check.IsActive)
		require.Equal(t, "upd_changed", check.Username)
	})

	// --- 4. TEST DE BUSCADOR Y PAGINACIÓN ADMINISTRATIVA ---
	t.Run("List with Filters and Pagination", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewAdminRepository(tx)

		// Limpiar tabla solo para este test por seguridad de conteos
		tx.Exec("DELETE FROM user_admins")

		// Seed de datos (Usamos UpdateColumn para el IsActive=false y evitar la trampa de GORM default:true)
		a1 := domain.UserAdmin{ID: uuid.New(), Username: "alpha", Email: "al@mail.com"}
		tx.Create(&a1)

		a2 := domain.UserAdmin{ID: uuid.New(), Username: "bravo", Email: "br@mail.com"}
		tx.Create(&a2)

		a3 := domain.UserAdmin{ID: uuid.New(), Username: "charlie", Email: "ch@alpha.com"}
		tx.Create(&a3)
		tx.Model(&a3).UpdateColumn("is_active", false) // Inactivo

		// A. Filtro Activos
		res, total, err := r.List(ctx, pBool(true), "", 1, 10)
		require.NoError(t, err)
		require.Equal(t, int64(2), total)

		// B. Filtro Inactivos
		res, total, err = r.List(ctx, pBool(false), "", 1, 10)
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Equal(t, "charlie", res[0].Username)

		// C. Búsqueda ILIKE (Debe encontrar "alpha" en username y "ch@alpha.com" en email)
		res, total, err = r.List(ctx, nil, "alph", 1, 10)
		require.NoError(t, err)
		require.Equal(t, int64(2), total, "Búsqueda parcial falló")

		// D. Paginación (Página 1, Límite 2)
		res, total, err = r.List(ctx, nil, "", 1, 2)
		require.NoError(t, err)
		require.Equal(t, int64(3), total, "El total debe reflejar la tabla completa")
		require.Len(t, res, 2, "Solo debe devolver 2 registros")
		// Debe estar ordenado por fecha desc (Los más nuevos primero)
		require.Equal(t, "charlie", res[0].Username)
		require.Equal(t, "bravo", res[1].Username)
	})
}
