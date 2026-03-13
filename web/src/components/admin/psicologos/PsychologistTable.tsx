// web/src/components/admin/psicologos/PsychologistTable.tsx
import { For, Show, Suspense } from "solid-js";
import { PsiAdminListItem } from "~/types/admin";
import { PsychologistTableRow } from "./PsychologistTableRow";

interface Props {
  data: PsiAdminListItem[] | undefined;
  loading: boolean;
  hasQuery: boolean;
  query: string;
}

export function PsychologistTable(props: Props) {
  return (
    <div class="overflow-x-auto">
      <table class="w-full text-left border-collapse">
        <thead>
          <tr class="bg-gray-50/50 border-b border-gray-100">
            <th class="px-6 py-4 text-xs font-black text-gray-400 uppercase tracking-widest">Agremiado</th>
            <th class="px-6 py-4 text-xs font-black text-gray-400 uppercase tracking-widest">Credenciales</th>
            <th class="px-6 py-4 text-xs font-black text-gray-400 uppercase tracking-widest">Solvencia</th>
            <th class="px-6 py-4 text-xs font-black text-gray-400 uppercase tracking-widest">Estatus</th>
            <th class="px-6 py-4 text-xs font-black text-gray-400 uppercase tracking-widest text-right">Acciones</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          <Suspense fallback={
            <tr>
              <td colSpan="5" class="p-8 text-center text-gray-400 font-medium animate-pulse">
                Cargando base de datos...
              </td>
            </tr>
          }>
            <For
              each={props.data}
              fallback={
                <tr>
                  <td colSpan="5" class="p-20 text-center text-gray-500 font-medium">
                    {props.hasQuery
                      ? `Sin resultados para "${props.query}"`
                      : "No hay registros en la base de datos."}
                  </td>
                </tr>
              }
            >
              {(psi) => <PsychologistTableRow psi={psi} />}
            </For>
          </Suspense>
        </tbody>
      </table>
    </div>
  );
}