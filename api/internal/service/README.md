[← Volver al inicio](../../README.md)

# 🧠 `service` — El cerebro del sistema (Lógica de Negocio)

> **¿Qué es esto?** La capa más importante de toda la aplicación. Si el proyecto fuera un restaurante 🍽️:
> - `utils/` serían los cuchillos y tablas de cortar (herramientas)
> - `handler/` sería el mesonero que toma la orden (HTTP)
> - `repository/` sería la despensa (base de datos)
> - **`service/` es la cocina 👨‍🍳** — acá se decide qué se cocina, cómo, y si el cliente tiene permiso para comerlo
>
> **📍 Dónde está:** `api/internal/service/`
>
> **🧠 Para el junior:** Si `utils/` era la caja de herramientas, `service/` es el **manual de instrucciones del sistema**. Acá vive la lógica que hace que el colegio funcione: registrar psicólogos, iniciar sesión, enviar correos, controlar quién puede hacer qué.

---

## 📋 Índice rápido

| # | Archivo(s) | ¿Qué hace? | ¿Para qué sirve? |
|---|---|---|---|
| 1 | 👑 `admin_service.go` + test | CRUD de administradores, login, permisos | Gestionar al personal del colegio |
| 2 | 📊 `analytics_service.go` | Estadísticas y telemetría | Dashboard en tiempo real |
| 3 | 🗺️ `error_mapper.go` | Traduce errores de BD a mensajes en español | Que el usuario entienda qué pasó |
| 4 | 📧 `mail_service.go` | Envío de correos en segundo plano | Bienvenidas, notificaciones, alertas |
| 5 | 📰 `post_service.go` + test | Gestión de noticias (CMS) | Publicar contenido en el portal |
| 6 | 🧑‍⚕️ `psi_service.go` + test | Lógica principal de psicólogos | Perfiles, login, búsqueda, importación |
| 7 | 📑 `psi_service_xlsx.go` | Importar psicólogos desde Excel | Carga masiva de datos |
| 8 | 🔧 `psi_user_admin_service.go` + test | Admin editando psicólogos | Panel de control del staff |
| 9 | 📱 `social_media.go` + test | Redes sociales de psicólogos | Instagram, Facebook, etc. |
| 10 | 🏷️ `specialty_service.go` + test | Especialidades/áreas de trabajo | Catálogo de especialidades |

---

## 🔍 Antes de empezar: conceptos clave

### ¿Qué es un "Service"?

Un **Service** (servicio) es una estructura en Go que agrupa toda la lógica de negocio de un área específica. Por ejemplo, `AdminService` agrupa TODO lo relacionado con administradores: crearlos, listarlos, actualizarlos, borrarlos, iniciar sesión.

Cada Service:
1. ✅ Recibe datos del handler (controlador HTTP).
2. ✅ Valida permisos (¿quién puede hacer esto?).
3. ✅ Aplica reglas de negocio (¿esto tiene sentido?).
4. ✅ Llama al repository (base de datos) si todo está bien.
5. ✅ Devuelve el resultado o un error.

### Inyección de Dependencias (DI)

```go
type AdminService struct {
    repo        domain.UserAdminRepository  // 🔗 Se conecta a la BD
    cache       *cache.Cache                // ⚡ Guarda datos en RAM
    mailService *MailService                // 📧 Envía correos
}
```

En vez de que `AdminService` cree sus propias conexiones, **se las pasan desde afuera** ("inyección"). Esto se llama **Inversión de Dependencias** y es importante porque:
- ✅ **Testeable:** Podés pasar un "mock" (simulador) en vez de una BD real.
- ✅ **Flexible:** Si cambiás de base de datos, solo cambiás el repositorio, no el service.
- ✅ **Claro:** Ves todas las dependencias de un servicio de un solo vistazo.

### JWT (Jason Web Token)

Es un **carnet digital** que prueba que un usuario inició sesión. Cuando un admin o psicólogo se loguea, el sistema genera un JWT que contiene:
- `user_id`: Quién es el usuario.
- `role`: Si es admin, psicólogo, etc.
- `exp`: Cuándo expira (24 horas después).
- `iat`: Cuándo se emitió.

El JWT se firma con una **llave secreta** que cambia cada vez que el usuario inicia sesión (Key Rotation 🔑). Si alguien roba un JWT, al usuario hacer login de nuevo, el JWT robado deja de funcionar.

### Key Rotation (Rotación de Llaves)

Cada vez que un usuario inicia sesión:
1. Se genera una **nueva llave secreta** (UUID v7).
2. Se guarda en la base de datos.
3. Se firma el JWT con esa llave.
4. Cualquier JWT anterior (de otro dispositivo/sesión) **deja de funcionar**.

Esto significa que solo UN dispositivo puede tener sesión activa a la vez. Es una decisión de seguridad: si un psicólogo pierde su celular, el ladrón tiene el JWT, pero cuando el psicólogo inicie sesión desde otro lado, el JWT del ladrón se invalida solo.

### bcrypt

Es un algoritmo para guardar contraseñas de forma segura. En vez de guardar `"Pass1234!"` directamente (lo cual sería un desastre si alguien roba la BD), guarda un "hash" (texto revuelto) generado con bcrypt.

Características de bcrypt:
- **Lento a propósito:** Generar un hash toma ~100ms. Para el usuario que se loguea es imperceptible. Para un hacker que robe la BD y quiera probar millones de contraseñas, es **eternidad**.
- **Sal automática:** Cada hash incluye una "sal" (texto aleatorio) que hace que dos personas con la misma contraseña tengan hashes diferentes.

### PATCH vs PUT

- **PUT:** Reemplazá TODO el recurso. Si enviás `{name: "Fran"}`, los campos que no enviaste se borran.
- **PATCH:** Actualizá SOLO los campos que enviás. Si enviás `{name: "Fran"}`, el resto queda igual.

En este proyecto, todas las actualizaciones se hacen con **semántica PATCH** usando punteros (`*string`, `*bool`). Si el campo es `nil`, no se toca. Si tiene valor, se actualiza.

### Soft Delete (Borrado lógico)

En lugar de BORRAR un registro de la base de datos (lo que rompe relaciones y no se puede deshacer), se marca como "inactivo" o se le pone una fecha de borrado. El registro sigue existiendo para auditoría, pero no aparece en búsquedas ni en el sistema.

### Goroutines ( `go func()` )

```go
go func() {
    // Esto corre en SEGUNDO PLANO
    s.db.Create(&algo)
}()
```

`go` lanza una función en **segundo plano**. El programa principal sigue ejecutándose sin esperar a que termine. Es como pedir un delivery 📦: no te quedás mirando la puerta esperando, seguís con tu vida y la comida llega cuando esté lista.

Acá se usa para:
- Enviar correos (no querés que el usuario espere 2 segundos a que el SMTP responda).
- Registrar estadísticas (no querés ralentizar el login por guardar un evento).

---

## 👑 `admin_service.go` — Gestión de administradores

### 📖 ¿Qué es `AdminService`?

El servicio que maneja TODO lo relacionado con el personal del colegio: quienes tienen acceso al panel de control para gestionar psicólogos, publicar noticias, etc.

### 🏗️ Estructura

```go
type AdminService struct {
    repo        domain.UserAdminRepository  // 📦 Base de datos
    cache       *cache.Cache                // ⚡ Caché en memoria (RAM)
    mailService *MailService                // 📧 Correos electrónicos
}
```

Tres dependencias:
1. **`repo`** — Donde se guardan los datos de los admins (BD).
2. **`cache`** — Un "archivo rápido" en la memoria RAM para no consultar la BD cada 2 segundos.
3. **`mailService`** — Para enviar correos de bienvenida, notificaciones de login, etc.

### 🔄 `Login` — Inicio de sesión de staff

```go
func (s *AdminService) Login(ctx context.Context, identifier, password string) (string, error)
```

**Recibe:** El usuario (email o username) y la contraseña.
**Devuelve:** Un JWT (token) si las credenciales son correctas, o un error.

**Paso a paso:**

```
📥 Llega "admin@colpsi.com" y "Pass1234!"
│
▼
1. Sanitizar: pasa identifier a minúsculas
   "Admin@Colpsi.com" → "admin@colpsi.com"
   (Previene errores por capitalización en la BD)
│
▼
2. Buscar en la BD al admin por su identifier
   ¿Lo encontró?
   ├── ❌ No → error genérico "credenciales inválidas"
   └── ✅ Sí → continúa
│
▼
3. ¿Está activo el admin?
   ├── ❌ No → error "la cuenta está desactivada"
   └── ✅ Sí → continúa
│
▼
4. Verificar contraseña con bcrypt
   bcrypt.CompareHashAndPassword(hash_guardado, contraseña_ingresada)
   ├── ❌ No coincide → error genérico "credenciales inválidas"
   └── ✅ Coincide → continúa
│
▼
5. Key Rotation (Rotación de Llave Secreta)
   newKey = uuid.New()   // Ej: "550e8400-e29b-41d4-a716-446655440000"
   admin.Key = newKey    // Se guarda la nueva llave en la BD
   → Cualquier sesión anterior queda INVALIDADADA
│
▼
6. Generar JWT firmado con la nueva llave
   token = jwt.SigningMethodHS256({
       user_id: admin.ID,
       exp:     ahora + 24 horas,
       iat:     ahora,
       role:    "admin",
   }, newKey)
│
▼
7. Enviar correo de notificación (en segundo plano)
   "Alguien inició sesión con tu cuenta. ¿Fuiste vos?"
   (Si el SMTP falla, no importa, el login sigue funcionando)
│
▼
📤 Devuelve el token JWT
```

**¿Por qué el mensaje de error es tan genérico?**

Siempre dice "credenciales inválidas", nunca "el usuario existe pero la contraseña está mal".

Esto es **intencional** por seguridad. Si el sistema dijera "el usuario existe" pero "la contraseña está mal", un hacker podría:
1. Probar miles de usuarios hasta encontrar uno que existe.
2. Saber qué usuarios existen en el sistema.
3. Enfocar su ataque solo en esos.

Al decir siempre lo mismo, el atacante no sabe si falló el usuario o la contraseña.

### 📋 `GetAdmins` — Listado de administradores

```go
func (s *AdminService) GetAdmins(ctx context.Context, active *bool, search string, page, limit int) (interface{}, error)
```

**Característica especial: Caché en RAM**

```go
cacheKey := fmt.Sprintf("admins_l:%d_p:%d_s:%s_a:%v", limit, page, search, active)

// ¿Ya tengo este resultado guardado en la RAM?
if cached, found := s.cache.Get(cacheKey); found {
    return cached, nil  // ✅ Devuelve al toque, ni consulta la BD
}

// No estaba en caché → consulto la BD
result, err := s.repo.List(...)
s.cache.Set(cacheKey, result, 5*time.Minute)  // Guardar por 5 minutos
return result, nil
```

**¿Por qué es importante?**

Sin caché, cada vez que un admin abre el listado de otros admins, el sistema consulta la BD. Si 10 admins abren la página cada 30 segundos, son 10 consultas por minuto a la BD.

Con caché:
- **Primera vez:** Consulta a la BD (lento, ~50ms). Guarda en RAM.
- **Segunda vez (segundos después):** Lo saca de la RAM (instantáneo, ~1ms).
- Pasados 5 minutos, el caché expira y se vuelve a consultar la BD.

### 🛡️ La Matriz de Permisos

```go
type permissionUpdate struct {
    name       string      // "Crear Psi"
    requested  *bool       // ¿El request pide activarlo?
    current    bool        // ¿Está activo ahora?
    updaterHas bool        // ¿El que modifica tiene este permiso?
    setTarget  func(bool)  // Cómo cambiar el valor (callback)
}
```

**Analogía:** Tenés una hoja de cálculo con 15 permisos (Crear Psicólogo, Borrar Admin, Publicar Noticias, etc.). Cada permiso puede estar activo ✅ o inactivo ❌.

Cuando un admin quiere modificar los permisos de OTRO admin, el sistema construye esta matriz y verifica:
- ¿El que modifica tiene el permiso que quiere otorgar?
  - Si no → error "no puedes otorgar el permiso: X"
  - Si sí → se aplica el cambio

**¿Por qué no se puede usar `reflect` (reflexión) acá?**

Porque `reflect` es lento y hay que validar 15 permisos. Usar funciones lambda (closures) es más rápido y seguro.

```go
// En vez de hacer magia con reflect (lento y frágil):
for _, perm := range matrix {
    if perm.requested != nil && *perm.requested && !perm.updaterHas {
        return fmt.Errorf("no puedes otorgar el permiso: %s", perm.name)
    }
}
// Cada perm.update sabe cómo cambiar el valor sin adivinar
```

### 🧪 Cómo probar manualmente

Necesitarías un test real, pero la idea es:

```go
// 1. Crear el servicio con un repositorio falso
repo := &mockAdminRepo{}
svc := NewAdminService(repo, nil)

// 2. Programar el comportamiento del repo falso
repo.GetByIdentifierFunc = func(...) (*domain.UserAdmin, error) {
    return &domain.UserAdmin{...}, nil
}

// 3. Llamar al login
token, err := svc.Logion(context.Background(), "admin", "Pass1234!")

// 4. Verificar que funciona
if err != nil || token == "" {
    t.Error("Fallo el login")
}
```

---

## 📊 `analytics_service.go` — Estadísticas en tiempo real

### 📖 ¿Qué es `AnalyticsService`?

Un servicio que registra **todo lo que pasa** en el sistema (logins, búsquedas, visitas a perfiles) y genera un **dashboard** con estadísticas para los administradores.

### 🔥 "Fire-and-Forget" (Disparar y Olvidar)

Todos los métodos de escritura usan este patrón:

```go
func (s *AnalyticsService) RecordLogin(userID, username, role, ip, userAgent string) {
    go func() {  // ← Lanza en segundo plano
        s.db.Create(&domain.LoginEvent{...})        // Guardar el evento
        s.db.Where(...).Assign(session).FirstOrCreate(&session)  // Actualizar sesión activa
    }()
    // La función termina acá, el go func() sigue corriendo solo
}
```

**¿Por qué?**

Cuando un usuario inicia sesión, el sistema debe:
1. ✅ Validar credenciales (esto SÍ es urgente).
2. ✅ Generar el JWT (esto SÍ es urgente).
3. 📊 Guardar en la BD que el usuario inició sesión (esto NO es urgente).

La parte urgente se hace en el momento. La parte de estadísticas se **dispara** al fondo y sigue ejecutándose mientras el usuario ya está navegando.

### 📈 Dashboard completo

`GetDashboardStats()` devuelve TODO en UNA sola llamada:

```go
type DashboardStats struct {
    // 📊 Logins
    LoginsTotal      int64   // ¿Cuántos logins desde que existe el sistema?
    LoginsToday      int64   // ¿Cuántos hoy?
    LoginsThisWeek   int64   // ¿Esta semana?
    UniqueUsersToday int64   // ¿Cuántos usuarios distintos hoy?
    
    // 👀 Visitas al portal
    PageViewsTotal      int64
    UniqueVisitorsToday int64  // Visitantes únicos (por cookie/session)
    
    // 🔍 Búsquedas
    SearchesTotal    int64
    SearchesToday    int64
    
    // 👤 Visitas a perfiles
    ProfileViewsTotal int64
    ProfileViewsToday int64
    
    // 🟢 En vivo
    ActiveSessionsNow int64  // Usuarios conectados AHORA
    
    // 🏆 Rankings (top de los últimos 30 días)
    TopSpecialties []TopItem  // "Psicología Clínica" fue buscada 150 veces
    TopMunicipios  []TopItem  // "Valencia" fue buscada 300 veces
    TopSearchTerms []TopItem  // "ansiedad" fue buscada 80 veces
    TopProfiles    []TopProfile  // El psicólogo más visitado
    
    // 📈 Tendencias (gráficos de los últimos 14 días)
    LoginTrend []DailyCount  // Lunes: 50 logins, Martes: 45...
    ViewTrend  []DailyCount
}
```

### 🧹 Limpieza automática (`PurgeOldData`)

```go
func (s *AnalyticsService) PurgeOldData(olderThanDays int) {
    cutoff := time.Now().AddDate(0, 0, -olderThanDays)
    s.db.Where("created_at < ?", cutoff).Delete(&domain.PageView{})
    s.db.Where("created_at < ?", cutoff).Delete(&domain.SearchEvent{})
    s.db.Where("created_at < ?", cutoff).Delete(&domain.ProfileView{})
    // LoginEvent NUNCA se borra (es auditoría)
}
```

Esto debe ejecutarse como una **tarea programada** (cron job) una vez al día. Borra datos viejos de navegación y búsquedas para que la BD no crezca infinitamente.

Pero **NUNCA** borra los logins, porque los registros de inicio de sesión son **requisito legal** de auditoría de seguridad.

### 🧪 Cómo probar manualmente

```go
// Crear el servicio con una BD en memoria
// (esto requiere una instancia de gorm.DB de prueba)

svc := NewAnalyticsService(dbDePrueba)

// Registrar un login
svc.RecordLogin(uuid.New(), "psicologo1", "psi", "192.168.1.1", "Chrome/120")

// Obtener estadísticas
stats, _ := svc.GetDashboardStats()
fmt.Printf("Logins hoy: %d\n", stats.LoginsToday)
```

---

## 🗺️ `error_mapper.go` — El traductor de errores de BD

### 📖 ¿Qué hace?

```go
func MapDBError(err error) error
```

Cuando GORM (el ORM de Go) o PostgreSQL tiran un error, suelen ser cosas como:

> `ERROR: duplicate key value violates unique constraint "idx_psi_users_ci"`

Eso no se le puede mostrar a un usuario. `MapDBError` lo traduce a:

> "La Cédula de Identidad ya se encuentra registrada"

### 🎯 ¿Qué errores traduce?

| Error técnico (BD) | Mensaje para el usuario |
|---|---|
| `idx_psi_users_ci` o `uni_psi_users_ci` | "La Cédula de Identidad ya se encuentra registrada" |
| `idx_psi_users_fpv` o `uni_psi_users_fpv` | "El número de FPV ya está registrado por otro psicólogo" |
| `uni_psi_users_email` | "El correo electrónico ya está en uso" |
| `uni_psi_users_username` | "El nombre de usuario ya existe" |
| Cualquier otro | Se deja pasar como está (lo manejará el handler) |

### ¿Por qué hay dos nombres para cada índice? (`idx_` y `uni_`)?

Porque GORM/AutoMigrate a veces genera los índices con distintos prefijos según la versión. Para cubrir ambos casos, el mapper busca ambos patrones.

### 🧪 Cómo probar manualmente

```go
err1 := errors.New("duplicate key idx_psi_users_ci")
fmt.Println(MapDBError(err1))
// → "La Cédula de Identidad ya se encuentra registrada"

err2 := errors.New("duplicate key uni_psi_users_fpv")
fmt.Println(MapDBError(err2))
// → "El número de FPV ya está registrado por otro psicólogo"

err3 := errors.New("connection refused")
fmt.Println(MapDBError(err3))
// → "connection refused" (no se traduce, se deja pasar)
```

---

## 📧 `mail_service.go` — El cartero del sistema

### 📖 ¿Qué es `MailService`?

Un sistema de envío de correos electrónicos que trabaja en **segundo plano** usando el patrón **Productor-Consumidor**.

### 🤔 ¿Por qué es tan complicado?

Enviar un correo por SMTP toma ~2 segundos. Si el sistema enviara el correo **mientras** el usuario espera, el registro o login tardaría 2 segundos más de lo necesario.

**Solución:** Encolar el correo en una lista en memoria y que un trabajador de fondo (Worker) los vaya enviando de a poco.

### 🏗️ Arquitectura

```
📥 Llega una petición: "Registrar psicólogo"
│
▼
┌──────────────────────────────────────┐
│ Registro en BD (rápido, ~10ms)      │
└──────────────────────────────────────┘
│
▼
┌──────────────────────────────────────┐
│ s.mailService.SendEmail(...)         │
│ → Encola el correo en el canal      │
│ → TARDA ~0.001ms                    │
│ → El usuario ya recibió respuesta   │
└──────────────────────────────────────┘
│
▼
┌──────────────────────────────────────┐
│ Worker (corre en segundo plano)     │
│ Toma correos de la cola UNO POR UNO │
│                                                                     │
│ Cada 30 correos → pausa de 60-180 seg                             │
│ Entre cada correo → pausa de 500ms                                 │
│ (Para que Gmail no lo marque como spam)                            │
└──────────────────────────────────────┘
```

### 🔍 El canal (`chan MailJob`)

```go
queue: make(chan MailJob, 5000)
```

Un **canal** en Go es una tubería 🔧 por donde viajan datos entre diferentes partes del programa.

- `MailJob` = un correo a enviar.
- `5000` = el tamaño de la tubería. Puede tener hasta 5000 correos esperando.
- Si la tubería se llena (5000 correos esperando), el sistema se frena para no explotar la memoria RAM.

### 🐢 Anti-Spam (Throttling y Jittering)

```go
if sentInBatch >= 30 {
    waitTime := rand.Intn(120) + 60  // 60 a 180 segundos
    time.Sleep(waitTime * time.Second)
    sentInBatch = 0
}
```

**¿Por qué pausas?**

Gmail, Outlook y otros proveedores detectan como **spam** a cualquiera que envíe muchos correos seguidos. Si el colegio envía un boletín a 2000 psicólogos todos a la vez, el servidor SMTP bloquea la IP.

**Estrategias:**
1. 📦 **Batch de 30:** Cada 30 correos, el worker descansa.
2. 🎲 **Jittering:** El descanso no es fijo (ej: 60 segundos). Es aleatorio entre 1 y 3 minutos. Esto simula el comportamiento humano y evade los algoritmos anti-spam.
3. ⏸️ **Pausa entre correos:** 500ms entre cada uno para no saturar la conexión SMTP.

> [!WARNING]
> **Ojo:** Hay una línea comentada en el código:
> ```go
> // if err := s.client.DialAndSend(m); err != nil {
> //    return fmt.Errorf("fallo la conexión SMTP o el envío: %w", err)
> // }
> ```
> El envío **real** del correo está comentado. Esto sugiere que el sistema aún está en desarrollo/pruebas y los correos no se están enviando físicamente. Solo se encolan y el worker los procesa en vano.

### 🧪 Cómo probar manualmente

```go
// Inicializar el servicio (esto conecta con el SMTP real)
mailSvc, err := NewMailService()
if err != nil {
    log.Fatal("Error iniciando mail service:", err)
}

// Encolar un correo (esto es instantáneo)
err = mailSvc.SendEmail(
    "psicologo@example.com",
    "Bienvenido al Colegio",
    "welcome_psi",
    map[string]interface{}{
        "Name": "Juan",
        "Email": "juan@example.com",
        "Password": "Temp1234!",
    },
)

if err != nil {
    fmt.Println("Error encolando:", err)
} else {
    fmt.Println("✅ Correo encolado, el worker lo enviará en breve")
}
```

---

## 📰 `post_service.go` — Gestor de contenido (CMS)

### 📖 ¿Qué es `PostService`?

El servicio que maneja las **noticias y publicaciones** del portal del colegio. Es como el admin de WordPress.

### 🏗️ Estructura

```go
type PostService struct {
    repo      domain.PostRepository    // 📦 BD
    s3Client  *s3.S3Client             // ☁️ Almacenamiento de imágenes
    sanitizer *bluemonday.Policy       // 🧹 Limpiador de HTML
}
```

**`bluemonday`** es una librería que limpia HTML para prevenir **XSS** (Cross-Site Scripting). Si un admin escribe:

```html
<p>Bienvenidos al colegio</p>
<script>alert('hackeado!')</script>
```

`bluemonday` borra el `<script>` y solo deja:

```html
<p>Bienvenidos al colegio</p>
```

### 🔐 Control de Acceso por Rol (ACL)

Diferentes usuarios ven diferentes cosas. Esto se maneja con un `switch`:

```go
switch userRole {
case "admin":
    // 👑 TODO: borradores, programados, archivados, publicados
    filter.Status = nil  // Sin filtro = ve todo

case "psi":
    // 🧑‍⚕️ Psicólogos: solo lo publicado
    filter.Status = []domain.PostStatus{PostStatusPublished}

default: // "public"
    // 👤 Visitantes: solo lo publicado y público
    filter.Status = []domain.PostStatus{PostStatusPublished}
    filter.Type = "public"
}
```

**¿Por qué el admin no tiene filtro?**

Porque necesita ver borradores (para editarlos), programados (para revisar la fecha), archivados (para restaurarlos) y publicados. Todos los estados.

**¿Por qué el público solo ve tipo "public"?**

Porque hay posts que son solo para psicólogos (tipo "psi"), como comunicados internos del gremio.

### 🖼️ Manejo de imágenes con S3

```go
// Sanitizar la imagen (eliminar EXIF, comprimir a WebP)
cleanBytes, ext, contentType, err := utils.SanitizeDocument(src)

// Generar nombre único (UUID v7) para evitar colisiones
filename := uuid.Must(uuid.NewV7()).String() + ext

// Subir a Amazon S3 / MinIO
key, err := s.s3Client.UploadStream(ctx, bytes.NewReader(cleanBytes), "posts", filename, contentType)
```

**¿Por qué UUID para el nombre?**

Si dos admins suben `foto.jpg`, el segundo pisaría al primero. Usando UUID, los nombres son únicos:
- `550e8400-e29b-41d4-a716-446655440000.webp`
- `6ba7b810-9dad-11d1-80b4-00c04fd430c8.webp`

### 🔄 Transacciones distribuidas (Saga Pattern)

```go
// 1. Subir imagen a S3 (éxito ✅)
newS3Key = key

// 2. Guardar en BD (falla ❌ porque el título está vacío)
if err := s.repo.Update(ctx, post, textModel); err != nil {
    // 3. ROLLBACK: Borrar la imagen de S3 que acabamos de subir
    if newS3Key != oldS3Key {
        s.s3Client.DeleteFile(ctx, newS3Key)
    }
    return err
}

// 4. Si todo salió bien, borrar la imagen vieja (Garbage Collection)
if newS3Key != oldS3Key && oldS3Key != "" {
    s.s3Client.DeleteFile(ctx, oldS3Key)
}
```

**¿Por qué no se puede usar una transacción normal?**

Porque S3 (en la nube) y PostgreSQL (base de datos) son sistemas **separados**. No hay una transacción que abarque ambos. Si una falla y la otra no, quedan datos huérfanos.

La solución es **Saga Pattern**: si el paso 2 falla, ejecutamos un paso de "compensación" (borrar lo subido en S3).

---

## 🧑‍⚕️ `psi_service.go` — El archivo más grande del proyecto

### 📖 ¿Qué es `PsiService`?

**1435 líneas** de código. Es el servicio más grande y complejo del sistema. Maneja TODO lo relacionado con los psicólogos colegiados.

### 🏗️ Estructura

```go
type PsiService struct {
    repo        domain.PsiUserRepository  // 📦 BD
    s3Client    *s3.S3Client              // ☁️ Imágenes
    mailService IMailService              // 📧 Correos
    sanitizer   *bluemonday.Policy        // 🧹 Sanitización XSS
}
```

### 🔐 `Login` — Inicio de sesión de psicólogos

```go
func (s *PsiService) Login(ctx context.Context, identifier, password string) (string, *domain.PsiUserModel, error)
```

Misma lógica que `AdminService.Login`:
1. Buscá al psicólogo por identifier.
2. ¿Está activo? Si no → error "cuenta inactiva o suspendida".
3. Verificar contraseña con bcrypt.
4. **Key Rotation:** Generá nueva llave → JWT anterior invalidado.
5. Guardá la nueva llave en BD.
6. Generá JWT firmado.
7. Enviá correo de notificación de login (en segundo plano).
8. Devolvé el token + datos del psicólogo.

### 🌐 `LoginLibrary` — Inicio de sesión para la biblioteca digital

```go
func (s *PsiService) LoginLibrary(ctx context.Context, identifier, password string) (string, error)
```

**¿Qué es "Audiobookshelf"?** Es un sistema de **biblioteca digital** donde los psicólogos pueden acceder a libros y recursos académicos. El colegio tiene un microservicio separado para esto.

`LoginLibrary`:
1. Valida credenciales en el sistema central (ColPsi). ✅
2. Genera un JWT específico para Audiobookshelf. ✅
3. **Sincroniza la cuenta:** Si el usuario no existe en Audiobookshelf, lo crea automáticamente. 🔄
4. Si ya existe (HTTP 409 Conflict), no hace nada (es seguro, es "idempotente"). ✅

**El JWT de la biblioteca** tiene una expiración de **30 días** (no 24 horas como el normal) porque los psicólogos no inician sesión en la biblioteca todos los días.

### 🔄 Sincronización con Audiobookshelf

```go
func (s *PsiService) sincronizarConAudiobookshelf(ctx context.Context, username, password, email string) (string, error) {
    // Llamar a la API interna de Audiobookshelf
    url := "http://audiobookshelf:80/api/users"
    // ... envía POST con username, password, email ...
    
    if resp.StatusCode == 409 {
        // 409 = Conflict (ya existe)
        return "", nil  // No es error, sigue tranquilo
    }
    
    // Decodificar la respuesta para obtener el ID del usuario en la biblioteca
    return absData.User.ID, nil
}
```

**Dato importante:** `sincronizarConAudiobookshelf` tiene un **timeout de 5 segundos**.

```go
client := &http.Client{Timeout: 5 * time.Second}
```

Si la biblioteca está caída, el login no debe esperar para siempre. A los 5 segundos, se cancela la petición y el usuario puede iniciar sesión igual (sin acceso a la biblioteca por ahora).

### 👤 `GetPublicProfile` — Perfil público de psicólogo

```go
func (s *PsiService) GetPublicProfile(ctx context.Context, id int) (*request_structs.PsiFullProfileDTO, uuid.UUID, error)
```

Esta función construye la **ficha técnica** que ve cualquier persona en el portal público.

#### Escudo de Privacidad (Privacy Shield)

El psicólogo puede decidir qué datos mostrar y cuáles ocultar:

```go
if psi.ShowContactEmail {
    dto.Email = psi.ContactEmail  // ✅ Muestra el email
}
// Si ShowContactEmail es false, el email no se incluye en el DTO
```

Esto se aplica a: email, teléfono, dirección, celular, etc. Cada campo tiene su propio "switch" booleano de visibilidad.

#### Restricción de Solvencia

Si el psicólogo **no está solvente** (no pagó las cuotas gremiales), el perfil público se "degrada":

```go
if !psi.Solvent {
    return &request_structs.PsiFullProfileDTO{
        FirstName: psi.FirstName,
        // ... datos básicos ...
        PostGrades: []request_structs.PostGradeDTO{},  // ❌ Vacío
        // No se muestran postgrados, ni redes sociales, ni nada avanzado
    }, nil, nil
}
```

**¿Por qué mostrar algo y no un 404?** Por razones de SEO (posicionamiento en Google). Si el perfil diera error 404, Google lo desindexaría. Al mostrar datos básicos, el perfil sigue existiendo pero con información limitada hasta que el psicólogo se ponga al día.

### 📋 `GetPublicDirectory` — Listado público de psicólogos

```go
func (s *PsiService) GetPublicDirectory(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) (interface{}, error)
```

Construye el **directorio de psicólogos** que se ve en el portal. Aplica:
- Paginación (página 1 de 12 resultados).
- Filtro por género (M/F).
- Búsqueda por nombre, especialidad, ubicación.
- **NO muestra** el estado de solvencia (es información interna).

### 📥 `ImportFromCSV` — Carga masiva desde Excel

Este método es enorme y complejo. Veámoslo simplificado:

```go
func (s *PsiService) ImportFromCSV(ctx context.Context, reader io.Reader, adminID uuid.UUID) (int, []map[string]string) {
    // 1. Abrir archivo Excel
    f, _ := excelize.OpenReader(reader)
    
    // 2. Optimización: hashear la contraseña UNA SOLA VEZ
    defaultPassword := "Colpsi2025!"
    hashedPassword, _ := bcrypt.GenerateFromPassword(defaultPassword)
    
    // 3. Recorrer fila por fila
    for _, row := range rows {
        // Normalizar municipio
        municipio := NormalizeMunicipioCarabobo(raw)
        
        // Normalizar estado
        estado := NormalizeEstadoVenezuela(raw)
        
        // ... completar modelo ...
        
        // Guardar en BD (el método auto-crea: usuario + datos colegiales + solvencias)
        err := s.repo.CreateWithColData(ctx, psi, colData, solvencia, postgrados)
        if err != nil {
            failedRecords = append(failedRecords, "Fila 5 falló: " + err.Error())
            continue  // ❌ No se detiene, sigue con la siguiente fila
        }
        
        // Enviar correo de bienvenida (en segundo plano)
        go s.mailService.SendEmail(...)
        
        successCount++
    }
    
    return successCount, failedRecords
}
```

**Tolerancia a fallos:** Si una fila del Excel está mal (FPV duplicado, municipio inválido), se registra el error en `failedRecords` y se continúa con la siguiente fila. El administrador recibe un reporte al final con cuántos se importaron y cuáles fallaron.

---

## 📑 `psi_service_xlsx.go` — Importación desde Excel (versión mejorada)

### 📖 ¿Qué diferencia hay con `ImportFromCSV`?

`ImportFromCSV` (en `psi_service.go`) y `ImportFromXLSX` (en este archivo) hacen lo mismo pero:
- `ImportFromCSV` espera un formato TSV/CSV específico (columnas separadas por tabulación).
- `ImportFromXLSX` espera un archivo `.xlsx` de Excel con una hoja llamada **"BD ColPsiCarabobo 2026"**.

En la práctica, `ImportFromXLSX` es el método más completo y el que probablemente se usa en producción.

### 🧪 Cómo probar manualmente

```go
// Abrir un archivo Excel de prueba
file, _ := os.Open("psicologos.xlsx")
defer file.Close()

// Crear el servicio
svc := NewPsiService(repo, s3Client, mailService)

// Importar
success, failed := svc.ImportFromXLSX(context.Background(), file, adminID)

fmt.Printf("✅ Importados: %d\n", success)
fmt.Printf("❌ Fallaron: %d\n", len(failed))
for _, f := range failed {
    fmt.Printf("  Fila %s: %s\n", f["fila"], f["error"])
}
```

---

## 🔧 `psi_user_admin_service.go` — Admin editando psicólogos

### 📖 ¿Qué hace?

Este archivo implementa las operaciones que los **administradores** pueden hacer sobre los perfiles de los psicólogos, pero que los psicólogos NO pueden hacer sobre sí mismos:
- 👁️ Ver el expediente COMPLETO de cualquier psicólogo (sin escudo de privacidad).
- ✏️ Editar cualquier campo, incluso los que el psicólogo no puede auto-editar.
- 🗑️ Borrar lógicamente a un psicólogo.
- 📋 Ver el listado administrativo (con datos de solvencia incluidos).

### 👁️ `GetPsiByIDAdmin` — Expediente completo

```go
func (s *PsiService) GetPsiByIDAdmin(ctx context.Context, admin *domain.UserAdmin, targetID uuid.UUID) (*domain.PsiUserModel, error)
```

A diferencia de `GetPublicProfile` que oculta datos según la privacidad del psicólogo, esta función **los muestra todos** porque:
- El administrador necesita ver teléfonos, correos, dirección exacta, historial de solvencia.
- Es para uso operativo y legal (el colegio necesita tener estos datos aunque el psicólogo los oculte en el perfil público).

**Eso sí, hay que tener permiso:**
```go
if !admin.Sudo && !admin.CanUpdatePsi && !admin.CanCreatePsi {
    return nil, errors.New("permisos insuficientes")
}
```

### ✏️ `UpdatePsiByAdmin` — Edición maestra

Este método es **enorme** (700+ líneas) porque permite editar ABSOLUTAMENTE TODO del psicólogo:
- Datos personales (nombre, apellido, cédula, fecha de nacimiento).
- Credenciales (username, email, contraseña).
- Estado gremial (solvente, activo, prueba de vida).
- Contacto (teléfono, celular, dirección en Carabobo, fuera de Carabobo, fuera de Venezuela).
- Perfil profesional (áreas de trabajo, mini bio, biografía extensa).
- Datos colegiales (universidad, fecha de graduación, número de registro).
- Banderas gremiales (director, colaborador, empleado público, profesor universitario).
- Solvencias (pagos de cuotas).
- Imágenes (foto de perfil, títulos).

Y usa **Transacciones Distribuidas (Saga)** igual que `post_service.go`: si la BD falla, borra las imágenes recién subidas a S3.

### 🧪 Cómo probar manualmente

```go
// Crear un admin con permisos
admin := &domain.UserAdmin{
    ID: uuid.New(), Username: "super_admin", Sudo: true,
}

// Llamar a GetPsiByIDAdmin
psiCompleto, err := svc.GetPsiByIDAdmin(ctx, admin, targetID)
if err != nil {
    log.Fatal("Permiso denegado:", err)
}
fmt.Printf("📋 Expediente de %s %s\n", psiCompleto.FirstName, psiCompleto.LastName)
fmt.Printf("📧 Email interno: %s\n", psiCompleto.Email)
fmt.Printf("📞 Teléfono: %s\n", psiCompleto.ContactPhone)
```

---

## 📱 `social_media.go` — Redes sociales

### 📖 ¿Qué hace?

Gestiona las **redes sociales** de los psicólogos. Cada psicólogo puede tener hasta **10** redes sociales vinculadas a su perfil público.

### 🎯 ¿Por qué un servicio separado?

Porque las redes sociales tienen reglas específicas:
1. **Límite de 10** — Un psicólogo no puede agregar infinitas redes (previene saturación de BD).
2. **Normalización** — El nombre de la red se normaliza con `utils.NormalizePlatformName` ("ig" → "Instagram").
3. **Propiedad** — Un psicólogo no puede editar/borrar las redes de otro psicólogo (prevención IDOR).
4. **Roles** — Los admins pueden borrar cualquier red (moderación), los psicólogos solo las suyas.

### 🛡️ Prevención IDOR (Insecure Direct Object Reference)

```go
func (s *PsiService) UpdateSocialNetwork(ctx context.Context, psi *domain.PsiUserModel, netID uuid.UUID, req request_structs.UpdateSocialNetworkRequest) error {
    network, _ := s.repo.GetSocialNetworkByID(ctx, netID)
    
    // 🔴 EL PSICÓLOGO A NO PUEDE EDITAR LA RED DEL PSICÓLOGO B
    if network.PsiUserID != psi.ID {
        return errors.New("no tienes permiso para editar esta red social")
    }
    
    // ... aplicar cambios ...
}
```

**Ataque IDOR:** El Psicólogo A ve en la URL `/api/social-media/abc-123` y cambia el ID a `/api/social-media/abc-456` que es la red del Psicólogo B. Sin esta validación, podría editar la red de otro.

### 🧪 Cómo probar manualmente

```go
// Agregar red social
req := request_structs.CreateSocialNetworkRequest{
    Name: "ig",           // → Se normaliza a "Instagram"
    URL: "https://instagram.com/psicologo",
}

err := svc.AddSocialNetwork(ctx, psi, req)
if err != nil {
    if err.Error() == "límite de redes sociales alcanzado (10)" {
        fmt.Println("❌ Ya tiene 10 redes, no puede agregar más")
    }
}
```

---

## 🏷️ `specialty_service.go` — Especialidades (catálogo maestro)

### 📖 ¿Qué es `SpecialtyService`?

Gestiona las **áreas de trabajo** o especialidades (Psicología Clínica, Neuropsicología, Psicología Infantil, etc.). Estas son **datos maestros**: un catálogo que usan todos los psicólogos para describir su área.

### ¿Por qué un servicio separado?

Porque las especialidades son un **catálogo compartido**:
- Cuando un psicólogo dice "soy psicólogo clínico", elige de esta lista.
- Cuando alguien busca "psicólogo infantil en Valencia", usa esta lista.
- Si se borra una especialidad, todos los psicólogos que la usaban se quedan sin área.

Por eso las operaciones de escritura tienen **permisos específicos** (`CanCreateTags`, `CanEditTags`, `CanDeleteTags`).

### Fail-Safe (Seguro por Defecto)

```go
func (s *SpecialtyService) GetSpecialties(ctx context.Context, requestedStatus string, isAdmin bool) ([]domain.PsiSpecialtyModel, error) {
    finalStatus := "active"  // ← Por defecto, SOLO activas
    
    if isAdmin {
        finalStatus = requestedStatus  // ← Solo el admin puede pedir "all" o "inactive"
    }
    
    return s.repo.GetAll(ctx, finalStatus)
}
```

Si un usuario público envía `?status=all` para intentar ver especialidades ocultas, el servicio **ignora** el parámetro y fuerza `"active"`. Es **seguro por defecto**: ante la duda, mostrá lo mínimo.

---

## 🧪 Los tests — Seguridad y confianza

Cada servicio tiene su archivo de tests correspondiente. Usan **Mocks** (simuladores) para no necesitar base de datos ni servidores reales.

### 🔍 Patrón: Func Override

```go
type mockAdminRepo struct {
    domain.UserAdminRepository  // "Hereda" la interfaz
    
    // Sobrescribe cada método con una función personalizable
    GetByIdentifierFunc func(ctx, identifier) (*domain.UserAdmin, error)
    UpdateFunc          func(ctx, admin) error
}

// El mock implementa la interfaz llamando a la función personalizable
func (m *mockAdminRepo) GetByIdentifier(ctx, id) (*domain.UserAdmin, error) {
    return m.GetByIdentifierFunc(ctx, id)
}
```

**Ventaja:** Cada test puede programar el comportamiento del repositorio:

```go
t.Run("Login exitoso", func(t *testing.T) {
    repo.GetByIdentifierFunc = func(...) (*domain.UserAdmin, error) {
        return &domain.UserAdmin{...}, nil  // Simula que el usuario existe
    }
    
    repo.UpdateFunc = func(...) error {
        return nil  // Simula que la actualización funciona
    }
    
    token, err := svc.Login(ctx, "admin", "pass")
    // ...
})

t.Run("Usuario no encontrado", func(t *testing.T) {
    repo.GetByIdentifierFunc = func(...) (*domain.UserAdmin, error) {
        return nil, errors.New("not found")  // Simula que el usuario NO existe
    }
    
    _, err := svc.Login(ctx, "admin", "pass")
    // Debe devolver error "credenciales inválidas"
})
```

### 📊 Lo que prueban los tests

| Archivo de test | ¿Qué prueba? |
|---|---|
| `admin_service_test.go` | Login con Key Rotation, Caché, Prevención de escalada de privilegios, Protección de Sudo, Anti auto-eliminación |
| `post_service_test.go` | Reglas de acceso por rol (ACL), Sanitización XSS, Filtros de listado, Permisos de admin |
| `psi_service_test.go` | Privacy Shield (solvencia y email), Login y JWT, Lazy Loading de ColData |
| `psi_user_admin_service_test.go` | Registro manual por admin, Actualización parcial (PATCH), Bloqueo de seguridad a no-admins |
| `social_media_test.go` | Límite de cuota (máx 10 redes), IDOR (editar red ajena), Roles (psi vs admin) |
| `specialty_service_test.go` | RBAC (crear especialidad), PATCH con auditoría, Fail-safe público, Fuga de métricas |

### 🧪 Cómo correr los tests

```bash
# Ir a la carpeta de servicios
cd api/internal/service

# Correr todos los tests
go test -v

# Correr tests de un archivo específico
go test -v -run TestAdminService_All

# Ver cobertura
go test -cover

# Ver cobertura DETALLADA (qué líneas están cubiertas)
go test -coverprofile=coverage.out && go tool cover -html=coverage.out
```

---

## 📊 Resumen de servicios

| Servicio | Archivo(s) | Dependencias | Operaciones clave |
|---|---|---|---|
| 👑 **AdminService** | `admin_service.go` | BD, Caché, Mail | Login, CRUD admins, Matriz de permisos |
| 📊 **AnalyticsService** | `analytics_service.go` | BD | RecordLogin, DashboardStats, PurgeOldData |
| 🗺️ **MapDBError** | `error_mapper.go` | — | Traduce errores de BD |
| 📧 **MailService** | `mail_service.go` | SMTP | SendEmail (asíncrono, con cola y anti-spam) |
| 📰 **PostService** | `post_service.go` | BD, S3, Bluemonday | CRUD posts, Filtros por rol, Imágenes |
| 🧑‍⚕️ **PsiService** | `psi_service.go` + `xlsx.go` | BD, S3, Mail, Bluemonday | Login, Perfiles, Directorio, Importación Excel |
| 🔧 **PsiUserAdminService** | `psi_user_admin_service.go` | BD, S3, Mail, Bluemonday | CRUD admin de psicólogos, Solvencias |
| 📱 **SocialMediaService** | `social_media.go` | BD | CRUD redes, Límite 10, IDOR protection |
| 🏷️ **SpecialtyService** | `specialty_service.go` | BD | CRUD especialidades, Fail-safe visibility |

### 🔗 Cómo se relacionan los servicios

```
🌐 Handler HTTP (router)
    │
    ▼
┌─────────────────────────────────────────────────────────┐
│                    SERVICE LAYER                        │
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ AdminService │  │  PsiService  │  │ PostService  │  │
│  │              │  │              │  │              │  │
│  │ Gestión de   │  │ Psicólogos   │  │ Noticias del │  │
│  │ admins y     │  │ colegiados,  │  │ portal, con  │  │
│  │ permisos     │  │ perfiles,    │  │ imágenes y   │  │
│  │              │  │ importación  │  │ sanitización │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  │
│         │                 │                 │          │
│         ▼                 ▼                 ▼          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ MailService  │  │AnalyticsSvc │  │SpecialtySvc  │  │
│  │ Correos en   │  │ Estadísticas│  │ Catálogo de  │  │
│  │ segundo plano│  │ y telemetría│  │ especialid.  │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│                                                         │
│  ┌──────────────────────────────────────────────────┐   │
│  │            🧰 utils (herramientas)               │   │
│  │ CleanAlphaNumeric, NormalizePlatformName,        │   │
│  │ ParseAndValidateEmail, SanitizeDocument,         │   │
│  │ IsStrongPassword, NormalizeMunicipioCarabobo...  │   │
│  └──────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────┐
│                  REPOSITORY LAYER (BD)                  │
│  PostgreSQL + Amazon S3 / MinIO                         │
└─────────────────────────────────────────────────────────┘
```

---

## 💡 Tips para el junior

### ¿Qué pasa si modifico un service sin cuidado?

Los services son el **corazón del sistema**. Si rompés algo acá:
- Los psicólogos no pueden iniciar sesión 🔐.
- Los admins no pueden gestionar perfiles 👑.
- Las estadísticas se corrompen 📊.
- Los correos no se envían 📧.

**Siempre corré los tests después de tocar un service:**
```bash
go test ./api/internal/service/ -v
```

### ¿Cómo sé qué service usar?

| Si querés... | Usá... |
|---|---|
| Iniciar sesión como admin | `AdminService.Login()` |
| Iniciar sesión como psicólogo | `PsiService.Login()` |
| Ver el directorio público | `PsiService.GetPublicDirectory()` |
| Ver TODOS los datos de un psi (como admin) | `PsiService.GetPsiByIDAdmin()` |
| Publicar una noticia | `PostService.CreatePost()` |
| Agregar una red social a mi perfil | `PsiService.AddSocialNetwork()` |
| Crear una especialidad nueva | `SpecialtyService.Create()` |
| El dashboard de estadísticas | `AnalyticsService.GetDashboardStats()` |
| Enviar un correo de bienvenida | `MailService.SendEmail()` |
| Importar 500 psicólogos desde Excel | `PsiService.ImportFromXLSX()` |

### Patrones que se repiten en todos los services

1. **Gatekeeping (Portero):** Primero verificá permisos, después ejecutá la acción.
2. **Key Rotation:** Al loguearse, cambiá la llave para invalidar sesiones viejas.
3. **Fire-and-Forget:** Las operaciones lentas (correos, estadísticas) van en segundo plano.
4. **Saga Rollback:** Si la BD falla después de subir una imagen a S3, borrá la imagen.
5. **Fail-Safe:** Ante la duda, mostrá lo mínimo (seguro por defecto).
6. **Mensajes de error genéricos:** Nunca digas "el usuario existe pero la contraseña está mal".
