// web/src/components/admin/psicologos/edit/sections/LocationSection.tsx

import { ToggleSwitch } from "~/components/ui/ToggleSwitch";
import { Field, IC, SectionCard } from "../EditPrimitives";
import type { EditFormState } from "../types";

interface Props {
  form: EditFormState;
  setForm: (key: keyof EditFormState, value: any) => void;
}

export function LocationSection(props: Props) {
  return (
    <SectionCard title="Ubicación & Datos Regionales" accent="border-blue-300">
      <div class="space-y-8">

        {/* ── Carabobo ── */}
        <div>
          <h3 class="text-xs font-black text-blue-700 uppercase tracking-widest mb-4 pb-2 border-b border-blue-100">
            📍 Carabobo
          </h3>
          <p class="text-xs text-gray-400 mb-4 italic">
            Los teléfonos son privados — solo para uso interno del Colegio.
          </p>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Field label="Municipio">
              <input type="text" value={props.form.municipality_carabobo}
                onInput={(e) => props.setForm("municipality_carabobo", e.currentTarget.value)} class={IC} />
            </Field>
            <Field label="Teléfono Fijo">
              <input type="tel" value={props.form.phone_carabobo}
                onInput={(e) => props.setForm("phone_carabobo", e.currentTarget.value)} class={IC} />
            </Field>
            <Field label="Celular">
              <input type="tel" value={props.form.cel_phone_carabobo}
                onInput={(e) => props.setForm("cel_phone_carabobo", e.currentTarget.value)} class={IC} />
            </Field>
          </div>
        </div>

        {/* ── Otro Estado de Venezuela ── */}
        <div class="pt-4 border-t border-gray-100">
          <h3 class="text-xs font-black text-purple-700 uppercase tracking-widest mb-4 pb-2 border-b border-purple-100">
            🗺️ Otro Estado de Venezuela
          </h3>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field label="Estado">
              <input type="text" value={props.form.state_outside}
                onInput={(e) => props.setForm("state_outside", e.currentTarget.value)} class={IC} />
            </Field>
            <Field label="Ciudad / Municipio">
              <input type="text" value={props.form.municipality_outside_carabobo}
                onInput={(e) => props.setForm("municipality_outside_carabobo", e.currentTarget.value)} class={IC} />
            </Field>
            <Field label="Teléfono Fijo">
              <input type="tel" value={props.form.phone_outside_carabobo}
                onInput={(e) => props.setForm("phone_outside_carabobo", e.currentTarget.value)} class={IC} />
            </Field>
            <Field label="Celular">
              <input type="tel" value={props.form.cel_phone_outside_carabobo}
                onInput={(e) => props.setForm("cel_phone_outside_carabobo", e.currentTarget.value)} class={IC} />
            </Field>
            <div class="md:col-span-2">
              <Field label="Dirección de Consultorio">
                <input type="text" value={props.form.service_address_outside_carabobo}
                  onInput={(e) => props.setForm("service_address_outside_carabobo", e.currentTarget.value)} class={IC} />
              </Field>
            </div>
          </div>
          <div class="mt-4 bg-yellow-50/50 p-4 rounded-2xl border border-yellow-100 grid grid-cols-1 sm:grid-cols-3 gap-3">
            <ToggleSwitch label="Mostrar Teléfono Fijo"
              checked={props.form.show_phone_outside_carabobo}
              onChange={(v) => props.setForm("show_phone_outside_carabobo", v)} />
            <ToggleSwitch label="Mostrar Celular"
              checked={props.form.show_cel_phone_outside_carabobo}
              onChange={(v) => props.setForm("show_cel_phone_outside_carabobo", v)} />
            <ToggleSwitch label="Mostrar Dirección"
              checked={props.form.show_public_service_address_outside_carabobo}
              onChange={(v) => props.setForm("show_public_service_address_outside_carabobo", v)} />
          </div>
        </div>

        {/* ── Exterior ── */}
        <div class="pt-4 border-t border-gray-100">
          <h3 class="text-xs font-black text-green-700 uppercase tracking-widest mb-4 pb-2 border-b border-green-100">
            🌐 Exterior (Fuera de Venezuela)
          </h3>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Field label="País">
              <input type="text" value={props.form.country}
                onInput={(e) => props.setForm("country", e.currentTarget.value)} class={IC} />
            </Field>
            <Field label="Teléfono Internacional">
              <input type="tel" value={props.form.phone_outside_venezuela}
                onInput={(e) => props.setForm("phone_outside_venezuela", e.currentTarget.value)} class={IC} />
            </Field>
            <div class="md:col-span-2">
              <Field label="Dirección en el Exterior">
                <input type="text" value={props.form.service_address_outside_venezuela}
                  onInput={(e) => props.setForm("service_address_outside_venezuela", e.currentTarget.value)} class={IC} />
              </Field>
            </div>
          </div>
          <div class="mt-4 bg-green-50/50 p-4 rounded-2xl border border-green-100 grid grid-cols-1 sm:grid-cols-3 gap-3">
            <ToggleSwitch label="Mostrar Teléfono Internacional"
              checked={props.form.show_phone_outside_venezuela}
              onChange={(v) => props.setForm("show_phone_outside_venezuela", v)} />
            <ToggleSwitch label="Mostrar Celular Internacional"
              checked={props.form.show_cel_phone_outside_venezuela}
              onChange={(v) => props.setForm("show_cel_phone_outside_venezuela", v)} />
            <ToggleSwitch label="Mostrar Dirección en Exterior"
              checked={props.form.show_public_service_address_outside_venezuela}
              onChange={(v) => props.setForm("show_public_service_address_outside_venezuela", v)} />
          </div>
        </div>

      </div>
    </SectionCard>
  );
}