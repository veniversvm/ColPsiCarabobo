// Package postgres_test contiene las pruebas de integración para el adaptador de persistencia.
// Se asegura de que las consultas SQL generadas por GORM coincidan con las expectativas
// del dominio y las restricciones de PostgreSQL.
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
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
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
}
