// web/src/lib/terminos/parte-ii-agremiado.ts
//
// Documento del portal del agremiado. Solo lo ve quien tiene sesión: la ruta
// `/psi/terminos` redirige a `/terminos` si no la hay. Por eso este archivo
// SÍ puede mencionar el portal, el expediente y la solvencia sin filtrar nada.
//
// Aun así, dos reglas que vienen de la ronda anterior y que hay que conservar:
//   · No se dice "Parte I". Se dice "las condiciones de uso del sitio" (que es
//     donde está la sección 10.4 que este documento remite). La numeración en
//     romano confirmaría que hay otro documento.
//   · La sección 12 dice qué significa "retirar" datos: dejar de aparecer en el
//     directorio, NO borrar el expediente. Sin esa aclaración, el Colegio
//     promete algo que su propia normativa gremial le impide cumplir.
//
// Los datos del Colegio se interpolan desde `colegio.ts`, nunca a mano.

import type { TerminosParte, TerminosSeccion } from "./index";
import { COLEGIO, CORREO_PRIVACIDAD } from "./colegio";

/** El texto legal está escrito. El modal de aceptación deja de avisar. */
export const PENDIENTE = false;

const _SECCIONES: TerminosSeccion[] = [
  {
    numero: "1",
    titulo: "Quién es agremiado y cómo se accede",
    bloques: [
      {
        tipo: "lista",
        items: [
          "La cuenta del portal **la crea el Colegio** al aprobar una solicitud de inscripción. Usted no puede registrarse por sí mismo.",
          "El acceso está reservado a los profesionales inscritos en el Colegio de Psicólogos del Estado Carabobo.",
          "El Colegio puede **suspender o inactivar** una cuenta cuando corresponda conforme a sus normas.",
        ],
      },
    ],
  },
  {
    numero: "2",
    titulo: "Cuenta, contraseña y sesión",
    bloques: [
      {
        tipo: "lista",
        ordenada: true,
        items: [
          "Sus credenciales son **personales e intransferibles**. Usted es responsable de toda actividad realizada con su cuenta y debe avisar de inmediato al Colegio si sospecha un uso indebido.",
          "Su contraseña debe tener al menos 8 caracteres, sin espacios, con mayúscula, minúscula, número y símbolo.",
          "La sesión dura **24 horas**. **Cada inicio de sesión invalida los anteriores**: si entra desde otro dispositivo, la sesión del primero deja de funcionar. Cerrar sesión cierra todas las sesiones abiertas. Cambiar la contraseña también reinicia la sesión.",
          "Para **guardar cambios en su perfil** el portal le pedirá su contraseña actual, como confirmación de que es usted.",
          "El enlace de recuperación de contraseña vence en **1 hora** y solo se puede usar una vez.",
          "El acceso al portal está sujeto a **límites de intentos** de inicio de sesión por seguridad.",
        ],
      },
    ],
  },
  {
    numero: "3",
    titulo: "Correos y teléfono: qué debe registrar y por qué",
    bloques: [],
  },
  {
    numero: "3.1",
    titulo: "El correo de inicio de sesión no debe ser su correo de contacto público",
    bloques: [
      {
        tipo: "lista",
        items: [
          "El **correo principal (institucional)** es el que usted usa para iniciar sesión y al que el Colegio envía sus comunicados. **No lo publique como correo de contacto.**",
          "Como el inicio de sesión admite el correo como identificador, **publicarlo expone una de las dos credenciales de su cuenta**.",
          "Si desea recibir consultas del público, registre en «Email de contacto (gremial)» **un correo distinto** del de inicio de sesión. Usted se compromete a mantener ambos separados.",
        ],
      },
    ],
  },
  {
    numero: "3.2",
    titulo: "Teléfono de contacto",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "El teléfono registrado ante el Colegio debe ser, idealmente, **uno con el que usted pueda ser localizado con facilidad** (preferiblemente un número móvil o de WhatsApp que atienda con regularidad). Lo usa el Colegio para comunicarse con usted y, si usted lo decide, también puede mostrarse al público. Es su responsabilidad **mantenerlo actualizado**; el Colegio no responde por comunicaciones que no lleguen por datos de contacto desactualizados.",
      },
    ],
  },
  {
    numero: "4",
    titulo: "Su información es privada por defecto; usted decide qué publicar",
    bloques: [],
  },
  {
    numero: "4.1",
    titulo: "Principio general",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "**Su información completa es privada por defecto.** El Colegio no publica sus datos de contacto, ubicación ni información académica en el directorio a menos que **usted lo active manualmente**, dato por dato, en la pestaña **Privacidad y Visibilidad** de su perfil. Usted decide qué mostrar dentro de los campos que la plataforma ofrece y puede cambiar su decisión en cualquier momento.",
      },
    ],
  },
  {
    numero: "4.2",
    titulo: "Datos controlados por interruptores de visibilidad",
    bloques: [
      { tipo: "parrafo", texto: "Usted controla, uno por uno, si se publican:" },
      {
        tipo: "lista",
        items: [
          "**Carabobo:** correo de contacto, dirección de consulta, municipio, teléfono fijo, celular.",
          "**Fuera de Carabobo:** estado, municipio o ciudad, teléfono fijo, celular, dirección de consulta.",
          "**Exterior:** teléfono fijo, celular, dirección.",
          "**Académico:** universidad, fecha de grado, mención.",
          "**Modalidad de servicio** (presencial, en línea, telefónica).",
        ],
      },
    ],
  },
  {
    numero: "4.3",
    titulo: "Contenidos de presentación pública",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "Algunos contenidos están pensados precisamente para mostrarse en su ficha, y **se publican al cargarlos**: fotografía de perfil, mini-biografía, perfil completo, áreas de trabajo, redes sociales (que aparecen bajo el código QR) y **posgrados** registrados en «Formación Académica». **No los cargue si no desea que sean visibles.** Estos campos no tienen un interruptor propio: la forma de mantenerlos privados es no subirlos.",
      },
    ],
  },
  {
    numero: "4.4",
    titulo: "Identificación mínima",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "Sus **nombres, cédula y número de FPV** forman parte de su condición de profesional colegiado: se muestran siempre en su ficha y pueden consultarse también en el sitio web de la FPV. Esta identificación no se oculta con los interruptores de la sección 4.2.",
      },
    ],
  },
  {
    numero: "4.5",
    titulo: "Indexación y compartición",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "Las fichas públicas forman parte del sitio y **pueden ser encontradas por buscadores** (el sitio publica un mapa del sitio con las fichas del directorio), compartidas mediante enlace o código QR, y guardadas en copias por terceros. Si usted oculta un dato o retira su ficha, **el Colegio lo retirará de la plataforma, pero no controla las copias que buscadores o terceros ya hayan hecho**, cuya actualización puede tardar. Decida con cuidado qué publica.",
      },
    ],
  },
  {
    numero: "4.6",
    titulo: "Estadísticas de su ficha",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "El Colegio registra de forma **agregada** cuántas veces se visualizan los perfiles y cuántas búsquedas se realizan en el directorio. El Colegio **no le muestra estadísticas individuales sobre su propia ficha**: los datos se consulter en conjunto y no se atribuyen a un agremiado en particular.",
      },
    ],
  },
  {
    numero: "5",
    titulo: "Datos que gestiona el Colegio (usted los ve, pero no los edita)",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "Usted es responsable de su presencia pública y de su formación de posgrado. En cambio, el **título de pregrado, los datos del registro profesional, los documentos de su expediente, su expediente deontológico y las observaciones internas** los registra y gestiona el Colegio. Usted solo puede sustituir o retirar los soportes de imagen de su título. **Para corregir cualquier dato institucional debe solicitarlo al Colegio** mediante una solicitud, y el Colegio podrá pedirle documentos de respaldo.",
      },
      {
        tipo: "parrafo",
        texto:
          "Usted declara que la información y documentos que suministre son **veraces, vigentes y propios**, y se compromete a actualizarlos cuando cambien.",
      },
    ],
  },
  {
    numero: "6",
    titulo: "Solvencia: condición para el beneficio de publicar su información",
    bloques: [
      {
        tipo: "lista",
        ordenada: true,
        items: [
          "La publicación de su información en el directorio (dentro de lo que usted decida mostrar) es un **beneficio de la condición de agremiado y requiere estar solvente con el Colegio**.",
          "El portal le muestra su estatus (**AL DÍA** o **INSOLVENTE**). El público **no ve** su solvencia.",
          "Mientras usted esté insolvente, el Colegio **restringe de forma automática la visibilidad de su ficha**. El público sigue viendo sus datos de identificación —nombres, cédula, número de FPV, género, fotografía y universidad— y **deja de ver el resto de lo que usted publicó**: datos de contacto, ubicación, modalidad de servicio, presentación profesional, áreas de trabajo, posgrados y redes sociales. Sus datos **no se eliminan**: se restablece su visibilidad conforme a sus decisiones cuando regularice su situación.",
          "Si considera que su estatus es erróneo, puede pedir su revisión al Colegio mediante una solicitud.",
        ],
      },
    ],
  },
  {
    numero: "7",
    titulo: "Biblioteca virtual: solo para agremiados solventes",
    bloques: [
      {
        tipo: "lista",
        items: [
          "El acceso a la biblioteca virtual (libros y documentos digitales del Colegio) es **exclusivo para agremiados solventes**. Si usted está insolvente, el acceso aparece bloqueado y su sesión en la biblioteca podrá ser retirada.",
          "El acceso es **personal e intransferible**: no puede compartir su sesión ni credenciales con terceros.",
          "Los materiales están protegidos por derechos de autor. Se ofrecen para **uso personal y profesional**; **no puede copiarlos, redistribuirlos ni publicarlos**.",
          "El servicio depende de una plataforma externa y puede no estar disponible en ocasiones.",
        ],
      },
    ],
  },
  {
    numero: "8",
    titulo: "Notificaciones y noticias privadas del gremio: usted debe estar pendiente",
    bloques: [],
  },
  {
    numero: "8.1",
    titulo: "Notificaciones",
    bloques: [
      {
        tipo: "lista",
        ordenada: true,
        items: [
          "Las **notificaciones del portal** son un **canal oficial de comunicación** del Colegio con sus agremiados (convocatorias, avisos administrativos, solvencia, cambios de estatus, respuestas a solicitudes, etc.). Pueden dirigirse a todos los agremiados, a un grupo (por municipio, estado, género, área o solvencia) o a personas concretas.",
          "Usted debe **ingresar al portal con regularidad** y revisar sus notificaciones. El Colegio **no garantiza** que se envíe además un aviso por correo u otro medio: actualmente las notificaciones se entregan **solo dentro del portal**.",
          "**Abrir una notificación no la marca como leída**: usted debe pulsar «Marcar como leída». Esto es para su control; no cambia el hecho de que el Colegio la haya puesto a su disposición.",
          "Una vez publicada en su portal, la notificación se considera **puesta en su conocimiento**, y el desconocimiento por no haber ingresado no impide que surta efecto conforme a las normas gremiales.",
          "Las notificaciones **no pueden editarse ni reenviarse** una vez emitidas; si el Colegio corrige un aviso, lo hará mediante una nueva comunicación.",
          "Las fechas y horas se muestran en la **hora de Caracas**.",
        ],
      },
    ],
  },
  {
    numero: "8.2",
    titulo: "Noticias privadas (solo agremiados)",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "El Colegio publica noticias que **no son públicas** y están dirigidas **solo a los agremiados**. Usted debe **mantenerse atento a ellas**, porque pueden contener información relevante para el ejercicio profesional y la vida gremial. Su contenido es **confidencial del gremio**: no debe difundirlo fuera del colectivo sin autorización del Colegio, salvo que la propia noticia indique lo contrario.",
      },
    ],
  },
  {
    numero: "9",
    titulo: "Solicitudes al Colegio",
    bloques: [
      {
        tipo: "lista",
        items: [
          "Las solicitudes son el **canal formal** para consultas, correcciones de datos y trámites ante el Colegio.",
          "Cada motivo de atención tiene un **límite de solicitudes abiertas simultáneas** definido por el Colegio; no puede enviar más de **3 mensajes seguidos** sin respuesta del Colegio; y los textos tienen límites de extensión.",
          "El Colegio puede **pausar temporalmente la recepción** de solicitudes y lo indicará en el portal.",
          "Debe tratar al personal del Colegio con respeto. Las solicitudes abusivas, falsas o reiteradas sin fundamento podrán ser cerradas.",
          "Los cambios de estado de sus solicitudes le serán notificados en el portal.",
        ],
      },
    ],
  },
  {
    numero: "10",
    titulo: "Contenido que usted publica y conducta",
    bloques: [
      { tipo: "parrafo", texto: "Usted es **el único responsable** del contenido de su ficha. Se compromete a que:" },
      {
        tipo: "lista",
        ordenada: true,
        items: [
          "Sea **veraz** y no induzca a error sobre su formación, especialidad, títulos o resultados.",
          "Cumpla el **Código de Ética del Psicólogo**, la **Ley de Ejercicio de la Psicología**, los Estatutos y reglamentos aplicables, en especial en materia de **publicidad profesional**, confidencialidad y respeto a los usuarios de sus servicios.",
          "No contenga datos personales ni casos de pacientes, ni información de terceros sin su consentimiento.",
          "No vulnere derechos de autor o de imagen (use solo fotografías y textos propios o autorizados).",
          "No incluya contenido ofensivo, discriminatorio, ilícito o engañoso.",
          "Los enlaces a redes sociales lleven a perfiles **propios** y **profesionales**.",
        ],
      },
      {
        tipo: "parrafo",
        texto:
          "El contenido se **sanea** en el servidor por seguridad, y las imágenes se recomprimen a un tamaño estándar. El Colegio puede **retirar, ocultar o solicitar la corrección** de contenido que incumpla estos términos o las normas gremiales, sin perjuicio de las actuaciones disciplinarias que correspondan.",
      },
    ],
  },
  {
    numero: "11",
    titulo: "Contactos de emergencia y datos de terceros",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "Usted puede registrar **hasta 3 contactos de emergencia**. Al hacerlo declara que **cuenta con el consentimiento de esas personas** para que el Colegio use sus datos exclusivamente para localizarlas ante una situación de emergencia o cuando no sea posible contactarlo a usted. Estos datos **nunca se publican** en el directorio ni en la ficha pública. Usted debe mantenerlos actualizados.",
      },
    ],
  },
  {
    numero: "12",
    titulo: "Tratamiento de sus datos por el Colegio",
    bloques: [
      {
        tipo: "lista",
        items: [
          "El Colegio trata sus datos para **gestionar su condición de agremiado**: registro, solvencia, comunicaciones, expediente, soporte, biblioteca virtual y cumplimiento de sus funciones gremiales.",
          "El **expediente** (documentos, expediente deontológico y observaciones internas) es de **uso interno** y no se publica.",
          "Las **acciones administrativas sobre su registro quedan en una bitácora de auditoría** (qué cambió, quién y cuándo); las ediciones que el personal hace a su ficha exigen un **motivo escrito**.",
          "Si usted activa la opción **«Autorizar aviso de cumpleaños»**, el Colegio podrá mostrarlo en el aviso interno de cumpleaños del panel administrativo. Si la desactiva, no aparecerá.",
          "Usted puede ejercer sus derechos de **acceso, rectificación, actualización y retiro** de datos según lo indicado en la sección 10.4 de las condiciones de uso del sitio. Ciertos datos deben conservarse mientras dure su condición de agremiado y por los plazos que exijan las normas gremiales o legales.",
          "Para **dejar en blanco un campo ya guardado** (no solo ocultarlo) debe solicitarlo al Colegio.",
        ],
      },
      {
        tipo: "aviso",
        tono: "info",
        texto:
          "**Qué significa retirar sus datos en la práctica:** si el Colegio atiende un retiro de la información que usted publica, su ficha deja de aparecer en el directorio. Esto **no elimina su expediente** ni los documentos y datos que el Colegio debe conservar por normas gremiales o legales. Para cualquier otra rectification o eliminación, escriba a " +
          CORREO_PRIVACIDAD +
          ".",
      },
    ],
  },
  {
    numero: "13",
    titulo: "Suspensión y terminación",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "El Colegio podrá suspender o limitar el acceso al portal, a la biblioteca virtual o a la visibilidad de su ficha cuando: (a) exista insolvencia; (b) se incumplan estos términos o las normas gremiales; (c) se detecte uso indebido de la cuenta; o (d) cese su condición de agremiado. Usted podrá solicitar la revisión de estas medidas por el canal de solicitudes.",
      },
    ],
  },
  {
    numero: "14",
    titulo: "Cambios y aceptación",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "El Colegio puede modificar estas condiciones. Se le informará mediante **notificación en el portal**, y el uso continuado del portal implica su aceptación. Si no está de acuerdo, puede comunicarlo al Colegio por el canal de solicitudes.",
      },
      {
        tipo: "parrafo",
        texto:
          "La aceptación se registra con la **fecha, la hora y la versión** de estas condiciones, y puede consultarse en su portal.",
      },
    ],
  },
];

const parteIIAgremiado: TerminosParte = {
  clave: "agremiado",
  // Solo la ve un agremiado con sesión, pero el rótulo ya no usa numeración de
  // partes: se lee "Condiciones del portal" y no "Parte II".
  etiqueta: "Condiciones del portal",
  titulo: "Condiciones de uso del portal del agremiado",
  resumen: `Regula el acceso al portal privado del ${COLEGIO.nombre} y el tratamiento de los datos de los agremiados inscritos, sus documentos digitales y su expediente.`,
  secciones: _SECCIONES,
};

export default parteIIAgremiado;
