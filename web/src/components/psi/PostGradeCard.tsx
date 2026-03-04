// web/src/components/psi/PostGradeCard.tsx
import { For, Show, createSignal } from "solid-js";
import { PostGrade } from "~/types/psi";
import { ImageModal } from "~/components/ui/ImageModal";

interface PostGradeCardProps {
  postGrade: PostGrade;
  onClick: () => void; // Para abrir el modal de detalles del postgrado
}

export function PostGradeCard(props: PostGradeCardProps) {
  const [modalImage, setModalImage] = createSignal<{ src: string; alt: string } | null>(null);

  const getImageUrl = (url?: string) => {
    if (!url) return "";
    return `http://localhost:9000/colpsi-bucket/${url}`;
  };

  const hasImages = () => {
    return !!(props.postGrade.pic_one_url || props.postGrade.pic_two_url || props.postGrade.pic_three_url);
  };

  const images = () => {
    const imgs = [];
    if (props.postGrade.pic_one_url) imgs.push({ url: props.postGrade.pic_one_url, label: "Certificado 1" });
    if (props.postGrade.pic_two_url) imgs.push({ url: props.postGrade.pic_two_url, label: "Certificado 2" });
    if (props.postGrade.pic_three_url) imgs.push({ url: props.postGrade.pic_three_url, label: "Certificado 3" });
    return imgs;
  };

  const openImageModal = (url: string, label: string, e: Event) => {
    e.stopPropagation(); // Evitar que se active el onClick de la tarjeta
    setModalImage({ src: getImageUrl(url), alt: `${props.postGrade.title} - ${label}` });
  };

  const closeModal = () => {
    setModalImage(null);
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
        class="cursor-pointer hover:bg-gray-50 p-3 rounded-xl transition-all active:scale-[0.99]"
        onClick={props.onClick}
      >
        <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-1 md:gap-2">
          <h4 class="font-bold text-gray-900 text-base md:text-lg">
            {props.postGrade.title}
          </h4>
          <span class="text-xs bg-colpsi-yellow/20 text-colpsi-blue px-2 py-1 rounded-full font-medium w-fit">
            {props.postGrade.year}
          </span>
        </div>
        
        <p class="text-colpsi-blue text-sm md:text-base">{props.postGrade.university}</p>
        
        {/* Miniaturas de certificados */}
        <Show when={hasImages()}>
          <div class="mt-3 space-y-2">
            <p class="text-[10px] font-bold text-gray-400 uppercase tracking-wider">
              Certificados ({images().length})
            </p>
            <div class="flex flex-wrap gap-2">
              <For each={images()}>
                {({ url, label }, index) => (
                  <div class="relative group">
                    {/* Miniatura */}
                    <div 
                      class="w-16 h-16 rounded-lg overflow-hidden border-2 border-gray-200 hover:border-colpsi-yellow transition-all cursor-pointer shadow-sm"
                      onClick={(e) => openImageModal(url, label, e)}
                    >
                      <img 
                        src={getImageUrl(url)} 
                        alt={label}
                        class="w-full h-full object-cover group-hover:scale-110 transition-transform"
                      />
                    </div>
                    
                    {/* Tooltip */}
                    <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-2 py-1 bg-gray-800 text-white text-[8px] rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none">
                      {label} - Click para ver
                    </div>

                    {/* Badge de número */}
                    <div class="absolute -top-1 -right-1 w-5 h-5 bg-colpsi-blue text-white text-[10px] font-bold rounded-full flex items-center justify-center border-2 border-white">
                      {index() + 1}
                    </div>
                  </div>
                )}
              </For>
            </div>
          </div>
        </Show>

        {/* Badge de certificado (simplificado) - Solo mostrar si hay imágenes pero queremos mantener compatibilidad */}
        <Show when={hasImages()}>
          <div class="mt-2 flex items-center gap-1 text-[10px] text-gray-400">
            <span>📎 {images().length} certificado(s) disponible(s) - Click en las miniaturas para ver</span>
          </div>
        </Show>
      </div>
    </>
  );
}