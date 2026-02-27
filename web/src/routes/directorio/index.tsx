// web/src/routes/directorio/index.tsx

import { createResource, createSignal, For, Show, Suspense } from "solid-js";
import { apiGet } from "~/lib/api";
import { A } from "@solidjs/router";

export default function DirectoryPage() {
  // Estados para los filtros
  const [query, setQuery] = createSignal("");
  const [specialty, setSpecialty] = createSignal("");

  // 1. Cargamos el catálogo de especialidades para el buscador
  const [specialties] = createResource(() => apiGet<any[]>("/specialties"));

  // 2. Cargamos los psicólogos basándonos en los filtros (Reactividad pura)
  const [data] = createResource(
    () => ({ q: query(), spec: specialty() }),
    async (filter) => {
      return await apiGet<any>(`/psi/directory?q=${filter.q}&specialty=${filter.spec}&limit=12`);
    }
  );

  return (
    <main class="min-h-screen bg-[#fcfcfc] pb-24">
      {/* HEADER DE BÚSQUEDA */}
      <section class="bg-colpsi-blue pt-12 pb-24 px-6 text-center relative">
        <div class="max-w-4xl mx-auto space-y-6 relative z-10">
          <h1 class="text-white text-3xl md:text-5xl font-black tracking-tighter">
            DIRECTORIO PROFESIONAL
          </h1>
          <p class="text-blue-200 font-medium">Busque por nombre, cédula o especialidad</p>

          {/* CAJA DE BÚSQUEDA MOBILE-FIRST */}
          <div class="flex flex-col md:flex-row gap-3 max-w-3xl mx-auto">
            <div class="relative flex-grow">
              <input
                type="text"
                placeholder="Nombre, CI o FPV..."
                class="w-full bg-white rounded-2xl py-4 px-6 shadow-xl outline-none focus:ring-4 focus:ring-colpsi-yellow/50 transition-all text-colpsi-text"
                onInput={(e) => setQuery(e.currentTarget.value)}
              />
              <span class="absolute right-5 top-4 opacity-30">🔍</span>
            </div>
            
            {/* Filtro de Especialidad */}
            <select 
              class="bg-white rounded-2xl py-4 px-6 shadow-xl outline-none focus:ring-4 focus:ring-colpsi-yellow/50 transition-all text-colpsi-text font-bold appearance-none cursor-pointer"
              onChange={(e) => setSpecialty(e.currentTarget.value)}
            >
              <option value="">Todas las especialidades</option>
              <Suspense fallback={<option>Cargando...</option>}>
                <For each={specialties()}>
                  {(item) => <option value={item.name}>{item.name}</option>}
                </For>
              </Suspense>
            </select>
          </div>
        </div>
      </section>

      {/* RESULTADOS */}
      <div class="max-w-7xl mx-auto px-6 -mt-10 relative z-20">
        <Suspense fallback={
          <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <For each={[1, 2, 3, 4, 5, 6]}>
              {() => <div class="h-48 bg-white rounded-[2.5rem] animate-pulse border border-gray-100" />}
            </For>
          </div>
        }>
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <For each={data()?.data}>
              {(psi) => (
                <A 
                  href={`/directorio/${psi.id}`} 
                  class="bg-white border-2 border-transparent hover:border-colpsi-yellow p-6 rounded-[2.5rem] shadow-sm hover:shadow-2xl hover:-translate-y-1 transition-all group flex flex-col"
                >
                  <div class="flex items-center gap-4 mb-4">
                    <div class="w-16 h-16 bg-gray-50 rounded-2xl overflow-hidden flex-shrink-0 border-2 border-gray-50">
                      <Show when={psi.profile_picture} fallback={<div class="flex h-full items-center justify-center text-3xl">👤</div>}>
                        <img src={`http://localhost:9000/colpsi-bucket/${psi.profile_picture}`} class="w-full h-full object-cover" />
                      </Show>
                    </div>
                    <div>
                      <h3 class="text-colpsi-blue font-black leading-tight group-hover:underline">
                        {psi.first_name} {psi.last_name}
                      </h3>
                      <p class="text-[10px] font-black text-gray-400 uppercase tracking-widest">FPV: {psi.fpv}</p>
                    </div>
                  </div>

                  <p class="text-colpsi-text text-sm leading-relaxed line-clamp-3 italic mb-4">
                    "{psi.mini_bio || 'Profesional comprometido con la salud emocional en Carabobo.'}"
                  </p>

                  <div class="mt-auto flex flex-wrap gap-2">
                    <For each={psi.specialties}>
                      {(s) => (
                        <span class="text-[10px] bg-blue-50 text-colpsi-blue font-bold px-3 py-1 rounded-full uppercase">
                          {s}
                        </span>
                      )}
                    </For>
                  </div>
                </A>
              )}
            </For>
          </div>

          {/* Mensaje de No Resultados */}
          <Show when={data()?.data?.length === 0}>
            <div class="bg-white rounded-[2.5rem] p-20 text-center border-2 border-dashed border-gray-200">
               <span class="text-5xl mb-4 block">🔍</span>
               <h4 class="text-colpsi-blue font-black uppercase">No se encontraron resultados</h4>
               <p class="text-colpsi-muted text-sm mt-2">Intente con otros términos o especialidades.</p>
            </div>
          </Show>
        </Suspense>
      </div>

      {/* RE-UTILIZAMOS LA FRANJA DE LA BANDERA */}
      <div class="fixed bottom-0 left-0 w-full h-2 flex overflow-hidden z-50">
        <div class="flex-1 bg-colpsi-red"></div>
        <div class="flex-1 bg-green-700"></div>
        <div class="flex-1 bg-colpsi-blue"></div>
      </div>
    </main>
  );
}