// web/src/components/psi/profile/ProfessionalSection.tsx
// web/src/components/psi/profile/ProfessionalSection.tsx
import { RichTextEditor } from "~/components/ui/RichTextEditor";
const FULL_BIO_WORD_LIMIT = 5000;

function countWords(html: string): number {
  const text = html.replace(/<[^>]*>/g, " ").replace(/\s+/g, " ").trim();
  return text === "" ? 0 : text.split(" ").length;
}

export function ProfessionalSection(props: {
  primarySpecialty: string;
  secondarySpecialty: string;
  miniBio: string;
  fullBio: string;
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

  return (
    <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100">
      <h2 class="text-xl font-black text-colpsi-blue mb-6 border-l-4 border-colpsi-yellow pl-3">Perfil Profesional</h2>
      <div class="space-y-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="space-y-2">
            <label class="text-xs font-bold text-gray-500 uppercase ml-2">Especialidad Principal</label>
            <input type="text" value={props.primarySpecialty} onInput={(e) => props.onPrimarySpecialtyChange(e.currentTarget.value)} class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all" />
          </div>
          <div class="space-y-2">
            <label class="text-xs font-bold text-gray-500 uppercase ml-2">Especialidad Secundaria</label>
            <input type="text" value={props.secondarySpecialty} onInput={(e) => props.onSecondarySpecialtyChange(e.currentTarget.value)} class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all" />
          </div>
        </div>

        <div class="space-y-2">
          <label class="text-xs font-bold text-gray-500 uppercase ml-2">Mini Biografía (Max 250 caracteres)</label>
          <textarea
            value={props.miniBio}
            onInput={(e) => props.onMiniBioChange(e.currentTarget.value)}
            maxlength="250"
            class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-2xl px-5 py-4 outline-none text-colpsi-text transition-all min-h-[100px] resize-y"
          />
        </div>

        <div class="pt-4 border-t border-gray-100">
          <div class="flex justify-between items-center mb-2">
            <label class="text-xs font-bold text-gray-500 uppercase ml-2">Biografía Extensa (Perfil Detallado)</label>
            <span class={`text-xs font-semibold mr-2 ${isOverLimit() ? "text-red-500" : "text-gray-400"}`}>
              {wordCount()} / {FULL_BIO_WORD_LIMIT} palabras
            </span>
          </div>
          <RichTextEditor
            label=""
            content={props.fullBio}
            onUpdate={handleFullBioChange}
          />
          {isOverLimit() && (
            <p class="text-xs text-red-500 mt-1 ml-2">Has superado el límite de {FULL_BIO_WORD_LIMIT} palabras.</p>
          )}
        </div>
      </div>
    </section>
  );
}