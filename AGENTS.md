# Contexto para Agentes AI

## Stack

- **Frontend**: SolidStart (SolidJS) con TypeScript, renderizado SSR + CSR, rutas basadas en archivos
- **Backend**: API en Go (`api/`)
- **Auth**: JWT almacenado en `sessionStorage` (cliente) + HttpOnly cookie (server action)
- **Bucket**: MinIO/S3, URL centralizada vía `web/src/lib/bucket.ts`
- **Build**: Vinxi / Nitro, preset `deno-server-legacy`
- **CSS**: Tailwind

## Convenciones

- Nombres de archivos/rutas en minúsculas con guiones bajos (snake_case para español)
- Un solo commit por fix en la rama `docs`
- "no modifique el funcionamiento" — los cambios deben ser seguros y preservar el comportamiento existente

## Proyecto

Sitio web del Colegio de Psicólogos de Carabobo. Incluye:
- Páginas públicas (inicio, directorio, noticias, nosotros, explorar)
- Portal del psicólogo (`/psi/` — perfil, académico, modalidad de servicio, aviso de cumpleaños)
- Administración (`/admin/` — CRUD de psicólogos, noticias, áreas, staff, edad calculada, banner de cumpleaños, observaciones internas)
- Inscripciones (`/inscripcion` — ficha con validación de campos obligatorios que bloquea envío y edición admin)

## Seguridad aplicada (fixes recientes)

| # | Archivo(s) | Qué cambió |
|---|------------|-----------|
| 1 | `lib/sanitize-html.ts` | Wrapper DOMPurify para innerHTML |
| 2 | `lib/auth.tsx`, `lib/api.ts` | JWT de cookie → sessionStorage |
| 3 | `entry-server.tsx` | CSP meta tag |
| 4 | `lib/auth.tsx` | Cookie user_data secure + sameSite |
| 5 | `lib/bucket.ts`, 15+ archivos | Helper bucketUrl(), remove localhost:9000 |
| 6 | `deno.json`, `dockerfile` | Deno permisos granulares |
| 7 | `lib/errors.ts`, 12 archivos | Errores genéricos al usuario |
| 8 | `lib/utils.ts`, `crear.tsx` | parseInt radix 10 |
| 9 | `ImportXlsxModal.tsx` | Validación MIME + tamaño |
| 10 | `areas/crear/index.tsx` | Idempotency key |
| 11 | `directorio/index.tsx` | encodeURIComponent |
| 12 | `routes/robots.txt.ts` | robots.txt dinámico |
| 13 | `psi/academico.tsx` | rel noopener noreferrer |
| 14 | `aaaa[id].tsx` (eliminado) | Dead code removal |
| 15 | `api/internal/middleware/security_headers.go`, `api/cmd/api/main.go` | Headers de seguridad de la API: HSTS (via `HSTS_MAX_AGE`/`HSTS_PRELOAD`), Permissions-Policy y `Cache-Control: no-store` en auth/admin/psi-me |
| 16 | `psi/tickets/[id].tsx`, `admin/tickets/[id].tsx` | Chat sin recarga: eliminado `<Suspense>` global que causaba flash completo al enviar mensaje; respuesta de `apiPost` se añade al hilo al instante sin `refetch()` |
| 17 | RBAC, `admin_roles.go`, `admin_handler.go`, `user_admin_repo.go` | RBAC-liviano: 18 flags `can_*` en `user_admins`, presets de roles (Secretaría/Comunicación/Soporte/Proyectos/Lector), `GET /admin/me`, menú admin filtrado por permisos (ver `docs/plan-rbac-switches.md`) |
| 18 | `admin_permission_logs`, `POST /admin/transfer-sudo` | Sucesión de Sudo atómica con confirmación de contraseña y auditoría; botón "Ceder SUDO" en staff |
| 19 | `app_settings`, `settings_audit_logs`, `settings_service.go`, `settings_handler.go` | Interruptores globales de recepción (tickets/inscripciones), 409 `reception_disabled`, banners de UI, `ReceptionSwitchesCard` |
| 20 | `pkg/database/seed.go` | `SeedSudoPermissions` idempotente: fuerza la matriz completa `true` para `sudo=true` en cada arranque |
| 21 | `admin-access.tsx`, `login.tsx` | Aviso de rate-limit (429) con tiempo de espera real en login de admin y psicólogo |
| 22 | `admin/psicologos/[id]/detalle.tsx`, `psi_router.go`, `psi_user_social_admin.go`, `social_media.go` | Endpoints admin del detalle de psicólogo reparados: las 9 server actions se invocaban sin `useAction` (TypeError `singleFlight`) en Presencia Digital / Expediente Deontológico / Observaciones / Documentos; rutas admin de social en la API (POST/PATCH/DELETE `/admin/psi/:id/social`) con gate RBAC `Sudo || CanUpdatePsi || CanCreatePsi` (delete admite `CanDeletePsi`), cuota máx 10 y chequeo IDOR |
| 23 | bitácora de auditoría: `api_change_logs`, `audit_service.go`, `audit_router.go`, `audit_handler.go`, `audit_repo.go`, `/admin/auditoria` | Bitácora forense a nivel de API (qué/cuándo/quién) con escritura **diferida** (cola 5000 + worker en lotes de 50, best-effort, nunca revierte la operación); gates RBAC `can_view_logs`/`can_export_logs` (Sudo hereda, 404 enmascarado, sin `can_purge_logs`); endpoints `GET /admin/audit-logs{/,psi/:id,/export,/stats}` con paginación (20/100) y CSV sin paginar; retención `AUDIT_LOG_RETENTION_DAYS` (default 90) (ver `docs/audit-logs.md`) |
| 25 | `user_emergency_contacts`, migración `20260928132854_emergency_contacts.sql`, `psi_user_emergency*` (repo/service/handler/request/router), `psi_service_self_management.go`, `psi_service_admin_emergency.go` | **Persona de contacto para emergencias (1:N)**: tabla aparte con 2 `CHECK` (identificado + canal). Máximo 3, preload **solo** en `GetByID` (nunca en directorio/sitemap/perfil público), tag `json:"emergency_contacts,omitempty"` como barrera de privacidad, endpoints propios (psi auto-gestión y admin), IDOR/RBAC con mensajes genéricos, auditoría (`emergency_contact_*`) en alta/edición/baja; UI pestaña "Emergencia" en `/psi/perfil` y en ficha admin (ver `docs/plan-contacto-emergencia.md`). |
| 26 | `psi_users.last_change_reason`, migración `20260928160000_psi_users_last_change_reason.sql`, `psi_user_admin.go`, `psi_user_admin_service.go`, `psi_repository.go`, `detalle.tsx`, `types/audit.ts` | **Motivo del cambio al editar un psicólogo (admin)**: PATCH `/admin/psi/:id` exige `last_change_reason` (400 si hay cambios y falta motivo); se persiste en la ficha (columna interna, solo endpoints admin) y se anota en la bitácora (`Metadata.motivo`). El campo se excluye del chequeo de "request vacío" (un payload solo-con-motivo sigue siendo 400). Datos públicos (directorio, detalle, sitemap) no lo exponen. Max 500 chars. |
| 27 | `psi_service_directory.go`, `psi_repository.go` (`SearchDirectory`, `ResolveSpecialtyNames`), `psi_repository.go` domain interface, `request_structs/psi_user.go` (`PsiFullProfileDTO`), `[slug].tsx`, `types/psi.ts` | **Áreas de trabajo visibles solo para solventes**: chips del directorio y `work_areas` de la ficha pública se resuelven SOLO desde el catálogo `psi_specialty_models` (FK primero; fallback por coincidencia **exacta** del string legacy — jamás un área inventada) y únicamente se exponen si el psicólogo está solvente (la búsqueda de texto puede listar insolventes pero con la tarjeta SIN chips). `PsiFullProfileDTO` ya NO expone `solvent` (la solvencia es invisible al visitante; insolvente recibe solo el perfil de identidad). Frontend: `psi().work_areas` (antes `specialties`) + copy neutral de perfil no disponible. Sin migración (usa catálogo existente). |
| 28 | `inscription_service.go`, `audit_helpers.go`, `admin/inscripciones/[id].tsx`, `types/audit.ts` | **Bitácora + UX del guardado de notas en inscripciones**: `UpdateNotes` emite evento `ficha_inscripcion`/`update` con diff `notes {from→to}` (solo si el valor realmente cambió; `RecordAudit` global nil-safe). Web: el botón "Guardar notas" detecta el no-op (mismo texto ya persistido) y avisa "Las notas ya están guardadas" en lugar del engañoso "Notas guardadas"; etiqueta `notes` en `FIELD_LABELS`. |
| 29 | `inscription_request_model.go` (`PsiInscriptionNote`), migración `20260928173000_psi_inscription_note_history.sql`, `inscription_repository.go`, `inscription_service.go`, `inscription_handler.go`, `inscription_router.go`, `request_structs/inscription_requests.go`, `admin/inscripciones/[id].tsx`, `types/inscription.ts`, `docs/` | **Histórico de notas administrativas (1:N)**: cada guardado con cambio agrega una versión en `psi_inscription_notes` (texto completo + autor `create_by`/`create_by_id` + fecha; el histórico nace vacío, no migra notas existentes). `UpdateNotes` persiste nota actual + versión + bitácora; un guardado sin cambio real ya no reescribe (sin UPDATE, sin versión, sin evento). Nuevo `GET /admin/inscripciones/:id/notes` lista versiones desc (gate `canViewFicha`, 404 enmascarado). Web: bloque "Historial de notas (N)" en la tarjeta con cada versión (fecha · admin) que se refresca tras guardar. |

## Comandos (web)

```bash
npm run dev         # desarrollo (Vinxi)
npm run build       # build SSR + verificar errores → .output/
npm run start       # servir el build
```

> No hay typecheck configurado (TypeScript no está instalado); `npm run build` es la verificación.

Para el frontend en detalle (arquitectura, env vars, gotchas) ver [`web/README.md`](./web/README.md)
y [`web/AGENTS.md`](./web/AGENTS.md).
