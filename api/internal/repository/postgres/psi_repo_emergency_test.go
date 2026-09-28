// api/internal/repository/postgres/psi_repo_emergency_test.go
package postgres

// Suite de pruebas del repositorio de Persona de Contacto para Emergencias.
//
// Cubre lo que el mock de servicio no puede verificar contra PostgreSQL real:
// el preload de la relación (y su ausencia en la vista pública), el borrado
// lógico, la cuota y —lo más importante para la privacidad— que los datos de
// un tercero NO aparezcan en las consultas que alimentan el directorio.

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"gorm.io/gorm"
)

// nuevoPsiParaEmergencia crea un agremiado mínimo (con su TextModel de
// biografía, FK obligatoria) reutilizable por los sub-tests.
func nuevoPsiParaEmergencia(t *testing.T, tx *gorm.DB, ci int, username string) domain.PsiUserModel {
	t.Helper()
	bio := domain.TextModel{ID: uuid.New(), Content: ""}
	require.NoError(t, tx.Create(&bio).Error)

	psi := domain.PsiUserModel{
		ID: uuid.New(), CI: ci, FPV: ci, BornDate: time.Now(),
		Genre: "F", Nationality: "V", ContactEmail: username + "@t.com", ContactPhone: "04120000000",
		FirstName: "Ana", LastName: "Emergencia", BioTextID: bio.ID,
		// audio_book_shell_id tiene restricción UNIQUE: debe ser distinto por
		// agremiado o el INSERT choca con la clave duplicada.
		AudioBookShellId: "abs_" + username,
		Solvent:          true,
		Credentials:      domain.Credentials{Username: username, Email: username + "@t.com", IsActive: true},
	}
	require.NoError(t, tx.Create(&psi).Error)
	return psi
}

func TestPsiRepo_EmergencyContacts(t *testing.T) {
	mainDB := setupFullTestDB(t)
	ctx := context.Background()

	// El AutoMigrate de los tests no crea los CHECK del archivo de migración
	// (los ADD CONSTRAINT ... CHECK de Atlas no son derivables de los tags), así
	// que se replican aquí para probar también la defensa en profundidad.
	mainDB.Exec(`ALTER TABLE psi_user_emergency_contacts DROP CONSTRAINT IF EXISTS chk_emergency_contact_identified`)
	mainDB.Exec(`ALTER TABLE psi_user_emergency_contacts DROP CONSTRAINT IF EXISTS chk_emergency_contact_channel`)
	mainDB.Exec(`ALTER TABLE psi_user_emergency_contacts
		ADD CONSTRAINT chk_emergency_contact_identified CHECK (btrim(name) <> '' AND btrim(relationship) <> '')`)
	mainDB.Exec(`ALTER TABLE psi_user_emergency_contacts
		ADD CONSTRAINT chk_emergency_contact_channel CHECK (COALESCE(btrim(phone), '') <> '' OR COALESCE(btrim(email), '') <> '')`)

	t.Run("CRUD completo con preload en GetByID", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		psi := nuevoPsiParaEmergencia(t, tx, 31000, "emerg_crud")

		// Alta
		contacto := &domain.PsiUserEmergencyContact{
			ID:           uuid.New(),
			PsiUserID:    psi.ID,
			Name:         "María Rodríguez",
			Relationship: "madre",
			Phone:        "04121234567",
			Email:        "maria@correo.com",
		}
		require.NoError(t, r.CreateEmergencyContact(ctx, contacto))
		require.False(t, contacto.CreatedAt.IsZero(), "GORM debe sellar CreatedAt (orden por registro)")

		// Lectura individual
		obtenido, err := r.GetEmergencyContactByID(ctx, contacto.ID)
		require.NoError(t, err)
		require.Equal(t, "María Rodríguez", obtenido.Name)
		require.Equal(t, "04121234567", obtenido.Phone)

		// Listado
		lista, err := r.ListEmergencyContactsByPsiID(ctx, psi.ID)
		require.NoError(t, err)
		require.Len(t, lista, 1)

		// Cuota / conteo
		count, err := r.CountEmergencyContactsByPsiID(ctx, psi.ID)
		require.NoError(t, err)
		require.Equal(t, int64(1), count)

		// Edición (Updates con mapa explícito: un campo vacío debe poder guardarse)
		obtenido.Name = "  María   Rodríguez  "
		obtenido.Email = "" // se queda solo con teléfono
		obtenido.UpdateBy = "admin_test"
		require.NoError(t, r.UpdateEmergencyContact(ctx, obtenido))

		verificado, err := r.GetEmergencyContactByID(ctx, contacto.ID)
		require.NoError(t, err)
		require.Equal(t, "  María   Rodríguez  ", verificado.Name, "el repo guarda literal; normalizar es tarea del service")
		require.Equal(t, "", verificado.Email, "un campo vacío debe persistirse (no ignorarse por zero-value)")
		require.Equal(t, "04121234567", verificado.Phone)
		require.Equal(t, "admin_test", verificado.UpdateBy)

		// Preload en la vista de gestión
		conContactos, err := r.GetByID(ctx, psi.ID)
		require.NoError(t, err)
		require.Len(t, conContactos.EmergencyContacts, 1)
		require.Equal(t, "  María   Rodríguez  ", conContactos.EmergencyContacts[0].Name)

		// Borrado lógico
		require.NoError(t, r.DeleteEmergencyContact(ctx, contacto.ID))
		var borrado domain.PsiUserEmergencyContact
		err = tx.Unscoped().First(&borrado, "id = ?", contacto.ID).Error
		require.NoError(t, err, "la fila debe seguir existiendo físicamente (soft delete)")
		require.True(t, borrado.DeletedAt.Valid, "DeletedAt debe quedar poblado")

		// Y desaparecer de las consultas normales y del preload
		_, err = r.GetEmergencyContactByID(ctx, contacto.ID)
		require.Error(t, err, "un contacto borrado no debe ser recuperable")

		lista, err = r.ListEmergencyContactsByPsiID(ctx, psi.ID)
		require.NoError(t, err)
		require.Empty(t, lista)

		count, err = r.CountEmergencyContactsByPsiID(ctx, psi.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), count, "el borrado lógico libera la cuota")

		conContactos, err = r.GetByID(ctx, psi.ID)
		require.NoError(t, err)
		require.Empty(t, conContactos.EmergencyContacts)
	})

	t.Run("Orden de registro y aislamiento entre agremiados", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		psiA := nuevoPsiParaEmergencia(t, tx, 31100, "emerg_orden_a")
		psiB := nuevoPsiParaEmergencia(t, tx, 31200, "emerg_orden_b")

		// Tres contactos de A con created_at explícito para fijar el orden de
		// registro (el más antiguo se lista primero: es el contacto principal).
		primero := &domain.PsiUserEmergencyContact{ID: uuid.New(), PsiUserID: psiA.ID, Name: "Primero", Relationship: "padre", Phone: "04120000001"}
		segundo := &domain.PsiUserEmergencyContact{ID: uuid.New(), PsiUserID: psiA.ID, Name: "Segundo", Relationship: "madre", Phone: "04120000002"}
		tercero := &domain.PsiUserEmergencyContact{ID: uuid.New(), PsiUserID: psiA.ID, Name: "Tercero", Relationship: "hermano", Phone: "04120000003"}
		require.NoError(t, r.CreateEmergencyContact(ctx, primero))
		require.NoError(t, r.CreateEmergencyContact(ctx, segundo))
		require.NoError(t, r.CreateEmergencyContact(ctx, tercero))

		// Uno de B: no debe aparecer en el listado de A.
		ajeno := &domain.PsiUserEmergencyContact{ID: uuid.New(), PsiUserID: psiB.ID, Name: "Ajeno", Relationship: "amigo", Phone: "04120000009"}
		require.NoError(t, r.CreateEmergencyContact(ctx, ajeno))

		lista, err := r.ListEmergencyContactsByPsiID(ctx, psiA.ID)
		require.NoError(t, err)
		require.Len(t, lista, 3)
		require.Equal(t, "Primero", lista[0].Name, "el más antiguo se lista primero (contacto principal)")
		require.Equal(t, "Tercero", lista[2].Name)

		listaB, err := r.ListEmergencyContactsByPsiID(ctx, psiB.ID)
		require.NoError(t, err)
		require.Len(t, listaB, 1)
		require.Equal(t, "Ajeno", listaB[0].Name)

		// El preload de B no debe_traer los contactos de A.
		conB, err := r.GetByID(ctx, psiB.ID)
		require.NoError(t, err)
		require.Len(t, conB.EmergencyContacts, 1)
		require.Equal(t, "Ajeno", conB.EmergencyContacts[0].Name)
	})

	t.Run("PRIVACIDAD: la vista pública NO precarga los contactos de emergencia", func(t *testing.T) {
		tx := mainDB.Begin()
		defer tx.Rollback()
		r := NewPsiRepository(tx)

		psi := nuevoPsiParaEmergencia(t, tx, 31300, "emerg_privado")
		require.NoError(t, r.CreateEmergencyContact(ctx, &domain.PsiUserEmergencyContact{
			ID: uuid.New(), PsiUserID: psi.ID,
			Name: "Tercer Secreto", Relationship: "madre",
			Phone: "04129998888", Email: "secreto@correo.com",
		}))

		// GetByFPV es la fuente de la ficha pública (/psi/:fpv) y del detalle del
		// directorio: no debe traer la relación. Es la barrera más importante del
		// submódulo (datos de un tercero).
		publica, err := r.GetByFPV(ctx, psi.FPV)
		require.NoError(t, err)
		require.Empty(t, publica.EmergencyContacts, "GetByFPV no debe precargar contactos de emergencia")

		// El perfil del directorio tampoco.
		filtro := request_structs.PsiDirectoryFilterDTO{Page: 1, Limit: 20, SearchTerm: "Emergencia"}
		items, _, err := r.SearchDirectory(ctx, filtro)
		require.NoError(t, err)
		for _, item := range items {
			require.Equal(t, "Emergencia", item.LastName)
			require.Empty(t, item.EmergencyContacts, "el directorio no debe traer contactos de emergencia")
		}
	})

	t.Run("CHECK de la base de datos bloquea registros sin canal o sin nombre", func(t *testing.T) {
		// Cada caso de violación corre en su propia transacción: un CHECK violado
		// aborta la transacción completa en PostgreSQL, así que no se puede
		// seguir usando la misma para el siguiente intento.
		intentar := func(idx int, contacto *domain.PsiUserEmergencyContact) error {
			tx := mainDB.Begin()
			defer tx.Rollback()
			psi := nuevoPsiParaEmergencia(t, tx, 31400+idx, fmt.Sprintf("emerg_check_%d", idx))
			contacto.PsiUserID = psi.ID
			return NewPsiRepository(tx).CreateEmergencyContact(ctx, contacto)
		}

		// Sin teléfono ni correo: la regla "al menos uno de los dos" la sostiene
		// también la base de datos (defensa ante escrituras directas).
		err := intentar(1, &domain.PsiUserEmergencyContact{
			ID: uuid.New(), Name: "Sin Canal", Relationship: "amigo",
		})
		require.Error(t, err, "el CHECK chk_emergency_contact_channel debe rechazar el registro")

		// Sin nombre.
		err = intentar(2, &domain.PsiUserEmergencyContact{
			ID: uuid.New(), Name: "   ", Relationship: "amigo", Phone: "04121234567",
		})
		require.Error(t, err, "el CHECK chk_emergency_contact_identified debe rechazar el registro")

		// Sin parentesco.
		err = intentar(3, &domain.PsiUserEmergencyContact{
			ID: uuid.New(), Name: "Sin Parentesco", Relationship: "  ", Phone: "04121234567",
		})
		require.Error(t, err, "el CHECK chk_emergency_contact_identified debe rechazar el registro")

		// Válido (solo correo): pasa.
		require.NoError(t, intentar(4, &domain.PsiUserEmergencyContact{
			ID: uuid.New(), Name: "Válido", Relationship: "amigo", Email: "ok@correo.com",
		}))
	})
}
