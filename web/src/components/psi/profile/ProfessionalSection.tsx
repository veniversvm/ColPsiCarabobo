// web/src/components/psi/profile/ProfessionalSection.tsx
import { Show } from "solid-js";
import { RichTextEditor } from "~/components/ui/RichTextEditor";
import { DropdownSelect } from "~/components/ui/DropdownSelect";

const FULL_BIO_WORD_LIMIT = 5000;

function countWords(html: string): number {
  const text = html.replace(/<[^>]*>/g, " ").replace(/\s+/g, " ").trim();
  return text === "" ? 0 : text.split(" ").length;
}

interface WorkArea {
  id: number;
  name: string;
}

interface ProfessionalSectionProps {
  primaryWorkArea: string;     // Actualizado
  secondaryWorkArea: string;   // Actualizado
  miniBio: string;
  fullBio: string;
  specialties: WorkArea[] | undefined; // Mantiene el nombre de la lista o cámbialo a workAreas
  onPrimaryWorkAreaChange: (v: string) => void;   // Actualizado
  onSecondaryWorkAreaChange: (v: string) => void; // Actualizado
  onMiniBioChange: (v: string) => void;
  onFullBioChange: (v: string) => void;
}

export function ProfessionalSection(props: ProfessionalSectionProps) {
  const wordCount = () => countWords(props.fullBio);
  const isOverLimit = () => wordCount() > FULL_BIO_WORD_LIMIT;

  const handleFullBioChange = (v: string) => {
    // Permitimos el cambio pero el indicador visual mostrará el error si se pasa
    props.onFullBioChange(v);
  };

  const selectClass = "w-full bg-colpsi-surface border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 text-colpsi-text";

  return (
    <section>


      <div class="space-y-6">

        {/* ── ÁREAS DE TRABAJO (Antes Especialidades) ────────────────────── */}
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">

          {/* Área Principal */}
          <div class="space-y-2">
            <label class="text-xs font-bold text-gray-500 uppercase ml-2">
              Área de Trabajo Principal
            </label>
            <DropdownSelect
              value={props.primaryWorkArea}
              disabled={!props.specialties}
              loading={!props.specialties}
              loadingLabel="Cargando áreas..."
              placeholder="— Sin área asignada —"
              buttonClass={selectClass}
              options={
                props.specialties
                  ? [
                      { value: "", label: "— Sin área asignada —" },
                      ...props.specialties.map((item) => ({
                        value: item.name,
                        label: item.name,
                      })),
                    ]
                  : []
              }
              onChange={props.onPrimaryWorkAreaChange}
            />
          </div>

          {/* Área Secundaria */}
          <div class="space-y-2">
            <label class="text-xs font-bold text-gray-500 uppercase ml-2">
              Área de Trabajo Secundaria
            </label>
            <DropdownSelect
              value={props.secondaryWorkArea}
              disabled={!props.specialties}
              loading={!props.specialties}
              loadingLabel="Cargando áreas..."
              placeholder="— Sin área secundaria —"
              buttonClass={selectClass}
              options={
                props.specialties
                  ? [
                      { value: "", label: "— Sin área secundaria —" },
                      ...props.specialties.map((item) => ({
                        value: item.name,
                        label: item.name,
                        disabled: item.name === props.primaryWorkArea,
                      })),
                    ]
                  : []
              }
              onChange={props.onSecondaryWorkAreaChange}
            />
            <Show when={props.secondaryWorkArea && props.secondaryWorkArea === props.primaryWorkArea}>
              <p class="text-xs text-amber-500 font-bold ml-2">
                ⚠️ El área secundaria no puede ser igual a la principal.
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
            class="w-full bg-colpsi-surface border-2 border-transparent focus:border-colpsi-yellow rounded-2xl px-5 py-4 outline-none text-colpsi-text transition-all min-h-[100px] resize-y"
            placeholder="Escribe un breve resumen de tu práctica profesional..."
          />
          <p class="text-[11px] text-gray-400 text-right mr-2">
            {props.miniBio?.length || 0}/250
          </p>
        </div>

        {/* ── FULL BIO ──────────────────────────────────────────────────── */}
        <div class="pt-4 border-t border-colpsi-border">
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
            <p class="text-xs text-red-500 mt-1 ml-2 font-bold">
              ⚠️ Has superado el límite de {FULL_BIO_WORD_LIMIT} palabras. Por favor, resume tu contenido.
            </p>
          </Show>
        </div>

      </div>
    </section>
  );
}