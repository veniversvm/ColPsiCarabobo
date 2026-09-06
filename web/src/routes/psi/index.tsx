// web/src/routes/psi/index.ts
/**
 * Inicio del Portal del Psicólogo ya logueado. Aquí se muestra un resumen del perfil, estatus de solvencia y noticias gremiales.
 * Desde aquí el psicólogo puede navegar a su perfil completo, postgrados y otras secciones relevantes.
 * 
 * FIX SENIOR: Este es el dashboard principal del psicólogo. Se carga información sensible del perfil, por lo que se accede a través de la cookie de autenticación.
 * El endpoint /psi/me en Go se encarga de devolver toda la información relevante del perfil en una sola petición para optimizar la experiencia.
 */
import { createResource, createSignal, For, Show, Suspense } from "solid-js";
import { useAuth } from "~/lib/auth";
import { apiGet } from "~/lib/api";
import { bucketUrl } from "~/lib/bucket";
import { A } from "@solidjs/router";

export default function PsiDashboard() {
  const { user } = useAuth();

  // Cargamos los datos extendidos del perfil y las noticias gremiales
  const [profile] = createResource(() => apiGet<any>("/psi/me"));
  const [news] = createResource(() => apiGet<any>("/posts?limit=3"));

  // Contador de notificaciones no leídas (para el badge)
  const [unread] = createResource(
    () => apiGet<{ unread_count: number }>("/notifications/psi-user/unread-count"),
    { initialValue: { unread_count: 0 } }
  );

  // Acceso a la Biblioteca Virtual (Audiobookshelf). Solo agremiados solventes.
  const [opening, setOpening] = createSignal(false);
  const [libError, setLibError] = createSignal("");

  const openLibrary = async () => {
    if (!profile()?.solvent || opening()) return;
    setOpening(true);
    setLibError("");
    try {
      const res = await apiGet<{ url: string }>("/psi/me/audiobookshelf");
      window.open(res.url, "_blank", "noopener,noreferrer");
    } catch (e: any) {
      setLibError(
        e?.message === "OFFLINE_SERVICE"
          ? "La biblioteca no está disponible. Intenta de nuevo."
          : e?.message || "No se pudo abrir la biblioteca. Intenta de nuevo."
      );
    } finally {
      setOpening(false);
    }
  };

  return (
    <main class="bg-colpsi-bg min-h-screen pb-20">
      {/* Cabecera de Bienvenida */}
      <div class="bg-heraldic pt-12 pb-20 px-6 shadow-inner">
        <div class="max-w-4xl mx-auto">
          <h1 class="text-white text-2xl font-bold">
            Hola, {profile()?.first_name || user()?.username}
          </h1>
          <p class="text-blue-200 text-sm">Bienvenido a tu portal gremial</p>
        </div>
      </div>

      {/* Contenido Principal (Subido sobre la cabecera) */}
      <div class="max-w-4xl mx-auto px-4 -mt-12 space-y-6">
        
        {/* Card de Estatus de Solvencia */}
        <Suspense fallback={<div class="h-32 bg-white animate-pulse rounded-3xl" />}>
          <div class="bg-white rounded-3xl p-6 shadow-sm border border-colpsi-border flex items-center justify-between">
            <div>
              <p class="text-xs font-bold text-gray-400 uppercase tracking-widest">Estatus de Solvencia</p>
              <Show when={profile()?.solvent} 
                fallback={<h2 class="text-colpsi-red text-xl font-black mt-1">INSOLVENTE</h2>}
              >
                <h2 class="text-green-600 text-xl font-black mt-1">AL DÍA</h2>
              </Show>
              <p class="text-xs text-colpsi-muted mt-1">FPV: {profile()?.fpv}</p>
            </div>
            <div class={`w-12 h-12 rounded-full flex items-center justify-center ${profile()?.solvent ? 'bg-green-100 text-green-600' : 'bg-red-100 text-colpsi-red'}`}>
              <Show when={profile()?.solvent} fallback="!">✓</Show>
            </div>
          </div>
        </Suspense>

        {/* Biblioteca Virtual */}
        <div class={`col-span-2 bg-white rounded-3xl p-4 border shadow-sm ${profile()?.solvent ? 'border-colpsi-border' : 'border-dashed border-gray-300'}`}>
          <button
            type="button"
            onClick={openLibrary}
            disabled={!profile()?.solvent || opening()}
            class={`w-full flex items-center gap-3 text-left group ${profile()?.solvent ? 'active:scale-[0.99] transition-transform' : 'cursor-not-allowed'}`}
          >
            <div class={`w-12 h-12 rounded-2xl flex items-center justify-center shrink-0 ${profile()?.solvent ? 'bg-blue-50 text-colpsi-blue group-hover:bg-colpsi-yellow transition-colors' : 'bg-gray-100 text-gray-400'}`}>
              🎧
            </div>
            <div class="flex-1">
              <span class={`text-xs font-bold ${profile()?.solvent ? 'text-colpsi-blue' : 'text-gray-400'}`}>
                Biblioteca Virtual
              </span>
              <p class="text-[11px] text-colpsi-muted mt-0.5 leading-snug">
                <Show when={profile()?.solvent} fallback="Ponte al día con el gremio para acceder a la biblioteca virtual">
                  Audiolibros y contenidos exclusivos para agremiados
                </Show>
              </p>
            </div>
            <Show when={profile()?.solvent} fallback={<span class="text-lg text-gray-300">🔒</span>}>
              <span class={`text-lg text-colpsi-blue ${opening() ? 'opacity-40' : 'group-hover:translate-x-0.5 transition-transform'}`}>→</span>
            </Show>
          </button>
          <Show when={libError()}>
            <p class="text-xs text-red-600 mt-2 pl-16">{libError()}</p>
          </Show>
        </div>

        {/* Acceso Rápido (Grid) */}
        <div class="grid grid-cols-2 gap-4">
          <A href="/psi/perfil" class="bg-white p-4 rounded-3xl border border-colpsi-border shadow-sm flex flex-col items-center text-center group active:scale-95 transition-transform">
            <div class="w-12 h-12 bg-blue-50 text-colpsi-blue rounded-2xl flex items-center justify-center mb-2 group-hover:bg-colpsi-yellow transition-colors">
              👤
            </div>
            <span class="text-xs font-bold text-colpsi-blue">Mi Perfil</span>
          </A>
          <A href="/psi/academico" class="bg-white p-4 rounded-3xl border border-colpsi-border shadow-sm flex flex-col items-center text-center group active:scale-95 transition-transform">
            <div class="w-12 h-12 bg-blue-50 text-colpsi-blue rounded-2xl flex items-center justify-center mb-2 group-hover:bg-colpsi-yellow transition-colors">
              🎓
            </div>
            <span class="text-xs font-bold text-colpsi-blue">Postgrados</span>
          </A>
          <A href="/psi/documentos" class="bg-white p-4 rounded-3xl border border-colpsi-border shadow-sm flex flex-col items-center text-center group active:scale-95 transition-transform">
            <div class="w-12 h-12 bg-blue-50 text-colpsi-blue rounded-2xl flex items-center justify-center mb-2 group-hover:bg-colpsi-yellow transition-colors">
              📁
            </div>
            <span class="text-xs font-bold text-colpsi-blue">Mis Documentos</span>
          </A>
          <A href="/psi/notificaciones" class="relative bg-white p-4 rounded-3xl border border-colpsi-border shadow-sm flex flex-col items-center text-center group active:scale-95 transition-transform">
            <div class="w-12 h-12 bg-blue-50 text-colpsi-blue rounded-2xl flex items-center justify-center mb-2 group-hover:bg-colpsi-yellow transition-colors">
              🔔
            </div>
            <span class="text-xs font-bold text-colpsi-blue">Notificaciones</span>
            <Show when={(unread()?.unread_count ?? 0) > 0}>
              <span class="absolute -top-2 -right-2 bg-red-500 text-white text-[10px] font-black w-6 h-6 rounded-full flex items-center justify-center shadow">
                {unread()?.unread_count}
              </span>
            </Show>
          </A>
          <A href="/psi/tickets" class="bg-white p-4 rounded-3xl border border-colpsi-border shadow-sm flex flex-col items-center text-center group active:scale-95 transition-transform">
            <div class="w-12 h-12 bg-blue-50 text-colpsi-blue rounded-2xl flex items-center justify-center mb-2 group-hover:bg-colpsi-yellow transition-colors">
              🎫
            </div>
            <span class="text-xs font-bold text-colpsi-blue">Solicitudes</span>
          </A>
        </div>

        {/* Noticias Recientes */}
        <section class="space-y-4">
          <div class="flex justify-between items-end px-2">
            <h3 class="text-colpsi-blue font-bold">Noticias del Colegio</h3>
            <A href="/noticias" class="text-xs text-blue-500 font-bold">Ver todas</A>
          </div>

          <Suspense fallback={<div class="space-y-4"><div class="h-24 bg-gray-200 animate-pulse rounded-2xl" /></div>}>
            <div class="space-y-4">
              <For each={news()?.data}>
                {(post) => (
                  <div class="bg-white p-4 rounded-3xl border border-colpsi-border shadow-sm flex gap-4 items-center">
                    <Show when={post.image_url} fallback={<div class="w-16 h-16 bg-gray-100 rounded-2xl shrink-0" />}>
                       <img src={bucketUrl(post.image_url)} class="w-16 h-16 object-cover rounded-2xl shrink-0" />
                    </Show>
                    <div class="overflow-hidden">
                      <h4 class="font-bold text-sm text-gray-800 truncate">{post.title}</h4>
                      <p class="text-xs text-colpsi-muted line-clamp-2">{post.short_description}</p>
                    </div>
                  </div>
                )}
              </For>
            </div>
          </Suspense>
        </section>

      </div>

      {/* Tab Bar Inferior (Diseño Mobile-First) */}
      <nav class="md:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-colpsi-border px-6 py-3 flex justify-between items-center z-50">
        <A href="/psi" class="text-colpsi-blue flex flex-col items-center">
          <span class="text-xl">🏠</span>
          <span class="text-[10px] font-bold uppercase">Inicio</span>
        </A>
        <A href="/psi/notificaciones" class="relative text-gray-400 flex flex-col items-center">
          <span class="text-xl">🔔</span>
          <span class="text-[10px] font-bold uppercase">Avisos</span>
          <Show when={(unread()?.unread_count ?? 0) > 0}>
            <span class="absolute -top-1 -right-2 bg-red-500 text-white text-[9px] font-black min-w-[18px] h-[18px] rounded-full flex items-center justify-center px-1">
              {unread()?.unread_count}
            </span>
          </Show>
        </A>
        <A href="/directorio" class="text-gray-400 flex flex-col items-center">
          <span class="text-xl">🔍</span>
          <span class="text-[10px] font-bold uppercase">Buscar</span>
        </A>
        <A href="/psi/perfil" class="text-gray-400 flex flex-col items-center">
          <span class="text-xl">⚙️</span>
          <span class="text-[10px] font-bold uppercase">Ajustes</span>
        </A>
      </nav>
    </main>
  );
}