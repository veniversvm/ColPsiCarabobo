# 🗃️ Migraciones de Base de Datos (migrations/)

> **[⬆ API](../)** — `api/migrations/`

Migraciones de base de datos gestionadas por **[Atlas](https://atlasgo.io)**. El
esquema se construye aplicando los archivos `.sql` de este directorio **en orden
alfabético**: el nombre de cada archivo es su versión.

## Archivos

| Archivo | Añade |
|---------|-------|
| `20260906232227_baseline.sql` | Esquema inicial (889 líneas, 37 tablas, 86 índices) |
| `20260911162745_password_reset_tokens.sql` | Tabla de tokens de recuperación de contraseña + 2 índices (el del hash es único) |
| `20260914140000_posts_publish_scheduled_idx.sql` | Índice de noticias programadas |
| `20260924141347_audit_api_change_logs.sql` | Tabla de la bitácora de auditoría + 5 índices + 2 flags RBAC (`can_view_logs`, `can_export_logs`) |
| `20260928132854_emergency_contacts.sql` | Persona de contacto para emergencias + 2 índices + 2 `CHECK` |
| `20260928160000_psi_users_last_change_reason.sql` | Columna `psi_users.last_change_reason` |
| `20260928173000_psi_inscription_note_history.sql` | Histórico de notas administrativas de inscripciones + 2 índices |
| `20260928190000_inscription_unique_fpv_correo.sql` | 2 índices **únicos parciales** (FPV y correo) que cierran la ventana de carrera del doble envío |
| `20260928210000_inscription_reject_reason.sql` | Columna `psi_inscription_requests.reject_reason` |
| `20260929210000_psi_terms_acceptance.sql` | Aceptación de Términos y Condiciones + 3 índices (uno único por usuario y versión) |
| `20260930160000_unaccent_extension.sql` | Extensión `unaccent` (la del buscador) |
| `20261001100000_analytics_ip_hash.sql` | Renombra `page_views/search_events/profile_views.ip` a `ip_hash` (64), para que la huella SHA-256 entre tras el fix de privacidad. **No toca** `login_events`, `active_sessions` ni `psi_terms_acceptance` (IP en claro, bitácora de seguridad y registro de aceptación). |
| `atlas.sum` | Checksums del directorio (**lo genera Atlas, nunca se edita a mano**) |
| `atlas.hcl` (en `api/`, no aquí) | Define el env `gorm`: fuente de verdad = modelos, destino = `file://migrations` |
| `intrucciones.txt` | Guía rápida de uso de Atlas |

## Estado del esquema

**42 tablas**, **150 índices** y **220 constraints** en el schema `public`.

Los 150 índices se reparten en 42 primary keys, 22 índices únicos y 86 no
únicos. Las migraciones crean 103 índices de forma explícita (86 no únicos +
17 únicos); los 42 PK y los 5 `UNIQUE` restantes van declarados dentro del
`CREATE TABLE`.

**Extensiones:** `plpgsql` (del propio PostgreSQL) y `unaccent`, que declara
`20260930160000_unaccent_extension.sql` y necesita el buscador. La API además
activa `pgcrypto` al arrancar.

<details>
<summary>Las 42 tablas</summary>

- **Agremiados:** `psi_users`, `psi_user_col_data`, `psi_user_post_grades`,
  `psi_user_social_networks`, `psi_user_solvency`, `psi_user_documents`,
  `psi_user_emergency_contacts`, `psi_deontologia`, `psi_observations`
- **Áreas de trabajo:** `psi_specialty_models`
- **Texto:** `text_models`
- **Noticias:** `posts`
- **Inscripciones:** `psi_inscription_requests`, `psi_inscription_documents`,
  `psi_inscription_notes`
- **Términos:** `psi_terms_acceptance`
- **Recuperación de contraseña:** `psi_password_reset_tokens`
- **Administración:** `user_admins`, `admin_permission_logs`, `app_settings`,
  `settings_audit_logs`, `api_change_logs`
- **Tickets:** `tickets`, `ticket_mensajes`, `ticket_adjuntos`, `ticket_estados`,
  `ticket_motivos`, `ticket_status_logs`
- **Notificaciones:** `notifications`, `notification_targets`,
  `notification_filters`, `notification_attachments`
- **Kanban:** `kanban_projects`, `kanban_project_members`, `kanban_columns`,
  `kanban_cards`, `kanban_card_notes`
- **Analítica:** `page_views`, `profile_views`, `search_events`,
  `login_events`, `active_sessions`

</details>

> Los nombres de tabla cambiaron desde las primeras versiones: `post_grades` →
> `psi_user_post_grades`, `psi_user_psi_specialty` y
> `psi_user_psi_specialty_relations` → `psi_specialty_models`, y
> `psi_user_audit_models` / `user_admin_audit_models` desaparecieron al
> consolidar la auditoría en `api_change_logs`.

## Workflow para añadir una migración

Los modelos de dominio viven en `internal/domain/*.model.go` (**no** en un
`models.go` único). El env `gorm` de `atlas.hcl` los usa como fuente de verdad:

```bash
# 1. Editar el modelo en internal/domain/*.model.go

# 2. Generar el .sql (compara los modelos contra el estado de ./migrations)
atlas migrate diff <nombre_de_la_migracion> --env gorm

# 3. Aplicar
atlas migrate apply --env gorm --url "postgres://postgres:postgres@localhost:5432/colpsi_db?sslmode=disable"

# 4. Ver qué se aplicó y qué falta
atlas migrate status --env gorm --url "postgres://.../colpsi_db?sslmode=disable"
```

En Docker, el servicio `migrador` del compose hace el `apply` automáticamente
al levantar la API (y la API no arranca hasta que termina con éxito).

Si editas un `.sql` a mano, recalcula el checksum o Atlas rechazará el
directorio:

```bash
atlas migrate hash --env gorm
```

## Notas y trampas

- **Los objetos escritos a mano no se regeneran.** `atlas migrate diff`
  compara los *modelos* contra las migraciones, así que no conoce los
  `CHECK`, los índices **únicos parciales** ni las columnas que se añadieron a
  mano (`reject_reason`, `last_change_reason`, `can_view_logs`…). Con frecuencia
  propone **borrarlos** como huérfanos (por eso el diff de la persona de
  contacto viene con el `DROP INDEX idx_posts_status_publish_at` descartado a
  mano).
  Revisa siempre el `.sql` generado antes de aplicarlo, y si los borra,
  re-escríbelos a mano con su `COMMENT`.
- **`atlas.sum` no se edita a mano** y **nunca se borra**: es el control de
  integridad del directorio. Si el hash de un archivo no cuadra con el registro,
  Atlas se niega a aplicar nada.
- **Las revisiones aplicadas viven en el schema `atlas_schema_revisions`**
  (Atlas crea el schema, no una tabla en `public`), con una fila por archivo y
  su hash. `atlas migrate status` la consulta: si el directorio y la tabla
  discrepan, el error típico es *checksum mismatch*. Para volver a sincronizar
  una base ya aplicada con un directorio distinto: `atlas migrate set <versión>
  --env gorm --url ...`.
- **Tras una migración que altera una tabla, reinicia `colpsi_pgbouncer`.** El
  pooler guarda planes preparados con el esquema viejo; con
  `prefer_simple_protocol=true` (gotcha 14 de `api/AGENTS.md`) eso ya no ocurre,
  pero el reinicio sigue siendo una precaución barata.
- ⚠️ **La extensión `unaccent` la crea una migración, no un modelo.** El
  buscador la usa en 18 llamadas de
  `internal/repository/postgres/psi_repository.go` (`unaccent(first_name)
  ILIKE unaccent(?)`, etc.) y sin ella responde `42883 function
  unaccent(character varying) does not exist`. Está declarada en
  `20260930160000_unaccent_extension.sql`, pero como **ningún modelo de
  dominio la declara**, `atlas migrate diff` la marca como huérfana y propondrá
  `DROP EXTENSION unaccent`: bórralo del `.sql` generado antes de aplicar. La
  extensión `pgcrypto` sí la crea la propia API al arrancar
  (`pkg/database/migration.go`), con `IF NOT EXISTS`.
- La API además ejecuta `AutoMigrate` de GORM al arrancar
  (`pkg/database/migration.go`), que es *self-healing*: crea lo que falte. Eso
  puede **tapar** una migración mal aplicada, así que no confíes en que "la API
  arrancó" como prueba de que el esquema está bien. Lo autoritativo es
  `atlas migrate status`.

**[⬆ Volver a API](../)**
