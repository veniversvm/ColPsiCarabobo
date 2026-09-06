// web/src/components/admin/noticias/EmptyState.tsx
import { A } from "@solidjs/router";

interface Props {
  type: "no-posts" | "no-results";
  hasFilters?: boolean;
}

export function EmptyState(props: Props) {
  if (props.type === "no-posts") {
    return (
      <div class="text-center py-20 bg-white rounded-3xl border border-colpsi-border">
        <p class="text-5xl mb-4">📰</p>
        <p class="text-gray-400 font-bold">No hay publicaciones aún</p>
        <A href="/admin/noticias/crear" class="mt-4 inline-block text-blue-600 font-black text-sm hover:underline">
          Crear la primera →
        </A>
      </div>
    );
  }

  return (
    <div class="text-center py-16 bg-white rounded-3xl border border-colpsi-border">
      <p class="text-gray-400 font-bold">Ningún resultado para los filtros aplicados</p>
    </div>
  );
}