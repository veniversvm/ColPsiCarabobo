// web/src/components/psi/profile/LocationSection.tsx
import { InputField } from "./InputField";

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

export function LocationSection(props: LocationSectionProps) {
  return (
    <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100 space-y-10">

      {/* ── MENSAJE GLOBAL SOBRE PRIVACIDAD ────────────────────────────── */}
      <div class="bg-blue-50/80 p-5 rounded-3xl border border-blue-100 flex items-start gap-4">
        <span class="text-3xl mt-1">🛡️</span>
        <div>
          <p class="text-xs font-black text-blue-900 uppercase tracking-widest mb-1">
            Sobre tu Privacidad
          </p>
          <p class="text-xs text-blue-800 leading-relaxed font-medium">
            Completa aquí tus datos de ubicación y consulta por zona. Podrás elegir exactamente qué información ocultar o mostrar al público utilizando el <strong>Centro de Privacidad</strong> ubicado en la siguiente sección.
          </p>
        </div>
      </div>

      {/* ── CARABOBO ───────────────────────────────────────────────────── */}
      <div>
        <h2 class="text-lg font-black text-colpsi-blue mb-5 border-l-4 border-colpsi-yellow pl-3 uppercase tracking-tight">
          📍 Presencia en Carabobo
        </h2>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <InputField
            label="Municipio"
            value={props.municipalityCarabobo}
            onInput={props.onMunicipalityCaraboboChange}
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
      <div class="pt-6 border-t border-gray-100">
        <h2 class="text-lg font-black text-colpsi-blue mb-5 border-l-4 border-gray-300 pl-3 uppercase tracking-tight">
          🗺️ Otro Estado de Venezuela
        </h2>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <InputField
            label="Estado"
            value={props.stateOutside}
            onInput={props.onStateOutsideChange}
          />
          <InputField
            label="Ciudad / Municipio"
            value={props.municipalityOutside}
            onInput={props.onMunicipalityOutsideChange}
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
      <div class="pt-6 border-t border-gray-100">
        <h2 class="text-lg font-black text-colpsi-blue mb-5 border-l-4 border-gray-300 pl-3 uppercase tracking-tight">
          🌐 Exterior (Fuera de Venezuela)
        </h2>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
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