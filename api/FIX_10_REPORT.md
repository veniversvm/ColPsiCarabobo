# FIX-10 Report — Panic en startup si SMTP falla

| Campo | Valor |
|-------|-------|
| **Fix ID** | FIX-10 |
| **Hallazgo original** | HIGH-02 |
| **Archivo modificado** | `internal/router/psi_router.go` |
| **Fecha de implementación** | 2026-07-25 |
| **Estado** | Completado |

---

## Problema

Si el servidor SMTP no estaba disponible al iniciar la API, `panic("Error al inicializar...")` caía toda la aplicación. Nota: `admin_router.go` ya manejaba esto correctamente con `log.Printf`.

```go
// ANTES:
mailService, err := service.NewMailService()
if err != nil {
    panic("Error al inicializar el servicio de correo: " + err.Error())
}
```

---

## Corrección

Reemplazar `panic` por `log.Printf` (como ya hacía `admin_router.go`). La API continúa sin correo si SMTP no está disponible. `SendEmail` en `mail_service.go:132` ya tiene nil-check y retorna error si el servicio no está listo.

```go
// DESPUÉS:
mailService, err := service.NewMailService()
if err != nil {
    log.Printf("⚠️  Advertencia: No se pudo conectar al servidor SMTP: %v", err)
}
```

---

## Testing

- `go build ./...` — Pass
- `go vet ./...` — Pass
- Middleware tests: 15/15 Pass

---

## Impacto

- **Riesgo de regressión:** Bajo. `SendEmail` ya tiene nil-check en línea 132.
- **Producción:** Si SMTP falla, la API funciona pero no envía correos (se loguea la advertencia).

---

## Archivos relacionados

| Archivo | Relación |
|---------|----------|
| `internal/router/psi_router.go:20-23` | Archivo modificado |
| `internal/router/admin_router.go:20-22` | Referencia: ya manejaba el error correctamente |
| `internal/service/mail_service.go:132-134` | Nil-check en SendEmail |
