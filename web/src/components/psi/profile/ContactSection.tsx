import { InputField } from "./InputField";

// web/src/components/psi/profile/ContactSection.tsx
interface ContactSectionProps {
  contactEmail: string;
  publicPhone: string;
  serviceAddress: string;
  onContactEmailChange: (value: string) => void;
  onPublicPhoneChange: (value: string) => void;
  onServiceAddressChange: (value: string) => void;
}

export function ContactSection(props: ContactSectionProps) {
  return (
    <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100">
      <div class="mb-6 border-l-4 border-colpsi-yellow pl-3">
        <h2 class="text-xl font-black text-colpsi-blue leading-tight">Contacto de Consulta</h2>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <InputField
          label="Email de Consulta"
          type="email"
          value={props.contactEmail}
          onInput={props.onContactEmailChange}
        />
        <InputField
          label="Teléfono Principal"
          type="tel"
          value={props.publicPhone}
          onInput={props.onPublicPhoneChange}
        />
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