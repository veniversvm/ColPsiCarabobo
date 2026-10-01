package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
)

// TestPsiRepo_FiltroDeAreaCoincideConElChipEsElContratoQueEstabaRoto.
//
// El bug: el área que la tarjeta MUESTRA se resolvía por FK + coincidencia exacta
// del legacy (COALESCE(sp1.name, sp1n.name, ”)) pero el WHERE filtraba solo por
// la FK (primary_specialty_id = ?). En los datos la FK está poblada en 2 de 103
// agremiados, así que el directorio pintaba un chip que su propio filtro no
// encontraba: 13 tarjetas con "Clínica" y el filtro devolvía 2, y las 7 de
// "Neuropsicología" devolvían 0.
//
// Aquí se monta exactamente esa asimetría (mitad con FK, mitad solo con el texto
// legacy) y se exige que filtrar devuelva lo mismo que se muestra. Si alguien
// vuelve a filtrar por FK, esto falla.
func TestPsiRepo_FiltroDeAreaCoincideConElChip(t *testing.T) {
	mainDB := setupFullTestDB(t)
	ctx := context.Background()

	tx := mainDB.Begin()
	defer tx.Rollback()

	r := NewPsiRepository(tx)

	// `bio_text_id` es una FK obligatoria y el campo Go es `uuid.UUID` POR VALOR:
	// sin este TextModel, un struct literal vacío escribe uuid.Nil y el INSERT
	// revienta con 23503 (gotcha 19).
	dummyBio := domain.TextModel{ID: uuid.New(), Content: "Bio de prueba"}
	require.NoError(t, tx.Create(&dummyBio).Error)

	clinica := domain.PsiSpecialtyModel{Name: "Clínica", Active: true}
	educativa := domain.PsiSpecialtyModel{Name: "Educativa", Active: true}
	require.NoError(t, tx.Create(&clinica).Error)
	require.NoError(t, tx.Create(&educativa).Error)

	crear := func(nombre string, ci, fpv int, area string, specialtyID *uint32) {
		u := domain.PsiUserModel{
			FirstName: nombre, LastName: "Prueba",
			CI: ci, FPV: fpv, Nationality: "V", BioTextID: dummyBio.ID,
			// `audio_book_shell_id` es UNIQUE y el string vacío cuenta como valor:
			// dos usuarios sin valor chocan con 23505.
			AudioBookShellId:   "abs_" + nombre,
			Solvent:            true,
			PrimaryWorkArea:    area,
			PrimarySpecialtyID: specialtyID,
			Credentials:        domain.Credentials{Username: "u" + nombre, Email: nombre + "@t.com", Password: "x", IsActive: true},
		}
		require.NoError(t, tx.Create(&u).Error)
	}

	crear("Ana", 111, 9001, "", &clinica.ID) // por FK
	crear("Luis", 222, 9002, "Clínica", nil) // solo texto legacy
	crear("Carlos", 333, 9003, "Educativa", nil)
	crear("Marta", 444, 9004, "Deportiva", nil) // legacy que NO está en el catálogo

	buscar := func(id uint32) ([]domain.PsiUserModel, int64, error) {
		return r.SearchDirectory(ctx, request_structs.PsiDirectoryFilterDTO{
			SpecialtyID: id, Page: 1, Limit: 50,
		})
	}

	// El caso que fallaba: el filtro por FK devolvía 1 (solo Ana) en vez de 2.
	rows, total, err := buscar(clinica.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), total,
		"filtrar por Clínica debe devolver a los 2 solventes que MUESTRAN Clínica (uno por FK, otro solo por legacy)")

	nombres := make([]string, 0, len(rows))
	for _, row := range rows {
		nombres = append(nombres, row.FirstName)
	}
	require.ElementsMatch(t, []string{"Ana", "Luis"}, nombres,
		"el filtro debe traer tanto el que tiene la FK como el que solo tiene el legacy")

	// El área que no tiene a nadie con FK pero sí con legacy (el caso de
	// Neuropsicología, que devolvía 0 y dejaba el directorio vacío).
	rows, total, err = buscar(educativa.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, rows, 1)
	require.Equal(t, "Educativa", rows[0].PrimaryWorkArea)

	// Un legacy fuera del catálogo no debe aparecer por ninguna vía (gotcha 11:
	// jamás un área inventada).
	_, total, err = buscar(999)
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
}

// TestPsiRepo_FiltroDeAreaDelPanelAdminEsElMismo cubre SearchAdmin: repetía el
// mismo WHERE solo-FK y además necesita los JOIN del área, lo que obliga a
// cualificar `id` y `created_at` (psi_specialty_models comparte esas columnas con
// psi_users: sin calificar, 42702 "column reference is ambiguous").
func TestPsiRepo_FiltroDeAreaDelPanelAdminEsElMismo(t *testing.T) {
	mainDB := setupFullTestDB(t)
	ctx := context.Background()

	tx := mainDB.Begin()
	defer tx.Rollback()

	r := NewPsiRepository(tx)

	clinica := domain.PsiSpecialtyModel{Name: "Clínica", Active: true}
	require.NoError(t, tx.Create(&clinica).Error)

	dummyBio := domain.TextModel{ID: uuid.New(), Content: "Bio de prueba"}
	require.NoError(t, tx.Create(&dummyBio).Error)

	l := domain.PsiUserModel{
		FirstName: "Luis", LastName: "Prueba",
		CI: 222, FPV: 9002, Nationality: "V", BioTextID: dummyBio.ID,
		AudioBookShellId: "abs_luis",
		Solvent:          true, PrimaryWorkArea: "Clínica", // sin FK
		Credentials: domain.Credentials{Username: "uluis", Email: "luis@t.com", Password: "x", IsActive: true},
	}
	require.NoError(t, tx.Create(&l).Error)

	rows, total, err := r.SearchAdmin(ctx, request_structs.PsiDirectoryFilterDTO{
		SpecialtyID: clinica.ID, Page: 1, Limit: 50,
	})
	require.NoError(t, err, "SearchAdmin con filtro de área no debe fallar por columnas ambiguas")
	require.Equal(t, int64(1), total, "el panel debe encontrar al que solo tiene el legacy")
	require.Len(t, rows, 1)
	require.Equal(t, "Luis", rows[0].FirstName)
}
