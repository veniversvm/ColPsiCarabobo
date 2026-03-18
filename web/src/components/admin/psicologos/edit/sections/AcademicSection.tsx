// web/src/components/admin/psicologos/edit/sections/AcademicSection.tsx

import { ToggleSwitch } from "~/components/ui/ToggleSwitch";
import { Field, IC, SectionCard } from "../EditPrimitives";
import type { EditFormState } from "../types";

interface Props {
  form: EditFormState;
  setForm: (key: keyof EditFormState, value: any) => void;
}

export function AcademicSection(props: Props) {
  return (
    <SectionCard title="Registro de Título (Pregrado)" accent="border-colpsi-yellow">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-5">

        <div class="md:col-span-2">
          <Field label="Universidad de Egreso">
            <input type="text"
              value={props.form.university_undergraduate}
              onInput={(e) => props.setForm("university_undergraduate", e.currentTarget.value)}
              class={IC}
            />
          </Field>
        </div>

        <Field label="Fecha de Egreso">
          <input type="date"
            value={props.form.graduate_date}
            onInput={(e) => props.setForm("graduate_date", e.currentTarget.value)}
            class={IC}
          />
        </Field>

        <div class="md:col-span-3">
          <Field label="Mención">
            <input type="text"
              value={props.form.mention_undergraduate}
              onInput={(e) => props.setForm("mention_undergraduate", e.currentTarget.value)}
              class={IC}
            />
          </Field>
        </div>

        {/* Datos de Registro de Título en Estado */}
        <div class="col-span-full border-t border-gray-100 pt-4">
          <p class="text-[10px] font-black text-gray-400 uppercase tracking-widest mb-4">
            Datos de Registro de Título en Estado
          </p>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
            <Field label="Nro. Registro">
              <input type="number"
                value={props.form.register_number}
                onInput={(e) => props.setForm("register_number", e.currentTarget.value)}
                class={IC}
              />
            </Field>
            <Field label="Estado de Registro">
              <input type="text"
                value={props.form.register_title_state}
                onInput={(e) => props.setForm("register_title_state", e.currentTarget.value)}
                class={IC}
              />
            </Field>
            <Field label="Fecha de Registro">
              <input type="date"
                value={props.form.register_title_date}
                onInput={(e) => props.setForm("register_title_date", e.currentTarget.value)}
                class={IC}
              />
            </Field>
            <Field label="Folio">
              <input type="text"
                value={props.form.register_folio}
                onInput={(e) => props.setForm("register_folio", e.currentTarget.value)}
                class={IC}
              />
            </Field>
            <Field label="Tomo">
              <input type="text"
                value={props.form.register_tome}
                onInput={(e) => props.setForm("register_tome", e.currentTarget.value)}
                class={IC}
              />
            </Field>
          </div>
        </div>

      </div>

      {/* Privacidad académica */}
      <div class="mt-6 bg-gray-50 p-5 rounded-2xl border border-gray-100">
        <p class="text-[10px] font-black text-gray-400 uppercase tracking-widest mb-3">
          Visibilidad en Directorio Público
        </p>
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <ToggleSwitch
            label="Mostrar Universidad"
            checked={props.form.show_university_undergraduate}
            onChange={(v) => props.setForm("show_university_undergraduate", v)}
          />
          <ToggleSwitch
            label="Mostrar Fecha de Egreso"
            checked={props.form.show_graduate_date}
            onChange={(v) => props.setForm("show_graduate_date", v)}
          />
          <ToggleSwitch
            label="Mostrar Mención de Grado"
            checked={props.form.show_mention_undergraduate}
            onChange={(v) => props.setForm("show_mention_undergraduate", v)}
          />
        </div>
      </div>
    </SectionCard>
  );
}