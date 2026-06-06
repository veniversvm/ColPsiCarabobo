# Proyecto API: Directorio y Gestión

Este proyecto es una API robusta construida en Go, siguiendo principios de Clean Architecture. Permite la gestión de psicólogos (psi), posts, analíticas y administración.

## 📂 Estructura de Documentación

- [📁 api/migrations](./api/README.md) - Historial de cambios de base de datos.
- [📁 cmd](./cmd/README.md) - Puntos de entrada de la aplicación.
- [📁 docs](./docs/README.md) - Documentación Swagger/OpenAPI.
- [📁 internal/config](./internal/config/README.md) - Variables de entorno.
- [📁 internal/domain](./internal/domain/README.md) - Modelos e interfaces de dominio.
- [📁 internal/handler](./internal/handler/README.md) - Controladores HTTP.
- [📁 internal/middleware](./internal/middleware/README.md) - Filtros y seguridad.
- [📁 internal/repository](./internal/repository/README.md) - Acceso a datos (Postgres).
- [📁 internal/service](./internal/service/README.md) - Lógica de negocio.
- [📁 internal/utils](./internal/utils/README.md) - Utilidades y helpers.
- [📁 pkg](./pkg/README.md) - Paquetes compartidos (DB, S3).

## 🚀 Listado Global de Funciones Principales

| Módulo             | Función / Método    | Descripción                                      |
| :------------------ | :-------------------- | :------------------------------------------------ |
| **Auth**      | `GenerateToken`     | Genera JWT para usuarios y administradores.       |
| **Auth**      | `ValidatePassword`  | Compara hashes de contraseñas.                   |
| **Psi**       | `GetPsiByID`        | Retorna el perfil detallado de un psicólogo.     |
| **Psi**       | `FilterDirectory`   | Filtra psicólogos por especialidad o ubicación. |
| **Posts**     | `CreatePost`        | Crea una nueva publicación en el blog.           |
| **Analytics** | `TrackVisit`        | Registra métricas de visualización.             |
| **Utils**     | `CleanAlphaNumeric` | Sanitiza strings para prevenir inyecciones.       |
| **S3**        | `UploadImage`       | Sube archivos multimedia a AWS S3.                |

## 🔑 Variables Globales / Configuración

- `DB_URL`: Cadena de conexión a PostgreSQL.
- `PORT`: Puerto de ejecución del servidor.
- `JWT_SECRET`: Llave para firma de tokens.
- `AWS_BUCKET`: Nombre del bucket para imágenes.
