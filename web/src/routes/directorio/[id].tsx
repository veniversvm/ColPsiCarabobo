import { createResource, For, Show, Suspense } from "solid-js";
import { useParams, A } from "@solidjs/router";
import { apiGet } from "~/lib/api";

// Tipado rápido para autocompletado basado en tu DTO de Go
type PsiProfile = {
  id: string;
  first_name: string;
  last_name: string;
  fpv: number;
  gender: string;
  profile_picture: string;
  solvent: boolean;
  email?: string;
  phone?: string;
  address?: string;
  location: { state: string; municipality: string };
  specialties: string[];
  mini_bio: string;
  undergraduate: { university?: string; date?: string; mention?: string };
  post_grades?: { title: string; university: string; year: string }[];
  social_networks?: { name: string; url: string }[];
};

export default function PsiProfilePage() {
  const params = useParams();

  // Fetcher con SSR: Deno hace la petición al backend de Go antes de enviar el HTML
  const [profile] = createResource(() => 
    apiGet<PsiProfile>(`/psi/${params.id}`)
  );

  return (
    <main class="min-h-screen bg-colpsi-bg pb-20 font-sans">
      {/* HEADER DE NAVEGACIÓN */}
      <div class="bg-colpsi-blue py-6 px-4 md:px-8 shadow-md sticky top-0 z-40">
        <div class="max-w-5xl mx-auto flex items-center justify-between">
          <A href="/directorio" class="text-white hover:text-colpsi-yellow font-bold flex items-center gap-2 transition-colors">
            <span>←</span> Volver al Directorio
          </A>
          <span class="text-blue-200 text-sm font-medium tracking-widest uppercase hidden sm:block">
            Ficha Técnica Profesional
          </span>
        </div>
      </div>

      <div class="max-w-5xl mx-auto px-4 md:px-8 mt-8">
        <Suspense fallback={<ProfileSkeleton />}>
          <Show 
            when={profile()} 
            fallback={
              <div class="text-center py-20">
                <h2 class="text-2xl font-black text-colpsi-blue">Perfil no encontrado</h2>
                <p class="text-colpsi-muted mt-2">El profesional no existe o su perfil es privado.</p>
              </div>
            }
          >
            {(psi) => (
              <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
                
                {/* COLUMNA IZQUIERDA: Identidad y Contacto */}
                <div class="space-y-6">
                  {/* Tarjeta Principal */}
                  <div class="bg-white rounded-[2rem] p-8 shadow-premium border border-gray-100 text-center relative overflow-hidden">
                    {/* Badge de Solvencia (Incentivo gremial) */}
                    <Show when={psi().solvent}>
                      <div class="absolute top-4 right-4 bg-green-100 text-green-700 p-2 rounded-full shadow-sm" title="Miembro Solvente">
                        ✓
                      </div>
                    </Show>

                    <div class="w-32 h-32 mx-auto bg-gray-50 rounded-full overflow-hidden border-4 border-colpsi-yellow shadow-inner mb-6">
                      <Show 
                        when={psi().profile_picture} 
                        fallback={<div class="w-full h-full flex items-center justify-center text-5xl bg-blue-50">👤</div>}
                      >
                        <img src={`http://localhost:9000/colpsi-bucket/${psi().profile_picture}`} alt={`Dr(a). ${psi().last_name}`} class="w-full h-full object-cover" />
                      </Show>
                    </div>

                    <h1 class="text-2xl font-black text-colpsi-blue leading-tight">
                      {psi().first_name} {psi().last_name}
                    </h1>
                    <p class="text-colpsi-muted font-bold tracking-widest uppercase mt-1 text-sm">
                      FPV: {psi().fpv}
                    </p>

                    <div class="mt-4 flex justify-center gap-2 flex-wrap">
                      <For each={psi().specialties}>
                        {(spec) => <span class="bg-blue-50 text-colpsi-blue text-xs font-bold px-3 py-1.5 rounded-lg">{spec}</span>}
                      </For>
                    </div>
                  </div>

                  {/* Tarjeta de Contacto (Solo visible si el backend envía los datos) */}
                  <Show when={psi().email || psi().phone || psi().social_networks?.length}>
                    <div class="bg-white rounded-[2rem] p-8 shadow-premium border border-gray-100 space-y-4">
                      <h3 class="text-sm font-black text-colpsi-blue uppercase tracking-widest border-b border-gray-100 pb-2 mb-4">Contacto</h3>
                      
                      <Show when={psi().email}>
                        <div class="flex items-center gap-3 text-sm">
                          <span class="text-colpsi-yellow text-lg">✉️</span>
                          <a href={`mailto:${psi().email}`} class="text-gray-600 hover:text-colpsi-blue transition-colors break-all">{psi().email}</a>
                        </div>
                      </Show>
                      
                      <Show when={psi().phone}>
                        <div class="flex items-center gap-3 text-sm">
                          <span class="text-colpsi-yellow text-lg">📞</span>
                          <a href={`tel:${psi().phone}`} class="text-gray-600 hover:text-colpsi-blue transition-colors">{psi().phone}</a>
                        </div>
                      </Show>

                      <Show when={psi().location.municipality}>
                        <div class="flex items-center gap-3 text-sm">
                          <span class="text-colpsi-yellow text-lg">📍</span>
                          <span class="text-gray-600">{psi().location.municipality}, {psi().location.state}</span>
                        </div>
                      </Show>

                      {/* Redes Sociales */}
                      <Show when={(psi().social_networks?.length ?? 0) > 0}>
                        <div class="pt-4 mt-4 border-t border-gray-100 flex flex-wrap gap-3">
                          <For each={psi().social_networks}>
                            {(net) => (
                              <a href={net.url} target="_blank" rel="noopener noreferrer" class="text-xs bg-gray-50 text-colpsi-blue font-bold px-3 py-2 rounded-lg hover:bg-colpsi-yellow transition-colors">
                                {net.name}
                              </a>
                            )}
                          </For>
                        </div>
                      </Show>
                    </div>
                  </Show>
                </div>

                {/* COLUMNA DERECHA: Bio y Credenciales */}
                <div class="lg:col-span-2 space-y-6">
                  
                  {/* Biografía */}
                  <Show when={psi().mini_bio}>
                    <div class="bg-white rounded-[2rem] p-8 shadow-premium border border-gray-100">
                      <h3 class="text-sm font-black text-colpsi-blue uppercase tracking-widest mb-4">Perfil Profesional</h3>
                      <p class="text-colpsi-text leading-relaxed whitespace-pre-wrap">{psi().mini_bio}</p>
                    </div>
                  </Show>

                  {/* Formación Académica */}
                  <div class="bg-white rounded-[2rem] p-8 shadow-premium border border-gray-100">
                    <h3 class="text-sm font-black text-colpsi-blue uppercase tracking-widest mb-6">Formación Académica</h3>
                    
                    {/* Pregrado (Si el usuario autorizó mostrarlo) */}
                    <Show when={psi().undergraduate?.university}>
                      <div class="relative pl-6 border-l-2 border-colpsi-yellow pb-8">
                        <div class="absolute w-4 h-4 bg-colpsi-yellow rounded-full -left-[9px] top-0 border-4 border-white"></div>
                        <h4 class="font-bold text-gray-900 text-lg">Psicólogo</h4>
                        <p class="text-colpsi-blue font-medium">{psi().undergraduate.university}</p>
                        <div class="text-sm text-colpsi-muted mt-1 flex gap-4">
                          <Show when={psi().undergraduate.date}><span>Egreso: {psi().undergraduate.date}</span></Show>
                          <Show when={psi().undergraduate.mention}><span>Mención: {psi().undergraduate.mention}</span></Show>
                        </div>
                      </div>
                    </Show>

                    {/* Postgrados (Si está solvente y los tiene) */}
                    <For each={psi().post_grades}>
                      {(pg, index) => (
                        <div class={`relative pl-6 border-l-2 border-colpsi-yellow ${index() === (psi().post_grades?.length ?? 0) - 1 ? 'border-transparent' : 'pb-8'}`}>
                          <div class="absolute w-4 h-4 bg-colpsi-blue rounded-full -left-[9px] top-0 border-4 border-white"></div>
                          <h4 class="font-bold text-gray-900 text-lg">{pg.title}</h4>
                          <p class="text-colpsi-blue font-medium">{pg.university}</p>
                          <Show when={pg.year}><p class="text-sm text-colpsi-muted mt-1">Año: {pg.year}</p></Show>
                        </div>
                      )}
                    </For>

                    <Show when={!psi().undergraduate?.university && (!psi().post_grades || psi().post_grades?.length === 0)}>
                      <p class="text-gray-400 italic text-sm">Información académica no configurada públicamente.</p>
                    </Show>
                  </div>

                </div>

              </div>
            )}
          </Show>
        </Suspense>
      </div>
    </main>
  );
}

// Componente "Skeleton" para mostrar mientras Deno procesa la petición a Go
function ProfileSkeleton() {
  return (
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8 animate-pulse">
      <div class="bg-white rounded-[2rem] h-96 shadow-sm border border-gray-100"></div>
      <div class="lg:col-span-2 space-y-6">
        <div class="bg-white rounded-[2rem] h-48 shadow-sm border border-gray-100"></div>
        <div class="bg-white rounded-[2rem] h-64 shadow-sm border border-gray-100"></div>
      </div>
    </div>
  );
}