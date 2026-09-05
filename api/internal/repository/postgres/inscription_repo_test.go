// Package postgres contiene la suite de pruebas de integración para el módulo
// de pre-inscripción (ficha de inscripción cercana a la ficha interna).
//
// Cubre: unicidad excluyente (cédula / FPV / correo) al editar la ficha y el
// CRUD de fotos de documentos requeridos (psi_inscription_documents).
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

// setupInscriptionTestDB inicializa la base de pruebas para el módulo de
// pre-inscripción (tablas psi_inscription_requests + psi_inscription_documents).
func setupInscriptionTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5433 user=postgres password=postgres dbname=postgres sslmode=disable"
	}

	tmpDb, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	tmpDb.Exec("CREATE DATABASE colpsi_test")
	sqlTmp, _ := tmpDb.DB()
	sqlTmp.Close()

	testDSN := strings.Replace(dsn, "dbname=postgres", "dbname=colpsi_test", 1)
	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{})
	require.NoError(t, err)

	db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";")
	err = db.AutoMigrate(&domain.PsiInscriptionRequest{}, &domain.PsiInscriptionDocument{})
	require.NoError(t, err)

	db.Exec("TRUNCATE TABLE psi_inscription_documents RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE psi_inscription_requests RESTART IDENTITY CASCADE")
	return db
}

// makeInscription crea una solicitud pendiente mínima para la suite.
func makeInscription(ci, fpv int, email string) *domain.PsiInscriptionRequest {
	return &domain.PsiInscriptionRequest{
		Cedula:       ci,
		Nacionalidad: "V",
		Nombres:      "Ana",
		Apellidos:    "Prueba",
		FPV:          fpv,
		Correo:       email,
		Status:       domain.InscriptionPending,
	}
}

// TestInscriptionRepo_FichaSuite agrupa los flujos de la ficha de inscripción.
// Emplea el patrón "Transacción por Sub-Test" (Begin/Rollback) para aislamiento.
func TestInscriptionRepo_FichaSuite(t *testing.T) {
	mainDB := setupInscriptionTestDB(t)
	ctx := context.Background()

	t.Run("Unicidad excluyente al editar la ficha", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewInscriptionRepository(tx)

		a := makeInscription(100, 400100, "a@test.com")
		require.NoError(t, r.Create(ctx, a))

		// Otra solicitud distinta con el mismo CI pendiente → existe.
		exists, err := r.ExistsPendingCI(ctx, 100)
		require.NoError(t, err)
		require.True(t, exists)

		// Excluyendo la propia solicitud → no existe OTRA con ese CI.
		exists, err = r.ExistsPendingCIExcluding(ctx, 100, a.ID)
		require.NoError(t, err)
		require.False(t, exists)

		// Excluyendo otra solicitud → sí existe.
		exists, err = r.ExistsPendingCIExcluding(ctx, 100, uuid.New())
		require.NoError(t, err)
		require.True(t, exists)

		// FPV
		exists, err = r.ExistsPendingFPVExcluding(ctx, a.FPV, a.ID)
		require.NoError(t, err)
		require.False(t, exists)

		// Correo (case-insensitive, excluyendo la propia).
		exists, err = r.ExistsPendingEmailExcluding(ctx, "A@TEST.COM", a.ID)
		require.NoError(t, err)
		require.False(t, exists)
	})

	t.Run("CRUD de fotos de documentos", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewInscriptionRepository(tx)

		req := makeInscription(200, 400200, "docs@test.com")
		require.NoError(t, r.Create(ctx, req))

		docs := []domain.PsiInscriptionDocument{
			{
				ID:                 uuid.New(),
				InscriptionRequestID: req.ID,
				DocumentType:       domain.DocumentCedula,
				S3Key:              "inscripciones/documentos/cedula/a.png",
				OriginalFilename:   "cedula.png",
			},
			{
				ID:                 uuid.New(),
				InscriptionRequestID: req.ID,
				DocumentType:       domain.DocumentTitulo,
				S3Key:              "inscripciones/documentos/titulo/b.png",
				OriginalFilename:   "titulo.png",
			},
		}
		require.NoError(t, r.CreateDocuments(ctx, docs))

		got, err := r.ListDocumentsByRequestID(ctx, req.ID)
		require.NoError(t, err)
		require.Len(t, got, 2)

		// Actualizar metadatos + clave S3.
		got[0].S3Key = "inscripciones/documentos/cedula/nuevo.png"
		require.NoError(t, r.UpdateInscriptionDocument(ctx, &got[0]))
		byID, err := r.GetInscriptionDocumentByID(ctx, got[0].ID)
		require.NoError(t, err)
		require.Equal(t, "inscripciones/documentos/cedula/nuevo.png", byID.S3Key)

		// Borrado físico.
		require.NoError(t, r.DeleteInscriptionDocument(ctx, got[0].ID))
		_, err = r.GetInscriptionDocumentByID(ctx, got[0].ID)
		require.Error(t, err)

		// Borrado masivo por solicitud.
		require.NoError(t, r.DeleteInscriptionDocumentsByRequestID(ctx, req.ID))
		remaining, err := r.ListDocumentsByRequestID(ctx, req.ID)
		require.NoError(t, err)
		require.Len(t, remaining, 0)
	})

	t.Run("Un documento por categoría (unique)", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewInscriptionRepository(tx)

		req := makeInscription(300, 400300, "unique@test.com")
		require.NoError(t, r.Create(ctx, req))

		require.NoError(t, r.CreateDocuments(ctx, []domain.PsiInscriptionDocument{
			{ID: uuid.New(), InscriptionRequestID: req.ID, DocumentType: domain.DocumentCedula, S3Key: "a.png"},
		}))
		dup := &domain.PsiInscriptionDocument{
			ID:                   uuid.New(),
			InscriptionRequestID: req.ID,
			DocumentType:         domain.DocumentCedula,
			S3Key:                "b.png",
		}
		require.Error(t, r.CreateDocuments(ctx, []domain.PsiInscriptionDocument{*dup}))
	})
}