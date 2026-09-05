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
  // Términos del dominio de tickets (api/internal/domain/errors.go)
  "no puedes publicar más de 3 mensajes seguidos en la conversación":
    "Espere la respuesta del colegio antes de enviar otro mensaje (máximo 3 seguidos).",
  "el ticket está cerrado y no admite más comentarios":
    "Esta solicitud está cerrada y ya no admite comentarios.",
  "el mensaje excede el límite de caracteres":
    "El comentario supera el máximo de caracteres permitidos.",
  "el mensaje no puede estar vacío":
    "Escriba un comentario antes de enviar.",
  "límite de tickets abiertos alcanzado para este motivo":
    "Ya alcanzó el límite de solicitudes abiertas para este motivo.",
  "debes indicar un motivo de cierre":
    "Indique el motivo por el que cierra la solicitud.",
  "no puedes acceder a un ticket que no te pertenece":
    "No puede acceder a esta solicitud.",
  "ticket no encontrado": "Solicitud no encontrada.",
  "motivo de ticket no encontrado": "El motivo seleccionado ya no existe.",
  "estado de ticket no encontrado": "El estado seleccionado ya no existe.",
  "el estado no pertenece al motivo del ticket":
    "El estado seleccionado no corresponde al motivo de la solicitud.",
  "el motivo tiene tickets asociados y no se puede eliminar":
    "El motivo tiene solicitudes asociadas y no se puede eliminar.",
  "el estado está en uso por algún ticket y no se puede eliminar":
    "El estado está en uso por alguna solicitud y no se puede eliminar.",
  "el límite de tickets por psicólogo para este motivo debe ser al menos 1":
    "El límite de solicitudes por psicólogo debe ser al menos 1.",
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
