// web/src/components/ui/FlagFooter.tsx
// Componente reutilizable para la franja de la bandera
export function FlagFooter() {
  return (
    <div class="fixed bottom-0 left-0 w-full h-2 flex overflow-hidden z-50">
      <div class="flex-1 bg-colpsi-red"></div>
      <div class="flex-1 bg-green-700"></div>
      <div class="flex-1 bg-colpsi-blue"></div>
    </div>
  );
}