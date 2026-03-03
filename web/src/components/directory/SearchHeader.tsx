// web/src/components/directory/SearchHeader.tsx
import { For, Suspense } from "solid-js";

interface SearchHeaderProps {
  query: string;
  specialty: string;
  specialties: any[] | undefined;
  onQueryChange: (value: string) => void;
  onSpecialtyChange: (value: string) => void;
  onSearch: (e: Event) => void;
}

export function SearchHeader(props: SearchHeaderProps) {
  return (
    <section class="bg-colpsi-blue pt-12 pb-24 px-6 text-center relative shadow-2xl">
      <div class="max-w-4xl mx-auto space-y-6 relative z-10">
        <h1 class="text-white text-3xl md:text-5xl font-black tracking-tighter italic">
          DIRECTORIO PROFESIONAL
        </h1>
        
        <form onSubmit={props.onSearch} class="space-y-4">
          <div class="flex flex-col md:flex-row gap-3 max-w-3xl mx-auto">
            {/* Input de texto */}
            <div class="relative flex-grow">
              <input
                type="text"
                placeholder="Nombre, Cédula o FPV..."
                value={props.query}
                class="w-full bg-white rounded-2xl py-4 px-6 shadow-xl outline-none focus:ring-4 focus:ring-colpsi-yellow/50 transition-all text-colpsi-text font-medium"
                onInput={(e) => props.onQueryChange(e.currentTarget.value)}
              />
              <span class="absolute right-5 top-4 opacity-30">🔍</span>
            </div>
            
            {/* Select de Especialidad */}
            <select 
              value={props.specialty}
              class="bg-white rounded-2xl py-4 px-6 shadow-xl outline-none focus:ring-4 focus:ring-colpsi-yellow/50 transition-all text-colpsi-text font-bold appearance-none cursor-pointer"
              onChange={(e) => props.onSpecialtyChange(e.currentTarget.value)}
            >
              <option value="">Todas las áreas</option>
              <Suspense fallback={<option>Cargando...</option>}>
                <For each={props.specialties}>
                  {(item) => <option value={item.name}>{item.name}</option>}
                </For>
              </Suspense>
            </select>

            {/* BOTÓN DE ACCIÓN */}
            <button 
              type="submit"
              class="bg-colpsi-yellow text-colpsi-blue px-8 py-4 rounded-2xl font-black shadow-lg shadow-yellow-500/20 hover:scale-105 active:scale-95 transition-all whitespace-nowrap"
            >
              BUSCAR PROFESIONAL
            </button>
          </div>
        </form>
      </div>
    </section>
  );
}