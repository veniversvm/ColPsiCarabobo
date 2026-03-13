// web/src/components/admin/psicologos/create/InstitutionalStatusSection.tsx
import { ToggleSwitch } from "~/components/ui/ToggleSwitch";
import { PsicologoForm } from "~/types/admin";

interface Props {
  form: any;
  setForm: <K extends keyof PsicologoForm>(field: K, value: PsicologoForm[K]) => void;
  today: string;
}

export function InstitutionalStatusSection(props: Props) {
  const inputClass = "w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-blue rounded-xl px-4 py-2.5 outline-none";

  return (
    <section class="bg-white rounded-[2rem] p-6 shadow-sm border border-gray-100">
      <h2 class="text-lg font-black text-colpsi-blue mb-4 border-l-4 border-colpsi-yellow pl-3">Estatus Institucional</h2>
      
      <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
        <div class="bg-gray-50 p-5 rounded-2xl space-y-3 border border-gray-200">
          <h3 class="text-xs font-bold text-gray-500 uppercase tracking-widest border-b pb-2">Estado Principal</h3>
          <ToggleSwitch 
            label="Cuenta Activa (Acceso al sistema)" 
            checked={props.form.is_active} 
            onChange={(v) => props.setForm("is_active", v)} 
          />
          <ToggleSwitch 
            label="Miembro Solvente (Al día con pagos)" 
            checked={props.form.solvent} 
            onChange={(v) => props.setForm("solvent", v)} 
          />
          <ToggleSwitch 
            label="Fe de Vida Activa" 
            checked={props.form.proof_of_life} 
            onChange={(v) => props.setForm("proof_of_life", v)} 
          />
          
          <div class="pt-2">
            <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Fecha Última Solvencia <span class="text-red-500">*</span></label>
            <input 
              type="date" 
              required 
              value={props.form.date_of_last_solvency} 
              onInput={(e) => props.setForm("date_of_last_solvency", e.currentTarget.value)} 
              class={inputClass} 
            />
          </div>
        </div>

        <div class="bg-blue-50/50 p-5 rounded-2xl space-y-3 border border-blue-100">
          <h3 class="text-xs font-bold text-colpsi-blue uppercase tracking-widest border-b border-blue-100 pb-2">Roles Gremiales</h3>
          <ToggleSwitch 
            label="Director del Gremio" 
            checked={props.form.guild_director} 
            onChange={(v) => props.setForm("guild_director", v)} 
          />
          <ToggleSwitch 
            label="Colaborador del Gremio" 
            checked={props.form.guild_collaborator} 
            onChange={(v) => props.setForm("guild_collaborator", v)} 
          />
          <ToggleSwitch 
            label="Profesor Universitario" 
            checked={props.form.university_professor} 
            onChange={(v) => props.setForm("university_professor", v)} 
          />
          <ToggleSwitch 
            label="Empleado Público" 
            checked={props.form.public_employee} 
            onChange={(v) => props.setForm("public_employee", v)} 
          />
          <ToggleSwitch 
            label="Doble Gremio" 
            checked={props.form.double_guild} 
            onChange={(v) => props.setForm("double_guild", v)} 
          />
          <ToggleSwitch 
            label="Beneficio 65+ Años" 
            checked={props.form.sixty_five_or_plus} 
            onChange={(v) => props.setForm("sixty_five_or_plus", v)} 
          />
        </div>
      </div>
    </section>
  );
}