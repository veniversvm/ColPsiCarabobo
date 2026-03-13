// web/src/components/admin/noticias/edit/EditHeader.tsx
import { Show } from "solid-js";
import { PostDetail, STATUS_BADGE } from "./types";

interface Props {
  post: PostDetail | null;
  loading: boolean;
  onBack: () => void;
}

export function EditHeader(props: Props) {
  return (
    <div class="flex items-center gap-4 mb-8 bg-white p-6 rounded-3xl shadow-sm border border-gray-100">
      <button
        onClick={props.onBack}
        class="w-10 h-10 bg-gray-50 hover:bg-gray-100 text-gray-600 rounded-full font-bold flex items-center justify-center transition-colors flex-shrink-0"
      >←</button>
      <div class="flex-1 min-w-0">
        <h1 class="text-2xl font-black text-blue-900 uppercase tracking-tight">Editar Publicación</h1>
        <p class="text-gray-400 text-sm mt-0.5 font-medium truncate">
          {props.loading ? "Cargando..." : props.post?.title ?? ""}
        </p>
      </div>
      <Show when={props.post}>
        <span class={`text-[10px] font-black px-3 py-1.5 rounded-full uppercase tracking-widest flex-shrink-0 ${STATUS_BADGE[props.post!.status]}`}>
          {{
            draft: "Borrador",
            published: "Publicado",
            archived: "Archivado",
            scheduled: "Programado"
          }[props.post!.status]}
        </span>
      </Show>
    </div>
  );
}