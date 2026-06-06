# 🧩 Módulo de Dominio (Domain)

[⬅ Volver al Inicio](../../README.md)

El paquete `domain` es el corazón del sistema. Contiene las **Entidades** (estructuras de datos) y las **Interfaces** (contratos) que definen las reglas de negocio. Este módulo no tiene dependencias de ningún otro paquete interno (`service`, `handler`, `repository`), cumpliendo con la arquitectura de cebolla (Clean Architecture).

## 🗂️ Entidades y Modelos (Variables y Estructuras)

### 📋 AuditModel (`audit.model.go`)

Estructura base que se inyecta en casi todos los modelos para estandarizar la trazabilidad.

- **`CreatedAt` / `UpdatedAt`**: Timestamps automáticos de GORM.
- **`DeletedAt`**: Soporte para *Soft Delete* (los datos no se borran, se marcan como eliminados).
- **`CreateBy` / `UpdateBy`**: Nombre textual del responsable.
- **`CreateById` / `UpdateById`**: UUID del usuario responsable para integridad referencial.

### 👤 UserAdmin (`user.model.go`)

Define al personal con acceso al panel administrativo.

- **Variables Clave**:
  - `Sudo`: Booleano de superusuario (bypass de permisos).
  - `Key`: Semilla de seguridad para invalidar todas las sesiones de un admin si se ve comprometido.
  - `Permissions`: Serie de flags `CanCreate...`, `CanDelete...` que implementan un control de acceso basado en roles (RBAC).

### 🩺 PsiUserModel (`user.model.go`)

El modelo más extenso. Representa a un Psicólogo colegiado.

- **Gestión de Privacidad**: Utiliza múltiples flags `Show...` (ej. `ShowContactEmail`) para determinar qué datos se exponen en el directorio público vs. qué datos son solo para uso interno del gremio.
- **Ubicación Geográfica**: Dividido en tres zonas (Carabobo, Fuera de Carabobo, Internacional) para optimizar búsquedas regionales.
- **Vínculos**: Se relaciona con `ColData`, `PostGrades`, `SocialNetworks` y `Solvencies`.

### 🎓 PsiUserColData (`user.model.go`)

Datos estrictamente académicos y legales ante el colegio (Número de registro, tomo, folio, universidad de pregrado).

### 📜 TextModel (`text.model.go`)

Modelo diseñado para almacenar contenido extenso (HTML/Markdown) como biografías o cuerpo de posts, separado de la metadata para agilizar consultas rápidas.

---

## 🛠️ Interfaces de Repositorio (Contratos de Acción)

Las interfaces definen **qué** acciones se pueden realizar. La implementación real (el **cómo**) se encuentra en el paquete `internal/repository`.

### 1. `PsiUserRepository` (`psi_repository.go`)

| Función                     | Descripción                                                                    |
| :--------------------------- | :------------------------------------------------------------------------------ |
| `CreateWithColData`        | Operación atómica para registrar un psicólogo con todos sus datos iniciales. |
| `SearchDirectory`          | Lógica de filtrado para usuarios finales (respeta solvencia y visibilidad).    |
| `SearchAdmin`              | Vista de "Rayos X" para administradores, ignora restricciones de visibilidad.   |
| `UpdateKey`                | Rotación de credenciales para seguridad de sesión.                            |
| `CreateOrUpdateSolvencies` | Gestión masiva de estados de pago.                                             |

### 2. `PostRepository` (`post_respository.go`)

| Función                  | Descripción                                                               |
| :------------------------ | :------------------------------------------------------------------------- |
| `Create(post, content)` | Guarda metadatos y el cuerpo del texto en una sola transacción.           |
| `List`                  | Recupera posts con filtros de estado (Borrador/Publicado/Programado).      |
| `PublishScheduled`      | Función de "Cron" que activa posts cuya fecha de publicación ha llegado. |

### 3. `UserAdminRepository` (`admin_repository.go`)

| Función            | Descripción                                                                      |
| :------------------ | :-------------------------------------------------------------------------------- |
| `GetByIdentifier` | Busca por Email o Username (utilizado en el Login).                               |
| `CountSudos`      | Medida de seguridad para evitar que el sistema se quede sin superadministradores. |

### 4. `SpecialtyRepository` (`specialty_repository.go`)

| Función           | Descripción                                                                           |
| :----------------- | :------------------------------------------------------------------------------------- |
| `GetAll(status)` | Lista especialidades (clínica, organizacional, etc.) según su estado de activación. |

---

## 🔗 Navegación

- [Ir a Servicios (Lógica de Negocio) ➡](../service/README.md)
- [Ir a Repositorios (Persistencia) ➡](../repository/README.md)
