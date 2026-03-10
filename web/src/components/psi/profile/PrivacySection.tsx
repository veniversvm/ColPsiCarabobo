// web/src/components/psi/profile/PrivacySection.tsx
import { ToggleSwitch } from "~/components/ui/ToggleSwitch";

interface PrivacySectionProps {
  // Contacto principal
  showContactEmail: boolean;
  showPublicPhone: boolean;
  showServiceAddress: boolean;

  // Fuera de Carabobo
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
  onShowPublicPhoneChange: (value: boolean) => void;
  onShowServiceAddressChange: (value: boolean) => void;

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
        <p class="text-sm text-colpsi-muted mt-1">Elige qué información deseas hacer pública en el directorio.</p>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">

        {/* ── Contacto Principal ──────────────────────────────────────── */}
        <div class="space-y-4 bg-blue-50/50 p-6 rounded-3xl border border-blue-100">
          <h3 class="text-sm font-bold text-colpsi-blue uppercase tracking-widest border-b border-blue-100 pb-2">
            Contacto Principal
          </h3>
          <ToggleSwitch
            label="Mostrar Email de Contacto"
            checked={props.showContactEmail}
            onChange={props.onShowContactEmailChange}
          />
          <ToggleSwitch
            label="Mostrar Teléfono Público"
            checked={props.showPublicPhone}
            onChange={props.onShowPublicPhoneChange}
          />
          <ToggleSwitch
            label="Mostrar Dirección de Consulta"
            checked={props.showServiceAddress}
            onChange={props.onShowServiceAddressChange}
          />
        </div>

        {/* ── Datos Académicos ────────────────────────────────────────── */}
        <div class="space-y-4 bg-gray-50 p-6 rounded-3xl border border-gray-100">
          <h3 class="text-sm font-bold text-colpsi-text uppercase tracking-widest border-b border-gray-200 pb-2">
            Datos Académicos (Pregrado)
          </h3>
          <ToggleSwitch
            label="Mostrar Universidad Egreso"
            checked={props.showUniversity}
            onChange={props.onShowUniversity}
          />
          <ToggleSwitch
            label="Mostrar Fecha de Egreso"
            checked={props.showGraduateDate}
            onChange={props.onShowGraduateDateChange}
          />
          <ToggleSwitch
            label="Mostrar Mención de Grado"
            checked={props.showMention}
            onChange={props.onShowMentionChange}
          />
        </div>

        {/* ── Otro Estado Venezuela ───────────────────────────────────── */}
        <div class="space-y-4 bg-yellow-50/40 p-6 rounded-3xl border border-yellow-100">
          <h3 class="text-sm font-bold text-colpsi-blue uppercase tracking-widest border-b border-yellow-100 pb-2">
            Otro Estado de Venezuela
          </h3>
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
            Exterior
          </h3>
          <ToggleSwitch
            label="Mostrar Teléfono Internacional"
            checked={props.showPhoneOutsideVenezuela}
            onChange={props.onShowPhoneOutsideVenezuelaChange}
          />
          <ToggleSwitch
            label="Mostrar Celular Internacional"
            checked={props.showCelPhoneOutsideVenezuela}
            onChange={props.onShowCelPhoneOutsideVenezuelaChange}
          />
          <ToggleSwitch
            label="Mostrar Dirección en el Exterior"
            checked={props.showServiceAddressOutsideVenezuela}
            onChange={props.onShowServiceAddressOutsideVenezuelaChange}
          />
        </div>

      </div>
    </section>
  );
}