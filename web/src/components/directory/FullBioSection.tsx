import { Show } from "solid-js";
import { sanitizeHtml } from "~/lib/sanitize-html";

interface FullBioSectionProps {
  content?: string;
}

export function FullBioSection(props: FullBioSectionProps) {
  // Verificamos si realmente hay contenido útil (evitamos párrafos vacíos de Tiptap)
  const hasContent = () => {
    const html = props.content?.trim();
    return html && html !== "" && html !== "<p></p>";
  };

  return (
    <Show when={hasContent()}>
      <div class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100">
        <h3 class="text-xs md:text-sm font-black text-colpsi-blue uppercase tracking-widest mb-6 border-b-2 border-gray-50 pb-4">
          Biografía y Experiencia Detallada
        </h3>
        
        {/* Renderizado seguro del HTML con clases de Tailwind Typography */}
        <div 
          class="prose prose-slate max-w-none text-colpsi-text 
                 prose-headings:text-colpsi-blue prose-headings:font-black
                 prose-a:text-colpsi-yellow hover:prose-a:text-colpsi-blue transition-colors 
                 prose-li:marker:text-colpsi-yellow prose-strong:text-colpsi-blue"
          innerHTML={sanitizeHtml(props.content)} 
        />
      </div>
    </Show>
  );
}