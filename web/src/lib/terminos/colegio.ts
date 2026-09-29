// Datos institucionales del Colegio de Psicólogos del Estado Carabobo.
//
// CONVENCIÓN DEL PROYECTO: todo dato que aparece en los documentos legales
// (Términos y Condiciones) vive AQUÍ y en ningún otro lado. Razón práctica: el
// teléfono, la dirección y el correo cambian; si el texto legal los repite
// en 1.500 líneas, corregirlos es una búsqueda difícil y es fácil dejar una
// copia desactualizada. Para actualizar uno se edita esta constante y se
// reconstruye.
//
// OJO — leer antes de editar:
//   · No hay teléfono. El Colegio no publica ninguno en ningún canal del sitio,
//     así que el dato no existe y los documentos NO lo inventan (ver `telefono`).
//     Si el Colegio publica uno, se añade aquí y se usa en los documentos.
//   · `correo` es una casilla de Gmail (`admon.`). Funciona, pero no es un dominio
//     propio del Colegio: para un órgano de derecho público es una debilidad que
//     conviene corregir cuando se tenga el dominio. Es un solo cambio de línea.
//   · Los documentos legales se corrigen contra la LEY, no contra este archivo:
//     aquí solo cambian los datos, nunca las obligaciones.

/** Nombre legal completo del Colegio. */
export const COLEGIO_NOMBRE = "Colegio de Psicólogos del Estado Carabobo";

/** Registro de Información Fiscal. */
export const COLEGIO_RIF = "J-508172418";

/** Domicilio. Sin abreviaturas ("Edo."): en un documento legal se escriben completas. */
export const COLEGIO_DIRECCION =
  "Edificio Profesional y Comercial, Oficentro 108, piso 4, oficina D, Avenida 101 Díaz Moreno, Valencia, estado Carabobo, Venezuela";

/** Ciudad y jurisdicción. */
export const COLEGIO_CIUDAD = "Valencia, estado Carabobo, Venezuela";

/**
 * Correo institucional. Sirve también como canal de privacidad y de reportes
 * (ver `CORREO_PRIVACIDAD` y `CORREO_REPORTES`): los documentos lo dicen de
 * forma expresa para que nadie tenga que adivinar a dónde escribir.
 */
export const COLEGIO_CORREO = "admon.colpsicarabobo@gmail.com";

/** Alias explícito del canal de privacidad — el mismo buzón, declarado aparte. */
export const CORREO_PRIVACIDAD = COLEGIO_CORREO;

/** Alias explícito del canal de reportes — el mismo buzón, declarado aparte. */
export const CORREO_REPORTES = COLEGIO_CORREO;

/**
 * Teléfono: NO EXISTE.
 *
 * Se declara como `null` en vez de como cadena vacía para que un documento que
 * lo interpole muestre una omisión consciente y no un hueco invisible. Si algún
 * día el Colegio publica un número, se cambia a `null` → `"+58 241 000 0000"` y
 * los documentos que hoy lo omiten lo empezarán a mostrar.
 */
export const COLEGIO_TELEFONO: string | null = null;

/** Sitio web del Colegio. */
export const COLEGIO_URL = "https://colegio-psicologos-carabobo.com";

/** Federación de Psicólogos de Venezuela. */
export const FPV_URL = "https://fpv.org.ve/";

/**
 * Versión de los Términos y Condiciones.
 *
 * ADVERTENCIA — esta constante es una COPIA de la que declara el backend
 * (`domain.TermsVersion` en la API). La fuente de verdad es el backend: el
 * frontend la recibe por `GET /psi/me/terms` y usa ESE valor para decidir si el
 * modal aparece. La copia de aquí solo se usa como valor de arranque si el
 * endpoint falla, y para pintar la fecha en la cabecera del documento.
 *
 * Al publicar una versión nueva hay que cambiar las DOS (backend y este
 * archivo) en el mismo despliegue. Si quedan desalineadas, el síntoma es
 * visible: el texto muestra una fecha y el modal sigue pidiendo aceptar otra.
 */
export const TERMINOS_VERSION = "2026-09-29";

/** Bloque institucional listo para interpolar en cualquier documento. */
export const COLEGIO = {
  nombre: COLEGIO_NOMBRE,
  rif: COLEGIO_RIF,
  direccion: COLEGIO_DIRECCION,
  ciudad: COLEGIO_CIUDAD,
  correo: COLEGIO_CORREO,
  correoPrivacidad: CORREO_PRIVACIDAD,
  correoReportes: CORREO_REPORTES,
  telefono: COLEGIO_TELEFONO,
  url: COLEGIO_URL,
  fpvUrl: FPV_URL,
} as const;
