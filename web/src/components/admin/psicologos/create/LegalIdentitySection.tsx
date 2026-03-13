import { PsicologoForm } from "~/types/admin";

// web/src/components/admin/psicologos/create/LegalIdentitySection.tsx
interface Props {
  form: any;
  setForm: <K extends keyof PsicologoForm>(field: K, value: PsicologoForm[K]) => void;
}

export function LegalIdentitySection(props: Props) {
  const inputClass = "w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-blue rounded-xl px-4 py-2.5 outline-none";

  return (
    <section class="bg-white rounded-[2rem] p-6 shadow-sm border border-gray-100">
      <h2 class="text-lg font-black text-colpsi-blue mb-4 border-l-4 border-colpsi-yellow pl-3">Identidad Legal</h2>
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div class="space-y-1 md:col-span-2">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Primer Nombre <span class="text-red-500">*</span></label>
          <input 
            type="text" 
            required 
            value={props.form.first_name} 
            onInput={(e) => props.setForm("first_name", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>
        <div class="space-y-1 md:col-span-2">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Segundo Nombre</label>
          <input 
            type="text" 
            value={props.form.second_name} 
            onInput={(e) => props.setForm("second_name", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>
        <div class="space-y-1 md:col-span-2">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Primer Apellido <span class="text-red-500">*</span></label>
          <input 
            type="text" 
            required 
            value={props.form.last_name} 
            onInput={(e) => props.setForm("last_name", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>
        <div class="space-y-1 md:col-span-2">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Segundo Apellido</label>
          <input 
            type="text" 
            value={props.form.second_last_name} 
            onInput={(e) => props.setForm("second_last_name", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>

        <div class="space-y-1">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Nacionalidad <span class="text-red-500">*</span></label>
          <select 
            required 
            value={props.form.nationality} 
            onChange={(e) => props.setForm("nationality", e.currentTarget.value)} 
            class={inputClass}
          >
            <option value="V">V - Venezolano</option>
            <option value="E">E - Extranjero</option>
          </select>
        </div>
        <div class="space-y-1">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Cédula <span class="text-red-500">*</span></label>
          <input 
            type="number" 
            required 
            value={props.form.ci} 
            onInput={(e) => props.setForm("ci", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>
        <div class="space-y-1">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Nro. FPV <span class="text-red-500">*</span></label>
          <input 
            type="number" 
            required 
            value={props.form.fpv} 
            onInput={(e) => props.setForm("fpv", e.currentTarget.value)} 
            class="w-full bg-colpsi-yellow/20 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-4 py-2.5 outline-none font-bold text-colpsi-blue" 
          />
        </div>
        <div class="space-y-1">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Género <span class="text-red-500">*</span></label>
          <select 
            required 
            value={props.form.genre} 
            onChange={(e) => props.setForm("genre", e.currentTarget.value)} 
            class={inputClass}
          >
            <option value="M">Masculino</option>
            <option value="F">Femenino</option>
          </select>
        </div>
        <div class="space-y-1 md:col-span-2">
          <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Fecha de Nacimiento <span class="text-red-500">*</span></label>
          <input 
            type="date" 
            required 
            value={props.form.born_date} 
            onInput={(e) => props.setForm("born_date", e.currentTarget.value)} 
            class={inputClass} 
          />
        </div>
      </div>
    </section>
  );
}