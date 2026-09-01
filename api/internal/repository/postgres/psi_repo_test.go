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
		dsn = "host=localhost port=5433 user=postgres password=postgres dbname=postgres sslmode=disable"
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
	db.Exec("CREATE EXTENSION IF NOT EXISTS unaccent;")
	err = db.AutoMigrate(
		&domain.TextModel{},
		&domain.PsiUserModel{},
		&domain.PsiUserColData{},
		&domain.PsiUserPostGrade{},
		&domain.PsiUserSocialNetwork{},
		&domain.PsiSpecialtyModel{},
		&domain.PsiUserSolvency{},
		&domain.PsiODeontologia{},
	)
	require.NoError(t, err)

	// TRUCO SENIOR: Limpiar tablas antes de correr los tests
	db.Exec("TRUNCATE TABLE psi_user_social_networks RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE psi_user_post_grades RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE psi_user_col_data RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE psi_deontologia RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE psi_users RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE text_models RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE psi_specialty_models RESTART IDENTITY CASCADE")

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
				Solvent: true, Genre: "M", Nationality: "V",
				ContactEmail: "p1@t.com", ContactPhone: "04141234567",
				MunicipalityCarabobo: "Valencia", ShowMunicipalityCarabobo: true, // Visible
				PrimaryWorkArea:    "Neuropsicología",
				PrimarySpecialtyID: &spec.ID,
				BornDate:           now, BioTextID: dummyBio.ID, AudioBookShellId: "abs_p1",
				Credentials: domain.Credentials{
					IsActive: true,
					Username: "p1", Email: "p1@t.com",
				},
			},
			{
				// Maria: Solvente, Activa, Privacidad de Ubicación APAGADA (Oculta en búsquedas por zona)
				ID: uuid.New(), FirstName: "Maria", LastName: "Zeta", CI: 20, FPV: 20,
				Solvent: true, Genre: "F", Nationality: "V",
				ContactEmail: "m2@t.com", ContactPhone: "04120000000",
				MunicipalityCarabobo: "San Diego", ShowMunicipalityCarabobo: false, // Oculta
				BornDate: now, BioTextID: dummyBio.ID, AudioBookShellId: "abs_m2",
				Credentials: domain.Credentials{
					IsActive: true,
					Username: "m2", Email: "m2@t.com",
				},
			},
			{
				// Insolvente: No debe aparecer en búsquedas por navegación general
				ID: uuid.New(), FirstName: "Insolvente", LastName: "Busquedame", CI: 30, FPV: 30,
				Solvent: false, Genre: "M", Nationality: "V",
				ContactEmail: "i3@t.com", BornDate: now, BioTextID: dummyBio.ID, AudioBookShellId: "abs_i3",
				Credentials: domain.Credentials{
					IsActive: true,
					Username: "i3", Email: "i3@t.com",
				},
			},
			{
				// Baneado/Inactivo: Jamás debe aparecer en el directorio público
				ID: uuid.New(), FirstName: "Baneado", LastName: "Invisible", CI: 40, FPV: 40,
				Solvent: true, Genre: "M", Nationality: "V",
				ContactEmail: "b4@t.com", BornDate: now, BioTextID: dummyBio.ID, AudioBookShellId: "abs_b4",
				Credentials: domain.Credentials{
					IsActive: false,
					Username: "b4", Email: "b4@t.com",
				},
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
			ID: uuid.New(), CI: 999, FPV: 999, BornDate: time.Now(), BioTextID: dummyBio.ID,
			Genre: "M", Nationality: "V", ContactEmail: "del@t.com", ContactPhone: "123",
			FirstName: "Del", LastName: "Me", AudioBookShellId: "abs_del",
			Credentials: domain.Credentials{
				Username: "delete_me", Email: "del@t.com",
			},
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
			ID: uuid.New(), CI: 77, FPV: 77, BornDate: time.Now(), MiniBio: "Original",
			Genre: "F", Nationality: "V", ContactEmail: "upd@t.com", ContactPhone: "123",
			FirstName: "Upd", LastName: "User", AudioBookShellId: "abs_upd",
			BioTextID: bio.ID, // Asignación de FK obligatoria
			Credentials: domain.Credentials{
				Username: "upd", Email: "upd@t.com",
			},
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
			ID: uuid.New(), CI: 1, FPV: 100,
			Solvent: false, BornDate: now, Genre: "M",
			Nationality: "V", ContactEmail: "a1@t.com", ContactPhone: "111", FirstName: "A1", LastName: "T1", BioTextID: dummyBio.ID,
			AudioBookShellId: "abs_a1",
			Credentials: domain.Credentials{
				Username: "a1", Email: "a1@t.com", IsActive: true,
			},
		})

		// Usuario 2: Activo pero Insolvente
		tx.Create(&domain.PsiUserModel{
			ID: uuid.New(), CI: 2, FPV: 200,
			Solvent: true, BornDate: now, Genre: "F",
			Nationality: "V", ContactEmail: "a2@t.com", ContactPhone: "222", FirstName: "A2", LastName: "T2", BioTextID: dummyBio.ID,
			AudioBookShellId: "abs_a2",
			Credentials: domain.Credentials{
				Username: "a2", Email: "a2@t.com", IsActive: false,
			},
		})

		// Ejecutar búsqueda genérica desde el panel Admin
		res, total, err := r.SearchAdmin(ctx, request_structs.PsiDirectoryFilterDTO{Page: 1, Limit: 10})

		require.NoError(t, err)
		require.GreaterOrEqual(t, total, int64(2), "El conteo total debe incluir usuarios inactivos o insolventes")
		require.Len(t, res, 2, "La paginación debe recuperar a ambos usuarios sin importar sus banderas")
	})

	t.Run("SearchAdmin filters by solvency, status and gender", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		bio := domain.TextModel{ID: uuid.New(), Content: "..."}
		tx.Create(&bio)

		now := time.Now()
		create := func(ci int, solvent, isActive bool, genre string, abs string) {
			require.NoError(t, tx.Create(&domain.PsiUserModel{
				ID: uuid.New(), CI: ci, FPV: ci,
				Solvent: solvent, BornDate: now, Genre: genre,
				Nationality: "V", ContactEmail: "f@t.com", ContactPhone: "111",
				FirstName: "Filtro", LastName: "Test", BioTextID: bio.ID,
				AudioBookShellId: abs,
				Credentials: domain.Credentials{
					Username: abs, Email: abs + "@t.com", IsActive: isActive,
				},
			}).Error)
			if !isActive {
				require.NoError(t, tx.Model(&domain.PsiUserModel{}).
					Where("audio_book_shell_id = ?", abs).
					Update("is_active", false).Error)
			}
		}

		// Solvente + Activo + M
		create(3001, true, true, "M", "abs_f1")
		// Solvente + Inactivo + F
		create(3002, true, false, "F", "abs_f2")
		// Insolvente + Activo + F
		create(3003, false, true, "F", "abs_f3")
		// Insolvente + Inactivo + M
		create(3004, false, false, "M", "abs_f4")

		// Solo solventes
		_, total, err := r.SearchAdmin(ctx, request_structs.PsiDirectoryFilterDTO{
			Page: 1, Limit: 10, Solvent: boolPtr(true),
		})
		require.NoError(t, err)
		require.Equal(t, int64(2), total)

		// Solo inactivos
		_, total, err = r.SearchAdmin(ctx, request_structs.PsiDirectoryFilterDTO{
			Page: 1, Limit: 10, Active: boolPtr(false),
		})
		require.NoError(t, err)
		require.Equal(t, int64(2), total)

		// Solventes Y activos
		res, total, err := r.SearchAdmin(ctx, request_structs.PsiDirectoryFilterDTO{
			Page: 1, Limit: 10, Solvent: boolPtr(true), Active: boolPtr(true),
		})
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Equal(t, 3001, res[0].CI)

		// Solo género femenino
		_, total, err = r.SearchAdmin(ctx, request_structs.PsiDirectoryFilterDTO{
			Page: 1, Limit: 10, Gender: "F",
		})
		require.NoError(t, err)
		require.Equal(t, int64(2), total)
	})

	t.Run("CreateWithColData Atomic Insert", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		psi := &domain.PsiUserModel{
			ID: uuid.New(), CI: 5000, FPV: 5000, BornDate: time.Now(),
			Genre: "M", Nationality: "V", ContactEmail: "atomic@t.com", ContactPhone: "555",
			FirstName: "Atomic", LastName: "Test", AudioBookShellId: "abs_atomic",
			Credentials: domain.Credentials{Username: "atomic", Email: "atomic@t.com", IsActive: true},
		}
		col := &domain.PsiUserColData{
			UniversityUndergraduate: "UC",
		}
		solvencies := []domain.PsiUserSolvency{
			{ID: uuid.New(), Date: time.Now()},
		}

		err := r.CreateWithColData(ctx, psi, col, solvencies, nil)
		require.NoError(t, err)

		// Verify user, col data, and solvency all exist
		var user domain.PsiUserModel
		err = tx.First(&user, "id = ?", psi.ID).Error
		require.NoError(t, err)
		require.Equal(t, "Atomic", user.FirstName)

		var colCheck domain.PsiUserColData
		err = tx.First(&colCheck, "psi_user_model_id = ?", psi.ID).Error
		require.NoError(t, err)
		require.Equal(t, "UC", colCheck.UniversityUndergraduate)

		var solCheck domain.PsiUserSolvency
		err = tx.First(&solCheck, "psi_user_model_id = ?", psi.ID).Error
		require.NoError(t, err)
	})

	t.Run("GetByID Eager Loading", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		bio := domain.TextModel{ID: uuid.New(), Content: "<p>Full Bio</p>"}
		tx.Create(&bio)

		psi := domain.PsiUserModel{
			ID: uuid.New(), CI: 6000, FPV: 6000, BornDate: time.Now(),
			Genre: "F", Nationality: "V", ContactEmail: "eager@t.com", ContactPhone: "666",
			FirstName: "Eager", LastName: "Load", BioTextID: bio.ID, AudioBookShellId: "abs_eager",
			Credentials: domain.Credentials{Username: "eager", Email: "eager@t.com"},
		}
		tx.Create(&psi)

		col := domain.PsiUserColData{PsiUserModelID: psi.ID, UniversityUndergraduate: "UCV"}
		tx.Create(&col)

		sn := domain.PsiUserSocialNetwork{PsiUserID: psi.ID, Name: "LinkedIn", URL: "https://linkedin.com/test"}
		tx.Create(&sn)

		found, err := r.GetByID(ctx, psi.ID)
		require.NoError(t, err)
		require.Equal(t, "Eager", found.FirstName)
		require.Equal(t, "UCV", found.ColData.UniversityUndergraduate)
		require.Len(t, found.SocialNetworks, 1)
		require.Equal(t, "LinkedIn", found.SocialNetworks[0].Name)
	})

	t.Run("GetByIdentifier Login Logic", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		bio := domain.TextModel{ID: uuid.New(), Content: ""}
		tx.Create(&bio)

		psi := domain.PsiUserModel{
			ID: uuid.New(), CI: 7000, FPV: 7000, BornDate: time.Now(),
			Genre: "M", Nationality: "V", ContactEmail: "login@t.com", ContactPhone: "777",
			FirstName: "Login", LastName: "User", BioTextID: bio.ID, AudioBookShellId: "abs_login",
			Credentials: domain.Credentials{Username: "login_user", Email: "login@t.com"},
		}
		tx.Create(&psi)

		// Find by username
		found, err := r.GetByIdentifier(ctx, "login_user")
		require.NoError(t, err)
		require.Equal(t, psi.ID, found.ID)

		// Find by email
		found, err = r.GetByIdentifier(ctx, "login@t.com")
		require.NoError(t, err)
		require.Equal(t, psi.ID, found.ID)

		// Not found
		_, err = r.GetByIdentifier(ctx, "ghost")
		require.Error(t, err)
	})

	t.Run("UpdateKey Rotation", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		bio := domain.TextModel{ID: uuid.New(), Content: ""}
		tx.Create(&bio)

		psi := domain.PsiUserModel{
			ID: uuid.New(), CI: 8000, FPV: 8000, BornDate: time.Now(),
			Genre: "M", Nationality: "V", ContactEmail: "key@t.com", ContactPhone: "888",
			FirstName: "Key", LastName: "User", BioTextID: bio.ID, AudioBookShellId: "abs_key",
			Credentials: domain.Credentials{Username: "key_user", Email: "key@t.com", Key: "old_key_value"},
		}
		tx.Create(&psi)

		psi.Key = "new_secret_key"
		err := r.UpdateKey(ctx, &psi)
		require.NoError(t, err)

		var check domain.PsiUserModel
		tx.First(&check, psi.ID)
		require.Equal(t, "new_secret_key", check.Key)
	})

	t.Run("GetPsiUserColData and GetTextContentByID", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		bio := domain.TextModel{ID: uuid.New(), Content: "<p>Bio Content</p>"}
		tx.Create(&bio)

		psi := domain.PsiUserModel{
			ID: uuid.New(), CI: 9000, FPV: 9000, BornDate: time.Now(),
			Genre: "F", Nationality: "V", ContactEmail: "coldata@t.com", ContactPhone: "999",
			FirstName: "ColData", LastName: "Test", BioTextID: bio.ID, AudioBookShellId: "abs_coldata",
			Credentials: domain.Credentials{Username: "coldata", Email: "coldata@t.com"},
		}
		tx.Create(&psi)

		col := domain.PsiUserColData{PsiUserModelID: psi.ID, UniversityUndergraduate: "LUZ"}
		tx.Create(&col)

		foundCol, err := r.GetPsiUserColData(ctx, psi.ID)
		require.NoError(t, err)
		require.Equal(t, "LUZ", foundCol.UniversityUndergraduate)

		content, err := r.GetTextContentByID(ctx, bio.ID)
		require.NoError(t, err)
		require.Equal(t, "<p>Bio Content</p>", content)

		_, err = r.GetTextContentByID(ctx, uuid.New())
		require.Error(t, err)
	})

	t.Run("ValidateUniqueCredentials", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		bio := domain.TextModel{ID: uuid.New(), Content: ""}
		tx.Create(&bio)

		psi := domain.PsiUserModel{
			ID: uuid.New(), CI: 11000, FPV: 11000, BornDate: time.Now(),
			Genre: "M", Nationality: "V", ContactEmail: "uniq@t.com", ContactPhone: "1111",
			FirstName: "Unique", LastName: "Test", BioTextID: bio.ID, AudioBookShellId: "abs_uniq",
			Credentials: domain.Credentials{Username: "unique_user", Email: "uniq@t.com"},
		}
		tx.Create(&psi)

		// Duplicate username
		err := r.ValidateUniqueCredentials(ctx, "unique_user", "", uuid.Nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "nombre de usuario")

		// Duplicate email
		err = r.ValidateUniqueCredentials(ctx, "", "uniq@t.com", uuid.Nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "email")

		// Unique (excluding self)
		err = r.ValidateUniqueCredentials(ctx, "unique_user", "uniq@t.com", psi.ID)
		require.NoError(t, err)
	})

	t.Run("GetSitemapData", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		bio := domain.TextModel{ID: uuid.New(), Content: ""}
		tx.Create(&bio)

		// Active + Solvent — should appear
		tx.Create(&domain.PsiUserModel{
			ID: uuid.New(), CI: 12000, FPV: 12000, BornDate: time.Now(),
			Genre: "M", Nationality: "V", ContactEmail: "s1@t.com", ContactPhone: "1212",
			FirstName: "Sitemap", LastName: "One", BioTextID: bio.ID,
			AudioBookShellId: "abs_sm1", Solvent: true,
			Credentials: domain.Credentials{Username: "sm1", Email: "s1@t.com", IsActive: true},
		})

		// Inactive — should NOT appear
		inactive := domain.PsiUserModel{
			ID: uuid.New(), CI: 13000, FPV: 13000, BornDate: time.Now(),
			Genre: "F", Nationality: "V", ContactEmail: "s2@t.com", ContactPhone: "1313",
			FirstName: "Hidden", LastName: "Two", BioTextID: bio.ID,
			AudioBookShellId: "abs_sm2", Solvent: true,
			Credentials: domain.Credentials{Username: "sm2", Email: "s2@t.com", IsActive: true},
		}
		tx.Create(&inactive)
		tx.Model(&inactive).UpdateColumn("is_active", false)

		users, err := r.GetSitemapData(ctx)
		require.NoError(t, err)
		require.Len(t, users, 1)
		require.Equal(t, "Sitemap", users[0].FirstName)
	})

	t.Run("GetAllForABSSync", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		// TextModel es una Foreign Key obligatoria (BioTextID) para PsiUserModel.
		bio := domain.TextModel{ID: uuid.New(), Content: ""}
		require.NoError(t, tx.Create(&bio).Error)

		// Solvente activo — debe aparecer.
		solvent := domain.PsiUserModel{
			ID: uuid.New(), CI: 14000, FPV: 14000, BornDate: time.Now(),
			Genre: "M", Nationality: "V", ContactEmail: "abs1@t.com", ContactPhone: "1414",
			FirstName: "Abs", LastName: "Sync", BioTextID: bio.ID, AudioBookShellId: "abs_14000", Solvent: true,
			Credentials: domain.Credentials{Username: "abs1", Email: "abs1@t.com", IsActive: true},
		}
		require.NoError(t, tx.Create(&solvent).Error)

		// Insolvente activo — debe aparecer (para poder desactivarlo en ABS).
		insolvent := domain.PsiUserModel{
			ID: uuid.New(), CI: 15000, FPV: 15000, BornDate: time.Now(),
			Genre: "F", Nationality: "V", ContactEmail: "abs2@t.com", ContactPhone: "1515",
			FirstName: "Abs", LastName: "Insolvent", BioTextID: bio.ID, AudioBookShellId: "abs_15000", Solvent: false,
			Credentials: domain.Credentials{Username: "abs2", Email: "abs2@t.com", IsActive: true},
		}
		require.NoError(t, tx.Create(&insolvent).Error)

		// Soft-deleted — debe aparecer también (para revocar su acceso ABS).
		deleted := domain.PsiUserModel{
			ID: uuid.New(), CI: 16000, FPV: 16000, BornDate: time.Now(),
			Genre: "M", Nationality: "V", ContactEmail: "abs3@t.com", ContactPhone: "1616",
			FirstName: "Abs", LastName: "Deleted", BioTextID: bio.ID, AudioBookShellId: "abs_16000", Solvent: true,
			Credentials: domain.Credentials{Username: "abs3", Email: "abs3@t.com", IsActive: true},
		}
		require.NoError(t, tx.Create(&deleted).Error)
		require.NoError(t, tx.Delete(&deleted).Error)

		users, err := r.GetAllForABSSync(ctx)
		require.NoError(t, err)

		// Debe incluir solvente, insolvente y soft-deleted (Unscoped).
		found := map[int]domain.PsiUserModel{}
		for _, u := range users {
			found[u.CI] = u
		}
		require.Contains(t, found, 14000)
		require.Contains(t, found, 15000)
		require.Contains(t, found, 16000)
		require.True(t, found[14000].Solvent)
		require.False(t, found[15000].Solvent)
		require.True(t, found[16000].DeletedAt.Valid, "soft-deleted debe traer DeletedAt poblado")
	})

	t.Run("Deontologia CRUD", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		// TextModel es una Foreign Key obligatoria (BioTextID) para PsiUserModel.
		bio := domain.TextModel{ID: uuid.New(), Content: ""}
		require.NoError(t, tx.Create(&bio).Error)

		psi := domain.PsiUserModel{
			ID: uuid.New(), CI: 20000, FPV: 20000, BornDate: time.Now(),
			Genre: "M", Nationality: "V", ContactEmail: "deon@t.com", ContactPhone: "2020",
			FirstName: "Deon", LastName: "Prueba", BioTextID: bio.ID, AudioBookShellId: "abs_deon",
			Credentials: domain.Credentials{Username: "deon", Email: "deon@t.com", IsActive: true},
		}
		require.NoError(t, tx.Create(&psi).Error)

		// Crear dos entradas
		e1 := domain.PsiODeontologia{ID: uuid.New(), PsiUserID: psi.ID, Content: "Primera entrada"}
		e2 := domain.PsiODeontologia{ID: uuid.New(), PsiUserID: psi.ID, Content: "Segunda entrada"}
		require.NoError(t, r.CreateDeontologia(ctx, &e1))
		require.NoError(t, r.CreateDeontologia(ctx, &e2))

		// Listar: ambas presentes, la más reciente primero
		entries, err := r.ListDeontologiaByPsiID(ctx, psi.ID)
		require.NoError(t, err)
		require.Len(t, entries, 2)
		require.Equal(t, "Segunda entrada", entries[0].Content)

		// GetDeontologiaByID
		fetched, err := r.GetDeontologiaByID(ctx, e1.ID)
		require.NoError(t, err)
		require.Equal(t, "Primera entrada", fetched.Content)

		// Update: corrige el contenido de e1 (el expediente no se puede eliminar)
		require.NoError(t, r.UpdateDeontologia(ctx, e1.ID, "Primera entrada corregida", "admin", uuid.New()))
		fetched, err = r.GetDeontologiaByID(ctx, e1.ID)
		require.NoError(t, err)
		require.Equal(t, "Primera entrada corregida", fetched.Content)

		// Aislamiento entre psicólogos: otro psi no ve estas entradas
		otherPsi := domain.PsiUserModel{
			ID: uuid.New(), CI: 20001, FPV: 20001, BornDate: time.Now(),
			Genre: "F", Nationality: "V", ContactEmail: "deon2@t.com", ContactPhone: "2021",
			FirstName: "Otra", LastName: "Prueba", BioTextID: bio.ID, AudioBookShellId: "abs_deon2",
			Credentials: domain.Credentials{Username: "deon2", Email: "deon2@t.com", IsActive: true},
		}
		require.NoError(t, tx.Create(&otherPsi).Error)
		other := domain.PsiODeontologia{ID: uuid.New(), PsiUserID: otherPsi.ID, Content: "De otro"}
		require.NoError(t, r.CreateDeontologia(ctx, &other))
		entries, err = r.ListDeontologiaByPsiID(ctx, psi.ID)
		require.NoError(t, err)
		require.Len(t, entries, 1)
	})
}

func boolPtr(b bool) *bool { return &b }
