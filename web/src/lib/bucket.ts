const BUCKET_URL = import.meta.env.VITE_BUCKET_URL || "";
const SITE_URL = import.meta.env.VITE_SITE_URL || "";

/**
 * Construye la URL completa para un objeto en el bucket de S3/MinIO.
 * Retorna vacío si no hay key o si BUCKET_URL no está configurada.
 */
export function bucketUrl(key: string | null | undefined): string {
  if (!key || !BUCKET_URL) return "";
  return `${BUCKET_URL}/${key}`;
}

/**
 * Construye una URL absoluta del sitio combinando SITE_URL base con el path dado.
 */
export function siteUrl(path: string): string {
  return `${SITE_URL}${path}`;
}
