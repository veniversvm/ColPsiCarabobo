# FIX-09 Report — Debug-monitor sin autenticación

| Campo | Valor |
|-------|-------|
| **Fix ID** | FIX-09 |
| **Hallazgo original** | HIGH-01 |
| **Archivo modificado** | `internal/router/admin_router.go` |
| **Fecha de implementación** | 2026-07-25 |
| **Estado** | Completado |

---

## Problema

La ruta `/debug-monitor` (Fiber Monitor) estaba registrada públicamente sin autenticación. Cualquier usuario podía acceder a métricas internas del sistema (CPU, memoria, goroutines, uptime).

```go
// ANTES:
router.Get("/debug-monitor", monitor.New(monitor.Config{Title: "DEV ONLY - Monitor"}))
```

---

## Corrección

La ruta ahora solo se registra cuando `APP_ENV=development`. En producción (default), la ruta no existe y retorna 404.

```go
// DESPUÉS:
if config.Envs.Environment == "development" {
    router.Get("/debug-monitor", monitor.New(monitor.Config{Title: "DEV ONLY - Monitor"}))
}
```

---

## Testing

- `go build ./...` — Pass
- `go vet ./...` — Pass
- Middleware tests: 15/15 Pass

---

## Impacto

- **Riesgo de regressión:** Mínimo. Solo condiciona el registro de una ruta.
- **Producción:** La ruta `/debug-monitor` no existe (404).
- **Desarrollo:** La ruta funciona normalmente.

---

## Archivos relacionados

| Archivo | Relación |
|---------|----------|
| `internal/router/admin_router.go:33` | Archivo modificado |
| `internal/config/env.config.go:95` | `Environment` default a `"production"` |
