// web/src/components/admin/psicologos/edit/sections/AdminStatusSection.tsx

import { ToggleSwitch } from "~/components/ui/ToggleSwitch";
import { Field, IC } from "../EditPrimitives";
import FlatDatePicker from "~/components/ui/FlatDatePicker";
import type { EditFormState } from "../types";

interface Props {
  form: EditFormState;
  setForm: (key: keyof EditFormState, value: any) => void;
}

// Requisitos legales de activación (Art. 5 Ley de Ejercicio + Art. 18 Estatutos FPV)
function ActivationSection(props: { form: EditFormState }) {
  const reqs = () => [
    {
      label: "N° de FPV asignado",
      ok: !!props.form.fpv,
    },
    {
      label: "Psicólogo solvente",
      ok: !!props.form.solvent,
    },
  ];
  const allOk = () => reqs().every((r) => r.ok);

  return (
    <div class={`pt-3 border-t ${allOk() ? "border-green-200" : "border-amber-200"}`}>
      <div class="flex items-center justify-between mb-2">
        <h3 class="text-xs font-bold text-gray-600 uppercase tracking-widest">
          Verificaciones de Activación
        </h3>
        <span class={`text-[10px] font-black px-2 py-0.5 rounded-full ${allOk() ? "bg-green-100 text-green-700" : "bg-amber-100 text-amber-700"}`}>
          {allOk() ? "Completos" : "Pendientes"}
        </span>
      </div>
      <ul class="space-y-1.5">
        {reqs().map((r) => (
          <li class="flex items-center gap-2 text-xs font-semibold text-gray-600">
            <span class={`w-4 h-4 rounded-full flex items-center justify-center text-[9px] font-black ${r.ok ? "bg-green-100 text-green-700" : "bg-amber-100 text-amber-700"}`}>
              {r.ok ? "✓" : "!"}
            </span>
            {r.label}
          </li>
        ))}
      </ul>
      <p class="mt-2 text-[10px] text-gray-400 leading-relaxed">
        Confirmaciones que realiza la administración con el expediente en mano
        antes de activar la cuenta. El N° de FPV acredita la inscripción
        ministerial; la solvencia nace al aprobar la inscripción
        (Art. 5 Ley de Ejercicio de la Psicología · Art. 18 Estatutos FPV).
      </p>
    </div>
  );
}

export function AdminStatusSection(props: Props) {
  return (
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">

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
              <FlatDatePicker
                value={props.form.date_of_last_solvency}
                onChange={(v) => props.setForm("date_of_last_solvency", v)}
                class={IC}
              />
            </Field>
          </div>
          <ActivationSection form={props.form} />
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
  );
}