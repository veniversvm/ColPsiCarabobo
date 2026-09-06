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

## Comandos (web)

```bash
npm run dev         # desarrollo (Vinxi)
npm run build       # build SSR + verificar errores → .output/
npm run start       # servir el build
```

> No hay typecheck configurado (TypeScript no está instalado); `npm run build` es la verificación.

Para el frontend en detalle (arquitectura, env vars, gotchas) ver [`web/README.md`](./web/README.md)
y [`web/AGENTS.md`](./web/AGENTS.md).
