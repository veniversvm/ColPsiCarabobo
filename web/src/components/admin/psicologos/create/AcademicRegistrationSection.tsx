import { PsicologoForm } from "~/types/admin";
import FlatDatePicker from "~/components/ui/FlatDatePicker";

// web/src/components/admin/psicologos/create/AcademicRegistrationSection.tsx
interface Props {
  form: any;
  setForm: <K extends keyof PsicologoForm>(field: K, value: PsicologoForm[K]) => void;
}

export function AcademicRegistrationSection(props: Props) {
  const inputClass = "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors";

  return (
    <section class="bg-white rounded-lg p-5 border border-colpsi-border">
      <h2 class="text-base font-semibold text-colpsi-text mb-4">Registro Académico Base</h2>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="space-y-1 md:col-span-2">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Universidad de Egreso <span class="text-red-500">*</span></label>
          <input
            type="text"
            required
            value={props.form.university_undergraduate}
            onInput={(e) => props.setForm("university_undergraduate", e.currentTarget.value)}
            class={inputClass}
          />
        </div>
        <div class="space-y-1">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Fecha de Egreso <span class="text-red-500">*</span></label>
          <FlatDatePicker
            value={props.form.graduate_date}
            onChange={(v) => props.setForm("graduate_date", v)}
            class={inputClass}
          />
        </div>
        <div class="space-y-1 md:col-span-3">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Mención</label>
          <input
            type="text"
            value={props.form.mention_undergraduate}
            onInput={(e) => props.setForm("mention_undergraduate", e.currentTarget.value)}
            class={inputClass}
          />
        </div>

        <div class="col-span-full mt-4"><hr class="border-colpsi-border"/></div>

        <div class="space-y-1">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Estado de Registro <span class="text-red-500">*</span></label>
          <input
            type="text"
            required
            value={props.form.register_title_state}
            onInput={(e) => props.setForm("register_title_state", e.currentTarget.value)}
            class={inputClass}
          />
        </div>
        <div class="space-y-1">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Fecha de Registro <span class="text-red-500">*</span></label>
          <FlatDatePicker
            value={props.form.register_title_date}
            onChange={(v) => props.setForm("register_title_date", v)}
            class={inputClass}
          />
        </div>
        <div class="space-y-1">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Número de Registro <span class="text-red-500">*</span></label>
          <input
            type="number"
            required
            value={props.form.register_number}
            onInput={(e) => props.setForm("register_number", e.currentTarget.value)}
            class={inputClass}
          />
        </div>
        <div class="space-y-1">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Folio <span class="text-red-500">*</span></label>
          <input
            type="text"
            required
            value={props.form.register_folio}
            onInput={(e) => props.setForm("register_folio", e.currentTarget.value)}
            class={inputClass}
          />
        </div>
        <div class="space-y-1">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Tomo <span class="text-red-500">*</span></label>
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