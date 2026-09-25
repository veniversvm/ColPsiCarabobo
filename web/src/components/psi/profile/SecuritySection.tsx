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


      
      <div class="grid grid-cols-1 gap-4">
        <div class="space-y-1 w-full md:max-w-md">
          <label class="text-[11px] font-semibold text-colpsi-red uppercase tracking-wide ml-1 mb-1">
            Contraseña Actual <span class="lowercase font-medium text-colpsi-muted">(obligatoria para guardar)</span>
          </label>

          <PasswordInputComponent
            required
            value={props.password}
            // Pasamos el valor directamente al handler como esperan los otros componentes
            onInput={(e) => props.onPasswordChange(e.currentTarget.value)}
            class="h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm text-colpsi-text placeholder:text-slate-400 outline-none transition-colors focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15"
            placeholder="Introduce tu contraseña para confirmar"
          />

          <div class="flex items-start gap-2 mt-2 ml-1">
            <span class="text-[11px] text-colpsi-muted leading-tight">
              Esta medida protege tu cuenta. Ningún cambio será procesado por el servidor sin esta validación.
            </span>
          </div>
        </div>
      </div>
    </section>
  );
}