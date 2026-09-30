// Package postgres_test contiene las pruebas de integración para el adaptador de persistencia.
// Se asegura de que las consultas SQL generadas por GORM coincidan con las expectativas
// del dominio y las restricciones de PostgreSQL.
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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestDB gestiona el ciclo de vida de la base de datos de pruebas.
//
// Realiza tres tareas críticas:
//  1. Asegura la existencia de una base de datos física dedicada a tests (colpsi_test).
//  2. Ejecuta las extensiones necesarias (pgcrypto para UUIDs si aplica).
//  3. Realiza la migración automática de esquemas para garantizar que las tablas
//     estén sincronizadas con los modelos del dominio.
func setupTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5433 user=postgres password=postgres dbname=postgres sslmode=disable"
	}

	// Paso 1: Conexión administrativa para garantizar la base de datos de pruebas.
	tmpDb, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	tmpDb.Exec("CREATE DATABASE colpsi_test")
	sqlTmp, _ := tmpDb.DB()
	sqlTmp.Close()

	// Paso 2: Conexión a la base de datos de pruebas definitiva.
	testDSN := strings.Replace(dsn, "dbname=postgres", "dbname=colpsi_test", 1)
	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{})
	require.NoError(t, err)

	// Paso 3: Configuración del entorno (Extensiones y Esquemas).
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";")
	err = db.AutoMigrate(&domain.TextModel{}, &domain.Post{})
	require.NoError(t, err)

	return db
}

// TestPostRepo_FullLifecycle valida el comportamiento del repositorio en un entorno real.
// Utiliza una estrategia de transacciones anidadas (Rollback) para asegurar que cada
// sub-test sea independiente y no deje basura en la base de datos.
func TestPostRepo_FullLifecycle(t *testing.T) {
	mainDB := setupTestDB(t)

	// Inyectamos una transacción en el repositorio para que los tests sean aislados.
	repoFactory := func(tx *gorm.DB) domain.PostRepository {
		return NewPostRepository(tx)
	}
	ctx := context.Background()

	// # Escenario: Listado, RBAC y Búsqueda
	// Verifica que la lógica de visibilidad y filtros de búsqueda funcione correctamente
	// a nivel de base de datos.
	t.Run("List with RBAC and Search Filters", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		repo := repoFactory(tx)

		// Preparación: Debido a las restricciones de Foreign Key, necesitamos un contenido base.
		dummyText := &domain.TextModel{ID: uuid.New(), Content: "Contenido de prueba"}
		tx.Create(dummyText)

		publishedStatus := domain.PostStatus("published")

		// Insertamos casos de prueba para validar la segregación de contenido (Público vs Gremial).
		tx.Create(&domain.Post{
			ID:     uuid.New(),
			Title:  "Aviso Publico",
			Type:   "public",
			Status: publishedStatus,
			TextID: dummyText.ID,
		})
		tx.Create(&domain.Post{
			ID:     uuid.New(),
			Title:  "Aviso Gremial",
			Type:   "psi",
			Status: publishedStatus,
			TextID: dummyText.ID,
		})

		// Ejecución: Filtramos solo contenido público.
		res, total, err := repo.List(ctx, domain.PostFilter{
			Type:   "public",
			Status: []domain.PostStatus{publishedStatus},
		}, 1, 10)

		// Verificación: Se debe ignorar el contenido 'psi' aunque esté publicado.
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Equal(t, "Aviso Publico", res[0].Title)
	})

	// # Escenario: Borrado Lógico (Soft Delete)
	// GORM implementa borrado lógico mediante el campo DeletedAt. Este test confirma
	// que los registros "borrados" permanecen en la DB pero son ignorados por las consultas.
	t.Run("Soft Delete Verification", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		repo := repoFactory(tx)

		// Preparación de datos.
		text := &domain.TextModel{ID: uuid.New(), Content: "Contenido a eliminar"}
		tx.Create(text)
		post := &domain.Post{ID: uuid.New(), Title: "To be deleted", TextID: text.ID}
		tx.Create(post)

		// Ejecución del borrado.
		err := repo.Delete(ctx, post.ID)
		require.NoError(t, err)

		// Verificación: Una consulta estándar no debe encontrar el registro.
		var check domain.Post
		err = tx.First(&check, "id = ?", post.ID).Error
		require.Error(t, err, "GORM debe retornar error porque el registro tiene fecha en deleted_at")
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("Create and GetByID with Eager Loading", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		repo := repoFactory(tx)

		text := &domain.TextModel{ID: uuid.New(), Content: "<p>Contenido Completo</p>"}
		post := &domain.Post{
			ID:    uuid.New(),
			Title: "Post con Contenido",
			Type:  "public",
		}

		err := repo.Create(ctx, post, text)
		require.NoError(t, err)
		require.Equal(t, text.ID, post.TextID, "Create debe vincular el TextID automáticamente")

		found, err := repo.GetByID(ctx, post.ID)
		require.NoError(t, err)
		require.Equal(t, "Post con Contenido", found.Title)
		require.Equal(t, "<p>Contenido Completo</p>", found.Text.Content)
	})

	// # Escenario: Actualizar el cuerpo de una publicación que NO tiene fila de texto
	// Las publicaciones anteriores a la creación del TextModel quedaron con
	// posts.text_id = NULL, que GORM carga como uuid.Nil. Un Update sobre un modelo
	// con primary key en cero no genera WHERE y GORM lo rechaza con
	// "WHERE conditions required" (500). El repo debe crear la fila y enlazarla.
	t.Run("Update crea y enlaza el texto si el post no tiene text_id", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		repo := repoFactory(tx)

		// Preparación: post huérfano de texto, tal como queda en producción
		// (text_id NULL). Se inserta con SQL crudo porque el AutoMigrate del banco
		// de pruebas deja text_id NOT NULL (uuid.UUID es tipo valor, GORM nunca
		// envía NULL), mientras que la migración de producción sí lo permite.
		postID := uuid.New()
		require.NoError(t, tx.Exec(
			`INSERT INTO posts (id, title, type, status, text_id) VALUES (?::uuid, ?, 'public', 'published', NULL)`,
			postID, "Noticia sin cuerpo").Error)

		var loaded domain.Post
		require.NoError(t, tx.First(&loaded, "id = ?", postID).Error)
		require.Equal(t, uuid.Nil, loaded.TextID, "text_id NULL debe cargarse como uuid.Nil")

		// Ejecución: el servicio entrega el TextModel con el ID que venía del post.
		text := &domain.TextModel{ID: loaded.TextID, Content: "<p>Cuerpo nuevo</p>"}
		require.NoError(t, repo.Update(ctx, &loaded, text))

		// Verificación: se creó la fila de texto y el post quedó enlazado a ella.
		require.NotEqual(t, uuid.Nil, text.ID, "el repo debe generar el ID de la fila creada")
		require.Equal(t, text.ID, loaded.TextID)

		found, err := repo.GetByID(ctx, postID)
		require.NoError(t, err)
		require.Equal(t, text.ID, found.TextID)
		require.Equal(t, "<p>Cuerpo nuevo</p>", found.Text.Content)
	})

	// # Escenario: el mismo post, ahora con texto, se actualiza en el sitio (no crea otra fila)
	t.Run("Update reutiliza la fila de texto existente", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		repo := repoFactory(tx)

		text := &domain.TextModel{ID: uuid.New(), Content: "contenido original"}
		require.NoError(t, tx.Create(text).Error)
		post := &domain.Post{ID: uuid.New(), Title: "Noticia con cuerpo", Type: "public", Status: domain.PostStatusPublished, TextID: text.ID}
		require.NoError(t, tx.Create(post).Error)

		var before, after int64
		require.NoError(t, tx.Model(&domain.TextModel{}).Count(&before).Error)

		require.NoError(t, repo.Update(ctx, post, &domain.TextModel{ID: post.TextID, Content: "contenido editado"}))

		require.NoError(t, tx.Model(&domain.TextModel{}).Count(&after).Error)
		require.Equal(t, before, after, "no debe crear una fila nueva si el texto ya existía")

		var reread domain.TextModel
		require.NoError(t, tx.First(&reread, "id = ?", text.ID).Error)
		require.Equal(t, "contenido editado", reread.Content)
	})

	// # Escenario: nil en el contenido no toca la fila de texto (PATCH parcial de metadata)
	t.Run("Update sin contenido no toca el texto", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		repo := repoFactory(tx)

		text := &domain.TextModel{ID: uuid.New(), Content: "cuerpo intacto"}
		require.NoError(t, tx.Create(text).Error)
		post := &domain.Post{ID: uuid.New(), Title: "Titulo viejo", Type: "public", TextID: text.ID}
		require.NoError(t, tx.Create(post).Error)

		post.Title = "Titulo nuevo"
		require.NoError(t, repo.Update(ctx, post, nil))

		found, err := repo.GetByID(ctx, post.ID)
		require.NoError(t, err)
		require.Equal(t, "Titulo nuevo", found.Title)
		require.Equal(t, "cuerpo intacto", found.Text.Content)
	})

	t.Run("PublishScheduled transitions due posts", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		repo := repoFactory(tx)

		text := &domain.TextModel{ID: uuid.New(), Content: "scheduled content"}
		tx.Create(text)

		past := time.Now().Add(-1 * time.Hour)
		future := time.Now().Add(1 * time.Hour)

		// Post scheduled in the past — should be published
		scheduledPast := domain.Post{
			ID: uuid.New(), Title: "Due Post", Type: "public",
			Status: domain.PostStatusScheduled, TextID: text.ID, PublishAt: &past,
		}
		tx.Create(&scheduledPast)

		// Post scheduled in the future — should stay scheduled
		text2 := &domain.TextModel{ID: uuid.New(), Content: "future content"}
		tx.Create(text2)
		scheduledFuture := domain.Post{
			ID: uuid.New(), Title: "Future Post", Type: "public",
			Status: domain.PostStatusScheduled, TextID: text2.ID, PublishAt: &future,
		}
		tx.Create(&scheduledFuture)

		affected := repo.PublishScheduled(ctx)
		require.Equal(t, int64(1), affected, "Solo el post vencido debe ser publicado")

		var checkPast domain.Post
		tx.First(&checkPast, scheduledPast.ID)
		require.Equal(t, domain.PostStatusPublished, checkPast.Status)

		var checkFuture domain.Post
		tx.First(&checkFuture, scheduledFuture.ID)
		require.Equal(t, domain.PostStatusScheduled, checkFuture.Status)
	})

	t.Run("GetSitemapPosts returns only published public posts", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		repo := repoFactory(tx)

		text := &domain.TextModel{ID: uuid.New(), Content: "sitemap content"}
		tx.Create(text)

		tx.Create(&domain.Post{
			ID: uuid.New(), Title: "Published Public", Type: "public",
			Status: domain.PostStatusPublished, TextID: text.ID,
		})
		tx.Create(&domain.Post{
			ID: uuid.New(), Title: "Published Psi", Type: "psi",
			Status: domain.PostStatusPublished, TextID: text.ID,
		})
		tx.Create(&domain.Post{
			ID: uuid.New(), Title: "Draft Public", Type: "public",
			Status: domain.PostStatusDraft, TextID: text.ID,
		})

		posts, err := repo.GetSitemapPosts(ctx)
		require.NoError(t, err)
		require.Len(t, posts, 1, "Solo publicaciones publicadas y públicas")
		require.Equal(t, "Published Public", posts[0].Title)
	})

	t.Run("List with Search filter", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		repo := repoFactory(tx)

		text := &domain.TextModel{ID: uuid.New(), Content: "search content"}
		tx.Create(text)

		tx.Create(&domain.Post{
			ID: uuid.New(), Title: "Aviso Importante", Type: "public",
			Status: domain.PostStatusPublished, TextID: text.ID,
		})
		tx.Create(&domain.Post{
			ID: uuid.New(), Title: "Otro Aviso", Type: "public",
			Status: domain.PostStatusPublished, TextID: text.ID,
		})
		tx.Create(&domain.Post{
			ID: uuid.New(), Title: "Noticia General", Type: "public",
			Status: domain.PostStatusPublished, TextID: text.ID,
		})

		res, total, err := repo.List(ctx, domain.PostFilter{
			Search: "Importante",
			Status: []domain.PostStatus{domain.PostStatusPublished},
		}, 1, 10)
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Equal(t, "Aviso Importante", res[0].Title)
	})

	t.Run("List pagination and all_visible type", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		repo := repoFactory(tx)

		text := &domain.TextModel{ID: uuid.New(), Content: "page content"}
		tx.Create(text)

		tx.Create(&domain.Post{
			ID: uuid.New(), Title: "Public 1", Type: "public",
			Status: domain.PostStatusPublished, TextID: text.ID,
		})
		tx.Create(&domain.Post{
			ID: uuid.New(), Title: "Psi 1", Type: "psi",
			Status: domain.PostStatusPublished, TextID: text.ID,
		})

		// all_visible should return both public and psi
		res, total, err := repo.List(ctx, domain.PostFilter{
			Type: "all_visible",
		}, 1, 10)
		require.NoError(t, err)
		require.Equal(t, int64(2), total)
		require.Len(t, res, 2)

		// Page 1, limit 1
		res, total, err = repo.List(ctx, domain.PostFilter{}, 1, 1)
		require.NoError(t, err)
		require.Equal(t, int64(2), total)
		require.Len(t, res, 1)
	})
}
