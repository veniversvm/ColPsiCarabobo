// web/src/routes/noticias/[slug].tsx
import { createSignal, onMount, Show } from "solid-js";
import { useParams, useLocation, A } from "@solidjs/router";
import { apiGet } from "~/lib/api";

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

const BUCKET_URL = import.meta.env.VITE_BUCKET_URL || "http://localhost:9000/colpsi-bucket";
const imgUrl = (key: string) => key ? `${BUCKET_URL}/${key}` : "";

const formatDate = (iso: string) =>
  new Date(iso).toLocaleDateString("es-VE", { day: "numeric", month: "long", year: "numeric" });

const formatDateShort = (iso: string) =>
  new Date(iso).toLocaleDateString("es-VE", { day: "2-digit", month: "short", year: "numeric" });

export default function PublicNoticiaDetailPage() {
  const params = useParams<{ slug: string }>();
  const location = useLocation();

  // undefined = cargando | null = no encontrado | PostDetail = éxito
  const [post, setPost] = createSignal<PostDetail | null | undefined>(undefined);

  const extractUuid = (): string => {
    // Prioridad 1: Si viene del state (navegación interna), usamos el ID completo
    const stateId = (location.state as any)?.id;
    if (stateId) return stateId;
    
    // Prioridad 2: Extraer del slug
    const slug = params.slug ?? "";
    
    // Buscar el patrón: todo después del último guión debería ser el primer segmento del UUID (8 caracteres hex)
    // Pero necesitamos reconstruir el UUID completo para la API
    const match = slug.match(/-([0-9a-f]{8})$/i);
    if (match) {
      const firstSegment = match[1];
      
      // Aquí necesitamos una forma de obtener el UUID completo
      // Como no podemos saber los segmentos restantes, usamos el state o 
      // podríamos modificar la API para que acepte búsqueda por primer segmento
      
      // Opción 1: Mantener compatibilidad con UUID completo usando el primer segmento
      // Esto asume que el UUID completo tiene el formato estándar 8-4-4-4-12
      // Pero no podemos reconstruirlo sin los otros segmentos
      
      console.warn("Slug con primer segmento detectado, pero no se puede reconstruir UUID completo");
      return firstSegment; // Esto fallará en la API si espera UUID completo
    }
    
    // Fallback: devolver el slug completo (puede ser un ID directo o un formato antiguo)
    return slug;
  };

  onMount(async () => {
    const id = extractUuid();
    if (!id) { setPost(null); return; }
    
    try {
      // Intentar buscar por ID (puede ser UUID completo o primer segmento)
      // Si tu API soporta búsqueda por fragmento de UUID, mejor
      const data = await apiGet<PostDetail>(`/posts/${id}`);
      setPost(data ?? null);
    } catch {
      setPost(null);
    }
  });

  return (
    <main class="min-h-screen bg-[#f7f5f0]">

      {/* Skeleton mientras carga */}
      <Show when={post() === undefined}>
        <ArticleSkeleton />
      </Show>

      {/* No encontrado */}
      <Show when={post() === null}>
        <div class="max-w-3xl mx-auto px-4 py-24 text-center">
          <p class="text-5xl mb-4">😕</p>
          <h1 class="text-xl font-black text-gray-700 mb-2">Publicación no encontrada</h1>
          <p class="text-gray-400 text-sm mb-6">Es posible que haya sido eliminada o no exista.</p>
          <A href="/noticias" class="inline-flex items-center gap-2 text-blue-700 font-black text-sm hover:underline">
            ← Volver a noticias
          </A>
        </div>
      </Show>

      {/* Contenido */}
      <Show when={post() !== undefined && post() !== null}>
        {(() => {
          const data = post() as PostDetail;
          return (
            <>
              {/* HERO */}
              <div class="relative bg-[#0d2b5e] overflow-hidden">
                <Show when={data.image_url}>
                  <img src={imgUrl(data.image_url)} alt={data.title}
                    class="absolute inset-0 w-full h-full object-cover opacity-20" />
                </Show>
                <div class="absolute inset-0 opacity-5"
                  style="background-image: repeating-linear-gradient(45deg, #fff 0, #fff 1px, transparent 0, transparent 50%); background-size: 12px 12px;" />
                <div class="relative max-w-3xl mx-auto px-4 py-16 md:py-20">
                  <nav class="flex items-center gap-2 text-[11px] text-blue-300 font-bold mb-8">
                    <A href="/" class="hover:text-white transition-colors">Inicio</A>
                    <span class="opacity-40">/</span>
                    <A href="/noticias" class="hover:text-white transition-colors">Noticias</A>
                    <span class="opacity-40">/</span>
                    <span class="text-white/60 truncate max-w-[200px]">{data.title}</span>
                  </nav>
                  <p class="text-[11px] font-black text-yellow-400 uppercase tracking-[0.25em] mb-4">
                    {formatDate(data.created_at)}
                  </p>
                  <h1 class="text-3xl md:text-4xl font-black text-white leading-tight mb-5 max-w-2xl">
                    {data.title}
                  </h1>
                  <Show when={data.short_description}>
                    <p class="text-blue-200 text-base leading-relaxed max-w-xl mb-8">
                      {data.short_description}
                    </p>
                  </Show>
                  <div class="flex items-center gap-3 pt-6 border-t border-white/10">
                    <div class="w-8 h-8 rounded-full bg-blue-700 flex items-center justify-center font-black text-white text-xs">CP</div>
                    <div>
                      <p class="text-white text-xs font-black">Colegio de Psicólogos de Carabobo</p>
                      <p class="text-blue-300 text-[10px]">Publicado el {formatDateShort(data.created_at)}</p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Imagen portada */}
              <Show when={data.image_url}>
                <div class="max-w-3xl mx-auto px-4 -mt-8 mb-0 relative z-10">
                  <div class="rounded-2xl overflow-hidden shadow-2xl border border-white/20 aspect-video">
                    <img src={imgUrl(data.image_url)} alt={data.title} class="w-full h-full object-cover" />
                  </div>
                </div>
              </Show>

              {/* Contenido enriquecido */}
              <div class="max-w-3xl mx-auto px-4 py-12">
                <Show
                  when={data.text?.content}
                  fallback={
                    <div class="bg-white rounded-3xl p-12 text-center border border-gray-100">
                      <p class="text-gray-400 italic">Esta publicación no tiene contenido detallado.</p>
                    </div>
                  }
                >
                  <article
                    class="bg-white rounded-3xl shadow-sm border border-gray-100 px-8 md:px-14 py-12
                      prose prose-lg max-w-none
                      prose-headings:font-black prose-headings:text-blue-900 prose-headings:tracking-tight
                      prose-p:text-gray-700 prose-p:leading-relaxed prose-p:text-[1.05rem]
                      prose-a:text-blue-700 prose-a:font-bold prose-a:no-underline hover:prose-a:underline
                      prose-strong:text-gray-900 prose-strong:font-black
                      prose-blockquote:border-l-4 prose-blockquote:border-yellow-400 prose-blockquote:bg-yellow-50
                      prose-blockquote:rounded-r-xl prose-blockquote:py-2 prose-blockquote:not-italic
                      prose-blockquote:text-gray-700 prose-li:text-gray-700
                      prose-img:rounded-2xl prose-img:shadow-md"
                    innerHTML={data.text.content}
                  />
                </Show>

                {/* Pie */}
                <div class="mt-10 flex flex-col sm:flex-row items-center justify-between gap-4 pt-8 border-t border-gray-200">
                  <A href="/noticias" class="inline-flex items-center gap-2 text-blue-700 font-black text-sm hover:gap-3 transition-all">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M11 17l-5-5m0 0l5-5m-5 5h12" />
                    </svg>
                    Volver a noticias
                  </A>
                  <div class="flex items-center gap-2 text-[10px] text-gray-400 font-bold">
                    <Show when={data.updated_at !== data.created_at}>
                      <span>Actualizado: {formatDateShort(data.updated_at)}</span>
                      <span>·</span>
                    </Show>
                    <span>ColPsi Carabobo © {new Date().getFullYear()}</span>
                  </div>
                </div>
              </div>
            </>
          );
        })()}
      </Show>

    </main>
  );
}

function ArticleSkeleton() {
  return (
    <div class="animate-pulse">
      <div class="bg-gray-200 h-72" />
      <div class="max-w-3xl mx-auto px-4 py-12 space-y-6">
        <div class="bg-white rounded-3xl p-12 space-y-4 border border-gray-100">
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