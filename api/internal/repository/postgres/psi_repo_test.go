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
		&domain.TextModel{},
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

	t.Run("Advanced Search Logic and Ranking", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		// 1. IMPORTANTE: Crear primero una biografía para satisfacer la FK
		dummyBio := domain.TextModel{ID: uuid.New(), Content: "Bio de prueba"}
		require.NoError(t, tx.Create(&dummyBio).Error)

		spec := domain.PsiSpecialtyModel{Name: "Neuropsicología", Active: true}
		tx.Create(&spec)

		now := time.Now()
		users := []domain.PsiUserModel{
			{
				ID: uuid.New(), FirstName: "Pedro", LastName: "Alfonso", CI: 10, FPV: 10,
				Solvent: true, IsActive: true, Genre: "M", Nationality: "V",
				ContactEmail: "p1@t.com", ContactPhone: "04141234567",
				MunicipalityCarabobo: "Valencia", ShowMunicipalityCarabobo: true, // 👈 Bandera de privacidad EN ENCENDIDO
				PrimaryWorkArea: "Neuropsicología", // 👈 Cambiado a WorkArea
				Username:        "p1", Email: "p1@t.com", BornDate: now, BioTextID: dummyBio.ID,
			},
			{
				ID: uuid.New(), FirstName: "Maria", LastName: "Zeta", CI: 20, FPV: 20,
				Solvent: true, IsActive: true, Genre: "F", Nationality: "V",
				ContactEmail: "m2@t.com", ContactPhone: "04120000000",
				MunicipalityCarabobo: "San Diego", ShowMunicipalityCarabobo: false, // 👈 Bandera de privacidad APAGADA
				Username: "m2", Email: "m2@t.com", BornDate: now, BioTextID: dummyBio.ID,
			},
			{
				ID: uuid.New(), FirstName: "Insolvente", LastName: "Busquedame", CI: 30, FPV: 30,
				Solvent: false, IsActive: true, Genre: "M", Nationality: "V",
				ContactEmail: "i3@t.com", Username: "i3", Email: "i3@t.com", BornDate: now, BioTextID: dummyBio.ID,
			},
			{
				ID: uuid.New(), FirstName: "Baneado", LastName: "Invisible", CI: 40, FPV: 40,
				Solvent: true, IsActive: false, Genre: "M", Nationality: "V",
				ContactEmail: "b4@t.com", Username: "b4", Email: "b4@t.com", BornDate: now, BioTextID: dummyBio.ID,
			},
		}
		for i := range users {
			require.NoError(t, tx.Create(&users[i]).Error)
		}

		// Prueba 1: Búsqueda por Área de Trabajo (Especialidad)
		res, total, err := r.SearchDirectory(ctx, request_structs.PsiDirectoryFilterDTO{SpecialtyID: spec.ID, Page: 1, Limit: 10})
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Equal(t, "Pedro", res[0].FirstName)

		// Prueba 2 (Nueva): Validar Escudo de Privacidad en Búsqueda por Ubicación
		locRes, locTotal, locErr := r.SearchDirectory(ctx, request_structs.PsiDirectoryFilterDTO{Location: "Valencia", Page: 1, Limit: 10})
		require.NoError(t, locErr)
		require.Equal(t, int64(1), locTotal, "Debe retornar 1 porque solo Pedro tiene ShowMunicipalityCarabobo=true")
		require.Equal(t, "Pedro", locRes[0].FirstName)
	})

	t.Run("Soft Delete and Unscoped Recovery", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		dummyBio := domain.TextModel{ID: uuid.New(), Content: "..."}
		tx.Create(&dummyBio)

		psi := domain.PsiUserModel{
			ID: uuid.New(), Username: "delete_me", Email: "del@t.com",
			CI: 999, FPV: 999, BornDate: time.Now(), BioTextID: dummyBio.ID,
			Genre: "M", Nationality: "V", ContactEmail: "del@t.com", ContactPhone: "123", // 👈 Añadido ContactPhone
			FirstName: "Del", LastName: "Me",
		}
		require.NoError(t, tx.Create(&psi).Error)

		err := r.Delete(ctx, psi.ID)
		require.NoError(t, err)

		var check domain.PsiUserModel
		err = tx.First(&check, "id = ?", psi.ID).Error
		require.Error(t, err)
	})

	t.Run("UpdatePublicProfile with Omit Logic", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		// 1. IMPORTANTE: Crear Bio ANTES que el usuario
		bio := domain.TextModel{ID: uuid.New(), Content: "<p>Bio Original</p>"}
		require.NoError(t, tx.Create(&bio).Error)

		psi := domain.PsiUserModel{
			ID: uuid.New(), Username: "upd", Email: "upd@t.com",
			CI: 77, FPV: 77, BornDate: time.Now(), MiniBio: "Original",
			Genre: "F", Nationality: "V", ContactEmail: "upd@t.com", ContactPhone: "123",
			FirstName: "Upd", LastName: "User",
			BioTextID: bio.ID, // Asignar la FK
		}
		require.NoError(t, tx.Create(&psi).Error)

		col := domain.PsiUserColData{PsiUserModelID: psi.ID, UniversityUndergraduate: "Original Uni"}
		require.NoError(t, tx.Create(&col).Error)

		// Simular cambios
		psi.MiniBio = "Actualizada"
		psi.PrimaryWorkArea = "Clínica" // 👈 Simular actualización de nueva columna
		psi.CI = 88                     // No debe cambiar
		bio.Content = "<p>Bio Actualizada</p>"

		err := r.UpdatePublicProfile(ctx, &psi, &col, &bio)
		require.NoError(t, err)

		var check domain.PsiUserModel
		tx.First(&check, psi.ID)
		require.Equal(t, 77, check.CI, "El campo CI no debió cambiar por Omit")
		require.Equal(t, "Clínica", check.PrimaryWorkArea, "El área de trabajo debió actualizarse")
	})

	t.Run("SearchAdmin ignores filters", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		dummyBio := domain.TextModel{ID: uuid.New(), Content: "..."}
		tx.Create(&dummyBio)

		now := time.Now()
		tx.Create(&domain.PsiUserModel{
			ID: uuid.New(), Username: "a1", Email: "a1@t.com", CI: 1, FPV: 100,
			Solvent: false, IsActive: true, BornDate: now, Genre: "M",
			Nationality: "V", ContactEmail: "a1@t.com", ContactPhone: "111", FirstName: "A1", LastName: "T1", BioTextID: dummyBio.ID,
		})

		tx.Create(&domain.PsiUserModel{
			ID: uuid.New(), Username: "a2", Email: "a2@t.com", CI: 2, FPV: 200,
			Solvent: true, IsActive: false, BornDate: now, Genre: "F",
			Nationality: "V", ContactEmail: "a2@t.com", ContactPhone: "222", FirstName: "A2", LastName: "T2", BioTextID: dummyBio.ID,
		})

		res, total, err := r.SearchAdmin(ctx, request_structs.PsiDirectoryFilterDTO{Page: 1, Limit: 10})
		require.NoError(t, err)
		require.GreaterOrEqual(t, total, int64(2))
		require.Len(t, res, 2)
	})
}
