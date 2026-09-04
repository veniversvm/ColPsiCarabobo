# Plan: Sistema de Gestión de Ingreso de Profesionales

**Estado:** Documento de planificación. No aplicar todavía.

**Fecha:** 2026-09-03

---

## 1. Contexto y problema

El Colegio de Psicólogos del Estado Carabobo actualmente gestiona el ingreso de nuevos
profesionales de manera completamente manual: el interesado envía documentos por correo
electrónico, el personal administrativo verifica físicamente los recaudos en una cita
presencial, y finalmente un admin crea el registro en el sistema importando desde un Excel
maestro o creándolo manualmente en `/admin/psicologos/crear`.

**Problemas identificados:**

- No hay control digital del flujo de inscripciones — las solicitudes llegan por correo y se
  pierden entre emails.
- No se valida la unicidad de cédula ni FPV antes de la cita presencial.
- No existe un estado "pendiente" — un psicólogo se crea directamente como activo.
- El número de control solo se asigna desde el Excel, no durante la creación manual.
- No hay trazabilidad del proceso de verificación de requisitos.

**Objetivo:** Implementar un sistema de pre-inscripción digital que permita:
1. Recoger datos del profesional interesado vía formulario público.
2. Validar la unicidad de cédula y FPV en tiempo real.
3. Mantener las solicitudes en estado "pendiente" hasta verificación admin.
4. Asignar número de control secuencial al aprobar.
5. Crear el registro del psicólogo con `is_active: false` hasta confirmar requisitos.
6. Activar manualmente solo cuando se cumplan los 3 requisitos legales.

---

## 2. Marco regulatorio

Extraído de los documentos en `Docs/analisis/`:

### 2.1 Requisitos para ejercer (Art. 5, Ley de Ejercicio de la Psicología)

Para poder ejercer la Psicología, los titulares deben cumplir:
1. Inscribirse en el Ministerio de Educación.
2. Estar inscritos en el Colegio de Psicólogos de la jurisdicción.
3. Ser miembro del Instituto de Previsión del Psicólogo (INPREPSI).

### 2.2 Condiciones de admisión (Art. 14-17, Estatutos FPV)

- Ser profesional de la Psicología con Título de Licenciado o Psicólogo de universidad
  reconocida por el CNU, o título extranjero revalidado.
- El Colegio tramita el registro ante la FPV, que asigna el **N° FPV**.
- Solo con N° FPV se puede ejercer plenamente (Art. 16).

### 2.3 Miembro activo (Art. 18, Estatutos FPV)

Se considera Psicólogo Agremiado Activo aquel que:
- Esté registrado en la FPV.
- Esté inscrito en el Ministerio de Educación.
- Tenga asignado el N° de FPV.
- Esté inscrito y solvente en el Colegio de Psicólogos.
- Esté inscrito y solvente en el INPREPSI.
- No esté suspendido por un Tribunal Disciplinario.

### 2.4 Inscripción en el Colegio (Art. 15-16, Reglamento Interno Carabobo)

- Formulario de solicitud con todos los datos.
- Presentar título original de Licenciado en Psicología para cotejo y certificación.
- El Presidente o Secretario del Colegio certifican la copia del título.

### 2.5 Proceso de inscripción vigente (5 pasos)

1. **Determinación de deuda:** $30 inscripción (pago único) + $40/año desde 2024.
   Conversión a bolívares con tasa BCV del día.
2. **Pago bancario:** Transferencia a BancoProvincial, cuenta corriente
   0108-0558-94-0100208134, RIF J-508172418.
3. **Postulación digital:** Envío por correo a admon.colpsicarabobo@gmail.com de:
   foto tipo carnet, título, cédula, RIF (con dirección en Carabobo), comprobante de pago.
4. **Confirmación y espera:** 5 días hábiles para verificación técnica y asignación de cita.
5. **Cita presencial:** Entrega de 2 expedientes (FPV + archivo regional), verificación de
   título original, registro fotográfico de seguridad.

---

## 3. Estado actual del sistema

### 3.1 Modelo de datos existente (`psi_users`)

| Campo | Tipo | Restricción | Notas |
|-------|------|-------------|-------|
| `id` | UUID | PK | uuidv7 |
| `ci` | BIGINT | NOT NULL, UNIQUE INDEX | Cédula de identidad |
| `fpv` | BIGINT | NOT NULL, UNIQUE INDEX | N° FPV |
| `control_number` | VARCHAR(50) | UNIQUE INDEX (parcial, solo no vacío) | Viene del Excel |
| `is_active` | BOOLEAN | DEFAULT true | Control de acceso |
| `solvent` | BOOLEAN | DEFAULT false | Solvencia financiera |
| `proof_of_life` | BOOLEAN | | Fe de vida |
| `username` | VARCHAR(25) | UNIQUE, NOT NULL | |
| `email` | VARCHAR(255) | UNIQUE, NOT NULL | |
| `password` | VARCHAR(512) | | Bcrypt hash |

### 3.2 Tabla relacionada `psi_user_col_data`

| Campo | Tipo | Notas |
|-------|------|-------|
| `guild_inscription_date` | DATE | Fecha de colegiatura |
| `university_undergraduate` | VARCHAR(255) | Universidad |
| `graduate_date` | DATE | Fecha de graduación |
| `mention_undergraduate` | VARCHAR(255) | Mención |
| `register_number` | VARCHAR(100) | N° registro del título |
| `register_title_state` | VARCHAR(100) | Estado del registro |
| `register_folio` | VARCHAR(100) | Folio |
| `register_tome` | VARCHAR(100) | Tomo |

### 3.3 Cómo se crea un psicólogo actualmente

- **Admin create** (`POST /admin/psi/create`): JSON con todos los campos. `is_active`
  default `true`. Sin número de control. Sin subida de archivos.
- **Excel import** (`POST /admin/psi/upload-csv`): Bulk desde Excel. Columna 0 = Nº control.
  `is_active: true` para todos.

### 3.4 Envío de email de bienvenida

- Template: `welcome_psi.html` (HTML con branding COLPSI).
- Datos: `Name`, `Email`, `Password`.
- Subject: "Bienvenido(a) a la plataforma COLPSI Carabobo".
- Se envía desde `psi_user_admin_service.go:141-150` (create) y desde los importadores.

### 3.5 Mecanismo de subida de archivos

- **Frontend → API → S3** vía `multipart/form-data` (no presigned URLs).
- API parsea con `c.FormFile()`, sanitiza con `utils.SanitizeImage/SanitizeDocument`,
  sube con `s3Client.UploadFile()`.
- Solo se persiste la S3 key en la DB.

### 3.6 Número de control

- Viene de la columna 0 ("Nº") del Excel maestro.
- Se almacena como string tal cual (tras `cleanDash` que elimina "-" y "0").
- **No se asigna** durante creación manual desde admin.

---

## 4. Decisiones de diseño

| # | Decisión | Resolución |
|---|----------|------------|
| 1 | Acceso al formulario | **Público**, sin login |
| 2 | Subida de archivos | **Directo al bucket S3** vía API (multipart) |
| 3 | Formato Nº de control | **Secuencial numérico**: MAX(control_number existente) + 1 |
| 4 | Activación del psicólogo | **Manual**: admin con checklist de 3 requisitos |
| 5 | Email al aprobar | **Sí**: con credenciales (usuario + contraseña temporal) |
| 6 | Solicitud rechazada | **Eliminar** registro + archivos S3 |
| 7 | Duplicados de CI | **Prohibidos** entre solicitudes pending y en `psi_users` |

---

## 5. Arquitectura propuesta

### 5.1 Diagrama de flujo

```
INTERESADO                    API                         ADMIN
    │                          │                           │
    │  GET /inscripcion        │                           │
    │─────────────────────────>│                           │
    │<─── formulario HTML      │                           │
    │                          │                           │
    │  POST /inscripcion/      │                           │
    │  check-ci?ci=N           │                           │
    │─────────────────────────>│                           │
    │<─── { exists: false }    │                           │
    │                          │                           │
    │  POST /inscripcion/      │                           │
    │  submit (FormData)       │                           │
    │─────────────────────────>│                           │
    │                          │  INSERT inscription_req   │
    │                          │  status = 'pending'       │
    │<─── 201 Created          │                           │
    │                          │                           │
    │                          │  GET /admin/inscripciones │
    │                          │  /list?status=pending     │
    │                          │<──────────────────────────│
    │                          │──── lista de solicitudes ─>│
    │                          │                           │
    │                          │  GET /admin/inscripciones │
    │                          │  /:id                     │
    │                          │<──────────────────────────│
    │                          │──── detalle + archivos ──>│
    │                          │                           │
    │                          │  POST /admin/inscripciones│
    │                          │  /:id/approve             │
    │                          │<──────────────────────────│
    │                          │  1. Gen Nº control        │
    │                          │  2. Crear psi_users       │
    │                          │     (is_active: false)    │
    │                          │  3. Mover archivos S3     │
    │                          │  4. Enviar email          │
    │                          │                           │
    │<──── email con creds ────│                           │
    │                          │                           │
    │                          │  POST /admin/psi/:id      │
    │                          │  (activate)               │
    │                          │<──────────────────────────│
    │                          │  Checklist:               │
    │                          │  ☑ Ministerio             │
    │                          │  ☑ FPV                    │
    │                          │  ☑ Solvente               │
    │                          │  → is_active = true       │
```

### 5.2 Nuevo modelo de datos: `psi_inscription_requests`

```
Tabla: psi_inscription_requests
─────────────────────────────────────────────────────────
id                      UUID PK (uuidv7)
cedula                  BIGINT NOT NULL
nacionalidad            VARCHAR(1) NOT NULL (V/E)
nombres                 VARCHAR(255) NOT NULL
apellidos               VARCHAR(255) NOT NULL
fpv                     BIGINT (nullable)
telefono                VARCHAR(50)
correo                  VARCHAR(255) NOT NULL
fecha_nacimiento        DATE
titulo_universidad      VARCHAR(255)
titulo_fecha_graduacion DATE
titulo_mencion          VARCHAR(255)
titulo_registro_numero  VARCHAR(100)
titulo_registro_estado  VARCHAR(100)
rif                     VARCHAR(50)
foto_s3_key             VARCHAR(512)
comprobante_s3_key      VARCHAR(512)
status                  VARCHAR(20) DEFAULT 'pending'
control_number          VARCHAR(50)
notes                   TEXT
created_at              TIMESTAMPTZ
updated_at              TIMESTAMPTZ

Restricciones:
  UNIQUE INDEX en (cedula) WHERE status = 'pending'
  UNIQUE INDEX en (control_number) WHERE control_number <> ''
  CHECK (status IN ('pending', 'approved', 'rejected'))
```

### 5.3 Campo nuevo en `psi_user_col_data`

```
ministry_registration_confirmed BOOLEAN DEFAULT false
```

### 5.4 Cambio en `psi_users`

```
is_active: DEFAULT true → DEFAULT false
```

Esto afecta:
- Creación manual desde admin (`CreatePsiByAdmin`)
- Creación desde inscripción aprobada
- **No afecta** la importación Excel (que setea `is_active: true` explícitamente)

---

## 6. API backend (Go)

### 6.1 Nuevos archivos

| Archivo | Responsabilidad |
|---------|-----------------|
| `api/internal/domain/inscription_request_model.go` | Modelo de dominio + DTOs |
| `api/internal/repository/postgres/inscription_repository.go` | Queries contra `psi_inscription_requests` |
| `api/internal/service/inscription_service.go` | Lógica de negocio (validar, crear, aprobar, rechazar) |
| `api/internal/handler/inscription_handler.go` | Handlers HTTP |
| `api/internal/router/inscription_router.go` | Registro de rutas |
| `api/migrations/XXXX_add_inscription_requests.sql` | Migración de tabla |
| `api/migrations/XXXX_add_ministry_confirmed.sql` | Campo nuevo en `psi_user_col_data` |

### 6.2 Archivos a modificar

| Archivo | Cambio |
|---------|--------|
| `api/internal/router/router.go` | Registrar `SetupInscriptionRoutes()` |
| `api/cmd/api/main.go` | Inyectar dependencias del nouveau service |
| `api/internal/domain/user.model.go` | `is_active` default: `true` → `false` |
| `api/internal/domain/psi_user_col_data.go` | Agregar `MinistryRegistrationConfirmed` |
| `api/internal/service/psi_user_admin_service.go` | Validación de activación |
| `api/internal/service/psi_user_admin_service.go` | Creación desde inscripción |

### 6.3 Endpoints

#### Públicos (sin auth)

| Método | Ruta | Descripción |
|--------|------|-------------|
| `GET` | `/api/v1/inscripcion/check-ci` | Validar unicidad de cédula |
| `GET` | `/api/v1/inscripcion/check-fpv` | Validar unicidad de FPV |
| `POST` | `/api/v1/inscripcion/submit` | Enviar solicitud de pre-inscripción |

#### Admin (con `ProtectedAdmin404()`)

| Método | Ruta | Descripción |
|--------|------|-------------|
| `GET` | `/api/v1/admin/inscripciones/list` | Listar solicitudes con filtros |
| `GET` | `/api/v1/admin/inscripciones/:id` | Detalle de solicitud |
| `POST` | `/api/v1/admin/inscripciones/:id/approve` | Aprobar y crear psicólogo |
| `DELETE` | `/api/v1/admin/inscripciones/:id` | Rechazar y eliminar solicitud |

### 6.4 Detalle de implementación por endpoint

#### `GET /api/v1/inscripcion/check-ci`

Query params: `ci` (int, requerido)

Lógica:
1. Buscar en `psi_users` WHERE `ci = ?` → si existe, `{ exists: true, message: "La cédula ya está registrada en el sistema" }`
2. Buscar en `psi_inscription_requests` WHERE `cedula = ? AND status = 'pending'` → si existe, `{ exists: true, message: "Ya existe una solicitud activa con esta cédula" }`
3. Si no existe en ninguno: `{ exists: false }`

#### `GET /api/v1/inscripcion/check-fpv`

Query params: `fpv` (int, requerido)

Misma lógica que check-ci pero contra el campo `fpv`.

#### `POST /api/v1/inscripcion/submit`

Body: `multipart/form-data`

Campos del formulario:
- `cedula` (int, requerido)
- `nacionalidad` (string: V/E, requerido)
- `nombres` (string, requerido)
- `apellidos` (string, requerido)
- `fpv` (int, opcional)
- `telefono` (string, requerido)
- `correo` (string, requerido, email válido)
- `fecha_nacimiento` (date, requerido)
- `titulo_universidad` (string, requerido)
- `titulo_fecha_graduacion` (date, requerido)
- `titulo_mencion` (string)
- `titulo_registro_numero` (string)
- `titulo_registro_estado` (string)
- `rif` (string, requerido)
- `foto` (file, imagen, max 5MB, requerido)
- `comprobante` (file, imagen/PDF, max 5MB, requerido)

Lógica:
1. Sanitizar todos los strings con bluemonday (misma política que el resto del API)
2. Validar unicidad de cédula (misma lógica que check-ci)
3. Validar unicidad de FPV si se proporciona
4. Subir foto → `s3Client.UploadFile(ctx, foto, "inscripciones/fotos")`
5. Subir comprobante → `s3Client.UploadFile(ctx, comprobante, "inscripciones/comprobantes")`
6. Insertar en `psi_inscription_requests` con `status = 'pending'`
7. Responder 201 Created

Error handling:
- Si la cédula ya existe → 409 Conflict con mensaje descriptivo
- Si el FPV ya existe → 409 Conflict
- Si falla la subida a S3 → 500 (los archivos se limpiarían con rollback)

#### `GET /api/v1/admin/inscripciones/list`

Query params:
- `status` (string: pending|approved|rejected, default: pending)
- `q` (string, búsqueda por nombre o cédula)
- `page` (int, default 1)
- `limit` (int, default 20)

Retorna lista paginada con los campos del modelo.

#### `GET /api/v1/admin/inscripciones/:id`

Retorna el registro completo con URLs públicas de los archivos S3
(usando `s3Client.GetPublicURL(key)`).

#### `POST /api/v1/admin/inscripciones/:id/approve`

Lógica:
1. Buscar solicitud, verificar `status = 'pending'`
2. Generar número de control:
   ```sql
   SELECT COALESCE(MAX(CAST(control_number AS INTEGER)), 0) + 1
   FROM psi_users
   WHERE control_number ~ '^\d+$';
   ```
   Resultado → string (ej: "451")
3. Generar username: transliterar nombre + apellido (snake_case, sin tildes, sin espacios)
4. Generar contraseña temporal: 12 caracteres aleatorios (letras + números)
5. Hashear contraseña con bcrypt
6. Crear registro en `psi_users`:
   - `is_active: false`
   - `solvent: false`
   - `control_number`: el generado
   - `username`: generado
   - `email`: correo de la solicitud
   - `password`: hash
   - Todos los campos personales de la solicitud
7. Crear `psi_user_col_data`:
   - `guild_inscription_date`: fecha actual
   - Datos académicos de la solicitud
8. Crear bio vacía (patrón existente)
9. Mover archivos S3:
   - `inscripciones/fotos/{key}` → `avatars/{psiID}.webp` (sanitizado)
   - `inscripciones/comprobantes/{key}` → `titles/{psiID}_comprobante.webp`
10. Actualizar solicitud: `status = 'approved'`, `control_number = el generado`
11. Enviar email `welcome_psi` con:
    - `Name`: nombres
    - `Email`: correo
    - `Password`: contraseña temporal en claro
12. Responder 200 con los datos del psicólogo creado

#### `DELETE /api/v1/admin/inscripciones/:id`

Lógica:
1. Buscar solicitud
2. Eliminar archivos del S3 (foto + comprobante)
3. Eliminar registro de `psi_inscription_requests`
4. Responder 204 No Content

### 6.5 Validación de activación en `psi_user_admin_service.go`

En `UpdatePsiByAdmin`, antes de procesar el update:

```go
// Si se intenta activar (is_active: false → true), verificar requisitos
if req.IsActive != nil && *req.IsActive && !existingUser.IsActive {
    colData, _ := repo.GetColData(existingUser.ID)
    if colData == nil || !colData.MinistryRegistrationConfirmed {
        return ApiError{Status: 400, Message: "Debe confirmar la inscripción en el Ministerio"}
    }
    if existingUser.FPV == 0 {
        return ApiError{Status: 400, Message: "Debe tener un N° de FPV asignado"}
    }
    if !existingUser.Solvent {
        return ApiError{Status: 400, Message: "El psicólogo debe estar solvente para activarse"}
    }
}
```

---

## 7. Frontend (SolidStart)

### 7.1 Nuevos archivos

| Archivo | Responsabilidad |
|---------|-----------------|
| `web/src/types/inscription.ts` | Tipos TypeScript para inscripción |
| `web/src/components/inscripcion/CheckField.tsx` | Campo con validación async (CI/FPV) |
| `web/src/components/inscripcion/FileUpload.tsx` | Subida de archivos con preview |
| `web/src/components/inscripcion/SuccessMessage.tsx` | Mensaje post-submit |
| `web/src/routes/admin/inscripciones/index.tsx` | Listado admin |
| `web/src/routes/admin/inscripciones/[id].tsx` | Detalle admin |
| `web/src/components/admin/inscripciones/InscriptionRow.tsx` | Fila de la tabla |
| `web/src/components/admin/inscripciones/InscriptionDetail.tsx` | Vista de detalle |
| `web/src/components/admin/psicologos/edit/sections/ActivationSection.tsx` | Checklist de activación |

### 7.2 Archivos a modificar

| Archivo | Cambio |
|---------|--------|
| `web/src/routes/inscripcion.tsx` | Reescribir: de estático a formulario interactivo |
| `web/src/types/admin.ts` | Agregar tipos de inscripción |
| `web/src/routes/admin/psicologos/[id]/detalle.tsx` | Integrar ActivationSection |

### 7.3 Página de inscripción (`/inscripcion`)

#### Estructura del formulario

```
┌─────────────────────────────────────────────────┐
│  Procedimiento de Inscripción (resumen)         │
│  ─────────────────────────────────────────────  │
│  [Sección 1: Datos Personales]                  │
│  Cédula: [________] ✓/✗ (validación async)      │
│  Nacionalidad: [V ▾] [E]                         │
│  Nombres: [________]                             │
│  Apellidos: [________]                           │
│  FPV: [________] (opcional, validación async)   │
│  Teléfono: [________]                            │
│  Correo: [________]                              │
│  Fecha de nacimiento: [________]                 │
│  ─────────────────────────────────────────────  │
│  [Sección 2: Datos Académicos]                  │
│  Universidad: [________]                         │
│  Fecha de graduación: [________]                 │
│  Mención: [________]                             │
│  N° Registro del título: [________]             │
│  Estado del registro: [________]                 │
│  RIF: [________]                                 │
│  ─────────────────────────────────────────────  │
│  [Sección 3: Documentos]                        │
│  Foto tipo carnet: [arrastrar o seleccionar]     │
│  Comprobante de pago: [arrastrar o seleccionar]  │
│  ─────────────────────────────────────────────  │
│  [Enviar solicitud]                              │
│  ─────────────────────────────────────────────  │
│  Tras enviar exitosamente:                       │
│  ✅ Solicitud recibida correctamente.           │
│  Recibirás un correo con tus credenciales       │
│  en un plazo de 5 días hábiles.                 │
└─────────────────────────────────────────────────┘
```

#### Componente `CheckField.tsx`

Campo de texto que al perder foco (o tras 500ms de debounce) consulta la API:

```tsx
// Ejemplo de uso
<CheckField
  label="Cédula de Identidad"
  endpoint="/api/v1/inscripcion/check-ci"
  param="ci"
  onValid={(value) => setCedula(value)}
  onInvalid={(message) => setError(message)}
/>
```

Estados visuales:
- **Normal:** borde gris
- **Validando:** borde azul con spinner
- **Válido:** borde verde con check ✓
- **Inválido:** borde rojo con mensaje de error

#### Validación del formulario

- Todos los campos requeridos marcados con asterisco (*)
- Validación client-side antes del submit:
  - Cédula: numérico, mínimo 6 dígitos
  - Email: formato válido
  - Archivos: tipo MIME correcto (imagen para foto, imagen/PDF para comprobante), max 5MB
  - Si la validación async de CI/FPV falló, no permitir submit

#### Comportamiento post-submit

- Botón "Enviar solicitud" se deshabilita durante el submit
- Spinner en el botón
- Éxito: formulario se reemplaza por `SuccessMessage`
- Error 409 (CI duplicada): mostrar mensaje específico
- Error de red: mostrar "Error de conexión, intente nuevamente"

### 7.4 Panel admin de inscripciones

#### Listado (`/admin/inscripciones/index.tsx`)

Filtros:
- Estado: Todos | Pendientes | Aprobadas | Rechazadas (tabs o dropdown)
- Búsqueda: por nombre o cédula (input con debounce)

Tabla:

| Cédula | Nombre completo | FPV | Fecha solicitud | Acciones |
|--------|----------------|-----|-----------------|----------|
| 12345678 | María García | 5678 | 03/09/2026 | Ver |
| 23456789 | Juan López | — | 02/09/2026 | Ver |

Cada fila tiene un botón "Ver" que lleva al detalle.

Paginación al pie.

#### Detalle (`/admin/inscripciones/[id].tsx`)

Layout de 2 columnas:

**Columna izquierda:**
- Datos personales (nombre, cédula, FPV, teléfono, email, fecha nacimiento, RIF)
- Datos académicos (universidad, graduación, mención, registro)
- N° de control asignado (si está aprobado)

**Columna derecha:**
- Foto del solicitante (visualizada con `bucketUrl` o URL pública del S3)
- Comprobante de pago (visualizado)

**Acciones:**
- Botón verde "Aprobar inscripción" → abre modal de confirmación:
  "Se creará la cuenta del psicólogo con is_active=false. Se enviará un correo
  con las credenciales. ¿Confirmar?"
- Botón rojo "Rechazar solicitud" → abre modal de confirmación:
  "Se eliminará permanentemente esta solicitud y los archivos adjuntos.
  ¿Está seguro?"
- Botón "Volver" → lista

### 7.5 ActivationSection en admin edit de psicólogo

Se agrega una nueva sección en la página de detalle/edición del psicólogo
(`/admin/psicologos/[id]/detalle.tsx`), visible solo cuando `is_active = false`:

```
┌─────────────────────────────────────────────────┐
│  Activación de cuenta                           │
│  ─────────────────────────────────────────────  │
│  Para activar esta cuenta, confirme los         │
│  siguientes requisitos:                         │
│                                                 │
│  ☐ Inscripción en Ministerio confirmada         │
│    [toggle switch]                              │
│                                                 │
│  ☐ N° FPV asignado: #5678                      │
│    [✅ ya cumplido - readonly]                  │
│                                                 │
│  ☐ Estado solvente                              │
│    [toggle switch]                              │
│                                                 │
│  ─────────────────────────────────────────────  │
│  [Activar cuenta] ← solo habilitado si los 3   │
│  están marcados                                 │
└─────────────────────────────────────────────────┘
```

Cuando `is_active = true`, la sección muestra:
```
┌─────────────────────────────────────────────────┐
│  Estado de cuenta: ✅ Activa                    │
│  Activada el: 15/09/2026                        │
│  Requisitos verificados: Ministerio ✓ FPV ✓     │
│  Solvente ✓                                     │
└─────────────────────────────────────────────────┘
```

---

## 8. Migraciones de base de datos

### 8.1 `XXXX_add_inscription_requests.sql`

```sql
-- Crear tabla de solicitudes de inscripción
CREATE TABLE "psi_inscription_requests" (
    "id" uuid NOT NULL DEFAULT uuidv7(),
    "cedula" bigint NOT NULL,
    "nacionalidad" character varying(1) NOT NULL,
    "nombres" character varying(255) NOT NULL,
    "apellidos" character varying(255) NOT NULL,
    "fpv" bigint NULL,
    "telefono" character varying(50) NULL,
    "correo" character varying(255) NOT NULL,
    "fecha_nacimiento" date NULL,
    "titulo_universidad" character varying(255) NULL,
    "titulo_fecha_graduacion" date NULL,
    "titulo_mencion" character varying(255) NULL,
    "titulo_registro_numero" character varying(100) NULL,
    "titulo_registro_estado" character varying(100) NULL,
    "rif" character varying(50) NULL,
    "foto_s3_key" character varying(512) NULL,
    "comprobante_s3_key" character varying(512) NULL,
    "status" character varying(20) NULL DEFAULT 'pending',
    "control_number" character varying(50) NULL,
    "notes" text NULL,
    "created_at" timestamptz NULL,
    "updated_at" timestamptz NULL,
    PRIMARY KEY ("id")
);

-- Unicidad de cédula solo entre solicitudes pendientes
CREATE UNIQUE INDEX "idx_inscription_requests_cedula_pending"
    ON "psi_inscription_requests" ("cedula")
    WHERE "status" = 'pending';

-- Unicidad de número de control (cuando no esté vacío)
CREATE UNIQUE INDEX "idx_inscription_requests_control_number"
    ON "psi_inscription_requests" ("control_number")
    WHERE "control_number" <> '' AND "control_number" IS NOT NULL;

-- Índice para búsquedas por status
CREATE INDEX "idx_inscription_requests_status"
    ON "psi_inscription_requests" ("status");
```

### 8.2 `XXXX_add_ministry_confirmed.sql`

```sql
-- Agregar campo de confirmación de inscripción en Ministerio
ALTER TABLE "psi_user_col_data"
    ADD COLUMN "ministry_registration_confirmed" boolean NULL DEFAULT false;
```

---

## 9. Seguridad

### 9.1 Validaciones

- **Cédula/FPV:** validación en tiempo real contra la DB (endpoint check-*)
- **Archivos:** validación MIME + tamaño en el handler Go (misma práctica que `ImportXlsxModal`)
- **Strings:** sanitización con bluemonday UCG policy (patrón existente)
- **Email:** validación de formato
- **Idempotency:** no se aplica en submit público (cada submit es único por CI)

### 9.2 Protección de archivos

- Los archivos de solicitudes pendientes están en `inscripciones/` (prefijo separado)
- Solo accesibles vía URL pública (mismo patrón que el resto del bucket)
- Al aprobar, se mueven a `avatars/` y `titles/` (sanitizados con `SanitizeImage`)
- Al rechazar, se eliminan con `s3Client.DeleteFile()`

### 9.3 Prevención de abuso

- Rate limiting en `submit` (configurable, sugerido: 5 intentos/minuto por IP)
- No hay autenticación → el rate limit es la única protección
- Validación estricta de tipos MIME (solo imagen para foto, imagen/PDF para comprobante)
- Tamaño máximo de archivo: 5MB

---

## 10. Orden de implementación

| Paso | Capa | Descripción | Archivos |
|------|------|-------------|----------|
| 1 | DB | Migración: tabla `psi_inscription_requests` | `migrations/XXXX_add_inscription_requests.sql` |
| 2 | DB | Migración: campo `ministry_registration_confirmed` | `migrations/XXXX_add_ministry_confirmed.sql` |
| 3 | API | Modelo de dominio + DTOs | `domain/inscription_request_model.go` |
| 4 | API | Repository (CRUD inscripciones) | `repository/postgres/inscription_repository.go` |
| 5 | API | Service (lógica de negocio) | `service/inscription_service.go` |
| 6 | API | Handler (endpoints HTTP) | `handler/inscription_handler.go` |
| 7 | API | Router (registro de rutas) | `router/inscription_router.go` |
| 8 | API | Wiring (router.go + main.go) | `router/router.go`, `cmd/api/main.go` |
| 9 | API | Validación de activación | `service/psi_user_admin_service.go` |
| 10 | API | Default `is_active: false` | `domain/user.model.go` |
| 11 | API | Campo en modelo | `domain/psi_user_col_data.go` |
| 12 | FE | Tipos TypeScript | `types/inscription.ts` |
| 13 | FE | Componentes reutilizables | `components/inscripcion/` |
| 14 | FE | Página `/inscripcion` | `routes/inscripcion.tsx` |
| 15 | FE | Admin listado inscripciones | `routes/admin/inscripciones/index.tsx` |
| 16 | FE | Admin detalle inscripción | `routes/admin/inscripciones/[id].tsx` |
| 17 | FE | Admin ActivationSection | `components/admin/psicologos/edit/sections/ActivationSection.tsx` |
| 18 | FE | Integrar ActivationSection | `routes/admin/psicologos/[id]/detalle.tsx` |
| 19 | VER | Build completo | `npm run build` + `go build ./...` |

---

## 11. Verificación

### 11.1 Checklist de pruebas

- [ ] Formulario público carga sin errores
- [ ] Validación async de cédula funciona (reject duplicada)
- [ ] Validación async de FPV funciona (reject duplicada)
- [ ] Submit exitoso crea solicitud con status=pending
- [ ] Archivos se suben al bucket correctamente
- [ ] Admin puede listar solicitudes pendientes
- [ ] Admin puede ver detalle con archivos
- [ ] Admin puede aprobar → se crea psicólogo con `is_active=false`
- [ ] Email de bienvenida se envía con credenciales correctas
- [ ] Nº de control se asigna secuencialmente
- [ ] Admin puede rechazar → solicitud + archivos se eliminan
- [ ] CI duplicada en solicitud pending retorna 409
- [ ] Admin NO puede activar psicólogo sin los 3 requisitos
- [ ] Admin SÍ puede activar con los 3 requisitos completados
- [ ] `npm run build` compila sin errores
- [ ] `go build ./...` compila sin errores

### 11.2 Comandos de verificación

```bash
# Frontend
cd web && npm run build

# Backend
cd api && go build ./...

# Docker (opcional)
docker compose build web && docker compose build api && docker compose up -d

# Verificar endpoints
curl http://localhost:3000/inscripcion  # formulario
curl http://localhost:28080/api/v1/inscripcion/check-ci?ci=12345678  # check
```

---

## 12. Riesgos y mitigaciones

| Riesgo | Impacto | Mitigación |
|--------|---------|------------|
| Spam en formulario público | Alto | Rate limiting por IP, validación estricta |
| Archivos maliciosos en uploads | Alto | Sanitización con `SanitizeImage`/`SanitizeDocument`, validación MIME |
| Race condition en Nº control | Bajo | Usar transacción DB con `SELECT MAX + INSERT` |
| Duplicados de CI entre check y submit | Bajo | Unique index en DB + validación en service layer |
| Email no se envía | Medio | Log de errores, posibilidad de reenviar manualmente |

---

## 13. Futuras mejoras (no incluidas en esta fase)

- Formulario de seguimiento de solicitud (el interesado consulta estado con su cédula)
- Notificaciones push al admin cuando llega una solicitud nueva
- Firma digital del formulario de inscripción FPV
- Generación automática del expediente en PDF
- Integración con pasarela de pago para calcular monto automáticamente
