// web/src/components/admin/psicologos/edit/sections/LocationSection.tsx

import { Show } from "solid-js";
import { ToggleSwitch } from "~/components/ui/ToggleSwitch";
import { MUNICIPIOS_CARABOBO, ESTADOS_VENEZUELA, municipiosDe } from "~/lib/geo";
import { Field, IC } from "../EditPrimitives";
import type { EditFormState } from "../types";
import { Icon } from "~/components/admin/ui/icons";

interface Props {
  form: EditFormState;
  setForm: (key: keyof EditFormState, value: any) => void;
}

export function LocationSection(props: Props) {
  const hasLegacyMunicipio = () =>
    props.form.municipality_carabobo &&
    !MUNICIPIOS_CARABOBO.includes(props.form.municipality_carabobo);
  const hasLegacyEstado = () =>
    props.form.state_outside && !ESTADOS_VENEZUELA.includes(props.form.state_outside);

  return (
    <div class="space-y-12">

        {/* ── 1. CARABOBO ────────────────────────────────────────────────── */}
        <div class="space-y-6">
          <div class="flex items-center gap-2 border-b border-colpsi-border pb-2.5">
            <h3 class="text-sm font-semibold text-colpsi-blue uppercase tracking-wide">
              <Icon name="mapPin" class="w-3.5 h-3.5 inline-block mr-1.5 align-text-bottom" /> Presencia en Carabobo
            </h3>
          </div>
          
          <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
            <Field label="Municipio">
              <select
                value={props.form.municipality_carabobo}
                onChange={(e) => props.setForm("municipality_carabobo", e.currentTarget.value)}
                class={IC}
              >
                <option value="">Seleccionar municipio…</option>
                {MUNICIPIOS_CARABOBO.map((m) => (
                  <option value={m}>{m}</option>
                ))}
                <Show when={hasLegacyMunicipio()}>
                  <option value={props.form.municipality_carabobo}>
                    {props.form.municipality_carabobo} (no estándar)
                  </option>
                </Show>
              </select>
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

          <div class="bg-blue-50/50 p-5 rounded-lg border border-blue-100">
            <p class="text-[11px] font-semibold text-blue-600 uppercase tracking-wide mb-3 ml-1">Configuración de Visibilidad (Carabobo)</p>
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
        <div class="space-y-6 pt-4 border-t border-colpsi-border">
          <div class="flex items-center gap-2 border-b border-colpsi-border pb-2.5">
            <h3 class="text-sm font-semibold text-purple-900 uppercase tracking-wide">
              <Icon name="map" class="w-3.5 h-3.5 inline-block mr-1.5 align-text-bottom" /> Otro Estado de Venezuela
            </h3>
          </div>
          
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5">
            <Field label="Estado">
              <select
                value={props.form.state_outside}
                onChange={(e) => {
                  const v = e.currentTarget.value;
                  props.setForm("state_outside", v);
                  const cur = props.form.municipality_outside_carabobo;
                  if (cur && !municipiosDe(v).includes(cur)) {
                    props.setForm("municipality_outside_carabobo", "");
                  }
                }}
                class={IC}
              >
                <option value="">Seleccionar estado…</option>
                {ESTADOS_VENEZUELA.map((e) => (
                  <option value={e}>{e}</option>
                ))}
                <Show when={hasLegacyEstado()}>
                  <option value={props.form.state_outside}>
                    {props.form.state_outside} (no estándar)
                  </option>
                </Show>
              </select>
            </Field>
            <Field label="Ciudad / Municipio">
              <select
                value={props.form.municipality_outside_carabobo}
                onChange={(e) => props.setForm("municipality_outside_carabobo", e.currentTarget.value)}
                disabled={!props.form.state_outside}
                class={IC}
              >
                <option value="">
                  {!props.form.state_outside
                    ? "Primero selecciona un estado"
                    : municipiosDe(props.form.state_outside).length === 0
                      ? "Este estado no tiene municipios"
                      : "Seleccionar municipio…"}
                </option>
                {municipiosDe(props.form.state_outside).map((m) => (
                  <option value={m}>{m}</option>
                ))}
                {props.form.municipality_outside_carabobo &&
                  !municipiosDe(props.form.state_outside).includes(props.form.municipality_outside_carabobo) && (
                    <option value={props.form.municipality_outside_carabobo}>
                      {props.form.municipality_outside_carabobo} (no estándar)
                    </option>
                  )}
              </select>
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

          <div class="bg-purple-50/50 p-5 rounded-lg border border-purple-100">
            <p class="text-[11px] font-semibold text-purple-600 uppercase tracking-wide mb-3 ml-1">Configuración de Visibilidad (Nacional)</p>
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
        <div class="space-y-6 pt-4 border-t border-colpsi-border">
          <div class="flex items-center gap-2 border-b border-colpsi-border pb-2.5">
            <h3 class="text-sm font-semibold text-emerald-900 uppercase tracking-wide">
              <Icon name="globe" class="w-3.5 h-3.5 inline-block mr-1.5 align-text-bottom" /> Exterior (Fuera de Venezuela)
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

          <div class="bg-emerald-50/50 p-5 rounded-lg border border-emerald-100">
            <p class="text-[11px] font-semibold text-emerald-600 uppercase tracking-wide mb-3 ml-1">Configuración de Visibilidad (Exterior)</p>
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
  );
}