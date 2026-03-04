// web/src/components/psi/PostGradeModal.tsx (actualizado para usar ImageModal)
import { Show, For, createSignal } from "solid-js";
import { Portal } from "solid-js/web";
import { PostGrade } from "~/types/psi";
import { ImageModal } from "~/components/ui/ImageModal";
import { ModalHeader } from "./ModalHeader";

interface PostGradeModalProps {
  postGrade: PostGrade | null;
  onClose: () => void;
}

export function PostGradeModal(props: PostGradeModalProps) {
  const [modalImage, setModalImage] = createSignal<{ src: string; alt: string } | null>(null);

  const images = () => {
    const urls = [
      props.postGrade?.pic_one_url,
      props.postGrade?.pic_two_url,
      props.postGrade?.pic_three_url,
    ];
    
    return urls.filter((url): url is string => 
      typeof url === 'string' && url.trim() !== ""
    );
  };

  const hasImages = () => images().length > 0;

  const getImageUrl = (url: string) => {
    return `http://localhost:9000/colpsi-bucket/${url}`;
  };

  const openImageModal = (url: string, label: string) => {
    setModalImage({ 
      src: getImageUrl(url), 
      alt: `${props.postGrade?.title} - ${label}` 
    });
  };

  const closeModal = () => {
    setModalImage(null);
  };

  return (
    <Show when={props.postGrade}>
      <Portal>
        {/* Modal de imagen (anidado) */}
        <ImageModal 
          src={modalImage()?.src || ""}
          alt={modalImage()?.alt || ""}
          isOpen={!!modalImage()}
          onClose={closeModal}
        />

        {/* Modal principal */}
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

              {/* Certificados con miniaturas clickeables */}
              <Show when={hasImages()}>
                <div>
                  <h4 class="text-xs font-black text-colpsi-blue uppercase tracking-wider mb-3">
                    Certificados ({images().length})
                  </h4>
                  <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
                    <For each={images()}>
                      {(imgUrl, index) => (
                        <div class="space-y-2">
                          <div 
                            class="relative aspect-square rounded-xl overflow-hidden border-2 border-gray-100 hover:border-colpsi-yellow transition-all cursor-pointer group shadow-md"
                            onClick={() => openImageModal(imgUrl, `Certificado ${index() + 1}`)}
                          >
                            <img 
                              src={getImageUrl(imgUrl)}
                              alt={`Certificado ${index() + 1}`}
                              class="w-full h-full object-cover group-hover:scale-110 transition-transform duration-300"
                            />
                            <div class="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-colors flex items-center justify-center">
                              <span class="bg-white text-colpsi-blue p-2 rounded-full shadow-lg transform scale-90 group-hover:scale-100 transition-transform opacity-0 group-hover:opacity-100">
                                🔍
                              </span>
                            </div>
                          </div>
                          <p class="text-[10px] text-center text-gray-500">
                            Certificado {index() + 1}
                          </p>
                        </div>
                      )}
                    </For>
                  </div>
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