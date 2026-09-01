// web/src/components/psi/profile/ContactSection.tsx
import { InputField } from "./InputField";

interface ContactSectionProps {
  contactEmail: string;
  contactPhone: string;     
  contactCellPhone: string;  
  onContactEmailChange: (value: string) => void;
  onContactPhoneChange: (value: string) => void;     
  onContactCellPhoneChange: (value: string) => void; 
}

export function ContactSection(props: ContactSectionProps) {
  return (
    <section>
      <p class="text-[11px] text-gray-500 mb-4 font-medium leading-relaxed">
        Estos datos son de <span class="font-bold">uso exclusivo e interno</span> del Colegio de Psicólogos.
        No serán publicados en el directorio bajo ninguna circunstancia.
      </p>
      
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Fila 1 - Ocupa todo el ancho */}
        <div class="md:col-span-2">
          <InputField
            label="Email de Contacto (Gremial)"
            type="email"
            value={props.contactEmail}
            onInput={props.onContactEmailChange}
          />
        </div>

        {/* Fila 2 */}
        <InputField
          label="Teléfono Fijo / Local (Gremial)"
          type="tel"
          value={props.contactPhone}
          onInput={props.onContactPhoneChange}
        />
        <InputField
          label="Teléfono Móvil / WhatsApp (Gremial)"
          type="tel"
          value={props.contactCellPhone}
          onInput={props.onContactCellPhoneChange}
        />
      </div>
    </section>
  );
}