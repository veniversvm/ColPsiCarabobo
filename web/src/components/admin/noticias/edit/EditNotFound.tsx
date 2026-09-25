// web/src/components/admin/noticias/edit/EditNotFound.tsx
import { Icon } from "~/components/admin/ui/icons";

interface Props {
  onBack: () => void;
}

export function EditNotFound(props: Props) {
  return (
    <div class="text-center py-16 bg-white rounded-lg border border-colpsi-border">
      <span class="inline-flex h-12 w-12 items-center justify-center rounded-md bg-colpsi-bg text-colpsi-muted mb-4">
        <Icon name="fileText" class="w-6 h-6" />
      </span>
      <h2 class="text-base font-semibold text-colpsi-text mb-1">Publicación no encontrada</h2>
      <p class="text-sm text-colpsi-muted mb-5">Es posible que haya sido eliminada.</p>
      <button
        onClick={props.onBack}
        class="inline-flex items-center gap-1.5 text-colpsi-blue font-semibold text-sm hover:underline"
      >
        <Icon name="chevronRight" class="w-3.5 h-3.5 rotate-180" />
        Volver al listado
      </button>
    </div>
  );
}