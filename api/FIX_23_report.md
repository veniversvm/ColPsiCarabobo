# FIX-23: Desacoplar AnalyticsService de *gorm.DB

## Metadatos

| Campo | Valor |
|-------|-------|
| **FIX** | FIX-23 |
| **Hallazgo** | MED-04 |
| **Severidad** | MEDIO |
| **Archivos** | `domain/analytics_repository.go` (nuevo), `repository/postgres/analytics_repository.go` (nuevo), `domain/analytics.go`, `service/analytics_service.go`, `middleware/analytics.go`, `router/router.go`, `cmd/api/main.go`, `middleware/auth_test.go` |
| **Estado** | COMPLETADO |
| **Fecha** | 2026-07-25 |

---

## Problema

`AnalyticsService` dependía directamente de `*gorm.DB`, violando el Principio de Inversión de Dependencias (DIP). El middleware `AnalyticsMiddleware` también usaba `*gorm.DB` directamente.

---

## Solución

1. Creada interfaz `AnalyticsRepository` en `domain/analytics_repository.go` (12 métodos)
2. Creada implementación GORM en `repository/postgres/analytics_repository.go`
3. Movidas structs de dashboard (`DashboardStats`, `TopItem`, `TopProfile`, `DailyCount`) de `service/` a `domain/` para que la interfaz compile
4. Refactorizado `AnalyticsService` para recibir `domain.AnalyticsRepository` en vez de `*gorm.DB`
5. Refactorizado `AnalyticsMiddleware` para usar `*service.AnalyticsService` en vez de `*gorm.DB`
6. Actualizado `router.go` y `cmd/api/main.go` para crear el repo y pasarlo
7. Creado mock `mockAnalyticsRepo` en `auth_test.go` (eliminó dependencia de GORM DryRun)

---

## Verificación

- `go build ./...` → sin errores ✅
- `go test ./internal/middleware/...` → PASS ✅
- 2 failures pre-existentes (`TestAdminService_All`, `TestSpecialtyService_Update`) no son causados por este cambio ✅
