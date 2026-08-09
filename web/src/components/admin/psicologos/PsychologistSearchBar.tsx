// web/src/components/admin/psicologos/PsychologistSearchBar.tsx
import { Show } from "solid-js";
import { DropdownSelect } from "~/components/ui/DropdownSelect";

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
  "w-full bg-gray-50 rounded-xl px-4 py-2.5 text-sm text-colpsi-text font-bold border border-gray-100 hover:bg-gray-100";

export function PsychologistSearchBar(props: Props) {
  return (
    <div class="bg-white p-4 rounded-2xl shadow-premium border-2 border-colpsi-blue/20 space-y-3">
      <div class="flex items-center gap-2 bg-gray-50 border-2 border-gray-200 rounded-xl px-3 py-2.5 focus-within:border-colpsi-blue focus-within:ring-2 focus-within:ring-colpsi-blue/30 transition-all">
        <span class="text-xl">🔍</span>
        <input
          type="text"
          value={props.value}
          onInput={props.onInput}
          placeholder={props.placeholder || "Buscar por nombre, apellido, cédula o FPV..."}
          class="flex-grow bg-transparent outline-none text-colpsi-text font-medium"
        />
        <Show when={props.loading}>
          <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-colpsi-yellow mr-2 flex-shrink-0" />
        </Show>
        <Show when={props.value}>
          <button
            onClick={props.onClear}
            class="text-gray-400 hover:text-gray-600 font-black text-lg leading-none flex-shrink-0 mr-1"
            title="Limpiar búsqueda"
          >
            ×
          </button>
        </Show>
      </div>

      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
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
