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

const boxClass =
  "space-y-3 bg-colpsi-bg/50 p-4 rounded-lg border border-colpsi-border";
const boxTitleClass =
  "text-[11px] font-semibold text-colpsi-text uppercase tracking-wide border-b border-colpsi-border pb-2";

export function PrivacySection(props: PrivacySectionProps) {
  return (
    <section>


      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">

        {/* ── Contacto Principal ──────────────────────────────────────── */}
        <div class={boxClass}>
          <h3 class={boxTitleClass}>
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
        <div class={boxClass}>
          <h3 class={boxTitleClass}>
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
        <div class={boxClass}>
          <h3 class={boxTitleClass}>
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
        <div class={boxClass}>
          <h3 class={boxTitleClass}>
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
        <div class={`${boxClass} md:col-span-2`}>
          <h3 class={boxTitleClass}>
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