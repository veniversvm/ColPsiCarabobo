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
    <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100 transition-all hover:border-red-100">
      
      {/* 
        Muestra alertas de éxito o error específicas del proceso de guardado 
        justo encima del campo de validación.
      */}
      <Show when={props.message}>
        <div class="mb-6">
          <MessageAlert type={props.message!.type} text={props.message!.text} />
        </div>
      </Show>

      <div class="flex items-center gap-3 mb-6 border-l-4 border-colpsi-red pl-3">
        <div class="bg-red-50 p-2 rounded-lg">
           <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-colpsi-red" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M2.166 4.999A11.954 11.954 0 0010 1.944 11.954 11.954 0 0017.834 5c.11.65.166 1.32.166 2.001 0 5.225-3.34 9.67-8 11.317C5.34 16.67 2 12.225 2 7c0-.682.057-1.35.166-2.001zm11.541 3.708a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
          </svg>
        </div>
        <h2 class="text-xl font-black text-colpsi-red leading-tight">
          Validación de Seguridad
        </h2>
      </div>
      
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