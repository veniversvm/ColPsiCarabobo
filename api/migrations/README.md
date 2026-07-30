# 🗃️ Migraciones de Base de Datos (migrations/)

> **[⬆ API](../)** — `api/migrations/`

Migraciones de base de datos gestionadas por **Atlas**.

## Archivos

| Archivo | Descripción |
|---------|-------------|
| `20260604165811_init.sql` | Esquema inicial (408 líneas, 15 tablas) |
| `intrucciones.txt` | Guía y instrucciones de migración |
| `atlas.sum` | Archivo de checksum de Atlas |

## Tablas creadas

`psi_users`, `psi_user_col_data`, `psi_user_audit_models`, `psi_user_social_networks`, `psi_user_solvency`, `psi_user_psi_specialty`, `psi_user_psi_specialty_relations`, `psi_user_deontologia`, `text_models`, `post_grades`, `user_admins`, `user_admin_audit_models`, `page_views`, `profile_views`, `search_events`, `login_events`, `active_sessions`

## Workflow de migración

1. Actualizar los modelos de dominio
2. Ejecutar `go run cmd/exp/migrate/main.go` para generar el esquema
3. Ejecutar `atlas migrate diff` para generar el SQL de migración
4. Ejecutar `atlas migrate apply` para aplicar los cambios

**[⬆ Volver a API](../)**
