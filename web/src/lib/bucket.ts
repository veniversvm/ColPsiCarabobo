const BUCKET_URL = import.meta.env.VITE_BUCKET_URL || "";
const SITE_URL = import.meta.env.VITE_SITE_URL || "";

/**
 * Construye la URL completa para un objeto en el bucket de S3/MinIO.
 * Retorna vacío si no hay key o si BUCKET_URL no está configurada.
 * Si la key ya es una URL absoluta (el API puede devolverla completa),
 * la retorna sin modificar para evitar prefijos duplicados.
 */
export function bucketUrl(key: string | null | undefined): string {
  if (!key || !BUCKET_URL) return "";
  if (/^https?:\/\//i.test(key)) return key;
  return `${BUCKET_URL}/${key}`;
}

/**
 * Construye una URL absoluta del sitio combinando SITE_URL base con el path dado.
 */
export function siteUrl(path: string): string {
  return `${SITE_URL}${path}`;
}
