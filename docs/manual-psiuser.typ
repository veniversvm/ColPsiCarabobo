// ============================================================================
// Manual del Usuario — Portal del Psicólogo (Colegio de Psicólogos de Carabobo)
// Compilar:  typst compile manual-psiuser.typ manual-psiuser.pdf
// Desde:     docs/  (las imágenes son relativas a este archivo → capturas/*.png)
// ============================================================================

// ── Paleta institucional ─────────────────────────────────────────────────────
#let azul   = rgb("#14328c")   // azul institucional
#let amar   = rgb("#facc15")   // amarillo institucional
#let rojo   = rgb("#b91c1c")
#let gris   = rgb("#6b7280")
#let fondo  = rgb("#f3f6fb")

#set document(title: "Manual del Usuario — Portal del Psicólogo")
#set page(
  paper: "a4",
  margin: (x: 1.9cm, y: 2cm),
  footer: context {
    set text(fill: gris, size: 8.5pt)
    align(center + horizon)[
      Colegio de Psicólogos del Estado Carabobo · Portal del Psicólogo · Página #counter(page).get().first()
    ]
  },
)
#set text(font: ("Noto Sans", "Noto Sans CJK SC"), lang: "es", region: "VE", size: 10pt)
#set par(justify: true, leading: 0.62em)
// Los títulos llevan su número en el texto (1. Introducción, 4.8 Contacto de
// Emergencia…). La numeración automática está desactivada para no duplicar el
// número en el render ni en el índice (#outline).
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

// ============================================================================
// PORTADA
// ============================================================================
#align(center)[
  #block(width: 100%, fill: azul, inset: (y: 26pt), radius: 10pt)[
    #text(size: 22pt, weight: "bold", fill: white)[Manual del Usuario]
    #v(4pt)
    #text(size: 17pt, fill: white)[Portal del Psicólogo]
  ]
  #v(10pt)
  #text(size: 26pt, weight: "bold", fill: azul)[Colegio de Psicólogos]\
  #text(size: 26pt, weight: "bold", fill: azul)[del Estado Carabobo]
  #v(6pt)
  #text(size: 12pt, fill: gris)[Portal gremial del agremiado — acceso, perfil, documentos y solicitudes]
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
    #text(size: 10pt)[Psicólogos colegiados con acceso al portal gremial]
  ]],
)

#v(18pt)

#block(width: 100%, fill: rgb("#fef3c7"), stroke: 0.8pt + amar, radius: 8pt, inset: 12pt)[
  #text(weight: "bold", fill: rgb("#8a6d00"))[Aviso] — Las capturas que acompañan este manual
  son ilustrativas y pueden variar ligeramente según la información de cada
  agremiado y la versión vigente del portal.
]

#v(16pt)

// ── Índice ───────────────────────────────────────────────────────────────────
#outline(indent: auto, depth: 2)

// ============================================================================
// 1. INTRODUCCIÓN
// ============================================================================
= 1. Introducción

El *Portal del Psicólogo* es el espacio personal del agremiado dentro del sitio
del Colegio de Psicólogos del Estado Carabobo. Desde allí cada colegiado puede:

+ Consultar su *estatus de solvencia* y su número de Federación de Psicólogos
  de Venezuela (FPV).
+ Mantener actualizado su *perfil profesional* (identidad, contacto, ubicación,
  formación, modalidad de servicio, privacidad y contacto de emergencia).
+ Registrar sus *postgrados* y mantener sus *documentos* al día.
+ Recibir *notificaciones* gremiales.
+ Consultar y descargar el material de la *Biblioteca Virtual* del Colegio (ver
  sección 9).
+ Abrir y dar seguimiento a *solicitudes (tickets)* ante la administración.

El portal es de uso personal: lo que se guarda en el perfil alimenta el
*Directorio Público* solo si el propio agremiado decide mostrarlo (ver sección
4.7). El resto de la información es interna del Colegio.

#tip[Anota tu *usuario o correo* y tu *contraseña* en un lugar seguro.
Sin ellos no podrás ingresar, y la recuperación de contraseña lleva unos minutos
de verificación.]

// ============================================================================
// 2. INGRESO AL PORTAL
// ============================================================================
= 2. Ingreso al portal

El portal se abre desde la página principal del Colegio, botón *Portal de
Agremiados*, o directamente en la dirección del portal. La pantalla de ingreso
solicita dos datos:

+ *Identificación*: tu nombre de usuario o el correo principal con el que estás
  registrado.
+ *Contraseña*: la clave personal de acceso.

#figure(
  image("capturas/01-login.png", width: 78%),
  caption: [Pantalla de ingreso al portal],
)

Si la identificación o la contraseña son incorrectas, el sistema muestra un
aviso de *Error de acceso* y no permite continuar:

#figure(
  image("capturas/02-login-error.png", width: 78%),
  caption: [Aviso de credenciales incorrectas],
)

+ Si olvidaste la contraseña, usa el enlace _¿Olvidaste tu contraseña?_ que
  aparece bajo el formulario y sigue las instrucciones que llegarán a tu correo.
+ Recuerda: el usuario y el correo se registran en minúsculas; no importa cómo
  los escribas, el sistema los normaliza. La contraseña, en cambio, distingue
  mayúsculas y minúsculas.

#tip[Si recibes repetidamente el mensaje *"Demasiadas solicitudes"*, espera un
minuto e inténtalo de nuevo: es una medida de seguridad del Colegio contra
intentos masivos, y protege tu cuenta.]

#nota[Tu sesión queda activa hasta por *24 horas*. Al cerrar sesión o cambiar
la contraseña, las sesiones anteriores se invalidan; vuelve a ingresar con tus
credenciales.]

// ============================================================================
// 3. PANEL PRINCIPAL
// ============================================================================
= 3. Panel principal (inicio)

Al ingresar se muestra el *panel principal* (dashboard), que reúne en una sola
pantalla la información más consultada:

#figure(
  image("capturas/03-dashboard.png", width: 82%),
  caption: [Panel principal del agremiado],
)

+ *Estatus de Solvencia*: muestra si el agremiado está *AL DÍA* o *INSOLVENTE*,
  junto con su número de FPV. Los agremiados solventes tienen acceso a todos
  los servicios, incluida la Biblioteca Virtual.
+ *Biblioteca Virtual*: acceso al material de lectura del Colegio, que puedes
  consultar y descargar. Solo está disponible para agremiados solventes; si no
  lo estás, el acceso aparece bloqueado con un candado (ver sección 9).
+ *Acceso rápido*: seis atajos a *Mi Perfil*, *Académico*, *Mis Documentos*,
  *Notificaciones*, *Solicitudes* y *Manual*.
+ *Noticias gremiales*: las publicaciones más recientes del Colegio.

Debajo del acceso rápido, el Colegio recuerda que el uso del portal supone
aceptar los *Términos y Condiciones* del portal, con un enlace a su texto
completo.

En pantallas de celular el panel se reorganiza automáticamente para verse y
usarse con una mano:

#figure(
  image("capturas/18-dashboard-movil.png", width: 34%),
  caption: [Panel principal en versión móvil],
)

#tip[Consulta el *estatus de solvencia* antes de realizar trámites: varias
gestiones y la Biblioteca Virtual requieren estar al día con el gremio.]

// ============================================================================
// 4. MI PERFIL
// ============================================================================
= 4. Mi Perfil

Desde *Mi Perfil* (botón _Mi Perfil_ del panel principal) se administra toda la
información personal y profesional. El perfil se organiza como una *libreta de
pestañas*; cada pestaña guarda un grupo de datos. Al terminar los cambios pulsa
el botón *Guardar* que aparece al final del formulario y confirma con tu
*contraseña actual*.

#nota[Los datos del perfil son personales. El Colegio solo publica en el
Directorio lo que tú decidas mostrar en la pestaña *Privacidad y Visibilidad*
(ver sección 4.7).]

== 4.1 Cuenta y Seguridad

Identidad de acceso al portal: *nombre de usuario*, *correo principal* y
*cambio de contraseña*.

#figure(
  image("capturas/04-perfil-cuenta.png", width: 82%),
  caption: [Pestaña Cuenta y Seguridad],
)

+ Para cambiar la contraseña escribe la nueva clave en *Nueva Contraseña* y
  confírmala; el sistema valida en vivo que tenga al menos 8 caracteres e
  incluya mayúscula, minúscula, número y símbolo.
+ Si cambias la contraseña, recibirás una *nueva sesión* automáticamente; las
  demás pestañas o dispositivos donde estuvieras logueado quedarán sin acceso.

#tip[No compartas tu contraseña con terceros. El Colegio jamás te la pedirá por
correo, teléfono o redes sociales.]

== 4.2 Información de Contacto

Datos de contacto gremiales: *correo de contacto*, *teléfono fijo* y *teléfono
móvil / WhatsApp*.

#figure(
  image("capturas/05-perfil-contacto.png", width: 82%),
  caption: [Pestaña Información de Contacto],
)

+ El *correo de contacto* puede ser distinto del correo principal de acceso.
+ Los números se registran con el formato venezolano (p. ej. 0412-1234567);
  el sistema los normaliza al guardar.

== 4.3 Expediente Académico

Datos de tu carrera de pregrado que respaldan tu ejercicio profesional, con sus
documentos.

#figure(
  image("capturas/06-perfil-expediente.png", width: 82%),
  caption: [Pestaña Expediente Académico],
)

+ Registra la *universidad*, *fecha de egreso*, *mención*, *número de registro*
  y *folio/tomo*.
+ Adjunta la *imagen del título* y hasta dos documentos adicionales de respaldo.
+ Los documentos se eliminan suavemente al guardar si marcas la opción de
  quitarlos.

== 4.4 Ubicación Geográfica

Lugares donde ejerces tu consulta: *dentro de Carabobo*, *fuera de Carabobo* e
*internacional*.

#figure(
  image("capturas/07-perfil-ubicacion.png", width: 82%),
  caption: [Pestaña Ubicación Geográfica],
)

+ Para cada ubicación puedes indicar si es *Carabobo*, *Venezuela (fuera de
  Carabobo)* o *exterior*, con su municipio/estado/país, teléfonos y dirección
  del consultorio.
+ Se exige al menos un bloque de ubicación completado para que el perfil esté
  aprobado para el Directorio.

== 4.5 Perfil Profesional

Tu presentación ante el gremio y el público: *área de trabajo principal* y
*secundaria*, y un resumen de tu práctica profesional.

#figure(
  image("capturas/08-perfil-profesional.png", width: 82%),
  caption: [Pestaña Perfil Profesional],
)

+ El área de trabajo se elige del catálogo del Colegio (la que no esté en el
  catálogo no podrá mostrarse públicamente).
+ El *resumen profesional* se muestra en tu ficha del Directorio si así lo
  autorizas en Privacidad.

== 4.6 Servicio y Preferencias

Modalidad de atención que ofreces: *presencial*, *a distancia (en línea)* y
*telefónica*, y si deseas mostrarla en el directorio público.

#figure(
  image("capturas/09-perfil-servicio.png", width: 82%),
  caption: [Pestaña Servicio y Preferencias],
)

+ Marca una o varias modalidades y activa "Mostrar mi modalidad en el directorio
  público" si quieres que los visitantes la vean.
+ La *autorización de aviso de cumpleaños* solo informa de tu cumpleaños a la
  administración; no se publica.

== 4.7 Privacidad y Visibilidad

El corazón del control de datos: decide, campo por campo, qué se muestra en el
Directorio público y qué queda solo para el Colegio.

#figure(
  image("capturas/10-perfil-privacidad.png", width: 82%),
  caption: [Pestaña Privacidad y Visibilidad],
)

+ Puedes alternar la visibilidad de *email de contacto*, *dirección de
  consulta*, *municipio*, *teléfonos*, *universidad*, *modalidad de servicio*,
  entre otros.
+ *Redes Sociales*: se gestiona en el mismo perfil; vincula tus perfiles
  profesionales para que otros psicólogos y el público te encuentren. Cada red
  se quita de tu perfil público al eliminarla.

#tip[Revisa esta pestaña periódicamente: controla qué información personal ves
públicamente expuesta en el Directorio.]

== 4.8 Contacto de Emergencia

Personas a las que el Colegio puede avisar si no se logra ubicar al agremiado
o ante una situación de emergencia.

#figure(
  image("capturas/11-perfil-emergencia.png", width: 82%),
  caption: [Pestaña Contacto de Emergencia],
)

+ Puedes registrar hasta *3 personas*. Para cada una se exige *nombre*,
  *parentesco* y *al menos un medio de contacto* (teléfono o correo).
+ Son *datos privados*: jamás se publican en el Directorio ni en el perfil
  público; solo los usa la administración.

#nota[El contacto de emergencia se guarda con su propio botón
(no necesita la contraseña del perfil).]

// ============================================================================
// 5. FORMACIÓN ACADÉMICA (POSTGRADOS)
// ============================================================================
= 5. Formación Académica (postgrados)

Desde el panel principal, botón *Postgrados*, puedes registrar tus estudios de
posgrado: especializaciones, maestrías, doctorados y diplomados.

#figure(
  image("capturas/12-postgrados.png", width: 82%),
  caption: [Listado de formación académica],
)

+ Pulsa *+ Nuevo Título* para agregar un postgrado: *tipo de estudio*, *título*,
  *universidad*, *año de egreso*, *descripción* y hasta *3 archivos de respaldo*
  (certificados).
+ Los títulos ya registrados pueden *editarse* o *eliminarse* desde el listado.
+ La eliminación es segura: conserva un registro de auditoría y los archivos se
  limpian del almacenamiento del Colegio.

#tip[Mantén los certificados al día: un expediente académico completo respalda
tu perfil ante el Colegio y ante terceros que consulten el Directorio.]

// ============================================================================
// 6. MIS DOCUMENTOS
// ============================================================================
= 6. Mis Documentos

Desde *Mis Documentos* puedes ver los documentos registrados en tu ficha
(foto, cédula, título, RIF y otros) según su tipo.

#figure(
  image("capturas/13-documentos.png", width: 82%),
  caption: [Vista de Mis Documentos],
)

+ Los documentos se clasifican por tipo y se muestran con su estado de
  registro.
+ Si un documento figura como *pendiente*, contacta a la administración para
  regularizar tu expediente.

#tip[Tu expediente completo es clave para trámites como constancias de
solvencia. Si detectas un documento faltante, escríbenos por *Solicitudes*
(sección 8).]

// ============================================================================
// 7. NOTIFICACIONES
// ============================================================================
= 7. Notificaciones

El Colegio te comunica novedades gremiales a través de *Notificaciones*; el
badle rojo del panel principal muestra cuántas tienes sin leer.

#figure(
  image("capturas/14-notificaciones.png", width: 82%),
  caption: [Bandeja de notificaciones],
)

+ Cada notificación muestra su *título*, *mensaje* y *fecha*; las no leídas se
  distinguen visualmente.
+ Al abrir el panel, las notificaciones marcadas como leídas en esta sesión se
  actualizan en el contador.

#tip[Revisa las notificaciones con frecuencia: ahí se informan asambleas,
jornadas y comunicados importantes del Colegio.]

// ============================================================================
// 8. MIS SOLICITUDES (TICKETS)
// ============================================================================
= 8. Mis Solicitudes (tickets)

Las *Solicitudes* comunican con la administración: constancias, actualización
de datos, incidencias con el portal y otros motivos definidos por el Colegio.

== 8.1 Listado de solicitudes

El listado muestra tus solicitudes con su *estado* (Recibido, En proceso,
Cerrado), el *motivo*, la *fecha* y si tienen mensajes nuevos o no leídos.

#figure(
  image("capturas/15-tickets.png", width: 82%),
  caption: [Listado de mis solicitudes],
)

+ Pulsa *Nueva Solicitud* para abrir una nueva.
+ Las solicitudes cerradas quedan visibles en el historial para consulta.

== 8.2 Crear una solicitud

Completa el *motivo* (el desplegable muestra los motivos que ofrece el Colegio),
un *título* breve, la *descripción* del caso y, si lo deseas, *archivos
adjuntos* de respaldo.

#figure(
  image("capturas/16-ticket-crear.png", width: 82%),
  caption: [Formulario de nueva solicitud],
)

+ El motivo, el título (hasta 200 caracteres) y la descripción son
  obligatorios; el envío queda deshabilitado hasta completarlos.
+ Puedes adjuntar uno o varios archivos; el Colegio los conserva como respaldo
  de la conversación.

== 8.3 Seguimiento y conversación

Una vez abierta, la solicitud se convierte en una *conversación* entre tú y la
administración. Puedes responder desde el mismo hilo y seguir el estado en el
encabezado.

#figure(
  image("capturas/17-ticket-detalle.png", width: 82%),
  caption: [Detalle de solicitud con conversación],
)

+ Cada mensaje identifica su autor (agremiado o administración) y su fecha.
+ La conversación queda abierta mientras la solicitud no esté cerrada; al
  cerrarla (por ti o por la administración) ya no se pueden enviar mensajes.
+ Los mensajes no se pueden editar ni borrar: respóndelos con cuidado.

#tip[Describe tu solicitud con claridad y, si aplica, adjunta la documentación
de respaldo. Una solicitud bien explicada se resuelve en menos pasadas.]

// ============================================================================
// 9. BIBLIOTECA VIRTUAL
// ============================================================================
= 9. Biblioteca virtual

El Colegio pone a disposición de sus agremiados una Biblioteca Virtual con
material de lectura para la formación continua: libros y documentos digitales
que se consultan y se descargan desde el mismo lugar.

#figure(
  image("capturas/19-biblioteca.png", width: 82%),
  caption: [Tarjeta de la Biblioteca Virtual en el panel principal],
)

== 9.1 Quién tiene acceso

El acceso es exclusivo de los agremiados solventes. La tarjeta del panel
principal lo muestra siempre, pero:

+ Si estás al día, aparece con su flecha azul y abre la biblioteca.
+ Si estás insolvente, aparece con un candado y el mensaje "Ponte al día con el
  gremio para acceder a la biblioteca virtual"; no se puede abrir.

El requisito es el mismo del resto de los servicios: la solvencia anual. Si
regularizas tu cuota, el acceso se restablece automáticamente, sin que tengas
que pedir nada.

#nota[
  El acceso depende de tu estatus de solvencia, que se calcula a partir de las
  constancias registradas en la pestaña *Solvencias* de tu ficha. Si crees que
  estás al día y sin embargo te aparece bloqueado, abre una Solicitud
  (sección 8.2) indicando tu número de FPV.
]

== 9.2 Cómo entrar

Pulsa la tarjeta *Biblioteca Virtual*. La biblioteca se abre en una pestaña
nueva del navegador, con tu sesión ya iniciada: no te pide usuario ni
contraseña.

+ No necesitas recordar ninguna clave propia para la biblioteca. El Colegio
  genera internamente una clave asociada a tu cuenta y se encarga de
  entregártela por ti.
+ Puedes volver a entrar tantas veces como quieras desde la misma tarjeta, sin
  volver a escribir nada.
+ La sesión de la biblioteca es independiente de la del portal: cerrar sesión
  en el portal no te cierra la biblioteca, y viceversa.

== 9.3 Qué puedes hacer dentro

La biblioteca reúne libros y documentos digitales del Colegio. Dentro puedes:

+ *Consultar* el catálogo y abrir la ficha de cada título.
+ *Descargar* el material a tu equipo para leerlo con el programa que prefieras.

#figure(
  image("capturas/20-biblioteca-catalogo.png", width: 82%),
  caption: [Catálogo de la biblioteca abierto en el navegador],
)

Los títulos aparecen con el nombre del archivo, así que la primera vez conviene
abrir la ficha de cada uno y comprobar cuál es.

#tip[
  Descarga los documentos y guárdalos en una carpeta con un orden claro. El
  Colegio amplía la colección con el tiempo, y tenerlos separados por área o por
  año te facilita volver a consultarlos.
]

// ============================================================================
// 10. SOPORTE Y RECOMENDACIONES
// ============================================================================
= 10. Soporte y recomendaciones

+ *No recuerdo mi contraseña*: usa _¿Olvidaste tu contraseña?_ en la pantalla
  de ingreso.
+ *No puedo ingresar y el sistema dice "Demasiadas solicitudes"*: espera un
  minuto y reintenta (protección anti-abuso del Colegio).
+ *Mi perfil se ve incompleto en el Directorio*: revisa las pestañas
  *Expediente Académico*, *Ubicación Geográfica* y *Privacidad y Visibilidad*;
  las áreas de trabajo solo se muestran si están en el catálogo del Colegio.
+ *La Biblioteca Virtual me sale con candado*: revisa tu estatus de solvencia
  en el panel principal; solo los agremiados al día pueden entrar (sección 9.1).
+ *La biblioteca no me abre o se queda cargando*: es posible que el servicio
  esté temporalmente caído. Inténtalo en unos minutos y, si persiste, abre una
  Solicitud (sección 8.2) contándonos qué pasa.
+ *Necesito una constancia o corregir mis datos*: abre una *Solicitud* (ver
  sección 8.2) con el motivo correspondiente.
+ *No encuentro un documento*: verifica *Mis Documentos* y contacta a la
  administración si figura pendiente.

#block(
  width: 100%,
  fill: azul,
  radius: 8pt,
  inset: 14pt,
)[
  #text(fill: white, size: 11pt)[
    *Colegio de Psicólogos del Estado Carabobo* — Portal del Psicólogo.
    Ante cualquier duda, abre una solicitud desde tu portal o acude a la sede.
  ]
]