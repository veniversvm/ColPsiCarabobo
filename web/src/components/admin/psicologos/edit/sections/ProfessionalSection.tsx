// web/src/components/admin/psicologos/edit/sections/ProfessionalSection.tsx

import { For, Show } from "solid-js";
import { RichTextEditor } from "~/components/ui/RichTextEditor";
import { Field, IC, SectionCard } from "../EditPrimitives";
import type { EditFormState } from "../types";

interface Props {
  form: EditFormState;
  setForm: (key: keyof EditFormState, value: any) => void;
  specialties: any[] | undefined;
}

export function ProfessionalSection(props: Props) {
  return (
    <SectionCard title="Perfil Profesional">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-5">

        <div class="space-y-2">
          <label class="text-[10px] font-black text-gray-500 uppercase tracking-widest ml-1">
            Especialidad Principal
          </label>
          <select
            value={props.form.primary_specialty}
            disabled={!props.specialties}
            class={IC}
            onChange={(e) => props.setForm("primary_specialty", e.currentTarget.value)}
          >
            <Show when={props.specialties} fallback={<option value="">Cargando...</option>}>
              <option value="">— Sin especialidad —</option>
              <For each={props.specialties}>
                {(item: any) => <option value={item.name}>{item.name}</option>}
              </For>
            </Show>
          </select>
        </div>

        <div class="space-y-2">
          <label class="text-[10px] font-black text-gray-500 uppercase tracking-widest ml-1">
            Especialidad Secundaria
          </label>
          <select
            value={props.form.secondary_specialty}
            disabled={!props.specialties}
            class={IC}
            onChange={(e) => props.setForm("secondary_specialty", e.currentTarget.value)}
          >
            <Show when={props.specialties} fallback={<option value="">Cargando...</option>}>
              <option value="">— Sin especialidad —</option>
              <For each={props.specialties}>
                {(item: any) => (
                  <option value={item.name} disabled={item.name === props.form.primary_specialty}>
                    {item.name}
                  </option>
                )}
              </For>
            </Show>
          </select>
        </div>

        <div class="md:col-span-2">
          <Field label="Mini Bio (máx. 500 caracteres)">
            <textarea
              rows={4}
              value={props.form.mini_bio}
              onInput={(e) => props.setForm("mini_bio", e.currentTarget.value)}
              class={`${IC} resize-none`}
              maxLength={500}
            />
            <p class="text-[10px] text-gray-400 mt-1 text-right">
              {(props.form.mini_bio || "").length}/500
            </p>
          </Field>
        </div>

        <div class="md:col-span-2">
          <RichTextEditor
            label="Biografía Extensa (Bio Pública Completa)"
            content={props.form.full_bio}
            onUpdate={(html) => props.setForm("full_bio", html)}
          />
        </div>

      </div>
    </SectionCard>
  );
}