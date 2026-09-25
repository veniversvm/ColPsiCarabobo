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
      <table class="w-full border-collapse">
        <thead>
          <tr>
            <th class="th-cell">Nº Control</th>
            <th class="th-cell">Agremiado</th>
            <th class="th-cell">Credenciales</th>
            <th class="th-cell">Solvencia</th>
            <th class="th-cell">Estatus</th>
            <th class="th-cell text-right">Acciones</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-colpsi-border">
          <Suspense fallback={
            <tr>
              <td colSpan="6" class="p-8 text-center text-slate-400 font-medium animate-pulse">
                Cargando base de datos...
              </td>
            </tr>
          }>
            <For
              each={props.data}
              fallback={
                <tr>
                  <td colSpan="6" class="py-16 text-center text-colpsi-muted font-medium">
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