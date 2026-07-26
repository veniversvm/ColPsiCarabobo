package service

import (
	"testing"
	"time"
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
		name  string
		email string
		fpv   string
		nombre string
		want  string
	}{
		{
			name:  "email_with_at",
			email: "juan@test.com",
			fpv:   "12345",
			nombre: "Juan Perez",
			want:  "juan12345",
		},
		{
			name:  "email_without_at",
			email: "",
			fpv:   "99",
			nombre: "Maria Lopez",
			want:  "marialopez99",
		},
		{
			name:  "truncation_at_25_chars",
			email: "extremelylongusername@test.com",
			fpv:   "12345",
			nombre: "Test",
			want:  "extremelylongusernam12345",
		},
		{
			name:  "fpv_with_dots_commas_spaces",
			email: "a@b.com",
			fpv:   "1.234,5",
			nombre: "Test",
			want:  "a12345",
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
