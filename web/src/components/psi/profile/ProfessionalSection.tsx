// web/src/components/psi/profile/ProfessionalSection.tsx
import { For, Show } from "solid-js";
import { RichTextEditor } from "~/components/ui/RichTextEditor";

const FULL_BIO_WORD_LIMIT = 5000;

function countWords(html: string): number {
  const text = html.replace(/<[^>]*>/g, " ").replace(/\s+/g, " ").trim();
  return text === "" ? 0 : text.split(" ").length;
}

interface Specialty {
  id: number;
  name: string;
}

export function ProfessionalSection(props: {
  primarySpecialty: string;
  secondarySpecialty: string;
  miniBio: string;
  fullBio: string;
  specialties: Specialty[] | undefined;
  onPrimarySpecialtyChange: (v: string) => void;
  onSecondarySpecialtyChange: (v: string) => void;
  onMiniBioChange: (v: string) => void;
  onFullBioChange: (v: string) => void;
}) {
  const wordCount = () => countWords(props.fullBio);
  const isOverLimit = () => wordCount() > FULL_BIO_WORD_LIMIT;

  const handleFullBioChange = (v: string) => {
    if (countWords(v) <= FULL_BIO_WORD_LIMIT) {
      props.onFullBioChange(v);
    }
  };

  // Clases compartidas para los selects
  const selectClass = "w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all appearance-none cursor-pointer disabled:opacity-60 disabled:cursor-wait";

  return (
    <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100">
      <h2 class="text-xl font-black text-colpsi-blue mb-6 border-l-4 border-colpsi-yellow pl-3">
        Perfil Profesional
      </h2>

      <div class="space-y-6">

        {/* ── ESPECIALIDADES ────────────────────────────────────────────── */}
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">

          {/* Especialidad Principal */}
          <div class="space-y-2">
            <label class="text-xs font-bold text-gray-500 uppercase ml-2">
              Especialidad Principal
            </label>
            <select
              value={props.primarySpecialty}
              disabled={!props.specialties}
              class={selectClass}
              onChange={(e) => props.onPrimarySpecialtyChange(e.currentTarget.value)}
            >
              <Show
                when={props.specialties}
                fallback={<option value="">Cargando...</option>}
              >
                <option value="">— Sin especialidad —</option>
                <For each={props.specialties}>
                  {(item) => (
                    <option value={item.name}>{item.name}</option>
                  )}
                </For>
              </Show>
            </select>
          </div>

          {/* Especialidad Secundaria */}
          <div class="space-y-2">
            <label class="text-xs font-bold text-gray-500 uppercase ml-2">
              Especialidad Secundaria
            </label>
            <select
              value={props.secondarySpecialty}
              disabled={!props.specialties}
              class={selectClass}
              onChange={(e) => props.onSecondarySpecialtyChange(e.currentTarget.value)}
            >
              <Show
                when={props.specialties}
                fallback={<option value="">Cargando...</option>}
              >
                <option value="">— Sin especialidad —</option>
                <For each={props.specialties}>
                  {(item) => (
                    <option
                      value={item.name}
                      // Deshabilita la opción si ya está seleccionada como principal
                      disabled={item.name === props.primarySpecialty}
                    >
                      {item.name}
                    </option>
                  )}
                </For>
              </Show>
            </select>
            <Show when={props.secondarySpecialty && props.secondarySpecialty === props.primarySpecialty}>
              <p class="text-xs text-amber-500 font-bold ml-2">
                ⚠️ La especialidad secundaria no puede ser igual a la principal.
              </p>
            </Show>
          </div>
        </div>

        {/* ── MINI BIO ──────────────────────────────────────────────────── */}
        <div class="space-y-2">
          <label class="text-xs font-bold text-gray-500 uppercase ml-2">
            Mini Biografía <span class="text-gray-400 font-medium normal-case">(máx. 250 caracteres)</span>
          </label>
          <textarea
            value={props.miniBio}
            onInput={(e) => props.onMiniBioChange(e.currentTarget.value)}
            maxlength="250"
            class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-2xl px-5 py-4 outline-none text-colpsi-text transition-all min-h-[100px] resize-y"
          />
          <p class="text-[11px] text-gray-400 text-right mr-2">
            {props.miniBio.length}/250
          </p>
        </div>

        {/* ── FULL BIO ──────────────────────────────────────────────────── */}
        <div class="pt-4 border-t border-gray-100">
          <div class="flex justify-between items-center mb-2">
            <label class="text-xs font-bold text-gray-500 uppercase ml-2">
              Biografía Extensa <span class="text-gray-400 font-medium normal-case">(Perfil Detallado)</span>
            </label>
            <span class={`text-xs font-semibold mr-2 ${isOverLimit() ? "text-red-500" : "text-gray-400"}`}>
              {wordCount()} / {FULL_BIO_WORD_LIMIT} palabras
            </span>
          </div>
          <RichTextEditor
            label=""
            content={props.fullBio}
            onUpdate={handleFullBioChange}
          />
          <Show when={isOverLimit()}>
            <p class="text-xs text-red-500 mt-1 ml-2">
              Has superado el límite de {FULL_BIO_WORD_LIMIT} palabras.
            </p>
          </Show>
        </div>

      </div>
    </section>
  );
}