// web/src/components/ui/ImageModal.tsx
import { Show } from "solid-js";
import { Portal } from "solid-js/web";
import { Animate, Presence } from "~/components/ui/Motion";

interface ImageModalProps {
  src: string;
  alt: string;
  isOpen: boolean;
  onClose: () => void;
}

export function ImageModal(props: ImageModalProps) {
  return (
    <Presence>
      <Show when={props.isOpen}>
        <Portal>
          {/* Overlay */}
          <Animate
            variant="fade"
            class="fixed inset-0 bg-black/80 z-50 flex items-center justify-center p-4"
            exit={{ opacity: 0 }}
            onClick={props.onClose}
          >
            {/* Modal */}
            <Animate variant="zoom" class="relative max-w-5xl w-full max-h-[90vh]" exit={{ opacity: 0, scale: 0.95 }} onClick={(e) => e.stopPropagation()}>
              {/* Botón cerrar */}
              <button
                onClick={props.onClose}
                class="absolute -top-12 right-0 md:-top-14 md:right-4 text-white hover:text-colpsi-yellow transition-colors text-3xl"
                title="Cerrar"
              >
                ✕
              </button>

              {/* Imagen */}
              <div class="bg-white rounded-2xl overflow-hidden shadow-2xl">
                <img
                  src={props.src}
                  alt={props.alt}
                  class="w-full h-full object-contain max-h-[85vh]"
                />
              </div>

              {/* Info de la imagen */}
              <div class="absolute bottom-4 left-1/2 -translate-x-1/2 bg-black/50 text-white text-xs px-3 py-1.5 rounded-full backdrop-blur-sm">
                {props.alt}
              </div>
            </Animate>
          </Animate>
        </Portal>
      </Show>
    </Presence>
  );
}