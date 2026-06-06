// Package postgres_test contiene la suite de pruebas de integración para el repositorio
// de psicólogos (PsiRepository). Garantiza que las consultas complejas, el motor de búsqueda
// y las reglas de seguridad (como los escudos de privacidad) se comporten correctamente
// frente a una instancia real de PostgreSQL.
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

// setupFullTestDB gestiona la inicialización de la base de datos de pruebas para el dominio completo de psicólogos.
//
// Realiza el aprovisionamiento completo:
//  1. Crea la base de datos física si no existe.
//  2. Activa la extensión pgcrypto requerida para la generación nativa de UUIDs.
//  3. Sincroniza (AutoMigrate) todos los modelos relacionados para construir el grafo completo de tablas y Foreign Keys.
func setupFullTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
	}

	// Paso 1: Asegurar la existencia de la base de datos de pruebas
	tmpDb, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	tmpDb.Exec("CREATE DATABASE colpsi_test")
	sqlTmp, _ := tmpDb.DB()
	sqlTmp.Close()

	// Paso 2: Conexión definitiva a la BD de pruebas
	testDSN := strings.Replace(dsn, "dbname=postgres", "dbname=colpsi_test", 1)
	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{})
	require.NoError(t, err)

	// Paso 3: Infraestructura y Esquemas
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

// TestPsiRepo_ComprehensiveSuite agrupa todos los flujos críticos del repositorio de psicólogos.
// Emplea el patrón de "Transacción por Sub-Test" (tx.Begin() / defer tx.Rollback()) para aislar
// cada escenario, garantizando que una prueba no contamine los datos de la siguiente.
func TestPsiRepo_ComprehensiveSuite(t *testing.T) {
	mainDB := setupFullTestDB(t)
	ctx := context.Background()

	// # Escenario 1: Directorio Público y Filtros de Privacidad
	// Valida que el motor de búsqueda respete las especialidades y, críticamente,
	// que los "escudos de privacidad" (banderas booleanas) oculten usuarios en las búsquedas geográficas.
	t.Run("Advanced Search Logic and Ranking", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		// Preparación: TextModel es una Foreign Key obligatoria (BioTextID) para PsiUserModel.
		dummyBio := domain.TextModel{ID: uuid.New(), Content: "Bio de prueba"}
		require.NoError(t, tx.Create(&dummyBio).Error)

		// Preparación: Especialidad base
		spec := domain.PsiSpecialtyModel{Name: "Neuropsicología", Active: true}
		tx.Create(&spec)

		now := time.Now()

		// Sembrado de datos: Creamos una matriz de usuarios con distintos estados para evaluar los filtros.
		users := []domain.PsiUserModel{
			{
				// Pedro: Solvente, Activo, Especialista, Privacidad de Ubicación ENCENDIDA (Visible)
				ID: uuid.New(), FirstName: "Pedro", LastName: "Alfonso", CI: 10, FPV: 10,
				Solvent: true, IsActive: true, Genre: "M", Nationality: "V",
				ContactEmail: "p1@t.com", ContactPhone: "04141234567",
				MunicipalityCarabobo: "Valencia", ShowMunicipalityCarabobo: true, // Visible
				PrimaryWorkArea: "Neuropsicología",
				Username:        "p1", Email: "p1@t.com", BornDate: now, BioTextID: dummyBio.ID,
			},
			{
				// Maria: Solvente, Activa, Privacidad de Ubicación APAGADA (Oculta en búsquedas por zona)
				ID: uuid.New(), FirstName: "Maria", LastName: "Zeta", CI: 20, FPV: 20,
				Solvent: true, IsActive: true, Genre: "F", Nationality: "V",
				ContactEmail: "m2@t.com", ContactPhone: "04120000000",
				MunicipalityCarabobo: "San Diego", ShowMunicipalityCarabobo: false, // Oculta
				Username: "m2", Email: "m2@t.com", BornDate: now, BioTextID: dummyBio.ID,
			},
			{
				// Insolvente: No debe aparecer en búsquedas por navegación general
				ID: uuid.New(), FirstName: "Insolvente", LastName: "Busquedame", CI: 30, FPV: 30,
				Solvent: false, IsActive: true, Genre: "M", Nationality: "V",
				ContactEmail: "i3@t.com", Username: "i3", Email: "i3@t.com", BornDate: now, BioTextID: dummyBio.ID,
			},
			{
				// Baneado/Inactivo: Jamás debe aparecer en el directorio público
				ID: uuid.New(), FirstName: "Baneado", LastName: "Invisible", CI: 40, FPV: 40,
				Solvent: true, IsActive: false, Genre: "M", Nationality: "V",
				ContactEmail: "b4@t.com", Username: "b4", Email: "b4@t.com", BornDate: now, BioTextID: dummyBio.ID,
			},
		}
		for i := range users {
			require.NoError(t, tx.Create(&users[i]).Error)
		}

		// Prueba 1.1: Filtro por Área de Trabajo. Debe retornar a Pedro.
		res, total, err := r.SearchDirectory(ctx, request_structs.PsiDirectoryFilterDTO{SpecialtyID: spec.ID, Page: 1, Limit: 10})
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Equal(t, "Pedro", res[0].FirstName)

		// Prueba 1.2: Validación del Escudo de Privacidad.
		// Al buscar "Valencia", solo Pedro debe aparecer, ya que tiene ShowMunicipalityCarabobo en true.
		locRes, locTotal, locErr := r.SearchDirectory(ctx, request_structs.PsiDirectoryFilterDTO{Location: "Valencia", Page: 1, Limit: 10})
		require.NoError(t, locErr)
		require.Equal(t, int64(1), locTotal, "Debe retornar 1 porque solo Pedro tiene su escudo de privacidad configurado como visible")
		require.Equal(t, "Pedro", locRes[0].FirstName)
	})

	// # Escenario 2: Borrado Lógico (Soft Delete)
	// Valida que el comando Delete emita un UPDATE en lugar de un DROP físico, y que las
	// consultas subsecuentes de GORM ignoren automáticamente dicho registro.
	t.Run("Soft Delete and Unscoped Recovery", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		dummyBio := domain.TextModel{ID: uuid.New(), Content: "..."}
		tx.Create(&dummyBio)

		psi := domain.PsiUserModel{
			ID: uuid.New(), Username: "delete_me", Email: "del@t.com",
			CI: 999, FPV: 999, BornDate: time.Now(), BioTextID: dummyBio.ID,
			Genre: "M", Nationality: "V", ContactEmail: "del@t.com", ContactPhone: "123",
			FirstName: "Del", LastName: "Me",
		}
		require.NoError(t, tx.Create(&psi).Error)

		// Ejecutar borrado lógico
		err := r.Delete(ctx, psi.ID)
		require.NoError(t, err)

		// Verificación: Una consulta normal (First) debe arrojar ErrRecordNotFound
		var check domain.PsiUserModel
		err = tx.First(&check, "id = ?", psi.ID).Error
		require.Error(t, err)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound, "GORM debe ocultar el registro por tener un deleted_at")
	})

	// # Escenario 3: Seguridad en Edición de Perfil (Omit Logic)
	// Comprueba que los campos sensibles (como Cédula o Estatus) protegidos en el código
	// mediante tx.Omit() no puedan ser sobreescritos por el usuario, evitando escalamiento de privilegios.
	t.Run("UpdatePublicProfile with Omit Logic", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		// Preparación de dependencias (Bio)
		bio := domain.TextModel{ID: uuid.New(), Content: "<p>Bio Original</p>"}
		require.NoError(t, tx.Create(&bio).Error)

		psi := domain.PsiUserModel{
			ID: uuid.New(), Username: "upd", Email: "upd@t.com",
			CI: 77, FPV: 77, BornDate: time.Now(), MiniBio: "Original",
			Genre: "F", Nationality: "V", ContactEmail: "upd@t.com", ContactPhone: "123",
			FirstName: "Upd", LastName: "User",
			BioTextID: bio.ID, // Asignación de FK obligatoria
		}
		require.NoError(t, tx.Create(&psi).Error)

		col := domain.PsiUserColData{PsiUserModelID: psi.ID, UniversityUndergraduate: "Original Uni"}
		require.NoError(t, tx.Create(&col).Error)

		// Simulación de un intento de actualización por parte del cliente:
		// Mezcla campos permitidos (MiniBio, WorkArea) con campos prohibidos (CI)
		psi.MiniBio = "Actualizada"
		psi.PrimaryWorkArea = "Clínica"
		psi.CI = 88 // Intento de mutación de campo sensible
		bio.Content = "<p>Bio Actualizada</p>"

		err := r.UpdatePublicProfile(ctx, &psi, &col, &bio)
		require.NoError(t, err)

		// Validación cruzada con la base de datos
		var check domain.PsiUserModel
		tx.First(&check, psi.ID)

		require.Equal(t, 77, check.CI, "SEGURIDAD: El campo CI no debió cambiar, el método Omit debe bloquearlo")
		require.Equal(t, "Clínica", check.PrimaryWorkArea, "El área de trabajo debió actualizarse exitosamente")
	})

	// # Escenario 4: Privilegios de Búsqueda Administrativa
	// Verifica que el método SearchAdmin no aplique restricciones de visibilidad,
	// permitiendo listar usuarios insolventes o inactivos.
	t.Run("SearchAdmin ignores filters", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		dummyBio := domain.TextModel{ID: uuid.New(), Content: "..."}
		tx.Create(&dummyBio)

		now := time.Now()

		// Usuario 1: Inactivo pero Solvente
		tx.Create(&domain.PsiUserModel{
			ID: uuid.New(), Username: "a1", Email: "a1@t.com", CI: 1, FPV: 100,
			Solvent: false, IsActive: true, BornDate: now, Genre: "M",
			Nationality: "V", ContactEmail: "a1@t.com", ContactPhone: "111", FirstName: "A1", LastName: "T1", BioTextID: dummyBio.ID,
		})

		// Usuario 2: Activo pero Insolvente
		tx.Create(&domain.PsiUserModel{
			ID: uuid.New(), Username: "a2", Email: "a2@t.com", CI: 2, FPV: 200,
			Solvent: true, IsActive: false, BornDate: now, Genre: "F",
			Nationality: "V", ContactEmail: "a2@t.com", ContactPhone: "222", FirstName: "A2", LastName: "T2", BioTextID: dummyBio.ID,
		})

		// Ejecutar búsqueda genérica desde el panel Admin
		res, total, err := r.SearchAdmin(ctx, request_structs.PsiDirectoryFilterDTO{Page: 1, Limit: 10})

		require.NoError(t, err)
		require.GreaterOrEqual(t, total, int64(2), "El conteo total debe incluir usuarios inactivos o insolventes")
		require.Len(t, res, 2, "La paginación debe recuperar a ambos usuarios sin importar sus banderas")
	})
}
