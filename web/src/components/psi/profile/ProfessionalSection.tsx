// web/src/components/psi/profile/ProfessionalSection.tsx
import { enforceMaxLength } from "~/lib/sanitizer";
import { InputField } from "./InputField";

interface ProfessionalSectionProps {
  primarySpecialty: string;
  secondarySpecialty: string;
  miniBio: string;
  onPrimarySpecialtyChange: (value: string) => void;
  onSecondarySpecialtyChange: (value: string) => void;
  onMiniBioChange: (value: string) => void;
}

export function ProfessionalSection(props: ProfessionalSectionProps) {
  return (
    <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100">
      <h2 class="text-xl font-black text-colpsi-blue mb-6 border-l-4 border-colpsi-yellow pl-3">
        Perfil Profesional
      </h2>
      
      <div class="space-y-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <InputField
            label="Especialidad Principal"
            value={props.primarySpecialty}
            onInput={props.onPrimarySpecialtyChange}
          />
          <InputField
            label="Especialidad Secundaria"
            value={props.secondarySpecialty}
            onInput={props.onSecondarySpecialtyChange}
          />
        </div>
        
        <div class="space-y-2">
          <label class="text-xs font-bold text-gray-500 uppercase ml-2">Presentación</label>
          <textarea 
            value={props.miniBio} 
            onInput={(e) => {
              const text = enforceMaxLength(e.currentTarget.value, 250);
              props.onMiniBioChange(text);
            }}
            maxlength="250"
            class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-2xl px-5 py-4 outline-none text-colpsi-text transition-all min-h-[120px] resize-y" 
            placeholder="Describe brevemente tu práctica profesional..."
          />
          <div class="text-right text-xs text-gray-400">
            {props.miniBio?.length || 0} / 250
          </div>
        </div>
      </div>
    </section>
  );
}