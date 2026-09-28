// ============================================================================
// Manual de Administración — Panel Administrativo del Colegio
// de Psicólogos de Carabobo
// Compilar:  typst compile manual-admin.typ manual-admin.pdf
// Desde:     docs/  (las imágenes son relativas a este archivo → capturas-admin/*.png)
// ============================================================================

// ── Paleta institucional ─────────────────────────────────────────────────────
#let azul   = rgb("#14328c")   // azul institucional
#let amar   = rgb("#facc15")   // amarillo institucional
#let rojo   = rgb("#b91c1c")
#let gris   = rgb("#6b7280")
#let fondo  = rgb("#f3f6fb")

#set document(title: "Manual de Administración — Panel Administrativo")
#set page(
  paper: "a4",
  margin: (x: 1.9cm, y: 2cm),
  footer: context {
    set text(fill: gris, size: 8.5pt)
    align(center + horizon)[
      Colegio de Psicólogos del Estado Carabobo · Panel Administrativo · Página #counter(page).get().first()
    ]
  },
)
#set text(font: ("Noto Sans", "Noto Sans CJK SC"), lang: "es", region: "VE", size: 10pt)
#set par(justify: true, leading: 0.62em)
// Los títulos llevan su número en el texto (1. Ingreso, 4.1 Identidad…). La
// numeración automática está desactivada para no duplicar el número.
#show heading: set text(fill: azul, weight: "bold")
// Cada título principal (nivel 1) inicia en página nueva.
#show heading.where(level: 1): it => {
  pagebreak()
  it
}
#set list(marker: [•], tight: true)

// ── Bloque de consejo ("tip") ────────────────────────────────────────────────
#let tip(body) = block(
  width: 100%,
  fill: rgb("#fffbe6"),
  stroke: (left: 3pt + amar),
  radius: 4pt,
  inset: (x: 12pt, y: 9pt),
  breakable: true,
)[
  #text(weight: "bold", fill: rgb("#8a6d00"))[Consejo útil — ] #body
]

// ── Caja de información / advertencia ────────────────────────────────────────
#let nota(body, color: azul) = block(
  width: 100%,
  fill: fondo,
  stroke: (left: 3pt + color),
  radius: 4pt,
  inset: (x: 12pt, y: 9pt),
  breakable: true,
)[#body]

// ── Caja de advertencia (rojo) ───────────────────────────────────────────────
#let aviso(body) = nota(body, color: rojo)

// ============================================================================
// PORTADA
// ============================================================================
#align(center)[
  #block(width: 100%, fill: azul, inset: (y: 26pt), radius: 10pt)[
    #text(size: 22pt, weight: "bold", fill: white)[Manual de Administración]
    #v(4pt)
    #text(size: 17pt, fill: white)[Panel Administrativo del Colegio]
  ]
  #v(10pt)
  #text(size: 26pt, weight: "bold", fill: azul)[Colegio de Psicólogos]\
  #text(size: 26pt, weight: "bold", fill: azul)[del Estado Carabobo]
  #v(6pt)
  #text(size: 12pt, fill: gris)[Sistema de gestión gremial — psicólogos, inscripciones, comunicaciones y auditoría]
]

#v(14pt)

#grid(
  columns: (1fr, 1fr),
  gutter: 12pt,
  [#block(radius: 8pt, fill: fondo, inset: 12pt)[
    #text(size: 13pt, weight: "bold", fill: azul)[Versión del manual]
    #v(3pt)
    #text(size: 10pt)[1.0 — septiembre de 2026]
  ]],
  [#block(radius: 8pt, fill: fondo, inset: 12pt)[
    #text(size: 13pt, weight: "bold", fill: azul)[Audiencia]
    #v(3pt)
    #text(size: 10pt)[Personal administrativo del Colegio con acceso al panel]
  ]],
)

#v(18pt)

#block(width: 100%, fill: rgb("#fef3c7"), stroke: 0.8pt + amar, radius: 8pt, inset: 12pt)[
  #text(weight: "bold", fill: rgb("#8a6d00"))[Aviso] — Las capturas que acompañan este manual
  son ilustrativas y pueden variar según la información registrada en cada momento
  y la versión vigente del sistema.
]

#v(16pt)

// ── Índice ───────────────────────────────────────────────────────────────────
#outline(indent: auto, depth: 2)

// ============================================================================
// 1. INTRODUCCIÓN
// ============================================================================
= 1. Introducción

El *Panel Administrativo* es el sistema interno del Colegio de Psicólogos del
Estado Carabobo. Permite al personal autorizado gestionar el padrón de
agremiados, las solicitudes de inscripción, las comunicaciones gremiales
(noticias, notificaciones y tickets) y los proyectos internos, con una
*bitácora de auditoría* que registra cada cambio.

El panel funciona bajo un esquema de *roles y permisos* (RBAC-liviano):

+ *Super Usuario (SUDO)*: acceso total e irrevocable a todos los módulos.
+ *Presets de rol*: plantillas de permisos según el cargo (Secretaría,
  Comunicación, Soporte, Proyectos, Lector).
+ *Perfil personalizado*: cada administrador puede recibir permisos puntuales
  marcados uno a uno.

#nota[El sistema filtra el *menú* según los permisos de cada administrador, pero
la barrera real es la API: un permiso ausente responde siempre como
"recurso no encontrado" (oculto ante quien no tiene acceso).]

#tip[Anota tu *usuario* y *contraseña* de acceso en un lugar seguro. El cambio
de contraseña o un nuevo ingreso invalidan las sesiones anteriores de esa
cuenta (medida anti-replay del Colegio).]

// ============================================================================
// 2. INGRESO AL PANEL
// ============================================================================
= 2. Ingreso al panel

El acceso está restringido al personal administrativo. Se ingresa desde la
pantalla *Acceso Restringido*, que solicita:

+ *Username o Correo institucional*: el identificador que la administración
  asignó al operador.
+ *Contraseña*: la clave personal. El sistema normaliza el identificador a
  minúsculas; la contraseña, en cambio, distingue mayúsculas y minúsculas.

#figure(
  image("capturas-admin/01-admin-acceso.png", width: 78%),
  caption: [Pantalla de acceso restringido del panel],
)

Si las credenciales son incorrectas, aparece el aviso _Fallo de Seguridad_ y el
acceso se bloquea:

#figure(
  image("capturas-admin/02-admin-acceso-error.png", width: 78%),
  caption: [Aviso de credenciales incorrectas],
)

+ Si se recibe repetidamente *"Demasiadas solicitudes"*, espera un minuto e
  inténtalo de nuevo: es la protección anti-abuso del Colegio.
+ No compartas tu sesión. Si otra persona ingresa con tu usuario, tu sesión
  anterior queda invalidada automáticamente.
+ La sesión del panel se mantiene *por pestaña*: al abrir una pestaña nueva, el
  sistema recupera la sesión de forma silenciosa cuando la cookie sigue vigente.

// ============================================================================
// 3. PANEL PRINCIPAL (DASHBOARD)
// ============================================================================
= 3. Panel principal (dashboard)

Al ingresar se muestra el *panel de gestión*, que resume la actividad del gremio
en una sola pantalla:

#figure(
  image("capturas-admin/03-dashboard.png", width: 82%),
  caption: [Panel principal de gestión],
)

+ *Indicadores (KPIs)*: inicios de sesión, páginas vistas, búsquedas, visitas a
  perfiles y usuarios activos, con tendencias de los últimos días.
+ *Ranking de actividad*: especialidades, municipios y términos de búsqueda más
  consultados en el Directorio.
+ *Cumpleaños de la semana*: muestra los agremiados que cumplen años, para
  coordinar saludos y gestiones.
+ *Interruptores de recepción*: permiten *pausar o reanudar* la recepción de
  nuevas solicitudes (tickets) e inscripciones. Cuando un canal está pausado,
  los formularios públicos responden con el aviso y el psicólogo no puede
  enviar hasta reabrirlo.
+ *Sesiones activas*: aviso de accesos concurrentes detectados.

#tip[Revisa diariamente el dashboard: los picos de búsqueda y de visitas a
perfiles ayudan a detectar campañas, incidencias o interés del público en una
especialidad.]

// ============================================================================
// 4. PSICÓLOGOS
// ============================================================================
= 4. Psicólogos

El módulo de *Psicólogos* administra el padrón de agremiados: alta, consulta,
edición, estados gremiales (solvencia y fe de vida) y expediente completo de
cada colegiado.

== 4.1 Listado de psicólogos

El listado permite *buscar* por nombre, cédula o número de FPV y filtrar los
resultados, mostrando el estado de cada colega (activo, solvente, control
interno) en la misma tabla.

#figure(
  image("capturas-admin/04-psicologos-listado.png", width: 82%),
  caption: [Listado de psicólogos con búsqueda y filtros],
)

== 4.2 Alta de un psicólogo

Para registrar a un agremiado de forma manual, usa *Nuevo psicólogo* y completa
la ficha. El sistema exige la identidad legal (cédula, FPV, nombres, apellidos,
nacionalidad) y crea las credenciales de acceso al portal gremial.

#figure(
  image("capturas-admin/05-psicologo-crear.png", width: 82%),
  caption: [Formulario de alta de un psicólogo],
)

#nota[La vía preferente de alta es la *aprobar una inscripción* (sección 5):
así el expediente nace completo y validado por el comité.]

== 4.3 Ficha del psicólogo

La ficha del agremiado se organiza en dos *libretas de pestañas*. La primera
cubre la cuenta y el expediente: *Cuenta, Estatus, Solvencias, Identidad,
Contacto, Ubicación, Perfil y Académico*.

#figure(
  image("capturas-admin/06-psi-detalle-identidad.png", width: 82%),
  caption: [Ficha del psicólogo — pestaña Identidad (con edad calculada)],
)

En *Identidad* el sistema muestra la *edad calculada* a partir de la fecha de
nacimiento. *Solvencias* permite registrar la solvencia anual del colega y
*Estatus* controla el estado de la cuenta y el número de control interno.

La segunda libreta concentra los módulos de gestión: *Redes, Deontológico,
Observaciones, Documentos, Contacto Emergencia y Auditoría*.

#figure(
  image("capturas-admin/07-psi-detalle-academico.png", width: 82%),
  caption: [Ficha del psicólogo — pestaña Académico],
)

+ *Redes* (presencia digital): vincula los perfiles profesionales del agremiado.
+ *Deontológico*: registro de incidencias del expediente deontológico.
+ *Observaciones*: notas internas del gremio.
+ *Documentos*: archivos digitales del agremiado (título, RIF, copias).
+ *Contacto Emergencia*: personas de contacto ante emergencias, siempre
  privadas.
+ *Auditoría*: trazabilidad de cambios sobre la ficha.

#figure(
  image("capturas-admin/08-psi-detalle-redes.png", width: 82%),
  caption: [Ficha del psicólogo — presencia digital (Redes)],
)

#figure(
  image("capturas-admin/09-psi-detalle-emergencia.png", width: 82%),
  caption: [Ficha del psicólogo — contacto de emergencia],
)

#figure(
  image("capturas-admin/10-psi-detalle-documentos.png", width: 82%),
  caption: [Ficha del psicólogo — documentos],
)

== 4.4 Motivo del cambio

Al editar cualquier dato del expediente, el sistema exige declarar el *motivo
del cambio* (obligatorio, máximo 500 caracteres). El motivo queda registrado en
la ficha (como último motivo) y en la bitácora de auditoría, como parte del
expediente interno del gremio.

#aviso[Los datos que el psicólogo marca como privados *no* se modifican desde el
panel sin justificación: la cuenta del agremiado y su visibilidad en el
Directorio son gestionadas por el propio colega desde su portal.]

// ============================================================================
// 5. INSCRIPCIONES
// ============================================================================
= 5. Inscripciones

El módulo de *Inscripciones* gestiona las solicitudes de incorporación enviadas
desde el formulario público de inscripción.

== 5.1 Listado de solicitudes

El listado muestra cada solicitud con su *N° de Control*, nombre del aspirante,
estado (pendiente, aprobada, rechazada) y fecha. Los estados *aprobada* y
*rechazada* abren la ficha en modo lectura y las solicitudes pendientes se
pueden editar y completar.

#figure(
  image("capturas-admin/11-inscripciones-listado.png", width: 82%),
  caption: [Solicitudes de inscripción con N° de Control],
)

== 5.2 Ficha de la solicitud y pendientes

Al abrir una solicitud *pendiente*, se muestran los datos cargados por el
aspirante (identidad, académico, ubicación, documentos) y un bloque de *notas
administrativas*. Si a la ficha le faltan datos para constituir al psicólogo,
un indicador enumera los *pendientes para aprobar* y el botón de aprobación
queda deshabilitado hasta completarlos y guardar.

#figure(
  image("capturas-admin/12-inscripcion-pendiente-incompleta.png", width: 82%),
  caption: [Ficha con pendientes: se detallan los campos que faltan],
)

== 5.3 Aprobar una inscripción

Cuando la ficha está completa, el botón *Aprobar inscripción* se habilita. La
aprobación crea la cuenta del psicólogo *activa, solvente y con fe de vida*, y
su foto tipo carnet pasa a ser la foto de perfil. El sistema muestra un diálogo
de confirmación antes de ejecutar la acción.

#figure(
  image("capturas-admin/13-inscripcion-aprobar-modal.png", width: 78%),
  caption: [Confirmación antes de aprobar la inscripción],
)

#nota[La aprobación es integral: si faltan datos o el CI, FPV o correo del
aspirante ya existen en el Directorio, el sistema devuelve la lista completa de
impedimentos en lugar de aprobar a medias.]

== 5.4 Rechazar una inscripción

*Rechazar solicitud* permite rechazar el trámite indicando un *motivo*
(opcional). La solicitud pasa a estado *rechazada* pero conserva la ficha, las
notas, los documentos y los archivos, para revisión o re-aplicación futura del
aspirante. La ficha rechazada se abre en *modo lectura* y muestra el motivo
consignado.

#figure(
  image("capturas-admin/14-inscripcion-rechazada.png", width: 82%),
  caption: [Solicitud rechazada: el expediente se conserva para revisión],
)

== 5.5 Historial de notas administrativas

Cada vez que se guardan las *notas* de una ficha con un cambio real, se agrega
una *versión* al historial con el texto, el autor y la fecha. El bloque
*Historial de notas* muestra la evolución del seguimiento del trámite.

#figure(
  image("capturas-admin/15-inscripcion-historial-notas.png", width: 82%),
  caption: [Historial de notas administrativas de la ficha],
)

// ============================================================================
// 6. ÁREAS DE EJERCICIO PROFESIONAL
// ============================================================================
= 6. Áreas de Ejercicio Profesional

El catálogo de *áreas de ejercicio* contiene las especialidades profesionales
que se muestran en el Directorio público. Solo los agremiados *solventes*
exhiben sus áreas, y siempre que existan en este catálogo.

#figure(
  image("capturas-admin/16-areas-listado.png", width: 82%),
  caption: [Catálogo de áreas de ejercicio profesional],
)

Después de *crear o editar* un área, esta queda disponible para asignarla en la
ficha de los psicólogos (areas primaria y secundaria) y para filtrar el
Directorio:

#figure(
  image("capturas-admin/17-areas-crear.png", width: 78%),
  caption: [Alta de una nueva área de ejercicio],
)

#tip[Mantén el catálogo limpio: un área mal escrita no genera coincidencias en
el Directorio y un psicólogo con un área fuera del catálogo no la verá
publicada.]

// ============================================================================
// 7. NOTICIAS
// ============================================================================
= 7. Noticias

El módulo de *Noticias* administra las publicaciones del sitio (inicio,
directorio y portal). Cada publicación tiene un *estado*: publicada, borrador o
archivada.

== 7.1 Listado y estados

El listado muestra todas las publicaciones con su estado y permite *editar* o
*eliminar* cada una. Publicar hace visible la noticia de inmediato.

#figure(
  image("capturas-admin/18-noticias-listado.png", width: 82%),
  caption: [Listado de noticias con sus estados],
)

== 7.2 Redacción y edición

El editor incluye un *procesador de texto enriquecido (RTF)*: títulos,
negritas, listas, imágenes y enlaces, igual que el sitio público. Antes de
publicar se puede guardar como borrador.

#figure(
  image("capturas-admin/19-noticias-crear.png", width: 82%),
  caption: [Editor de noticias con formato enriquecido],
)

#figure(
  image("capturas-admin/20-noticias-detalle.png", width: 82%),
  caption: [Vista de detalle de la noticia],
)

#nota[El contenido se valida y sanitiza en el servidor antes de publicarse; una
noticia con código no permitido se limpia automáticamente.]

// ============================================================================
// 8. NOTIFICACIONES
// ============================================================================
= 8. Notificaciones

El módulo de *Notificaciones* envía avisos del Colegio a los agremiados, que
los reciben en su portal (campana de notificaciones) y, opcionalmente, por
correo.

== 8.1 Listado

El listado muestra las notificaciones enviadas, su *destino* (global o
individual) y su estado de envío:

#figure(
  image("capturas-admin/21-notificaciones-listado.png", width: 82%),
  caption: [Listado de notificaciones enviadas],
)

== 8.2 Redactar y enviar

Para crear una notificación se completa el *título*, el *mensaje* y el
*tipo de destino*. Las notificaciones individuales se dirigen a un agremiado
específico; las globales, a todos los colegiados. El envío por correo es
opcional.

#figure(
  image("capturas-admin/22-notificaciones-crear.png", width: 82%),
  caption: [Redacción de una nueva notificación],
)

#tip[Usa destinatarios *globales* con moderación: un aviso excesivo satura la
bandeja de los psicólogos. Las noticias se usan para divulgación general y las
notificaciones para avisos accionables.]

// ============================================================================
// 9. TICKETS (SOLICITUDES)
// ============================================================================
= 9. Tickets (solicitudes)

El módulo de *Tickets* gestiona las solicitudes que los psicólogos abren desde
su portal (constancias, actualización de datos, incidencias del portal).

== 9.1 Cola de solicitudes

El menú muestra un *indicador* con la cantidad de solicitudes pendientes por
atender (se actualiza periódicamente). El listado presenta la cola con su
motivo, estado y prioridad de atención.

#figure(
  image("capturas-admin/23-tickets-listado.png", width: 82%),
  caption: [Cola de solicitudes (tickets) por atender],
)

== 9.2 Tramitación y conversación

Al abrir una solicitud se ven el motivo, la descripción del agremiado y un
*hilo de conversación* con el psicólogo. La respuesta se publica al instante,
sin recargar la página, y el estado de la solicitud se actualiza conforme avanza
el trámite (recibido → en proceso → cerrado).

#figure(
  image("capturas-admin/24-tickets-detalle.png", width: 82%),
  caption: [Tramitación de una solicitud con conversación],
)

== 9.3 Configuración de motivos y estados

En *Configuración de Tickets* la administración mantiene el catálogo de
*motivos* que ven los psicólogos al crear una solicitud y los *estados* del
flujo de atención. Se pueden crear, renombrar y reordenar:

#figure(
  image("capturas-admin/25-tickets-configuracion.png", width: 82%),
  caption: [Configuración de motivos y estados de las solicitudes],
)

// ============================================================================
// 10. PROYECTOS (KANBAN)
// ============================================================================
= 10. Proyectos (kanban)

El módulo de *Proyectos* organiza el trabajo interno del Colegio en tableros
*kanban* con columnas y tarjetas. Solo los administradores con permiso de
proyectos (o *Super Usuario*) acceden a este módulo.

== 10.1 Listado de proyectos

El listado muestra los proyectos creados y permite abrirlos o crear nuevos:

#figure(
  image("capturas-admin/26-proyectos-listado.png", width: 82%),
  caption: [Listado de proyectos internos],
)

== 10.2 Tablero kanban

Dentro del proyecto, las tarjetas se *arrastran* entre columnas (por ejemplo,
"Por hacer", "En progreso", "Hecho"); al soltar una tarjeta, el sistema guarda
el cambio automáticamente.

#figure(
  image("capturas-admin/27-proyectos-kanban.png", width: 88%),
  caption: [Tablero kanban del proyecto con sus tarjetas],
)

#tip[Describe cada tarjeta con el resultado esperado, no solo la tarea: el
tablero funciona como acta de trabajo del equipo directivo.]

// ============================================================================
// 11. AUDITORÍA
// ============================================================================
= 11. Auditoría

La *bitácora de auditoría* registra, a nivel de API, *qué* se cambió, *cuándo*
y *quién* lo hizo. Es la herramienta forense del Colegio: cada acción sobre los
expedientes, inscripciones, publicaciones y administradores queda trazada.

#figure(
  image("capturas-admin/28-auditoria-listado.png", width: 82%),
  caption: [Bitácora de auditoría con filtros y paginación],
)

+ *Filtros*: por suceso, entidad, acción, actor o rango de fechas.
+ *Paginación*: 20 registros por página; se navega desde el listado.
+ *Exportación*: el historial completo se puede exportar (CSV) para archivo
  institucional.

#aviso[La auditoría es de solo lectura: nadie puede eliminar o editar entradas.
El registro se conserva por un período definido por el Colegio (por defecto
90 días).]

// ============================================================================
// 12. STAFF (ADMINISTRADORES)
// ============================================================================
= 12. Staff (administradores)

El módulo de *Staff* administra las cuentas del personal con acceso al panel:
altas, permisos, estados y sucesión del rol de Super Usuario.

== 12.1 Listado de administradores

El listado muestra cada administrador, su rol, su permiso de *Super Usuario*
(SUDO) y su estado (activo/inactivo):

#figure(
  image("capturas-admin/29-staff-listado.png", width: 82%),
  caption: [Listado del personal administrativo],
)

== 12.2 Alta de un administrador

Al crear a un administrador se elige un *preset de rol* (Secretaría,
Comunicación, Soporte, Proyectos, Lector) o se configura el *perfil
personalizado*. El preset asigna de una vez el conjunto de permisos típico del
cargo; el perfil personalizado activa permisos puntuales uno a uno.

#figure(
  image("capturas-admin/30-staff-crear.png", width: 82%),
  caption: [Alta de un administrador con preset de rol],
)

== 12.3 Permisos (RBAC)

El detalle del administrador muestra los *18 permisos granulares* del sistema,
agrupados por dominio (gestión de colegiados, contenido, notificaciones,
catálogos, proyectos, tickets y auditoría). El rol es solo una etiqueta
descriptiva: *siempre mandan los permisos marcados*.

#figure(
  image("capturas-admin/31-staff-detalle-permisos.png", width: 82%),
  caption: [Permisos granulares de un administrador],
)

== 12.4 Ceder el rol de Super Usuario

El rol de *SUDO* solo puede transferirse: el botón *Ceder SUDO* abre un diálogo
para elegir al administrador destinatario y *confirmar con la contraseña* del
cedente. La sucesión se registra en la bitácora de auditoría.

#figure(
  image("capturas-admin/32-staff-ceder-sudo.png", width: 78%),
  caption: [Diálogo de cesión del rol de Super Usuario],
)

#aviso[Nunca crees cuentas con SUDO por comodidad: el acceso total no es
necesario para tareas operativas y concentra el riesgo en pocas cuentas. Usa
presets o permisos puntuales.]

// ============================================================================
// 13. USO EN MOVIL
// ============================================================================
= 13. Uso en móvil

El panel es utilizable desde el celular: el acceso y el dashboard se reorganizan
para pantallas pequeñas.

#figure(
  image("capturas-admin/33-acceso-movil.png", width: 34%),
  caption: [Acceso al panel en versión móvil],
)

#figure(
  image("capturas-admin/34-dashboard-movil.png", width: 34%),
  caption: [Panel de gestión en versión móvil],
)

#tip[Las vistas de *gestión* (fichas, kanban, auditoría) se recomiendan en
pantalla amplia; en móvil se usan para consultas rápidas y atención de tickets.]

// ============================================================================
// 14. SOPORTE Y RECOMENDACIONES
// ============================================================================
= 14. Soporte y recomendaciones

+ *No recuerdo mi contraseña*: la administración puede reemplazarla desde la
  cuenta del administrador (gestión de personal).
+ *"Demasiadas solicitudes" al operar*: espera un minuto y reintenta; es la
  protección anti-abuso compartida por todo el sistema.
+ *Un listado no muestra datos*: revisa los filtros y el buscador antes de
  registrar el incidente; si el problema persiste, consulta la bitácora de
  auditoría y restringe la causa.
+ *Cambios que no se guardan en la ficha de un psicólogo*: verifica que se haya
  indicado el *motivo del cambio*; sin motivo, el sistema rechaza la edición.
+ *Recargar una sección no debe cerrar tu sesión*: si vuelves a ver la pantalla
  de acceso sin haber salido, reporta la fecha y hora para revisar la bitácora.

#block(
  width: 100%,
  fill: azul,
  radius: 8pt,
  inset: 14pt,
)[
  #text(fill: white, size: 11pt)[
    *Colegio de Psicólogos del Estado Carabobo* — Panel Administrativo.
    Ante cualquier duda operativa, acude a la administración del sistema o abre
    una incidencia desde el panel.
  ]
]