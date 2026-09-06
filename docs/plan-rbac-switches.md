# RBAC-liviano + interruptores de recepción (`feat/rbac-switches`)

Sistema de permisos por bandera para el staff (`/admin`) y control global de
recepción de tickets e inscripciones. Implementado en 4 fases sobre
Clean Architecture (Go) sin migración pesada de roles: **los flags booleanos en
`user_admins` son la fuente de verdad**; el campo `role` es solo una etiqueta.

## Modelo

- `UserAdmin` expone **18 flags** (`can_*`): `read/create/update/delete` psi,
  `create/update/delete` admin, `publish`, `update/delete` publish,
  `send/manage/read` notifications, `create/edit/delete` tags,
  `manage_projects`, `manage_tickets`, `read_psi` (lector sin editar).
- `sudo` sigue siendo la llave maestra: toda autorización lo bypassa
  (`admin.Sudo || flag`). 37 gates en el backend.
- Presets de roles en `api/internal/service/admin_roles.go`
  (`AdminPermissionSet` / `PresetRole`): Secretaría, Comunicación, Soporte,
  Proyectos y Lector. Servidos vía `GET /admin/roles/presets`.

## Fases

| Fase | Implementación | Endpoints / UI |
|------|----------------|----------------|
| 0 — `e65648c` | `can_manage_tickets` expuesto, gates de tags corregidos, dead code | — |
| 1 — `31e5a05` | `Role` + `CanReadPsi`, `admin_roles.go`, 5 presets | `GET /admin/roles/presets`; selector de rol + badge X/18 en `admin/staff` |
| 2 — `8053abe`, `1ee36f0` | `GET /admin/me` (`AdminPermissionSet`) | menú admin en `web/src/routes/admin.tsx` filtrado por flags (bypass si `sudo`) |
| 3 — `fb77434`, `9666c4f` | tabla `admin_permission_logs`, `POST /admin/transfer-sudo` | "Ceder SUDO" con confirmación de contraseña, switch atómico (índice único parcial `sudo`) + auditoría |
| 4 — `8e6cc4a`, `c6bfa64` | KV `app_settings` + `settings_audit_logs`, migración `20260905235529_add_app_settings.sql` | `GET/POST /admin/settings/reception` (POST solo Sudo); `GET /inscripcion/status` y `GET /psi/tickets/status` (públicos); 409 `reception_disabled` bloquea `InscriptionService.Submit` y `TicketService.CreateTicket`; banners en `InscriptionForm.tsx` y `psi/tickets/crear.tsx`; `ReceptionSwitchesCard.tsx` en el dashboard (visible solo Sudo) |

## Fixes asociados

- **`SeedSudoPermissions`** (`api/pkg/database/seed.go`, se ejecuta en cada
  boot tras `SeedAdmin`): idempotente, fuerza toda la matriz a `true` para
  `sudo = true`. Los SUDO pre-existentes (creados antes de que nacieran
  `can_read_psi` / `can_manage_tickets` con `DEFAULT false`) quedaban con flags
  apagados que la UI mostraba mal aunque la autorización ya bypassaba.
- **Aviso de rate-limit** (`web/src/routes/admin-access.tsx`,
  `web/src/routes/login.tsx`): cuando el limitador corta con 429, ambos logins
  muestran el aviso real del backend con el tiempo de espera ("Intenta de nuevo
  en 30/15 minutos") en vez del fallback genérico de `getUserFacingError`.

## Notas operativas

- **Limitadores por IP**: admin 5 intentos / 30 min; psi 10 / 15 min.
  Persisten en Valkey (`VALKEY_ADDR`). Gotcha en local: todo el tráfico del host
  comparte el contador de la IP `172.20.0.1` (gateway Docker); desbloquear con
  `docker exec colpsi_valkey valkey-cli del 172.20.0.1`.
- **Rutas admin enmascaran 404 sin JWT** (`ProtectedAdmin404`), nunca 403.
- La batería de tests de integración / E2E (`make test-security`) requiere el
  stack Docker completo (PostgreSQL `:5433`); fuera de él no corren pero no son
  fallos de código.