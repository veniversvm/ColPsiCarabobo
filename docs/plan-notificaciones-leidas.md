# Plan: Notificaciones leídas / no leídas (portal psi)

## Contexto

El portal psi mostraba las notificaciones sin distinción de lectura y el simple acto de
abrir una la marcaba como leída (comportamiento implícito). Como consecuencia no había
forma de saber cuáles comunicados quedaban pendientes de revisión.

## Semántica decidida

- **Abrir una notificación NO la marca como leída.**
- La lectura es una acción **explícita** del psicólogo: pulsar `✓ Marcar como leída`.
- La UI refleja el cambio **al instante** (actualización optimista) y lo revierte si el
  servidor no lo confirma.

## Backend (API Go)

- `GET /api/v1/notifications/psi-user` — cada item expone `targets[0]` (el target del
  psicólogo autenticado) con `is_read`/`read_at`, gracias a un `Preload` condicionado por
  `psi_user_id` en `ListByUser`.
- `GET /api/v1/notifications/psi-user/unread-count` — contador para badges.
- `PATCH /api/v1/notifications/psi-user/{id}/read` — marca como leída (semántica
  explícita). Responde `403` si la notificación no pertenece al agremiado
  (`ErrNotificationTargetNotOwned`).
- El auto-marcar al abrir (vía `GET /:id`) queda deprecado en favor del PATCH: la UI ya no
  llama a `GET /:id` cuando se abre un comunicado.

## Frontend (SolidStart)

`web/src/routes/psi/notificaciones.tsx`:

- Chip `● No leída` (azul) / `✓ Leída` (gris) en cada tarjeta.
- Botón `✓ Marcar como leída` visible **en la tarjeta colapsada** (sin expandir) y también
  dentro del panel expandido. La tarjeta es un `<div role="button">` (con soporte de
  teclado `Enter/Espacio`) para permitir el botón anidado.
- Estado local `readIds` (Set) para el reflejo inmediato; `targets[0].is_read` del API como
  respaldo persistido.
- Actualización optimista con revert en caso de error; el badge `X nuevas` se refresca al
  confirmarse en el servidor.

## Ficheros

- `api/internal/handler/notification_handler.go` (`MarkNotificationRead`)
- `api/internal/router/notification_router.go` (`PATCH /:id/read`)
- `api/internal/service/notification_service.go` (`MarkAsRead`)
- `api/internal/repository/postgres/notification_repo.go` (`ListByUser` con Preload de targets)
- `web/src/routes/psi/notificaciones.tsx`