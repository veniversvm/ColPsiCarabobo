// web/src/components/psi/UndergraduateCard.tsx
import { Show, createSignal, For } from "solid-js";
import { bucketUrl } from "~/lib/bucket";
import { Undergraduate } from "~/types/psi";
import { ImageModal } from "~/components/ui/ImageModal";

interface UndergraduateCardProps {
  undergraduate: Undergraduate;
  onClick?: () => void; // Opcional, por si queremos abrir un modal de detalles
}

export function UndergraduateCard(props: UndergraduateCardProps) {
  const [modalImage, setModalImage] = createSignal<{ src: string; alt: string } | null>(null);

  const getImageUrl = (url?: string) => {
    if (!url) return "";
    return bucketUrl(url);
  };

  const hasImages = () => {
    return !!(
      props.undergraduate.title_image_one_url || 
      props.undergraduate.title_image_two_url || 
      props.undergraduate.title_image_three_url
    );
  };

  const images = () => {
    const imgs = [];
    if (props.undergraduate.title_image_one_url) 
      imgs.push({ url: props.undergraduate.title_image_one_url, label: "Título de Pregrado" });
    if (props.undergraduate.title_image_two_url) 
      imgs.push({ url: props.undergraduate.title_image_two_url, label: "Documento Adicional 1" });
    if (props.undergraduate.title_image_three_url) 
      imgs.push({ url: props.undergraduate.title_image_three_url, label: "Documento Adicional 2" });
    return imgs;
  };

  const openImageModal = (url: string, label: string, e: Event) => {
    e.stopPropagation();
    setModalImage({ src: getImageUrl(url), alt: `Título de Pregrado - ${label}` });
  };

  const closeModal = () => {
    setModalImage(null);
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return "";
    try {
      return new Date(dateString).toLocaleDateString("es-VE", {
        year: "numeric",
        month: "long",
        day: "numeric"
      });
    } catch {
      return dateString;
    }
  };

  return (
    <>
      {/* Modal para imágenes */}
      <ImageModal 
        src={modalImage()?.src || ""}
        alt={modalImage()?.alt || ""}
        isOpen={!!modalImage()}
        onClose={closeModal}
      />

      <div 
        class={`${props.onClick ? 'cursor-pointer' : ''} hover:bg-colpsi-surface p-4 rounded-xl transition-all`}
        onClick={props.onClick}
      >
        {/* Información principal */}
        <div class="mb-3">
          <div class="flex items-center justify-between gap-2 mb-1">
            <h4 class="font-bold text-gray-900 text-base md:text-lg">Psicólogo</h4>
            <Show when={hasImages()}>
              <span class="text-[10px] bg-blue-50 text-colpsi-blue px-2 py-1 rounded-full">
                {images().length} {images().length === 1 ? 'documento' : 'documentos'}
              </span>
            </Show>
          </div>
          
          <p class="text-colpsi-blue text-sm md:text-base font-medium">
            {props.undergraduate.university || "Universidad no especificada"}
          </p>
          
          <Show when={props.undergraduate.date || props.undergraduate.mention}>
            <div class="text-xs md:text-sm text-gray-500 mt-2 flex flex-wrap gap-x-4 gap-y-1">
              <Show when={props.undergraduate.date}>
                <span class="inline-flex items-center gap-1">
                  <span class="text-colpsi-yellow">📅</span>
                  Egreso: {formatDate(props.undergraduate.date)}
                </span>
              </Show>
              <Show when={props.undergraduate.mention}>
                <span class="inline-flex items-center gap-1">
                  <span class="text-colpsi-yellow">🎓</span>
                  Mención: {props.undergraduate.mention}
                </span>
              </Show>
            </div>
          </Show>
        </div>

        {/* Miniaturas de documentos */}
        <Show when={hasImages()}>
          <div class="mt-4 pt-3 border-t border-colpsi-border">
            <p class="text-[10px] font-bold text-gray-400 uppercase tracking-wider mb-3">
              Documentos del Título
            </p>
            <div class="flex flex-wrap gap-3">
              <For each={images()}>
                {({ url, label }, index) => (
                  <div class="relative group">
                    {/* Miniatura */}
                    <div 
                      class="w-20 h-20 rounded-lg overflow-hidden border-2 border-gray-200 hover:border-colpsi-yellow transition-all cursor-pointer shadow-sm"
                      onClick={(e) => openImageModal(url, label, e)}
                    >
                      <img 
                        src={getImageUrl(url)} 
                        alt={label}
                        class="w-full h-full object-cover group-hover:scale-110 transition-transform duration-300"
                      />
                      
                      {/* Overlay con lupa al hover */}
                      <div class="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-colors flex items-center justify-center opacity-0 group-hover:opacity-100">
                        <span class="bg-white text-colpsi-blue p-1.5 rounded-full shadow-lg transform scale-90 group-hover:scale-100 transition-transform">
                          🔍
                        </span>
                      </div>
                    </div>

                    {/* Tooltip */}
                    <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-2 py-1 bg-gray-800 text-white text-[8px] rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none z-10">
                      {label} - Click para ver
                    </div>

                    {/* Badge de número */}
                    <div class="absolute -top-1 -right-1 w-5 h-5 bg-colpsi-blue text-white text-[10px] font-bold rounded-full flex items-center justify-center border-2 border-white shadow-sm">
                      {index() + 1}
                    </div>
                  </div>
                )}
              </For>
            </div>
            
            {/* Texto informativo */}
            <p class="text-[9px] text-gray-400 mt-3">
              Haz clic en cualquier documento para verlo en tamaño completo
            </p>
          </div>
        </Show>
      </div>
    </>
  );
}