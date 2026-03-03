import { InputField, PasswordField } from "./InputField";

// web/src/components/psi/profile/AccountSection.tsx
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
        <PasswordField
          label="Nueva Contraseña (Opcional)"
          value={props.newPassword1}
          onInput={props.onNewPassword1Change}
          placeholder="Dejar en blanco para no cambiar"
        />
        <PasswordField
          label="Confirmar Nueva Contraseña"
          value={props.newPassword2}
          onInput={props.onNewPassword2Change}
          placeholder="••••••••"
        />
      </div>
    </section>
  );
}