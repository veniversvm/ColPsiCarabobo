// web/src/components/ui/SearchingIndicator.tsx (mejorado)
export function SearchingIndicator() {
  return (
    <div class="flex flex-col items-center justify-center py-16 px-4">
      <div class="relative">
        {/* Spinner animado */}
        <div class="w-16 h-16 border-4 border-colpsi-yellow/20 border-t-colpsi-yellow rounded-full animate-spin"></div>
        {/* Icono de lupa */}
        <div class="absolute inset-0 flex items-center justify-center">
          <span class="text-2xl animate-pulse">🔍</span>
        </div>
      </div>
      <p class="text-colpsi-blue font-bold mt-6 text-sm uppercase tracking-wider animate-pulse">
        Buscando profesionales...
      </p>
      <p class="text-gray-400 text-xs mt-2">
        Esto puede tomar unos segundos
      </p>
    </div>
  );
}