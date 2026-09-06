// web/src/components/admin/psicologos/PsychologistHeader.tsx
import { A } from "@solidjs/router";

interface Props {
  title?: string;
  onImportClick: () => void;
}

export function PsychologistHeader(props: Props) {
  return (
    <div class="flex flex-col md:flex-row justify-between md:items-end gap-4">
      <div>
        <h1 class="text-2xl font-black text-colpsi-blue">Gestión de Agremiados</h1>
        <p class="text-gray-500 text-sm mt-1">Base de datos maestra de profesionales colegiados.</p>
      </div>
      <div class="flex gap-3">
        <button
          onClick={props.onImportClick}
          class="bg-white border-2 border-gray-200 text-gray-700 px-4 py-2.5 rounded-xl font-bold hover:bg-colpsi-surface transition-colors text-sm"
        >
          📥 Importar CSV
        </button>
        <A
          href="/admin/psicologos/crear"
          class="bg-colpsi-blue text-white px-4 py-2.5 rounded-xl font-bold hover:bg-blue-800 transition-colors text-sm shadow-md"
        >
          ➕ Nuevo Registro
        </A>
      </div>
    </div>
  );
}