// web/src/components/psi/profile/InputField.tsx
// Componente utilitario para inputs
interface InputFieldProps {
  label: string;
  value: string;
  onInput: (value: string) => void;
  type?: string;
  placeholder?: string;
}

export function InputField(props: InputFieldProps) {
  return (
    <div class="space-y-2">
      <label class="text-xs font-bold text-gray-500 uppercase ml-2">{props.label}</label>
      <input
        type={props.type || "text"}
        value={props.value}
        onInput={(e) => props.onInput(e.currentTarget.value)}
        placeholder={props.placeholder}
        class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all"
      />
    </div>
  );
}

// web/src/components/psi/profile/PasswordField.tsx
interface PasswordFieldProps {
  label: string;
  value: string;
  onInput: (value: string) => void;
  placeholder?: string;
}

export function PasswordField(props: PasswordFieldProps) {
  return (
    <div class="space-y-2">
      <label class="text-xs font-bold text-gray-500 uppercase ml-2">{props.label}</label>
      <input
        type="password"
        value={props.value}
        onInput={(e) => props.onInput(e.currentTarget.value)}
        placeholder={props.placeholder}
        class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all"
      />
    </div>
  );
}