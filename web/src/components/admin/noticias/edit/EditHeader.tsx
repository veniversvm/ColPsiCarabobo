// web/src/components/admin/noticias/edit/EditHeader.tsx
import { Show } from "solid-js";
import { PostDetail, STATUS_BADGE } from "./types";
import { Icon } from "~/components/admin/ui/icons";

interface Props {
  post: PostDetail | null;
  loading: boolean;
  onBack: () => void;
}

const STATUS_LABEL: Record<PostDetail["status"], string> = {
  draft: "Borrador",
  published: "Publicado",
  archived: "Archivado",
  scheduled: "Programado",
};

export function EditHeader(props: Props) {
  return (
    <div class="flex items-center gap-3 pb-4 border-b border-colpsi-border">
      <button
        onClick={props.onBack}
        class="inline-flex items-center justify-center h-8 w-8 rounded-md border border-colpsi-border bg-white text-colpsi-muted hover:text-colpsi-blue hover:bg-colpsi-bg transition-colors"
        title="Volver"
      >
        <Icon name="chevronRight" class="w-4 h-4 rotate-180" />
      </button>
      <div class="flex-1 min-w-0">
        <h1 class="text-lg font-semibold text-colpsi-text">Editar Publicación</h1>
        <p class="text-sm text-colpsi-muted mt-0.5 truncate">
          {props.loading ? "Cargando..." : props.post?.title ?? ""}
        </p>
      </div>
      <Show when={props.post}>
        <span class={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-md border text-[11px] font-medium whitespace-nowrap ${STATUS_BADGE[props.post!.status]}`}>
          <span class="w-1.5 h-1.5 rounded-full bg-current" />
          {STATUS_LABEL[props.post!.status]}
        </span>
      </Show>
    </div>
  );
}