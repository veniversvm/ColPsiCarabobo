import { PsicologoForm } from "~/types/admin";

// web/src/components/admin/psicologos/create/ContactSection.tsx
interface Props {
  form: any;
  setForm: <K extends keyof PsicologoForm>(field: K, value: PsicologoForm[K]) => void;
}

export function ContactSection(props: Props) {
  const inputClass = "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors";

  return (
    <section class="bg-white rounded-lg p-5 border border-colpsi-border">
      <h2 class="text-base font-semibold text-colpsi-text mb-4">Datos de Contacto</h2>
      <p class="text-xs text-colpsi-muted mb-4 bg-colpsi-bg p-3 rounded-md border border-colpsi-border">
        Información requerida para mantener comunicación oficial con el agremiado.
      </p>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="space-y-1 md:col-span-2">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Teléfono Fijo <span class="text-red-500">*</span></label>
          <input
            type="tel"
            required
            value={props.form.public_phone}
            onInput={(e) => props.setForm("public_phone", e.currentTarget.value)}
            class={inputClass}
          />
        </div>
        <div class="space-y-1 md:col-span-2">
          <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1">Dirección Exacta <span class="text-red-500">*</span></label>
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