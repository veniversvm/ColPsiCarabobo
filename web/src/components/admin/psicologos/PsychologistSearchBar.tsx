// web/src/components/admin/psicologos/PsychologistSearchBar.tsx
import { Show } from "solid-js";
import { DropdownSelect } from "~/components/ui/DropdownSelect";
import { Icon } from "~/components/admin/ui/icons";

interface WorkArea {
  id: number;
  name: string;
}

interface Props {
  value: string;
  onInput: (e: Event) => void;
  onClear: () => void;
  loading: boolean;
  placeholder?: string;
  solvent: string;
  onSolventChange: (v: string) => void;
  active: string;
  onActiveChange: (v: string) => void;
  gender: string;
  onGenderChange: (v: string) => void;
  specialty: string;
  onSpecialtyChange: (v: string) => void;
  workAreas: WorkArea[] | undefined;
}

const selectButtonClass =
  "w-full h-9 rounded-md border border-slate-300 bg-white px-3 text-sm font-medium text-colpsi-text hover:bg-colpsi-bg focus:border-colpsi-blue";

export function PsychologistSearchBar(props: Props) {
  return (
    <div class="border border-colpsi-border rounded-lg bg-white p-3 space-y-3">
      <div class="relative">
        <Icon name="search" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
        <input
          type="text"
          value={props.value}
          onInput={props.onInput}
          placeholder={props.placeholder || "Buscar por nombre, apellido, cédula o FPV..."}
          class="h-9 w-full rounded-md border border-slate-300 bg-white pl-9 pr-9 text-sm text-colpsi-text outline-none transition-colors placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15"
        />
        <Show when={props.loading}>
          <div class="absolute right-9 top-1/2 -translate-y-1/2 animate-spin rounded-full h-4 w-4 border-2 border-colpsi-yellow border-t-transparent" />
        </Show>
        <Show when={props.value}>
          <button
            onClick={props.onClear}
            class="absolute right-2.5 top-1/2 -translate-y-1/2 h-6 w-6 rounded-md text-slate-400 hover:text-slate-600 hover:bg-colpsi-bg flex items-center justify-center text-base leading-none"
            title="Limpiar búsqueda"
          >
            ×
          </button>
        </Show>
      </div>

      <div class="grid grid-cols-2 lg:grid-cols-4 gap-2">
        <DropdownSelect
          value={props.solvent}
          onChange={props.onSolventChange}
          placeholder="Solvencia: Todos"
          buttonClass={selectButtonClass}
          options={[
            { value: "", label: "Solvencia: Todos" },
            { value: "1", label: "Solvente" },
            { value: "0", label: "Insolvente" },
          ]}
        />
        <DropdownSelect
          value={props.active}
          onChange={props.onActiveChange}
          placeholder="Estatus: Todos"
          buttonClass={selectButtonClass}
          options={[
            { value: "", label: "Estatus: Todos" },
            { value: "1", label: "Activo" },
            { value: "0", label: "Inactivo" },
          ]}
        />
        <DropdownSelect
          value={props.gender}
          onChange={props.onGenderChange}
          placeholder="Género: Todos"
          buttonClass={selectButtonClass}
          options={[
            { value: "", label: "Género: Todos" },
            { value: "M", label: "Masculino" },
            { value: "F", label: "Femenino" },
          ]}
        />
        <DropdownSelect
          value={props.specialty}
          onChange={props.onSpecialtyChange}
          placeholder="Especialidad: Todas"
          disabled={!props.workAreas}
          loading={!props.workAreas}
          loadingLabel="Cargando áreas..."
          buttonClass={selectButtonClass}
          options={
            props.workAreas
              ? [
                  { value: "", label: "Especialidad: Todas" },
                  ...props.workAreas.map((a) => ({ value: String(a.id), label: a.name })),
                ]
              : []
          }
        />
      </div>
    </div>
  );
}