# Persona de Contacto para Emergencias (`feat/contacto-emergencia`)

Permite que cada agremiado declare hasta **3 personas** a las que el Colegio
puede avisar cuando no successfully localizable al profesional o ante un
accidente. Los datos son de un **tercero** y **nunca se publican**: ni en el
directorio, ni en el perfil público, ni en el sitemap.

**Regla de negocio** (idéntica en Go y en el cliente): nombre y parentesco
siempre, y **al menos uno** de teléfono o correo.

---

## 1. Por qué una tabla aparte y no columnas en `psi_users`

Un contacto plural (1:N) no cabe en el perfil, y guardar estos datos en el
`FormData` del perfil fue descartado por una razón concreta: el guardado del
perfil **omite los campos vacíos**, así que un contacto guardado únicamente con
teléfono se borraría en el siguiente guardado del perfil que no lo incluyera.
Tabla nueva, soft delete y endpoints propios.

## 2. Modelo y migración

`PsiUserEmergencyContact` (`api/internal/domain/user.model.go`) →
`api/migrations/20260928132854_emergency_contacts.sql`:

| Columna | Tipo | Regla |
|---|---|---|
| `psi_user_id` | `uuid` | FK → `psi_users(id)`, indexada |
| `name` | `varchar(255)` | NOT NULL |
| `relationship` | `varchar(100)` | NOT NULL |
| `phone` | `varchar(20)` | — |
| `email` | `varchar(255)` | — |

Dos `CHECK` (**defensa en profundidad**, no sustituyen la validación del
servicio): `chk_emergency_contact_identified` (nombre y parentesco no vacíos) y
`chk_emergency_contact_channel` (teléfono o correo). Borrado lógico con
`deleted_at` (`gorm.DeletedAt`).

La API no conoce el catálogo de parentescos: solo exige que no venga vacío, así
que el texto libre de "Otro…" se persiste tal cual.

## 3. Endpoints

| Método | Ruta | Actor | Gate |
|---|---|---|---|
| `POST` | `/psi/me/emergency` | agremiado | `ProtectedPsiUser` (dueño) |
| `PATCH` | `/psi/me/emergency/:id` | agremiado | IDOR (contacto debe ser suyo) |
| `DELETE` | `/psi/me/emergency/:id` | agremiado | IDOR |
| `POST` | `/admin/psi/:id/emergency` | staff | `Sudo \|\| CanCreatePsi` |
| `PATCH` | `/admin/psi/:id/emergency/:contactId` | staff | `Sudo \|\| CanUpdatePsi` |
| `DELETE` | `/admin/psi/:id/emergency/:contactId` | staff | `Sudo \|\| CanDeletePsi` |

**No hay endpoint de listado**: el array llega embebido en `GET /psi/me` y en
`GET /admin/psi/:id` (mismo gate que ya abría la ficha, sin flags nuevos).

Errores de dominio (`internal/domain/errors.go` → `emergencyContactError` en
`internal/handler/psi_user_emergency.go`): 400 incompleto / correo inválido,
403 cuota + IDOR + RBAC (**mensaje genérico**, nunca datos de terceros), 404
inexistente. `MaxEmergencyContacts = 3` en el servicio.

## 4. Privacidad (la parte delicate)

`EmergencyContacts []PsiUserEmergencyContact` se preloadea **solo** en
`GetByID`. `GetByFPV`, `SearchDirectory` y `GetSitemapData` no la tocan, así que
el directorio y el perfil público no pueden verla ni por accidente.

El tag es `json:"emergency_contacts,omitempty"` a propósito: el endpoint público
del sitemap serializa el modelo crudo, y un tag plano expondría la clave (como
`null`) en respuestas públicas. Con `omitempty` la clave **desaparece** cuando
está vacía. Por eso el frontend la lee siempre como `emergencyContactsOf(x)` →
`?? []`.

## 5. Auditoría

Alta, edición y baja se registran en `api_change_logs` (mismo
`Record()` diferido y best-effort del resto del módulo) con diff
`emergency_contact_{name,relationship,phone,email}` en snake_case, y la etiqueta
de entidad `Prueba Notif (FPV …)`. Los `403` por IDOR **no** generan entrada
(nada se persistió). Rótulos en `FIELD_LABELS` (`web/src/types/audit.ts`).

## 6. Frontend

| Pieza | Ubicación |
|---|---|
| Catálogo de parentescos + regla de validación | `web/src/lib/emergency-contact.ts` |
| Sección (alta / edición / borrado / cuota) | `web/src/components/psi/profile/EmergencyContactsSection.tsx` |
| Pestaña del agremiado | `routes/psi/perfil.tsx` → `NotebookPage id="emergencia"` |
| Pestaña del admin | `routes/admin/psicologos/[id]/detalle.tsx` → `NotebookPage id="emergencia"` (Notebook de gestión) |

La misma sección sirve en los dos bandos (props `contacts` + `onAdd/onUpdate/
onDelete`); `audience` solo cambia los textos del aviso. El catálogo y las
reglas viven en el cliente para dar feedback inmediato, **replicando** (no
sustituyendo) la validación de Go.

## 7. Pruebas

| Suite | Qué cubre |
|---|---|
| `internal/service/emergency_contact_test.go` | 11 casos de tabla (normalización, campos cruzados, cuota), IDOR, RBAC, roles, reflexión del DTO público, `GetPublicProfile` sin fuga |
| `internal/repository/postgres/psi_repo_emergency_test.go` | CRUD, preload, orden, aislamiento entre agremiados, privacidad frente a `GetByFPV`/`SearchDirectory`, violación de los `CHECK` |
| `internal/integration/emergency_contact_flow_test.go` | flujo E2E psi (alta→edición→baja), flujo admin, privacidad pública incluido el sitemap |
| `web/src/lib/emergency-contact.ts` (Deno) | catálogo, `resolve/option`, cada rama de `validateEmergencyContact`, normalización del payload, `?? []` |

Además: smoke manual contra el stack Docker (alta con solo correo, 4º contacto
→ 403, IDOR de otro agremiado → 403 en `PATCH` y `DELETE`, cuota, bitácora con
actor `sudo`, y ausencia de la clave en directorio / perfil público / sitemap).

## 8. Notas operativas

- **Tests de repo/integración** requieren PostgreSQL en `:5433` (el `Makefile`
  lo levanta; en local, `socat TCP-LISTEN:5433,fork,reuseaddr
  TCP:127.0.0.1:5432`).
- **Limitador de login por IP**: todo el tráfico del host comparte el contador
  de `172.20.0.1`; para desbloquear,
  `docker exec colpsi_valkey valkey-cli del 172.20.0.1`.
- La sección del agremiado vive **dentro** del `<form>` del perfil: sus botones
  son `type="button"` y **no** anida otro `<form>` (el navegador descartaría la
  etiqueta al parsear el SSR). Enter se intercepta con `onKeyDown` para que no
  dispare el guardado del perfil. En esa pestaña la barra de guardado se oculta
  (`NotebookTabWatcher`), porque exige contraseña y tiene su propio guardado.
