// web/src/components/psi/profile/InputField.tsx
import { createSignal, Show } from "solid-js";

interface InputFieldProps {
  label: string;
  value: string;
  onInput: (value: string) => void;
  type?: string;
  placeholder?: string;
  required?: boolean;
  error?: string;
  helper?: string;
}

export function InputField(props: InputFieldProps) {
  return (
    <div class="space-y-1">
      <div class="flex items-center justify-between">
        <label class="text-xs font-bold text-gray-500 uppercase ml-2">
          {props.label}
          {props.required && <span class="text-red-500 ml-1">*</span>}
        </label>
        <Show when={props.helper}>
          <span class="text-[10px] text-gray-400">{props.helper}</span>
        </Show>
      </div>
      <input
        type={props.type || "text"}
        value={props.value}
        onInput={(e) => props.onInput(e.currentTarget.value)}
        placeholder={props.placeholder}
        class={`w-full bg-gray-50 border-2 rounded-xl px-5 py-3 outline-none transition-all ${
          props.error
            ? "border-red-300 focus:border-red-500"
            : "border-transparent focus:border-colpsi-yellow"
        }`}
      />
      <Show when={props.error}>
        <p class="text-xs text-red-500 mt-1 ml-2">{props.error}</p>
      </Show>
    </div>
  );
}

interface PasswordFieldProps {
  label: string;
  value: string;
  onInput: (value: string) => void;
  placeholder?: string;
  required?: boolean;
  error?: string;
  showStrength?: boolean;
}

// Ojo abierto: contraseña visible
function IconEyeOpen() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
    >
      <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  );
}

// Ojo tachado: contraseña oculta
function IconEyeOff() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
    >
      <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24" />
      <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68" />
      <path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61" />
      <line x1="2" x2="22" y1="2" y2="22" />
    </svg>
  );
}

export function PasswordField(props: PasswordFieldProps) {
  const [showPassword, setShowPassword] = createSignal(false);

  return (
    <div class="space-y-1">
      <label class="text-xs font-bold text-gray-500 uppercase ml-2">
        {props.label}
        {props.required && <span class="text-red-500 ml-1">*</span>}
      </label>

      <div class="relative">
        <input
          type={showPassword() ? "text" : "password"}
          value={props.value}
          onInput={(e) => props.onInput(e.currentTarget.value)}
          placeholder={props.placeholder}
          class={`w-full bg-gray-50 border-2 rounded-xl px-5 py-3 outline-none transition-all font-mono pr-12 ${
            props.error
              ? "border-red-300 focus:border-red-500"
              : "border-transparent focus:border-colpsi-yellow"
          }`}
        />

        {/* Botón para mostrar/ocultar contraseña */}
        <button
          type="button"
          onClick={() => setShowPassword(!showPassword())}
          class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-colpsi-blue transition-colors p-1"
          tabIndex={-1}
          title={showPassword() ? "Ocultar contraseña" : "Mostrar contraseña"}
        >
          {showPassword() ? <IconEyeOpen /> : <IconEyeOff />}
        </button>
      </div>

      <Show when={props.error}>
        <p class="text-xs text-red-500 mt-1 ml-2">{props.error}</p>
      </Show>
    </div>
  );
}