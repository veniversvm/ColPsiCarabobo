# Funcionalidad de la aplicación — Colegio de Psicólogos del Estado Carabobo

> **Qué es este documento.** Referencia funcional de todo lo que hace la
> aplicación, organizada en las dos grandes superficies privadas:
> **Parte 1 — el agremiado (psi user)** y **Parte 2 — los administradores**.
> Cada módulo describe qué se ve en pantalla, qué se puede hacer, qué no, qué
> límites aplica y qué mensajes ve la persona.
>
> **Qué NO es.** No es el manual de usuario. Esos existen y son tutoriales paso
> a paso con capturas, generados desde Typst:
> `docs/manual-admin.pdf` y `docs/manual-psiuser.pdf` (embebidos en el binario de
> la API y servidos en `/admin/manual` y `/psi/manual`). Este documento es la
> especificación: es más largo, más denso y describe también lo que hay **debajo**
> (endpoints, permisos, cuotas, arquitectura) en los anexos.
>
> **Cómo leerlo.** Si buscas "qué puede hacer X", quédate en el cuerpo. Si buscas
> "por qué pasa esto" o "dónde está implementado", ve al anexo técnico.
>
> **Estado del código.** Todo lo descrito está verificado en el código fuente.
> Los puntos marcados como **[KI]** (*known issue*) son inconsistencias reales
> detectadas durante la redacción: están descritas en el Anexo E y **no** se han
> corregido.

---

## Índice

- [0. Introducción](#0-introducción)
  - [0.1 Qué es la aplicación](#01-qué-es-la-aplicación)
  - [0.2 Los tres actorajes](#02-los-tres-actorajes)
  - [0.3 Contexto: el sitio público](#03-contexto-el-sitio-público)
  - [0.4 Mapa de rutas](#04-mapa-de-rutas)
- [Parte 1 — El agremiado](#parte-1--el-agremiado-psi-user)
  - [1. Acceso y sesión](#1-acceso-y-sesión)
  - [2. Panel de inicio — `/psi`](#2-panel-de-inicio--psi)
  - [3. Tu Identidad Digital (perfil)](#3-tu-identidad-digital-perfil--psiperfil)
  - [4. Formación Académica](#4-formación-académica--psiacademico)
  - [5. Expediente documental](#5-expediente-documental--psidocumentos)
  - [6. Notificaciones](#6-notificaciones--psinotificaciones)
  - [7. Solicitudes (tickets)](#7-solicitudes-tickets--psitickets)
  - [8. Manual](#8-manual--psimanual)
  - [9. Qué NO puede hacer el agremiado](#9-qué-no-puede-hacer-el-agremiado)
  - [10. Límites, cuotas y validaciones](#10-límites-cuotas-y-validaciones)
- [Parte 2 — Los administradores](#parte-2--los-administradores)
  - [11. Acceso, sesión y peculiaridades](#11-acceso-sesión-y-peculiaridades)
  - [12. Roles y permisos](#12-roles-y-permisos)
  - [13. Mapa del panel](#13-mapa-del-panel)
  - [14. Dashboard](#14-dashboard--admin)
  - [15. Psicólogos](#15-psicólogos--adminpsicologos)
  - [16. Inscripciones](#16-inscripciones--admininscripciones)
  - [17. Áreas de ejercicio profesional](#17-áreas-de-ejercicio-profesional--adminareas_de_ejercicio_profesional)
  - [18. Noticias](#18-noticias--adminnoticias)
  - [19. Notificaciones](#19-notificaciones--adminnotificaciones)
  - [20. Tickets](#20-tickets--admintickets)
  - [21. Proyectos (Kanban)](#21-proyectos-kanban--adminproyectos)
  - [22. Bitácora de auditoría](#22-bitácora-de-auditoría--adminauditoria)
  - [23. Staff](#23-staff--adminstaff)
  - [24. Manual](#24-manual--adminmanual)
  - [25. Qué NO puede hacer un administrador](#25-qué-no-puede-hacer-un-administrador)
- [Anexos técnicos](#anexos-técnicos)
  - [Anexo A — Mapa de endpoints](#anexo-a--mapa-de-endpoints)
  - [Anexo B — Matriz de permisos por módulo](#anexo-b--matriz-de-permisos-por-módulo)
  - [Anexo C — Consolidado de cuotas](#anexo-c--consolidado-de-cuotas)
  - [Anexo D — Gotchas de arquitectura](#anexo-d--gotchas-de-arquitectura)
  - [Anexo E — Inconsistencias detectadas](#anexo-e--inconsistencias-detectadas)
  - [Anexo F — Glosario](#anexo-f--glosario)

---

## 0. Introducción

### 0.1 Qué es la aplicación

Plataforma web del Colegio de Psicólogos del Estado Carabobo (CPEC). Tiene tres
mitades que comparten una misma base de datos:

| Mitad | Qué es | Dónde vive |
|---|---|---|
| **Sitio público** | Portal de cara a la ciudadanía: información del gremio, directorio de profesionales, noticias, marco legal y formulario de inscripción. | Rutas públicas del frontend |
| **Portal del agremiado** (`/psi`) | Autogestión del profesional: perfil público que se publica, formación, expediente, notificaciones y solicitudes al Colegio. | `web/src/routes/psi/**` |
| **Panel administrativo** (`/admin`) | Gestión interna del Colegio: agremiados, inscripciones, publicaciones, comunicación, tickets, proyectos, bitácora y personal. | `web/src/routes/admin/**` |

**Stack técnico**

- **Frontend**: SolidStart (SolidJS + TypeScript) con renderizado SSR + CSR, Deno como runtime (`deno-server-legacy`), Tailwind v4, editor TipTap v3, flatpickr para fechas.
- **Backend**: Go 1.x con Fiber v2, arquitectura limpia `router → handler → service → repository → DB`. Documentación Swagger en `/swagger/`.
- **Datos**: PostgreSQL 18 detrás de PgBouncer (modo transacción), migraciones versionadas con Atlas, ORM GORM.
- **Archivos**: MinIO / S3. Las imágenes se recomprimen a WebP en el servidor antes de guardarse.
- **Otros**: Valkey (rate limiting), Audiobookshelf (biblioteca virtual), Resend (correo), Avispa/analytics propio.

### 0.2 Los tres actorajes

```
                    VISITANTE                    sin sesión
                        │
                        │  ve sitio público, se inscribe
                        ▼
   ┌──────────────────────────────────────────────────────────┐
   │  AGREMIADO (psi user)                    con sesión /psi  │
   │  Gestiona SU perfil, formación, expediente y solicitudes  │
   └──────────────────────────────────────────────────────────┘
                        │
                        │  el Colegio lo aprueba y crea la cuenta
                        ▼
   ┌──────────────────────────────────────────────────────────┐
   │  STAFF (administrador)                  con sesión /admin │
   │  Gestiona agremiados, inscripciones, comunicación, etc.   │
   │  con 20 permisos finos + un único SUDO                     │
   └──────────────────────────────────────────────────────────┘
```

**Regla de oro de separación de datos**: lo que un agremiado escribe en su perfil
alimenta su **ficha pública** en el directorio, pero cada campo tiene un
interruptor de visibilidad que el agremiado controla. Los datos que el Colegio
maneja (documentos del expediente, expediente deontológico, observaciones
internas, motivo de los cambios) **nunca** salen hacia el visitante.

### 0.3 Contexto: el sitio público

Sección de referencia; no es el objeto de este documento pero ubica a los dos
paneles.

| Ruta | Qué hace |
|---|---|
| `/` | Portada: banner, noticias destacadas, accesos a las secciones. |
| `/nosotros` | Información institucional (misión, visión, historia). Estática. |
| `/explorar` | Guía de navegación del sitio. Estática. |
| `/directorio` | Buscador de profesionales. Filtros por texto, área de trabajo y ubicación. Carga por scroll infinito. |
| `/directorio/[slug]` | Ficha pública del profesional: perfil, áreas, ubicaciones con badge de alcance, contacto, redes y QR para compartir. |
| `/noticias` | Listado paginado de publicaciones públicas. |
| `/noticias/[slug]` | Detalle de la noticia. El cuerpo HTML se sanea en cliente (la barrera real es el backend). |
| `/inscripcion` | Formulario público de solicitud de ingreso al Colegio. |
| `/documentos` | Marco legal: Código de Ética, Estatutos FPV, Ley de Ejercicio de la Psicología y Reglamento Interno, servidos como texto en el propio repo (`web/src/lib/documentos/*.ts`). |
| `/robots.txt`, `/sitemap.xml` | Generados dinámicamente en el servidor. El sitemap incluye las fichas del directorio y las noticias publicadas. |

**Formulario público de inscripción** (`/inscripcion`)

Es la entrada del proceso que después gestiona el admin en el módulo
*Inscripciones*. Como contexto de lo que el Colegio recibe:

- Secciones de datos personales, formación, contacto y ubicación (con selects en
  cascada de estado → municipio).
- **Verificación de unicidad en vivo** contra tres endpoints públicos:
  `/inscripcion/check-ci`, `/inscripcion/check-fpv` y `/inscripcion/check-email`.
  Marcan el campo en rojo si ya existe una solicitud pendiente.
- **Documentos obligatorios** marcados con `*` en rojo: foto, comprobante y
  copias de cédula, título y RIF, con leyenda *"Los campos marcados con * son
  obligatorios"*.
- Consulta el interruptor de recepción en `/inscripcion/status`: si el Colegio
 pausa la recepción, el formulario avisa y no permite enviar.
- Al enviar (`POST /inscripcion/submit`) la solicitud queda en estado `pending`.
  **Nadie más que el Colegio la ve**; no se envía ningún correo de confirmación al
  solicitante.

> **[KI]** La carpeta `web/src/routes/public/**` contiene 6 archivos de **0 bytes**
> (`index.tsx`, `login.tsx`, `directorio/{index,[id]}.tsx`,
> `noticias/{index,[id]}.tsx`) y ninguna ruta del sitio enlaza esas rutas. Es
> código muerto de una reorganización anterior; las rutas vivas son
> `/directorio` y `/noticias` en la raíz de `routes/`.

### 0.4 Mapa de rutas

**Público**

```
/                       /nosotros            /explorar
/directorio             /directorio/[slug]   /noticias            /noticias/[slug]
/inscripcion            /documentos          /documentos/{codigo-etica,estatutos-fpv,
                                                   ley-ejercicio-psicologia,reglamento-interno}
/login                  /forgot-password     /reset-password
/robots.txt             /sitemap.xml         /[...404]
```

**Portal del agremiado** — todas requieren sesión

```
/psi                    /psi/perfil          /psi/academico       /psi/documentos
/psi/notificaciones     /psi/manual
/psi/tickets            /psi/tickets/crear   /psi/tickets/[id]
```

**Panel administrativo** — todas requieren sesión

```
/admin-access                                       (login del staff)
/admin                  /admin/manual
/admin/psicologos      /admin/psicologos/crear      /admin/psicologos/[id]
/admin/inscripciones    /admin/inscripciones/[id]
/admin/areas_de_ejercicio_profesional    /admin/areas_de_ejercicio_profesional/crear
/admin/areas_de_ejercicio_profesional/[id]
/admin/noticias         /admin/noticias/crear       /admin/noticias/[id]
/admin/notificaciones   /admin/notificaciones/crear  /admin/notificaciones/[id]
/admin/tickets          /admin/tickets/configuracion /admin/tickets/[id]
/admin/proyectos        /admin/proyectos/crear      /admin/proyectos/[id]
/admin/auditoria
/admin/staff            /admin/staff/crear          /admin/staff/[id]
```

> `web/src/routes/admin/dashboard/index.tsx` existe pero es un stub de 5 líneas
> que el menú no enlaza: el dashboard real vive en `/admin`. **[KI]**

---
---

# Parte 1 — El agremiado (psi user)

Todo lo que un profesional agremiado puede hacer desde el portal `/psi`.

**Principio de diseño del portal**: el agremiado es dueño de su **presencia
pública** (foto, bio, áreas, contacto, ubicaciones, redes) y de su **formación
académica** (posgrados). En cambio, es **espectador** de todo lo que el Colegio
registra sobre él: el título de pregrado, el número de registro profesional, los
documentos del expediente, el expediente deontológico y las observaciones
internas.

## 1. Acceso y sesión

### 1.1 Iniciar sesión — `/login`

Dos campos, ambos obligatorios:

| Campo | Descripción |
|---|---|
| **Identificación** | Nombre de usuario o correo electrónico. Se normaliza a minúsculas y sin espacios. |
| **Contraseña** | Con botón de ojo para mostrar/ocultar. |

**Qué pasa al enviar**: el portal valida las credenciales, guarda un token en
una cookie invisible al navegador y además una copia por pestaña, y entra al
panel. El botón muestra *"VERIFICANDO…"* mientras tanto.

**Si falla**:

- Credenciales incorrectas o cuenta suspendida → *"Credenciales inválidas"* /
  *"Cuenta inactiva o suspendida"*.
- Si se han hecho demasiados intentos → el mensaje real del servidor con el
  **tiempo de espera concreto** que queda.
- Cualquier otro fallo → *"Ocurrió un error inesperado"*.

**Límite de intentos**: 15 envíos por cada 5 minutos desde la misma conexión.

**Consecuencia importante**: cada inicio de sesión **invalida los inicios de
sesión anteriores** del mismo agremiado. Si alguien abre sesión en otro
dispositivo, la sesión del primero deja de servir. Es una medida anti-replay
deliberada.

### 1.2 Recuperar contraseña — `/forgot-password`

Un solo campo: **correo registrado**. La respuesta es **siempre la misma**,
exista o no esa cuenta en el sistema, para no revelar qué correos están
registrados.

Si el correo existe, el Colegio envía un enlace que:

- expira en **1 hora**;
- se puede usar **una sola vez**;
- lleva a `/reset-password?token=...`.

El enlace se guarda en la base de datos **solo como hash**, nunca en claro.

### 1.3 Cambiar la contraseña por enlace — `/reset-password`

Dos campos: nueva contraseña y su confirmación.

- Si el enlace no trae token, la pantalla avisa *"Enlace inválido"* y ofrece
  pedir uno nuevo.
- Las dos contraseñas deben coincidir y tener al menos 8 caracteres.
- Al enviarlo, **el agremiado entra automáticamente** a su portal: no tiene que
  volver a iniciar sesión.

### 1.4 Cómo funciona la sesión por dentro

| Pieza | Detalle |
|---|---|
| Duración | 24 horas |
| Dónde vive el token | Cookie invisible (para el servidor) + copia en la memoria de la pestaña (para el navegador) |
| Verificación | Cada 60 segundos el portal pregunta al servidor si la sesión sigue viva |
| Cierre | El botón de salir invalida **todas** las sesiones abiertas de ese agremiado |
| Pestaña nueva o navegador reiniciado | La sesión **se recupera sola**: el portal lee la cookie y reconstruye su copia sin pedir login otra vez |

**Cambiar la contraseña desde el perfil también reinicia la sesión** (por
seguridad) y el portal lo resuelve solo, sin expulsar al agremiado.

## 2. Panel de inicio — `/psi`

Es la pantalla de aterrizaje del portal. Tiene tres bloques.

### 2.1 Saludo y estado de solvencia

Muestra el nombre del agremiado y una tarjeta de **Estatus de Solvencia**:

| Estado | Cómo se ve |
|---|---|
| **AL DÍA** | Fondo verde, marca de verificación. |
| **INSOLVENTE** | Fondo rojo, signo de advertencia. |

Aparece también el número de FPV (Ficha de Psiquiólogo Veterinario) o el número
de cédula según corresponda.

### 2.2 Biblioteca virtual

Tarjeta con acceso a la biblioteca de audio-libros del Colegio (plataforma
Audiobookshelf).

- **Solo disponible para agremiados solventes.** Si no lo es, el botón aparece
  bloqueado con un candado y el texto *"Ponte al día con el gremio para acceder
  a la biblioteca virtual"*.
- Al pulsarlo se abre en una pestaña nueva la sesión personal ya iniciada en la
  biblioteca. **La clave nunca pasa por la web**: se genera en el servidor.
- Si la biblioteca no responde, el portal avisa y reintenta después.

### 2.3 Acceso rápido

Seis tarjetas: Perfil, Académico, Documentos, Notificaciones (con contador de
no leídas), Solicitudes y Manual.

Debajo, una barra inferior con Panel, Notificaciones, Directorio y Perfil.

## 3. Tu Identidad Digital (perfil) — `/psi/perfil`

La pantalla central del portal: **"Tu Identidad Digital"**. Un cuaderno de ocho
pestañas más dos bloques fuera del cuaderno.

Antes de las pestañas hay un **selector de foto de perfil** (fuera del formulario)
y al final, fuera del cuaderno, el bloque de **Redes Sociales**.

> **Un solo botón guarda todo.** Debajo del cuaderno hay una barra con el botón
> *Guardar* y, al lado, un campo de **contraseña actual obligatorio**. No se puede
> guardar ninguna pestaña sin escribir la contraseña: es la confirmación de que
> es el propio agremiado quien está al otro lado. La barra se oculta en la
> pestaña de emergencia (que tiene su propio guardado).

### 3.1 Pestaña 1 — Cuenta y Seguridad

| Campo | Editable | Notas |
|---|---|---|
| Nombre de usuario (para el login) | Sí | Se identifica en la base de datos como único. |
| Correo principal (institucional) | Sí | Es el correo al que el Colegio envía comunicados. |
| Contraseña actual | Sí | **Obligatoria para guardar cualquier cosa.** |
| Nueva contraseña | Sí | Opcional. Mínimo 8 caracteres, sin espacios, con mayúscula, minúscula, número y símbolo. |
| Confirmar nueva contraseña | Sí | Debe coincidir con la anterior. |

Al cambiar la contraseña, el portal renueva la sesión automáticamente y muestra
los requisitos de robustez como lista de comprobación.

### 3.2 Pestaña 2 — Información de Contacto

| Campo | Uso |
|---|---|
| Email de contacto (gremial) | Correo distinto al institucional, para consultas del público. |
| Teléfono fijo / local (gremial) | Línea fija de consulta. |
| Teléfono móvil / WhatsApp (gremial) | Celular, el canal más usado. |

Los tres son públicos **solo si el agremiado lo autoriza** (ver pestaña 7).

### 3.3 Pestaña 3 — Expediente Académico (pregrado)

Esta pestaña es **mitad de lectura, mitad de escritura**:

**Lo que el agremiado VE pero NO puede editar** (lo registra el Colegio):

| Dato | Origen |
|---|---|
| Título de pregrado | Colegio |
| Universidad | Colegio |
| Fecha de graduación | Colegio |
| Mención | Colegio |
| Número de registro | Colegio |
| Folio | Colegio |
| Tomo | Colegio |
| Fecha del título | Colegio |
| Estado del registro | Colegio |

**Lo que el agremiado SÍ puede hacer**: subir hasta **3 soportes** de su título
— *Imagen del Título*, *Documento Adicional 1* y *Documento Adicional 2*.

Cada soporte muestra: la imagen actual, si se ha subido, y tres estados posibles
— *"Subida"* (verde), *"Se eliminará al guardar"* (rojo, con botón *Cancelar*) o
la vista previa del archivo nuevo. Al hacer clic se abre en grande.

> **Consecuencia práctica**: desde el portal el agremiado puede **sustituir o
> quitar** un soporte del título, pero **no puede corregir** los datos del
> registro profesional. Para eso debe acudir al Colegio.

### 3.4 Pestaña 4 — Ubicación Geográfica

Tres bloques, porque el Colegio clasifica el alcance de cada consultorio:

**Consultorio en Carabobo**

| Campo | Etiqueta en pantalla |
|---|---|
| Municipio | Desplegable en cascada (los municipios del estado seleccionado) |
| Teléfono fijo de consulta | |
| Celular de consulta | |
| Dirección de consultorio en Carabobo | |

**Consultorio fuera de Carabobo (Venezuela)**

| Campo | Etiqueta |
|---|---|
| Estado | Desplegable |
| Ciudad / Municipio | Desplegable en cascada |
| Teléfono fijo | |
| Celular | |
| Dirección de consultorio secundario | |

**Consultorio en el exterior**

| Campo | Etiqueta |
|---|---|
| País | |
| Teléfono internacional | |
| Celular / Móvil | |
| Dirección en el Exterior | |

En la ficha pública cada una de estas ubicaciones aparece **una debajo de la
otra** con una etiqueta de alcance: *Carabobo* (azul), *Venezuela* (índigo) o
*Internacional* (verde).

### 3.5 Pestaña 5 — Perfil Profesional

| Campo | Detalle |
|---|---|
| Área de trabajo principal | Desplegable con el **catálogo oficial** del Colegio. |
| Área de trabajo secundaria | Desplegable con el mismo catálogo. |
| Mini-bio | Texto corto, **máximo 250 caracteres**, con contador. Se muestra en la tarjeta del directorio. |
| Perfil completo | Texto largo con formato (negrita, listas, enlaces). Límite-orientativo de 5.000 palabras, con contador que avisa en rojo al superarlo. |

> **Las áreas de trabajo nunca se escriben a mano**: el desplegable solo ofrece las
> áreas que existen en el catálogo del Colegio, y el directorio público solo las
> muestra si además el agremiado está solvente. Nunca se muestra un área
> inventada.

### 3.6 Pestaña 6 — Servicio y Preferencias

**Modalidad de servicio** (se pueden marcar varias):

- Presencial
- A distancia (en línea)
- Telefónica

**Preferencias**:

- *Mostrar mi modalidad en el directorio público* — decide si la modalidad de
  atención se publica en su ficha.
- *Autorizar aviso de cumpleaños a la administración* — si está apagado, el
  agremiado **no aparece** en el banner de cumpleaños del panel administrativo,
  aunque cumpla años.

### 3.7 Pestaña 7 — Privacidad y Visibilidad

Dieciséis interruptores, agrupados por bloque. Controlan **uno por uno** qué datos
se publican en la ficha del directorio. Todo apagado = perfil de identidad sin
contacto ni ubicación.

| Bloque | Interruptores |
|---|---|
| **Carabobo** | Mostrar email de contacto · Mostrar dirección de consulta · Mostrar municipio · Mostrar teléfono fijo · Mostrar celular |
| **Fuera de Carabobo** | Mostrar estado · Mostrar municipio/ciudad · Mostrar teléfono fijo · Mostrar celular · Mostrar dirección de consulta |
| **Exterior** | Mostrar teléfono fijo · Mostrar celular · Mostrar dirección exterior |
| **Académico** | Mostrar universidad · Mostrar fecha de grado · Mostrar mención |

### 3.8 Pestaña 8 — Contacto de Emergencia

Bloque de gestión propia, **fuera del formulario general** (por eso el botón
*Guardar* no aparece aquí).

Es un listado de **hasta 3 personas** a las que el Colegio puede avisar si no
logra localizar al profesional o ocurre un accidente.

**Datos de cada contacto**:

| Campo | Regla |
|---|---|
| Nombre o Cédula | Entre 2 y 255 caracteres. |
| Parentesco / relación | Obligatorio, hasta 100 caracteres. Catálogo de 9 opciones (madre, padre, cónyuge, hermano, hijo, etc.) más la opción *Otro* con texto libre. |
| Teléfono | Hasta 20 dígitos. |
| Correo | Opcional, con formato válido. |
| — | **Se exige al menos un canal de contacto**: teléfono o correo. |

**Acciones**: añadir, editar y eliminar. Al llegar al máximo, el aviso es
*"Alcanzaste el máximo de 3 personas. Edita o elimina una para registrar otra."*

> **Privacidad**: estos datos son de un tercero y **nunca se publican**. No
> aparecen en el directorio, ni en la ficha pública, ni en el buscador del sitio.
> La pantalla lo dice de forma explícita: *"Son datos privados: nunca se publican
> en el directorio"*.

### 3.9 Redes Sociales (fuera del cuaderno)

Listado de enlaces (nombre + URL) que se publican bajo el código QR de la ficha
pública. El Colegio limita a **10 redes** por agremiado.

- **Añadir**: nombre y URL son obligatorios.
- **Eliminar**: pide confirmación.

> **[KI]** Si una de estas dos operaciones falla, el portal **no muestra error**:
> la lista simplemente no cambia. El agremiado no se entera de que su red no se
> guardó.

### 3.10 Cómo se guarda el perfil

Todo el perfil se envía en **una sola operación** al pulsar *Guardar*:

1. Se comprueba la contraseña actual. Sin ella, no se envía nada.
2. Se empaquetan los campos de texto, los 21 interruptores (siempre con valor
   sí/no, por eso sí se pueden apagar) y hasta 4 archivos.
3. El servidor limpia y sanea cada campo: teléfonos quedan solo con dígitos,
   el texto pierde caracteres peligrosos, el perfil completo se sanea
   como HTML, la mini-bio se recorta a 250 caracteres.
4. Se guarda y se muestra el banner *"Perfil actualizado correctamente."*

> **Limitación conocida**: si un agremiado **borra** un campo dejándolo en blanco,
> el sistema **no lo borra**: el valor anterior se conserva. Apagar sí funciona
> (los interruptores van siempre); vaciar un texto, no. Para dejar un campo en
> blanco hay que solicitarlo al Colegio.

> **[KI]** Los errores al añadir o borrar redes sociales no se muestran al
> usuario (ver 3.9).

## 4. Formación Académica — `/psi/academico`

Pant independiente del perfil dedicada a los **posgrados**.

**Alta de un posgrado**:

| Campo | Obligatorio | Notas |
|---|---|---|
| Título obtenido | Sí | |
| Universidad | Sí | |
| Año de egreso | Sí | |
| Breve descripción | No | Texto libre. |
| Soporte 1 — *Título* | No | Archivo. |
| Soporte 2 — *Notas* | No | Archivo. |
| Soporte 3 — *Extra* | No | Archivo. |

**Acciones**: crear, editar y eliminar. Cada posgrado se muestra en una tarjeta
con su título, universidad, año, descripción y las miniaturas de los soportes
(cada una abre en una pestaña nueva). La pantalla aclara que el posgrado *quedará
visible en tu perfil público*.

**Al editar**, si no se vuelven a elegir los archivos, los soportes anteriores
se conservan (no se borran). Al eliminar un posgrado, se borra también su
certificado del almacenamiento.

> **[KI]** La etiqueta de la pantalla dice *"Soportes Digitales (PDF o Imagen)"* y
> el servidor acepta ambos, pero el selector de archivos del navegador solo
> permite elegir imágenes: **un agremiado no puede subir un PDF** pese a que la
> interfaz lo promete.

## 5. Expediente documental — `/psi/documentos`

**Solo lectura.** No hay forma de subir, cambiar ni borrar nada desde aquí.

Muestra los documentos que el Colegio tiene registrados en el expediente del
agremiado, agrupados por tipo con un contador por grupo:

| Tipo | Documento |
|---|---|
| `cedula` | Copia de la cédula |
| `titulo` | Título de grado |
| `rif` | RIF |
| `solvencia` | Comprobante de solvencia |
| `comprobante` | Otro comprobante |
| `otro` | Otros documentos |

Cada documento se muestra con miniatura y enlace de descarga. Solo se admite un
documento por categoría; para corregirlo hay que dirigirse al Colegio.

## 6. Notificaciones — `/psi/notificaciones`

Listado de los comunicados que el Colegio ha dirigido al agremiado.

**Características**:

- **Abrir una notificación NO la marca como leída.** El contenido se despliega en
  un panel y hay que pulsar explícitamente *Marcar como leída*.
- El marcado es **inmediato**: la notificación se ve como leída al instante y, si
  el servidor rechaza el cambio, **vuelve a marcarse como no leída**.
- El contador de no leídas se actualiza en el panel, en el acceso rápido del
  inicio y en la barra inferior.
- **Carga progresiva**: al llegar al final se piden las siguientes 20 y se
  acumulan. Hay un esqueleto de carga mientras tanto.
- Las fechas se muestran en la zona horaria de Caracas.
- Un **sonido** suena cuando llega una notificación nueva, solo si la pestaña
  está visible (para no molestar con notificaciones en segundo plano).

## 7. Solicitudes (tickets) — `/psi/tickets`

El canal formal para que el agremiado consulte al Colegio. Lo que el Colegio
puede ofrecer como motivos (solicitudes, preguntas, otros) y los estados de cada
motivo los define el propio Colegio en el panel administrativo.

### 7.1 Ver mis solicitudes — `/psi/tickets`

Listado paginado con título, estado (con color) y fecha. Al entrar a una
solicitud que no es suya, el portal responde *"No puede acceder a esta
solicitud"* y lo trata igual que si no existiera.

### 7.2 Crear una solicitud — `/psi/tickets/crear`

| Campo | Regla |
|---|---|
| Motivo | Obligatorio. Debajo se indica **cuántas solicitudes abiertas permite ese motivo** ("Por este motivo puedes tener hasta N solicitudes abiertas a la vez"). |
| Título | Obligatorio, hasta **200 caracteres**, con contador. |
| Descripción | Obligatoria, hasta **2.000 caracteres**, con contador. |
| Archivos | Opcional, varios a la vez. |

**Si el Colegio tiene la recepción de solicitudes pausada**, el portal lo avisa
antes de dejar enviar, con el mensaje que el Colegio haya configurado. Si el
pausa se activa entre la lectura y el envío, el servidor lo rechaza y el portal
lo explica.

**Al crear**, la solicitud:

1. Se registra en el motivo elegido, en el primer estado no cerrado que el
   Colegio definió para ese motivo.
2. La descripción se convierte en el **primer mensaje** de la conversación, con
   los archivos adjuntos.
3. El portal abre la conversación.

### 7.3 Conversación y cierre — `/psi/tickets/[id]`

**Enviar un mensaje**:

| Regla | Valor |
|---|---|
| Caracteres | Hasta **1.000** por mensaje |
| Mensajes seguidos | **Máximo 3** sin respuesta del Colegio |
| Si se supera | *"Espere la respuesta del colegio antes de enviar otro mensaje"* |

El mensaje se suma a la conversación **al instante, sin recargar la página**, y
se pueden adjuntar archivos.

**Cerrar la solicitud**: botón con un motivo obligatorio de hasta **500
caracteres**. Al cerrar, la solicitud pasa al primer estado de cierre que el
Colegio definió para ese motivo.

**Historial**: debajo del hilo se ven los cambios de estado con su fecha y su
motivo.

> Si la solicitud no existe, no es del agremiado, o el servidor falla, el portal
> muestra siempre lo mismo: *"Solicitud no encontrada. Puede que no exista o que
> no te pertenezca."*

## 8. Manual — `/psi/manual`

Tarjeta con el **manual de usuario del portal del psicólogo** en PDF, integrado
en la propia aplicación y protegido por sesión (nadie puede descargar el archivo
sin estar registrado como agremiado).

- Se abre embebido dentro de la página.
- Botón **Descargar PDF**.
- Si la sesión ya no es válida, avisa que hay que volver a iniciar sesión.

## 9. Qué NO puede hacer el agremiado

| Restricción | Detalle |
|---|---|
| Editar el título de pregrado | Universidad, fecha, mención y datos de registro los gestiona el Colegio. |
| Subir o cambiar documentos del expediente | `/psi/documentos` es de lectura; el registro lo hace el Colegio. |
| Editar sus redes sociales | Puede añadirlas y borrarlas, pero no modificarlas. |
| Ver su propio estado de solvencia en el directorio | La solvencia es información interna; el visitante no la ve. |
| Ver sus propios tickets si no son suyos | Se rejectsan con mensaje genérico. |
| Publicar noticias o avisos | La comunicación es del Colegio. |
| Crear solicitudes si el Colegio pausó la recepción | El portal avisa y bloquea. |
| Enviar más de 3 mensajes seguidos en un ticket | Debe esperar respuesta del Colegio. |
| Registrarse a sí mismo como agremiado | El alta la hace el Colegio al aprobar una inscripción. |
| Borrar un campo de su perfil dejándolo en blanco | Solo puede *ocultarlo* con los interruptores. |
| Poner más de 3 contactos de emergencia | El portal avisa del límite. |
| Poner más de 10 redes sociales | El servidor rechaza. |
| Subir un PDF como soporte de posgrado | El selector del navegador solo admite imágenes. |

## 10. Límites, cuotas y validaciones

### 10.1 Contenido

| Límite | Valor | Dónde se aplica |
|---|---|---|
| Mini-bio | 250 caracteres | Recortado en el servidor |
| Perfil completo | 5.000 palabras | Aviso visual, no bloquea el guardado |
| Mensaje en una solicitud | 1.000 caracteres | Servidor + contador en pantalla |
| Mensajes seguidos del agremiado | 3 | Servidor |
| Título de solicitud | 200 caracteres | Servidor + contador |
| Descripción de solicitud | 2.000 caracteres | Servidor + contador |
| Motivo de cierre de solicitud | 500 caracteres | Servidor + contador |
| Solicitudes abiertas simultáneas | Las que fije el Colegio por motivo (1 a 50) | Servidor |
| Postgrados | Sin límite | — |
| Redes sociales | 10 | Servidor |
| Contactos de emergencia | 3 | Servidor + cliente |
| Nombre de contacto de emergencia | 2 a 255 caracteres | Servidor + cliente |
| Parentesco del contacto | 1 a 100 caracteres | Servidor + cliente |
| Teléfono del contacto | 20 dígitos | Servidor + cliente |

### 10.2 Archivos

| Tipo | Límite del navegador | Límite real del servidor |
|---|---|---|
| Foto de perfil | JPG/PNG/GIF, 5 MB | Se recomprime a WebP: 800 px, 150 KB |
| Soportes de título (3) | — | Imagen: 1.600 px, 400 KB · PDF: 4 MB |
| Soportes de posgrado (3 por posgrado) | Solo imágenes (el `accept` bloquea el PDF) | Imagen: 1.600 px, 400 KB · PDF: 4 MB |
| Archivos de solicitud (sin tope de cantidad) | — | Imagen: 1.600 px, 400 KB · PDF: 4 MB |

Todas las imágenes se **recomprimen a WebP** en el servidor, se aplana el
transparente sobre blanco y se guardan con un tamaño fijo. Los PDF se validan
por su firma de archivo, no por su extensión.

### 10.3 Sesión y seguridad

| Aspecto | Valor |
|---|---|
| Duración de la sesión | 24 horas |
| Intentos de inicio de sesión | 15 por cada 5 minutos |
| Recuperación de contraseña | Enlace de 1 hora, de un solo uso, guardado como hash |
| Contraseña | Mínimo 8 caracteres, sin espacios, con mayúscula, minúscula, número y símbolo |
| Verificación de sesión | Cada 60 segundos |
| Cierre de sesión | Invalida todas las sesiones abiertas del agremiado |
| Ventana de sesión de la biblioteca | 30 días (sesión independiente, en Audiobookshelf) |
| Límite general de la API | 60 peticiones por minuto y conexión, en toda la aplicación |

---
---

# Parte 2 — Los administradores

Personal del Colegio que gestiona la institución. Todo bajo `/admin`, detrás de
`/admin-access`.

**Modelo mental**: los administradores **no tienen roles fijos**. Cada persona
tiene una matriz de **20 permisos individuales** que el Colegio asigna. Existen
5 perfiles preconfigurados como punto de partida (Secretaría, Comunicación,
Soporte, Proyectos, Lector), pero cualquier combinación es válida y se guarda
como *Personalizado*. Existe además un único **SUDO** (Super Usuario) que puede
todo y que puede transferir ese poder a otra persona.

## 11. Acceso, sesión y peculiaridades

### 11.1 Iniciar sesión — `/admin-access`

Usuario y contraseña. A diferencia del portal del agremiado, el login del
staff es **muy restrictivo**: **5 intentos por cada 30 minutos** desde la misma
conexión. Agotado el límite, el mensaje indica el tiempo exacto restante.

> En pruebas automatizadas o al depurar, cuidado: encadenar varios inicios de
> sesión con la misma cuenta agota el límite y bloquea el panel durante media
> hora.

### 11.2 Sesión

Mecánicamente igual a la del agremiado (cookie invisible + copia por pestaña,
verificación cada 60 segundos, recuperación automática al abrir una pestaña
nueva), con cuatro diferencias que hay que tener presentes:

1. **Cada inicio de sesión del staff invalida sus sesiones anteriores**, en
   todas las pestañas y todos los dispositivos. Consecuencia visible: si alguien
   vuelve a iniciar sesión en otra pestaña, la primera deja de responder.
2. **Cambiar la contraseña de un miembro del personal** invalida todas sus
   sesiones. Si un administrador cambia la suya, se cae a sí mismo en las demás
   pestañas (y el portal lo reconnecta en la que está escribiendo).
3. **Cerrar sesión** invalida todas las sesiones de ese administrador.
4. **Desactivar** a un administrador (`is_active = false`) **no** le cierra la
   sesión actual: solo impide que vuelva a iniciar sesión.

### 11.3 El detalle que más confunde al personal: el error "no existe"

Cuando una acción del panel responde *"Cannot POST /api/v1/admin/..."* o
*"Cannot PATCH ..."*, **no es que la función no exista**. Casi siempre significa
que **la petición llegó sin credencial válida** y el sistema la disfraza de
"página no encontrada" para no revelar qué rutas existen.

Regla práctica: si una acción del panel falla con ese mensaje, el problema está
en la sesión, no en la función.

Se reservaron dos rutas con respuesta real de "no autorizado" (`/session/me` y
`/session/validate`) para que el portal pueda distinguir "tu sesión caducó" de
"esta función no existe". Por eso existe el botón *Reintentar* en la pantalla de
error y no hace falta volver a iniciar sesión.

### 11.4 Pantallas que se degradan sin romper el panel

Si un módulo falla al cargar (por ejemplo, la API devuelve un error puntual), el
panel **no se cae a pantalla completa**: muestra una tarjeta de error dentro del
contenido, con el detalle técnico, un botón *Reintentar* y un enlace *Volver al
panel*. El menú lateral sigue funcionando. Los listados además conservan el
último resultado que se cargó bien.

## 12. Roles y permisos

### 12.1 Los 20 permisos

Se agrupan en ocho bloques en la interfaz de edición:

| Bloque | Permisos |
|---|---|
| **Gestión de Colegiados** | Ver · Crear · Editar · Eliminar |
| **Gestión de Staff** | Crear · Editar · Eliminar |
| **Publicaciones** | Publicar · Editar · Eliminar |
| **Notificaciones** | Enviar · Gestionar · Ver |
| **Áreas de ejercicio** | Crear · Editar · Eliminar |
| **Proyectos** | Gestionar |
| **Tickets** | Gestionar |
| **Auditoría** | Ver · Exportar |

> **[KI]** `can_delete_publish` (Eliminar publicaciones) existe en el modelo, en
> la interfaz de permisos y habilita el acceso al módulo de Noticias, pero
> **ningún servicio del backend lo consulta**. El borrado real de una noticia no
> existe: solo se archiva, y archivar requiere el permiso de editar. Ver 18.2.

### 12.2 Los 5 perfiles preconfigurados

Al crear o editar un miembro del personal se puede partir de un perfil y luego
ajustar la matriz a mano:

| Perfil | Qué puede hacer |
|---|---|
| **Secretaría** | Ver, crear y editar colegiados; ver notificaciones. |
| **Comunicación** | Publicar y editar noticias; enviar y gestionar notificaciones; crear áreas de ejercicio. |
| **Soporte** | Ver colegiados y notificaciones; gestionar las solicitudes de los agremiados. |
| **Proyectos** | Gestionar los tableros de proyectos. |
| **Lector** | Ver colegiados y notificaciones. Sin escribir nada. |
| *(Personalizado)* | Cuando la matriz no coincide exactamente con ninguno de los anteriores. |

### 12.3 El SUDO

Existe **un único Super Usuario**. Puede hacer todo, saltándose cualquier
comprobación de permisos.

Tres cosas que solo puede hacer el SUDO:

1. **Modificar los interruptores globales** de recepción de solicitudes y
   inscripciones. Cualquier otro administrador ve la tarjeta pero no puede
   tocarla, con el aviso *"Solo el Super Usuario puede modificar esta
   configuración."*
2. **Ceder el SUDO** a otra persona (§23.5).
3. **Purgar la bitácora** por antigüedad. No hay ni botón ni permiso: es una
   tarea programada que borra lo que tenga más de 90 días.

Al arrancar, el sistema **fuerza la matriz completa de permisos en `true`** para
el SUDO, de modo que nunca se queda sin acceso a nada aunque alguien edite la
base de datos a mano.

### 12.4 Reglas que impiden la escalada de privilegios

Son defensas en el servidor, no en la interfaz:

| Intento | Resultado |
|---|---|
| Crear otro SUDO desde el panel | **Imposible**: el campo SUDO se fuerza siempre en falso al crear. |
| Dar a otro un permiso que uno mismo no tiene | Rechazado: *"no puedes otorgar el permiso: …"* |
| Editar o eliminar a un SUDO | Rechazado: *"no puedes editar/eliminar un Super Usuario"*. |
| Eliminar la propia cuenta | Rechazado: *"no puedes eliminar tu propia cuenta"*. |
| Quitarse a uno mismo el permiso que le permite administrar | El sistema compara el rango y lo rechaza: *"no tienes rango para modificar: …"*. |
| Acceder a un módulo sin permiso | El módulo no aparece en el menú **y** el servidor rechaza la acción. |

## 13. Mapa del panel

Once módulos en el menú lateral. El menú **se arma según los permisos** de quien
entra: si un módulo no está autorizado, no aparece.

| # | Módulo | Ruta | Quién lo ve |
|---|---|---|---|
| 1 | **Dashboard** | `/admin` | Todos |
| 2 | **Psicólogos** | `/admin/psicologos` | Con permisos de colegiados |
| 3 | **Inscripciones** | `/admin/inscripciones` | Con permisos de colegiados |
| 4 | **Áreas de Ejercicio** | `/admin/areas_de_ejercicio_profesional` | Con permisos de áreas |
| 5 | **Noticias** | `/admin/noticias` | Con permisos de publicaciones |
| 6 | **Notificaciones** | `/admin/notificaciones` | Con permisos de notificaciones |
| 7 | **Tickets** | `/admin/tickets` | Con permisos de tickets |
| 8 | **Proyectos** | `/admin/proyectos` | Con permisos de proyectos |
| 9 | **Auditoría** | `/admin/auditoria` | Con permiso de ver bitácora |
| 10 | **Staff** | `/admin/staff` | Con permisos de personal |
| 11 | **Manual** | `/admin/manual` | Todos |

**Sobre el menú y la sidebar**

- La barra lateral se puede **colapsar** a un modo estrecho (solo iconos).
- La barra superior muestra la sección actual y permite salir de la sesión.
- Un **contador de solicitudes pendientes** parpadea en el icono de Tickets; se
  refresca cada 30 segundos y se satura en "99+".

## 14. Dashboard — `/admin`

Vista general de la operación del Colegio. Los datos se recargan cada 15
minutos.

### 14.1 Tres bloques de indicadores

**Inicios de sesión**: total, hoy, esta semana, este mes, y cuántos usuarios
distintos entraron hoy.

**Visitas al sitio**: total, hoy, esta semana, visitantes únicos de hoy y de la
semana, y búsquedas (total, hoy, semana).

**Perfiles**: vistas de perfil totales, de hoy y de la semana, más las **sesiones
activas en este momento** (destacadas en una banda superior).

### 14.2 Tendencias

Dos gráficas de línea con los últimos **14 días**: una de inicios de sesión y
otra de visitas al sitio.

### 14.3 Rankings

| Ranking | Cantidad |
|---|---|
| Áreas de trabajo más buscadas | 11 |
| Municipios más visitados | 10 |
| Términos de búsqueda más usados | 10 |
| Perfiles de agremiados más vistos | 10 |

### 14.4 Banner de cumpleaños

Muestra los agremiados que cumplen años **hoy o durante la semana**, con su
nombre. Solo aparecen los que **autorizaron** el aviso en su perfil
(pestaña *Servicio y Preferencias*), de modo que la lista respeta su
preferencia de privacidad.

### 14.5 Interruptores de recepción

Tarjeta con dos filas:

| Interruptor | Efecto cuando está apagado |
|---|---|
| **Solicitudes de agremiados** | El portal `/psi/tickets` deja de permitir abrir nuevas solicitudes. Las existentes se siguen gestionando con normalidad. |
| **Inscripciones públicas** | El formulario público `/inscripcion` deja de aceptar nuevos envíos. |

Cada fila tiene un interruptor y un **mensaje público** (hasta 500 caracteres)
que se escribe **solo cuando el interruptor está apagado**: es el texto que verá
el agremiado o el visitante que intente enviar. Botón *Guardar* por fila y
confirmación efímera *✓ Guardado*.

Al apagar o encender cada interruptor queda registrado quién lo hizo, cuándo y
qué mensaje había antes y después. **Solo el SUDO puede modificarlos.**

## 15. Psicólogos — `/admin/psicologos`

El módulo central: el padrón de agremiados.

### 15.1 Listado

**Búsqueda y filtros** (se combinan entre sí):

| Filtro | Valores |
|---|---|
| Texto libre | Nombre, cédula, FPV o área de desempeño (con espera de medio segundo al escribir). |
| Solvencia | Todos · Solvente · Insolvente |
| Estatus | Todos · Activo · Inactivo |
| Género | Todos · Masculino · Femenino |
| Área de ejercicio | Todas · (catálogo del Colegio) |

**Tabla**, 6 columnas:

| Columna | Contenido |
|---|---|
| N.º de Control | El número interno del Colegio. |
| Colegiado | Nombre y apellido. |
| Credenciales | FPV, cédula y **edad calculada** (la calcula el servidor a partir de la fecha de nacimiento). |
| Solvencia | Semáforo. |
| Estatus | Activo / Inactivo. |
| Acciones | Abre la ficha. |

Paginación arriba y abajo con selector de cantidad por página; al cambiar de
página la vista vuelve arriba con desplazamiento suave.

> No hay ordenamiento por columna ni acciones masivas (seleccionar varios y
> aplicar una acción en bloque).

### 15.2 Alta de agremiado — `/admin/psicologos/crear`

Formulario en cinco secciones:

| Sección | Contenido |
|---|---|
| **Cuenta** | Usuario, correo, contraseña inicial. |
| **Identidad legal** | Cédula, FPV, nombres, apellidos, nacionalidad, fecha de nacimiento, género. |
| **Registro académico** | Título, universidad, fecha de graduación, mención, número de registro, folio, tomo, fecha del título, estado. |
| **Contacto** | Correo y teléfonos de contacto. |
| **Estatus institucional** | Solvencia, número de control, avisos. |

**Validaciones**: la contraseña cumple la política de robustez; la cédula y el
FPV deben ser mayores que cero. El servidor vuelve a comprobar la unicidad de
cédula, FPV y correo antes de guardar.

**Al crear**, el sistema: genera el número de control, envía un correo de
bienvenida, y si el agremiado nace **solvente** le provisiona automáticamente
el acceso a la biblioteca virtual.

> **[KI]** Si el servidor rechaza el alta, el panel muestra un mensaje genérico
> y **no muestra qué campo concreto causó el rechazo**.
>
> **[KI]** La sección de contacto escribe el teléfono en un campo que no existe
> en el modelo de datos (`public_phone` en lugar de `contact_phone`): ese
> teléfono **no se guarda**.
>
> **[KI]** El formulario declara los campos de área de trabajo principal y
> secundaria, pero **no dibuja ningún campo para ellos**: quedan vacíos en el
> alta y el agremiado debe elegirlas después desde su portal.

### 15.3 Importación masiva

Botón en la cabecera del listado (el botón dice *"Importar CSV"* pero el archivo
es un Excel **[KI]**).

**Modal de 4 pasos**: elegir archivo → confirmar → subir → resultado.

| Requisito | Valor |
|---|---|
| Formato | `.xlsx` / `.xls` (se valida extensión **y** tipo real del archivo) |
| Tamaño máximo | 5 MB |
| Hoja obligatoria | *"BD ColPsiCarabobo 2026"* |
| Filas saltadas | Las 2 primeras (encabezados) |
| Columnas leídas | 48 |

**Al importar**, el sistema: genera una contraseña para cada agremiado nuevo (aleatoria fuera del entorno de desarrollo), inventa un correo sintético para los que no traen, deduce la solvencia inicial de la última solvencia registrada, e importa también sus posgrados. Se genera un log de la importación.

**Resultado**: pantalla con *importados*, *fallidos* y una tabla de errores
detallando **fila, nombre, cédula, FPV y el motivo** de cada fallo.

### 15.4 Ficha del agremiado — `/admin/psicologos/[id]`

La pantalla más completa del panel: **dos cuadernos de pestañas**.

#### Cuaderno A — Expediente (8 pestañas)

| Pestaña | Contenido |
|---|---|
| **Cuenta** | Usuario, correo, estado de la cuenta. |
| **Estatus** | Activación del agremiado, datos de estado institucional. |
| **Solvencias** | Registro y renovación de solvencias. |
| **Identidad** | Cédula, FPV, nombres, apellidos, nacionalidad, nacimiento, género. |
| **Contacto** | Correos y teléfonos, y su visibilidad. |
| **Ubicación** | Los tres bloques de consultorio (Carabobo / Venezuela / exterior). |
| **Perfil** | Áreas de trabajo, mini-bio y perfil completo. |
| **Académico** | Pregrado, registro profesional y los 3 soportes del título. |

**Regla fundamental: el motivo del cambio es obligatorio.** Antes de guardar,
la pantalla pide escribir **por qué** se hace el cambio (hasta 500 caracteres).
El texto queda:

1. Guardado en la ficha del agremiado, visible en la próxima edición
   (*"Último motivo registrado"*).
2. Copiado en la bitácora de auditoría, asociado a la modificación.

Si se intenta guardar un cambio sin motivo, el servidor lo rechaza. Un envío
que solo lleva el motivo (sin ningún cambio real) también se rechaza, para que
quede claro que no se está "tocando" la ficha sin propósito.

**Reglas del servidor al guardar**:

- Cédula, FPV y correo se vuelven a validar contra el padrón.
- **Activar** un agremiado exige que tenga FPV válido **y** que esté solvente.
- Hasta 3 imágenes de título; si el guardado falla después de subirlas, se
  deshace la subida.
- Si cambió el correo, se re-sincroniza el acceso a la biblioteca virtual.

**Lo que NO se puede editar desde aquí** (es autogestión del agremiado): la
modalidad de servicio, el interruptor de mostrarla, el aviso de cumpleaños y el
resto de los interruptores de privacidad.

#### Cuaderno B — Gestión (6 pestañas)

| Pestaña | Qué gestiona | Límite |
|---|---|---|
| **Redes** | Redes sociales del perfil público. | 10 |
| **Deontológico** | Entradas del expediente deontológico. Texto plano, hasta **10.000 caracteres**. Editable. | — |
| **Observaciones** | Observaciones internas del Colegio. Hasta **10.000 caracteres**. **Nunca visibles para el agremiado ni para el público.** | — |
| **Documentos** | Expediente documental: cédula, título, RIF, solvencia, comprobante y otros. **Un documento por categoría**, con carga y descarga. | 1 por categoría |
| **Contacto Emergencia** | Alta, edición y baja de los contactos de emergencia (los mismos 3 que ve el agremiado, pero **el Colegio puede gestionarlos**). | 3 |
| **Auditoría** | Historial de cambios de **ese** agremiado, con enlace a la bitácora completa. | — |

#### Otras acciones de la ficha

| Acción | Qué hace |
|---|---|
| **Restablecer contraseña** | Genera una contraseña aleatoria, **invalida las sesiones abiertas** de ese agremiado y le envía un correo con la nueva clave. **La contraseña nunca se muestra en pantalla, ni siquiera al administrador.** |
| **Sincronizar biblioteca** | Re-crea o reactiva la cuenta de biblioteca virtual de ese agremiado. Muestra un informe: creadas, reactivadas, desactivadas, omitidas y errores. |
| **Quitar foto de perfil** | Borra la imagen (pide confirmación). |
| **Eliminar agremiado** | Baja lógica. Requiere permiso de eliminar. |

## 16. Inscripciones — `/admin/inscripciones`

Gestiona las solicitudes públicas de ingreso (§0.3).

### 16.1 Listado

Cuatro pestañas de estado: **Pendientes** (por defecto), **Aprobadas**,
**Rechazadas** y **Todas**. Búsqueda por nombre o cédula.

**Tabla**, 7 columnas: Cédula · Nombre completo · FPV · N.º de Control · Fecha
· Estado · Acciones. Las celdas vacías muestran "—".

### 16.2 Ficha de la solicitud

Pestañas de datos (personales, formación, contacto, ubicación), documentos y
un panel lateral de gestión.

**Acciones disponibles**:

| Acción | Detalle |
|---|---|
| Guardar la ficha completa | Con los mismos campos que la ficha interna del agremiado. |
| Subir foto y comprobante | Hasta 5 MB por archivo. |
| Subir documentos | Un archivo por categoría, con borrado. |
| Guardar notas internas | Hasta guardar. **Con historial completo**: cada guardado con cambio real crea una versión más con fecha y autor. Guardar sin cambios no crea versión ni evento (y el panel avisa *"Las notas ya están guardadas"* en vez de un "guardado" engañoso). El historial **nace vacío**: las notas anteriores a esta función no se migraron. |
| Enviar correo al solicitante | Aviso manual desde la ficha. |
| **Aprobar** | Ver 16.3. |
| **Rechazar** | Ver 16.4. |

**Marks de obligatoriedad**: todos los campos que el Colegio exige para poder
aprobar están marcados con `*` rojo, con la leyenda *"son obligatorios para
aprobar"*. Mientras la ficha guardada tenga huecos, el botón **Aprobar** aparece
**desactivado** con una explicación, y una línea compacta resume qué falta
("Pendientes para aprobar: …"). Se reactiva al guardar la ficha completa.

Las fichas **aprobadas o rechazadas** pasan a **modo lectura**: no se puede
editar ni subir/borrar archivos.

### 16.3 Aprobar una inscripción

Es la operación más delicada del panel, porque crea un agremiado real.

**Comprobaciones previas (todas se evalúan y se reportan de una sola vez)**:

1. **Ficha completa**: los campos obligatorios marcados con `*` y al menos un
   bloque de ubicación completo.
2. **Identidad legal**: cédula y FPV mayores que cero, nombres, apellidos,
   nacionalidad y correo.
3. **Unicidad contra el padrón**: que la cédula, el FPV, el correo y el nombre de
   usuario que se va a generar **no existan** ya entre los agremiados.

Si falta algo, la pantalla **no** muestra un error suelto: devuelve la **lista
completa de pendientes** de una vez, y el panel la muestra sobre los botones de
aprobar/rechazar.

> **Por qué es tan estricta**: antes de este control, aprobar una ficha
> incompleta creaba agremiados con fecha de nacimiento inexistente y género
> vacío, o fallaba con un error técnico inentendible por colisión de datos.

**Al aprobar, el sistema**:

1. Asigna el siguiente **número de control**.
2. Genera el **nombre de usuario** a partir del nombre, en minúsculas y sin
   tildes.
3. Genera una **contraseña aleatoria** y marca la cuenta como *"debe cambiar la
   contraseña"*.
4. Activa el agremiado y lo marca **solvente**, con vencimiento al 31 de
   diciembre del año en curso.
5. **Traslada los documentos** de la solicitud al expediente del agremiado, para
   que no haya que volver a subirlos.
6. Envía el correo de bienvenida.
7. Responde con el número de control y si el correo salió.

**Deshacer el recibo**: una inscripción aprobada **no tiene botón de deshacer**.

### 16.4 Rechazar una inscripción

Abre un modal con un **motivo opcional** (hasta 500 caracteres) que se guarda
con la solicitud.

> **Punto clave**: rechazar **no borra nada**. La solicitud pasa a estado
> "rechazada" y **conserva todos sus datos, notas, documentos y archivos**, para
> que el Colegio pueda revisarla más adelante. En la ficha aparece un bloque
> "Solicitud rechazada" con el motivo.

Efecto secundario importante: al rechazar, la cédula, el FPV y el correo quedan
**libres**, así que esa persona puede presentar una **nueva** solicitud de
inmediato. (El Colegio no tiene una lista de bloqueados por cédula: el bloqueo es
solo de "solicitudes pendientes en curso".)

## 17. Áreas de ejercicio profesional — `/admin/areas_de_ejercicio_profesional`

Catálogo oficial de áreas de trabajo del Colegio. Es una pantalla pequeña pero de
consecuencias grandes, porque **de aquí salen las opciones de todos los
desplegables** del portal del agremiado y **de aquí salen los chips del
directorio público**.

**Listado**: sin búsqueda, sin paginación. Filtro por estado (todas / activas /
inactivas) y dos acciones por fila: **activar/desactivar** (inmediato, sin
confirmación) y **eliminar** (baja lógica, con confirmación).

**Alta y edición**:

| Campo | Regla |
|---|---|
| Nombre | Máximo 100 caracteres. Se filtra en vivo: no deja escribir caracteres no permitidos. |
| Descripción | Máximo 500 caracteres. |

**Eliminar un área no la borra**: la marca como inactiva, y los agremiados que la
tuvieran asignada **dejan de mostrarla en el directorio público** (el aviso lo
dice explícitamente en el modal de confirmación). Los datos no se pierden y se
puede reactivar.

> **Regla de integridad del directorio** (importante): un área solo se muestra
> públicamente si **existe en este catálogo**. Si un agremiado tuviera guardada
> un área que no está en el catálogo, el sistema la descarta en lugar de
> inventarla. Por eso, crear un área que luego se borre o desactive **oculta**
> información que antes era visible.

## 18. Noticias — `/admin/noticias`

Publicaciones del Colegio. Cada una tiene **audiencia** (pública o solo
agremiados) y **estado**.

### 18.1 Estados

| Estado | Significado |
|---|---|
| **Borrador** | No se publica; solo la ve el staff. |
| **Publicado** | Visible para su audiencia. |
| **Programado** | Se publica solo en la fecha y hora indicadas. **Exige fecha de publicación.** |
| **Archivado** | Retirado de la web, pero conservado. |

### 18.2 Listado

Tarjetas con imagen, título, resumen, tipo, estado y fecha. Filtros por
**audiencia**, por **estado** y **búsqueda por texto** (todo en el navegador).
Acciones por tarjeta: **editar**, **ver**, y **cambiar entre publicado y
borrador** con un clic.

> **No existe el borrado de noticias.** No hay ningún botón ni ninguna función
> que elimine una publicación: lo que hay es **archivar**, que la retira de la
> web conservando todo. El botón se titula *"¿Archivar publicación?"*.
>
> **[KI]** El permiso *"Eliminar publicaciones"* existe y se muestra en la
> pantalla de permisos, pero **no lo usa ningún servicio**: el sistema de
> permisos tiene un flag para una capacidad que el producto no implementó.

### 18.3 Crear y editar

| Campo | Regla |
|---|---|
| Título | Máximo 100 caracteres. |
| Descripción corta | Máximo 250 caracteres. |
| Contenido | Editor de texto enriquecido (negrita, cursiva, listas, enlaces, imágenes). |
| Audiencia | Pública · Solo agremiados. |
| Imagen de portada | JPG/PNG/WebP, hasta 5 MB. Se recomprime en el servidor. |
| Estado | Al editar. |
| Fecha de publicación | Al editar, con selector de fecha y hora. **Obligatoria si el estado es "programado"**. |

El contenido se **sanea en el servidor** antes de guardarse: es la barrera real
contra inyección de código en la web pública.

## 19. Notificaciones — `/admin/notificaciones`

Comunicados dirigidos a los agremiados, dentro de su portal. Es el canal de
comunicación interna del Colegio.

### 19.1 Crear

**Destinatarios** (tres modos, excluyentes):

| Modo | Cómo se elige |
|---|---|
| **Global** | Todos los agremiados. |
| **Individual** | Buscador con selección múltiple de personas concretas. |
| **Grupo** | Por filtros: municipio, estado, género, área de ejercicio y solvencia. |

**Contenido**: título (máximo 255 caracteres), mensaje, y opcionalmente una
fecha y hora de envío programado. Hay un botón de **vista previa** antes de
enviar.

**El flujo de envío** es en segundo plano: el envío de miles de notificaciones no
bloquea la pantalla. Si se programa, queda en cola hasta su hora.

> **[KI]** El interruptor de "enviar también por correo" está **desactivado en el
> código a propósito**: las notificaciones llegan solo al portal.

### 19.2 Listado y detalle

Listado paginado con búsqueda por título. Por cada una: estado (pendiente /
enviada) y contadores de destinatarios, leídos y no leídos.

**Cancelar** una notificación solo es posible mientras esté **pendiente** (es
decir, si aún no se ha enviado o programado para el futuro).

> **No se puede editar ni reenviar** una notificación ya creada. Si el mensaje
> tenía un error, la única salida es cancelarla (si está pendiente) y crear otra.

## 20. Tickets — `/admin/tickets`

Gestión de las solicitudes que abren los agremiados. Es el módulo que traduce
las necesidades de los agremiados en un flujo de trabajo del Colegio.

### 20.1 Bandeja

Ordenada por **antigüedad** (primero lo más viejo que sigue abierto), que es
justo el orden en que hay que atenderlas.

| Filtro | Detalle |
|---|---|
| Solo abiertas | Oculta las ya cerradas. |
| Motivo | Los que definió el Colegio. |
| Estado | Los estados de ese motivo. |
| Texto libre | Búsqueda con espera de medio segundo. |

Paginación de 10 en 10. El contador del menú lateral se refresca cada 30
segundos.

### 20.2 Conversación

Cada solicitud tiene su hilo. Desde aquí el Colegio puede:

- **Responder** (hasta 4.000 caracteres, sin límite de mensajes seguidos) y
  adjuntar archivos.
- **Cambiar el estado** de la solicitud. El agremiado recibe una notificación del
  cambio.
- **Cerrar** la solicitud con un motivo (hasta 500 caracteres).

> Cada cambio de estado queda en el historial de la solicitud, visible también
> para el agremiado.

El mensaje que envías aparece **al instante** en el hilo, sin recargar la página.

### 20.3 Configuración

**Motivos de atención** — el Colegio define qué tipos de consulta existen:

| Campo | Regla |
|---|---|
| Nombre | Obligatorio. |
| Descripción | Máximo 500 caracteres. Se muestra al agremiado al elegir el motivo. |
| **Solicitudes abiertas por agremiado** | Entre 1 y 50. Por defecto 3. Es la **cuota**: al alcanzarla, el agremiado no puede abrir más solicitudes de ese motivo hasta cerrar una. |

**Estados por motivo** — cada motivo tiene sus propios estados, con un **orden**
y una marca de **"estado de cierre"**:

- El estado inicial de una solicitud nueva es el primero no cerrado.
- Al cerrar, pasa al primer estado marcado como cierre.
- Si un motivo se queda **sin ningún estado de cierre**, el agremiado no podrá
  cerrar sus solicitudes de ese motivo (y el sistema avisa).

Al crear el Colegio puede partir de tres estados precargados (*Por hacer*,
*En progreso*, *Hecho*) y ajustarlos.

> **[KI]** El borrado de un motivo o un estado pide confirmación con el diálogo
> genérico del navegador, no con el modal con textos propio que usa el resto del
> panel. Cosmético, pero rompe la consistencia.

> **Interacción con los interruptores**: si el Colegio apaga la recepción, el
> personal **sigue podendo gestionar** las solicitudes existentes; solo se
> impide abrir nuevas.

## 21. Proyectos (Kanban) — `/admin/proyectos`

Tableros de trabajo colaborativo del Colegio. Un tablero por proyecto.

### 21.1 Proyectos

**Crear**: nombre (hasta 120 caracteres) y descripción (hasta 500). Al crear el
proyecto se generan automáticamente tres columnas: *Por hacer*, *En progreso* y
*Hecho*.

**Listado**: una tarjeta por proyecto, con una insignia que indica el rol de
quien mira: **master**, **propietario**, **editor** o **lector**. Cada proyecto
tiene además su propio grupo de personas con acceso.

**Eliminar**: pide confirmación. El tablero y su contenido se pierden.

### 21.2 El tablero

| Acción | Detalle |
|---|---|
| Crear y renombrar columnas | Título hasta 120 caracteres. |
| Crear tarjetas | Título hasta 200, descripción hasta 2.000. |
| **Arrastrar tarjetas** entre columnas | Se guarda solo, con actualización inmediata y reversión si falla. |
| Notas por tarjeta | **Hasta 10 notas de 500 caracteres** cada una. |
| Gestionar acceso | Invitar a otras personas con nivel *lector*, *editor* o *propietario*. |

> **Nota técnica**: el tablero es una de las pocas pantallas que **no funciona
> con el servidor**: se construye en el navegador. Por eso el arrastre
> desaparece un instante mientras carga. No es un fallo.

## 22. Bitácora de auditoría — `/admin/auditoria`

Registro forense de **qué cambió, quién lo cambió y cuándo**. Es el módulo de
rastreo interno del Colegio.

### 22.1 Qué se registra

Todo cambio de datos hecho a través de la API: altas, ediciones, bajas, cambios
de estado, inicios y cierres de sesión, cambios de permisos, cesión de SUDO,
restablecimiento de contraseñas, movimientos de tarjetas, envíos de
notificaciones, edición de áreas, cambios en fichas de inscripción, altas y
bajas de contactos de emergencia, y cambios en la configuración.

**Entidades cubiertas**: agremiado, personal, autenticación, notificación,
solicitud, publicación, proyecto, área, ficha de inscripción y configuración.

**Qué guarda cada entrada**:

| Campo | Contenido |
|---|---|
| Qué pasó | Acción (crear, editar, eliminar, iniciar sesión, cambiar estado, mover, enviar…). |
| Sobre qué | Entidad e identificador. |
| Quién | El usuario que lo hizo. |
| Cuándo | Fecha y hora. |
| **Qué cambió** | Campo, valor anterior y valor nuevo. Solo los campos que **realmente** cambiaron. |
| **Por qué** | Motivo, cuando la pantalla lo exige (edición de un agremiado, rechazo de una inscripción). |

> **Dos garantías de diseño**: (1) un fallo al escribir en la bitácora **jamás**
> revierte la operación principal — una modificación de datos nunca se pierde
> por un problema del registro; (2) solo se registra lo que **se guardó de
> verdad**: si la operación falló, no hay entrada.

### 22.2 Consulta

Cuatro pestañas: **General**, **Por psicólogo**, **Por personal** y
**Estadísticas**.

**Filtros**: texto libre, tipo de suceso, tipo de entidad, quién lo hizo, sobre
qué registro, y rango de fechas (desde / hasta, con selector de calendario).

**Acceso directo desde otras pantallas**:

- Desde la ficha de un agremiado, el enlace *"Auditoría"* abre su historial.
- Desde la ficha de un miembro del personal, abre todo lo que ha hecho.

**Paginación**: 20 por página, arriba y abajo. Cualquier cambio de filtro vuelve
a la primera página.

### 22.3 Exportar

Con el permiso de exportar, hay un botón **Exportar CSV** que descarga **todo**
lo que filtra en ese momento, sin paginar. El archivo se genera con la sesión del
usuario (el token viaja en la cabecera, nunca en el enlace).

### 22.4 Retención y borrado

- Las entradas se conservan **90 días** por defecto.
- No hay ningún botón ni permiso de purga: la limpieza es una tarea programada
  interna, exclusiva del SUDO.
- Ajustable por configuración del servidor.

> **Los fallos de permiso en este módulo no se muestran como "no autorizado"**:
> responden como "no encontrado", igual que el resto del panel.

## 23. Staff — `/admin/staff`

Gestión del personal con acceso al panel.

### 23.1 Listado

| Columna | Contenido |
|---|---|
| Usuario | Nombre y correo. |
| Rol | La insignia del perfil preconfigurado, o "Personalizado". |
| Permisos | **N/20**: cuántos de los 20 permisos tiene activos. |
| Estatus | Activo / Inactivo. |
| Creado | Fecha de alta. |
| Acciones | Editar, eliminar, y **ver su actividad en la bitácora**. |

Filtro por estado (todos / activos / inactivos) y búsqueda por nombre o correo.

### 23.2 Alta

| Campo | Regla |
|---|---|
| Usuario | Máximo 25 caracteres, único. |
| Correo | Obligatorio, único. Es donde llegan sus credenciales. |
| Contraseña | Política de robustez completa. |
| Rol | Uno de los perfiles preconfigurados. |
| Permisos | Los 20, con la opción de aplicar un perfil de golpe. |

El correo de bienvenida incluye la **contraseña inicial en claro** (primera
contraseña; la política de seguridad la cambia en su primer uso).

No se puede crear a otro SUDO desde aquí.

### 23.3 Edición

Misma pantalla, con el selector de perfil que aplica los permisos de un clic
("aplicar perfil"). Si luego se toca un permiso a mano, el rol pasa a
*"Personalizado"*.

**Cambiar la contraseña de alguien invalida todas sus sesiones abiertas.** Si el
administrador cambia la suya desde aquí, se desconecta de sus demás pestañas.

Cambios de rol y de permisos quedan registrados en la bitácora con el detalle
completo de los 20 flags.

### 23.4 Baja

Baja lógica. Tres bloqueos: no se puede eliminar la propia cuenta, ni a un SUDO,
ni a alguien por encima en jerarquía. La baja **cierra la sesión** de esa persona.

### 23.5 Ceder el SUDO

Acción exclusiva del SUDO, disponible en el listado de personal.

1. El SUDO elige a quién cederle el poder.
2. **Confirma escribiendo su propia contraseña.**
3. La cesión es **atómica**: no puede quedar nadie sin Super Usuario ni haber dos.
4. Se registra en la bitácora quién cedió a quién.

**No cierra ninguna sesión**: la persona que recibe el SUDO lo obtiene de
inmediato, y la que lo cede sigue conectada (con menos permisos).

## 24. Manual — `/admin/manual`

Acceso a los dos manuales en PDF: el **del panel administrativo** y el **del
portal del agremiado**. Se abre embebido en la pantalla, con botón de descarga.

Están protegidos por sesión: no se pueden descargar sin iniciar sesión. Un
administrador puede leer el manual del agremiado y viceversa.

## 25. Qué NO puede hacer un administrador

| Restricción | Detalle |
|---|---|
| Crear a otro Super Usuario | El campo se fuerza en falso siempre. |
| Editar o eliminar a un Super Usuario | Bloqueado, incluso para otro SUDO. |
| Eliminar su propia cuenta | Bloqueado. |
| Dar un permiso que no tiene | Bloqueado (anti-escalada). |
| Modificar a alguien de rango superior | Bloqueado. |
| Cambiar los interruptores de recepción | Solo el SUDO. |
| Purgar la bitácora | No hay interfaz ni permiso: es tarea programada. |
| Ver la contraseña de un agremiado | El restablecimiento la genera y la envía por correo; nunca se muestra. |
| Editar un agremiado sin escribir el motivo | Obligatorio, o el servidor rechaza el cambio. |
| Aprobar una ficha incompleta | El sistema devuelve la lista completa de pendientes. |
| Editar los datos que el agremiado gestiona | Modalidad de servicio, aviso de cumpleaños y los 16 interruptores de privacidad son del agremiado. |
| Ver los contactos de emergencia en el público | Son datos privados por diseño. |
| Saber si un agremiado del directorio es solvente | La solvencia es interna; el visitante no la ve. |
| Editar o reenviar una notificación ya creada | Solo se puede cancelar si está pendiente. |
| **Borrar** una noticia | No existe: solo archivar. |
| Devolver la contraseña de un miembro del personal | No se guarda en reversible. |
| Acceder a un módulo sin permiso | No aparece en el menú y el servidor rechaza la acción. |

---
---

# Anexos técnicos

## Anexo A — Mapa de endpoints

Referencia rápida. `⛔` = la respuesta es 404 enmascarada ante falta de
credencial.

### Sesión y autenticación

| Método | Ruta | Gate |
|---|---|---|
| POST | `/api/v1/auth/login` | Público, limitado a 5/30 min |
| POST | `/api/v1/psi/login` | Público, limitado a 15/5 min |
| POST | `/api/v1/psi/forgot-password` | Público, limitado |
| POST | `/api/v1/psi/reset-password` | Público, limitado |
| GET | `/session/me` ⛔→ **401 real** | `ProtectedAdmin` |
| GET | `/session/validate` ⛔→ **401 real** | `ProtectedAdmin` |
| POST | `/admin/logout` | Sesión |
| POST | `/psi/me/logout` | Sesión psi |
| GET | `/psi/me/validate` | Sesión psi |

> `/admin/me` y `/admin/validate` **no existen**: se movieron a `/session/*`
> porque bajo `/admin` el sistema solo puede responder 404 (Anexo D).

### Agremiado — autogestión

| Método | Ruta | Gate |
|---|---|---|
| GET | `/psi/me` | Sesión psi |
| PATCH | `/psi/me` (multipart) | Sesión psi + contraseña actual |
| GET | `/psi/me/documents` | Sesión psi (solo lectura) |
| GET | `/psi/me/audiobookshelf` | Sesión psi + solvente |
| POST/PATCH/DELETE | `/psi/me/postgrades[/:id]` | Sesión psi (propio) |
| POST/PATCH/DELETE | `/psi/me/social[/:id]` | Sesión psi (propio) |
| POST/PATCH/DELETE | `/psi/me/emergency[/:id]` | Sesión psi (propio) |
| GET | `/notifications/psi-user` | Sesión psi |
| GET | `/notifications/psi-user/unread-count` | Sesión psi |
| PATCH | `/notifications/psi-user/:id/read` | Sesión psi (destinatario) |
| GET | `/psi/tickets`, `/:id` | Sesión psi (propio) |
| POST | `/psi/tickets` | Sesión psi + recepción abierta |
| POST | `/psi/tickets/:id/mensaje` | Sesión psi (propio, no cerrada) |
| POST | `/psi/tickets/:id/cerrar` | Sesión psi (propio, no cerrada) |
| GET | `/manuales/:file` | Sesión admin **o** psi |

### Personal — administración

| Método | Ruta | Gate |
|---|---|---|
| GET | `/admin/dashboard/stats` | Sesión |
| POST | `/admin/create` | `can_create_admin \|\| sudo` |
| GET | `/admin/list` | Sesión |
| PATCH | `/admin/update` | `can_update_admin \|\| sudo` + jerarquía |
| DELETE | `/admin/delete/:id` | `can_delete_admin \|\| sudo` + anti-auto |
| POST | `/admin/transfer-sudo` | `sudo` + contraseña |
| GET | `/admin/roles/presets` | Sesión |
| GET/POST | `/admin/settings/reception` | GET sesión · POST **`sudo`** |
| GET | `/admin/psi/list` | `can_read_psi \|\| sudo` |
| POST | `/admin/psi/create` | `can_create_psi \|\| sudo` + idempotencia |
| POST | `/admin/psi/upload-csv` | Sesión |
| GET | `/admin/psi/:id` | `can_read_psi \|\| sudo` |
| PATCH | `/admin/psi/:id` | `can_update_psi \|\| sudo` + `last_change_reason` |
| DELETE | `/admin/psi/:id` | `can_delete_psi \|\| sudo` |
| POST | `/admin/psi/:id/reset-password` | `can_update_psi \|\| sudo` |
| POST | `/admin/psi/:id/sync-abs` | Sesión |
| DELETE | `/admin/psi/:id/picture` | `can_update_psi \|\| sudo` |
| GET/POST/PATCH | `/admin/psi/:id/{deontologia,observaciones,documents,social,emergency}` | Según operación |
| DELETE | `/admin/psi/:id/{social,documents,emergency}/:subId` | `can_delete_psi` o lectura |
| GET | `/admin/psi/birthdays?range=today\|week` | Sesión |
| GET/PATCH | `/admin/inscripciones/list`, `/:id` | `can_read_psi \|\| sudo` |
| POST | `/admin/inscripciones/:id/approve` | `can_create_psi \|\| sudo` |
| DELETE | `/admin/inscripciones/:id` (= rechazar) | `can_delete_psi \|\| sudo` |
| GET/PATCH | `/admin/inscripciones/:id/notes` | `can_read_psi` / escritura |
| POST | `/admin/inscripciones/:id/{email,photo,documents}` | Escritura de ficha |
| GET/POST/PATCH/DELETE | `/admin/tickets/motivos[/:id]` | Sesión admin |
| GET/POST/PATCH/DELETE | `/admin/tickets/estados[/:id]` | Sesión admin |
| GET | `/admin/tickets`, `/:id` | Sesión admin |
| PATCH | `/admin/tickets/:id/estado` | Sesión admin |
| POST | `/admin/tickets/:id/{mensaje,cerrar}` | Sesión admin |
| GET | `/admin/tickets/pendientes-count` | Sesión |
| GET/POST/DELETE | `/admin/projects[/:id]`, `/cards/:id` | `can_manage_projects \|\| sudo` |
| GET/POST/DELETE | `/admin/specialties/all`, `/admin/specialties` | `can_*_tags \|\| sudo` |
| POST/PATCH | `/admin/posts[/:id]` | `can_publish` / `can_update_publish` |
| GET/POST/DELETE | `/notifications/admin[/:id]` | `can_*_notifications \|\| sudo` |
| POST | `/notifications/admin/preview` | `can_send_notifications \|\| sudo` |
| GET | `/admin/audit-logs/{,psi/:id,/export,/stats}` | `can_view_logs` / `can_export_logs` |
| GET | `/live`, `/ready` | Público (excluido del límite de tasa) |

### Públicos

| Método | Ruta |
|---|---|
| GET | `/psi/directory`, `/psi/:fpv` || GET | `/posts`, `/posts/:id` |
| GET | `/specialties`, `/specialties/count` |
| GET | `/inscripcion/status`, `/inscripcion/check-ci`, `/check-fpv`, `/check-email` |
| POST | `/inscripcion/submit` |

## Anexo B — Matriz de permisos por módulo

| Módulo | Permiso | Acción |
|---|---|---|
| Psicólogos | `can_read_psi` | Ver listado y ficha |
| | `can_create_psi` | Dar de alta · **Aprobar inscripciones** |
| | `can_update_psi` | Editar ficha · Reset de clave |
| | `can_delete_psi` | Eliminar · **Rechazar inscripciones** |
| Inscripciones | `can_read_psi` | Ver listado y ficha · Notas (lectura) |
| | *escritura de ficha* | Guardar datos, subir/borrar archivos, notas, correos |
| Áreas | `can_create_tags` | Crear |
| | `can_edit_tags` | Editar, activar/desactivar |
| | `can_delete_tags` | Eliminar |
| Noticias | `can_publish` | Crear |
| | `can_update_publish` | Editar · **Archivar** |
| | `can_delete_publish` | *(flag huérfano, ver Anexo E)* |
| Notificaciones | `can_send_notifications` | Crear, previsualizar, adjuntar |
| | `can_manage_notifications` | Cancelar cualquier notificación |
| | `can_read_notifications` | Ver detalle y destinatarios |
| Tickets | `can_manage_tickets` | Bandeja, chat, estados, configuración |
| Proyectos | `can_manage_projects` | Crear, mover tarjetas, gestionar acceso |
| Auditoría | `can_view_logs` | Consultar |
| | `can_export_logs` | Exportar CSV |
| Staff | `can_create_admin` | Crear |
| | `can_update_admin` | Editar |
| | `can_delete_admin` | Eliminar |
| Recepción | *(solo `sudo`)* | Interruptores de tickets e inscripciones |
| Menú | `always` | Dashboard y Manual, visibles para todos |

## Anexo C — Consolidado de cuotas

### Archivos y sanitización

| Regla | Valor |
|---|---|
| Foto de perfil | 800 px · 150 KB · WebP calidad 80 |
| Documentos y soportes | 1.600 px · 400 KB (imagen) |
| PDF | 4 MB |
| XLSX de importación | 5 MB (validado en el modal) |
| Archivos de inscripción | 5 MB |
| Compresión | Se reduce un 20 % iterativamente hasta que quepa |

### Textos

| Regla | Valor |
|---|---|
| Contraseña | ≥ 8 caracteres, sin espacios, con mayúscula, minúscula, número y símbolo |
| Contraseña generada (alta, reset, aprobación) | 12 caracteres aleatorios (16 en la importación masiva) |
| Usuario generado al aprobar | 25 caracteres, sin tildes, minúsculas |
| Nombre de área de ejercicio | 100 caracteres |
| Motivo / descripción de ticket | 500 caracteres |
| Título de ticket | 200 |
| Descripción de ticket | 2.000 |
| Mensaje del agremiado | 1.000 |
| Mensaje del Colegio | 4.000 |
| Motivo de cierre / de cambio de estado | 500 |
| Motivo de cambio de un agremiado | 500 |
| Motivo de rechazo de inscripción | 500 |
| Mensaje público de un interruptor | 500 (recortado, no rechazado) |
| Deontología / Observaciones | 10.000 |
| Título de noticia | 100 |
| Descripción corta de noticia | 250 |
| Título de notificación | 255 |
| Nombre de proyecto / columna | 120 |
| Título de tarjeta | 200 |
| Descripción de tarjeta / proyecto | 2.000 / 500 |
| Nota de tarjeta | 500 · **máximo 10 por tarjeta** |
| Motivos de ticket | Descripción 500 · **cuota 1–50 por agremiado** (por defecto 3) |

### Sesión, tasa y retención

| Regla | Valor |
|---|---|
| Duración de sesión | 24 h (biblioteca: 30 días) |
| Login del agremiado | 15 / 5 min por conexión |
| Login del personal | 5 / 30 min por conexión |
| Límite general de la API | 60 req/min por conexión (no cuenta los preflight) |
| Token de recuperación | 1 h, un solo uso |
| Retención de la bitácora | 90 días (`AUDIT_LOG_RETENTION_DAYS`) |
| Cola de auditoría | 5.000 entradas, escritura por lotes de 50/s |
| **Idempotencia** | 30 min, cabecera `X-Idempotency-Key`, en altas de agremiado, noticias, notificaciones, personal, áreas y proyectos |

## Anexo D — Gotchas de arquitectura

Conocimientos que costaron tiempo descubrir. Si tocas estas áreas, léelos
primero.

### 1. Por qué el panel devuelve "no encontrado" en vez de "no autorizado"

El grupo de rutas `/admin` está protegido por un middleware que, ante falta de
credencial, responde **404** en lugar de 401. Es deliberado: no revela qué rutas
existen. Consecuencia práctica: **"Cannot PATCH /api/v1/admin/..." significa
"la petición llegó sin sesión válida"**, no "la función no existe".

Para que el portal pudiera distinguir "sesión caducada" de "ruta inexistente" se
crearon dos rutas fuera de `/admin` (`/session/me`, `/session/validate`) que sí
responden 401 real.

> **Trampa de Fiber**: registrar un segundo grupo de rutas sobre un prefijo ya
> usado hace que herede el middleware del primero y el suyo nunca se ejecute.
> Fue exactamente lo que pasó con las rutas de validación entre el 5 y el 28 de
> septiembre: estaban registradas y respondían 404. **Solución adoptada:
> prefijos distintos**, nunca dos grupos sobre el mismo prefijo.

### 2. Por qué "se cerró la sesión" al hacer una acción

Fue el problema más reportado. Tiene tres causas distintas, todas resueltas:

1. **El navegador pierde la copia del token al abrir una pestaña nueva.** Ahora
   el portal la recupera sola de la cookie, sin pedir login otra vez.
2. **Un 404 de la ruta de validación se trataba como sesión revocada.** Ahora
   solo un 401/403 cierra la sesión: un 404 significa ruta inexistente o
   desfase de despliegue.
3. **Un fallo puntual de la API en un módulo tumbaba todo el panel.** Ahora cada
   módulo falla de forma aislada, con su propia pantalla de error y sin perder
   el menú.

Además, un 401 en una petición de datos **ya no borra la sesión**: la sesión solo
muere cuando la verificación periódica lo confirma.

> **Consecuencia de diseño a conocer**: re-iniciar sesión (o cambiar la
> contraseña) **invalida las sesiones anteriores de esa misma persona**. Es
> anti-replay, pero hay que explicárselo al personal del Colegio.

### 3. La conexión con la base de datos

El acceso a PostgreSQL pasa por un pooler (PgBouncer) en modo transacción. Con la
configuración habitual, el driver de Go cacheaba consultas preparadas que
sobrevivían a las migraciones, provocando errores técnicos (500) que el portal
mostraban como "Conexión en pausa" a pantalla completa.

**Solución en tres capas** (y son tres cosas que **no se deben deshacer**):

1. La API usa protocolo simple, que no cachea consultas preparadas.
2. El pooler está configurado para ignorar ese parámetro al arrancar.
3. El frontend aísla los errores por módulo.

> **Operativo**: tras aplicar una migración que altera una tabla en uso,
> **reiniciar el pooler**. Ha hecho falta más de una vez: el plan compilado
> queda con el esquema viejo.

### 4. Errores de "demasiadas solicitudes" (429)

El límite general de la API es de 60 peticiones por minuto y conexión. Como cada
petición del navegador genera además una comprobación previa (OPTIONS) que
**antes** contaba, y como en la red de contenedores todo el tráfico sale desde
la misma dirección, una uso normal agotaba el cupo y el panel entero respondía
"demasiadas solicitudes" durante un minuto.

**Solución**: las comprobaciones previas y las rutas de salud ya no cuentan, y el
mensaje al usuario es accionable. **No revertir esto.**

> **Operativo**: si el servicio de Valkey está caído, el contador pasa a
> memoria del proceso y **se reinicia con cada reinicio de la API**. Con Valkey
> arriba es persistente y compartido entre réplicas.

### 5. Fechas

**Ningún formulario usa el selector nativo del navegador.** Todos usan un
componente propio con calendario, en español. El valor que viaja al servidor es
siempre el mismo formato de antes (`YYYY-MM-DD`, o con hora `YYYY-MM-DDTHH:MM`),
así que la API no se enteró del cambio.

### 6. Sanitización de contenido

El contenido que escriben los usuarios (perfiles, noticias, tickets) se
**sanitiza en el servidor antes de guardarse**. La limpieza del navegador es solo
una capa extra: **la barrera que importa es la del backend**.

### 7. Los manuales en PDF

Están **incrustados en el binario del backend**, no como archivos públicos de la
web. Se sirven por una ruta propia protegida por sesión, con lista blanca de
nombres (imposible el recorrido de rutas) y sin caché. Para actualizarlos hay que
copiarlos al directorio del backend y **reconstruir la API**.

### 8. El tablero de proyectos no se renderiza en el servidor

Arrastrar tarjetas necesita montar una capa flotante sobre el documento, lo que
rompe el renderizado en servidor de esa ruta. Por eso el tablero solo se
construye en el navegador. **No mover esa parte fuera del bloque "solo
navegador"** o la página dejará de cargar al refrescarse.

## Anexo E — Inconsistencias detectadas

Detectadas al redactar este documento. **Ninguna está corregida.** Se listan
para que no se confundan con comportamiento intencionado.

| # | Dónde | Qué ocurre |
|---|---|---|
| 1 | `web/src/routes/public/**` | 6 archivos de **0 bytes** sin ninguna ruta que los enlace. Restos de una reorganización anterior. |
| 2 | `web/src/routes/admin/dashboard/index.tsx` | Stub de 5 líneas; el menú no lo enlaza (el dashboard real está en `/admin`). |
| 3 | Alta de agremiado, sección Contacto | El teléfono se escribe en un campo que no existe en el modelo (`public_phone` en vez de `contact_phone`): **el dato no se guarda**. |
| 4 | Alta de agremiado | Declara los campos de área de trabajo sin dibujar los campos: quedan vacíos. |
| 5 | Alta de agremiado | Si el servidor rechaza el alta, no se muestra **qué campo** causó el rechazo. |
| 6 | Listado de colegiados | El botón dice **"Importar CSV"** pero importa un Excel y el endpoint también se llama `upload-csv`. |
| 7 | Noticias | El permiso **"Eliminar publicaciones"** existe y se muestra, pero **ningún servicio lo usa**: no hay borrado, solo archivado. |
| 8 | Noticias | El botón de archivar está rotulado *"¿Archivar publicación?"* dentro de un diálogo titulado como si fuera borrado. |
| 9 | Notificaciones | El interruptor de envío por correo está **desactivado a propósito** en el código. |
| 10 | Tickets → Configuración | El borrado de un motivo o estado usa la confirmación genérica del navegador, no el modal con texto propio del resto del panel. |
| 11 | Comentarios en el backend de roles | Dicen "18 permisos" en tres sitios; el código maneja **20**. |
| 12 | Perfil del agremiado | **Borrar un campo** dejándolo en blanco **no lo borra** (el valor anterior sobrevive). Apagar con los interruptores sí funciona. |
| 13 | Perfil → Redes sociales | Si añadir o borrar una red falla, **no se muestra error**: la lista simplemente no cambia. |
| 14 | Académico (agremiado) | La etiqueta promete "PDF o imagen" pero el selector **solo admite imágenes**. |
| 15 | `admin_roles.go` vs `staff-permissions.ts` | La numeración de permisos no coincide entre el comentario del backend y la interfaz. |
| 16 | Gestión de personal | Los endpoints de alta, edición, baja y cesión de SUDO responden **403 con el mensaje real** del servidor, mientras el resto del panel enmascara la falta de credencial como 404. Inconsistente, aunque cada mensaje es correcto para quien lo ve. |

## Anexo F — Glosario

| Término | Significado |
|---|---|
| **Agremiado / psi user** | El psychologist inscrito en el Colegio. Su cuenta en el portal `/psi`. |
| **Staff** | El personal del Colegio con acceso al panel `/admin`. |
| **SUDO** | Super Usuario. La única cuenta que pasa por alto todos los permisos. |
| **FPV** | Ficha del Psicólogo. Es el número que identifica al profesional ante el Colegio (en el código aparece como "FPV" por herencia del nombre original del campo). |
| **Solvencia** | Si el agremiado está al día con la cuota del Colegio. Determina el acceso a la biblioteca virtual y si sus áreas se muestran en el directorio. |
| **Número de control** | Identificador interno correlativo que asigna el Colegio. |
| **Expediente** | La ficha completa de un agremiado en el panel. |
| **Expediente deontológico** | Registro de conductas y procedimientos del Colegio sobre un profesional. |
| **Observaciones internas** | Notas del Colegio sobre un agremiado. **Nunca visibles** para él ni para el público. |
| **Ficha de inscripción** | La solicitud de ingreso que un profesional presenta por el formulario público. |
| **Motivo de atención** | La categoría de una solicitud del agremiado (solicitud, pregunta, otro…). Lo define el Colegio. |
| **Bitácora / auditoría** | El registro interno de qué cambió, quién y cuándo. |
| **Motivo del cambio** | La explicación obligatoria que el Colegio escribe al editar un agremiado. Se guarda en su ficha y en la bitácora. |
| **Interruptor de recepción** | El apagador global que impide abrir nuevas solicitudes o inscripciones. |
| **Biblioteca virtual** | La colección de audio-libros del Colegio (Audiobookshelf). Solo para solventes. |
| **Idempotencia** | Poder reintentar un alta sin crear duplicados. |

---

*Documento generado el 28 de septiembre de 2026 a partir del código fuente del
proyecto. Estructura y reglas del proyecto en `AGENTS.md`; detalle técnico por
módulo en `docs/`.*

