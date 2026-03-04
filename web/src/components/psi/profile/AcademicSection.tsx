// web/src/components/psi/profile/AcademicSection.tsx
import { Show, createSignal } from "solid-js";
import { FileUploader } from "~/components/ui/fileUploader";
import { ImageModal } from "~/components/ui/ImageModal";

interface AcademicSectionProps {
  undergraduateData: {
    university_undergraduate?: string;
    graduate_date?: string;
    mention_undergraduate?: string;
    title_image_one_url?: string;
    title_image_two_url?: string;
    title_image_three_url?: string;
    register_number?: number;
    register_folio?: string;
    register_tome?: string;
    register_title_date?: string;
    register_title_state?: string;
  };
  showUniversity: boolean;
  showGraduateDate: boolean;
  showMention: boolean;
  files: { [key: string]: File };
  setFiles: (files: any) => void;
}

export function AcademicSection(props: AcademicSectionProps) {
  const [modalImage, setModalImage] = createSignal<{ src: string; alt: string } | null>(null);

  const handleFileChange = (e: Event, key: string) => {
    const target = e.target as HTMLInputElement;
    if (target.files && target.files[0]) {
      props.setFiles({ ...props.files, [key]: target.files[0] });
    }
  };

  const handleRemoveFile = (key: string) => {
    const newFiles = { ...props.files };
    delete newFiles[key];
    props.setFiles(newFiles);
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

  const getImageUrl = (url?: string) => {
    if (!url) return "";
    return `http://localhost:9000/colpsi-bucket/${url}`;
  };

  const openModal = (url: string, title: string) => {
    setModalImage({ src: getImageUrl(url), alt: title });
  };

  const closeModal = () => {
    setModalImage(null);
  };

  return (
    <section class="bg-white rounded-[2.5rem] p-8 shadow-premium border border-gray-100">
      {/* Modal para imágenes */}
      <ImageModal 
        src={modalImage()?.src || ""}
        alt={modalImage()?.alt || ""}
        isOpen={!!modalImage()}
        onClose={closeModal}
      />

      <div class="mb-6 border-l-4 border-colpsi-yellow pl-3">
        <h2 class="text-xl font-black text-colpsi-blue">Formación de Pregrado</h2>
        <p class="text-xs text-colpsi-muted mt-1">
          Datos certificados por el colegio. Contacta a administración para corregir datos erróneos.
        </p>
      </div>

      <div class="space-y-6">
        {/* Tarjeta de Pregrado */}
        <div class="bg-gray-50 p-6 rounded-2xl border border-gray-100">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
            {/* Universidad */}
            <Show when={props.showUniversity}>
              <div>
                <label class="text-[10px] font-black text-gray-400 uppercase block mb-1">
                  Universidad
                </label>
                <p class="font-bold text-colpsi-blue text-lg">
                  {props.undergraduateData.university_undergraduate || "No especificada"}
                </p>
              </div>
            </Show>

            {/* Fecha de Egreso */}
            <Show when={props.showGraduateDate}>
              <div>
                <label class="text-[10px] font-black text-gray-400 uppercase block mb-1">
                  Fecha de Egreso
                </label>
                <p class="font-bold text-colpsi-blue">
                  {formatDate(props.undergraduateData.graduate_date) || "No especificada"}
                </p>
              </div>
            </Show>

            {/* Mención */}
            <Show when={props.showMention}>
              <div>
                <label class="text-[10px] font-black text-gray-400 uppercase block mb-1">
                  Mención
                </label>
                <p class="font-bold text-colpsi-blue">
                  {props.undergraduateData.mention_undergraduate || "No especificada"}
                </p>
              </div>
            </Show>

            {/* Número de Registro */}
            <Show when={props.undergraduateData.register_number}>
              <div>
                <label class="text-[10px] font-black text-gray-400 uppercase block mb-1">
                  N° Registro
                </label>
                <p class="font-bold text-colpsi-blue">
                  {props.undergraduateData.register_number}
                </p>
              </div>
            </Show>

            {/* Folio y Tomo */}
            <Show when={props.undergraduateData.register_folio || props.undergraduateData.register_tome}>
              <div class="md:col-span-2">
                <label class="text-[10px] font-black text-gray-400 uppercase block mb-1">
                  Folio / Tomo
                </label>
                <p class="font-bold text-colpsi-blue">
                  {props.undergraduateData.register_folio && `Folio: ${props.undergraduateData.register_folio}`}
                  {props.undergraduateData.register_folio && props.undergraduateData.register_tome && " • "}
                  {props.undergraduateData.register_tome && `Tomo: ${props.undergraduateData.register_tome}`}
                </p>
              </div>
            </Show>
          </div>

          {/* Gestión de imágenes del título */}
          <div class="space-y-4">
            <label class="text-[10px] font-black text-gray-400 uppercase">
              Documentos del Título
            </label>
            
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
              {/* Imagen 1 */}
              <div class="space-y-3">
                <div class="flex items-center justify-between">
                  <span class="text-xs font-medium text-gray-600">Imagen del Título</span>
                  <Show when={props.undergraduateData.title_image_one_url && !props.files.title_image_one}>
                    <span class="text-[10px] text-green-600 bg-green-50 px-2 py-0.5 rounded-full">
                      Subida
                    </span>
                  </Show>
                </div>

                {/* Preview de imagen existente */}
                <Show when={props.undergraduateData.title_image_one_url && !props.files.title_image_one}>
                  <div 
                    class="relative group cursor-pointer"
                    onClick={() => openModal(
                      props.undergraduateData.title_image_one_url!, 
                      "Imagen del Título"
                    )}
                  >
                    <img 
                      src={getImageUrl(props.undergraduateData.title_image_one_url)} 
                      alt="Título"
                      class="w-full h-32 object-cover rounded-lg border-2 border-gray-200 group-hover:border-colpsi-yellow transition-all"
                    />
                    <div class="absolute inset-0 bg-black/0 group-hover:bg-black/30 rounded-lg transition-all flex items-center justify-center opacity-0 group-hover:opacity-100">
                      <span class="bg-white text-colpsi-blue p-2 rounded-full shadow-lg transform scale-90 group-hover:scale-100 transition-transform">
                        🔍
                      </span>
                    </div>
                  </div>
                </Show>

                {/* Preview de archivo seleccionado */}
                <Show when={props.files.title_image_one}>
                  <div class="relative">
                    <div 
                      class="w-full h-32 bg-blue-50 rounded-lg border-2 border-colpsi-blue flex flex-col items-center justify-center cursor-pointer hover:bg-blue-100 transition-colors"
                      onClick={() => {
                        const file = props.files.title_image_one;
                        const url = URL.createObjectURL(file);
                        openModal(url, file.name);
                      }}
                    >
                      <span class="text-3xl mb-2">📄</span>
                      <span class="text-[10px] font-medium text-colpsi-blue text-center px-2 truncate max-w-full">
                        {props.files.title_image_one?.name}
                      </span>
                      <span class="text-[8px] text-gray-500">
                        {(props.files.title_image_one!.size / 1024).toFixed(1)}KB
                      </span>
                    </div>
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleRemoveFile("title_image_one");
                      }}
                      class="absolute -top-2 -right-2 bg-red-500 text-white rounded-full w-6 h-6 flex items-center justify-center text-sm hover:bg-red-600 transition-colors"
                    >
                      ✕
                    </button>
                  </div>
                </Show>

                <FileUploader 
                  id="title_image_one" 
                  label={props.files.title_image_one ? "Cambiar archivo" : "Seleccionar archivo"} 
                  onChange={(e) => handleFileChange(e, "title_image_one")} 
                />
              </div>

              {/* Imagen 2 - similar a la primera */}
              <div class="space-y-3">
                <div class="flex items-center justify-between">
                  <span class="text-xs font-medium text-gray-600">Documento Adicional 1</span>
                  <Show when={props.undergraduateData.title_image_two_url && !props.files.title_image_two}>
                    <span class="text-[10px] text-green-600 bg-green-50 px-2 py-0.5 rounded-full">
                      Subida
                    </span>
                  </Show>
                </div>

                <Show when={props.undergraduateData.title_image_two_url && !props.files.title_image_two}>
                  <div 
                    class="relative group cursor-pointer"
                    onClick={() => openModal(
                      props.undergraduateData.title_image_two_url!, 
                      "Documento Adicional 1"
                    )}
                  >
                    <img 
                      src={getImageUrl(props.undergraduateData.title_image_two_url)} 
                      alt="Documento 1"
                      class="w-full h-32 object-cover rounded-lg border-2 border-gray-200 group-hover:border-colpsi-yellow transition-all"
                    />
                    <div class="absolute inset-0 bg-black/0 group-hover:bg-black/30 rounded-lg transition-all flex items-center justify-center opacity-0 group-hover:opacity-100">
                      <span class="bg-white text-colpsi-blue p-2 rounded-full shadow-lg transform scale-90 group-hover:scale-100 transition-transform">
                        🔍
                      </span>
                    </div>
                  </div>
                </Show>

                <Show when={props.files.title_image_two}>
                  <div class="relative">
                    <div 
                      class="w-full h-32 bg-blue-50 rounded-lg border-2 border-colpsi-blue flex flex-col items-center justify-center cursor-pointer hover:bg-blue-100 transition-colors"
                      onClick={() => {
                        const file = props.files.title_image_two;
                        const url = URL.createObjectURL(file);
                        openModal(url, file.name);
                      }}
                    >
                      <span class="text-3xl mb-2">📄</span>
                      <span class="text-[10px] font-medium text-colpsi-blue text-center px-2 truncate max-w-full">
                        {props.files.title_image_two?.name}
                      </span>
                      <span class="text-[8px] text-gray-500">
                        {(props.files.title_image_two!.size / 1024).toFixed(1)}KB
                      </span>
                    </div>
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleRemoveFile("title_image_two");
                      }}
                      class="absolute -top-2 -right-2 bg-red-500 text-white rounded-full w-6 h-6 flex items-center justify-center text-sm hover:bg-red-600 transition-colors"
                    >
                      ✕
                    </button>
                  </div>
                </Show>

                <FileUploader 
                  id="title_image_two" 
                  label={props.files.title_image_two ? "Cambiar archivo" : "Seleccionar archivo"} 
                  onChange={(e) => handleFileChange(e, "title_image_two")} 
                />
              </div>

              {/* Imagen 3 - similar a la primera */}
              <div class="space-y-3">
                <div class="flex items-center justify-between">
                  <span class="text-xs font-medium text-gray-600">Documento Adicional 2</span>
                  <Show when={props.undergraduateData.title_image_three_url && !props.files.title_image_three}>
                    <span class="text-[10px] text-green-600 bg-green-50 px-2 py-0.5 rounded-full">
                      Subida
                    </span>
                  </Show>
                </div>

                <Show when={props.undergraduateData.title_image_three_url && !props.files.title_image_three}>
                  <div 
                    class="relative group cursor-pointer"
                    onClick={() => openModal(
                      props.undergraduateData.title_image_three_url!, 
                      "Documento Adicional 2"
                    )}
                  >
                    <img 
                      src={getImageUrl(props.undergraduateData.title_image_three_url)} 
                      alt="Documento 2"
                      class="w-full h-32 object-cover rounded-lg border-2 border-gray-200 group-hover:border-colpsi-yellow transition-all"
                    />
                    <div class="absolute inset-0 bg-black/0 group-hover:bg-black/30 rounded-lg transition-all flex items-center justify-center opacity-0 group-hover:opacity-100">
                      <span class="bg-white text-colpsi-blue p-2 rounded-full shadow-lg transform scale-90 group-hover:scale-100 transition-transform">
                        🔍
                      </span>
                    </div>
                  </div>
                </Show>

                <Show when={props.files.title_image_three}>
                  <div class="relative">
                    <div 
                      class="w-full h-32 bg-blue-50 rounded-lg border-2 border-colpsi-blue flex flex-col items-center justify-center cursor-pointer hover:bg-blue-100 transition-colors"
                      onClick={() => {
                        const file = props.files.title_image_three;
                        const url = URL.createObjectURL(file);
                        openModal(url, file.name);
                      }}
                    >
                      <span class="text-3xl mb-2">📄</span>
                      <span class="text-[10px] font-medium text-colpsi-blue text-center px-2 truncate max-w-full">
                        {props.files.title_image_three?.name}
                      </span>
                      <span class="text-[8px] text-gray-500">
                        {(props.files.title_image_three!.size / 1024).toFixed(1)}KB
                      </span>
                    </div>
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleRemoveFile("title_image_three");
                      }}
                      class="absolute -top-2 -right-2 bg-red-500 text-white rounded-full w-6 h-6 flex items-center justify-center text-sm hover:bg-red-600 transition-colors"
                    >
                      ✕
                    </button>
                  </div>
                </Show>

                <FileUploader 
                  id="title_image_three" 
                  label={props.files.title_image_three ? "Cambiar archivo" : "Seleccionar archivo"} 
                  onChange={(e) => handleFileChange(e, "title_image_three")} 
                />
              </div>
            </div>

            <p class="text-[9px] text-gray-400 mt-4">
              ✓ Haz clic en cualquier imagen para verla en tamaño completo.
              <br />
              Las imágenes subidas se muestran arriba. Para reemplazar una imagen, selecciona un nuevo archivo.
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}