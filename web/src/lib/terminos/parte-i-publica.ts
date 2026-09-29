// web/src/lib/terminos/parte-i-publica.ts
//
// Documento público de Condiciones de Uso. Es el ÚNICO documento que ve quien
// no tiene sesión: nada en este archivo menciona otra parte, otro documento ni
// la existencia de condiciones adicionales. Si alguna vez se le agrega una
// frase que sugiera un segundo texto, se rompe la regla que sostiene esta
// página: un visitante anónimo no debe poder inferir que hay algo más.
//
// Los datos del Colegio NO se escriben a mano: se interpolan desde `colegio.ts`.
// Si aparece un RIF, un correo o una dirección repetidos, es un error.

import type { TerminosParte, TerminosSeccion } from "./index";
import { COLEGIO, CORREO_PRIVACIDAD, CORREO_REPORTES, FPV_URL } from "./colegio";

/** El texto legal está escrito. La banda roja de "pendiente" no debe verse.
//
// Si se edita este texto, lee `docs/plan-terminos-condiciones.md`: las
// afirmaciones de la sección 10.1 son un contrato con el backend
// (`ANALYTICS_RETENTION_DAYS`, la huella de la IP, la cookie de 30 días), y
// cambiarlas sin cambiar el código vuelve a hacer falsa la promesa. */
export const PENDIENTE = false;

const _SECCIONES: TerminosSeccion[] = [
  {
    numero: "1",
    titulo: "Quiénes somos y qué es esta plataforma",
    bloques: [
      {
        tipo: "parrafo",
        texto: `Esta plataforma es operada por el **${COLEGIO.nombre}** (en adelante, el Colegio), RIF ${COLEGIO.rif}, con domicilio en ${COLEGIO.direccion}, correo de contacto ${COLEGIO.correo}.`,
      },
      {
        tipo: "parrafo",
        texto:
          "La plataforma reúne tres espacios: el **sitio público**, con información institucional, el directorio de profesionales, las noticias de audiencia pública, el marco legal y el formulario de solicitud de inscripción; el **portal del agremiado**, de acceso restringido a los profesionales inscritos; y el **panel administrativo**, de uso interno del personal del Colegio.",
      },
      {
        tipo: "parrafo",
        texto: "Estos términos regulan el uso del sitio público.",
      },
    ],
  },
  {
    numero: "2",
    titulo: "Aceptación",
    bloques: [
      {
        tipo: "parrafo",
        texto: `Al navegar por el sitio, consultar el directorio, enviar una solicitud de inscripción o iniciar sesión, usted declara que ha leído y acepta estos términos. **Esto aplica también a quien navega sin haber iniciado sesión.** Si no está de acuerdo, le pedimos que no utilice la plataforma.`,
      },
      {
        tipo: "parrafo",
        texto:
          "Si usted es menor de edad, debe utilizar la plataforma con la supervisión de su representante legal.",
      },
    ],
  },
  {
    numero: "3",
    titulo: "Naturaleza informativa: esto no es un servicio de atención psicológica",
    bloques: [
      {
        tipo: "lista",
        items: [
          "La plataforma es **informativa e institucional**. **No presta servicios de atención psicológica, psicoterapia, diagnóstico ni orientación clínica.**",
          "El Colegio **no asigna, recomienda ni garantiza** a ningún profesional en particular, ni interviene en la relación que usted establezca con quien contacte.",
          "**La plataforma no es un servicio de emergencia.** Si usted o alguien cercano se encuentra en riesgo inmediato, comuníquese con los servicios de emergencia de su localidad o acuda al centro de salud más cercano.",
        ],
      },
    ],
  },
  {
    numero: "4",
    titulo: "Qué puede hacer sin iniciar sesión",
    bloques: [
      {
        tipo: "parrafo",
        texto: "Cualquier visitante, sin necesidad de cuenta, puede:",
      },
      {
        tipo: "lista",
        items: [
          "Consultar la portada, las secciones institucionales (`/nosotros`, `/explorar`) y las noticias de audiencia pública.",
          "**Buscar en el directorio** de profesionales (por texto, área de trabajo y ubicación) y abrir la **ficha pública** de cada profesional, incluido su código QR para compartirla.",
          "Leer el **marco legal** publicado en `/documentos`: Código de Ética, Estatutos de la FPV, Ley de Ejercicio de la Psicología y Reglamento Interno.",
          "Enviar una **solicitud de inscripción** al Colegio (ver la sección 8).",
          "Acceder a las pantallas de inicio de sesión y recuperación de contraseña.",
        ],
      },
      {
        tipo: "parrafo",
        texto:
          "Lo que **no** puede hacer sin sesión es acceder al portal del agremiado, a las noticias dirigidas solo a agremiados, a la biblioteca virtual, a las notificaciones internas ni a ningún dato que el profesional no haya decidido hacer público.",
      },
    ],
  },
  {
    numero: "5",
    titulo: "El directorio: qué información muestra y de dónde proviene",
    bloques: [],
  },
  {
    numero: "5.1",
    titulo: "Origen de la información",
    bloques: [
      {
        tipo: "parrafo",
        texto: `**La información del directorio corresponde a psicólogos inscritos en el Colegio de Psicólogos del Estado Carabobo.** El Colegio mantiene el registro de sus agremiados y publica en el directorio únicamente la información de esos profesionales.`,
      },
    ],
  },
  {
    numero: "5.2",
    titulo: "Información mínima de identificación",
    bloques: [
      {
        tipo: "parrafo",
        texto: `La información mínima de cada profesional (**nombres, cédula de identidad y número de FPV**) forma parte de su condición de profesional colegiado y **también puede consultarse en el sitio web de la [Federación de Psicólogos de Venezuela](${FPV_URL})**. Usted puede contrastar allí lo que ve en esta plataforma.`,
      },
    ],
  },
  {
    numero: "5.3",
    titulo: "Información adicional: publicada por decisión del profesional",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "Todo dato adicional que aparezca en una ficha (por ejemplo: correo o teléfono de contacto, dirección o ubicación del consultorio, modalidad de atención, datos académicos, presentación profesional, áreas de trabajo, redes sociales o fotografía) **es publicado bajo la responsabilidad y por decisión del propio profesional**, dentro de los campos que la plataforma permite. Cada profesional puede mostrar u ocultar buena parte de sus datos de contacto, ubicación y formación en cualquier momento.",
      },
      { tipo: "parrafo", texto: "Por eso:" },
      {
        tipo: "lista",
        items: [
          "**Que una ficha no muestre un dato (por ejemplo, un teléfono) no significa que el profesional no lo tenga**, sino que decidió no publicarlo.",
          "Los datos de contacto, cuando aparecen, deben usarse **únicamente para contactar al profesional** a propósito de sus servicios profesionales.",
        ],
      },
    ],
  },
  {
    numero: "5.4",
    titulo: "Lo que el directorio no muestra",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "El directorio **no muestra** los documentos del expediente del profesional, su expediente deontológico, observaciones internas del Colegio, sus contactos de emergencia, ni su situación de solvencia con el Colegio. **La ausencia de un profesional o de determinada información en el directorio no constituye juicio alguno del Colegio sobre su idoneidad, conducta o estado gremial.**",
      },
    ],
  },
  {
    numero: "5.5",
    titulo: "Áreas de trabajo",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "Las áreas de trabajo se muestran a partir del **catálogo oficial** del Colegio; los profesionales no pueden inventar áreas. El Colegio puede actualizar ese catálogo, por lo que un área puede dejar de mostrarse.",
      },
    ],
  },
  {
    numero: "6",
    titulo: "Alcance de la verificación",
    bloques: [
      { tipo: "parrafo", texto: "El Colegio verifica la inscripción de los profesionales en su registro. Sin embargo:" },
      {
        tipo: "lista",
        items: [
          "**No garantiza** la exactitud, actualización ni calidad de la información que cada profesional publica de forma voluntaria (contacto, biografía, formación complementaria, redes, etc.).",
          "**No garantiza** los resultados, la calidad ni la disponibilidad de los servicios de los profesionales listados.",
          "Le recomendamos **verificar la condición del profesional** con el Colegio o con la FPV antes de contratar sus servicios.",
        ],
      },
      {
        tipo: "parrafo",
        texto: `Si usted detecta información incorrecta, desactualizada o que considera indebida en una ficha, puede reportarla a ${CORREO_REPORTES}.`,
      },
    ],
  },
  {
    numero: "7",
    titulo: "Uso aceptable",
    bloques: [
      { tipo: "parrafo", texto: "Al usar la plataforma usted se compromete a **no**:" },
      {
        tipo: "lista",
        ordenada: true,
        items: [
          "**Extraer de forma masiva o automatizada** datos del directorio (scraping, recolección de correos o teléfonos, creación de bases de datos paralelas), ni utilizarlos para **publicidad no solicitada, spam, ventas o actividades de mercadeo**.",
          "Usar los datos de contacto de un profesional con fines distintos a contactarlo profesionalmente, ni para acosar, amenazar, suplantar o dañar su reputación.",
          "Hacerse pasar por otra persona, por un profesional inscrito o por el Colegio.",
          "Intentar acceder a áreas restringidas (portal del agremiado, panel administrativo), vulnerar la seguridad, sobrecargar el servicio o eludir sus límites de uso. La plataforma aplica **límites de peticiones** por conexión; su superación puede provocar bloqueos temporales.",
          "Introducir código malicioso o contenido ilícito.",
          "Reproducir el contenido del sitio de forma que sugiera respaldo del Colegio sin su autorización.",
        ],
      },
      {
        tipo: "parrafo",
        texto:
          "El Colegio podrá **bloquear accesos, limitar el uso y ejercer las acciones que correspondan** ante conductas contrarias a estos términos.",
      },
    ],
  },
  {
    numero: "8",
    titulo: "Solicitud de inscripción (quien aspira a ser agremiado)",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "El formulario público `/inscripcion` permite solicitar el ingreso al Colegio. Al usarlo, usted acepta que:",
      },
      {
        tipo: "lista",
        items: [
          "Debe suministrar **información veraz y completa**, y adjuntar los documentos exigidos (foto, comprobante, copias de cédula, título y RIF). Los campos marcados como obligatorios deben completarse.",
          "El formulario **verifica en tiempo real** si la cédula, el número de FPV o el correo ya constan en una solicitud pendiente; esto se hace únicamente para evitar duplicados.",
          "Su solicitud queda en estado **pendiente** y **solo el personal autorizado del Colegio puede verla**.",
          "**El envío de la solicitud no implica su aprobación.** El Colegio evalúa cada solicitud conforme a sus normas y puede rechazarla o pedir información adicional.",
          "**El Colegio le comunicará el resultado por el correo electrónico o el teléfono que usted registró en el formulario.** La plataforma no envía un aviso automático, porque la solicitud todavía no tiene una cuenta asociada; la comunicación la hace el personal del Colegio, dentro de un plazo razonable.",
          "El Colegio puede **pausar temporalmente la recepción** de solicitudes; en ese caso el formulario lo indicará y no permitirá el envío.",
          "Si su solicitud es rechazada, el Colegio podrá conservar los datos identificativos indispensables (cédula, número de FPV y correo) para evitar solicitudes duplicadas o fraudulentas.",
        ],
      },
    ],
  },
  {
    numero: "9",
    titulo: "Noticias, documentos y propiedad intelectual",
    bloques: [
      {
        tipo: "lista",
        items: [
          "Las **noticias públicas** son comunicaciones del Colegio para la ciudadanía. Algunas noticias están dirigidas **solo a agremiados** y no son visibles para el público.",
          "Los **textos legales** publicados (Código de Ética, Estatutos, Ley, Reglamento) se ofrecen como referencia de consulta; ante cualquier diferencia prevalecen los textos oficiales.",
          "Los contenidos, marcas, logotipos, diseño y textos originales del Colegio están protegidos por la normativa de propiedad intelectual. Puede compartir enlaces a las noticias y fichas. **No puede reproducir, modificar o explotar comercialmente** el contenido sin autorización escrita.",
          "Las fotografías y textos de cada ficha son responsabilidad de cada profesional.",
        ],
      },
    ],
  },
  {
    numero: "10",
    titulo: "Datos personales de quienes visitan la plataforma",
    bloques: [],
  },
  {
    numero: "10.1",
    titulo: "Sin iniciar sesión",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "Para navegar por el sitio público **no se le exige registrarse ni entregar datos personales**. Sin embargo, la plataforma registra información técnica de la visita con el fin único de conocer, en cifras agregadas, cómo se usa el sitio:",
      },
      {
        tipo: "lista",
        items: [
          "Las páginas que visita y la hora de la visita.",
          "El texto que escribe en el buscador del directorio y los filtros que aplica (área de trabajo, ubicación).",
          "Las fichas de profesionales que consulta.",
          "Una **huella criptográfica** de su dirección de conexión, en lugar de la dirección en sí. No se guarda la dirección IP.",
          "El **sitio web de origen** desde el que llegó a esta plataforma, sin la ruta ni los parámetros de esa dirección. No se guarda la dirección completa.",
        ],
      },
      {
        tipo: "aviso",
        tono: "info",
        texto:
          "Esta información se conserva **90 días** y se elimina automáticamente después de ese plazo. Las cifras se consultan **solo de forma agregada** (cuántas visitas, cuántas búsquedas, qué áreas se buscan más); el Colegio no construye perfiles de visitantes ni utiliza estos datos con fines publicitarios. **El sitio no instala cookies de terceros.**",
      },
      {
        tipo: "parrafo",
        texto:
          "Además, la plataforma aplica **límites de uso por conexión** para prevenir abusos, y mantiene una cookie técnica propia, sin datos personales y con una vigencia de 30 días, para no contar dos veces la misma visita.",
      },
    ],
  },
  {
    numero: "10.2",
    titulo: "Si envía una solicitud de inscripción",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "Los datos y documentos que envíe (datos personales, formación, contacto, ubicación, cédula, título, RIF, fotografía, comprobante) se usarán **para evaluar su solicitud** y, si es aprobada, para gestionar su condición de agremiado. Son de acceso restringido al personal autorizado del Colegio.",
      },
    ],
  },
  {
    numero: "10.3",
    titulo: "Seguridad",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "El Colegio aplica medidas técnicas razonables para proteger la información: control de acceso, sesiones con vencimiento, contraseñas almacenadas de forma protegida, limpieza de contenidos y registro de las acciones administrativas que se realizan sobre los datos de los agremiados. Ningún sistema es infalible; el Colegio no puede garantizar seguridad absoluta.",
      },
    ],
  },
  {
    numero: "10.4",
    titulo: "Sus derechos",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "Conforme a la Constitución de la República Bolivariana de Venezuela (en particular el derecho a conocer y rectificar los datos que sobre usted consten en registros, y a la protección de la honra, la vida privada y la propia imagen) y demás normas aplicables, usted puede solicitar al Colegio **conocer, rectificar, actualizar o retirar** los datos personales que aparezcan sobre usted en la plataforma, escribiendo a " +
          CORREO_PRIVACIDAD +
          ". El Colegio atenderá su solicitud por ese mismo correo o por el teléfono que usted haya facilitado, dentro de un plazo razonable, salvo que exista una obligación legal o gremial de conservar los datos.",
      },
    ],
  },
  {
    numero: "11",
    titulo: "Enlaces y servicios de terceros",
    bloques: [
      {
        tipo: "parrafo",
        texto: `La plataforma puede enlazar a sitios de terceros, incluidas las redes sociales que cada profesional decide publicar en su ficha y el sitio de la [FPV](${FPV_URL}). El Colegio **no controla ni responde** por su contenido, disponibilidad o políticas de privacidad.`,
      },
    ],
  },
  {
    numero: "12",
    titulo: "Disponibilidad y limitación de responsabilidad",
    bloques: [
      {
        tipo: "lista",
        items: [
          "La plataforma se ofrece «tal como está» y puede tener interrupciones por mantenimiento, fallas técnicas o causas ajenas al Colegio.",
          "El Colegio **no responde** por decisiones que usted tome con base en la información publicada, ni por los actos, omisiones, servicios o comunicaciones de los profesionales listados, ni por el uso que terceros hagan de los datos publicados por los propios profesionales.",
          "Nada de lo anterior limita las responsabilidades que la ley no permita excluir.",
        ],
      },
    ],
  },
  {
    numero: "13",
    titulo: "Modificaciones",
    bloques: [
      {
        tipo: "parrafo",
        texto:
          "El Colegio puede modificar estos términos. La versión vigente estará publicada en esta plataforma con su fecha de actualización. El uso continuado del sitio tras una modificación implica su aceptación.",
      },
    ],
  },
  {
    numero: "14",
    titulo: "Ley aplicable y jurisdicción",
    bloques: [
      {
        tipo: "parrafo",
        texto: `Estos términos se rigen por las leyes de la República Bolivariana de Venezuela. Para cualquier controversia, las partes se someten a los tribunales competentes de ${COLEGIO.ciudad}.`,
      },
    ],
  },
  {
    numero: "15",
    titulo: "Contacto",
    bloques: [
      {
        tipo: "parrafo",
        texto: `${COLEGIO.nombre} · ${COLEGIO.direccion} · ${COLEGIO.correo}`,
      },
    ],
  },
];

const parteIPublica: TerminosParte = {
  clave: "publica",
  // Rótulo neutro. Antes decía "Parte I", y en pantalla (badge del héroe, tarjeta
  // del índice) eso ya anunciaba que existía una Parte II.
  etiqueta: "Condiciones de uso del sitio",
  // Sin "y de la plataforma": un título que reparte el alcance entre el sitio y la
  // plataforma dice que la otra mitad del alcance está en otro documento.
  titulo: "Condiciones de Uso",
  // Sin "sin importar si tiene o no cuenta en el portal del agremiado":
  // mencionar cuentas dentro del texto público dice que hay otro documento para
  // quienes sí tienen cuenta.
  resumen: `Regula el acceso y uso del sitio del ${COLEGIO.nombre} por parte de cualquier visitante.`,
  secciones: _SECCIONES,
};

export default parteIPublica;
