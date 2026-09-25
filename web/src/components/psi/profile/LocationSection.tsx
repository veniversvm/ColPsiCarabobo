// web/src/components/psi/profile/LocationSection.tsx
import { For, Show } from "solid-js";
import { InputField } from "./InputField";
import { MUNICIPIOS_CARABOBO, ESTADOS_VENEZUELA, municipiosDe } from "~/lib/geo";
import { Icon } from "~/components/admin/ui/icons";

interface LocationSectionProps {
  // ── Carabobo ──
  municipalityCarabobo: string;
  phoneCarabobo: string;
  celPhoneCarabobo: string;
  serviceAddress: string; // NUEVO: Se muda aquí

  // ── Fuera de Carabobo (Venezuela) ──
  stateOutside: string;
  municipalityOutside: string;
  phoneOutside: string;
  celPhoneOutside: string;
  serviceAddressOutsideCarabobo: string;

  // ── Exterior (fuera de Venezuela) ──
  country: string;
  phoneOutsideVenezuela: string;
  cellPhoneOutsideVenezuela: string; 
  serviceAddressOutsideVenezuela: string;

  // ── Handlers ──
  onMunicipalityCaraboboChange: (value: string) => void;
  onPhoneCaraboboChange: (value: string) => void;
  onCelPhoneCaraboboChange: (value: string) => void;
  onServiceAddressChange: (value: string) => void; // NUEVO: Handler de Carabobo

  onStateOutsideChange: (value: string) => void;
  onMunicipalityOutsideChange: (value: string) => void;
  onPhoneOutsideChange: (value: string) => void;
  onCelPhoneOutsideChange: (value: string) => void;
  onServiceAddressOutsideCaraboboChange: (value: string) => void;

  onCountryChange: (value: string) => void;
  onPhoneOutsideVenezuelaChange: (value: string) => void;
  onCellPhoneOutsideVenezuelaChange: (value: string) => void; 
  onServiceAddressOutsideVenezuelaChange: (value: string) => void;
}

// Select con el mismo estilo visual de InputField, más opción legacy "(no estándar)".
interface SelectFieldProps {
  label: string;
  value: string;
  onChange: (value: string) => void;
  options: string[];
  placeholder?: string;
  disabled?: boolean;
  legacy?: string;
}

function SelectField(props: SelectFieldProps) {
  return (
    <div class="space-y-1">
      <div class="flex items-center justify-between">
        <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1 mb-1">
          {props.label}
        </label>
      </div>
      <select
        value={props.value}
        onChange={(e) => props.onChange(e.currentTarget.value)}
        disabled={props.disabled}
        class={`h-9 w-full rounded-md border bg-white px-3 text-sm text-colpsi-text outline-none transition-colors ${
          props.disabled
            ? "border-slate-200 bg-slate-50 opacity-60 cursor-not-allowed"
            : "border-slate-300 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15"
        }`}
      >
        <option value="">{props.placeholder ?? "Seleccionar…"}</option>
        <For each={props.options}>{(opt) => <option value={opt}>{opt}</option>}</For>
        <Show when={props.legacy}>
          <option value={props.legacy}>{props.legacy} (no estándar)</option>
        </Show>
      </select>
    </div>
  );
}

export function LocationSection(props: LocationSectionProps) {
  // Valores persistidos fuera del catálogo se conservan como opción "(no estándar)".
  const legacyMunicipioCarabobo = () =>
    props.municipalityCarabobo !== "" && !MUNICIPIOS_CARABOBO.includes(props.municipalityCarabobo);
  const legacyEstado = () => props.stateOutside !== "" && !ESTADOS_VENEZUELA.includes(props.stateOutside);
  const munisFuera = () => municipiosDe(props.stateOutside);
  const legacyMunicipioFuera = () =>
    props.municipalityOutside !== "" && !munisFuera().includes(props.municipalityOutside);

  return (
    <section class="space-y-10">

      {/* ── MENSAJE GLOBAL SOBRE PRIVACIDAD ────────────────────────────── */}
      <div class="bg-colpsi-bg/60 p-4 rounded-md border border-colpsi-border flex items-start gap-3">
        <Icon name="shield" class="w-5 h-5 text-colpsi-blue mt-0.5 shrink-0" />
        <div>
          <p class="text-[11px] font-semibold text-colpsi-text uppercase tracking-wide mb-1">
            Sobre tu Privacidad
          </p>
          <p class="text-xs text-colpsi-muted leading-relaxed">
            Completa aquí tus datos de ubicación y consulta por zona. Podrás elegir exactamente qué información ocultar o mostrar al público utilizando el <strong class="font-semibold text-colpsi-text">Centro de Privacidad</strong> ubicado en la siguiente sección.
          </p>
        </div>
      </div>

      {/* ── CARABOBO ───────────────────────────────────────────────────── */}
      <div>
        <h2 class="text-sm font-semibold text-colpsi-text uppercase tracking-wide mb-4 flex items-center gap-2">
          <Icon name="mapPin" class="w-4 h-4 text-colpsi-blue" />
          Presencia en Carabobo
        </h2>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <SelectField
            label="Municipio"
            value={props.municipalityCarabobo}
            onChange={props.onMunicipalityCaraboboChange}
            options={MUNICIPIOS_CARABOBO}
            placeholder="Seleccionar municipio…"
            legacy={legacyMunicipioCarabobo() ? props.municipalityCarabobo : ""}
          />
          <InputField
            label="Teléfono Fijo de Consulta"
            type="tel"
            value={props.phoneCarabobo}
            onInput={props.onPhoneCaraboboChange}
          />
          <InputField
            label="Celular de Consulta"
            type="tel"
            value={props.celPhoneCarabobo}
            onInput={props.onCelPhoneCaraboboChange}
          />
          {/* NUEVO: Dirección de Consulta Carabobo */}
          <div class="md:col-span-3">
            <InputField
              label="Dirección de Consultorio en Carabobo"
              value={props.serviceAddress}
              onInput={props.onServiceAddressChange}
            />
          </div>
        </div>
      </div>

      {/* ── FUERA DE CARABOBO (VENEZUELA) ──────────────────────────────── */}
      <div class="pt-5 border-t border-colpsi-border">
        <h2 class="text-sm font-semibold text-colpsi-text uppercase tracking-wide mb-4 flex items-center gap-2">
          <Icon name="map" class="w-4 h-4 text-colpsi-blue" />
          Otro Estado de Venezuela
        </h2>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <SelectField
            label="Estado"
            value={props.stateOutside}
            onChange={(v) => {
              props.onStateOutsideChange(v);
              const cur = props.municipalityOutside;
              if (cur && !municipiosDe(v).includes(cur)) props.onMunicipalityOutsideChange("");
            }}
            options={ESTADOS_VENEZUELA}
            placeholder="Seleccionar estado…"
            legacy={legacyEstado() ? props.stateOutside : ""}
          />
          <SelectField
            label="Ciudad / Municipio"
            value={props.municipalityOutside}
            onChange={props.onMunicipalityOutsideChange}
            options={munisFuera()}
            placeholder={
              !props.stateOutside
                ? "Primero selecciona un estado"
                : munisFuera().length === 0
                  ? "Este estado no tiene municipios"
                  : "Seleccionar municipio…"
            }
            disabled={!props.stateOutside}
            legacy={legacyMunicipioFuera() ? props.municipalityOutside : ""}
          />
          <InputField
            label="Teléfono Fijo"
            type="tel"
            value={props.phoneOutside}
            onInput={props.onPhoneOutsideChange}
          />
          <InputField
            label="Celular"
            type="tel"
            value={props.celPhoneOutside}
            onInput={props.onCelPhoneOutsideChange}
          />
          <div class="lg:col-span-4">
            <InputField
              label="Dirección de Consultorio Secundario"
              value={props.serviceAddressOutsideCarabobo}
              onInput={props.onServiceAddressOutsideCaraboboChange}
            />
          </div>
        </div>
      </div>

      {/* ── EXTERIOR (FUERA DE VENEZUELA) ──────────────────────────────── */}
      <div class="pt-5 border-t border-colpsi-border">
        <h2 class="text-sm font-semibold text-colpsi-text uppercase tracking-wide mb-4 flex items-center gap-2">
          <Icon name="globe" class="w-4 h-4 text-colpsi-blue" />
          Exterior (Fuera de Venezuela)
        </h2>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <InputField
            label="País"
            value={props.country}
            onInput={props.onCountryChange}
          />
          <InputField
            label="Teléfono Internacional"
            type="tel"
            value={props.phoneOutsideVenezuela}
            onInput={props.onPhoneOutsideVenezuelaChange}
          />
          <InputField
            label="Celular / Móvil"
            type="tel"
            value={props.cellPhoneOutsideVenezuela}
            onInput={props.onCellPhoneOutsideVenezuelaChange}
          />
          <div class="md:col-span-3">
            <InputField
              label="Dirección en el Exterior"
              value={props.serviceAddressOutsideVenezuela}
              onInput={props.onServiceAddressOutsideVenezuelaChange}
            />
          </div>
        </div>
      </div>

    </section>
  );
}