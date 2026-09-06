// web/src/routes/noticias/[slug].tsx
import { createEffect, createResource, Show } from "solid-js";
import { useParams, useLocation, A } from "@solidjs/router";
import { apiGet, ApiError } from "~/lib/api";
import { Meta, Title, Link } from "@solidjs/meta";
import { sanitizeHtml } from "~/lib/sanitize-html";
import { bucketUrl as imgUrl } from "~/lib/bucket";

export const ssr = true;

interface PostDetail {
  id: string;
  title: string;
  short_description: string;
  image_url: string;
  created_at: string;
  updated_at: string;
  create_by: string;
  text: { id: string; content: string };
}

const SITE_URL = import.meta.env.VITE_SITE_URL || "https://colpsi-carabobo.org";

const formatDate = (iso: string) =>
  new Date(iso).toLocaleDateString("es-VE", { day: "numeric", month: "long", year: "numeric" });

const formatDateShort = (iso: string) =>
  new Date(iso).toLocaleDateString("es-VE", { day: "2-digit", month: "short", year: "numeric" });

const extractIdFromSlug = (slug: string): string | null => {
  const match = slug.match(/-([0-9a-f]{8})$/i);
  return match ? match[1] : null;
};

async function fetchPost(id: string): Promise<PostDetail | null> {
  "use server";
  if (!id) return null;
  try {
    return await apiGet<PostDetail>(`/posts/${id}`);
  } catch (err: any) {
    if (err instanceof ApiError) {
      console.warn(`[fetchPost] ${err.status} – id=${id}: ${err.message}`);
      return null;
    }
    console.error(`[fetchPost] Error inesperado – id=${id}:`, err);
    return null;
  }
}

export default function PublicNoticiaDetailPage() {
  const params = useParams<{ slug: string }>();
  const location = useLocation();

  const [post] = createResource(async () => {
    try {
      const stateId = (location.state as any)?.id as string | undefined;
      const id = stateId || extractIdFromSlug(params.slug);
      if (!id) return null;
      return await fetchPost(id);
    } catch {
      return null;
    }
  });

  const postData = () => post();
  const canonicalUrl = `${SITE_URL}/noticias/${params.slug}`;
  const ogImage = () =>
    postData()?.image_url ? imgUrl(postData()!.image_url) : `${SITE_URL}/og-default.jpg`;

  // createEffect(() => {
  //   if (postData())
  //     console.log(postData())
  // })

  const title: string = postData() ? `${postData()!.title} | COLPSI Carabobo` : "Noticia | COLPSI Carabobo"

  return (
    <>
      <Title>
        {title}
      </Title>
      <Meta
        name="description"
        content={
          postData()?.short_description ||
          "Noticias y comunicados oficiales del Colegio de Psicólogos del estado Carabobo, Venezuela."
        }
      />
      <Meta name="keywords" content="psicología, Carabobo, Venezuela, noticias, comunicados, colegio de psicólogos" />

      <Meta property="og:type" content="article" />
      <Meta property="og:url" content={canonicalUrl} />
      <Meta property="og:title" content={postData()?.title || "Noticia | COLPSI Carabobo"} />
      <Meta
        property="og:description"
        content={
          postData()?.short_description ||
          "Noticias y comunicados oficiales del Colegio de Psicólogos del estado Carabobo."
        }
      />
      <Meta property="og:image" content={ogImage()} />
      <Meta property="og:image:alt" content={postData()?.title || "COLPSI Carabobo"} />
      <Meta property="og:site_name" content="COLPSI Carabobo" />
      <Meta property="og:locale" content="es_VE" />

      <Show when={postData()}>
        {(data) => (
          <>
            <Meta property="article:published_time" content={data().created_at} />
            <Meta property="article:modified_time" content={data().updated_at} />
            <Meta property="article:author" content="Colegio de Psicólogos de Carabobo" />
            <Meta property="article:section" content="Noticias" />
          </>
        )}
      </Show>

      <Meta name="twitter:card" content="summary_large_image" />
      <Meta name="twitter:title" content={postData()?.title || "Noticia | COLPSI Carabobo"} />
      <Meta
        name="twitter:description"
        content={
          postData()?.short_description ||
          "Noticias y comunicados oficiales del Colegio de Psicólogos del estado Carabobo."
        }
      />
      <Meta name="twitter:image" content={ogImage()} />
      <Meta name="twitter:image:alt" content={postData()?.title || "COLPSI Carabobo"} />

      <Link rel="canonical" href={canonicalUrl} />
      <Meta name="robots" content="index, follow" />
      <Meta name="googlebot" content="index, follow" />

      <main class="min-h-screen bg-colpsi-bg">

        <Show when={post.loading}>
          <ArticleSkeleton />
        </Show>

        <Show when={!post.loading && postData() === null}>
          <NotFoundState />
        </Show>

        <Show when={postData()}>
          {(data) => (
            <>
              {/* HERO */}
              <div class="relative bg-colpsi-blue-dark overflow-hidden">
                <Show when={data().image_url}>
                  <img
                    src={imgUrl(data().image_url)}
                    alt={data().title}
                    class="absolute inset-0 w-full h-full object-cover opacity-20"
                  />
                </Show>
                <div
                  class="absolute inset-0 opacity-5"
                  style="background-image: repeating-linear-gradient(45deg,#fff 0,#fff 1px,transparent 0,transparent 50%);background-size:12px 12px;"
                />
                <div class="relative max-w-3xl mx-auto px-4 py-16 md:py-20">
                  <nav class="flex items-center gap-2 text-[11px] text-blue-300 font-bold mb-8" aria-label="Breadcrumb">
                    <A href="/" class="hover:text-white transition-colors">Inicio</A>
                    <span class="opacity-40" aria-hidden="true">/</span>
                    <A href="/noticias" class="hover:text-white transition-colors">Noticias</A>
                    <span class="opacity-40" aria-hidden="true">/</span>
                    <span class="text-white/60 truncate max-w-[200px]">{data().title}</span>
                  </nav>

                  <time
                    class="text-[11px] font-black text-colpsi-yellow uppercase tracking-[0.25em] mb-4"
                    datetime={data().created_at}
                  >
                    {formatDate(data().created_at)}
                  </time>

                  <h1 class="text-3xl md:text-4xl font-black text-white leading-tight mb-5 max-w-2xl">
                    {data().title}
                  </h1>

                  <Show when={data().short_description}>
                    <p class="text-blue-200 text-base leading-relaxed max-w-xl mb-8">
                      {data().short_description}
                    </p>
                  </Show>

                  <div class="flex items-center gap-3 pt-6 border-t border-white/10">
                    <div
                      class="w-8 h-8 rounded-full bg-blue-700 flex items-center justify-center font-black text-white text-xs"
                      aria-label="Logo COLPSI"
                    >
                      CP
                    </div>
                    <div>
                      <p class="text-white text-xs font-black">Colegio de Psicólogos de Carabobo</p>
                      <p class="text-blue-300 text-[10px]">
                        Publicado el{" "}
                        <time datetime={data().created_at}>{formatDateShort(data().created_at)}</time>
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Imagen portada — se adapta a cualquier proporción */}
              <Show when={data().image_url}>
                <div class="max-w-3xl mx-auto px-4 -mt-8 mb-0 relative z-10">
                  <figure class="rounded-2xl overflow-hidden shadow-2xl border border-white/20 bg-gray-900 flex items-center justify-center">
                    <img
                      src={imgUrl(data().image_url)}
                      alt={data().title}
                      class="w-full h-auto max-h-[70vh] object-contain"
                      loading="eager"
                    />
                    <figcaption class="sr-only">Imagen destacada: {data().title}</figcaption>
                  </figure>
                </div>
              </Show>

              {/* Contenido */}
              <div class="max-w-3xl mx-auto px-4 py-12">
                <Show
                  when={data().text?.content}
                  fallback={
                    <div class="bg-white rounded-3xl p-12 text-center border border-colpsi-border">
                      <p class="text-gray-400 italic">Esta publicación no tiene contenido detallado.</p>
                    </div>
                  }
                >
                  <article
                    class="bg-white rounded-3xl shadow-sm border border-colpsi-border px-8 md:px-14 py-12
                      prose prose-lg max-w-none
                      prose-headings:font-black prose-headings:text-blue-900 prose-headings:tracking-tight
                      prose-p:text-gray-700 prose-p:leading-relaxed prose-p:text-[1.05rem]
                      prose-a:text-blue-700 prose-a:font-bold prose-a:no-underline hover:prose-a:underline
                      prose-strong:text-gray-900 prose-strong:font-black
                      prose-blockquote:border-l-4 prose-blockquote:border-yellow-400 prose-blockquote:bg-yellow-50
                      prose-blockquote:rounded-r-xl prose-blockquote:py-2 prose-blockquote:not-italic
                      prose-blockquote:text-gray-700 prose-li:text-gray-700
                      prose-img:rounded-2xl prose-img:shadow-md"
                    innerHTML={sanitizeHtml(data().text.content)}
                  />
                </Show>

                <footer class="mt-10 flex flex-col sm:flex-row items-center justify-between gap-4 pt-8 border-t border-gray-200">
                  <A
                    href="/noticias"
                    class="inline-flex items-center gap-2 text-blue-700 font-black text-sm hover:gap-3 transition-all"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M11 17l-5-5m0 0l5-5m-5 5h12" />
                    </svg>
                    Volver a noticias
                  </A>

                  <div class="flex items-center gap-2 text-[10px] text-gray-400 font-bold">
                    <Show when={data().updated_at !== data().created_at}>
                      <span>
                        Actualizado:{" "}
                        <time datetime={data().updated_at}>{formatDateShort(data().updated_at)}</time>
                      </span>
                      <span aria-hidden="true">·</span>
                    </Show>
                    <span>ColPsi Carabobo © {new Date().getFullYear()}</span>
                  </div>
                </footer>
              </div>
            </>
          )}
        </Show>

      </main>
    </>
  );
}

function ArticleSkeleton() {
  return (
    <div class="animate-pulse" role="status" aria-label="Cargando contenido...">
      <div class="bg-gray-200 h-72" />
      <div class="max-w-3xl mx-auto px-4 py-12 space-y-6">
        <div class="bg-white rounded-3xl p-12 space-y-4 border border-colpsi-border">
          <div class="h-4 bg-gray-100 rounded w-1/4" />
          <div class="h-6 bg-gray-100 rounded w-full" />
          <div class="h-6 bg-gray-100 rounded w-5/6" />
          <div class="h-4 bg-gray-100 rounded w-full" />
          <div class="h-4 bg-gray-100 rounded w-3/4" />
          <div class="h-4 bg-gray-100 rounded w-full" />
          <div class="h-4 bg-gray-100 rounded w-4/5" />
        </div>
      </div>
    </div>
  );
}

function NotFoundState() {
  return (
    <div class="max-w-3xl mx-auto px-4 py-24 text-center">
      <Meta name="robots" content="noindex, follow" />
      <p class="text-5xl mb-4" aria-hidden="true">😕</p>
      <h1 class="text-xl font-black text-gray-700 mb-2">Publicación no encontrada</h1>
      <p class="text-gray-400 text-sm mb-6">Es posible que haya sido eliminada o no exista.</p>
      <A
        href="/noticias"
        class="inline-flex items-center gap-2 text-blue-700 font-black text-sm hover:underline"
        aria-label="Volver al listado de noticias"
      >
        <span aria-hidden="true">←</span> Volver a noticias
      </A>
    </div>
  );
}