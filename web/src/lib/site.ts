// web/src/lib/site.ts
//
// URL base del sitio — único lugar donde se decide.
//
// POR QUÉ ESTE ARCHIVO EXISTE: antes, nueve archivos declaraban su propia
// constante `SITE_URL` y había CUATRO fallbacks distintos repartidos:
//
//   http://localhost:3000            (6 archivos)
//   https://colpsi-carabobo.org      (noticias/[slug])
//   https://colpsicarabobo.org       (explorar, inscripcion)
//   http://localhost:28080/api/v1    (psi/perfil.tsx — un error: una URL de API
//                                     usada como `canonical`, que producía
//                                     "http://localhost:28080/api/v1/directorio/…")
//
// Con `VITE_SITE_URL` sin definir (es el caso normal: la clave no está en
// `web/.env`, solo en `.env.example`), cada página quemaba su propio dominio en
// los `canonical` y en las etiquetas de Open Graph. Para el Colegio eso
// significa que los buscadores podían ver dos o tres versiones distintas de la
// misma página, y que el `canonical` de una ficha de agremiado apuntaba a la API.
//
// Uso:  import { SITE_URL } from "~/lib/site";
//
// NO confundir con `lib/bucket.ts`: ese usa SITE_URL con fallback "" a
// propósito, porque solo necesita construir URLs absolutas cuando la variable
// está configurada.

/**
 * URL base del sitio, sin barra final.
 *
 * Se lee de `VITE_SITE_URL` y, si no está, cae al dominio de producción. El
 * fallback importa: con un `localhost` los canonical de producción apuntarían
 * al equipo de desarrollo.
 *
 * Se normaliza quitando la barra final porque el dominio se suele escribir con
 * ella ("https://…com/") y `${SITE_URL}/directorio` produciría "//directorio".
 */
export const SITE_URL = (
  import.meta.env.VITE_SITE_URL || "https://colegio-psicologos-carabobo.com"
).replace(/\/+$/, "");

/**
 * Construye una URL absoluta del sitio a partir de un path con barra inicial.
 * `absoluteUrl("/terminos")` → `https://…com/terminos`
 */
export function absoluteUrl(path: string): string {
  return `${SITE_URL}${path.startsWith("/") ? path : `/${path}`}`;
}
