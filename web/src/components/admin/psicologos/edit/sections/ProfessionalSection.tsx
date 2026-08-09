// web/src/components/admin/psicologos/edit/sections/ProfessionalSection.tsx

import { For, Show } from "solid-js";
import { RichTextEditor } from "~/components/ui/RichTextEditor";
import { Field, IC } from "../EditPrimitives";
import type { EditFormState } from "../types";

interface Props {
  form: EditFormState;
  setForm: (key: keyof EditFormState, value: any) => void;
  workAreas: any[] | undefined;
}

export function ProfessionalSection(props: Props) {
  // Función auxiliar para obtener la longitud de forma segura
  const getBioLength = () => props.form.mini_bio?.length || 0;

  return (
    <div class="grid grid-cols-1 md:grid-cols-2 gap-5">

        {/* Área de Desempeño 1 */}
        <div class="space-y-2">
          <label class="text-[10px] font-black text-gray-500 uppercase tracking-widest ml-1">
            Área de Desempeño
          </label>
          <select
            value={props.form.primary_work_area || ""} // Safe fallback
            disabled={!props.workAreas}
            class={IC}
            onChange={(e) => props.setForm("primary_work_area", e.currentTarget.value)}
          >
            <Show when={props.workAreas} fallback={<option value="">Cargando áreas...</option>}>
              <option value="">— Sin área asignada —</option>
              <For each={props.workAreas}>
                {(item: any) => <option value={item.name}>{item.name}</option>}
              </For>
            </Show>
          </select>
        </div>

        {/* Área de Desempeño 2 */}
        <div class="space-y-2">
          <label class="text-[10px] font-black text-gray-400 uppercase tracking-widest ml-1">
            Área de Desempeño (Adicional)
          </label>
          <select
            value={props.form.secondary_work_area || ""} // Safe fallback
            disabled={!props.workAreas}
            class={IC}
            onChange={(e) => props.setForm("secondary_work_area", e.currentTarget.value)}
          >
            <Show when={props.workAreas} fallback={<option value="">Cargando áreas...</option>}>
              <option value="">— Sin área asignada —</option>
              <For each={props.workAreas}>
                {(item: any) => (
                  <option 
                    value={item.name} 
                    disabled={item.name === props.form.primary_work_area}
                  >
                    {item.name}
                  </option>
                )}
              </For>
            </Show>
          </select>
        </div>

        {/* Mini Bio con corrección de ERROR de .length */}
        <div class="md:col-span-2">
          <Field label="Resumen Profesional / Mini Bio (máx. 250 caracteres)">
            <textarea
              rows={3}
              value={props.form.mini_bio || ""} // Safe fallback
              onInput={(e) => props.setForm("mini_bio", e.currentTarget.value)}
              class={`${IC} resize-none`}
              maxLength={250}
            />
            <div class="flex justify-between mt-1 px-1">
               <p class="text-[9px] text-gray-400 italic">Este resumen aparece en las tarjetas del directorio público.</p>
               {/* 
                  FIX: Usamos getBioLength() que ya tiene el fallback || 0 
                  para evitar el error "Cannot read properties of undefined"
               */}
               <p class={`text-[10px] font-bold ${getBioLength() >= 250 ? 'text-red-500' : 'text-gray-400'}`}>
                 {getBioLength()}/250
               </p>
            </div>
          </Field>
        </div>

        {/* Full Bio */}
        <div class="md:col-span-2">
          <RichTextEditor
            label="Biografía Detallada (Contenido HTML)"
            content={props.form.full_bio || ""} // Safe fallback
            onUpdate={(html) => props.setForm("full_bio", html)}
          />
        </div>

      </div>
  );
}