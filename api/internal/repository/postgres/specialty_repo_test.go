package postgres

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupSpecialtyTestDB(t *testing.T) *gorm.DB {
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

	err = db.AutoMigrate(&domain.PsiSpecialtyModel{})
	require.NoError(t, err)

	// TRUCO SENIOR: Limpiar la tabla antes de correr los tests
	// para evitar que datos "fantasma" de tests anteriores arruinen los conteos.
	db.Exec("TRUNCATE TABLE psi_specialty_models RESTART IDENTITY CASCADE")

	return db
}

func ptrBool(b bool) *bool {
	return &b
}

func TestSpecialtyRepo_ComprehensiveSuite(t *testing.T) {
	mainDB := setupSpecialtyTestDB(t)
	ctx := context.Background()

	t.Run("Create and Unique Constraint", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewSpecialtyRepository(tx)

		spec1 := &domain.PsiSpecialtyModel{Name: "Clínica", Description: "Desc", Active: true}
		err := r.Create(ctx, spec1)
		require.NoError(t, err)

		spec2 := &domain.PsiSpecialtyModel{Name: "Clínica", Description: "Otra", Active: true}
		err = r.Create(ctx, spec2)
		require.Error(t, err, "Debe fallar por restricción de unicidad en el nombre")
	})

	t.Run("GetByID respects Active flag", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewSpecialtyRepository(tx)

		activeSpec := domain.PsiSpecialtyModel{Name: "Forense", Active: true}
		inactiveSpec := domain.PsiSpecialtyModel{Name: "Deportiva"} // Nace como true por default

		tx.Create(&activeSpec)
		tx.Create(&inactiveSpec)
		// TRUCO SENIOR PARA GORM: Forzamos físicamente el valor a FALSE con UpdateColumn
		// Esto ignora cualquier regla de default que GORM intente aplicar.
		tx.Model(&inactiveSpec).UpdateColumn("active", false)

		found, err := r.GetByID(ctx, activeSpec.ID)
		require.NoError(t, err)
		require.Equal(t, "Forense", found.Name)

		_, err = r.GetByID(ctx, inactiveSpec.ID)
		require.Error(t, err)
		require.True(t, errors.Is(err, gorm.ErrRecordNotFound), "No debe encontrar inactivas")
	})

	t.Run("GetAll with Status Filters and Sorting", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewSpecialtyRepository(tx)

		tx.Create(&domain.PsiSpecialtyModel{Name: "Zeta", Active: true})
		tx.Create(&domain.PsiSpecialtyModel{Name: "Alfa", Active: true})

		beta := domain.PsiSpecialtyModel{Name: "Beta"}
		tx.Create(&beta)
		tx.Model(&beta).UpdateColumn("active", false) // Forzamos inactivo

		activas, err := r.GetAll(ctx, "active")
		require.NoError(t, err)
		require.Len(t, activas, 2)
		require.Equal(t, "Alfa", activas[0].Name, "Debe ordenar alfabéticamente")

		inactivas, err := r.GetAll(ctx, "inactive")
		require.NoError(t, err)
		require.Len(t, inactivas, 1)
		require.Equal(t, "Beta", inactivas[0].Name)

		todas, err := r.GetAll(ctx, "all")
		require.NoError(t, err)
		require.Len(t, todas, 3)
	})

	t.Run("Count with Pointer Logic", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewSpecialtyRepository(tx)

		tx.Create(&domain.PsiSpecialtyModel{Name: "E1", Active: true})
		tx.Create(&domain.PsiSpecialtyModel{Name: "E2", Active: true})

		e3 := domain.PsiSpecialtyModel{Name: "E3"}
		tx.Create(&e3)
		tx.Model(&e3).UpdateColumn("active", false) // Forzamos inactivo

		cActivas, err := r.Count(ctx, ptrBool(true))
		require.NoError(t, err)
		require.Equal(t, int64(2), cActivas)

		cInactivas, err := r.Count(ctx, ptrBool(false))
		require.NoError(t, err)
		require.Equal(t, int64(1), cInactivas)

		cTodas, err := r.Count(ctx, nil)
		require.NoError(t, err)
		require.Equal(t, int64(3), cTodas)
	})

	t.Run("Update and Manual Soft Delete", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewSpecialtyRepository(tx)

		spec := domain.PsiSpecialtyModel{Name: "Social", Description: "Original", Active: true}
		tx.Create(&spec)

		spec.Description = "Modificada"
		err := r.Update(ctx, &spec)
		require.NoError(t, err)

		err = r.Delete(ctx, spec.ID)
		require.NoError(t, err)

		_, err = r.GetByID(ctx, spec.ID)
		require.Error(t, err, "Debe estar inactiva y no encontrarse")

		adminList, err := r.GetAllAdmin(ctx)
		require.NoError(t, err)
		require.Len(t, adminList, 1)
		require.False(t, adminList[0].Active, "El admin debe verla como inactiva")
	})
}
