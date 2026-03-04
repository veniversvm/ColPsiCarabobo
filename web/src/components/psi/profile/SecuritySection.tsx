// web/src/components/psi/profile/SecuritySection.tsx
import { Show } from "solid-js";
import { PasswordInputComponent } from "~/components/ui/PasswordInput";
import { MessageAlert } from "./MessageAlert";

interface SecuritySectionProps {
  password: string;
  message: { type: "success" | "error"; text: string } | null;
  onPasswordChange: (value: string) => void;
}

export function SecuritySection(props: SecuritySectionProps) {
  return (
    <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100">
      <Show when={props.message}>
        <MessageAlert type={props.message!.type} text={props.message!.text} />
      </Show>

      <h2 class="text-xl font-black text-colpsi-red mb-6 border-l-4 border-colpsi-red pl-3">
        Validación de Seguridad
      </h2>
      
      <div class="grid grid-cols-1 gap-6">
        <div class="space-y-1 w-full md:max-w-md">
          <label class="text-[10px] font-black text-colpsi-red uppercase ml-2 italic">
            Contraseña Actual (Obligatoria para guardar cambios)
          </label>
          
          <PasswordInputComponent 
            required 
            value={props.password} 
            onInput={(e) => props.onPasswordChange(e.currentTarget.value)} 
            class="w-full bg-red-50/30 border-2 border-red-100 focus:border-colpsi-red rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all"
            placeholder="••••••••" 
          />
          
          <p class="text-[9px] text-gray-400 ml-2">
            Por seguridad, confirma tu identidad para aplicar los cambios en tu perfil.
          </p>
        </div>
      </div>
    </section>
  );
}