// web/src/components/admin/noticias/EmptyState.tsx
import { A } from "@solidjs/router";
import { Icon } from "~/components/admin/ui/icons";

interface Props {
  type: "no-posts" | "no-results";
  hasFilters?: boolean;
}

export function EmptyState(props: Props) {
  if (props.type === "no-posts") {
    return (
      <div class="text-center py-16 bg-white rounded-lg border border-colpsi-border">
        <span class="inline-flex h-12 w-12 items-center justify-center rounded-md bg-colpsi-bg text-colpsi-muted mb-4">
          <Icon name="newspaper" class="w-6 h-6" />
        </span>
        <p class="text-colpsi-muted font-medium">No hay publicaciones aún</p>
        <A href="/admin/noticias/crear" class="mt-3 inline-block text-colpsi-blue font-semibold text-sm hover:underline">
          Crear la primera →
        </A>
      </div>
    );
  }

  return (
    <div class="text-center py-14 bg-white rounded-lg border border-colpsi-border">
      <p class="text-colpsi-muted font-medium">Ningún resultado para los filtros aplicados</p>
    </div>
  );
}