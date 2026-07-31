package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/xuri/excelize/v2"
)

// =========================================================================
// TEST: getValorSeguro (utility function)
// =========================================================================

func TestGetValorSeguro(t *testing.T) {
	tests := []struct {
		name   string
		row    []string
		index  int
		expect string
	}{
		{"index in bounds", []string{"a", "b", "c"}, 1, "b"},
		{"index out of bounds", []string{"a", "b"}, 5, ""},
		{"index zero", []string{"first", "second"}, 0, "first"},
		{"last element", []string{"a", "b", "c"}, 2, "c"},
		{"empty string in cell", []string{"a", "", "c"}, 1, ""},
		{"whitespace trimmed", []string{"  hello  "}, 0, "hello"},
		{"nil row", nil, 0, ""},
		{"empty row", []string{}, 0, ""},
		{"exact boundary", []string{"a", "b"}, 2, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := getValorSeguro(tc.row, tc.index)
			require.Equal(t, tc.expect, result)
		})
	}
}

// =========================================================================
// HELPERS: Create test XLSX files
// =========================================================================

func createTestXLSX(t *testing.T, rows [][]string) *bytes.Buffer {
	t.Helper()

	f := excelize.NewFile()
	sheetName := "BD ColPsiCarabobo 2026"

	index, err := f.GetSheetIndex("Sheet1")
	require.NoError(t, err)
	_ = index
	f.SetSheetName("Sheet1", sheetName)

	for i, row := range rows {
		for j, val := range row {
			cell, err := excelize.CoordinatesToCellName(j+1, i+1)
			require.NoError(t, err)
			require.NoError(t, f.SetCellValue(sheetName, cell, val))
		}
	}

	var buf bytes.Buffer
	require.NoError(t, f.Write(&buf))
	require.NoError(t, f.Close())

	return &buf
}

// emptyRow genera una fila de 50 strings vacios
func emptyRow() []string {
	return make([]string, 50)
}

// dataRow construye una fila de datos con campos clave
type testRow struct {
	FPV          string
	CI           string
	FirstName    string
	SecondName   string
	LastName     string
	SecLastName  string
	Email        string
	Genre        string
	Municipio    string
	Estado       string
	Postgrade1   string
	Postgrade2   string
	Postgrade3   string
	Postgrade4   string
	Status       string
	ProofOfLife  string
	SolvencyDate string
}

func buildDataRow(d testRow) []string {
	row := emptyRow()
	if d.FPV != "" {
		row[3] = d.FPV
	}
	if d.CI != "" {
		row[6] = d.CI
	}
	if d.FirstName != "" {
		row[7] = d.FirstName
	}
	if d.SecondName != "" {
		row[8] = d.SecondName
	}
	if d.LastName != "" {
		row[9] = d.LastName
	}
	if d.SecLastName != "" {
		row[10] = d.SecLastName
	}
	if d.Email != "" {
		row[15] = d.Email
	}
	if d.Genre != "" {
		row[13] = d.Genre
	}
	if d.Municipio != "" {
		row[16] = d.Municipio
	}
	if d.Estado != "" {
		row[19] = d.Estado
	}
	if d.Postgrade1 != "" {
		row[33] = d.Postgrade1
	}
	if d.Postgrade2 != "" {
		row[34] = d.Postgrade2
	}
	if d.Postgrade3 != "" {
		row[35] = d.Postgrade3
	}
	if d.Postgrade4 != "" {
		row[36] = d.Postgrade4
	}
	if d.Status != "" {
		row[45] = d.Status
	}
	if d.ProofOfLife != "" {
		row[14] = d.ProofOfLife
	}
	if d.SolvencyDate != "" {
		row[44] = d.SolvencyDate
	}
	return row
}

func testRows(dataRows ...testRow) [][]string {
	rows := [][]string{emptyRow(), emptyRow()} // titulo + headers
	for _, d := range dataRows {
		rows = append(rows, buildDataRow(d))
	}
	return rows
}

// =========================================================================
// TEST: ImportFromXLSX — Apertura
// =========================================================================

func TestImportFromXLSX_Apertura(t *testing.T) {
	t.Run("archivo corrupto retorna error", func(t *testing.T) {
		repo := &mockPsiRepoSvc{}
		svc := NewPsiService(repo, nil, nil)

		reader := bytes.NewBufferString("esto no es un xlsx")
		success, failed := svc.ImportFromXLSX(context.Background(), reader, uuid.Must(uuid.NewV7()))

		require.Equal(t, 0, success)
		require.Len(t, failed, 1)
		require.Contains(t, failed[0]["error"], "no se pudo abrir el archivo Excel")
	})

	t.Run("hoja inexistente retorna error", func(t *testing.T) {
		f := excelize.NewFile()
		var buf bytes.Buffer
		require.NoError(t, f.Write(&buf))
		require.NoError(t, f.Close())

		repo := &mockPsiRepoSvc{}
		svc := NewPsiService(repo, nil, nil)

		success, failed := svc.ImportFromXLSX(context.Background(), &buf, uuid.Must(uuid.NewV7()))
		require.Equal(t, 0, success)
		require.Len(t, failed, 1)
		require.Contains(t, failed[0]["error"], "no se pudo leer la hoja especificada")
	})
}

// =========================================================================
// TEST: ImportFromXLSX — Ghost Rows
// =========================================================================

func TestImportFromXLSX_GhostRows(t *testing.T) {
	t.Run("filas vacias son ignoradas", func(t *testing.T) {
		rows := [][]string{emptyRow(), emptyRow(), emptyRow(), emptyRow()}

		createCalled := false
		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				createCalled = true
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		success, failed := svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))

		require.Equal(t, 0, success)
		require.Empty(t, failed)
		require.False(t, createCalled)
	})
}

// =========================================================================
// TEST: ImportFromXLSX — Municipality Validation
// =========================================================================

func TestImportFromXLSX_MunicipalityValidation(t *testing.T) {
	t.Run("municipio invalido registra fallo", func(t *testing.T) {
		rows := testRows(testRow{FPV: "100", CI: "200", FirstName: "Test", LastName: "User", Genre: "M", Municipio: "FalsoMunicipio"})

		createCalled := false
		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				createCalled = true
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		success, failed := svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))

		require.Equal(t, 0, success)
		require.Len(t, failed, 1)
		require.Contains(t, failed[0]["error"], "municipio de Carabobo inválido")
		require.False(t, createCalled)
	})

	t.Run("municipio valido permite insertar", func(t *testing.T) {
		rows := testRows(testRow{FPV: "100", CI: "200", FirstName: "Test", LastName: "User", Genre: "M", Municipio: "Valencia"})

		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		success, failed := svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))

		require.Equal(t, 1, success)
		require.Empty(t, failed)
	})
}

// =========================================================================
// TEST: ImportFromXLSX — DB Error Handling
// =========================================================================

func TestImportFromXLSX_DBError(t *testing.T) {
	t.Run("error de DB en primera fila no impide segunda", func(t *testing.T) {
		rows := testRows(
			testRow{FPV: "100", CI: "200", FirstName: "A", LastName: "B", Genre: "M", Municipio: "Valencia"},
			testRow{FPV: "101", CI: "201", FirstName: "C", LastName: "D", Genre: "F", Municipio: "Valencia"},
		)

		callCount := 0
		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				callCount++
				if callCount == 1 {
					return errors.New("duplicate key value violates unique constraint \"idx_psi_users_fpv\"")
				}
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		success, failed := svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))

		require.Equal(t, 1, success)
		require.Len(t, failed, 1)
		require.Contains(t, failed[0]["error"], "FPV")
	})

	t.Run("todas las filas fallan", func(t *testing.T) {
		rows := testRows(testRow{FPV: "100", CI: "200", FirstName: "A", LastName: "B", Genre: "M", Municipio: "Valencia"})

		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				return errors.New("constraint violation")
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		success, failed := svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))

		require.Equal(t, 0, success)
		require.Len(t, failed, 1)
	})
}

// =========================================================================
// TEST: ImportFromXLSX — Welcome Email
// =========================================================================

func TestImportFromXLSX_WelcomeEmail(t *testing.T) {
	t.Run("email valido envia welcome", func(t *testing.T) {
		var emailSent bool
		var sentTo string

		rows := testRows(testRow{FPV: "100", CI: "200", FirstName: "Test", LastName: "User", Genre: "M", Municipio: "Valencia", Email: "test@email.com"})

		mailSvc := &mockMailSvc{
			SendEmailFunc: func(to, subject, template string, data interface{}) error {
				emailSent = true
				sentTo = to
				return nil
			},
		}
		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				return nil
			},
		}
		svc := NewPsiService(repo, nil, mailSvc)

		buf := createTestXLSX(t, rows)
		svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))

		require.True(t, emailSent)
		require.Equal(t, "test@email.com", sentTo)
	})

	t.Run("email vacio no envia welcome", func(t *testing.T) {
		emailSent := false

		rows := testRows(testRow{FPV: "100", CI: "200", FirstName: "Test", LastName: "User", Genre: "M", Municipio: "Valencia"})

		mailSvc := &mockMailSvc{
			SendEmailFunc: func(to, subject, template string, data interface{}) error {
				emailSent = true
				return nil
			},
		}
		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				return nil
			},
		}
		svc := NewPsiService(repo, nil, mailSvc)

		buf := createTestXLSX(t, rows)
		svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))

		require.False(t, emailSent)
	})

	t.Run("error de email no es fatal", func(t *testing.T) {
		rows := testRows(testRow{FPV: "100", CI: "200", FirstName: "Test", LastName: "User", Genre: "M", Municipio: "Valencia", Email: "test@email.com"})

		mailSvc := &mockMailSvc{
			SendEmailFunc: func(to, subject, template string, data interface{}) error {
				return errors.New("smtp connection refused")
			},
		}
		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				return nil
			},
		}
		svc := NewPsiService(repo, nil, mailSvc)

		buf := createTestXLSX(t, rows)
		success, failed := svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))

		require.Equal(t, 1, success)
		require.Empty(t, failed)
	})
}

// =========================================================================
// TEST: ImportFromXLSX — ProofOfLife
// =========================================================================

func TestImportFromXLSX_ProofOfLife(t *testing.T) {
	t.Run("fallecido设置ProofOfLife a false", func(t *testing.T) {
		row := emptyRow()
		row[3] = "100"
		row[6] = "200"
		row[7] = "Test"
		row[9] = "User"
		row[13] = "M"
		row[14] = "fallecido"
		row[16] = "Valencia"

		rows := [][]string{emptyRow(), emptyRow(), row}

		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				require.False(t, psi.ProofOfLife, "ProofOfLife debe ser false cuando es fallecido")
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		success, _ := svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))
		require.Equal(t, 1, success)
	})

	t.Run("valor normal设置ProofOfLife a true", func(t *testing.T) {
		rows := testRows(testRow{FPV: "100", CI: "200", FirstName: "Test", LastName: "User", Genre: "M", Municipio: "Valencia", ProofOfLife: "SI"})

		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				require.True(t, psi.ProofOfLife, "ProofOfLife debe ser true para registros normales")
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		success, _ := svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))
		require.Equal(t, 1, success)
	})
}

// =========================================================================
// TEST: ImportFromXLSX — PostGrades
// =========================================================================

func TestImportFromXLSX_PostGrades(t *testing.T) {
	t.Run("4 postgrados son capturados", func(t *testing.T) {
		rows := testRows(testRow{
			FPV: "100", CI: "200", FirstName: "Test", LastName: "User", Genre: "M",
			Municipio: "Valencia", Postgrade1: "Dipl Psic Clinica", Postgrade2: "Esp Neuropsicologia",
			Postgrade3: "Maestria Educacion", Postgrade4: "Doctorado Psicologia",
		})

		var capturedPGs []domain.PsiUserPostGrade
		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				capturedPGs = pg
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		success, failed := svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))

		require.Equal(t, 1, success)
		require.Empty(t, failed)
		require.Len(t, capturedPGs, 4)
		require.Equal(t, domain.Diplomado, capturedPGs[0].Type)
		require.Equal(t, domain.Especializacion, capturedPGs[1].Type)
		require.Equal(t, domain.Maestria, capturedPGs[2].Type)
		require.Equal(t, domain.Doctorado, capturedPGs[3].Type)
	})

	t.Run("postgrados vacios no se agregan", func(t *testing.T) {
		rows := testRows(testRow{FPV: "100", CI: "200", FirstName: "Test", LastName: "User", Genre: "M", Municipio: "Valencia"})

		var capturedPGs []domain.PsiUserPostGrade
		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				capturedPGs = pg
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))

		require.Len(t, capturedPGs, 0)
	})

	t.Run("solo diplomado presente", func(t *testing.T) {
		rows := testRows(testRow{
			FPV: "100", CI: "200", FirstName: "Test", LastName: "User", Genre: "M",
			Municipio: "Valencia", Postgrade1: "Solo Diplomado",
		})

		var capturedPGs []domain.PsiUserPostGrade
		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				capturedPGs = pg
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))

		require.Len(t, capturedPGs, 1)
		require.Equal(t, domain.Diplomado, capturedPGs[0].Type)
		require.Equal(t, "Solo Diplomado", capturedPGs[0].Title)
	})
}

// =========================================================================
// TEST: ImportFromXLSX — Fault Tolerance Multi-Row
// =========================================================================

func TestImportFromXLSX_FaultTolerance(t *testing.T) {
	t.Run("falla en fila 1 no impide fila 2", func(t *testing.T) {
		rows := testRows(
			testRow{FPV: "100", CI: "200", FirstName: "A", LastName: "B", Genre: "M", Municipio: "FalsoMunicipio"},
			testRow{FPV: "101", CI: "201", FirstName: "C", LastName: "D", Genre: "F", Municipio: "Valencia"},
		)

		createCount := 0
		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				createCount++
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		success, failed := svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))

		require.Equal(t, 1, success)
		require.Len(t, failed, 1)
		require.Equal(t, 1, createCount)
	})

	t.Run("error DB en fila 1 no impide fila 2", func(t *testing.T) {
		rows := testRows(
			testRow{FPV: "100", CI: "200", FirstName: "A", LastName: "B", Genre: "M", Municipio: "Valencia"},
			testRow{FPV: "101", CI: "201", FirstName: "C", LastName: "D", Genre: "F", Municipio: "Valencia"},
		)

		callCount := 0
		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				callCount++
				if callCount == 1 {
					return errors.New("duplicate key")
				}
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		success, failed := svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))

		require.Equal(t, 1, success)
		require.Len(t, failed, 1)
	})
}

// =========================================================================
// TEST: ImportFromXLSX — Credentials
// =========================================================================

func TestImportFromXLSX_Credentials(t *testing.T) {
	t.Run("must_change_password es true", func(t *testing.T) {
		rows := testRows(testRow{FPV: "100", CI: "200", FirstName: "Test", LastName: "User", Genre: "M", Municipio: "Valencia"})

		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				require.True(t, psi.MustChangePassword, "MustChangePassword debe ser true en importacion")
				require.NotEmpty(t, psi.Password, "Password no debe estar vacio")
				require.NotEmpty(t, psi.Key, "Key no debe estar vacio")
				require.Empty(t, psi.AudioBookShellId, "AudioBookShellId debe estar vacio al importar")
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		success, _ := svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))
		require.Equal(t, 1, success)
	})

	t.Run("privacidad default es false", func(t *testing.T) {
		rows := testRows(testRow{FPV: "100", CI: "200", FirstName: "Test", LastName: "User", Genre: "M", Municipio: "Valencia"})

		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				require.False(t, psi.ShowContactEmail)
				require.False(t, psi.ShowMunicipalityCarabobo)
				require.False(t, psi.ShowPhoneCarabobo)
				require.False(t, psi.ShowCelPhoneCarabobo)
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))
	})
}

// =========================================================================
// TEST: ImportFromXLSX — MustChangePassword (from import code)
// =========================================================================

func TestImportFromXLSX_GeneratedUsername(t *testing.T) {
	t.Run("username con email", func(t *testing.T) {
		rows := testRows(testRow{FPV: "999", CI: "111", FirstName: "Juan", LastName: "Perez", Genre: "M", Municipio: "Valencia", Email: "juan@test.com"})

		var capturedUser string
		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				capturedUser = psi.Username
				return nil
			},
		}
		mail := &mockMailSvc{SendEmailFunc: func(to, subject, template string, data interface{}) error { return nil }}
		svc := NewPsiService(repo, nil, mail)

		buf := createTestXLSX(t, rows)
		svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))

		require.Equal(t, "juan999", capturedUser, "Username = localPart + FPV")
	})

	t.Run("username sin email", func(t *testing.T) {
		rows := testRows(testRow{FPV: "888", CI: "222", FirstName: "Maria", LastName: "Lopez", Genre: "F", Municipio: "Valencia"})

		var capturedUser string
		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				capturedUser = psi.Username
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))

		require.Equal(t, "psi888222", capturedUser, "Username = psi + FPV + CI")
	})
}

// =========================================================================
// TEST: ImportFromXLSX — Estado Outside
// =========================================================================

func TestImportFromXLSX_EstadoOutside(t *testing.T) {
	t.Run("estado valido se normaliza", func(t *testing.T) {
		rows := testRows(testRow{FPV: "100", CI: "200", FirstName: "T", LastName: "U", Genre: "M", Estado: "lara"})

		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				require.Equal(t, "Lara", psi.StateOutside, "Estado debe normalizarse")
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		success, _ := svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))
		require.Equal(t, 1, success)
	})

	t.Run("estado invalido se mantiene raw", func(t *testing.T) {
		rows := testRows(testRow{FPV: "100", CI: "200", FirstName: "T", LastName: "U", Genre: "M", Estado: "EstadoFalso"})

		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				require.Equal(t, "EstadoFalso", psi.StateOutside, "Estado invalido se mantiene raw")
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		success, _ := svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))
		require.Equal(t, 1, success)
	})
}

// =========================================================================
// TEST: ImportFromXLSX — IsActive y Solvent (nuevas reglas de lanzamiento)
// =========================================================================
//
// Reglas acordadas:
//   - Todos los registros importados quedan activos (is_active = true).
//   - Solvente únicamente si la última solvencia registrada es del año vigente.

func TestImportFromXLSX_IsActive(t *testing.T) {
	t.Run("todos los importados quedan activos", func(t *testing.T) {
		rows := testRows(testRow{FPV: "100", CI: "200", FirstName: "T", LastName: "U", Genre: "M", Municipio: "Valencia"})

		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				require.True(t, psi.IsActive, "IsActive debe ser true para todos los importados")
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))
	})

	t.Run("solvente con solvencia del año vigente", func(t *testing.T) {
		rows := testRows(testRow{
			FPV: "100", CI: "200", FirstName: "T", LastName: "U", Genre: "M", Municipio: "Valencia",
			SolvencyDate: fmt.Sprintf("%d-07-14", time.Now().Year()),
		})

		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				require.True(t, psi.Solvent, "Solvent debe ser true con solvencia del año vigente")
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))
	})

	t.Run("insolvente sin solvencia del año vigente", func(t *testing.T) {
		rows := testRows(testRow{
			FPV: "100", CI: "200", FirstName: "T", LastName: "U", Genre: "M", Municipio: "Valencia",
			SolvencyDate: fmt.Sprintf("%d-07-14", time.Now().Year()-1),
		})

		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				require.False(t, psi.Solvent, "Solvent debe ser false sin solvencia del año vigente")
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))
	})

	t.Run("historial de solvencias desde la colegiatura", func(t *testing.T) {
		inscriptionYear := 2024
		rows := testRows(testRow{
			FPV: "100", CI: "200", FirstName: "T", LastName: "U", Genre: "M", Municipio: "Valencia",
			SolvencyDate: fmt.Sprintf("%d-07-14", time.Now().Year()),
		})

		var captured []domain.PsiUserSolvency
		repo := &mockPsiRepoSvc{
			CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
				captured = sol
				return nil
			},
		}
		svc := NewPsiService(repo, nil, nil)

		buf := createTestXLSX(t, rows)
		svc.ImportFromXLSX(context.Background(), buf, uuid.Must(uuid.NewV7()))

		require.Len(t, captured, time.Now().Year()-inscriptionYear+1, "debe sembrar solvencias anuales desde 2024")
		for i, s := range captured {
			require.Equal(t, inscriptionYear+i, s.Date.Year())
			require.Equal(t, time.December, s.Date.Month())
		}
	})
}

// verify unused import
var _ = strings.Contains
