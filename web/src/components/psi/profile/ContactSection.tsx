import { InputField } from "./InputField";

// web/src/components/psi/profile/ContactSection.tsx
interface ContactSectionProps {
  contactEmail: string;
  contactPhone: string;      // Actualizado: Antes publicPhone
  contactCellPhone: string;  // Nuevo
  serviceAddress: string;
  onContactEmailChange: (value: string) => void;
  onContactPhoneChange: (value: string) => void;     // Actualizado
  onContactCellPhoneChange: (value: string) => void; // Nuevo
  onServiceAddressChange: (value: string) => void;
}

export function ContactSection(props: ContactSectionProps) {
  return (
    <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100">
      <div class="mb-6 border-l-4 border-colpsi-yellow pl-3">
        <h2 class="text-xl font-black text-colpsi-blue leading-tight">Contacto de Consulta</h2>
      </div>
      
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Fila 1 */}
        <InputField
          label="Email de Consulta"
          type="email"
          value={props.contactEmail}
          onInput={props.onContactEmailChange}
        />
        <InputField
          label="Teléfono Fijo / Local"
          type="tel"
          value={props.contactPhone}
          onInput={props.onContactPhoneChange}
        />

        {/* Fila 2 */}
        <InputField
          label="Teléfono Móvil (WhatsApp)"
          type="tel"
          value={props.contactCellPhone}
          onInput={props.onContactCellPhoneChange}
        />

        {/* Fila 3 - Ocupa todo el ancho */}
        <div class="md:col-span-2">
          <InputField
            label="Dirección de Consultorio"
            value={props.serviceAddress}
            onInput={props.onServiceAddressChange}
          />
        </div>
      </div>
    </section>
  );
}