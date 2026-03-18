// web/src/components/admin/psicologos/edit/sections/AdminStatusSection.tsx

import { ToggleSwitch } from "~/components/ui/ToggleSwitch";
import { Field, IC } from "../EditPrimitives";
import type { EditFormState } from "../types";

interface Props {
  form: EditFormState;
  setForm: (key: keyof EditFormState, value: any) => void;
}

export function AdminStatusSection(props: Props) {
  return (
    <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-blue-100 relative overflow-hidden">
      <div class="absolute top-0 left-0 w-2 h-full bg-yellow-400" />
      <h2 class="text-lg font-black text-blue-800 mb-6 ml-2">Estatus Administrativo</h2>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 ml-2">

        {/* Cuenta + Solvencia */}
        <div class="space-y-4 bg-gray-50 p-6 rounded-2xl border border-gray-100">
          <ToggleSwitch
            label="Cuenta Activa en Sistema"
            checked={props.form.is_active}
            onChange={(v) => props.setForm("is_active", v)}
          />
          <ToggleSwitch
            label="Estado: Solvente"
            checked={props.form.solvent}
            onChange={(v) => props.setForm("solvent", v)}
          />
          <ToggleSwitch
            label="Fe de Vida Activa"
            checked={props.form.proof_of_life}
            onChange={(v) => props.setForm("proof_of_life", v)}
          />
          <div class="pt-3">
            <Field label="Fecha Última Solvencia">
              <input
                type="date"
                value={props.form.date_of_last_solvency}
                onInput={(e) => props.setForm("date_of_last_solvency", e.currentTarget.value)}
                class={IC}
              />
            </Field>
          </div>
        </div>

        {/* Roles Gremiales */}
        <div class="space-y-3 bg-blue-50/50 p-6 rounded-2xl border border-blue-100">
          <h3 class="text-xs font-bold text-blue-800 uppercase tracking-widest border-b border-blue-100 pb-2 mb-3">
            Roles Gremiales
          </h3>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
            <ToggleSwitch label="Director"             checked={props.form.guild_director}       onChange={(v) => props.setForm("guild_director",       v)} />
            <ToggleSwitch label="Colaborador"          checked={props.form.guild_collaborator}   onChange={(v) => props.setForm("guild_collaborator",   v)} />
            <ToggleSwitch label="Prof. Universitario"  checked={props.form.university_professor} onChange={(v) => props.setForm("university_professor", v)} />
            <ToggleSwitch label="Empleado Público"     checked={props.form.public_employee}      onChange={(v) => props.setForm("public_employee",      v)} />
            <ToggleSwitch label="Doble Gremio"         checked={props.form.double_guild}         onChange={(v) => props.setForm("double_guild",         v)} />
            <ToggleSwitch label="65+ Años"             checked={props.form.sixty_five_or_plus}   onChange={(v) => props.setForm("sixty_five_or_plus",   v)} />
            <ToggleSwitch label="CPSM"                 checked={props.form.cpsm}                 onChange={(v) => props.setForm("cpsm",                 v)} />
          </div>
        </div>

      </div>
    </section>
  );
}