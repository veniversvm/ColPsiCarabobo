// web/src/components/admin/noticias/NoticiasFilters.tsx
import { Accessor, Setter } from "solid-js";
import { PostStatus } from "./types";

type FilterType = "all" | "public" | "psi";
type FilterStatus = "all" | PostStatus;

interface Props {
  search: Accessor<string>;
  setSearch: Setter<string>;
  filterType: Accessor<FilterType>;
  setFilterType: Setter<FilterType>;
  filterStatus: Accessor<FilterStatus>;
  setFilterStatus: Setter<FilterStatus>;
}

export function NoticiasFilters(props: Props) {
  return (
    <div class="flex flex-col md:flex-row gap-3 mb-6">
      {/* Búsqueda */}
      <div class="relative flex-1">
        <svg class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-4.35-4.35M17 11A6 6 0 1 1 5 11a6 6 0 0 1 12 0z" />
        </svg>
        <input
          type="text"
          placeholder="Buscar por título o resumen..."
          value={props.search()}
          onInput={(e) => props.setSearch(e.currentTarget.value)}
          class="w-full pl-10 pr-4 py-2.5 bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl outline-none text-sm text-gray-800 transition-all"
        />
      </div>

      {/* Filtro tipo */}
      <div class="flex gap-2">
        {(["all", "public", "psi"] as const).map((t) => (
          <button
            onClick={() => props.setFilterType(t)}
            class={`px-4 py-2.5 rounded-xl text-xs font-black uppercase tracking-wide transition-all border-2 ${
              props.filterType() === t
                ? "bg-blue-800 text-white border-blue-800"
                : "bg-white text-gray-500 border-gray-200 hover:border-gray-300"
            }`}
          >
            {t === "all" ? "Todos" : t === "public" ? "Públicos" : "Colegiados"}
          </button>
        ))}
      </div>

      {/* Filtro estado */}
      <div class="flex gap-2 flex-wrap">
        {(["all", "published", "draft", "archived", "scheduled"] as const).map((s) => (
          <button
            onClick={() => props.setFilterStatus(s)}
            class={`px-4 py-2.5 rounded-xl text-xs font-black uppercase tracking-wide transition-all border-2 ${
              props.filterStatus() === s
                ? "bg-blue-800 text-white border-blue-800"
                : "bg-white text-gray-500 border-gray-200 hover:border-gray-300"
            }`}
          >
            {{ all: "Todos", published: "Publicados", draft: "Borradores", archived: "Archivados", scheduled: "Programados" }[s]}
          </button>
        ))}
      </div>
    </div>
  );
}