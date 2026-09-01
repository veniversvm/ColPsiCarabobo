// web/src/lib/errors.ts
//
// Traducción de errores internos a mensajes amigables.
// Evita filtrar detalles internos (stack traces, códigos de la API)
// al usuario final. Agregar entradas nuevas a KNOWN_MESSAGES cuando
// la API introduzca mensajes de error adicionales.

const KNOWN_MESSAGES: Record<string, string> = {
  "OFFLINE_SERVICE": "El servicio no está disponible en este momento. Intente más tarde.",
  "credenciales inválidas": "Usuario o contraseña incorrectos.",
  "error al guardar": "Ocurrió un error al guardar. Verifique los datos e intente de nuevo.",
  "error al publicar": "Ocurrió un error al publicar. Intente de nuevo.",
};

/**
 * Normaliza un mensaje de error a una forma canónica para facilitar el match
 * contra el mapa de mensajes conocidos (ignora mayúsculas, espacios y el punto final).
 */
function normalize(msg: string): string {
  return msg.trim().toLowerCase().replace(/[.\s]+$/, "");
}

/**
 * Traduce errores internos (err.message) a mensajes amigables para el usuario final.
 * Si el mensaje no está en el mapa, retorna un mensaje genérico.
 */
export function getUserFacingError(error: unknown): string {
  const msg = normalize((error as any)?.message || String(error));
  return KNOWN_MESSAGES[msg] || "Ocurrió un error inesperado. Si persiste, contacte al administrador.";
}
