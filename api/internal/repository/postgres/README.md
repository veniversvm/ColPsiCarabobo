# 🗄️ Capa de Persistencia: PostgreSQL + GORM

[⬅ Volver al Inicio](../../../README.md)

Este módulo contiene las implementaciones concretas de los repositorios. Aquí se traduce la lógica de negocio a consultas SQL eficientes, transacciones seguras y gestión de relaciones complejas.

---

## 🩺 1. Repositorio de Psicólogos (`psi_repository.go`)

Maneja el núcleo de la aplicación, interactuando con al menos 6 tablas relacionadas.

### Constructores y Funciones Base

* **`NewPsiRepository(db *gorm.DB)`**: Inyecta la dependencia de la base de datos y retorna la interfaz del dominio.

### Gestión Core del Perfil

* **`CreateWithColData(ctx, psi, colData, solvencies, postgrades)`**:
  * *Qué hace:* Registra un psicólogo de forma atómica.
  * *El Porqué:* Usa una transacción (`tx`) para crear la `TextModel` (Bio vacía) -> `PsiUserModel` -> `PsiUserColData` -> `PsiUSerSolvency` -> Bulk Insert de `PsiUserPostGrade`. Si algo falla, revierte todo para evitar perfiles huérfanos.
* **`GetByID(ctx, id)`**:
  * *Qué hace:* Busca un psicólogo por UUID.
  * *Técnica:* Usa `Preload` para traer `ColData`, `PostGrades` (ordenados cronológicamente), `SocialNetworks` y `FullBio` de un solo golpe.
* **`GetByFPV(ctx, id)`**:
  * *Qué hace:* Igual que `GetByID`, pero filtra por el número de federación (FPV), útil para rutas públicas (ej. `/directorio/1234`).
* **`GetByIdentifier(ctx, identifier)`**:
  * *Qué hace:* Búsqueda para inicio de sesión.
  * *Por qué:* Busca por `username OR email`, dando flexibilidad al usuario al hacer login.
* **`Delete(ctx, id)`**:
  * *Qué hace:* Ejecuta un borrado lógico (Soft Delete) aprovechando el campo `DeletedAt` heredado de `AuditModel`.

### Mutaciones y Actualizaciones

* **`Update(ctx, psi, colData, bioText, solvencies)`**:
  * *Qué hace:* Edición global reservada para administradores.
  * *Técnica:* Usa `clause.OnConflict` (Upsert) para procesar las solvencias; si existe una solvencia con la misma fecha, la actualiza, si no, la crea.
* **`UpdatePublicProfile(ctx, psi, colData, bioText)`**:
  * *Qué hace:* Actualización hecha por el propio psicólogo.
  * *Seguridad (El Porqué):* Usa `Omit("ci", "fpv", "is_active", "solvent")` para impedir que un usuario malintencionado cambie sus datos legales o de pago. Además, usa `gorm.Expr("?", valor)` para forzar a GORM a guardar booleanos falsos.
* **`UpdateKey(ctx, psi)`**:
  * *Qué hace:* Cambia la semilla de firma (Key) del JWT. Se usa un `Select` estricto para no tocar otros campos del modelo.
* **`GetPsiUserColData(ctx, psiID)`**: Recuperación ligera solo de la información académica y de solvencia.

### Búsqueda y Filtros

* **`SearchDirectory(ctx, filter)`**:
  * *Qué hace:* Motor de búsqueda público.
  * *Técnica:* Busca por identidad dividiendo en palabras cruzadas. Aplica cláusulas `ILIKE unaccent(?)` para ignorar tildes. Respeta siempre los escudos de privacidad (ej. oculta el municipio si `ShowMunicipalityCarabobo == false`).
* **`SearchAdmin(ctx, filter)`**:
  * *Qué hace:* Listado para el panel de control.
  * *Diferencia:* Ignora totalmente las banderas de privacidad y de solvencia.
* **`Count(ctx, active)`**: Cuenta psicólogos usando lógica tri-estatal (`*bool`).
* **`Search(ctx, filters, page, limit)`**: Búsqueda genérica por mapa dinámico (usada internamente).

### Sub-módulos: Postgrados, Redes y Solvencias

* **`CreatePostGrade`**, **`GetPostGradeByID`**, **`UpdatePostGrade`**: CRUD transaccional simple para estudios adicionales.
* **`CreateSolvency`**, **`GetSolvencies`**, **`CreateOrUpdateSolvencies`**: Manejo del estado de pagos del agremiado.
* **`CreateSocialNetwork`**, **`GetSocialNetworkByID`**, **`UpdateSocialNetwork`**, **`DeleteSocialNetwork`**: CRUD de redes sociales.
* **`CountSocialNetworksByPsiID`**: Cuenta cuántas redes tiene un Psi (para validación de límites).
* **`GetTextContentByID(ctx, id)`**: Extrae el string HTML de una biografía sin cargar la metadata.
* **`ValidateUniqueCredentials(ctx, username, email, excludeID)`**: Evita choques de credenciales en actualizaciones.
* **`GetSitemapData(ctx)`**: Extrae UUIDs y fechas para el crawler de Google.

---

## 📰 2. Repositorio de Publicaciones (`post_repo.go`)

Gestión del módulo de noticias, aplicando separación de recursos pesados.

* **`NewPostRepository(db)`**: Constructor.
* **`Create(ctx, post, text)`**:
  * *Qué hace:* Crea una noticia.
  * *Por qué:* Usa transacción para crear primero el contenido HTML (`TextModel`), obtiene el ID y se lo asigna a la metadata (`Post`).
* **`GetByID(ctx, id)`**: Extrae noticia usando `Preload("Text")`.
* **`List(ctx, filter, page, limit)`**:
  * *Qué hace:* Listado paginado.
  * *Rendimiento:* Deliberadamente **OMITE** cargar el `TextModel` para ahorrar memoria. Controla visibilidad usando `filter.Type` (`public`, `psi`, `all_visible`).
* **`Update(ctx, post, text)`**: Actualiza la noticia. Si el parámetro `text` es nulo, solo guarda la metadata (título, foto).
* **`Delete(ctx, id)`**: Soft delete automático de GORM.
* **`PublishScheduled(ctx)`**: Busca posts con estado `scheduled` cuya fecha límite pasó y los pasa a `published`.
* **`GetSitemapPosts(ctx)`**: Query ultraligera (`id`, `title`, `updated_at`) para archivos XML.

---

## 🏷️ 3. Repositorio de Especialidades (`specialty_repo.go`)

Manejo del catálogo de áreas de trabajo (tags).

* **`NewSpecialtyRepository(db)`**: Constructor.
* **`Create(ctx, s)`**: Añade una nueva especialidad al catálogo.
* **`GetAll(ctx, status)`**: Devuelve la lista ordenada alfabéticamente filtrando por estado (`active`, `inactive`, `all`).
* **`GetByID(ctx, id, active)`**:
  * *Qué hace:* Busca una especialidad.
  * *Seguridad:* Si el flag `active` es `true`, bloquea el acceso si la especialidad está dada de baja.
* **`GetByAdminID(ctx, id)`**: Versión privilegiada que ignora si está activa o inactiva.
* **`Update(ctx, s)`**: Sobreescribe el modelo completo (`Save()`).
* **`Delete(ctx, id)`**:
  * *Qué hace:* Desactiva la especialidad.
  * *Técnica:* No hace un DELETE de SQL (para no romper el perfil de los psicólogos). Simplemente hace `Updates(map{"active": false})`.
* **`Count(ctx, active)`**: Conteo tri-estatal (`*bool`) de la tabla.
* **`GetAllAdmin(ctx)`**: Listado sin filtros de visibilidad.

---

## 👮 4. Repositorio de Administradores (`user_admin_repo.go`)

Gestión del staff interno del colegio.

* **`NewAdminRepository(db)`**: Constructor.
* **`GetByIdentifier(ctx, identifier)`**:
  * *Qué hace:* Busca un admin.
  * *Técnica:* Realiza la consulta por la columna `email` o `username`.
* **`GetByID(ctx, id)`**: Recuperación estándar por UUID.
* **`Create(ctx, user)`**: Registra un nuevo administrador.
* **`Update(ctx, user)`**: Guarda cambios en el modelo.
* **`Delete(ctx, id)`**: Soft Delete automático heredado.
* **`CountSudos(ctx)`**:
  * *Qué hace:* Valida el estado del sistema.
  * *Seguridad:* Cuenta cuántos Superusuarios quedan. Evita que el sistema permita borrar a todos los admins, bloqueando el acceso permanentemente.
* **`List(ctx, active, search, page, limit)`**: Buscador paginado que aplica `ILIKE` para búsqueda parcial.

---
