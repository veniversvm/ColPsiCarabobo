// web/src/components/admin/noticias/NoticiasFilters.tsx
import { Accessor, Setter } from "solid-js";
import { PostStatus } from "./types";
import { Input } from "~/components/admin/ui/Input";
import { Icon } from "~/components/admin/ui/icons";

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

const segClass = (active: boolean) =>
  `h-8 px-3 rounded-md text-xs font-medium transition-all border ${
    active
      ? "bg-white text-colpsi-blue border-colpsi-border shadow-sm"
      : "bg-colpsi-bg text-colpsi-muted border-transparent hover:text-colpsi-blue hover:bg-white/70"
  }`;

export function NoticiasFilters(props: Props) {
  return (
    <div class="flex flex-col lg:flex-row gap-3 lg:items-center">
      {/* Búsqueda */}
      <div class="relative flex-1 min-w-[220px]">
        <Icon name="search" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
        <Input
          type="text"
          placeholder="Buscar por título o resumen..."
          value={props.search()}
          onInput={(e) => props.setSearch(e.currentTarget.value)}
          class="pl-9"
        />
      </div>

      {/* Filtro tipo */}
      <div class="flex gap-1 p-1 rounded-md bg-colpsi-bg border border-colpsi-border">
        {(["all", "public", "psi"] as const).map((t) => (
          <button
            onClick={() => props.setFilterType(t)}
            class={segClass(props.filterType() === t)}
          >
            {t === "all" ? "Todos" : t === "public" ? "Públicos" : "Colegiados"}
          </button>
        ))}
      </div>

      {/* Filtro estado */}
      <div class="flex gap-1 p-1 rounded-md bg-colpsi-bg border border-colpsi-border flex-wrap">
        {(["all", "published", "draft", "archived", "scheduled"] as const).map((s) => (
          <button
            onClick={() => props.setFilterStatus(s)}
            class={segClass(props.filterStatus() === s)}
          >
            {{ all: "Todos", published: "Publicados", draft: "Borradores", archived: "Archivados", scheduled: "Programados" }[s]}
          </button>
        ))}
      </div>
    </div>
  );
}