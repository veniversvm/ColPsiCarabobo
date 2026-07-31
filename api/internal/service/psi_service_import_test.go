package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

func TestParseInt(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "empty", input: "", want: 0},
		{name: "dash", input: "-", want: 0},
		{name: "simple", input: "12345", want: 12345},
		{name: "with_commas", input: "20,493", want: 20493},
		{name: "with_dots", input: "1.234", want: 1234},
		{name: "spaces", input: " 123 ", want: 123},
		{name: "mixed", input: "1,234.56", want: 123456},
		{name: "zero", input: "0", want: 0},
		{name: "non_numeric", input: "abc", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseInt(tt.input); got != tt.want {
				t.Errorf("parseInt(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateSecureUsername(t *testing.T) {
	tests := []struct {
		name   string
		email  string
		fpv    string
		nombre string
		want   string
	}{
		{
			name:   "email_with_at",
			email:  "juan@test.com",
			fpv:    "12345",
			nombre: "Juan Perez",
			want:   "juan12345",
		},
		{
			name:   "email_without_at",
			email:  "",
			fpv:    "99",
			nombre: "Maria Lopez",
			want:   "marialopez99",
		},
		{
			name:   "truncation_at_25_chars",
			email:  "extremelylongusername@test.com",
			fpv:    "12345",
			nombre: "Test",
			want:   "extremelylongusernam12345",
		},
		{
			name:   "fpv_with_dots_commas_spaces",
			email:  "a@b.com",
			fpv:    "1.234,5",
			nombre: "Test",
			want:   "a12345",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSecureUsername(tt.email, tt.fpv, tt.nombre)
			if got != tt.want {
				t.Errorf("generateSecureUsername(%q, %q, %q) = %q, want %q", tt.email, tt.fpv, tt.nombre, got, tt.want)
			}
			if len(got) > 25 {
				t.Errorf("username length %d exceeds 25 chars", len(got))
			}
		})
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"true", true}, {"TRUE", true}, {"1", true},
		{"v", true}, {"s", true}, {" true ", true},
		{"false", false}, {"0", false}, {"", false},
		{"no", false}, {"f", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := parseBool(tt.input); got != tt.want {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNil bool
		year    int
	}{
		{name: "empty", input: "", wantNil: true},
		{name: "dash", input: "-", wantNil: true},
		{name: "zero", input: "0", wantNil: true},
		{name: "dd_mm_yyyy", input: "15/03/2020", year: 2020},
		{name: "yyyy_mm_dd", input: "2020-03-15", year: 2020},
		{name: "mm_dd_yy", input: "3/15/20", year: 2020},
		{name: "invalid", input: "not-a-date", wantNil: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDate(tt.input)
			if tt.wantNil {
				if !got.IsZero() {
					t.Errorf("parseDate(%q) expected zero time, got %v", tt.input, got)
				}
			} else {
				if got.IsZero() {
					t.Errorf("parseDate(%q) returned zero time unexpectedly", tt.input)
				} else if got.Year() != tt.year {
					t.Errorf("parseDate(%q) year = %d, want %d", tt.input, got.Year(), tt.year)
				}
			}
		})
	}
}

func TestCleanDash(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"-", ""}, {"0", ""}, {"hello", "hello"}, {"", ""}, {"-", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := cleanDash(tt.input); got != tt.want {
				t.Errorf("cleanDash(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDate_RangeLoop(t *testing.T) {
	// Stress test: parseDate should not allocate on empty inputs (GC-friendly)
	for i := 0; i < 10000; i++ {
		_ = parseDate("")
		_ = parseDate("-")
		_ = parseDate("0")
	}
	// Valid date still works after stress
	d := parseDate("25/12/2025")
	if d.IsZero() || d.Year() != 2025 {
		t.Errorf("parseDate stress test failed: got %v", d)
	}
}

func TestParseDate_PastYear(t *testing.T) {
	got := parseDate("01/01/2010")
	if got.IsZero() || got != time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC) {
		t.Errorf("parseDate(01/01/2010) = %v, want 2010-01-01", got)
	}
}

func TestParseDate_ExcelSerial(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{name: "serial 45852", input: "45852", want: time.Date(2025, 7, 14, 0, 0, 0, 0, time.UTC)},
		{name: "serial 33154", input: "33154", want: time.Date(1990, 10, 8, 0, 0, 0, 0, time.UTC)},
		{name: "serial 46230", input: "46230", want: time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC)},
		{name: "serial con decimales", input: "45852.0", want: time.Date(2025, 7, 14, 0, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDate(tt.input)
			if got.IsZero() || !got.Equal(tt.want) {
				t.Errorf("parseDate(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseNationality(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"V", "V"}, {"v", "V"}, {"E", "E"}, {"e", "E"},
		{"", "V"}, {"VE", "V"}, {"123", "V"},
	}
	for _, tt := range tests {
		if got := parseNationality(tt.input); got != tt.want {
			t.Errorf("parseNationality(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsProofOfLife(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"SI", true}, {"si", true}, {"Sí", true}, {"s", true}, {"1", true}, {"true", true},
		{"NO", false}, {"no", false}, {"Fallecido", false}, {"fallecido", false}, {"", false},
	}
	for _, tt := range tests {
		if got := isProofOfLife(tt.input); got != tt.want {
			t.Errorf("isProofOfLife(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsSolventByLastSolvency(t *testing.T) {
	currentYear := time.Now().Year()
	tests := []struct {
		name  string
		input time.Time
		want  bool
	}{
		{"sin fecha", time.Time{}, false},
		{"año vigente", time.Date(currentYear, 6, 1, 0, 0, 0, 0, time.UTC), true},
		{"año anterior", time.Date(currentYear-1, 12, 31, 0, 0, 0, 0, time.UTC), false},
		{"año futuro", time.Date(currentYear+1, 1, 1, 0, 0, 0, 0, time.UTC), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSolventByLastSolvency(tt.input); got != tt.want {
				t.Errorf("isSolventByLastSolvency(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildSolvencyHistory(t *testing.T) {
	audit := domain.AuditModel{CreateBy: "test"}
	userID := uuid.Must(uuid.NewV7())
	currentYear := time.Now().Year()

	t.Run("sin fecha retorna nil", func(t *testing.T) {
		if got := buildSolvencyHistory(time.Time{}, 0, userID, audit); got != nil {
			t.Errorf("esperaba nil, obtuve %v", got)
		}
	})

	t.Run("año futuro retorna nil", func(t *testing.T) {
		if got := buildSolvencyHistory(time.Date(currentYear+1, 1, 1, 0, 0, 0, 0, time.UTC), 0, userID, audit); got != nil {
			t.Errorf("esperaba nil, obtuve %v", got)
		}
	})

	t.Run("año previo a 2024 retorna nil", func(t *testing.T) {
		if got := buildSolvencyHistory(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), 0, userID, audit); got != nil {
			t.Errorf("esperaba nil, obtuve %v", got)
		}
	})

	t.Run("desde colegiatura 2025 hasta año vigente", func(t *testing.T) {
		got := buildSolvencyHistory(time.Date(currentYear, 7, 14, 0, 0, 0, 0, time.UTC), 2025, userID, audit)
		wantLen := currentYear - 2025 + 1
		if len(got) != wantLen {
			t.Fatalf("len = %d, want %d", len(got), wantLen)
		}
		for i, s := range got {
			wantYear := 2025 + i
			if s.Date.Year() != wantYear || s.Date.Month() != time.December || s.Date.Day() != 31 {
				t.Errorf("solvencia[%d] = %v, want 31/12/%d", i, s.Date, wantYear)
			}
			if s.PsiUserModelID != userID {
				t.Errorf("solvencia[%d] no está vinculada al usuario", i)
			}
		}
	})

	t.Run("colegiatura vacía usa límite 2024", func(t *testing.T) {
		got := buildSolvencyHistory(time.Date(currentYear, 7, 14, 0, 0, 0, 0, time.UTC), 0, userID, audit)
		if len(got) == 0 || got[0].Date.Year() != 2024 {
			t.Errorf("primera solvencia = %v, want 2024", got[0].Date)
		}
	})
}

func TestImportFromCSV_FullMapping(t *testing.T) {
	config.InitConfig()
	rows := [][]string{emptyRow(), emptyRow()}
	data := emptyRow()
	data[0] = "42"
	data[3] = "1042"
	data[4] = "33154"
	data[5] = "V"
	data[6] = "12345678"
	data[7] = "Ana"
	data[8] = "María"
	data[9] = "Pérez"
	data[10] = "Gómez"
	data[11] = "33000"
	data[13] = "F"
	data[14] = "SI"
	data[15] = "ana@test.com"
	data[16] = "valencia"
	data[17] = "0241-5555"
	data[18] = "0414-5555"
	data[19] = "lara"
	data[25] = "UC"
	data[26] = "33500"
	data[27] = "Psicología"
	data[28] = "Carabobo"
	data[29] = "34000"
	data[30] = "123"
	data[31] = "5"
	data[32] = "2"
	data[33] = "Dipl Clínica"
	data[34] = "Esp Neuro"
	data[38] = "SI"
	data[39] = "SI"
	data[43] = "SI"
	data[44] = "46230"
	data[46] = "Zulia"
	data[47] = "Aprobado"
	rows = append(rows, data, emptyRow())

	var capturedPsi *domain.PsiUserModel
	var capturedCol *domain.PsiUserColData
	var capturedSol []domain.PsiUserSolvency
	var capturedPG []domain.PsiUserPostGrade

	repo := &mockPsiRepoSvc{
		CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
			capturedPsi = psi
			capturedCol = col
			capturedSol = sol
			capturedPG = pg
			return nil
		},
	}
	svc := NewPsiService(repo, nil, nil)

	buf := createTestXLSX(t, rows)
	success, failed := svc.ImportFromCSV(context.Background(), buf, uuid.Must(uuid.NewV7()))

	require := func(cond bool, msg string) {
		if !cond {
			t.Error(msg)
		}
	}

	require(success == 1, "debe importar 1 fila")
	require(len(failed) == 0, "no debe haber filas fallidas")

	require(capturedPsi.ControlNumber == "42", "ControlNumber no mapeado")
	require(capturedPsi.Nationality == "V", "Nationality no mapeada")
	require(capturedPsi.SecondName == "María", "SecondName no mapeado")
	require(capturedPsi.SecondLastName == "Gómez", "SecondLastName no mapeado")
	require(capturedPsi.IsActive, "IsActive debe ser true")
	require(capturedPsi.Solvent, "Solvent debe ser true (solvencia 2026)")
	require(capturedPsi.ProofOfLife, "ProofOfLife debe ser true (SI)")
	require(capturedPsi.MunicipalityCarabobo == "Valencia", "Municipio no normalizado")
	require(capturedPsi.StateOutside == "Lara", "Estado no normalizado")
	require(capturedPsi.PrimaryWorkArea == "", "columna AL no usada")
	require(capturedPsi.BornDate.Equal(time.Date(1990, 5, 7, 0, 0, 0, 0, time.UTC)), "BornDate serial mal convertida")
	require(capturedCol.GuildInscriptionDate.Equal(time.Date(1990, 10, 8, 0, 0, 0, 0, time.UTC)), "InscriptionDate serial mal convertida")
	require(capturedCol.GraduateDate.Equal(time.Date(1991, 9, 19, 0, 0, 0, 0, time.UTC)), "GraduateDate serial mal convertida")
	require(capturedCol.RegisterTitleDate.Equal(time.Date(1993, 1, 31, 0, 0, 0, 0, time.UTC)), "RegisterTitleDate serial mal convertida")
	require(capturedCol.DateOfLastSolvency.Equal(time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC)), "DateOfLastSolvency serial mal convertida")
	require(capturedCol.GuildDirector && capturedCol.SixtyFiveOrPlus && capturedCol.UniversityProfessor, "Flags gremiales no mapeados")
	require(capturedCol.DoubleGuild && capturedCol.DoubleGuildLocation == "Zulia", "Doble colegiatura no mapeada")
	require(capturedCol.CPSM, "CPSM no mapeado")
	require(len(capturedPG) == 2, "postgrados no mapeados")
	require(capturedPG[0].Type == domain.Diplomado && capturedPG[1].Type == domain.Especializacion, "tipos de postgrado incorrectos")
	require(len(capturedSol) == 3, "historial de solvencias incorrecto (2024-2026)")
	for i, s := range capturedSol {
		require(s.Date.Year() == 2024+i && s.Date.Month() == time.December, "fecha de solvencia incorrecta")
	}
}
