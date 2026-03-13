// web/src/components/admin/noticias/NoticiaActions.tsx
import { A } from "@solidjs/router";
import { Show } from "solid-js";
import { Post } from "./types";

interface Props {
  post: Post;
  isBusy: boolean;
  onToggle: (post: Post) => void;
  onDelete: (id: string) => void;
}

export function NoticiaActions(props: Props) {
  return (
    <div class="flex-shrink-0 flex flex-col sm:flex-row items-center gap-2">
      {/* Toggle published/draft */}
      <button
        onClick={() => props.onToggle(props.post)}
        disabled={props.isBusy || props.post.status === "archived"}
        title={props.post.status === "published" ? "Despublicar" : "Publicar"}
        class={`w-9 h-9 rounded-xl flex items-center justify-center border-2 transition-all font-bold text-sm disabled:opacity-40 ${
          props.post.status === "published"
            ? "border-emerald-200 bg-emerald-50 text-emerald-600 hover:bg-emerald-100"
            : "border-gray-200 bg-gray-50 text-gray-400 hover:bg-gray-100"
        }`}
      >
        {props.isBusy ? "…" : props.post.status === "published" ? "✓" : "○"}
      </button>

      {/* Editar */}
      <A
        href={`/admin/noticias/${props.post.id}`}
        class="w-9 h-9 rounded-xl flex items-center justify-center border-2 border-blue-100 bg-blue-50 text-blue-600 hover:bg-blue-100 transition-all"
        title="Editar"
      >
        ✏
      </A>

      {/* Archivar */}
      <button
        onClick={() => props.onDelete(props.post.id)}
        disabled={props.isBusy || props.post.status === "archived"}
        title="Archivar"
        class="w-9 h-9 rounded-xl flex items-center justify-center border-2 border-red-100 bg-red-50 text-red-400 hover:bg-red-100 hover:text-red-600 transition-all disabled:opacity-40"
      >
        🗑
      </button>
    </div>
  );
}