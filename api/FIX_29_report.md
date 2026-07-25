# FIX-29: Deduplicar instanciación de repos en routers

## Metadatos

| Campo | Valor |
|-------|-------|
| **FIX** | FIX-29 |
| **Hallazgo** | MED-10 |
| **Severidad** | MEDIO |
| **Archivos** | `router.go`, `admin_router.go`, `psi_router.go`, `post_router.go`, `specialty_router.go` |
| **Estado** | COMPLETADO |
| **Fecha** | 2026-07-25 |

---

## Problema

Cada router file creaba sus propias instancias de repos:
- `AdminRepository`: instanciado 4 veces (1 por router)
- `PsiRepository`: instanciado 4 veces
- `PostRepository`: instanciado 2 veces
- `SpecialtyRepository`: instanciado 1 vez

---

## Solución

Repos se instancian una sola vez en `SetupRouter()` y se pasan como parámetros a los sub-routers usando interfaces del dominio:

| Router | Firma anterior | Firma nueva |
|--------|---------------|-------------|
| `SetupAdminRoutes` | `(router, db, analyticsSvc)` | `(router, adminRepo, psiRepo, analyticsSvc)` |
| `SetupPsiRoutes` | `(router, db, s3Client, analyticsSvc)` | `(router, psiRepo, adminRepo, s3Client, analyticsSvc)` |
| `SetupPostRoutes` | `(router, db, s3Client, analyticsSvc)` | `(router, adminRepo, psiRepo, postRepo, s3Client, analyticsSvc)` |
| `SetupSpecialtyRoutes` | `(router, db, analyticsSvc)` | `(router, psiRepo, adminRepo, specialtyRepo, analyticsSvc)` |

---

## Verificación

- `go build ./...` → sin errores ✅
- `go test ./internal/middleware/...` → PASS ✅
- Imports de `gorm.io/gorm` y `repository/postgres` eliminados de los 4 router files ✅
- Imports de `domain` agregados para usar interfaces ✅
