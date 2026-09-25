import { PsicologoForm } from "~/types/admin";
import FlatDatePicker from "~/components/ui/FlatDatePicker";

// web/src/components/admin/psicologos/create/LegalIdentitySection.tsx
interface Props {
  form: any;
  setForm: <K extends keyof PsicologoForm>(field: K, value: PsicologoForm[K]) => void;
}

export function LegalIdentitySection(props: Props) {
  const inputClass = "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors";

  return (
    <section class="bg-white rounded-lg p-5 border border-colpsi-border">
      <h2 class="text-base font-semibold text-colpsi-text mb-4">Identidad Legal</h2>
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div class="space-y-1 md:col-span-2">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Primer Nombre <span class="text-red-500">*</span></label>
          <input
            type="text"
            required
            value={props.form.first_name}
            onInput={(e) => props.setForm("first_name", e.currentTarget.value)}
            class={inputClass}
          />
        </div>
        <div class="space-y-1 md:col-span-2">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Segundo Nombre</label>
          <input
            type="text"
            value={props.form.second_name}
            onInput={(e) => props.setForm("second_name", e.currentTarget.value)}
            class={inputClass}
          />
        </div>
        <div class="space-y-1 md:col-span-2">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Primer Apellido <span class="text-red-500">*</span></label>
          <input
            type="text"
            required
            value={props.form.last_name}
            onInput={(e) => props.setForm("last_name", e.currentTarget.value)}
            class={inputClass}
          />
        </div>
        <div class="space-y-1 md:col-span-2">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Segundo Apellido</label>
          <input
            type="text"
            value={props.form.second_last_name}
            onInput={(e) => props.setForm("second_last_name", e.currentTarget.value)}
            class={inputClass}
          />
        </div>

        <div class="space-y-1">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Nacionalidad <span class="text-red-500">*</span></label>
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
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Cédula <span class="text-red-500">*</span></label>
          <input
            type="number"
            required
            value={props.form.ci}
            onInput={(e) => props.setForm("ci", e.currentTarget.value)}
            class={inputClass}
          />
        </div>
        <div class="space-y-1">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Nro. FPV <span class="text-red-500">*</span></label>
          <input
            type="number"
            required
            value={props.form.fpv}
            onInput={(e) => props.setForm("fpv", e.currentTarget.value)}
            class="h-9 w-full rounded-md border border-amber-300 bg-amber-50 px-3 text-sm font-semibold text-colpsi-blue outline-none focus:border-amber-400 focus:ring-2 focus:ring-amber-400/15 transition-colors"
          />
        </div>
        <div class="space-y-1">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Género <span class="text-red-500">*</span></label>
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
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Fecha de Nacimiento <span class="text-red-500">*</span></label>
          <FlatDatePicker
            value={props.form.born_date}
            onChange={(v) => props.setForm("born_date", v)}
            class={inputClass}
          />
        </div>
      </div>
    </section>
  );
}