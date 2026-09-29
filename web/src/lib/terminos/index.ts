// Términos y Condiciones — catálogo central del documento.
//
// Sigue el patrón de `lib/documentos/`: el texto es estático, vive en módulos
// tipados y NO tiene CRUD en el panel admin. Es deliberado. Un documento legal
// versionado y editable por el admin desde una pantalla invites a publicarlo a
// medio escribir; además, un cambio de texto tiene que pasar por revisión legal,
// no por un formulario.
//
// La versión vigente NO se declara aquí como fuente de verdad: la declara el
// backend (`domain.TermsVersion`) y el frontend la recibe por
// `GET /psi/me/terms`. Ver `TERMINOS_VERSION` en `colegio.ts`.

import parteIPublica from "./parte-i-publica";
import parteIIAgremiado from "./parte-ii-agremiado";
import { TERMINOS_VERSION } from "./colegio";

export { TERMINOS_VERSION } from "./colegio";
export * from "./colegio";

/** Identificador de cada parte del documento. */
export type TerminosParteClave = "publica" | "agremiado";

/**
 * Un bloque de contenido.
 *
 * El conjunto es corto a propósito: un documento legal se lee, no se maqueta.
 * Los cuatro tipos cubren párrafos, listas numeradas, avisos destacados y
 * cuadros comparativos, que es todo lo que hace falta.
 *
 * `texto` admite un subconjunto deliberado de Markdown en línea:
 *   **negrita**, *cursiva*, `código` y [enlace](https://ejemplo.ve).
 * Se renderiza con `renderInline()` de `TerminosLayout.tsx`, que NO usa
 * innerHTML: construye nodos, así que no hay superficie de XSS aunque alguien
 * edite el texto mañana.
 */
export type TerminosBloque =
  | { tipo: "parrafo"; texto: string }
  | { tipo: "lista"; items: string[]; ordenada?: boolean }
  | { tipo: "aviso"; texto: string; tono?: "info" | "alerta" }
  | { tipo: "tabla"; cabecera: string[]; filas: string[][] };

/** Una sección numerada del documento. */
export interface TerminosSeccion {
  /** Numeración jerárquica tal como se imprime: "1", "4.3", "10.4". */
  numero: string;
  titulo: string;
  bloques: TerminosBloque[];
}

/**
 * Una sección de primer nivel con sus subsecciones colgando.
 *
 * El agrupamiento NO es cosmético. El articulado se lee por niveles: si "5.1
 * Origen de la información" se dibuja al mismo tamaño y en la misma tarjeta que
 * "5. El directorio", el lector pierde la jerarquía que en un documento legal
 * es justamente lo que da sentido al texto. El nivel se deduce del propio
 * `numero` (un punto = subsección), así que no hay que marcar nada a mano.
 */
export interface TerminosGrupo {
  seccion: TerminosSeccion;
  hijos: TerminosSeccion[];
}

/**
 * Agrupa las secciones por nivel a partir de su numeración.
 *
 * Una sección cuyo número no lleva punto es de primer nivel; el resto cuelga
 * de la sección cuyo número es el prefijo anterior al último punto
 * ("10.3" → "10"). Si una subsección llega sin su padre —porque alguien pegó
 * mal el texto— se promueve a primer nivel en vez de desaparecer: es
 * preferible que se lea de más a que se pierda un artículo.
 */
export function agruparSecciones(secciones: TerminosSeccion[]): TerminosGrupo[] {
  const grupos: TerminosGrupo[] = [];
  const porNumero = new Map<string, TerminosGrupo>();

  for (const s of secciones) {
    if (!s.numero.includes(".")) {
      const g: TerminosGrupo = { seccion: s, hijos: [] };
      grupos.push(g);
      porNumero.set(s.numero, g);
      continue;
    }
    const padre = s.numero.slice(0, s.numero.lastIndexOf("."));
    const destino = porNumero.get(padre);
    if (destino) destino.hijos.push(s);
    else grupos.push({ seccion: s, hijos: [] });
  }

  return grupos;
}

export interface TerminosParte {
  clave: TerminosParteClave;
  /**
   * Rótulo corto del documento, para el enlace cruzado entre ambos. NO es una
   * numeración: solo el ve un agremiado con sesión, y decir "Parte II" en
   * pantalla confirmaría que hay dos partes.
   */
  etiqueta: string;
  titulo: string;
  /** Una o dos frases de qué regula. Se usa como descripción SEO. */
  resumen: string;
  secciones: TerminosSeccion[];
}

/** Las dos partes, en orden. */
export const TERMINOS_PARTES: TerminosParte[] = [parteIPublica, parteIIAgremiado];

/** Devuelve la parte por clave; `undefined` si la clave no existe. */
export function getParte(clave?: string): TerminosParte | undefined {
  if (!clave) return undefined;
  return TERMINOS_PARTES.find((p) => p.clave === clave);
}

/**
 * Versión vigente declarada por el frontend.
 *
 * Solo se usa como valor de arranque y para pintar la fecha en la cabecera.
 * La que decide si hay que aceptar es la que llega de la API; si difieren, la
 * API gana (ver `AceptarTerminosModal.tsx`).
 */
export const VERSION_TERMINOS = TERMINOS_VERSION;
