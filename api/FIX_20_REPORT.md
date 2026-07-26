# FIX-20: Split PsiService God Object

**Fecha:** 25 de Julio, 2026
**Estado:** ✅ IMPLEMENTADO
**Commit:** pendiente

## Problema

`psi_service.go` contenia 1439 lineas con 18+ metodos cubriendo autenticacion, directorio, importacion, academicos, Audiobookshelf y perfiles. Dificultaba mantenimiento, testing y review de codigo.

## Solucion

Split en **7 archivos** manteniendo un solo `PsiService` struct (zero breaking changes en router/main). Cada archivo es una extension del mismo tipo en el mismo paquete `service`.

### Archivos creados/modificados

| Archivo | Lineas | Metodos | Responsabilidad |
|---------|--------|---------|-----------------|
| `psi_service.go` | 64 | `NewPsiService`, `ResolvePsiModelURLs`, `publicURL` | Core: struct, constructor, S3 URL resolution |
| `psi_service_auth.go` | 119 | `Login`, `LoginLibrary`, `Logout` | Autenticacion y sesiones JWT |
| `psi_service_audiobookshelf.go` | 113 | `actualizarEnAudiobookshelf`, `sincronizarConAudiobookshelf` | Integracion con Audiobookshelf |
| `psi_service_self_management.go` | 242 | `UpdateProfileSelf` | Autogestion de perfil con S3 rollback |
| `psi_service_directory.go` | 253 | `GetPublicDirectory`, `GetPublicProfile`, `GetPsiBioByID`, `GetPsiSOlvency`, `GetSitemapPsis` | Consulta publica y SEO |
| `psi_service_import.go` | 207 | `ImportFromCSV` + helpers (`parseInt`, `generateSecureUsername`, `parseBool`, `parseDate`, `cleanDash`) | Importacion masiva CSV |
| `psi_service_academic.go` | 129 | `AddPostGrade`, `UpdatePostGrade` | Gestion academica con S3 |

### Archivos existentes sin cambios

- `psi_service_xlsx.go` (289 lineas) — ya existia separado, usa helpers de import
- `psi_user_admin_service.go` (907 lineas) — ya existia separado

## GC Optimizations

- Pre-allocation de slices con capacidad conocida: `make([]string, 0, 4)` en UpdateProfileSelf (1 profile pic + 3 title images)
- Reutilizacion de audit model por batch en imports

## Tests nuevos

### `psi_service_import_test.go` (9 subtests + stress test)

Funciones puras testeadas con table-driven tests:
- `TestParseInt` — 9 casos (empty, dash, simple, commas, dots, spaces, mixed, zero, non-numeric)
- `TestGenerateSecureUsername` — 4 casos (email with/without @, truncation at 25 chars, FPV cleanup)
- `TestParseBool` — 11 casos (true, TRUE, 1, v, s, false, 0, empty, no, f)
- `TestParseDate` — 7 casos (empty, dash, zero, dd/mm/yyyy, yyyy-mm-dd, mm/dd/yy, invalid)
- `TestCleanDash` — 5 casos
- `TestParseDate_RangeLoop` — Stress test: 10,000 empty parses + valid date
- `TestParseDate_PastYear` — Edge case: 2010

### `psi_service_auth_test.go` (5 subtests)

- `TestPsiService_Logout/Logout limpia la Key`
- `TestPsiService_Login_CredencialesInvalidas` — 3 subtests: usuario no encontrado, cuenta inactiva, contrasena incorrecta
- `TestPsiService_LoginLibrary_CredencialesInvalidas` — 2 subtests: usuario no encontrado, cuenta inactiva

### `psi_service_directory_test.go` (9 subtests)

- `TestPsiService_GetPublicDirectory_Pagination` — 3 subtests: default pagination, gender normalization, invalid gender cleanup
- `TestPsiService_GetPublicDirectory_MiniProfile` — 2 subtests: specialties mapping, total_pages calculation
- `TestPsiService_GetPsiBioByID` — 2 subtests: existing bio, DB error
- `TestPsiService_GetSolvencies` — 1 subtest
- `TestPsiService_GetSitemapPsis` — 1 subtest

## Mock updates

`mockPsiRepoSvc` en `psi_service_test.go` expandido con:
- `SearchDirectoryFunc` + `SearchDirectory`
- `GetSitemapDataFunc` + `GetSitemapData`
- `GetSolvenciesFunc` + `GetSolvencies`
- `CreateWithColDataFunc` + `CreateWithColData`

## Resultado

```
go build ./...     → PASS
go vet ./...       → PASS
go test ./internal/service/... → PASS (2 pre-existing failures unchanged)
```

### Line count comparison

| Metric | Before | After (largest file) |
|--------|--------|---------------------|
| Largest file | 1439 lines (`psi_service.go`) | 253 lines (`psi_service_directory.go`) |
| Total service files | 3 (`psi_service.go`, `psi_service_xlsx.go`, `psi_user_admin_service.go`) | 9 |
| Test files | 1 (`psi_service_test.go`, 252 lines) | 4 (+ 830 total lines) |

## Pre-existing test failures (unrelated)

1. `TestAdminService_All/CreateAdmin` — expects hierarchy error, gets email validation error
2. `TestSpecialtyService_Update` — nil pointer dereference (mock missing `GetByAdminID`)
