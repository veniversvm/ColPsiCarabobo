import { apiGet } from "~/lib/api";

// 1. Configuración de URL base (Asegúrate de tener VITE_SITE_URL en tu .env de producción)
const SITE_URL = import.meta.env.VITE_SITE_URL || "http://localhost:3000";

// ── HELPERS PARA SLUGS (Deben coincidir con tus rutas [slug].tsx) ──

const cleanString = (text: string) => {
  return text
    .toLowerCase()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "") // Quitar acentos
    .replace(/[^a-z0-9]/g, "-")      // Cambiar todo lo no alfanumérico por guion
    .replace(/-+/g, "-")             // Quitar guiones dobles
    .replace(/^-|-$/g, "");          // Quitar guiones al inicio o final
};

const toNewsSlug = (title: string, id: string) => {
  const slugTitle = cleanString(title).slice(0, 55);
  const firstSegment = id.split("-")[0]; // Tomamos el primer bloque del UUID
  return `${slugTitle}-${firstSegment}`;
};

const toPsiSlug = (p: any) => {
  const firstName = cleanString(p.first_name || "");
  const lastName = cleanString(p.last_name || "");
  return `${firstName}-${lastName}-fpv-${p.fpv}`;
};

// ── FUNCIÓN PRINCIPAL GET ──

export async function GET() {
  let psis: any[] = [];
  let posts: any[] = [];

  try {
    // Ejecutamos ambas peticiones al mismo tiempo para no perder tiempo
    const [psiData, postData] = await Promise.all([
      apiGet<any[]>("/psi/public/sitemap-data"),
      apiGet<any[]>("/posts/public/sitemap-posts")
    ]);
    psis = psiData || [];
    posts = postData || [];
  } catch (err) {
    console.error("[Sitemap Generation Error]", err);
    // Si falla la API, devolvemos un array vacío para que al menos las rutas estáticas se indexen
  }

  // 2. Construcción del XML
  const sitemap = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <!-- Páginas Estáticas Principales -->
  <url>
    <loc>${SITE_URL}/</loc>
    <priority>1.0</priority>
    <changefreq>weekly</changefreq>
  </url>
  <url>
    <loc>${SITE_URL}/directorio</loc>
    <priority>0.9</priority>
    <changefreq>weekly</changefreq>
  </url>
  <url>
    <loc>${SITE_URL}/noticias</loc>
    <priority>0.9</priority>
    <changefreq>daily</changefreq>
  </url>
  <url>
    <loc>${SITE_URL}/inscripcion</loc>
    <priority>0.8</priority>
    <changefreq>monthly</changefreq>
  </url>
  <url>
    <loc>${SITE_URL}/nosotros</loc>
    <priority>0.8</priority>
    <changefreq>monthly</changefreq>
  </url>
  <url>
    <loc>${SITE_URL}/documentos</loc>
    <priority>0.7</priority>
    <changefreq>monthly</changefreq>
  </url>
  <url>
    <loc>${SITE_URL}/documentos/estatutos-fpv</loc>
    <priority>0.6</priority>
    <changefreq>monthly</changefreq>
  </url>
  <url>
    <loc>${SITE_URL}/documentos/ley-ejercicio-psicologia</loc>
    <priority>0.6</priority>
    <changefreq>monthly</changefreq>
  </url>
  <url>
    <loc>${SITE_URL}/documentos/codigo-etica</loc>
    <priority>0.6</priority>
    <changefreq>monthly</changefreq>
  </url>
  <url>
    <loc>${SITE_URL}/documentos/reglamento-interno</loc>
    <priority>0.6</priority>
    <changefreq>monthly</changefreq>
  </url>

  <!-- Noticias y Comunicados -->
  ${posts.map(post => `
  <url>
    <loc>${SITE_URL}/noticias/${toNewsSlug(post.title, post.id)}</loc>
    <lastmod>${post.updated_at ? post.updated_at.split('T')[0] : new Date().toISOString().split('T')[0]}</lastmod>
    <changefreq>monthly</changefreq>
    <priority>0.7</priority>
  </url>`).join("")}

  <!-- Perfiles de Psicólogos -->
  ${psis.map(p => `
  <url>
    <loc>${SITE_URL}/directorio/${toPsiSlug(p)}</loc>
    <changefreq>monthly</changefreq>
    <priority>0.6</priority>
  </url>`).join("")}
</urlset>`.trim();

  return new Response(sitemap, {
    headers: {
      "Content-Type": "application/xml",
      "Cache-Control": "public, max-age=3600", // El navegador/Google lo guarda 1 hora
    },
  });
}