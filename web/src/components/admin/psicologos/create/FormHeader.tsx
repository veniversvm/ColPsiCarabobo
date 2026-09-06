// web/src/components/admin/psicologos/create/FormHeader.tsx
import { A } from "@solidjs/router";

export function FormHeader() {
  return (
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-black text-colpsi-blue">Alta de Colegiado</h1>
        <p class="text-gray-500 text-sm mt-1">Apertura de nuevo expediente institucional (* Campos obligatorios)</p>
      </div>
      <A href="/admin/psicologos" class="bg-white border-2 border-gray-200 text-gray-700 px-4 py-2.5 rounded-xl font-bold hover:bg-colpsi-surface transition-colors text-sm">
        Cancelar
      </A>
    </div>
  );
}