const KNOWN_MESSAGES: Record<string, string> = {
  "OFFLINE_SERVICE": "El servicio no está disponible en este momento. Intente más tarde.",
  "Credenciales inválidas.": "Usuario o contraseña incorrectos.",
  "Error al guardar.": "Ocurrió un error al guardar. Verifique los datos e intente de nuevo.",
  "Error al publicar.": "Ocurrió un error al publicar. Intente de nuevo.",
};

/**
 * Traduce errores internos (err.message) a mensajes amigables para el usuario final.
 * Si el mensaje no está en el mapa, retorna un mensaje genérico.
 */
export function getUserFacingError(error: unknown): string {
  const msg = (error as any)?.message || String(error);
  return KNOWN_MESSAGES[msg] || "Ocurrió un error inesperado. Si persiste, contacte al administrador.";
}
