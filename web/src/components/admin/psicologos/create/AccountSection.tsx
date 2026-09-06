import { SetStoreFunction } from "solid-js/store";
import { PasswordInputComponent } from "~/components/ui/PasswordInput";
import { PsicologoForm } from "~/types/admin";


interface Props {
  form: PsicologoForm;
  setForm: <K extends keyof PsicologoForm>(field: K, value: PsicologoForm[K]) => void;
}

export function AccountSection(props: Props) {
  const inputClass = "w-full bg-colpsi-surface border-2 border-transparent focus:border-colpsi-blue rounded-xl px-4 py-2.5 outline-none";

  return (
    <section class="bg-white rounded-[2rem] p-6 shadow-sm border border-colpsi-border">
      <h2 class="text-lg font-black text-colpsi-blue mb-4 border-l-4 border-colpsi-yellow pl-3">Cuenta y Acceso</h2>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="space-y-1">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Usuario <span class="text-red-500">*</span></label>
          <input 
            type="text" 
            required 
            value={props.form.username} 
            onInput={(e) => props.setForm("username", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>
        <div class="space-y-1">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Email <span class="text-red-500">*</span></label>
          <input 
            type="email" 
            required 
            value={props.form.email} 
            onInput={(e) => props.setForm("email", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>
        <div class="space-y-1">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Contraseña Inicial <span class="text-red-500">*</span></label>
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