// web/src/components/psi/profile/LocationSection.tsx
import { InputField } from "./InputField";

interface LocationSectionProps {
  // Carabobo
  municipalityCarabobo: string;
  phoneCarabobo: string;
  celPhoneCarabobo: string;

  // Fuera de Carabobo (Venezuela)
  stateOutside: string;
  municipalityOutside: string;
  phoneOutside: string;
  celPhoneOutside: string;
  serviceAddressOutsideCarabobo: string;

  // Exterior (fuera de Venezuela)
  country: string;
  phoneOutsideVenezuela: string;
  cellPhoneOutsideVenezuela: string; // NUEVO
  serviceAddressOutsideVenezuela: string;

  onMunicipalityCaraboboChange: (value: string) => void;
  onPhoneCaraboboChange: (value: string) => void;
  onCelPhoneCaraboboChange: (value: string) => void;

  onStateOutsideChange: (value: string) => void;
  onMunicipalityOutsideChange: (value: string) => void;
  onPhoneOutsideChange: (value: string) => void;
  onCelPhoneOutsideChange: (value: string) => void;
  onServiceAddressOutsideCaraboboChange: (value: string) => void;

  onCountryChange: (value: string) => void;
  onPhoneOutsideVenezuelaChange: (value: string) => void;
  onCellPhoneOutsideVenezuelaChange: (value: string) => void; // NUEVO
  onServiceAddressOutsideVenezuelaChange: (value: string) => void;
}

export function LocationSection(props: LocationSectionProps) {
  return (
    <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100 space-y-10">

      {/* ── CARABOBO ───────────────────────────────────────────────────── */}
      <div>
        <h2 class="text-xl font-black text-colpsi-blue mb-2 border-l-4 border-colpsi-yellow pl-3">
          Ubicación en Carabobo
        </h2>
        <p class="text-sm text-gray-500 mb-6 bg-gray-50 p-4 rounded-2xl border border-gray-100 leading-relaxed italic">
          El municipio es información pública. Los teléfonos son privados y solo los utiliza el Colegio internamente.
        </p>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <InputField
            label="Municipio"
            value={props.municipalityCarabobo}
            onInput={props.onMunicipalityCaraboboChange}
          />
          <InputField
            label="Teléfono Fijo"
            type="tel"
            value={props.phoneCarabobo}
            onInput={props.onPhoneCaraboboChange}
          />
          <InputField
            label="Celular"
            type="tel"
            value={props.celPhoneCarabobo}
            onInput={props.onCelPhoneCaraboboChange}
          />
        </div>
      </div>

      {/* ── FUERA DE CARABOBO (VENEZUELA) ──────────────────────────────── */}
      <div class="pt-6 border-t border-gray-100">
        <h2 class="text-xl font-black text-colpsi-blue mb-2 border-l-4 border-gray-300 pl-3">
          Otro Estado de Venezuela
        </h2>
        <p class="text-sm text-gray-500 mb-6 bg-gray-50 p-4 rounded-2xl border border-gray-100 leading-relaxed italic">
          Si también ejerces en otro estado venezolano. La visibilidad de estos datos se controla en el Centro de Privacidad.
        </p>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
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
          <div class="md:col-span-2">
            <InputField
              label="Dirección de Consultorio"
              value={props.serviceAddressOutsideCarabobo}
              onInput={props.onServiceAddressOutsideCaraboboChange}
            />
          </div>
        </div>
      </div>

      {/* ── EXTERIOR (FUERA DE VENEZUELA) ──────────────────────────────── */}
      <div class="pt-6 border-t border-gray-100">
        <h2 class="text-xl font-black text-colpsi-blue mb-2 border-l-4 border-gray-300 pl-3">
          Exterior
        </h2>
        <p class="text-sm text-gray-500 mb-6 bg-gray-50 p-4 rounded-2xl border border-gray-100 leading-relaxed italic">
          Si resides o ejerces fuera de Venezuela. Indica país y contactos internacionales.
        </p>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <InputField
            label="País"
            value={props.country}
            onInput={props.onCountryChange}
          />
          <InputField
            label="Teléfono Fijo"
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