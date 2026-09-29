// web/src/routes/terminos/index.tsx → /terminos
// Documento público de Términos y Condiciones.
//
// ESTE ARCHIVO ES EL DOCUMENTO, NO UN ÍNDICE. No hay lista de partes, ni
// numerales, ni nada que sugiera que exista un segundo documento: el visitante
// sin sesión ve un solo texto. El documento del agremiado cuelga del portal
// (`/psi/terminos`), no se anuncia desde aquí, y el enlace cruzado entre ambos
// solo se pinta con sesión (ver `TerminosLayout`, `mostrarCruce`).
//
// NOTA DE RUTING — este archivo vive DENTRO de la carpeta `terminos/`, no como
// `routes/terminos.tsx`. Motivo: si existiera `terminos.tsx` junto a la carpeta,
// SolidStart lo tomaría como layout del directorio (igual que `psi.tsx` envuelve
// `psi/`) y `/terminos` quedaría sin página mientras `/terminos/publica`
// renderizara el archivo como envoltura. Con `index.tsx` dentro, las rutas
// cuelgan de la carpeta y ninguna se solapa.

import TerminosLayout from "~/components/terminos/TerminosLayout";
import { PENDIENTE } from "~/lib/terminos/parte-i-publica";

export default function TerminosPublicos() {
  return <TerminosLayout clave="publica" pendiente={PENDIENTE} />;
}
