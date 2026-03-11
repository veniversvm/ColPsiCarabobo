// web/src/lib/sanitizer.ts

/**
 * Limpia números de teléfono.
 * Permite '+' solo al inicio y '-' en cualquier posición.
 */
export function sanitizePhone(val: string): string {
  if (!val) return "";
  // Mantener el + inicial si existe
  const hasPlus = val.startsWith("+");
  // Eliminar todo lo que no sea número o guion
  let cleaned = val.replace(/[^0-9-]/g, "");
  return (hasPlus ? "+" : "") + cleaned;
}

/**
 * Limpia texto general.
 * Elimina caracteres especiales peligrosos para SQL o HTML.
 * Solo permite Alfanuméricos, espacios, acentos y eñes.
 */
export function sanitizeText(val: string): string {
  if (!val) return "";
  // Regex: Permite letras (con acentos), números, espacios, puntos, comas y guiones.
  // Elimina: < > ; ' " -- (comentarios SQL) { } [ ]
  return val.replace(/[^a-zA-Z0-9áéíóúÁÉÍÓÚñÑ\s,.¿?¡!-]/g, "").trim();
}

/**
 * Normaliza y valida un email. 
 * Retorna el email limpio o null si no es válido.
 */
export function sanitizeEmail(email: string): string | null {
  if (!email) return null;

  // 1. Normalización básica
  const cleanEmail = email.trim().toLowerCase();

  // 2. Validación mediante RFC 5322 (Regex estándar balanceada)
  const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

  if (!emailRegex.test(cleanEmail)) {
    return null;
  }

  // 3. Sanitización adicional (opcional pero recomendada)
  // Evitamos inyecciones básicas eliminando caracteres peligrosos 
  // que el regex anterior podría haber dejado pasar en casos raros.
  return cleanEmail.replace(/[<>'"();]/g, "");
}

/**
 * Procesa un objeto completo (Request DTO) y limpia sus campos según el tipo
 */
export function sanitizeProfileRequest(data: any): any {
  const cleanData = { ...data };

  // Campos de teléfono
  const phoneFields = [
    "public_phone", "phone_carabobo", "cel_phone_carabobo", 
    "phone_outside_carabobo", "cel_phone_outside_carabobo"
  ];
  
  // Campos de texto alfanumérico
  const textFields = [
    "municipality_carabobo", "state_outside", 
    "municipality_outside_carabobo", "primary_specialty", 
    "secondary_specialty", "mini_bio", "service_address"
  ];

  Object.keys(cleanData).forEach(key => {
    if (typeof cleanData[key] === "string") {
      if (phoneFields.includes(key)) {
        cleanData[key] = sanitizePhone(cleanData[key]);
      } else if (textFields.includes(key)) {
        cleanData[key] = sanitizeText(cleanData[key]);
      } else if (key.includes("email")) {
        cleanData[key] = sanitizeEmail(cleanData[key]);
      }
    }
  });

  return cleanData;
}


// web/src/lib/sanitizer.ts

/**
 * Valida si una contraseña cumple con los estándares de seguridad de la API en Go.
 * Reglas:
 * 1. Mínimo 8 caracteres.
 * 2. Al menos una letra mayúscula.
 * 3. Al menos una letra minúscula.
 * 4. Al menos un número.
 * 5. Al menos un carácter especial (puntuación o símbolo).
 * 6. NO contiene espacios en blanco.
 */
export function isStrongPassword(password: string): boolean {
  if (!password || password.length < 8) return false;
  
  // Regla: Sin espacios en blanco
  if (/\s/.test(password)) return false;

  // Reglas de composición usando Regex
  const hasUpper = /[A-Z]/.test(password);
  const hasLower = /[a-z]/.test(password);
  const hasNumber = /[0-9]/.test(password);
  
  // Caracteres especiales: Cualquier cosa que no sea alfanumérica ni espacio
  const hasSpecial = /[^a-zA-Z0-9\s]/.test(password);

  return hasUpper && hasLower && hasNumber && hasSpecial;
}

/**
 * Trunca un texto al máximo permitido para evitar errores en la base de datos.
 */
export function enforceMaxLength(val: string, max: number): string {
  if (!val) return "";
  return val.length > max ? val.substring(0, max) : val;
}
