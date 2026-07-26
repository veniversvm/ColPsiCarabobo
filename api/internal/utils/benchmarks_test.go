package utils

import (
	"strings"
	"testing"
)

func BenchmarkIsStrongPassword(b *testing.B) {
	benchmarks := []struct {
		name     string
		password string
	}{
		{"weak", "pass"},
		{"medium", "Password1!"},
		{"strong", "V3ry$tr0ng&C0mpl3xP@ssw0rd!"},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				IsStrongPassword(bm.password)
			}
		})
	}
}

func BenchmarkGenerateSecureRandomString(b *testing.B) {
	benchmarks := []struct {
		name   string
		length int
	}{
		{"16chars", 16},
		{"32chars", 32},
		{"64chars", 64},
		{"128chars", 128},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				GenerateSecureRandomString(bm.length)
			}
		})
	}
}

func BenchmarkCleanAlphaNumeric(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{"short", "Hello@World!"},
		{"medium", "user@email.com (Work) - 123"},
		{"long", "This is a very long string with lots of special characters !@#$%^&*() and spaces mixed in"},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				CleanAlphaNumeric(bm.input)
			}
		})
	}
}

func BenchmarkNormalizeMunicipioCarabobo(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{"exact", "Valencia"},
		{"case_insensitive", "valencia"},
		{"tilde_tolerant", "San Joaquin"},
		{"with_spaces", "  San Diego  "},
		{"not_found", "FalsoMunicipio"},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				NormalizeMunicipioCarabobo(bm.input)
			}
		})
	}
}

func BenchmarkNormalizeEstadoVenezuela(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{"exact", "Lara"},
		{"case_insensitive", "lara"},
		{"tilde", "Anzoategui"},
		{"not_found", "Valencia"},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				NormalizeEstadoVenezuela(bm.input)
			}
		})
	}
}

func BenchmarkSanitizeImage(b *testing.B) {
	fakeImage := strings.NewReader("Esto no es un JPEG, es texto para engañar al sistema")
	for i := 0; i < b.N; i++ {
		SanitizeImage(fakeImage)
		fakeImage.Reset("Esto no es un JPEG, es texto para engañar al sistema")
	}
}

func BenchmarkParseAndValidateEmail(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{"valid", "user@test.com"},
		{"uppercase", "USER@TEST.COM"},
		{"with_name", "Fran <fran@colpsi.com>"},
		{"invalid", "not-an-email"},
		{"empty", ""},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ParseAndValidateEmail(bm.input)
			}
		})
	}
}

func BenchmarkBoolFromForm(b *testing.B) {
	benchmarks := []struct {
		name  string
		input string
	}{
		{"true", "true"},
		{"false", "false"},
		{"one", "1"},
		{"zero", "0"},
		{"empty", ""},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				BoolFromForm(bm.input)
			}
		})
	}
}

func BenchmarkIsEmptyReq(b *testing.B) {
	s := DummyStruct{Name: "test", Age: 25}
	for i := 0; i < b.N; i++ {
		IsEmptyReq(s)
	}
}

func BenchmarkNormalizePlatformName(b *testing.B) {
	inputs := []string{"ig", "facebook", "https://youtu.be/mivideo", "threads"}
	for _, in := range inputs {
		b.Run(in, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				NormalizePlatformName(in)
			}
		})
	}
}
