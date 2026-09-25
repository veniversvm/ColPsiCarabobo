# Bitácora de cambios en API — Auditoría (`/admin/auditoria`)

> **Estado: implementado** (merge de `feat/audit-logs` → `main`).
> Plan original y alcance acordado: [`docs/plan-audit-logs.md`](./plan-audit-logs.md).

Registro forense de cambios a nivel de API: **qué cambió, quién lo cambió y
cuándo**, visible para el staff autorizado (`/admin`). Búsqueda por suceso,
entidad, actor y rango de fechas, con diff por campo `{from, to}` y exportación
CSV. Gateado por RBAC (`can_view_logs` / `can_export_logs`).

## Arquitectura: escritura diferida

```
PETICIÓN HTTP
  ├─ middleware AuditRequestMeta: inyecta IP/User-Agent en el ctx
  ├─ servicio muta la entidad; si persiste OK → RecordAudit(evt)
  │    (snapshot previo → diff {campo:{from,to}} → encola)
  ▼
  cola chan AuditEvent (buffer 5000 → drop + log.Warn si se llena)
  ▼
  worker background (ctx propio, NO el del request):
      INSERT por lotes de 50 filas / cada 1s (timeout 5s)
  ▼
  api_change_logs
```

El log es **best-effort y no bloquea**: un fallo del worker o una cola llena
jamás revierte la operación principal (mismo patrón productor-consumidor que
`mail_service.go`). El worker usa su propio contexto de fondo (`bgCtx` en
`cmd/api/main.go`) y drena la cola al apagarse.

Registro global **nil-safe**: `InitAuditLogs(auditSvc)` en el arranque
(`main.go:164`) registra el servicio en una variable global; sin instancia (tests)
`RecordAudit` no hace nada. Esto evita inyectar el servicio en los constructores
de todos los servicios del paquete (ver `service/audit_registry.go`).

## Modelo — tabla `api_change_logs`

| Campo | Descripción |
|---|---|
| `id` | uuid v7 |
| `entity` | `psi \| staff \| auth \| notificacion \| ticket \| post \| proyecto \| area \| ficha_inscripcion \| config` |
| `entity_id` | uuid del registro afectado (o slug) |
| `entity_label` | legible sin joins: "Ana Pérez (FPV 12345)" |
| `action` | `create \| update \| delete \| login \| logout \| reset_password \| transfer_sudo \| estado \| send \| move` |
| `actor_id` / `actor_role` / `actor_username` | quién; `actor_role` = `sudo`, preset (`secretaria`, …) o `psi` |
| `ip` / `user_agent` | del contexto de la petición |
| `changes` | jsonb: `{ "campo": { "from": ..., "to": ... } }` (diff) |
| `metadata` | jsonb libre: motivo/nota (p. ej. cambio de password o imágenes) |
| `created_at` | timestamp |

Índices (en el modelo): `(entity, entity_id, created_at)`, `(actor_id, created_at)`,
`(entity, action, created_at)` y `created_at`.

## RBAC

- Flags nuevos: `can_view_logs` (leer listado/detalle) y `can_export_logs` (CSV).
  Sudo los hereda vía `SeedSudoPermissions` (fuerza `true` en cada arranque);
  **ningún preset** los incluye por defecto.
- **No existe `can_purge_logs`**: borrar/limpiar logs es exclusivo de Sudo.
- Toda ruta admin devuelve **404 enmascarado** sin permiso (nunca 401/403),
  consistente con `ProtectedAdmin404()`.

## Endpoints de lectura

Todos bajo `/admin/audit-logs` con `middleware.NoStore()` + `ProtectedAdmin404()`.

| Endpoint | Filtros / uso | Gate |
|---|---|---|
| `GET /admin/audit-logs` | `q` (texto sobre `entity_label`), `suceso`, `entidad`, `actor_id`, `desde`, `hasta`, `page`, `limit` | `Sudo \|\| CanViewLogs` |
| `GET /admin/audit-logs/psi/:id` | Historial completo de un psicólogo | `Sudo \|\| CanViewLogs` |
| `GET /admin/audit-logs/export` | CSV con los filtros aplicados (sin paginar) | `Sudo \|\| CanExportLogs` |
| `GET /admin/audit-logs/stats` | Agregados por (entity, action) desde `since` | `Sudo \|\| CanViewLogs` |

### Paginación

- API: `page` (default 1) y `limit` (default 20, tope 100). El repositorio
  normaliza (`page < 1 → 1`; `limit < 1 o > 100 → 20`), ordena
  `created_at DESC` y responde `{ data, total, page, limit }`.
- Frontend: 20 por página; la barra `PaginationBar` se renderiza **arriba y
  abajo** del listado (visible solo con `total > 20`) y comparte el estado
  `page`; cualquier cambio de filtro resetea a la página 1.
- La **exportación CSV no está paginada**: descarga el conjunto completo que
  cumpla los filtros.

## Instrumentación (puntos de `RecordAudit`)

Se registra **solo tras persistencia exitosa** de la mutación (0 llamadas si el
UPDATE falla). `actor_id` puede ser un admin o el propio psicólogo
(auto-gestión, `actor_role="psi"`).

**Admin:**

- `psi_user_admin_service.go`: `CreatePsiByAdmin`, `UpdatePsiByAdmin` (diff por
  campo + `entity_label` = nombre + FPV), `DeletePsiByAdmin`,
  `ResetPsiPasswordByAdmin`, `DeleteProfilePictureByAdmin`.
- `admin_service.go`: login/logout de admin (`entity=auth` y `entity=staff`),
  CRUD de staff, cambios de permisos y `transfer_sudo`.
- `notification_service.go`: crear notificación y cambio de estado.
- Redes sociales del psicólogo por admin (*variantes `ByAdmin`* en
  `social_media.go`: POST/PATCH/DELETE `social/:id/social`), con cuota máx 10 y
  chequeo IDOR.

**Auto-gestión del psicólogo (`/psi/me/*`, actor = el propio psi):**

- `UpdateProfileSelf` (`psi_service_self_management.go`): diff por campo del
  perfil + `metadata` para password, imágenes de portada/título y `full_bio`.
- Académico (`psi_service_academic.go`): `AddPostGrade`, `UpdatePostGrade`
  (diff + metadata de certificados) y **`DeletePostGrade`** — la ruta eliminada
  `DELETE /psi/me/postgrades/:id` (bug preexistente del frontend) es **soft
  delete** vía `gorm.DeletedAt` embebido, con limpieza del certificado en S3
  **best-effort** y respeto al IDOR.
- Redes sociales (self, sin varianza admin): todas las mutaciones de `social/me`.

## Contrato de las claves del diff

- Claves **snake_case** = tags JSON reales del modelo (p. ej.
  `cell_phone_outside_venezuela`, no `phone_out_side`), porque la UI las
  traduce con `FIELD_LABELS` y las agrupa por clave.
- Redes sociales: `social_name` / `social_url` / `social_active`.
- Postgrado: `post_grade_title` / `post_grade_university` /
  `post_grade_graduation_year` / `post_grade_description`.
- `from` se omite en JSON si es nil (campo nuevo); el diff usa `{campo:{from,to}}`.

## Frontend (`/admin/auditoria`)

- **Tabs**: General (default), Por psicólogo, Por staff, Estadísticas — todas
  comparten el mismo listado + paginación (todo menos `stats`).
- `web/src/types/audit.ts`: `ApiChangeLog` con `changes`/`metadata` tolerantes a
  **objeto o string serializado** (historia previa al fix de la columna),
  `ACTION_LABELS` / `ENTITY_LABELS` / `FIELD_LABELS`,
  `parseAuditJson()` (parser defensivo, nunca `JSON.parse` directo) y
  `auditChangesKeys()`.
- `AuditLogDrawer.tsx`: diff `from → to` por campo, coloreado.
- Tarjetas del listado: resumen "✏️ Cambios: campo1 · campo2 …" (máx 4).
- Menú lateral: item "Auditoría" visible solo con `can_view_logs`/Sudo.
- Acceso desde las fichas: botón "Ver bitácora" en el detalle del psicólogo y
  "Ver actividad" en staff (deep-links `?psi_id=` / `?actor_id=`).

## Retención

Purga diaria (cron en `main.go`) de `api_change_logs` más antiguas que
`AUDIT_LOG_RETENTION_DAYS` (default **90**; `<=0` desactiva el purge).
Exclusivo de Sudo, no expuesto como flag.

## Verificación

```bash
# Backend
go build ./internal/... ./cmd/... ./pkg/... && go vet ./...
make test-unit                 # incluye audit_service/handler/social/academic

# Frontend (no hay typecheck; el build es la verificación)
cd web && npm run build

# Swagger (tras tocar handlers)
cd api && swag init -g cmd/api/main.go -o docs/
```

Flujo manual end-to-end: login admin → mutar un psi → el worker difunde (≤1s) →
`GET /admin/audit-logs/psi/:id` muestra quién/cuándo/qué. Sin `can_view_logs` →
404 enmascarado y el menú "Auditoría" no aparece.