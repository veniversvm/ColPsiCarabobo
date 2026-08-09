// web/src/components/admin/psicologos/edit/sections/ContactVisibilitySection.tsx

import { ToggleSwitch } from "~/components/ui/ToggleSwitch";
import { Field, IC } from "../EditPrimitives";
import type { EditFormState } from "../types";

interface Props {
  form: EditFormState;
  setForm: (key: keyof EditFormState, value: any) => void;
}

export function ContactVisibilitySection(props: Props) {
  return (
    <div class="space-y-8">
        
        {/* ── SECCIÓN A: CONTACTO INTERNO (GREMIO) ── */}
        <div>
          <div class="flex items-center gap-2 mb-4 border-l-4 border-blue-500 pl-3">
            <h3 class="text-sm font-black text-blue-900 uppercase tracking-widest">
              Información de Contacto Interno
            </h3>
            <span class="text-[9px] bg-blue-100 text-blue-700 px-2 py-0.5 rounded-md font-bold uppercase">Uso Privado</span>
          </div>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-5 bg-gray-50/50 p-6 rounded-2xl border border-gray-100">
            <Field label="Teléfono Local (Gremio)">
              <input type="tel"
                value={props.form.contact_phone || ""}
                onInput={(e) => props.setForm("contact_phone", e.currentTarget.value)}
                class={IC}
                placeholder="Ej: 0241-0000000"
              />
            </Field>
            <Field label="Teléfono Celular (Gremio)">
              <input type="tel"
                value={props.form.contact_cell_phone || ""}
                onInput={(e) => props.setForm("contact_cell_phone", e.currentTarget.value)}
                class={IC}
                placeholder="Ej: 0414-0000000"
              />
            </Field>
            <p class="md:col-span-2 text-[10px] text-gray-400 italic">
              * Estos números son para comunicación exclusiva entre el Colegio y el agremiado. No se muestran en el directorio público.
            </p>
          </div>
        </div>

        {/* ── SECCIÓN B: CONTACTO DE CONSULTA (PÚBLICO) ── */}
        <div class="pt-2">
          <div class="flex items-center gap-2 mb-4 border-l-4 border-emerald-500 pl-3">
            <h3 class="text-sm font-black text-blue-900 uppercase tracking-widest">
              Datos de Consulta y Visibilidad Pública
            </h3>
            <span class="text-[9px] bg-emerald-100 text-emerald-700 px-2 py-0.5 rounded-md font-bold uppercase">Directorio</span>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Email de Consulta */}
            <div class="space-y-4">
              <Field label="Email de Contacto Público">
                <input type="email"
                  value={props.form.contact_email || ""}
                  onInput={(e) => props.setForm("contact_email", e.currentTarget.value)}
                  class={IC}
                  placeholder="email@consulta.com"
                />
              </Field>
              <ToggleSwitch
                label="Hacer público este email"
                checked={props.form.show_contact_email}
                onChange={(v) => props.setForm("show_contact_email", v)}
              />
            </div>

            {/* Dirección General */}
            <div class="space-y-4">
              <Field label="Dirección de Consultorio Principal">
                <input type="text"
                  value={props.form.service_address || ""}
                  onInput={(e) => props.setForm("service_address", e.currentTarget.value)}
                  class={IC}
                  placeholder="Ubicación física de la consulta"
                />
              </Field>
              <ToggleSwitch
                label="Hacer pública esta dirección"
                checked={props.form.show_public_service_address}
                onChange={(v) => props.setForm("show_public_service_address", v)}
              />
            </div>
          </div>
        </div>

        {/* ── NOTA SOBRE TELÉFONOS PÚBLICOS ── */}
        <div class="bg-amber-50 p-4 rounded-2xl border border-amber-100 flex items-start gap-3">
          <span class="text-lg">💡</span>
          <p class="text-[11px] text-amber-800 leading-relaxed">
            <span class="font-bold uppercase block mb-1">Nota sobre teléfonos públicos:</span>
            Los números de teléfono que aparecen en el directorio (Carabobo, Otros Estados o Exterior) se editan en la sección de <strong>"Ubicación Geográfica"</strong> más abajo, ya que están vinculados a la zona donde el psicólogo ejerce.
          </p>
        </div>

      </div>
  );
}