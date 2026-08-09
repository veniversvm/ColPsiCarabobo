// web/src/components/admin/psicologos/edit/SolvenciesSection.tsx
import { createSignal, For, Show } from "solid-js";

interface SolvenciesSectionProps {
  solvencies: any[];
  onAddLocalSolvency: (year: number) => void;
}

export const SolvenciesSection = (props: SolvenciesSectionProps) => {
  const [newYear, setNewYear] = createSignal<string>(new Date().getFullYear().toString());
  const currentYear = new Date().getFullYear();

  const handleAdd = () => {
    const yearNum = parseInt(newYear(), 10);
    
    // 1. Validaciones de rango (según tu requerimiento de negocio)
    if (isNaN(yearNum) || yearNum < 2024 || yearNum > currentYear) {
      alert(`Ingrese un año entre 2024 y ${currentYear}.`);
      return;
    }

    // 2. VALIDACIÓN DE DUPLICADOS:
    // Buscamos en el array usando la llave "date" que es la que viene de la API
    const isDuplicate = props.solvencies?.some(s => {
      const d = s.date || s.Date; // Compatibilidad con posibles variaciones de casing
      return d && new Date(d).getUTCFullYear() === yearNum;
    });

    if (isDuplicate) {
      alert(`El año ${yearNum} ya está registrado.`);
      return;
    }
    
    props.onAddLocalSolvency(yearNum);
    
    // Sugerir el siguiente año si no hemos llegado al actual
    if (yearNum < currentYear) {
      setNewYear((yearNum + 1).toString());
    }
  };

  return (
    <div>

      {/* Input de adición */}
          <div class="flex items-end gap-4 mb-8 bg-slate-50 p-5 rounded-2xl border border-slate-100">
            <div class="flex-1 max-w-[200px]">
              <label class="block text-[10px] font-black text-slate-400 mb-1.5 uppercase tracking-widest">Año fiscal</label>
              <input 
                type="number" 
                value={newYear()}
                onInput={(e) => setNewYear(e.currentTarget.value)}
                min="2024"
                max={currentYear}
                class="w-full bg-white p-3 border border-slate-200 rounded-xl focus:ring-4 focus:ring-emerald-500/10 focus:border-emerald-500 outline-none font-mono text-lg transition-all"
              />
            </div>
            <button 
              type="button"
              onClick={handleAdd}
              class="bg-slate-900 text-white px-6 py-3.5 rounded-xl font-bold text-sm hover:bg-emerald-600 active:scale-95 transition-all shadow-lg shadow-slate-200 flex items-center gap-2"
            >
              <span>+</span> Añadir Periodo
            </button>
          </div>

          <div class="overflow-hidden rounded-2xl border border-slate-100 shadow-sm">
            <table class="w-full text-left border-collapse">
              <thead class="bg-slate-50/50">
                <tr>
                  <th class="p-4 text-[11px] font-black text-slate-500 uppercase tracking-widest">Año de Solvencia</th>
                  <th class="p-4 text-[11px] font-black text-slate-500 uppercase text-center tracking-widest">Estado API</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-50">
                <Show when={props.solvencies?.length > 0} fallback={
                  <tr><td colspan="2" class="p-10 text-center text-slate-400 italic text-sm">No hay registros pendientes ni guardados.</td></tr>
                }>
                  <For each={[...props.solvencies].sort((a, b) => 
                    new Date(b.date || b.Date).getTime() - new Date(a.date || a.Date).getTime()
                  )}>
                    {(s) => {
                      const dateVal = s.date || s.Date;
                      const year = new Date(dateVal).getUTCFullYear();
                      // s.id existe si viene de la DB (UUID), si es nueva estará undefined
                      const isSaved = !!(s.id || s.ID);

                      return (
                        <tr class="hover:bg-slate-50/80 transition-colors group">
                          <td class="p-4 font-mono text-base text-slate-700">
                            <span class="font-black text-slate-900">{year}</span>
                            <span class="text-[10px] text-slate-400 ml-3 uppercase font-sans font-bold tracking-tighter opacity-0 group-hover:opacity-100 transition-opacity">Cierre 31 Dic</span>
                          </td>
                          <td class="p-4 text-center">
                            <span class={`inline-flex items-center px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-tighter border ${
                              isSaved 
                                ? "bg-emerald-50 text-emerald-600 border-emerald-100" 
                                : "bg-amber-50 text-amber-600 border-amber-100 animate-pulse"
                            }`}>
                              {isSaved ? "● Sincronizado" : "○ Pendiente de Envío"}
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
  );
};