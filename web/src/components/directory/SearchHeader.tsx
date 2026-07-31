// web/src/components/directory/SearchHeader.tsx
import { DropdownSelect } from "~/components/ui/DropdownSelect";

interface WorkArea {
  id: number;
  name: string;
}

interface SearchHeaderProps {
  query: string;
  workArea: string;              // Cambiado de specialty
  location: string;
  workAreas: WorkArea[] | undefined; // Cambiado de specialties
  onQueryChange: (value: string) => void;
  onWorkAreaChange: (value: string) => void; // Cambiado de onSpecialtyChange
  onLocationChange: (value: string) => void;
  onSearch: (e: Event) => void;
}

export function SearchHeader(props: SearchHeaderProps) {
  return (
    <section class="bg-colpsi-blue pt-12 pb-24 px-6 text-center relative shadow-2xl">
      <div class="max-w-5xl mx-auto space-y-6 relative z-10">
        <h1 class="text-white text-3xl md:text-5xl font-black tracking-tighter italic uppercase">
          Directorio Profesional
        </h1>

        <form onSubmit={props.onSearch} class="space-y-4">
          {/* Fila 1: Búsqueda por texto y Área de Desempeño */}
          <div class="flex flex-col md:flex-row gap-3 max-w-4xl mx-auto">

            {/* Input de Identidad */}
            <div class="relative flex-1">
              <input
                type="text"
                placeholder="Nombre, Cédula o FPV..."
                name="search_query"
                value={props.query}
                class="w-full bg-white rounded-2xl py-4 px-6 shadow-xl outline-none focus:ring-4 focus:ring-colpsi-yellow/50 transition-all text-colpsi-text font-medium pr-12"
                onInput={(e) => props.onQueryChange(e.currentTarget.value)}
              />
              <span class="absolute right-4 top-1/2 -translate-y-1/2 text-gray-400">🔍</span>
            </div>

            {/* Selector de Área de Desempeño */}
            <div class="relative md:w-80">
              <DropdownSelect
                value={props.workArea}
                disabled={!props.workAreas}
                loading={!props.workAreas}
                loadingLabel="Cargando áreas..."
                placeholder="Todas las Áreas de Desempeño"
                buttonClass="w-full bg-white rounded-2xl py-4 px-6 shadow-xl focus:ring-4 focus:ring-colpsi-yellow/50 text-colpsi-text font-bold"
                options={
                  props.workAreas
                    ? [
                        { value: "", label: "Todas las Áreas de Desempeño" },
                        ...props.workAreas.map((item) => ({
                          value: String(item.id),
                          label: item.name,
                        })),
                      ]
                    : []
                }
                onChange={props.onWorkAreaChange}
              />
            </div>
          </div>

          {/* Fila 2: Ubicación y Botón de Acción */}
          <div class="flex flex-col md:flex-row gap-3 max-w-4xl mx-auto">

            {/* Input de Ubicación */}
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

            {/* Botón Buscar */}
            <button
              type="submit"
              class="md:w-48 bg-colpsi-yellow text-colpsi-blue px-8 py-4 rounded-2xl font-black shadow-lg shadow-yellow-500/20 hover:bg-[#f3ca05] hover:scale-105 active:scale-95 transition-all whitespace-nowrap flex items-center justify-center gap-2"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
              BUSCAR
            </button>
          </div>

          {/* Texto de Ayuda */}
          <p class="text-blue-200 text-xs mt-4 text-left md:text-center font-medium">
            💡 Tip: Filtra por <span class="text-white font-bold">Área de Desempeño</span> para encontrar especialistas según tus necesidades.
          </p>
        </form>
      </div>
    </section>
  );
}