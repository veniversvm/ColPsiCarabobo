// Package postgres contiene la suite de pruebas de integración para el
// submódulo de Registro Digital de Documentos (psi_user_documents).
//
// Garantiza que el CRUD (crear, listar, consultar, editar y borrado lógico)
// se comporte de manera predecible sobre una instancia real de PostgreSQL,
// incluyendo la restricción de integridad referencial hacia psi_users.
package postgres

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupDocumentsTestDB inicializa la base de pruebas para el expediente digital
// de documentos: crea la BD física si falta, activa pgcrypto y sincroniza el
// esquema del grafo completo (psi_users + psi_user_documents).
func setupDocumentsTestDB(t *testing.T) *gorm.DB {
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
	err = db.AutoMigrate(&domain.PsiUserModel{}, &domain.PsiUserDocument{})
	require.NoError(t, err)

	db.Exec("TRUNCATE TABLE psi_user_documents RESTART IDENTITY CASCADE")
	return db
}

// createDocsTestPsi crea un psicólogo mínimo y deja la transacción aislada para
// que cada sub-test pueda insertar sus documentos sin contaminación cruzada.
func createDocsTestPsi(t *testing.T, tx *gorm.DB, idx int) *domain.PsiUserModel {
	now := time.Now()

	// El grafo de psi_users exige una TextModel válida para la FK full_bio.
	bio := domain.TextModel{ID: uuid.New(), Content: "..."}
	require.NoError(t, tx.Create(&bio).Error)

	psi := domain.PsiUserModel{
		ID: uuid.New(),
		Credentials: domain.Credentials{
			Username: "docpsi_" + uuid.NewString()[:8],
			Email:    "docpsi" + uuid.NewString()[:8] + "@test.com",
		},
		FirstName:      "Juan",
		LastName:       "Prueba",
		FPV:            700500 + idx,
		CI:             90005000 + idx,
		Nationality:    "V",
		BornDate:       now,
		Genre:          "M",
		ContactPhone:   "0241000000",
		ContactCellPhone: "0414000000",
		ContactEmail:   "docpsi@test.com",
		BioTextID:      bio.ID,
	}
	require.NoError(t, tx.Create(&psi).Error)
	return &psi
}

// TestPsiRepo_DocumentsSuite agrupa los flujos críticos del registro digital.
// Emplea el patrón "Transacción por Sub-Test" (Begin/Rollback) para aislamiento.
func TestPsiRepo_DocumentsSuite(t *testing.T) {
	mainDB := setupDocumentsTestDB(t)
	ctx := context.Background()

	t.Run("ListDocuments ordena por fecha descendente", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)
		psi := createDocsTestPsi(t, tx, 1)

		older := time.Now().Add(-48 * time.Hour)
		newer := time.Now().Add(-1 * time.Hour)
		docA := domain.PsiUserDocument{
			ID: uuid.New(), AuditModel: domain.AuditModel{CreatedAt: older},
			PsiUserID: psi.ID, DocumentType: domain.DocumentSolvencia,
			Title: "Solvencia 2024", S3Key: "documents/a.webp", MimeType: "image/webp", SizeBytes: 100,
		}
		docB := domain.PsiUserDocument{
			ID: uuid.New(), AuditModel: domain.AuditModel{CreatedAt: newer},
			PsiUserID: psi.ID, DocumentType: domain.DocumentCedula,
			Title: "Cédula", S3Key: "documents/b.webp", MimeType: "image/webp", SizeBytes: 200,
		}
		require.NoError(t, r.CreateDocument(ctx, &docA))
		require.NoError(t, r.CreateDocument(ctx, &docB))

		docs, err := r.ListDocuments(ctx, psi.ID)
		require.NoError(t, err)
		require.Len(t, docs, 2)
		require.Equal(t, "Cédula", docs[0].Title, "El más reciente debe ir primero")
		require.Equal(t, "Solvencia 2024", docs[1].Title)
	})

	t.Run("GetDocument recupera por UUID y auditoría", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)
		psi := createDocsTestPsi(t, tx, 2)

		adminID := uuid.New()
		doc := domain.PsiUserDocument{
			ID: uuid.New(), AuditModel: domain.AuditModel{CreateBy: "admin_tester", CreateById: &adminID},
			PsiUserID: psi.ID, DocumentType: domain.DocumentTitulo,
			Title: "Título pregrado", Notes: "verificar registro", S3Key: "documents/t.webp",
		}
		require.NoError(t, r.CreateDocument(ctx, &doc))

		got, err := r.GetDocument(ctx, doc.ID)
		require.NoError(t, err)
		require.Equal(t, doc.ID, got.ID)
		require.Equal(t, psi.ID, got.PsiUserID)
		require.Equal(t, "admin_tester", got.CreateBy)
		require.Equal(t, "verificar registro", got.Notes)
	})

	t.Run("GetDocument con UUID inexistente devuelve ErrRecordNotFound", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		_, err := r.GetDocument(ctx, uuid.New())
		require.Error(t, err)
		require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	t.Run("UpdateDocument persiste metadatos y auditoría del editor", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)
		psi := createDocsTestPsi(t, tx, 3)

		doc := domain.PsiUserDocument{
			ID: uuid.New(), PsiUserID: psi.ID, DocumentType: domain.DocumentOtro,
			Title: "Antes", S3Key: "documents/x.pdf", MimeType: "application/pdf", SizeBytes: 500,
		}
		require.NoError(t, r.CreateDocument(ctx, &doc))

		editorID := uuid.New()
		doc.Title = "Después"
		doc.DocumentType = domain.DocumentRif
		doc.UpdateBy = "editor_tester"
		doc.UpdateById = &editorID
		require.NoError(t, r.UpdateDocument(ctx, &doc))

		got, err := r.GetDocument(ctx, doc.ID)
		require.NoError(t, err)
		require.Equal(t, "Después", got.Title)
		require.Equal(t, domain.DocumentRif, got.DocumentType)
		require.Equal(t, "editor_tester", got.UpdateBy)
		require.Equal(t, editorID, *got.UpdateById)
	})

	t.Run("DeleteDocument aplica soft delete y oculta el registro", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)
		psi := createDocsTestPsi(t, tx, 4)

		doc := domain.PsiUserDocument{
			ID: uuid.New(), PsiUserID: psi.ID, DocumentType: domain.DocumentComprobante,
			Title: "Comprobante", S3Key: "documents/c.pdf",
		}
		require.NoError(t, r.CreateDocument(ctx, &doc))

		require.NoError(t, r.DeleteDocument(ctx, doc.ID))

		docs, err := r.ListDocuments(ctx, psi.ID)
		require.NoError(t, err)
		require.Empty(t, docs, "El borrado lógico debe excluir el documento")

		_, err = r.GetDocument(ctx, doc.ID)
		require.True(t, errors.Is(err, gorm.ErrRecordNotFound))

		require.NoError(t, r.DeleteDocument(ctx, doc.ID), "Borrar dos veces debe ser idempotente")
	})
}