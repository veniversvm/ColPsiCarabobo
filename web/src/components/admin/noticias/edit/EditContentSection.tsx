// web/src/components/admin/noticias/edit/EditContentSection.tsx
import { RichTextEditor } from "~/components/ui/RichTextEditor";

interface Props {
  content: string;
  onUpdate: (html: string) => void;
}

export function EditContentSection(props: Props) {
  return (
    <section class="bg-white rounded-lg p-5 border border-colpsi-border">
      <h2 class="text-base font-semibold text-colpsi-text border-b border-colpsi-border pb-3 mb-4">
        Contenido <span class="text-red-400">*</span>
      </h2>
      <RichTextEditor content={props.content} onUpdate={props.onUpdate} />
    </section>
  );
}