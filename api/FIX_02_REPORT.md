# FIX-02 Report — Password hardcodeada en CSV import

| Campo | Valor |
|-------|-------|
| **Fix ID** | FIX-02 |
| **Hallazgo original** | CRIT-02 |
| **Archivo modificado** | `internal/service/psi_service.go` |
| **Línea afectada** | 125 |
| **Fecha de implementación** | 2026-07-25 |
| **Estado** | Completado |

---

## Problema

La función `ImportFromCSV` en `psi_service.go:125` asignaba la misma password hardcodeada `"Colpsi2025!"` a TODOS los usuarios importados vía CSV:

```go
defaultPassword := "Colpsi2025!"
```

Si un atacante conocía el patrón o accedía al código fuente, podía loguearse como cualquier usuario importado.

Nota: La importación XLSX (`psi_service_xlsx.go:68`) **ya usaba** el patrón seguro con `utils.GenerateSecureRandomString(12)`.

---

## Corrección

```go
// ANTES:
defaultPassword := "Colpsi2025!"

// DESPUÉS:
var defaultPassword string
if config.Envs.Environment == "development" {
    defaultPassword = "Colpsi2025!"
} else {
    defaultPassword = utils.GenerateSecureRandomString(16)
}
```

| Entorno | Comportamiento |
|---------|---------------|
| `APP_ENV=development` | Password conocida `"Colpsi2025!"` (fácil para dev) |
| `APP_ENV=production` (default) | Password aleatoria de 16 caracteres por usuario |

---

## Testing

| Verificación | Estado |
|---|---|
| `go vet ./internal/service/...` | Pass |
| `go test ./internal/middleware/...` | Pass (15/15) |
| `go test ./internal/utils/...` | 4/5 pass, 1 fallo pre-existente (FIX-33) |

---

## Archivos relacionados

| Archivo | Relación |
|---------|----------|
| `internal/service/psi_service.go:125` | Línea modificada |
| `internal/service/psi_service_xlsx.go:68` | Patrón seguro ya existente (referencia) |
| `SECURITY_FIX_PLAN.md` (FIX-02) | Plan de seguridad referenciado |
