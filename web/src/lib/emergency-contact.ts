// web/src/lib/emergency-contact.ts
//
// Catálogo de parentescos y validación en cliente de la Persona de Contacto
// para Emergencias. Es la MISMA regla que aplica la API Go, replicada aquí para
// dar feedback inmediato sin gastar un round-trip:
//
//   nombre y parentesco siempre, y al menos UNO de teléfono o correo.
//
// El catálogo vive solo en el frontend: la API no lo conoce, solo exige que el
// parentesco no venga vacío (el texto libre de "Otro" se guarda tal cual).

import type { EmergencyContact } from "~/types/psi";

/** Máximo de contactos por agremiado (debe coincidir con `MaxEmergencyContacts` en Go). */
export const MAX_EMERGENCY_CONTACTS = 3;

/** Valor centinela del `<select>` que revela el campo de texto libre. */
export const OTHER_RELATIONSHIP = "__otro__";

/**
 * Catálogo de parentescos. El `value` es lo que se persiste; el `label` es lo que
 * ve el agremiado. Se guardan en minúsculas para no fragmentar el dato (el mismo
 * parentesco escrito de dos formas no debe producir dos registros distintos).
 */
export const RELATIONSHIPS: { value: string; label: string }[] = [
  { value: "madre", label: "Madre" },
  { value: "padre", label: "Padre" },
  { value: "conyuge", label: "Cónyuge / compañero(a)" },
  { value: "hijo", label: "Hijo(a)" },
  { value: "hermano", label: "Hermano(a)" },
  { value: "tio", label: "Tío(a)" },
  { value: "primo", label: "Primo(a)" },
  { value: "amigo", label: "Amigo(a) / conocida(o)" },
  { value: "colega", label: "Colega" },
];

/**
 * Resuelve el valor del `<select>` al texto que se persiste. Si el usuario eligió
 * "Otro" (o el registro ya tenía un parentesco fuera del catálogo), se devuelve
 * el texto libre tal cual.
 */
export function resolveRelationship(selected: string, custom: string): string {
  if (selected === OTHER_RELATIONSHIP) return custom.trim();
  // Un registro existente puede tener un parentesco en texto libre: se conserva.
  if (!RELATIONSHIPS.some((r) => r.value === selected)) return selected.trim();
  return selected;
}

/**
 * Determina qué opción del `<select>` corresponde a un parentesco guardado:
 * el del catálogo, o "Otro" si es texto libre.
 */
export function relationshipOption(relationship: string): string {
  return RELATIONSHIPS.some((r) => r.value === relationship)
    ? relationship
    : OTHER_RELATIONSHIP;
}

export type EmergencyContactDraft = {
  name: string;
  relationship: string;
  phone: string;
  email: string;
};

/** Cuerpo que aceptan `POST/PATCH .../emergency` (ver `psi_user_emergency.go`). */
export type EmergencyContactPayload = {
  name: string;
  relationship: string;
  phone: string;
  email: string;
};

/** Error de validación con el campo culpable, para Pintar el mensaje bajo el input. */
export type EmergencyContactError = { field: string; message: string };

const EMAIL_RE = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

/**
 * Valida el borrador contra la regla de negocio. Devuelve `null` si es válido.
 * Mismo criterio que `normalizeEmergencyContact` en Go: teléfono con 7..20
 * dígitos (los separadores se ignoran) y correo con formato RFC-ish.
 */
export function validateEmergencyContact(
  draft: EmergencyContactDraft,
): EmergencyContactError | null {
  const name = draft.name.trim();
  const relationship = draft.relationship.trim();
  const phoneDigits = draft.phone.replace(/\D/g, "");
  const email = draft.email.trim();

  if (name.length < 2) {
    return { field: "name", message: "Escribe el nombre de la persona de contacto." };
  }
  if (name.length > 255) {
    return { field: "name", message: "El nombre es demasiado largo." };
  }
  if (!relationship) {
    return { field: "relationship", message: "Indica el parentesco o vínculo." };
  }
  if (relationship.length > 100) {
    return { field: "relationship", message: "El parentesco es demasiado largo." };
  }
  if (phoneDigits.length > 20) {
    return { field: "phone", message: "El teléfono es demasiado largo." };
  }
  if (email && !EMAIL_RE.test(email)) {
    return { field: "email", message: "El correo no tiene un formato válido." };
  }
  if (!phoneDigits && !email) {
    return {
      field: "phone",
      message: "Escribe al menos un teléfono o un correo para poder avisarle.",
    };
  }
  return null;
}

/** Prepara el payload que espera la API, ya limpio. */
export function toEmergencyPayload(
  draft: EmergencyContactDraft,
): EmergencyContactPayload {
  const phoneDigits = draft.phone.replace(/\D/g, "");
  return {
    name: draft.name.trim(),
    relationship: draft.relationship.trim(),
    phone: phoneDigits ? (draft.phone.startsWith("+") ? `+${phoneDigits}` : phoneDigits) : "",
    email: draft.email.trim().toLowerCase(),
  };
}

/** Listado normalizado: la API omite la clave cuando no hay contactos (`omitempty`). */
export function emergencyContactsOf(source: { emergency_contacts?: EmergencyContact[] } | undefined) {
  return source?.emergency_contacts ?? [];
}
