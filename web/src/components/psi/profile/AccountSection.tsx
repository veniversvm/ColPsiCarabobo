// web/src/components/psi/profile/AccountSection.tsx
import { createMemo, Show } from "solid-js";
import { InputField, PasswordField } from "./InputField";

interface AccountSectionProps {
  username: string;
  email: string;
  newPassword1: string;
  newPassword2: string;
  onUsernameChange: (value: string) => void;
  onEmailChange: (value: string) => void;
  onNewPassword1Change: (value: string) => void;
  onNewPassword2Change: (value: string) => void;
}

export function AccountSection(props: AccountSectionProps) {
  // Validaciones memorizadas
  const passwordsMatch = createMemo(() => {
    if (!props.newPassword1 && !props.newPassword2) return true;
    return props.newPassword1 === props.newPassword2;
  });

  const hasPasswordError = createMemo(() => {
    return props.newPassword1 || props.newPassword2 
      ? !passwordsMatch()
      : false;
  });

  const passwordStrength = createMemo(() => {
    if (!props.newPassword1) return null;
    
    const password = props.newPassword1;
    let strength = 0;
    
    if (password.length >= 8) strength++;
    if (/[A-Z]/.test(password)) strength++;
    if (/[a-z]/.test(password)) strength++;
    if (/[0-9]/.test(password)) strength++;
    if (/[^A-Za-z0-9]/.test(password)) strength++;
    
    return strength;
  });

  const strengthColor = createMemo(() => {
    const strength = passwordStrength();
    if (!strength) return "";
    if (strength <= 2) return "text-red-500";
    if (strength <= 3) return "text-yellow-500";
    if (strength <= 4) return "text-green-500";
    return "text-colpsi-blue";
  });

  const strengthLabel = createMemo(() => {
    const strength = passwordStrength();
    if (!strength) return "";
    if (strength <= 2) return "Débil";
    if (strength <= 3) return "Media";
    if (strength <= 4) return "Fuerte";
    return "Muy Fuerte";
  });

  return (
    <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100">
      <h2 class="text-xl font-black text-colpsi-blue mb-6 border-l-4 border-colpsi-yellow pl-3">
        Datos de Cuenta
      </h2>
      
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <InputField
          label="Nombre de Usuario (Para Login)"
          value={props.username}
          onInput={props.onUsernameChange}
        />
        
        <InputField
          label="Email Principal (Institucional)"
          type="email"
          value={props.email}
          onInput={props.onEmailChange}
        />

        {/* Nueva Contraseña */}
        <div class="space-y-2">
          <PasswordField
            label="Nueva Contraseña (Opcional)"
            value={props.newPassword1}
            onInput={props.onNewPassword1Change}
            placeholder="••••••••"
          />
          
          {/* Indicador de fortaleza */}
          <Show when={props.newPassword1}>
            <div class="flex items-center gap-2 text-xs mt-1">
              <div class="flex-1 h-1.5 bg-gray-200 rounded-full overflow-hidden">
                <div 
                  class={`h-full transition-all duration-300 ${
                    (passwordStrength() ?? 0) <= 2 ? "bg-red-500" :
                    (passwordStrength() ?? 0) <= 3 ? "bg-yellow-500" :
                    (passwordStrength() ?? 0) <= 4 ? "bg-green-500" :
                    "bg-colpsi-blue"
                  }`}
                  style={{ width: `${(passwordStrength() ?? 0) * 20}%` }}
                />
              </div>
              <span class={`font-medium ${strengthColor()}`}>
                {strengthLabel()}
              </span>
            </div>
          </Show>
        </div>

        {/* Confirmar Contraseña */}
        <div class="space-y-2">
          <PasswordField
            label="Confirmar Nueva Contraseña"
            value={props.newPassword2}
            onInput={props.onNewPassword2Change}
            placeholder="••••••••"
          />
          
          <Show when={hasPasswordError()}>
            <div class="flex items-center gap-1 text-xs text-red-500 mt-1 animate-in fade-in slide-in-from-top-1">
              <span class="text-red-500">⚠️</span>
              <span>Las contraseñas no coinciden</span>
            </div>
          </Show>

          <Show when={passwordsMatch() && props.newPassword1 && props.newPassword2}>
            <div class="flex items-center gap-1 text-xs text-green-500 mt-1 animate-in fade-in slide-in-from-top-1">
              <span class="text-green-500">✅</span>
              <span>Las contraseñas coinciden</span>
            </div>
          </Show>
        </div>
      </div>

      {/* Requisitos de seguridad */}
      <Show when={props.newPassword1}>
        <div class="mt-6 p-4 bg-gray-50 rounded-xl border border-gray-100">
          <p class="text-xs font-bold text-gray-500 uppercase tracking-wider mb-3">
            Requisitos de seguridad
          </p>
          <div class="grid grid-cols-2 md:grid-cols-5 gap-3">
            <RequirementCheck 
              label="8+ caracteres" 
              met={props.newPassword1.length >= 8} 
            />
            <RequirementCheck 
              label="Mayúscula" 
              met={/[A-Z]/.test(props.newPassword1)} 
            />
            <RequirementCheck 
              label="Minúscula" 
              met={/[a-z]/.test(props.newPassword1)} 
            />
            <RequirementCheck 
              label="Número" 
              met={/[0-9]/.test(props.newPassword1)} 
            />
            <RequirementCheck 
              label="Símbolo" 
              met={/[^A-Za-z0-9]/.test(props.newPassword1)} 
            />
          </div>
        </div>
      </Show>
    </section>
  );
}

// Componente auxiliar para los checks de requisitos
function RequirementCheck(props: { label: string; met: boolean }) {
  return (
    <div class={`flex items-center gap-1.5 text-xs ${
      props.met ? 'text-green-600' : 'text-gray-400'
    }`}>
      <span class="text-sm">
        {props.met ? '✅' : '◻️'}
      </span>
      <span class={props.met ? 'font-medium' : ''}>
        {props.label}
      </span>
    </div>
  );
}