// web/src/components/psi/profile/SecuritySection.tsx
import { Show } from "solid-js";
import { PasswordInputComponent } from "~/components/ui/PasswordInput";
import { MessageAlert } from "./MessageAlert";

interface SecuritySectionProps {
  password: string;
  // El mensaje puede venir del estado global del formulario para mostrar errores de validación
  message: { type: "success" | "error"; text: string } | null;
  onPasswordChange: (value: string) => void;
}

export function SecuritySection(props: SecuritySectionProps) {
  return (
    <section>
      
      {/* 
        Muestra alertas de éxito o error específicas del proceso de guardado 
        justo encima del campo de validación.
      */}
      <Show when={props.message}>
        <div class="mb-6">
          <MessageAlert type={props.message!.type} text={props.message!.text} />
        </div>
      </Show>


      
      <div class="grid grid-cols-1 gap-6">
        <div class="space-y-1 w-full md:max-w-md">
          <label class="text-[10px] font-black text-colpsi-red uppercase ml-2 tracking-wider">
            Contraseña Actual <span class="lowercase font-medium text-red-400">(obligatoria para guardar)</span>
          </label>
          
          <PasswordInputComponent 
            required 
            value={props.password} 
            // Pasamos el valor directamente al handler como esperan los otros componentes
            onInput={(e) => props.onPasswordChange(e.currentTarget.value)} 
            class="w-full bg-red-50/20 border-2 border-red-50 focus:border-colpsi-red focus:bg-white rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all placeholder:text-red-200"
            placeholder="Introduce tu contraseña para confirmar" 
          />
          
          <div class="flex items-start gap-2 mt-2 ml-2">
            <span class="text-[10px] text-gray-400 leading-tight">
              Esta medida protege tu cuenta. Ningún cambio será procesado por el servidor sin esta validación.
            </span>
          </div>
        </div>
      </div>
    </section>
  );
}