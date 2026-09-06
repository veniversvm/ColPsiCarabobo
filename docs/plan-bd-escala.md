# Plan: Escalabilidad BD/API para 3.000+ Psicólogos

## Resumen

Optimización del esquema, repositorios y paginación del API para soportar 3.000+
psicólogos concurrentes sin degradación de latencia.

---

## 1. P0 — Índices óptimos

### notification_targets

Antes: dos índices simples (`notification_id`, `psi_user_id`).

Ahora:

| Índice | Tipo | Columnas | Propósito |
|--------|------|----------|-----------|
| `idx_notification_targets_notification_psi` | UNIQUE | `(notification_id, psi_user_id)` | Garantiza un target por notificación×psi, acelera bulk-insert |
| `idx_notification_targets_psi_read` | B-tree | `(psi_user_id, is_read)` | `CountUnread` single-table sin JOIN |

**CountUnread simplificado**: solo consulta `notification_targets` con filtro
`(psi_user_id = ? AND is_read = false)` — evita el JOIN con `notifications`.

### tickets

Antes: sin índices compuestos.

Ahora:

| Índice | Tipo | Columnas | Propósito |
|--------|------|----------|-----------|
| `idx_tickets_psi_user_id` | B-tree | `(psi_user_id, id)` | ListMyTickets + keyset |
| `idx_tickets_estado_id` | B-tree | `(estado_id, id)` | Filtro admin + keyset FIFO |
| `idx_tickets_motivo_id` | B-tree | `(motivo_id, id)` | Filtro admin + keyset |

Se logran vía el `priority` del tag GORM en el campo `ID` de `Ticket`.

### psi_inscription_requests

| Índice | Tipo | Columnas | Propósito |
|--------|------|----------|-----------|
| `idx_inscription_requests_cedula_pending` | UNIQUE parcial | `(cedula)` WHERE `status = 'pending'` | Cédula única solo para inscripciones activas |
| `idx_inscription_requests_control_number` | UNIQUE parcial | `(control_number)` WHERE no vacío | Número de control único cuando existe |
| `idx_inscription_requests_status` | B-tree | `(status)` | Filtro por estado |
| `idx_inscription_requests_psi_user_id` | B-tree | `(psi_user_id)` | Lookup por usuario |

### posts

Sin cambios — tabla pequeña (< 500 filas en producción), ya tiene `idx_posts_deleted_at` de GORM `AuditModel`.

---

## 2. P1 — Keyset pagination (`?cursor=&limit=`)

### Contrato API

```
GET /notifications/psi-user?limit=15&cursor=<id>
GET /posts/?limit=10&cursor=<uuid>
GET /admin/tickets?limit=20&cursor=<id>

Respuesta:
{
  "data": [...],
  "next_cursor": "<último id>" | null,  // null = sin más páginas
  "total": N,
  "page": P
}
```

**`next_cursor` se emite siempre** que `len(data) == limit`, sin importar si la
primera página usa cursor o no. El frontend solo necesita seguir `next_cursor`
hasta que sea `null`.

### Notificaciones

- `ListByUser(ctx, psiUserID, cursor, page, limit)`: ORDER BY `id DESC`
  - Sin cursor: OFFSET (page, limit)
  - Con cursor: `WHERE id < $cursor` ORDER BY `id DESC` LIMIT $limit
- `ListBySender` usa la misma estrategia.
- `parseCursor(c)` en `notification_handler.go` extrae el UUID del query param.

### Posts

- `PostFilter.Cursor uuid.UUID` (no nullable; `uuid.Nil` = sin cursor)
- `post_repo.List`: ORDER BY `id DESC`
  - Cursor nil: OFFSET (page, limit) — compatible con el SSR inicial del frontend
  - Cursor != nil: `WHERE id < $cursor` ORDER BY `id DESC` LIMIT $limit

### Tickets admin (FIFO)

- `TicketFilter.Cursor *uint`
- `ticket_repo.ListTickets`: ORDER BY `tickets.id ASC` (FIFO estricto)
  - Sin cursor: OFFSET (page, limit)
  - Con cursor: `WHERE tickets.id > $cursor` ORDER BY `tickets.id ASC` LIMIT $limit
- `ListMyTickets` (portal psi) se mantiene en OFFSET — listas diminutas por usuario.

---

## 3. Rebase de migraciones

Se eliminaron 14 migraciones históricas y se reemplazaron por una única baseline
generada desde los modelos GORM con `atlas migrate diff baseline --env gorm`.

### Archivo actual

`api/migrations/20260906232227_baseline.sql` (~883 líneas)

### Proceso

```bash
# 1. Backup
cp -r api/migrations /tmp/opencode/migrations_backup

# 2. Eliminar migraciones viejas (conservar intrucciones.txt y README.md)
rm api/migrations/20260*.sql api/migrations/atlas.sum

# 3. Generar nueva baseline
cd api && atlas migrate diff baseline --env gorm

# 4. Verificar drift
docker run --rm -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=colpsi_drift \
  -d --name colpsi_drift postgres:18-alpine
sleep 3
docker exec colpsi_drift psql -U postgres -d colpsi_drift -c 'CREATE EXTENSION IF NOT EXISTS unaccent'
docker exec colpsi_drift psql -U postgres -d colpsi_drift -f /path/to/baseline.sql
pg_dump | diff - against baseline → confirmar solo diffs intencionados
```

### Verificación de drift

Column sets idénticos en 37 tablas. Las únicas diferencias son los índices
compuestos nuevos (`notification_targets`, `tickets`) y los parciales de
`inscription_requests` — todos intencionados.

---

## 4. Despliegue en producción (DB existente)

**NO** recrear la DB. Usar baseline no destructiva:

```bash
# Atlas detecta las migraciones ya aplicadas y aplica solo el diff
docker compose exec migrador atlas migrate apply \
  --baseline 20260902161912 \
  --url postgres://user:pass@host:5432/colpsi?sslmode=disable
```

El flag `--baseline` le dice a Atlas que la DB ya tiene las migraciones hasta
esa versión, y que solo debe aplicar las diferencias entre esa baseline y el
estado actual de los archivos `.sql`.

> Si la baseline coincide exactamente con el estado de la DB, no aplica nada.

---

## 5. Hallazgo: `atlas_schema_revisions`

Tras aplicar la migración, la tabla `atlas_schema_revisions` **no existe** en
`public` — vive en su propio esquema dedicado `atlas_schema_revisions` (comportamiento
por defecto del atlas canary v1.0.1-e0b5899).

Esto es normal y **no afecta** el idempotencia del migrador: re-ejecutar
`docker compose up migrador` retorna `"No migration files to execute"` con exit 0.

---

## 6. Seeds de desarrollo

Tras recrear la DB dev, el seed automático crea el admin (`admin`/`admin123`)
si la tabla `user_admins` está vacía.

Cuentas de prueba manuales:

| Cuenta | Email | Contraseña | Notas |
|--------|-------|------------|-------|
| psico01 | psico01@test.colpsi.local | TestPsi9900! | 5 notifs, 2 leídas |
| psico02 | psico02@test.colpsi.local | TestPsi9900! | 5 notifs, 0 leídas |
| psico03 | psico03@test.colpsi.local | TestPsi9900! | 5 notifs, 0 leídas |

Rate-limit login: `docker compose exec -T valkey valkey-cli FLUSHDB`

---

## 7. Gotchas conocidos

| Problema | Causa | Solución |
|----------|-------|----------|
| `cached plan must not change result type` | Planes preparados de Postgres stale tras `DROP SCHEMA` con conexiones abiertas en pgbouncer | `docker compose restart pgbouncer api` |
| Admin login 429 tras intentos fallidos | Rate-limiter persistente en valkey | `valkey-cli FLUSHDB` |
| Posts sin `next_cursor` en 1ª página | Bug: condición `cursor != nil` en `post_service.go` | Corregido: se emite cuando `len == limit` |
| Admin tickets sin `next_cursor` | Mismo bug en `ticket_handler.go` | Corregido |
