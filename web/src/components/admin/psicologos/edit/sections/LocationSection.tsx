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
    <SectionCard title="Ubicación Geográfica y Privacidad" accent="border-indigo-400">
      <div class="space-y-12">

        {/* ── 1. CARABOBO ────────────────────────────────────────────────── */}
        <div class="space-y-6">
          <div class="flex items-center gap-2 border-l-4 border-blue-600 pl-3">
            <h3 class="text-sm font-black text-blue-900 uppercase tracking-widest">
              📍 Presencia en Carabobo
            </h3>
          </div>
          
          <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
            <Field label="Municipio">
              <input type="text" value={props.form.municipality_carabobo}
                onInput={(e) => props.setForm("municipality_carabobo", e.currentTarget.value)} class={IC} 
                placeholder="Ej: Valencia" />
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

          <div class="bg-blue-50/50 p-5 rounded-2xl border border-blue-100">
            <p class="text-[10px] font-black text-blue-600 uppercase tracking-widest mb-3 ml-1">Configuración de Visibilidad (Carabobo)</p>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <ToggleSwitch label="Públicar Municipio" 
                checked={props.form.show_municipality_carabobo}
                onChange={(v) => props.setForm("show_municipality_carabobo", v)} />
              <ToggleSwitch label="Publicar Fijo" 
                checked={props.form.show_phone_carabobo}
                onChange={(v) => props.setForm("show_phone_carabobo", v)} />
              <ToggleSwitch label="Publicar Celular" 
                checked={props.form.show_cel_phone_carabobo}
                onChange={(v) => props.setForm("show_cel_phone_carabobo", v)} />
            </div>
          </div>
        </div>

        {/* ── 2. OTRO ESTADO DE VENEZUELA ────────────────────────────────── */}
        <div class="space-y-6 pt-4 border-t border-gray-100">
          <div class="flex items-center gap-2 border-l-4 border-purple-600 pl-3">
            <h3 class="text-sm font-black text-purple-900 uppercase tracking-widest">
              🗺️ Otro Estado de Venezuela
            </h3>
          </div>
          
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5">
            <Field label="Estado">
              <input type="text" value={props.form.state_outside}
                onInput={(e) => props.setForm("state_outside", e.currentTarget.value)} class={IC} 
                placeholder="Ej: Aragua" />
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
            <div class="lg:col-span-4">
              <Field label="Dirección de Consultorio Secundario">
                <input type="text" value={props.form.service_address_outside_carabobo}
                  onInput={(e) => props.setForm("service_address_outside_carabobo", e.currentTarget.value)} class={IC} />
              </Field>
            </div>
          </div>

          <div class="bg-purple-50/50 p-5 rounded-2xl border border-purple-100">
            <p class="text-[10px] font-black text-purple-600 uppercase tracking-widest mb-3 ml-1">Configuración de Visibilidad (Nacional)</p>
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
              <ToggleSwitch label="Mostrar Estado" 
                checked={props.form.show_state_outside}
                onChange={(v) => props.setForm("show_state_outside", v)} />
              <ToggleSwitch label="Mostrar Ciudad" 
                checked={props.form.show_municipality_outside_carabobo}
                onChange={(v) => props.setForm("show_municipality_outside_carabobo", v)} />
              <ToggleSwitch label="Mostrar Fijo" 
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
        </div>

        {/* ── 3. EXTERIOR ────────────────────────────────────────────────── */}
        <div class="space-y-6 pt-4 border-t border-gray-100">
          <div class="flex items-center gap-2 border-l-4 border-emerald-600 pl-3">
            <h3 class="text-sm font-black text-emerald-900 uppercase tracking-widest">
              🌐 Exterior (Fuera de Venezuela)
            </h3>
          </div>
          
          <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
            <Field label="País">
              <input type="text" value={props.form.country}
                onInput={(e) => props.setForm("country", e.currentTarget.value)} class={IC} 
                placeholder="Ej: España" />
            </Field>
            <Field label="Teléfono Fijo">
              <input type="tel" value={props.form.phone_outside_venezuela}
                onInput={(e) => props.setForm("phone_outside_venezuela", e.currentTarget.value)} class={IC} />
            </Field>
            <Field label="Celular Internacional">
              <input type="tel" value={props.form.cell_phone_outside_venezuela}
                onInput={(e) => props.setForm("cell_phone_outside_venezuela", e.currentTarget.value)} class={IC} />
            </Field>
            <div class="md:col-span-3">
              <Field label="Dirección en el Exterior">
                <input type="text" value={props.form.service_address_outside_venezuela}
                  onInput={(e) => props.setForm("service_address_outside_venezuela", e.currentTarget.value)} class={IC} />
              </Field>
            </div>
          </div>

          <div class="bg-emerald-50/50 p-5 rounded-2xl border border-emerald-100">
            <p class="text-[10px] font-black text-emerald-600 uppercase tracking-widest mb-3 ml-1">Configuración de Visibilidad (Exterior)</p>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <ToggleSwitch label="Publicar Fijo Int." 
                checked={props.form.show_phone_outside_venezuela}
                onChange={(v) => props.setForm("show_phone_outside_venezuela", v)} />
              <ToggleSwitch label="Publicar Celular Int." 
                checked={props.form.show_cel_phone_outside_venezuela}
                onChange={(v) => props.setForm("show_cel_phone_outside_venezuela", v)} />
              <ToggleSwitch label="Publicar Dirección Int." 
                checked={props.form.show_public_service_address_outside_venezuela}
                onChange={(v) => props.setForm("show_public_service_address_outside_venezuela", v)} />
            </div>
          </div>
        </div>

      </div>
    </SectionCard>
  );
}