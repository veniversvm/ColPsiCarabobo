// web/src/components/admin/psicologos/edit/sections/ContactVisibilitySection.tsx

import { ToggleSwitch } from "~/components/ui/ToggleSwitch";
import { Field, IC2, SectionCard } from "../EditPrimitives";
import type { EditFormState } from "../types";

interface Props {
  form: EditFormState;
  setForm: (key: keyof EditFormState, value: any) => void;
}

export function ContactVisibilitySection(props: Props) {
  return (
    <SectionCard title="Contacto de Consulta & Visibilidad" accent="border-gray-300">
      <p class="text-xs text-gray-500 mb-5 bg-gray-50 p-3 rounded-xl border border-gray-100">
        Estos datos son los que el psicólogo muestra públicamente en el directorio. Puedes controlar su visibilidad con los toggles.
      </p>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-5">

        <div class="space-y-3 bg-gray-50 rounded-2xl p-5 border border-gray-100">
          <Field label="Email de Consulta Público">
            <input type="email"
              value={props.form.contact_email}
              onInput={(e) => props.setForm("contact_email", e.currentTarget.value)}
              class={IC2}
            />
          </Field>
          <ToggleSwitch
            label="Mostrar email de contacto"
            checked={props.form.show_contact_email}
            onChange={(v) => props.setForm("show_contact_email", v)}
          />
        </div>

        <div class="space-y-3 bg-gray-50 rounded-2xl p-5 border border-gray-100">
          <Field label="Teléfono Principal Público">
            <input type="tel"
              value={props.form.public_phone}
              onInput={(e) => props.setForm("public_phone", e.currentTarget.value)}
              class={IC2}
            />
          </Field>
          <ToggleSwitch
            label="Mostrar teléfono público"
            checked={props.form.show_public_phone}
            onChange={(v) => props.setForm("show_public_phone", v)}
          />
        </div>

        <div class="md:col-span-2 space-y-3 bg-gray-50 rounded-2xl p-5 border border-gray-100">
          <Field label="Dirección de Consultorio (Principal)">
            <input type="text"
              value={props.form.service_address}
              onInput={(e) => props.setForm("service_address", e.currentTarget.value)}
              class={IC2}
            />
          </Field>
          <ToggleSwitch
            label="Mostrar dirección de consultorio"
            checked={props.form.show_public_service_address}
            onChange={(v) => props.setForm("show_public_service_address", v)}
          />
        </div>

      </div>
    </SectionCard>
  );
}