// web/src/components/admin/noticias/edit/EditContentSection.tsx
import { RichTextEditor } from "~/components/ui/RichTextEditor";

interface Props {
  content: string;
  onUpdate: (html: string) => void;
}

export function EditContentSection(props: Props) {
  return (
    <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100">
      <h2 class="text-sm font-black text-blue-800 uppercase tracking-widest border-b border-gray-100 pb-3 mb-5">
        Contenido <span class="text-red-400">*</span>
      </h2>
      <RichTextEditor content={props.content} onUpdate={props.onUpdate} />
    </section>
  );
}