# Bitácora de cambios en API — audit logs (`feat/audit-logs`)

> Plan aprobado para implementar. Destino en repo: `docs/plan-audit-logs.md`
> (copiar/cumpleaños al pasar a implementación, un commit por fix en `docs`).

Registro de cambios a nivel de API para que el staff (`/admin`) pueda auditar
**qué cambió, quién lo cambió y cuándo**, con búsqueda por suceso, rango de
fechas, entidad y actor, siempre gateado por permisos RBAC.

El diseño clave es la **escritura diferida**: los eventos se capturan
síncronamente en el flujo de la petición pero se persisten en background por un
worker, de modo que **el log nunca interfiere con la operación principal**
(mismo patrón productor-consumidor que `api/internal/service/mail_service.go`).

## Alcance (qué se registra y qué no)

| Registra | No registra |
|---|---|
| Auth: login/logout de admin **y** de psicólogo | Eventos públicos/anónimos (directorio, inicio, etc.) |
| CRUD admin de psicólogos (con diff por campo) | Lecturas (`GET`) sin mutación |
| CRUD de staff, cambios de permisos, `transfer_sudo` | |
| Notificaciones (enviar/cancelar/programar) | |
| Posts/noticias, tickets, kanban, áreas, fichas de inscripción | |
| Autogestión del psicólogo (el actor es el propio psi) | |

Las tablas de auditoría previas (`admin_permission_logs`, `settings_audit_logs`,
`ticket_status_logs`, `LoginEvent`) **se mantienen intactas**; este sistema es
independiente y no las toca.

## Modelo — tabla `api_change_logs`

```go
type ApiChangeLog struct {
    ID             uuid.UUID     // v7
    Entity         string        // psi | staff | notificacion | ticket | post | proyecto | area | ficha_inscripcion | auth | config
    EntityID       string        // uuid del registro afectado (o slug)
    EntityLabel    string        // legible para la UI sin joins: "Ana Pérez (V-12345678)"
    Action         string        // create | update | delete | login | logout | reset_password | transfer_sudo | estado | send | move
    ActorID        uuid.UUID
    ActorRole      string        // sudo | secretaria | ... | psi
    ActorUsername  string
    IP             string
    UserAgent      string
    Changes        datatypes.JSON // diff: { "campo": { "from": ..., "to": ... } }
    Metadata       datatypes.JSON // libre: motivo, nota (opcional)
    CreatedAt      time.Time
}
```

**Índices** (cubren las consultas de búsqueda):

- `(entity, entity_id, created_at DESC)` → historial completo de un psi/ficha
- `(actor_id, created_at DESC)` → "todo lo que hizo este admin/psi"
- `(created_at DESC)` y `(action, created_at)` → listado general con rango
- `(entity, action, created_at DESC)` → agregados (`stats`)

## Escritura diferida (núcleo)

```
PETICIÓN HTTP
  ├─ Captura síncrona y barata: actor (c.Locals), IP, User-Agent,
  │    entidad, acción, diff {campo:{from,to}}, entity_label
  └─ auditSvc.Record(evt) ──► queue chan AuditEvent (buffer ~5000)
                                  │ (no bloquea; si está llena → drop + log.Warn)
                                  ▼
                        Worker background (1 goroutine, ctx propia)
                                  │ INSERT por lotes (p. ej. 50 filas/100ms)
                                  ▼
                            api_change_logs
```

- El worker usa su **propio `ctx` de fondo** (nunca el del request, que se
  cancela al terminar el handler). Conexión al arranque como `mail_service.go`.
- `Record()` captura todo en el momento de la petición y solo encola.
- Shutdown ordenado vía el `bgCtx/bgCancel` existente en `api/cmd/api/main.go`
  (línea 159/312): el worker drena la cola antes de apagar.
- Test del worker sin arrancarlo (inspección de cola): patrón ya documentado en
  `api/internal/service/mail_service_resend_test.go:59`.

## Permisos RBAC nuevos (18 → 20 flags)

| Flag | Alcance |
|---|---|
| `can_view_logs` | Ver listado y detalle de la bitácora |
| `can_export_logs` | Exportar CSV |

- Sudo los hereda automáticamente vía `SeedSudoPermissions`
  (`api/pkg/database/seed.go:98`).
- **Ningún preset** (Secretaría, Comunicación, Soporte, Proyectos, Lector) los
  incluye por defecto; se otorgan a medida en la UI de staff.
- **No existe `can_purge_logs`**: borrar/limpiar logs es exclusivo de Sudo.

## Fases

| Fase | Backend (Go) | Frontend | Tests |
|------|--------------|----------|-------|
| **F1 — Modelo + repo** | `domain/audit_log.model.go`; registro en `pkg/database/migration.go` (AutoMigrate) y `cmd/exp/migrate/main.go` (Atlas); `domain/audit_repository.go` (`CreateBatch`, `List`, `Count`) + impl en `repository/postgres/audit_repo.go` | — | `audit_repo_test.go` (PG real :5433): batch, filtros, paginación |
| **F2 — Worker diferido** | `service/audit_service.go`: `Record()` (encola), `startWorker(ctx)` (drena en lotes), `buildDiff(before, after)`; inyección en `cmd/api/main.go` junto a `mailSvc` | — | `audit_service_test.go` (cola sin worker, patrón `mail_service_resend_test.go:59`): encola sin bloquear, drop al lleno, lotes, `buildDiff` |
| **F3 — RBAC** | `domain/user.model.go` (+2 bool); `seed.go` (matriz + `SeedSudoPermissions`); `service/admin_roles.go` (`PermissionSet`, `AdminPermissionSet`, `diffPermissionSet`); `handler/admin_handler.go` (`/admin/me`) | `lib/staff-permissions.ts` (`PermissionState`, `PERM_KEYS`, 18→20); UI crear/editar admin (grupo "Auditoría") | Unit: `/admin/me` expone flags nuevos; `SeedSudoPermissions` los fuerza `true` en Sudo |
| **F4 — Endpoints de lectura** | `handler/audit_handler.go` + `router/audit_router.go` (grupo `/admin`, `NoStore` + `ProtectedAdmin404`) | — | `audit_handler_test.go`: parseo de filtros/rango, gate `Sudo \|\| CanViewLogs`, 404 si no hay permiso |
| **F5 — Instrumentación** | `Record()` post-persistencia en los servicios (tabla abajo) | — | Mocks por servicio: `Record` llamado tras éxito (0 llamadas si la persistencia falla) |
| **F6 — UI panel admin (secciones)** | — | Menú + rutas de secciones (abajo), `components/admin/audit-log-detail.tsx`, tipos en `types/admin.ts`, botones "Ver bitácora" | — |
| **F7 — Batería de tests** | — | — | Ver sección "Fases de testeo" más abajo |
| **F8 — Retención** | Purga por antigüedad `AUDIT_LOG_RETENTION_DAYS` (default 90); índices y `EXPLAIN` documentados en `docs/plan-bd-escala.md` | — | Unit de purga (borra solo > N días); `EXPLAIN` de cada consulta filtrada |

## Endpoints de lectura

Todos bajo `/admin` con `middleware.NoStore()` + `ProtectedAdmin404()` (404
enmascarado sin JWT, como el resto).

| Endpoint | Filtros / uso | Gate |
|---|---|---|
| `GET /admin/audit-logs` | `q` (texto sobre `entity_label`), `suceso` (acción), `entidad`, `actor_id`, `desde`, `hasta`, `page`, `limit` | `Sudo \|\| CanViewLogs` |
| `GET /admin/audit-logs/psi/:id` | Historial completo de un psicólogo (con su actor) | `Sudo \|\| CanViewLogs` |
| `GET /admin/audit-logs/export` | CSV con los filtros aplicados | `Sudo \|\| CanExportLogs` |
| `GET /admin/audit-logs/stats` | Agregados por suceso/entidad (horizonte, default 24h) | `Sudo \|\| CanViewLogs` |

## Instrumentación (puntos de `Record`)

**Prioridad alta:**
1. `admin_service.go:59-126` → **Login y Logout** de admin (`entity=auth`; hoy
   no audita nada).
2. Login/logout de psicólogo (véase `psi_handler.go:404,441`, junto al
   `LoginEvent` existente).
3. `psi_user_admin_service.go` → `CreatePsiByAdmin`, `UpdatePsiByAdmin`
   (**diff por campo**, `entity_label` = nombre + CI), `DeletePsiByAdmin`,
   `ResetPsiPasswordByAdmin`, `DeleteProfilePictureByAdmin`.
4. `admin_service.go` → CRUD staff, cambios de permisos, `transfer_sudo`.
5. `notification_service.go` → enviar / cancelar / programar notificaciones.

**Prioridad media (mismo patrón):**
6. `post_service.go` — posts/noticias.
7. Tickets — estados, cierre, motivos/estados.
8. Kanban — proyectos/columnas/tarjetas/miembros, mover tarjeta.
9. `specialty_*` — áreas de ejercicio profesional.
10. Fichas de inscripción — edición admin / cambio de estado.
11. Autogestión psi — perfil, académico, social, deontológico, documentos
    (actor = el propio psicólogo, `ActorRole="psi"`).

## Fases de testeo (F7)

Corre después de cada fase de código y agrupa la batería completa. Convención
del repo: tests con mocks propios (func-override, sin gomock), `-p 1` (serial).

### Unitarios (`make test-unit`, sin DB)

| Test | Cubre |
|---|---|
| `audit_service_test.go` | `Record()` encola sin bloquear (no difiere el flujo); drop + `log.Warn()` si la cola está llena; el worker (construido sin arrancarlo) procesa lotes; `buildDiff(before, after)` genera `{campo:{from,to}}` y omite campos iguales/cero |
| `audit_handler_test.go` | Parseo de filtros (`suceso`, `entidad`, `actor_id`, `desde`/`hasta`, `page`/`limit`); gate `Sudo \|\| CanViewLogs`; `CanExportLogs` solo en `/export`; sin permiso → 404 enmascarado; rango inválido → 400 genérico |
| RBAC | `/admin/me` incluye `can_view_logs`/`can_export_logs`; `diffPermissionSet` detecta cambio en los 2 flags nuevos; `SeedSudoPermissions` los fuerza `true` |
| Instrumentación | Por cada servicio instrumentado: mock que verifica que `Record` se invoca **solo tras persistencia exitosa** (0 llamadas si el `Update` falla → el log no registra mutaciones no aplicadas) |
| Purga (F8) | Solo borra filas con `created_at` más antiguas que `AUDIT_LOG_RETENTION_DAYS` |

### Repositorio (`make test-repo`, PostgreSQL real :5433)

`audit_repo_test.go`:
- `CreateBatch` inserta lote completo (y falla limpio si 1 fila es inválida).
- `List` con cada filtro y **combinaciones**: `entity+entity_id`, `actor_id`,
  `action`, rango `desde/hasta`, `q` sobre `entity_label`, orden
  `created_at DESC`, paginación consistente (`Count` = total real).
- Cobertura de índices: `EXPLAIN ANALYZE` de las 4 consultas de búsqueda
  (historial psi, actor, rango, general) → documentar planes en
  `docs/plan-bd-escala.md`.

### Integración (`make test-integration`, stack completo)

- Flujo completo: login admin → PATCH psi → evento difundido (espera de
  worker) → `GET /admin/audit-logs/psi/:id` devuelve la fila con actor correcto.
- Login/logout admin y psi → filas `auth` (hueco histórico cerrado).
- Cola llenada con `N > buffer` eventos no bloquea la petición HTTP que los
  origina (métrica de latencia o timeout del handler intacto).

### Seguridad (`make test-security`)

- Sin JWT → `GET /admin/audit-logs` responde 404 enmascarado (nunca 401/403).
- JWT de psi (no admin) en ruta admin → 404.
- Admin **sin** `can_view_logs` → 404; **con** flag → 200 (matriz de permisos).
- `export` sin `can_export_logs` → 404.
- Filtro `actor_id` ajeno no filtra datos de otros actores sin permiso
  (no hay "fuga por query param").

### E2E manual (panel admin)

Checklist de aceptación de la UI (sección siguiente): navegación de secciones,
búsqueda por suceso, rango de fechas, detalle con diff, export CSV, menú
oculto sin permiso.

### Ejecución final

```bash
make test-all    # unit + repo + integration + security, serial (-p 1)
go vet ./...
npm run build    # frontend
```

## Secciones del panel admin para ver los logs (F6)

La bitácora entra al panel como **sección propia del menú lateral** con
sub-secciones internas (tabs en `/admin/auditoria`), más puntos de integración
en las fichas existentes.

### Menú lateral (`web/src/routes/admin.tsx`)

- Nuevo item **"Auditoría"** (🧾, `path: /admin/auditoria`) visible solo con
  `can_view_logs` o Sudo (bypass). El backend sigue siendo la barrera real.

### Ruta `/admin/auditoria` — tabs (secciones internas)

| Tab | Contenido | Endpoint |
|---|---|---|
| **General** (default) | Listado con filtros: barra `q`, select de **suceso** (crear/editar/eliminar/login/logout/reset/…), select de **entidad**, select de **actor**, **rango de fechas** (`desde`/`hasta`), paginación | `GET /admin/audit-logs` |
| **Por psicólogo** | Buscador de psi (nombre/CI) → historial completo con quién cambió cada campo, cuándo, IP | `GET /admin/audit-logs/psi/:id` |
| **Por staff** | "Qué hizo cada admin": agrupado por `actor_id` con sucesos recientes (filtro por actor) | `GET /admin/audit-logs?actor_id=…` |
| **Estadísticas** | KPIs de sucesos por entidad/acción (horizonte 24h/7d/30d) — alimenta el dashboard | `GET /admin/audit-logs/stats` |

- **Tabla base** (todos los tabs): fecha · suceso (badge con color) · entidad
  (archivo: valor) · quién (badge con rol) · IP · botón "ver cambios".
- **Drawer de detalle** (`components/admin/audit-log-detail.tsx`): diff
  `from → to` coloreado por campo.
- **Exportar CSV** (botón en la barra de filtros): visible solo con
  `can_export_logs`; exporta lo que los filtros actuales seleccionan.

### Puntos de integración en secciones existentes del panel

| Sección | Qué se suma |
|---|---|
| **Psicólogos → ficha** (`admin/psicologos/[id]/detalle.tsx`) | Botón "Ver bitácora" → lleva a `/admin/auditoria` tab *Por psicólogo* preseleccionado con ese psi |
| **Staff** (`admin/staff`) | Por admin: acción "Ver actividad" → tab *Por staff* con `actor_id` del seleccionado |
| **Dashboard** (`/admin`) | Tarjeta "Actividad reciente" (últimos sucesos, solo si `can_view_logs`) — opcional, consume `stats` |
| **Notificaciones / Tickets / Proyectos** (listados) | Enlace discreto "Ver en auditoría" por registro → `entity=notificacion|ticket|proyecto` filtrado |

## Verificación (gate de cada fase)

1. **Por fase**: `go build ./... && go vet ./...` + los tests de esa fase
   (columna "Tests" de la tabla de fases).
2. **F7 — batería completa**: `make test-all` (unit + repo + integration +
   security, serial `-p 1`).
3. **Frontend**: `npm run build` y E2E manual del panel (checklist de F7).
4. **Prueba de humo funcional**: PATCH a un psi desde admin → el worker difunde
   → `GET /admin/audit-logs/psi/:id` muestra quién/cuándo/qué.
5. Login/logout admin y psi → filas `auth`.
6. Sin `can_view_logs` → `GET /admin/audit-logs` responde 404 enmascarado y el
   menú "Auditoría" no aparece en el panel.

## Notas operativas

- El log es **best-effort y diferido**: un fallo del worker jamás revierte la
  operación principal (convención del repo). Si la cola se llena se descartan
  eventos con `log.Warn()` — el log es evidencia, no carga.
- Retención obligatoria (purga por antigüedad) para evitar crecimiento infinito
  de `api_change_logs`.
- Respetar gotchas existentes: server actions con `useAction()`, `VITE_*` se
  inlinean en build, snake_case, un solo commit por fix en `docs`.