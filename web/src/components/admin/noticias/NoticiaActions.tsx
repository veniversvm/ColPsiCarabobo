// web/src/components/admin/noticias/NoticiaActions.tsx
import { A } from "@solidjs/router";
import { Show } from "solid-js";
import { Post } from "./types";
import { Icon } from "~/components/admin/ui/icons";

interface Props {
  post: Post;
  isBusy: boolean;
  onToggle: (post: Post) => void;
  onDelete: (id: string) => void;
}

const iconBtn =
  "w-8 h-8 rounded-md flex items-center justify-center border border-colpsi-border transition-colors disabled:opacity-40";

export function NoticiaActions(props: Props) {
  return (
    <div class="flex-shrink-0 flex items-center gap-1.5">
      {/* Toggle published/draft */}
      <button
        onClick={() => props.onToggle(props.post)}
        disabled={props.isBusy || props.post.status === "archived"}
        title={props.post.status === "published" ? "Despublicar" : "Publicar"}
        class={`${iconBtn} ${
          props.post.status === "published"
            ? "border-emerald-200 bg-emerald-50 text-emerald-600 hover:bg-emerald-100"
            : "bg-white text-slate-400 hover:text-colpsi-blue hover:bg-colpsi-bg"
        }`}
      >
        {props.isBusy ? "…" : props.post.status === "published" ? "✓" : "○"}
      </button>

      {/* Editar */}
      <A
        href={`/admin/noticias/${props.post.id}`}
        class={`${iconBtn} bg-white text-colpsi-muted hover:text-colpsi-blue hover:bg-colpsi-bg`}
        title="Editar"
      >
        <Icon name="pencil" class="w-4 h-4" />
      </A>

      {/* Archivar */}
      <button
        onClick={() => props.onDelete(props.post.id)}
        disabled={props.isBusy || props.post.status === "archived"}
        title="Archivar"
        class={`${iconBtn} bg-white text-slate-400 hover:text-colpsi-red hover:bg-red-50 hover:border-red-100`}
      >
        <Icon name="trash" class="w-4 h-4" />
      </button>
    </div>
  );
}