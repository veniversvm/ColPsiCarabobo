// web/src/components/admin/psicologos/edit/sections/AcademicSection.tsx

import { Show } from "solid-js";
import { ToggleSwitch } from "~/components/ui/ToggleSwitch";
import { Field, IC } from "../EditPrimitives";
import type { EditFormState } from "../types";

interface Props {
  form: EditFormState;
  setForm: (key: keyof EditFormState, value: any) => void;
  // --- Nuevas Props para manejo de archivos ---
  files: { [key: string]: File };
  setFiles: (files: any) => void;
}

export function AcademicSection(props: Props) {
  
  // Helper para manejar la selección de archivos de títulos
  const handleFileChange = (field: string, e: Event) => {
    const file = (e.currentTarget as HTMLInputElement).files?.[0];
    if (file) {
      props.setFiles((prev: any) => ({ ...prev, [field]: file }));
    }
  };

  const removeFile = (field: string) => {
    props.setFiles((prev: any) => {
      const newFiles = { ...prev };
      delete newFiles[field];
      return newFiles;
    });
  };

  return (
    <>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-5">

        {/* ── Datos de Egreso ────────────────────────────────────────── */}
        <div class="md:col-span-2">
          <Field label="Universidad de Egreso">
            <input type="text"
              value={props.form.university_undergraduate}
              onInput={(e) => props.setForm("university_undergraduate", e.currentTarget.value)}
              class={IC}
              placeholder="Ej: Universidad de Carabobo"
            />
          </Field>
        </div>

        <Field label="Fecha de Egreso">
          <input type="date"
            value={props.form.graduate_date}
            onInput={(e) => props.setForm("graduate_date", e.currentTarget.value)}
            class={IC}
          />
        </Field>

        <div class="md:col-span-2">
          <Field label="Mención / Título Obtenido">
            <input type="text"
              value={props.form.mention_undergraduate}
              onInput={(e) => props.setForm("mention_undergraduate", e.currentTarget.value)}
              class={IC}
              placeholder="Ej: Licenciado en Psicología"
            />
          </Field>
        </div>

        <Field label="Inscripción al Gremio (Fecha)">
          <input type="date"
            value={props.form.guild_inscription_date}
            onInput={(e) => props.setForm("guild_inscription_date", e.currentTarget.value)}
            class={IC}
          />
        </Field>

        {/* ── Imágenes del Título (S3) ────────────────────────────────── */}
        <div class="col-span-full pt-4 border-t border-gray-100">
          <p class="text-[10px] font-black text-gray-400 uppercase tracking-widest mb-4">
            Documentación Digital (Imágenes del Título)
          </p>
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            {[
              { id: "title_image_one", label: "Cara Principal" },
              { id: "title_image_two", label: "Reverso / Sellos" },
              { id: "title_image_three", label: "Anexo / Registro" }
            ].map((img) => (
              <div class="relative group">
                <label class={`flex flex-col items-center justify-center h-24 border-2 border-dashed rounded-2xl transition-all cursor-pointer ${
                  props.files[img.id] ? "border-emerald-300 bg-emerald-50" : "border-gray-200 bg-gray-50 hover:bg-blue-50"
                }`}>
                  <Show when={props.files[img.id]} fallback={
                    <>
                      <span class="text-xl">📄</span>
                      <span class="text-[9px] font-bold text-gray-400 uppercase mt-1">{img.label}</span>
                    </>
                  }>
                    <span class="text-emerald-600 text-[10px] font-black uppercase text-center px-2">
                      ✅ {props.files[img.id].name.slice(0, 15)}...
                    </span>
                  </Show>
                  <input type="file" accept="image/*" class="hidden" onChange={(e) => handleFileChange(img.id, e)} />
                </label>
                
                <Show when={props.files[img.id]}>
                  <button 
                    type="button"
                    onClick={() => removeFile(img.id)}
                    class="absolute -top-2 -right-2 bg-red-500 text-white w-5 h-5 rounded-full text-[10px] shadow-md flex items-center justify-center hover:bg-red-600 transition-colors"
                  >✕</button>
                </Show>
              </div>
            ))}
          </div>
        </div>

        {/* ── Registro Legal ──────────────────────────────────────────── */}
        <div class="col-span-full border-t border-gray-100 pt-6 mt-2">
          <p class="text-[10px] font-black text-gray-400 uppercase tracking-widest mb-4">
            Datos de Registro de Título (Principal/Estado)
          </p>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
            <Field label="Nro. Registro">
              <input type="number"
                value={props.form.register_number}
                onInput={(e) => props.setForm("register_number", e.currentTarget.value)}
                class={IC}
              />
            </Field>
            <Field label="Estado de Registro">
              <input type="text"
                value={props.form.register_title_state}
                onInput={(e) => props.setForm("register_title_state", e.currentTarget.value)}
                class={IC}
              />
            </Field>
            <Field label="Fecha de Registro">
              <input type="date"
                value={props.form.register_title_date}
                onInput={(e) => props.setForm("register_title_date", e.currentTarget.value)}
                class={IC}
              />
            </Field>
            <Field label="Folio">
              <input type="text"
                value={props.form.register_folio}
                onInput={(e) => props.setForm("register_folio", e.currentTarget.value)}
                class={IC}
              />
            </Field>
            <Field label="Tomo">
              <input type="text"
                value={props.form.register_tome}
                onInput={(e) => props.setForm("register_tome", e.currentTarget.value)}
                class={IC}
              />
            </Field>
          </div>
        </div>

      </div>

      {/* ── Privacidad ────────────────────────────────────────────────── */}
      <div class="mt-8 bg-blue-50/50 p-6 rounded-3xl border border-blue-100">
        <p class="text-[10px] font-black text-blue-900 uppercase tracking-widest mb-4">
          Privacidad de Formación Académica (Directorio Público)
        </p>
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <ToggleSwitch
            label="Mostrar Universidad"
            checked={props.form.show_university_undergraduate}
            onChange={(v) => props.setForm("show_university_undergraduate", v)}
          />
          <ToggleSwitch
            label="Mostrar Fecha Egreso"
            checked={props.form.show_graduate_date}
            onChange={(v) => props.setForm("show_graduate_date", v)}
          />
          <ToggleSwitch
            label="Mostrar Mención"
            checked={props.form.show_mention_undergraduate}
            onChange={(v) => props.setForm("show_mention_undergraduate", v)}
          />
        </div>
      </div>
    </>
  );
}