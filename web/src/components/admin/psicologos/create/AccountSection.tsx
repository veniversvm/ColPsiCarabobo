import { SetStoreFunction } from "solid-js/store";
import { PasswordInputComponent } from "~/components/ui/PasswordInput";
import { PsicologoForm } from "~/types/admin";


interface Props {
  form: PsicologoForm;
  setForm: <K extends keyof PsicologoForm>(field: K, value: PsicologoForm[K]) => void;
}

export function AccountSection(props: Props) {
  const inputClass = "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors";

  return (
    <section class="bg-white rounded-lg p-5 border border-colpsi-border">
      <h2 class="text-base font-semibold text-colpsi-text mb-4">Cuenta y Acceso</h2>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="space-y-1">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Usuario <span class="text-red-500">*</span></label>
          <input
            type="text"
            required
            value={props.form.username}
            onInput={(e) => props.setForm("username", e.currentTarget.value)}
            class={inputClass}
          />
        </div>
        <div class="space-y-1">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Email <span class="text-red-500">*</span></label>
          <input
            type="email"
            required
            value={props.form.email}
            onInput={(e) => props.setForm("email", e.currentTarget.value)}
            class={inputClass}
          />
        </div>
        <div class="space-y-1">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Contraseña Inicial <span class="text-red-500">*</span></label>
          <PasswordInputComponent
            required
            value={props.form.password}
            onInput={(e: any) => props.setForm("password", e.currentTarget.value)}
            class={inputClass}
          />
        </div>
      </div>
    </section>
  );
}