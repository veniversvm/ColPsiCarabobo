# 📖 Documentación API (docs/)

Especificaciones Swagger/OpenAPI de la API.

## Archivos

| Archivo | Descripción |
|---------|-------------|
| `docs.go` | Inicialización auto-generada de Swagger |
| `swagger.json` | Spec OpenAPI 2.0 (JSON) |
| `swagger.yaml` | Spec OpenAPI 2.0 (YAML) |

## Generar documentación

```bash
swag init -g cmd/api/main.go -o docs/
```

## Acceso

Los endpoints de Swagger están disponibles en `/swagger/*` (si están habilitados).

## Grupos documentados

| Grupo | Descripción |
|-------|-------------|
| **Admin** | CRUD de administradores |
| **Auth** | Login y registro |
| **Psi** | Perfiles de psicólogos |
| **Posts** | CMS y publicaciones |
| **Specialties** | Especialidades |
| **Analytics** | Métricas y eventos |
