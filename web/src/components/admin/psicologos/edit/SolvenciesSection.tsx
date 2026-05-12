// web/src/components/admin/psicologos/edit/SolvenciesSection.tsx
import { createSignal, For, Show } from "solid-js";

interface SolvenciesSectionProps {
  solvencies: any[];
  onAddLocalSolvency: (year: number) => void;
}

export const SolvenciesSection = (props: SolvenciesSectionProps) => {
  const [newYear, setNewYear] = createSignal<string>(new Date().getFullYear().toString());
  const currentYear = new Date().getFullYear()

  const handleAdd = () => {
    const yearNum = parseInt(newYear(), 10);
    if (isNaN(yearNum) || yearNum < 2024 || yearNum > currentYear) {
      alert("Por favor ingrese un año válido.");
      return;
    }
    
    // Disparamos la adición local
    props.onAddLocalSolvency(yearNum);
    
    // Preparamos el input para el siguiente año
    setNewYear((yearNum + 1).toString());
  };

  return (
    <section class="bg-white p-8 rounded-3xl shadow-sm border border-slate-100 mt-8">
      <details class="group">
        <summary class="flex items-center justify-between cursor-pointer list-none">
          <div class="flex items-center gap-4">
            <div class="bg-emerald-100 p-3 rounded-2xl">
              <span class="text-2xl">📜</span>
            </div>
            <div>
              <h2 class="text-xl font-black text-slate-800 uppercase">Historial de Solvencias</h2>
              <p class="text-sm text-slate-500">Añada años para incluirlos al guardar el expediente</p>
            </div>
          </div>
          <span class="text-slate-400 group-open:rotate-180 transition-transform duration-300">▼</span>
        </summary>

        <div class="mt-6 pt-6 border-t border-slate-100">
          
          {/* Formulario rápido */}
          <div class="flex items-end gap-4 mb-6 bg-slate-50 p-4 rounded-xl">
            <div class="flex-1 max-w-xs">
              <label class="block text-xs font-bold text-slate-500 mb-1">AÑADIR AÑO</label>
              <input 
                type="number" 
                value={newYear()}
                onInput={(e) => setNewYear(e.currentTarget.value)}
                placeholder="Ej. 2026"
                class="w-full p-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-emerald-500 outline-none"
              />
            </div>
            <button 
              type="button"
              onClick={handleAdd}
              class="bg-slate-800 text-white px-4 py-2 rounded-lg font-bold text-sm hover:bg-slate-700 active:scale-95 transition-all"
            >
              + Añadir a la lista
            </button>
          </div>

          {/* Tabla de registros locales */}
          <div class="overflow-hidden rounded-xl border border-slate-200">
            <table class="w-full text-left border-collapse">
              <thead class="bg-slate-100">
                <tr>
                  <th class="p-3 text-xs font-bold text-slate-600 uppercase">Año de Corte</th>
                  <th class="p-3 text-xs font-bold text-slate-600 uppercase text-center">Estado</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 bg-white">
                <Show when={props.solvencies?.length > 0} fallback={
                  <tr><td colspan="2" class="p-6 text-center text-slate-400 italic">No hay años en la lista.</td></tr>
                }>
                  {/* Ordenamos de mayor a menor localmente para mostrar el más reciente arriba */}
                  <For each={[...props.solvencies].sort((a, b) => new Date(b.date || b.Date).getTime() - new Date(a.date || a.Date).getTime())}>
                    {(s) => {
                      const year = new Date(s.date || s.Date).getFullYear();
                      return (
                        <tr class="hover:bg-slate-50">
                          <td class="p-3 font-mono text-sm text-slate-700">
                            {year} <span class="text-xs text-slate-400 ml-2">(31/12)</span>
                          </td>
                          <td class="p-3 text-center">
                            <span class={`px-2 py-1 rounded text-xs font-bold ${
                              s.id ? "bg-emerald-100 text-emerald-700" : "bg-amber-100 text-amber-700"
                            }`}>
                              {s.id ? "REGISTRADA" : "PENDIENTE POR GUARDAR"}
                            </span>
                          </td>
                        </tr>
                      );
                    }}
                  </For>
                </Show>
              </tbody>
            </table>
          </div>
          
        </div>
      </details>
    </section>
  );
};