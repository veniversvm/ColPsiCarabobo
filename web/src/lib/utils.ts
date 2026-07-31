// web/src/lib/utils.ts
//
// Utilidades de dominio compartidas (slugs, orden de postgrados).
// No colocar aquí utilidades genéricas de UI; crear un archivo
// dedicado en `src/lib/` por responsabilidad.

/**
 * Normaliza un texto para URL (elimina acentos y caracteres especiales)
 */
function normalizeForUrl(text: string): string {
  return text
    .toLowerCase()
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '') // Eliminar acentos
    .replace(/[^a-z0-9]/g, '-') // Reemplazar caracteres especiales con guiones
    .replace(/-+/g, '-') // Eliminar guiones múltiples
    .replace(/^-|-$/g, ''); // Eliminar guiones al inicio y final
}

/**
 * Crea un slug amigable para el perfil del psicólogo
 * Formato: nombre-segundoNombre-apellido-segundoApellido-fpv-1234
 */
export function createProfileSlug(profile: {
  first_name: string;
  second_name?: string | null;
  last_name: string;
  second_last_name?: string | null;
  fpv: number;
}): string {
  // Array con todos los nombres no vacíos
  const nameParts = [
    profile.first_name,
    profile.second_name,
    profile.last_name,
    profile.second_last_name,
  ].filter((part): part is string => Boolean(part && part.trim() !== ''));
  
  // Unir todas las partes con guiones
  const nameSlug = nameParts.map(part => normalizeForUrl(part)).join('-');
  
  return `${nameSlug}-fpv-${profile.fpv}`;
}

/**
 * Extrae el FPV del slug
 * El formato es: ...-fpv-1234
 */
export function extractFpvFromSlug(slug: string): number | null {
  const match = slug.match(/-fpv-(\d+)$/);
  if (!match) return null;
  
  const fpv = parseInt(match[1], 10);
  return isNaN(fpv) ? null : fpv;
}

/**
 * Ordena postgrados por año (del más antiguo al más reciente)
 */
export function sortPostGradesByYear(postGrades: any[] | undefined): any[] {
  if (!postGrades) return [];
  return [...postGrades].sort((a, b) => {
    if (!a.year) return 1;
    if (!b.year) return -1;
    return parseInt(a.year, 10) - parseInt(b.year, 10);
  });
}