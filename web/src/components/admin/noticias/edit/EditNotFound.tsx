// web/src/components/admin/noticias/edit/EditNotFound.tsx
interface Props {
  onBack: () => void;
}

export function EditNotFound(props: Props) {
  return (
    <div class="text-center py-24 bg-white rounded-3xl border border-colpsi-border">
      <p class="text-5xl mb-4">😕</p>
      <h2 class="text-lg font-black text-gray-700 mb-2">Publicación no encontrada</h2>
      <p class="text-gray-400 text-sm mb-6">Es posible que haya sido eliminada.</p>
      <button
        onClick={props.onBack}
        class="inline-flex items-center gap-2 text-blue-700 font-black text-sm hover:underline"
      >← Volver al listado</button>
    </div>
  );
}