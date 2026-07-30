const BUCKET_URL = import.meta.env.VITE_BUCKET_URL || "";
const SITE_URL = import.meta.env.VITE_SITE_URL || "";

export function bucketUrl(key: string | null | undefined): string {
  if (!key || !BUCKET_URL) return "";
  return `${BUCKET_URL}/${key}`;
}

export function siteUrl(path: string): string {
  return `${SITE_URL}${path}`;
}
