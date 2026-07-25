# FIX-13 Report — Logger GORM en modo Info en producción

| Campo | Valor |
|-------|-------|
| **Fix ID** | FIX-13 |
| **Hallazgo original** | HIGH-06 |
| **Archivo modificado** | `pkg/database/postgres.go` |
| **Fecha de implementación** | 2026-07-25 |
| **Estado** | Completado |

---

## Problema

GORM logueaba TODAS las queries SQL en nivel `Info`. En producción esto genera I/O excesivo y puede exponer datos sensibles en logs.

```go
// ANTES:
Logger: logger.Default.LogMode(logger.Info),
```

---

## Corrección

Logger condicional: `Info` en desarrollo, `Warn` en producción.

```go
// DESPUÉS:
Logger: logger.Default.LogMode(func() logger.LogLevel {
    if config.Envs.Environment == "development" {
        return logger.Info
    }
    return logger.Warn
}()),
```

---

## Testing

- `go build ./...` — Pass
- `go vet ./...` — Pass
- Middleware tests: 15/15 Pass

---

## Impacto

- **Riesgo de regressión:** Mínimo. Solo afecta verbosidad del log.
- **Producción:** Solo warnings y errores (sin I/O excesivo ni filtración de PII).
- **Desarrollo:** Todas las queries visibles para debugging.

---

## Archivos relacionados

| Archivo | Relación |
|---------|----------|
| `pkg/database/postgres.go:45` | Archivo modificado |
