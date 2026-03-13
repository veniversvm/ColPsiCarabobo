// web/src/components/admin/noticias/NoticiaImage.tsx
import { Show } from "solid-js";

interface Props {
  imageUrl?: string;
  alt: string;
}

export function NoticiaImage(props: Props) {
  return (
    <div class="flex-shrink-0 w-20 h-20 md:w-24 md:h-24 rounded-xl overflow-hidden bg-gray-100 border border-gray-200">
      <Show
        when={props.imageUrl}
        fallback={<div class="w-full h-full flex items-center justify-center text-gray-300 text-2xl">📄</div>}
      >
        <img
          src={`http://localhost:9000/colpsi-bucket/${props.imageUrl}`}
          alt={props.alt}
          class="w-full h-full object-cover"
          loading="lazy"
        />
      </Show>
    </div>
  );
}