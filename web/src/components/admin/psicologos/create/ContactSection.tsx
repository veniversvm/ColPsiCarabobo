import { PsicologoForm } from "~/types/admin";

// web/src/components/admin/psicologos/create/ContactSection.tsx
interface Props {
  form: any;
  setForm: <K extends keyof PsicologoForm>(field: K, value: PsicologoForm[K]) => void;
}

export function ContactSection(props: Props) {
  const inputClass = "w-full bg-colpsi-surface border-2 border-transparent focus:border-colpsi-blue rounded-xl px-4 py-2.5 outline-none";

  return (
    <section class="bg-white rounded-[2rem] p-6 shadow-sm border border-colpsi-border">
      <h2 class="text-lg font-black text-colpsi-blue mb-4 border-l-4 border-gray-300 pl-3">Datos de Contacto</h2>
      <p class="text-xs text-gray-500 mb-4 bg-colpsi-surface p-3 rounded-xl border border-colpsi-border">
        Información requerida para mantener comunicación oficial con el agremiado.
      </p>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="space-y-1 md:col-span-2">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Teléfono Fijo <span class="text-red-500">*</span></label>
          <input 
            type="tel" 
            required 
            value={props.form.public_phone} 
            onInput={(e) => props.setForm("public_phone", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>
        <div class="space-y-1 md:col-span-2">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Dirección Exacta <span class="text-red-500">*</span></label>
          <input 
            type="text" 
            required 
            value={props.form.service_address} 
            onInput={(e) => props.setForm("service_address", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>
      </div>
    </section>
  );
}