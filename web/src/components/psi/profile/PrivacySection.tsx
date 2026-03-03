// web/src/components/psi/profile/PrivacySection.tsx
import { ToggleSwitch } from "~/components/ui/ToggleSwitch";

interface PrivacySectionProps {
  showContactEmail: boolean;
  showPublicPhone: boolean;
  showServiceAddress: boolean;
  showUniversity: boolean;
  showGraduateDate: boolean;
  showMention: boolean;
  
  onShowContactEmailChange: (value: boolean) => void;
  onShowPublicPhoneChange: (value: boolean) => void;
  onShowServiceAddressChange: (value: boolean) => void;
  onShowUniversityChange: (value: boolean) => void;
  onShowGraduateDateChange: (value: boolean) => void;
  onShowMentionChange: (value: boolean) => void;
}

export function PrivacySection(props: PrivacySectionProps) {
  return (
    <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100">
      <div class="mb-6 border-l-4 border-colpsi-yellow pl-3">
        <h2 class="text-xl font-black text-colpsi-blue leading-tight">Centro de Privacidad</h2>
        <p class="text-sm text-colpsi-muted mt-1">Elige qué información deseas hacer pública en el directorio.</p>
      </div>
      
      <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
        <div class="space-y-4 bg-blue-50/50 p-6 rounded-3xl border border-blue-100">
          <h3 class="text-sm font-bold text-colpsi-blue uppercase tracking-widest border-b border-blue-100 pb-2">
            Información de Contacto
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

        <div class="space-y-4 bg-gray-50 p-6 rounded-3xl border border-gray-100">
          <h3 class="text-sm font-bold text-colpsi-text uppercase tracking-widest border-b border-gray-200 pb-2">
            Datos Académicos (Pregrado)
          </h3>
          <ToggleSwitch 
            label="Mostrar Universidad de Egreso" 
            checked={props.showUniversity} 
            onChange={props.onShowUniversityChange} 
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
      </div>
    </section>
  );
}