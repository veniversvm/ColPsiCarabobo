// web/src/components/directory/SearchHeader.tsx
import { For, Show } from "solid-js";

interface Specialty {
  id: number;
  name: string;
}

interface SearchHeaderProps {
  query: string;
  specialty: string;
  location: string;
  specialties: Specialty[] | undefined;
  onQueryChange: (value: string) => void;
  onSpecialtyChange: (value: string) => void;
  onLocationChange: (value: string) => void;
  onSearch: (e: Event) => void;
}

export function SearchHeader(props: SearchHeaderProps) {
  return (
    <section class="bg-colpsi-blue pt-12 pb-24 px-6 text-center relative shadow-2xl">
      <div class="max-w-5xl mx-auto space-y-6 relative z-10">
        <h1 class="text-white text-3xl md:text-5xl font-black tracking-tighter italic">
          DIRECTORIO PROFESIONAL
        </h1>

        <form onSubmit={props.onSearch} class="space-y-4">
          {/* Fila 1: Input de texto y especialidad */}
          <div class="flex flex-col md:flex-row gap-3 max-w-4xl mx-auto">

            {/* Input de texto principal */}
            <div class="relative flex-1">
              <input
                type="text"
                placeholder="Nombre, Cédula o FPV..."
                name="nombre, cédula o fpv"
                value={props.query}
                class="w-full bg-white rounded-2xl py-4 px-6 shadow-xl outline-none focus:ring-4 focus:ring-colpsi-yellow/50 transition-all text-colpsi-text font-medium pr-12"
                onInput={(e) => props.onQueryChange(e.currentTarget.value)}
              />
              <span class="absolute right-4 top-1/2 -translate-y-1/2 text-gray-400">🔍</span>
            </div>

            {/* Select de Especialidad */}
            <select
              value={props.specialty}
              disabled={!props.specialties}
              class="md:w-64 bg-white rounded-2xl py-4 px-6 shadow-xl outline-none focus:ring-4 focus:ring-colpsi-yellow/50 transition-all text-colpsi-text font-bold appearance-none cursor-pointer disabled:opacity-60 disabled:cursor-wait"
              onChange={(e) => props.onSpecialtyChange(e.currentTarget.value)}
              name="especialidad"
            >
              <Show
                when={props.specialties}
                fallback={<option value="">Cargando especialidades...</option>}
              >
                <option value="">Todas las áreas</option>
                <For each={props.specialties}>
                  {(item) => (
                    <option value={String(item.id)}>{item.name}</option>
                  )}
                </For>
              </Show>
            </select>
          </div>

          {/* Fila 2: Input de ubicación y botón */}
          <div class="flex flex-col md:flex-row gap-3 max-w-4xl mx-auto">

            {/* Input de ubicación */}
            <div class="relative flex-1">
              <input
                type="text"
                placeholder="Ubicación (Municipio, Ciudad o Estado)..."
                name="ubicacion"
                value={props.location}
                class="w-full bg-white rounded-2xl py-4 px-6 shadow-xl outline-none focus:ring-4 focus:ring-colpsi-yellow/50 transition-all text-colpsi-text font-medium pr-12"
                onInput={(e) => props.onLocationChange(e.currentTarget.value)}
              />
              <span class="absolute right-4 top-1/2 -translate-y-1/2 text-gray-400">📍</span>
            </div>

            {/* BOTÓN DE ACCIÓN */}
            <button
              type="submit"
              class="md:w-48 bg-colpsi-yellow text-colpsi-blue px-8 py-4 rounded-2xl font-black shadow-lg shadow-yellow-500/20 hover:scale-105 active:scale-95 transition-all whitespace-nowrap flex items-center justify-center gap-2"
            >
              <span>🔍</span>
              BUSCAR
            </button>
          </div>

          {/* Sugerencia */}
          <p class="text-blue-200 text-xs mt-4 text-left md:text-center">
            💡 Puedes buscar por nombre, número de cédula, FPV, especialidad o ubicación
          </p>
        </form>
      </div>
    </section>
  );
}