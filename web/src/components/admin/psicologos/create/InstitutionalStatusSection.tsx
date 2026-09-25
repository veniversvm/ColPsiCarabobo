// web/src/components/admin/psicologos/create/InstitutionalStatusSection.tsx
import { ToggleSwitch } from "~/components/ui/ToggleSwitch";
import { PsicologoForm } from "~/types/admin";
import FlatDatePicker from "~/components/ui/FlatDatePicker";

interface Props {
  form: any;
  setForm: <K extends keyof PsicologoForm>(field: K, value: PsicologoForm[K]) => void;
  today: string;
}

export function InstitutionalStatusSection(props: Props) {
  const inputClass = "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors";

  return (
    <section class="bg-white rounded-lg p-5 border border-colpsi-border">
      <h2 class="text-base font-semibold text-colpsi-text mb-4">Estatus Institucional</h2>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="bg-colpsi-bg p-4 rounded-md space-y-3 border border-colpsi-border">
          <h3 class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide border-b border-colpsi-border pb-2">Estado Principal</h3>
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
            <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Fecha Última Solvencia <span class="text-red-500">*</span></label>
            <FlatDatePicker
              value={props.form.date_of_last_solvency}
              onChange={(v) => props.setForm("date_of_last_solvency", v)}
              maxDate={props.today}
              class={inputClass}
            />
          </div>
        </div>

        <div class="bg-sky-50/50 p-4 rounded-md space-y-3 border border-sky-100">
          <h3 class="text-[11px] font-semibold text-colpsi-blue uppercase tracking-wide border-b border-sky-100 pb-2">Roles Gremiales</h3>
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