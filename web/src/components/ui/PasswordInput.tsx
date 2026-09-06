// web/src/components/ui/PasswordInput.tsx
import { createSignal, JSX, splitProps, Show } from "solid-js";

// Extendemos los atributos estándar de un input de HTML
interface PasswordInputProps extends JSX.InputHTMLAttributes<HTMLInputElement> {
  class?: string;
}

export function PasswordInputComponent(props: PasswordInputProps) {
  // Separamos la clase personalizada del resto de las propiedades para inyectarla con cuidado
  const [local, rest] = splitProps(props,["class"]);
  
  // Estado local para alternar el tipo de input
  const [show, setShow] = createSignal(false);

  return (
    <div class="relative w-full flex items-center">
      <input
        type={show() ? "text" : "password"}
        // Aplicamos un pr-12 (padding right) para que el texto no se monte sobre el ícono
        class={`w-full pr-12 outline-none transition-all ${local.class || "bg-colpsi-surface border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 text-colpsi-text"}`}
        {...rest}
      />
      
      {/* Botón Toggle */}
      <button
        type="button" // CRÍTICO: type="button" evita que presionar Enter envíe el formulario
        class="absolute right-2 p-2 text-gray-400 hover:text-colpsi-blue transition-colors focus:outline-none rounded-lg focus:ring-2 focus:ring-colpsi-yellow/50"
        onClick={() => setShow(!show())}
        title={show() ? "Ocultar contraseña" : "Mostrar contraseña"}
        aria-label={show() ? "Ocultar contraseña" : "Mostrar contraseña"}
      >
        <Show 
          when={show()} 
          fallback={
            // Ícono de "Ojo Abierto" (Mostrar)
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
            </svg>
          }
        >
          {/* Ícono de "Ojo Cerrado" (Ocultar) */}
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
          </svg>
        </Show>
      </button>
    </div>
  );
}