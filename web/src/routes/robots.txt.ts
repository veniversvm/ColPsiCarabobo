import { siteUrl } from "~/lib/bucket";

/**
 * Ruta que genera robots.txt dinámico para que la URL del Sitemap
 * use SITE_URL en vez de localhost.
 */
export async function GET() {
  const body = `Sitemap: ${siteUrl("/sitemap.xml")}

User-agent: *
Allow: /
Disallow: /admin/
Disallow: /psi/
`;
  return new Response(body, {
    headers: {
      "Content-Type": "text/plain",
      "Cache-Control": "public, max-age=86400",
    },
  });
}
