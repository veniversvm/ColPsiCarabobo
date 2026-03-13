// web/src/components/admin/noticias/NoticiaMetadata.tsx
import { Show } from "solid-js";
import { Post, TYPE_LABELS, STATUS_LABELS, formatDate } from "./types";

interface Props {
  post: Post;
}

export function NoticiaMetadata(props: Props) {
  const typeInfo = () => TYPE_LABELS[props.post.type] ?? { 
    label: props.post.type || "Desconocido", 
    color: "bg-gray-100 text-gray-600" 
  };
  
  const statusInfo = () => STATUS_LABELS[props.post.status] ?? { 
    label: props.post.status, 
    color: "bg-gray-100 text-gray-500" 
  };

  return (
    <>
      <div class="flex flex-wrap items-center gap-2 mb-1.5">
        <span class={`text-[10px] font-black px-2 py-0.5 rounded-lg uppercase tracking-wider ${typeInfo().color}`}>
          {typeInfo().label}
        </span>
        <span class={`text-[10px] font-black px-2 py-0.5 rounded-lg uppercase tracking-wider ${statusInfo().color}`}>
          {statusInfo().label}
        </span>
        <Show when={props.post.status === "scheduled" && props.post.publish_at}>
          <span class="text-[10px] text-purple-500 font-bold">
            ⏰ {formatDate(props.post.publish_at!)}
          </span>
        </Show>
      </div>

      <h2 class="font-black text-gray-900 text-base leading-tight truncate">{props.post.title}</h2>

      <p class="text-gray-500 text-sm mt-1 line-clamp-1">
        {props.post.short_description || <span class="italic text-gray-300">Sin resumen</span>}
      </p>

      <div class="flex items-center gap-3 mt-2 text-[11px] text-gray-400 font-medium">
        <span>Por <span class="font-bold text-gray-600">{props.post.create_by}</span></span>
        <span>·</span>
        <span>{formatDate(props.post.created_at)}</span>
      </div>
    </>
  );
}