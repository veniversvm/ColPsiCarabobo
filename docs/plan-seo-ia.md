# Plan: SEO + legibilidad por IA — frontend (`web/`)

> Documento de ejecución diferida. **No aplicar todavía**: este archivo es la guía
> para ejecutar después. Ante cualquier discrepancia con el código, releer los
> archivos antes de editar ("no modifique el funcionamiento").
>
> Referencia de contexto: `web/AGENTS.md` (gotchas 1–9) y `AGENTS.md` raíz.

## Objetivo

1. Que los buscadores (SEO) lean el contenido real de los perfiles de psicólogos
   y de todo el contenido público no `admin/`.
2. Que los agentes de IA (ChatGPT, Gemini, Claude, etc.) entiendan el sitio con
   facilidad: datos estructurados JSON-LD (Schema.org) + `llms.txt`.

## Hallazgos de la auditoría (estado actual)

| Página | SSR hoy | Meta/OG | Contenido indexable | Archivo |
|--------|---------|---------|---------------------|---------|
| `/` (home) | ✅ | ❌ cero meta | ⚠️ poco contenido | `web/src/routes/index.tsx` |
| `/explorar` | ✅ | ✅ | ✅ | `web/src/routes/explorar.tsx` |
| `/nosotros` | ✅ | ✅ | ✅ | `web/src/routes/nosotros.tsx` |
| `/inscripcion` | ✅ | ✅ | ✅ | `web/src/routes/inscripcion.tsx` |
| `/directorio` | ❌ `ssr=false` | ✅ | ❌ grid client-only | `web/src/routes/directorio/index.tsx` |
| `/directorio/[slug]` | ✅ | ✅ | ⚠️ `full_bio` en modal client | `web/src/routes/directorio/[slug].tsx` |
| `/noticias` | ⚠️ default | ❌ sin meta | ❌ posts en `onMount` | `web/src/routes/noticias/index.tsx` |
| `/noticias/[slug]` | ✅ | ✅ | ✅ | `web/src/routes/noticias/[slug].tsx` |
| `/login`, `/admin-access` | ✅ | — | deberían `noindex` | `web/src/routes/login.tsx`, `admin-access.tsx` |
| `/public/*` | ✅ | — | ⚠️ 5 rutas VACÍAS indexables | `web/src/routes/public/**` |
| robots.txt | — | — | ⚠️ no bloquea `/public/`, `/login` | `web/src/routes/robots.txt.ts` |
| sitemap.xml | — | — | ⚠️ slug ≠ canónico; falta `lastmod` psi | `web/src/routes/sitemap.xml.ts` |

**Problemas transversales:**
- Defaults de `SITE_URL` inconsistentes: `directorio/[slug].tsx:21` y `nosotros.tsx:6`
  usan `http://localhost:3000`; `noticias/[slug].tsx:22` usa `https://colpsi-carabobo.org`;
  `explorar.tsx:4` e `inscripcion.tsx:6` usan `https://colpsicarabobo.org`;
  `sitemap.xml.ts:4` usa `http://localhost:3000`. El valor real vive en `.env`
  (`VITE_SITE_URL=http://localhost:23000` en dev).
- El slug del sitemap (`toPsiSlug`, `sitemap.xml.ts:24`) usa SOLO
  `first_name-last_name-fpv-N`; el front usa `createProfileSlug` con los 4 nombres
  (`lib/utils.ts:24`). Las URLs del sitemap no coinciden con las que enlazan las
  tarjetas. (Funcionan por el parser tolerante `extractFpvFromSlug` pero conviene
  unificarlas.)
- Ninguna página emite JSON-LD.
- No existe `llms.txt`.

**Decisión del dueño (confirmada):**
- Dominio: aún no hay dominio de producción → el default de `SITE_URL` queda neutro
  (`http://localhost:3000`) y el dominio real se inyecta por `VITE_SITE_URL` en build.
- SSR inicial en los dos listados: sí.
- Bio completa: sección `<details>` nativa SSR (reemplaza el modal).
- IA: `llms.txt` índice + JSON-LD.

---

## Orden de ejecución

> Un solo commit por fix en la rama `docs` (AGENTS.md raíz). Los pasos están
> agrupados en fixes atómicos; cada grupo termina con `npm run build` en verde
> y, si es posible, un commit.

### Fix 0 — Centralizar `SITE_URL` (base para todos los demás)

**Archivo:** `web/src/lib/bucket.ts`

Añadir una exportación canónica y reutilizarla en `siteUrl()`:

```ts
// Default neutro; el dominio real se inyecta por VITE_SITE_URL en build.
export const SITE_URL = import.meta.env.VITE_SITE_URL || "http://localhost:3000";
```

Y cambiar la constante interna para que no se duplique:

```ts
const BUCKET_URL = import.meta.env.VITE_BUCKET_URL || "";
// SITE_URL ahora exportado arriba (eliminar la línea local `const SITE_URL = ...`)
```

`siteUrl(path)` queda igual (usa la constante exportada).

**Archivos a actualizar** (reemplazar el `const SITE_URL = ...` local por import de `~/lib/bucket`):
1. `web/src/routes/directorio/[slug].tsx:21` → `import { bucketUrl, siteUrl, SITE_URL } from "~/lib/bucket";`
2. `web/src/routes/noticias/[slug].tsx:22`
3. `web/src/routes/explorar.tsx:4`
4. `web/src/routes/nosotros.tsx:6`
5. `web/src/routes/inscripcion.tsx:6`
6. `web/src/routes/sitemap.xml.ts:4`

En todos, la URL canónica pasa a `siteUrl("/path")` (ej. `siteUrl("/explorar")`) o se usa
`SITE_URL` directamente donde ya se concatena.

**Verificación:** `npm run build` OK. No cambia comportamiento visible (mismo valor
de `.env`).

---

### Fix 1 — SSR inicial en `/directorio` (lista de tarjetas)

**Archivo:** `web/src/routes/directorio/index.tsx`

1. **Eliminar** `export const ssr = false;` (línea 28) — la ruta vuelve a SSR por defecto.
2. **Eliminar** el wrapper `clientOnly` (líneas 12–17) y el import de `clientOnly`.
   Reemplazar por import directo:

   ```tsx
   import { ResultsGrid } from "~/components/directory/ResultsGrid";
   ```

   `ResultsGrid` y `PsychologistCard` son SSR-safe (no usan APIs de browser), así
   que se renderizan en el HTML del servidor.
3. **Añadir** una server function para la búsqueda (patrón ya usado en
   `directorio/[slug].tsx:26` — las funciones `"use server"` SÍ pueden invocarse
   dentro de `createResource`; el gotcha de `useAction` aplica solo a actions de
   formulario):

   ```tsx
   async function searchDirectory(params: {
     q: string;
     area: string;
     loc: string;
     page: number;
   }): Promise<DirectoryResponse> {
     "use server";
     const url =
       `/psi/directory?q=${encodeURIComponent(params.q)}` +
       `&specialty=${encodeURIComponent(params.area)}` +
       `&location=${encodeURIComponent(params.loc)}` +
       `&limit=${LIMIT}&page=${params.page}`;
     return await apiGet<DirectoryResponse>(url);
   }
   ```
4. **Carga inicial SSR** con `createResource`. Llamarla como "first render" y
   poblar el store para que el grid salga en el HTML:

   ```tsx
   const [initial] = createResource(async () => {
     if (isCached()) return null;          // cache en memoria ya poblada (navegación SPA)
     const res = await searchDirectory({ q: "", area: "", loc: "", page: 1 });
     setAllItems(res?.data ?? []);
     setTotal(res?.total ?? 0);
     setPage(1);
     setHasMore((res?.total_pages ?? 1) > 1);
     setIsCached(true);
     setShowLoading(false);                // clave: sin esto el servidor solo emite el LoadingScreen
     return res;
   });
   ```
   - `createResource` suspende en SSR → el servidor espera, re-renderiza y la
     tarjeta queda en el HTML final (streaming).
   - `onMount` actual (líneas 78–85) conserva la llamada `executeSearch` SOLO como
     respaldo si el `createResource` no pobló nada (por seguridad, mantener la
     condición `if (!isCached())`).
   - **Gotcha estado global:** las señales de módulo (líneas 31–40) son compartidas
     entre requests SSR en el mismo proceso. La carga inicial DEBE venir del
     `createResource` (arriba), no de `executeSearch`, para no servir estado de
     otro request.
5. **Refactorizar** `executeSearch` y `loadMore` para reutilizar la server function
   (en cliente se resuelve por RPC; misma URL, mismo comportamiento):

   ```tsx
   const executeSearch = async (params) => {
     setLoading(true);
     try {
       const res = await searchDirectory({ ...params, page: 1 });
       setAllItems(res?.data ?? []);
       setTotal(res?.total ?? 0);
       setPage(1);
       setHasMore((res?.total_pages ?? 1) > 1);
       setIsCached(true);
     } catch (err) {
       console.error("[directorio] error de búsqueda:", err);
       setAllItems([]);
       setHasMore(false);
     } finally {
       setLoading(false);
     }
   };

   const loadMore = async () => {
     if (loadingMore() || !hasMore() || loading()) return;
     setLoadingMore(true);
     try {
       const params = searchParams();
       const res = await searchDirectory({ ...params, page: page() + 1 });
       setAllItems((prev) => [...prev, ...(res?.data ?? [])]);
       setPage((p) => p + 1);
       setHasMore(page() < (res?.total_pages ?? 1));
     } catch (err) {
       console.error("[directorio] error scroll:", err);
     } finally {
       setLoadingMore(false);
     }
   };
   ```
   Conservar intactos: `SearchHeader`, `handleSearch`, `clearSearch`,
   `getWorkAreaName`, el observer de scroll infinito, el estado de filtros en
   query params y el `LoadingScreen`.
6. **JSON-LD opcional (recomendado)** en el JSX, cuando haya datos:

   ```tsx
   <Show when={allItems.length > 0}>
     <script type="application/ld+json">
       {JSON.stringify({
         "@context": "https://schema.org",
         "@type": "ItemList",
         name: "Directorio de Psicólogos | COLPSI Carabobo",
         itemListElement: allItems.map((p, i) => ({
           "@type": "ListItem",
           position: i + 1,
           name: `${p.first_name} ${p.last_name}`,
           url: siteUrl(`/directorio/${createProfileSlug({ ...p })}`),
         })),
       }).replace(/</g, "\\u003c")}
     </script>
   </Show>
   ```
   (Requiere importar `createProfileSlug` de `~/lib/utils` y `siteUrl` de `~/lib/bucket`.)

**Verificación:** build OK. Servir y `curl localhost:23000/directorio` → el HTML
debe contener nombres y FPV de psicólogos (no solo el loading).

---

### Fix 2 — SSR inicial en `/noticias` (listado) + meta

**Archivo:** `web/src/routes/noticias/index.tsx`

1. **Añadir** server function:

   ```tsx
   async function fetchPosts(page: number, q: string): Promise<PaginatedPosts> {
     "use server";
     const params = new URLSearchParams({ page: String(page), limit: String(LIMIT) });
     if (q) params.set("search", q);
     return await apiGet<PaginatedPosts>(`/posts?${params.toString()}`);
   }
   ```
2. **Carga inicial** vía `createResource` (reemplaza el `loadPage(1, "")` del
   `onMount` inicial; el `onMount` conserva solo el setup del IntersectionObserver):

   ```tsx
   const [initial] = createResource(async () => {
     const data = await fetchPosts(1, "");
     const result = data?.data ?? [];
     setPosts(result);
     setHasMore(result.length === LIMIT);
     setInitialDone(true);   // clave: apaga el LoadingScreen en el HTML SSR
     return data;
   });
   ```
3. **Mantener** `loadPage` para el scroll infinito y la búsqueda con debounce
   (client-side), pero apuntando a `fetchPosts(page, q)` en lugar de `apiGet`
   directo (misma URL, mismo contrato).
4. **Añadir meta** al inicio del retorno de la página (hoy no tiene):

   ```tsx
   <Title>Noticias y Comunicados | COLPSI Carabobo</Title>
   <Meta name="description" content="Noticias, comunicados y novedades oficiales del Colegio de Psicólogos del Estado Carabobo, Venezuela." />
   <Meta name="robots" content="index, follow" />
   ```
   (Importar `Title, Meta` de `@solidjs/meta`.)

**Verificación:** build OK. `curl localhost:23000/noticias` → el HTML debe contener
títulos de publicaciones y el `<title>` de la página.

---

### Fix 3 — Perfil de psicólogo: bio completa SSR + JSON-LD

**Archivo:** `web/src/routes/directorio/[slug].tsx`

1. **Eliminar** el modal client-only:
   - Quitar import y uso de `FullBioModal` (componente `src/components/directory/FullBioModal.tsx`
     queda sin uso → retirarlo del árbol, precedente "dead code removal").
   - Quitar el signal `showBioModal` y el handler.
2. **Reemplazar** el botón "Leer biografía completa" (líneas 186–196) por una
   sección nativa colapsable que renderiza el contenido en el DOM:

   ```tsx
   <Show when={psi().full_bio_content && psi().full_bio_content !== "<p></p>"}>
     <details class="mt-6 group rounded-2xl border border-gray-100 bg-gray-50/50">
       <summary class="cursor-pointer select-none inline-flex items-center gap-2 bg-colpsi-blue/5 hover:bg-colpsi-blue hover:text-white text-colpsi-blue font-bold px-5 py-2.5 rounded-xl transition-all text-sm">
         Leer biografía completa
       </summary>
       <div
         class="mt-4 px-5 pb-5 prose prose-lg max-w-none prose-p:text-gray-700 prose-p:leading-relaxed prose-headings:font-black prose-headings:text-blue-900"
         innerHTML={sanitizeHtml(psi().full_bio_content)}
       />
     </details>
   </Show>
   ```
   - Importar `sanitizeHtml` de `~/lib/sanitize-html`.
   - Mantener el `<Show when={psi().mini_bio}>` tal cual.
3. **JSON-LD** (dentro del `<Show when={profileData()}>` que ya existe, además de
   las etiquetas `profile:*`):

   ```tsx
   <script type="application/ld+json">
     {JSON.stringify({
       "@context": "https://schema.org",
       "@type": "ProfessionalService",
       "@id": canonicalUrl,
       name: fullName(),
       url: canonicalUrl,
       image: ogImage(),
       description: description(),
       identifier: `FPV-${p().fpv}`,
       jobTitle: "Psicólogo(a)",
       areaServed: { "@type": "State", name: "Carabobo" },
       address: { "@type": "PostalAddress", addressCountry: "VE", addressRegion: "Carabobo" },
       sameAs: (p().social_networks ?? []).map((s) => s.url).filter(Boolean),
     }).replace(/</g, "\\u003c")}
   </script>
   ```
   - Añadir también `BreadcrumbList`:

   ```tsx
   <script type="application/ld+json">
     {JSON.stringify({
       "@context": "https://schema.org",
       "@type": "BreadcrumbList",
       itemListElement: [
         { "@type": "ListItem", position: 1, name: "Inicio", item: siteUrl("/") },
         { "@type": "ListItem", position: 2, name: "Directorio", item: siteUrl("/directorio") },
         { "@type": "ListItem", position: 3, name: fullName(), item: canonicalUrl },
       ],
     }).replace(/</g, "\\u003c")}
   </script>
   ```
   - `fullName()`, `ogImage()`, `description()` y `canonicalUrl` ya existen en el
     componente (líneas 57–75).

**Verificación:** build OK. `curl` de un perfil → el HTML debe incluir la bio
completa (HTML sanitizado) y los dos bloques `application/ld+json`.

---

### Fix 4 — Noticia: JSON-LD `NewsArticle` + `BreadcrumbList`

**Archivo:** `web/src/routes/noticias/[slug].tsx`

Dentro del `<Show when={postData()}>` existente (donde ya están los `article:*`),
añadir:

```tsx
<script type="application/ld+json">
  {JSON.stringify({
    "@context": "https://schema.org",
    "@type": "NewsArticle",
    headline: data().title,
    image: ogImage(),
    description: data().short_description,
    datePublished: data().created_at,
    dateModified: data().updated_at || data().created_at,
    mainEntityOfPage: { "@type": "WebPage", "@id": canonicalUrl },
    author: { "@type": "Organization", name: "Colegio de Psicólogos del Estado Carabobo" },
    publisher: {
      "@type": "Organization",
      name: "Colegio de Psicólogos del Estado Carabobo",
      url: siteUrl("/"),
    },
  }).replace(/</g, "\\u003c")}
</script>
```

Y `BreadcrumbList` (Inicio › Noticias › Título), usando `canonicalUrl` y `siteUrl`.
Importar `siteUrl` de `~/lib/bucket`.

**Verificación:** build OK.

---

### Fix 5 — Páginas estáticas: home + `noindex` de auth + limpieza `/public/`

#### 5a. Home (`web/src/routes/index.tsx`)
Hoy no tiene meta. Añadir al inicio del retorno:

```tsx
import { Title, Meta, Link } from "@solidjs/meta";
import { siteUrl } from "~/lib/bucket";
```

```tsx
const canonicalUrl = siteUrl("/");
const pageTitle = "Colegio de Psicólogos del Estado Carabobo";
const pageDescription =
  "Colegio de Psicólogos del Estado Carabobo. Directorio de psicólogos colegiados, noticias, trámites de inscripción e información institucional.";

<Title>{pageTitle}</Title>
<Meta name="description" content={pageDescription} />
<Meta name="keywords" content="psicólogos, Carabobo, Valencia, salud mental, colegio de psicólogos, directorio, Venezuela" />
<Meta name="robots" content="index, follow" />
<Meta property="og:type" content="website" />
<Meta property="og:url" content={canonicalUrl} />
<Meta property="og:title" content={pageTitle} />
<Meta property="og:description" content={pageDescription} />
<Meta property="og:site_name" content="Colegio de Psicólogos del Estado Carabobo" />
<Meta property="og:locale" content="es_VE" />
<Link rel="canonical" href={canonicalUrl} />

<script type="application/ld+json">
  {JSON.stringify({
    "@context": "https://schema.org",
    "@type": "Organization",
    name: "Colegio de Psicólogos del Estado Carabobo",
    url: siteUrl("/"),
    logo: siteUrl("/psi.png"),
    address: { "@type": "PostalAddress", addressCountry: "VE", addressRegion: "Carabobo", addressLocality: "Valencia" },
    description: pageDescription,
  }).replace(/</g, "\\u003c")}
</script>
```

#### 5b. `noindex` de páginas de acceso
- `web/src/routes/login.tsx` → `<Meta name="robots" content="noindex, nofollow" />`
- `web/src/routes/admin-access.tsx` → `<Meta name="robots" content="noindex, nofollow" />`

#### 5c. Limpiar rutas vacías `/public/*`
Borrar los 5 archivos vacíos (renderizan páginas 200 en blanco, "thin content"):
- `web/src/routes/public/index.tsx`
- `web/src/routes/public/login.tsx`
- `web/src/routes/public/directorio/index.tsx`
- `web/src/routes/public/directorio/[id].tsx`
- `web/src/routes/public/noticias/index.tsx`
- `web/src/routes/public/noticias/[id].tsx`

(No hay imports que los referencien; FileRoutes deja de generar `/public/*`.)

**Verificación:** build OK. `curl localhost:23000/public/...` → 404 (antes 200 vacío).

---

### Fix 6 — robots.txt, sitemap.xml, llms.txt

#### 6a. `web/src/routes/robots.txt.ts`
Ampliar los bloqueos (mantener todo lo existente):

```txt
User-agent: *
Allow: /
Disallow: /admin/
Disallow: /admin-access
Disallow: /psi/
Disallow: /public/
Disallow: /login

Sitemap: <siteUrl("/sitemap.xml")>
```

#### 6b. `web/src/routes/sitemap.xml.ts`
1. **Unificar slug de psicólogos** con el front: usar `createProfileSlug` de
   `~/lib/utils` (importar) en lugar del `toPsiSlug` local. Ajustar `toNewsSlug`
   igual que la ruta `noticias/index.tsx` (la regex actual es equivalente; dejarla
   pero alinearla con `noticias/[slug].tsx` si difiere).
2. **Requiere un micro-cambio en la API** para que el sitemap traiga los 4 nombres:
   `api/internal/repository/postgres/psi_repository.go:870`:
   ```go
   Select("first_name, second_name, last_name, second_last_name, fpv, updated_at").
   ```
   (Los campos fluyen al JSON porque el handler serializa `[]domain.PsiUserModel`
   completo; solo se llenan los seleccionados. Añadir `updated_at` para `lastmod`.)
3. Generar la URL del perfil:
   ```ts
   const psiUrl = (p: any) =>
     `${SITE_URL}/directorio/${createProfileSlug({
       first_name: p.first_name,
       second_name: p.second_name,
       last_name: p.last_name,
       second_last_name: p.second_last_name,
       fpv: p.fpv,
     })}`;
   ```
   Y en el bloque de psicólogos añadir `<lastmod>` cuando `p.updated_at` exista.
4. Añadir `<url>` para `/explorar` y `/nosotros` (priority 0.7/0.8), además de
   los ya existentes (`,` `/directorio`, `/noticias`, `/inscripcion`).
5. Sustituir `SITE_URL` local por el export de `~/lib/bucket`.

#### 6c. Nuevo `web/src/routes/llms.txt.ts`
`GET` → `text/plain; charset=utf-8`, `Cache-Control: public, max-age=3600`.
Contenido (se puede cargar `sitemap-data` en `try/catch` para listar un muestreo
de perfiles, igual que hace `sitemap.xml.ts`):

```text
# Colegio de Psicólogos del Estado Carabobo (COLPSI Carabobo)

Colegio profesional que agrupa, representa y regula a los psicólogos licenciados
que ejercen o residen en el Estado Carabobo, Venezuela. Sede en Valencia.
Registro profesional: número FPV.

Secciones principales:
- [Inicio](SITE_URL/): portada institucional.
- [Explorar](SITE_URL/explorar): acceso a directorio, noticias, inscripción y nosotros.
- [Directorio de Psicólogos](SITE_URL/directorio): listado público de psicólogos
  colegiados (búsqueda por nombre, área y ubicación).
  Perfil de cada profesional en: /directorio/{nombre}-fpv-{numero}
- [Noticias](SITE_URL/noticias): noticias, comunicados y avisos oficiales.
- [Inscripción](SITE_URL/inscripcion): procedimiento y requisitos de colegiatura.
- [Nosotros](SITE_URL/nosotros): historia, misión, visión y valores de la institución.

Accesos privados (no indexar): /admin/, /psi/, /login.

# Directorio (muestreo)
{por cada psi de sitemap-data:}
- [Nombre Apellido](SITE_URL/directorio/slug): psicólogo(a) colegiado(a) FPV N.
```

> El bloque "Directorio (muestreo)" es opcional; si la API falla, el archivo se
> sirve igual con las secciones estáticas (try/catch como en `sitemap.xml.ts:44`).

**Verificación:** `curl` de `/robots.txt`, `/sitemap.xml` y `/llms.txt` → contenido
esperado; las URLs del sitemap deben coincidir con los links de las tarjetas del
directorio.

---

### Fix 7 — Defaults globales (`web/src/entry-server.tsx`)

1. **OG defaults** dentro del `<head>` (las páginas con meta propia lo sobreescriben
   porque Solid Meta hace merge):

   ```tsx
   <meta name="description" content="Colegio de Psicólogos del Estado Carabobo: directorio de psicólogos colegiados, noticias e información institucional." />
   <meta property="og:site_name" content="Colegio de Psicólogos del Estado Carabobo" />
   <meta property="og:type" content="website" />
   <meta name="twitter:card" content="summary_large_image" />
   ```

2. **Enlace a `llms.txt`** (convención para crawlers de IA):

   ```tsx
   <link rel="alternate" type="text/markdown" href="/llms.txt" />
   ```

3. **NO tocar el CSP.** Los `<script type="application/ld+json">` son inline y ya
   están permitidos por `'unsafe-inline'`; el origen del bucket en `img-src` se
   mantiene intacto.

**Verificación:** build OK; el HTML base incluye los defaults y el link a llms.txt.

---

## Verificación global (final)

1. `npm run build` — compila (es la única verificación de tipos disponible).
2. Servir y comprobar con `curl`:
   - `/` → `<title>`, meta description, canonical, JSON-LD `Organization`.
   - `/directorio` → tarjetas con nombres/FPV en el HTML (no solo loading).
   - `/directorio/<slug>` → `mini_bio` + bio completa + JSON-LD `ProfessionalService`
     y `BreadcrumbList` + canonical.
   - `/noticias` → títulos de posts en el HTML + meta.
   - `/noticias/<slug>` → JSON-LD `NewsArticle` + contenido.
   - `/llms.txt`, `/robots.txt`, `/sitemap.xml` → OK y coherentes.
   - `/public/...` → 404.
   - `/login`, `/admin-access` → `noindex` en el HTML.
3. Con Docker (opcional): `docker compose build web && docker compose up -d`;
   el CSP debe incluir el origen del bucket en `img-src`.
4. Regresión manual de UX (lo interactivo NO debe cambiar):
   - Directorio: búsqueda por query/área/ubicación, filtros activos, "limpiar",
     scroll infinito.
   - Noticias: scroll infinito y búsqueda con debounce.
   - Perfil: `<details>` abre/cierra la bio; redes, ubicaciones y académico intactos.

## Notas de respeto al repo

- Server functions `"use server"` para **carga de datos** pueden invocarse dentro
  de `createResource` (patrón ya usado en `directorio/[slug].tsx:26` y
  `noticias/[slug].tsx:35`). El gotcha de `useAction()` (AGENTS 1) aplica solo a
  **actions** de formulario.
- `VITE_*` se inlinean en build: cambiar `.env` exige reconstruir la imagen.
- No agregar comentarios salvo gotchas/decisiones (cabeceras de `src/lib/*`).
- Un solo commit por fix en rama `docs`.
