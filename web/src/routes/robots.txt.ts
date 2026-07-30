import { siteUrl } from "~/lib/bucket";

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
