// web/src/components/psi/profile/PrivacySection.tsx
import { ToggleSwitch } from "~/components/ui/ToggleSwitch";

interface PrivacySectionProps {
  // Contacto principal
  showContactEmail: boolean;
  showServiceAddress: boolean;

  // Carabobo (NUEVOS)
  showMunicipalityCarabobo: boolean;
  showPhoneCarabobo: boolean;
  showCelPhoneCarabobo: boolean;

  // Fuera de Carabobo
  showStateOutside: boolean;
  showMunicipalityOutsideCarabobo: boolean;
  showPhoneOutsideCarabobo: boolean;
  showCelPhoneOutsideCarabobo: boolean;
  showServiceAddressOutsideCarabobo: boolean;

  // Exterior
  showPhoneOutsideVenezuela: boolean;
  showCelPhoneOutsideVenezuela: boolean;
  showServiceAddressOutsideVenezuela: boolean;

  // Datos académicos
  showGraduateDate: boolean;
  showMention: boolean;
  showUniversity: boolean;

  onShowContactEmailChange: (value: boolean) => void;
  onShowServiceAddressChange: (value: boolean) => void;

  onShowMunicipalityCaraboboChange: (value: boolean) => void;
  onShowPhoneCaraboboChange: (value: boolean) => void;
  onShowCelPhoneCaraboboChange: (value: boolean) => void;

  onShowStateOutsideChange: (value: boolean) => void;
  onShowMunicipalityOutsideCaraboboChange: (value: boolean) => void;
  onShowPhoneOutsideCaraboboChange: (value: boolean) => void;
  onShowCelPhoneOutsideCaraboboChange: (value: boolean) => void;
  onShowServiceAddressOutsideCaraboboChange: (value: boolean) => void;

  onShowPhoneOutsideVenezuelaChange: (value: boolean) => void;
  onShowCelPhoneOutsideVenezuelaChange: (value: boolean) => void;
  onShowServiceAddressOutsideVenezuelaChange: (value: boolean) => void;

  onShowGraduateDateChange: (value: boolean) => void;
  onShowMentionChange: (value: boolean) => void;
  onShowUniversity: (value: boolean) => void;
}

export function PrivacySection(props: PrivacySectionProps) {
  return (
    <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100">
      <div class="mb-6 border-l-4 border-colpsi-yellow pl-3">
        <h2 class="text-xl font-black text-colpsi-blue leading-tight">Centro de Privacidad</h2>
        <p class="text-sm text-colpsi-muted mt-1">Controla qué información es visible para el público en el directorio.</p>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">

        {/* ── Contacto Principal ──────────────────────────────────────── */}
        <div class="space-y-4 bg-blue-50/50 p-6 rounded-3xl border border-blue-100">
          <h3 class="text-sm font-bold text-colpsi-blue uppercase tracking-widest border-b border-blue-100 pb-2">
            Privacidad General
          </h3>
          <ToggleSwitch
            label="Mostrar Email de Contacto"
            checked={props.showContactEmail}
            onChange={props.onShowContactEmailChange}
          />
          <ToggleSwitch
            label="Mostrar Dirección de Consulta"
            checked={props.showServiceAddress}
            onChange={props.onShowServiceAddressChange}
          />
        </div>

        {/* ── Carabobo ────────────────────────────────────────────────── */}
        <div class="space-y-4 bg-indigo-50/40 p-6 rounded-3xl border border-indigo-100">
          <h3 class="text-sm font-bold text-indigo-900 uppercase tracking-widest border-b border-indigo-100 pb-2">
            Presencia en Carabobo
          </h3>
          <ToggleSwitch
            label="Mostrar Municipio (Carabobo)"
            checked={props.showMunicipalityCarabobo}
            onChange={props.onShowMunicipalityCaraboboChange}
          />
          <ToggleSwitch
            label="Mostrar Teléfono Fijo"
            checked={props.showPhoneCarabobo}
            onChange={props.onShowPhoneCaraboboChange}
          />
          <ToggleSwitch
            label="Mostrar Celular"
            checked={props.showCelPhoneCarabobo}
            onChange={props.onShowCelPhoneCaraboboChange}
          />
        </div>

        {/* ── Otro Estado Venezuela ───────────────────────────────────── */}
        <div class="space-y-4 bg-yellow-50/40 p-6 rounded-3xl border border-yellow-100">
          <h3 class="text-sm font-bold text-colpsi-blue uppercase tracking-widest border-b border-yellow-100 pb-2">
            Fuera de Carabobo
          </h3>
          <ToggleSwitch
            label="Mostrar Estado"
            checked={props.showStateOutside}
            onChange={props.onShowStateOutsideChange}
          />
          <ToggleSwitch
            label="Mostrar Municipio/Ciudad"
            checked={props.showMunicipalityOutsideCarabobo}
            onChange={props.onShowMunicipalityOutsideCaraboboChange}
          />
          <ToggleSwitch
            label="Mostrar Teléfono Fijo"
            checked={props.showPhoneOutsideCarabobo}
            onChange={props.onShowPhoneOutsideCaraboboChange}
          />
          <ToggleSwitch
            label="Mostrar Celular"
            checked={props.showCelPhoneOutsideCarabobo}
            onChange={props.onShowCelPhoneOutsideCaraboboChange}
          />
          <ToggleSwitch
            label="Mostrar Dirección de Consulta"
            checked={props.showServiceAddressOutsideCarabobo}
            onChange={props.onShowServiceAddressOutsideCaraboboChange}
          />
        </div>

        {/* ── Exterior ────────────────────────────────────────────────── */}
        <div class="space-y-4 bg-green-50/40 p-6 rounded-3xl border border-green-100">
          <h3 class="text-sm font-bold text-colpsi-blue uppercase tracking-widest border-b border-green-100 pb-2">
            Exterior (Internacional)
          </h3>
          <ToggleSwitch
            label="Mostrar Teléfono Fijo"
            checked={props.showPhoneOutsideVenezuela}
            onChange={props.onShowPhoneOutsideVenezuelaChange}
          />
          <ToggleSwitch
            label="Mostrar Celular"
            checked={props.showCelPhoneOutsideVenezuela}
            onChange={props.onShowCelPhoneOutsideVenezuelaChange}
          />
          <ToggleSwitch
            label="Mostrar Dirección Exterior"
            checked={props.showServiceAddressOutsideVenezuela}
            onChange={props.onShowServiceAddressOutsideVenezuelaChange}
          />
        </div>

        {/* ── Datos Académicos ────────────────────────────────────────── */}
        <div class="space-y-4 bg-gray-50 p-6 rounded-3xl border border-gray-100 md:col-span-2">
          <h3 class="text-sm font-bold text-colpsi-text uppercase tracking-widest border-b border-gray-200 pb-2">
            Formación Académica
          </h3>
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <ToggleSwitch
              label="Mostrar Universidad"
              checked={props.showUniversity}
              onChange={props.onShowUniversity}
            />
            <ToggleSwitch
              label="Mostrar Fecha de Grado"
              checked={props.showGraduateDate}
              onChange={props.onShowGraduateDateChange}
            />
            <ToggleSwitch
              label="Mostrar Mención"
              checked={props.showMention}
              onChange={props.onShowMentionChange}
            />
          </div>
        </div>

      </div>
    </section>
  );
}