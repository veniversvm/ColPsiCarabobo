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

// setupTestDB gestiona la conexión, creación de DB de test y migración inicial.
func setupTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
	}

	// 1. Conectar a postgres para asegurar existencia de la db de test
	tmpDb, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	// Ignoramos el error si ya existe
	tmpDb.Exec("CREATE DATABASE colpsi_test")
	sqlTmp, _ := tmpDb.DB()
	sqlTmp.Close()

	// 2. Conectar a la db de test
	testDSN := strings.Replace(dsn, "dbname=postgres", "dbname=colpsi_test", 1)
	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{})
	require.NoError(t, err)

	// 3. Migración
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";")
	err = db.AutoMigrate(&domain.TextModel{}, &domain.Post{})
	require.NoError(t, err)

	return db
}

func TestPostRepo_FullLifecycle(t *testing.T) {
	mainDB := setupTestDB(t)

	repoFactory := func(tx *gorm.DB) domain.PostRepository {
		return NewPostRepository(tx)
	}
	ctx := context.Background()

	// --- TEST: LIST & FILTERS ---
	t.Run("List with RBAC and Search Filters", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		repo := repoFactory(tx)

		// 1. Necesitamos un Texto real para los posts debido a la FK
		dummyText := &domain.TextModel{ID: uuid.New(), Content: "Contenido"}
		tx.Create(dummyText)

		// 2. Seed de datos con el Status en lugar de IsActive
		// Asumimos "published" como el estado activo para el test
		publishedStatus := domain.PostStatus("published")

		tx.Create(&domain.Post{
			ID:     uuid.New(),
			Title:  "Aviso Publico",
			Type:   "public",
			Status: publishedStatus, // Cambiado IsActive -> Status
			TextID: dummyText.ID,
		})
		tx.Create(&domain.Post{
			ID:     uuid.New(),
			Title:  "Aviso Gremial",
			Type:   "psi",
			Status: publishedStatus, // Cambiado IsActive -> Status
			TextID: dummyText.ID,
		})

		// 3. Validar filtros (asumiendo que PostFilter ahora usa Status)
		res, total, err := repo.List(ctx, domain.PostFilter{
			Type:   "public",
			Status: []domain.PostStatus{publishedStatus}, // Cambiado IsActive -> Status
		}, 1, 10)

		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Equal(t, "Aviso Publico", res[0].Title)
	})

	// --- TEST: SOFT DELETE ---
	t.Run("Soft Delete Verification", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		repo := repoFactory(tx)

		// Satisfacer FK
		text := &domain.TextModel{ID: uuid.New(), Content: "..."}
		tx.Create(text)

		post := &domain.Post{ID: uuid.New(), Title: "To be deleted", TextID: text.ID}
		tx.Create(post)

		err := repo.Delete(ctx, post.ID)
		require.NoError(t, err)

		var check domain.Post
		err = tx.First(&check, "id = ?", post.ID).Error
		require.Error(t, err, "GORM no debe encontrar registros con deleted_at")
	})
}
