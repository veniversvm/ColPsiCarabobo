// web/src/components/ui/ImageGrid.tsx
// Grid de imágenes reutilizable
import { For } from "solid-js";

interface ImageGridProps {
  images: string[];  // Ahora explícitamente string[], no opcional
  baseUrl?: string;
}

export function ImageGrid(props: ImageGridProps) {
  const baseUrl = props.baseUrl || "";

  return (
    <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
      <For each={props.images}>
        {(imgUrl, index) => (
          <button
            onClick={() => window.open(`${baseUrl}${imgUrl}`, '_blank')}
            class="relative aspect-square rounded-xl overflow-hidden border-2 border-colpsi-border hover:border-colpsi-yellow transition-all hover:scale-105 hover:shadow-lg group"
          >
            <img 
              src={`${baseUrl}${imgUrl}`}
              alt={`Imagen ${index() + 1}`}
              class="w-full h-full object-cover"
              loading="lazy"
            />
            <div class="absolute inset-0 bg-black/0 group-hover:bg-black/10 transition-colors flex items-center justify-center">
              <span class="opacity-0 group-hover:opacity-100 text-white text-xs bg-black/50 px-2 py-1 rounded-full">
                Ver
              </span>
            </div>
          </button>
        )}
      </For>
    </div>
  );
}