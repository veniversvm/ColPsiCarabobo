// web/src/components/psi/PostGradeModal.tsx
import { Show, For, createMemo } from "solid-js";
import { Portal } from "solid-js/web";
import { PostGradeModalProps } from "~/types/psi";
import { ImageGrid } from "~/components/ui/ImageGrid";
import { ModalHeader } from "./ModalHeader";

export function PostGradeModal(props: PostGradeModalProps) {
  // Usamos createMemo para crear un array tipado correctamente
  const images = createMemo<string[]>(() => {
    const urls = [
      props.postGrade?.pic_one_url,
      props.postGrade?.pic_two_url,
      props.postGrade?.pic_three_url,
    ];
    
    // Filtramos y aseguramos que solo strings no vacíos
    return urls.filter((url): url is string => 
      typeof url === 'string' && url.trim() !== ""
    );
  });

  const hasImages = () => images().length > 0;

  return (
    <Show when={props.postGrade}>
      <Portal>
        <div 
          class="fixed inset-0 bg-black/50 z-50 flex items-end md:items-center justify-center p-0 md:p-4 animate-in fade-in"
          onClick={props.onClose}
        >
          <div 
            class="bg-white w-full md:max-w-2xl md:rounded-3xl rounded-t-3xl max-h-[90vh] overflow-y-auto animate-in slide-in-from-bottom md:slide-in-from-bottom-0 md:fade-in"
            onClick={(e) => e.stopPropagation()}
          >
            <ModalHeader title="Detalles del Postgrado" onClose={props.onClose} />

            <div class="p-4 md:p-6 space-y-6">
              {/* Información principal */}
              <div>
                <h2 class="text-xl md:text-2xl font-bold text-gray-900">
                  {props.postGrade?.title}
                </h2>
                <p class="text-colpsi-blue text-base md:text-lg mt-1">
                  {props.postGrade?.university}
                </p>
                <div class="flex items-center gap-2 mt-2">
                  <span class="text-sm bg-colpsi-yellow/20 text-colpsi-blue px-3 py-1 rounded-full font-medium">
                    Año: {props.postGrade?.year}
                  </span>
                </div>
              </div>

              {/* Descripción */}
              <Show when={props.postGrade?.description}>
                <div>
                  <h4 class="text-xs font-black text-colpsi-blue uppercase tracking-wider mb-2">
                    Descripción
                  </h4>
                  <p class="text-gray-700 text-sm md:text-base leading-relaxed bg-gray-50 p-4 rounded-xl">
                    {props.postGrade?.description}
                  </p>
                </div>
              </Show>

              {/* Certificados */}
              <Show when={hasImages()}>
                <div>
                  <h4 class="text-xs font-black text-colpsi-blue uppercase tracking-wider mb-3">
                    Certificados ({images().length})
                  </h4>
                  <ImageGrid images={images()} />
                  <p class="text-[10px] text-gray-400 text-center mt-3">
                    Click en la imagen para ver en tamaño completo
                  </p>
                </div>
              </Show>

              {/* Mensaje sin info adicional */}
              <Show when={!hasImages() && !props.postGrade?.description}>
                <div class="text-center py-4 text-gray-400 text-sm">
                  No hay información adicional disponible
                </div>
              </Show>

              {/* Botón cerrar móvil */}
              <div class="sticky bottom-0 bg-white pt-4 pb-2 md:hidden">
                <button
                  onClick={props.onClose}
                  class="w-full bg-colpsi-blue text-white font-bold py-3 px-4 rounded-xl hover:bg-colpsi-blue/90 transition-colors"
                >
                  Cerrar
                </button>
              </div>
            </div>
          </div>
        </div>
      </Portal>
    </Show>
  );
}