// web/src/components/admin/noticias/NoticiaMetadata.tsx
import { Show } from "solid-js";
import { Post, TYPE_LABELS, STATUS_LABELS, formatDate } from "./types";

interface Props {
  post: Post;
}

export function NoticiaMetadata(props: Props) {
  const typeInfo = () => TYPE_LABELS[props.post.type] ?? {
    label: props.post.type || "Desconocido",
    color: "bg-slate-100 text-slate-600",
  };

  const statusInfo = () => STATUS_LABELS[props.post.status] ?? {
    label: props.post.status,
    color: "bg-slate-100 text-slate-500",
  };

  return (
    <>
      <div class="flex flex-wrap items-center gap-2 mb-1.5">
        <span class={`text-[11px] font-semibold px-2 py-0.5 rounded ${typeInfo().color}`}>
          {typeInfo().label}
        </span>
        <span class={`text-[11px] font-semibold px-2 py-0.5 rounded ${statusInfo().color}`}>
          {statusInfo().label}
        </span>
        <Show when={props.post.status === "scheduled" && props.post.publish_at}>
          <span class="text-[11px] text-purple-500 font-semibold">
            ⏰ {formatDate(props.post.publish_at!)}
          </span>
        </Show>
      </div>

      <h2 class="font-semibold text-colpsi-text text-base leading-tight truncate">{props.post.title}</h2>

      <p class="text-colpsi-muted text-sm mt-1 line-clamp-1">
        {props.post.short_description || <span class="italic text-slate-300">Sin resumen</span>}
      </p>

      <div class="flex items-center gap-3 mt-2 text-[11px] text-colpsi-muted font-medium">
        <span>Por <span class="font-semibold text-colpsi-text">{props.post.create_by}</span></span>
        <span>·</span>
        <span>{formatDate(props.post.created_at)}</span>
      </div>
    </>
  );
}