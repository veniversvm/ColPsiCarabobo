package postgres

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupFullTestDB asegura que todo el ecosistema de tablas esté listo
func setupFullTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
	}

	tmpDb, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	tmpDb.Exec("CREATE DATABASE colpsi_test")
	sqlTmp, _ := tmpDb.DB()
	sqlTmp.Close()

	testDSN := strings.Replace(dsn, "dbname=postgres", "dbname=colpsi_test", 1)
	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{})
	require.NoError(t, err)

	db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";")
	err = db.AutoMigrate(
		&domain.PsiUserModel{},
		&domain.PsiUserColData{},
		&domain.PsiUserPostGrade{},
		&domain.PsiUserSocialNetwork{},
		&domain.PsiSpecialtyModel{},
	)
	require.NoError(t, err)

	return db
}

func TestPsiRepo_ComprehensiveSuite(t *testing.T) {
	mainDB := setupFullTestDB(t)
	ctx := context.Background()

	// --- 1. TEST DE BÚSQUEDA AVANZADA (EL MOTOR DEL DIRECTORIO) ---
	t.Run("Advanced Search Logic and Ranking", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		// Seed de especialidad para el filtro por ID
		spec := domain.PsiSpecialtyModel{Name: "Neuropsicología", Active: true}
		tx.Create(&spec)

		// Seed de usuarios con diferentes perfiles
		users := []domain.PsiUserModel{
			{ID: uuid.New(), FirstName: "Pedro", LastName: "Alfonso", CI: 10, FPV: 10, Solvent: true, IsActive: true, Genre: "M", MunicipalityCarabobo: "Valencia", PrimarySpecialty: "Neuropsicología", Username: "p1", Email: "p1@t.com", BornDate: time.Now()},
			{ID: uuid.New(), FirstName: "Maria", LastName: "Zeta", CI: 20, FPV: 20, Solvent: true, IsActive: true, Genre: "F", MunicipalityCarabobo: "San Diego", Username: "m2", Email: "m2@t.com", BornDate: time.Now()},
			{ID: uuid.New(), FirstName: "Insolvente", LastName: "Busquedame", CI: 30, FPV: 30, Solvent: false, IsActive: true, Username: "i3", Email: "i3@t.com", BornDate: time.Now()},
			{ID: uuid.New(), FirstName: "Baneado", LastName: "Invisible", CI: 40, FPV: 40, Solvent: true, IsActive: false, Username: "b4", Email: "b4@t.com", BornDate: time.Now()},
		}
		for i := range users {
			tx.Create(&users[i])
		}

		// ESCENARIO A: Navegación por especialidad (Solo solventes y activos)
		res, total, _ := r.SearchDirectory(ctx, request_structs.PsiDirectoryFilterDTO{SpecialtyID: spec.ID, Page: 1, Limit: 10})
		require.Equal(t, int64(1), total)
		require.Equal(t, "Pedro", res[0].FirstName)

		// ESCENARIO B: Filtro por Género y Ubicación
		res, total, _ = r.SearchDirectory(ctx, request_structs.PsiDirectoryFilterDTO{Gender: "F", Location: "San Diego", Page: 1, Limit: 10})
		require.Equal(t, int64(1), total)
		require.Equal(t, "Maria", res[0].FirstName)

		// ESCENARIO C: Búsqueda por Identidad (CI) - Debe encontrar al insolvente
		res, total, _ = r.SearchDirectory(ctx, request_structs.PsiDirectoryFilterDTO{SearchTerm: "30", Page: 1, Limit: 10})
		require.Equal(t, int64(1), total)
		require.Equal(t, "Insolvente", res[0].FirstName)

		// ESCENARIO D: Búsqueda por Identidad parcial del nombre
		res, _, _ = r.SearchDirectory(ctx, request_structs.PsiDirectoryFilterDTO{SearchTerm: "Alfons", Page: 1, Limit: 10})
		require.Equal(t, "Pedro", res[0].FirstName)
	})

	// --- 2. TEST DE SOFT DELETE E INTEGRIDAD ---
	t.Run("Soft Delete and Unscoped Recovery", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		psi := domain.PsiUserModel{ID: uuid.New(), Username: "delete_me", Email: "del@t.com", CI: 999, FPV: 999, BornDate: time.Now()}
		tx.Create(&psi)

		// Ejecutar borrado
		err := r.Delete(ctx, psi.ID)
		require.NoError(t, err)

		// Verificar que no aparece en consultas normales
		var check domain.PsiUserModel
		err = tx.First(&check, "id = ?", psi.ID).Error
		require.Error(t, err, "GORM no debe encontrarlo")

		// Verificar que SIGUE en la base de datos (Unscoped)
		err = tx.Unscoped().First(&check, "id = ?", psi.ID).Error
		require.NoError(t, err, "El registro físico debe permanecer para auditoría")
	})

	// --- 3. TEST DE ACTUALIZACIÓN DE PERFIL PÚBLICO (OMIT LOGIC) ---
	t.Run("UpdatePublicProfile with Omit Logic", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		psi := domain.PsiUserModel{ID: uuid.New(), Username: "upd", Email: "upd@t.com", CI: 77, FPV: 77, BornDate: time.Now(), MiniBio: "Original"}
		col := domain.PsiUserColData{PsiUserModelID: psi.ID, UniversityUndergraduate: "Original Uni"}
		tx.Create(&psi)
		tx.Create(&col)

		// Intentar actualizar bio y un campo sensible que NO debería cambiar por este método
		psi.MiniBio = "Actualizada"
		psi.CI = 88 // Cambio "malicioso"
		col.UniversityUndergraduate = "Actualizada Uni"

		err := r.UpdatePublicProfile(ctx, &psi, &col)
		require.NoError(t, err)

		// Verificar
		var check domain.PsiUserModel
		var checkCol domain.PsiUserColData
		tx.First(&check, psi.ID)
		tx.Where("psi_user_model_id = ?", psi.ID).First(&checkCol)

		require.Equal(t, "Actualizada", check.MiniBio)
		require.Equal(t, 77, check.CI, "El campo CI no debió cambiar si el repo usa Omit")
		require.Equal(t, "Actualizada Uni", checkCol.UniversityUndergraduate)
	})

	// --- 4. TEST DE SEARCH PARA ADMINISTRADORES ---
	t.Run("SearchAdmin ignores filters", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		// FIX: Añadimos FPV único a cada registro de prueba
		tx.Create(&domain.PsiUserModel{
			ID: uuid.New(), Username: "a1", Email: "a1@t.com",
			CI: 1, FPV: 100, // <-- AQUÍ
			Solvent: false, IsActive: true, BornDate: time.Now(),
		})

		tx.Create(&domain.PsiUserModel{
			ID: uuid.New(), Username: "a2", Email: "a2@t.com",
			CI: 2, FPV: 200, // <-- AQUÍ
			Solvent: true, IsActive: false, BornDate: time.Now(),
		})

		// SearchAdmin debe ver a los 2 sin enviar query params
		res, total, err := r.SearchAdmin(ctx, request_structs.PsiDirectoryFilterDTO{Page: 1, Limit: 10})
		require.NoError(t, err)
		require.GreaterOrEqual(t, total, int64(2))
		require.Len(t, res, 2)
	})
}
