import { PsicologoForm } from "~/types/admin";

// web/src/components/admin/psicologos/create/AcademicRegistrationSection.tsx
interface Props {
  form: any;
  setForm: <K extends keyof PsicologoForm>(field: K, value: PsicologoForm[K]) => void;
}

export function AcademicRegistrationSection(props: Props) {
  const inputClass = "w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-blue rounded-xl px-4 py-2.5 outline-none";

  return (
    <section class="bg-white rounded-[2rem] p-6 shadow-sm border border-gray-100">
      <h2 class="text-lg font-black text-colpsi-blue mb-4 border-l-4 border-colpsi-yellow pl-3">Registro Académico Base</h2>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="space-y-1 md:col-span-2">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Universidad de Egreso <span class="text-red-500">*</span></label>
          <input 
            type="text" 
            required 
            value={props.form.university_undergraduate} 
            onInput={(e) => props.setForm("university_undergraduate", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>
        <div class="space-y-1">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Fecha de Egreso <span class="text-red-500">*</span></label>
          <input 
            type="date" 
            required 
            value={props.form.graduate_date} 
            onInput={(e) => props.setForm("graduate_date", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>
        <div class="space-y-1 md:col-span-3">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Mención</label>
          <input 
            type="text" 
            value={props.form.mention_undergraduate} 
            onInput={(e) => props.setForm("mention_undergraduate", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>

        <div class="col-span-full mt-4"><hr class="border-gray-100"/></div>
        
        <div class="space-y-1">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Estado de Registro <span class="text-red-500">*</span></label>
          <input 
            type="text" 
            required 
            value={props.form.register_title_state} 
            onInput={(e) => props.setForm("register_title_state", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>
        <div class="space-y-1">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Fecha de Registro <span class="text-red-500">*</span></label>
          <input 
            type="date" 
            required 
            value={props.form.register_title_date} 
            onInput={(e) => props.setForm("register_title_date", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>
        <div class="space-y-1">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Número de Registro <span class="text-red-500">*</span></label>
          <input 
            type="number" 
            required 
            value={props.form.register_number} 
            onInput={(e) => props.setForm("register_number", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>
        <div class="space-y-1">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Folio <span class="text-red-500">*</span></label>
          <input 
            type="text" 
            required 
            value={props.form.register_folio} 
            onInput={(e) => props.setForm("register_folio", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>
        <div class="space-y-1">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Tomo <span class="text-red-500">*</span></label>
          <input 
            type="text" 
            required 
            value={props.form.register_tome} 
            onInput={(e) => props.setForm("register_tome", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>
      </div>
    </section>
  );
}