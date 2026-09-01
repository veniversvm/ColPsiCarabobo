// web/src/components/psi/profile/AcademicSection.tsx
import { Show, createSignal } from "solid-js";
import { bucketUrl } from "~/lib/bucket";
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
  // Rastrea qué imágenes existentes el usuario quiere borrar
  const [pendingDeletes, setPendingDeletes] = createSignal<Record<string, boolean>>({});

  // console.log(props.undergraduateData)

  const markForDelete = (key: string) =>
    setPendingDeletes((prev) => ({ ...prev, [key]: true }));

  const cancelDelete = (key: string) =>
    setPendingDeletes((prev) => ({ ...prev, [key]: false }));

  const handleFileChange = (e: Event, key: string) => {
    const target = e.target as HTMLInputElement;
    if (target.files && target.files[0]) {
      // Si selecciona un nuevo archivo, cancelar borrado pendiente del slot
      cancelDelete(key);
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
    return bucketUrl(url);
  };

  const openModal = (url: string, title: string) => {
    setModalImage({ src: getImageUrl(url), alt: title });
  };

  const closeModal = () => {
    setModalImage(null);
  };

  // Renderiza un slot de imagen reutilizable
  const ImageSlot = (p: {
    label: string;
    fileKey: string;
    existingUrl?: string;
  }) => {
    const isPendingDelete = () => pendingDeletes()[p.fileKey];
    const newFile = () => props.files[p.fileKey];
    const showExisting = () => !!p.existingUrl && !isPendingDelete() && !newFile();

    return (
      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <span class="text-xs font-medium text-gray-600">{p.label}</span>
          <Show when={showExisting()}>
            <span class="text-[10px] text-green-600 bg-green-50 px-2 py-0.5 rounded-full">Subida</span>
          </Show>
          <Show when={isPendingDelete() && !newFile()}>
            <span class="text-[10px] text-red-500 bg-red-50 px-2 py-0.5 rounded-full">Se eliminará al guardar</span>
          </Show>
        </div>

        {/* Imagen existente */}
        <Show when={showExisting()}>
          <div class="relative group">
            <div
              class="cursor-pointer"
              onClick={() => openModal(p.existingUrl!, p.label)}
            >
              <img
                src={getImageUrl(p.existingUrl)}
                alt={p.label}
                class="w-full h-32 object-cover rounded-lg border-2 border-gray-200 group-hover:border-colpsi-yellow transition-all"
              />
              <div class="absolute inset-0 bg-black/0 group-hover:bg-black/30 rounded-lg transition-all flex items-center justify-center opacity-0 group-hover:opacity-100">
                <span class="bg-white text-colpsi-blue p-2 rounded-full shadow-lg transform scale-90 group-hover:scale-100 transition-transform">
                  🔍
                </span>
              </div>
            </div>
            <button
              type="button"
              onClick={() => markForDelete(p.fileKey)}
              class="absolute -top-2 -right-2 bg-red-500 hover:bg-red-600 text-white rounded-full w-7 h-7 flex items-center justify-center shadow transition-colors"
              title="Eliminar imagen"
            >
              🗑
            </button>
          </div>
        </Show>

        {/* Estado: borrado pendiente */}
        <Show when={isPendingDelete() && !newFile()}>
          <div class="w-full h-32 bg-red-50 rounded-lg border-2 border-dashed border-red-200 flex flex-col items-center justify-center gap-2">
            <span class="text-2xl opacity-40">🗑️</span>
            <p class="text-[10px] text-red-400 font-bold text-center">Se eliminará al guardar</p>
            <button
              type="button"
              onClick={() => cancelDelete(p.fileKey)}
              class="text-[10px] underline text-gray-400 hover:text-colpsi-blue"
            >
              Cancelar
            </button>
          </div>
        </Show>

        {/* Preview de nuevo archivo seleccionado */}
        <Show when={newFile()}>
          <div class="relative">
            <div
              class="w-full h-32 bg-blue-50 rounded-lg border-2 border-colpsi-blue flex flex-col items-center justify-center cursor-pointer hover:bg-blue-100 transition-colors"
              onClick={() => {
                const file = newFile();
                const url = URL.createObjectURL(file);
                setModalImage({ src: url, alt: file.name });
              }}
            >
              <span class="text-3xl mb-2">📄</span>
              <span class="text-[10px] font-medium text-colpsi-blue text-center px-2 truncate max-w-full">
                {newFile()?.name}
              </span>
              <span class="text-[8px] text-gray-500">
                {(newFile()!.size / 1024).toFixed(1)}KB
              </span>
            </div>
            <button
              type="button"
              onClick={() => handleRemoveFile(p.fileKey)}
              class="absolute -top-2 -right-2 bg-red-500 text-white rounded-full w-6 h-6 flex items-center justify-center text-sm hover:bg-red-600 transition-colors"
            >
              ✕
            </button>
          </div>
        </Show>

        <FileUploader
          id={p.fileKey}
          label={newFile() ? "Cambiar archivo" : "Seleccionar archivo"}
          onChange={(e) => handleFileChange(e, p.fileKey)}
        />
      </div>
    );
  };

  return (
    <section>
      <ImageModal
        src={modalImage()?.src || ""}
        alt={modalImage()?.alt || ""}
        isOpen={!!modalImage()}
        onClose={closeModal}
      />



      <div class="space-y-6">
        <div class="bg-gray-50 p-6 rounded-2xl border border-gray-100">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">

            <Show when={props.showUniversity}>
              <div>
                <label class="text-[10px] font-black text-gray-400 uppercase block mb-1">Universidad</label>
                <p class="font-bold text-colpsi-blue text-lg">
                  {props.undergraduateData.university_undergraduate || "No especificada"}
                </p>
              </div>
            </Show>

            <Show when={props.showGraduateDate}>
              <div>
                <label class="text-[10px] font-black text-gray-400 uppercase block mb-1">Fecha de Egreso</label>
                <p class="font-bold text-colpsi-blue">
                  {formatDate(props.undergraduateData.graduate_date) || "No especificada"}
                </p>
              </div>
            </Show>

            <Show when={props.showMention}>
              <div>
                <label class="text-[10px] font-black text-gray-400 uppercase block mb-1">Mención</label>
                <p class="font-bold text-colpsi-blue">
                  {props.undergraduateData.mention_undergraduate || "No especificada"}
                </p>
              </div>
            </Show>

            <Show when={props.undergraduateData.register_number}>
              <div>
                <label class="text-[10px] font-black text-gray-400 uppercase block mb-1">N° Registro</label>
                <p class="font-bold text-colpsi-blue">{props.undergraduateData.register_number}</p>
              </div>
            </Show>

            <Show when={props.undergraduateData.register_folio || props.undergraduateData.register_tome}>
              <div class="md:col-span-2">
                <label class="text-[10px] font-black text-gray-400 uppercase block mb-1">Folio / Tomo</label>
                <p class="font-bold text-colpsi-blue">
                  {props.undergraduateData.register_folio && `Folio: ${props.undergraduateData.register_folio}`}
                  {props.undergraduateData.register_folio && props.undergraduateData.register_tome && " • "}
                  {props.undergraduateData.register_tome && `Tomo: ${props.undergraduateData.register_tome}`}
                </p>
              </div>
            </Show>
          </div>

          <div class="space-y-4">
            <label class="text-[10px] font-black text-gray-400 uppercase">Documentos del Título</label>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
              <ImageSlot
                label="Imagen del Título"
                fileKey="title_image_one"
                existingUrl={props.undergraduateData.title_image_one_url}
              />
              <ImageSlot
                label="Documento Adicional 1"
                fileKey="title_image_two"
                existingUrl={props.undergraduateData.title_image_two_url}
              />
              <ImageSlot
                label="Documento Adicional 2"
                fileKey="title_image_three"
                existingUrl={props.undergraduateData.title_image_three_url}
              />
            </div>
            <p class="text-[9px] text-gray-400 mt-4">
              ✓ Haz clic en cualquier imagen para verla en tamaño completo.
              Para eliminar una imagen existente usa el botón 🗑 — el cambio se aplica al guardar.
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}