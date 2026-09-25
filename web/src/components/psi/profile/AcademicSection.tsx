// web/src/components/psi/profile/AcademicSection.tsx
import { Show, createSignal } from "solid-js";
import { bucketUrl } from "~/lib/bucket";
import { FileUploader } from "~/components/ui/fileUploader";
import { ImageModal } from "~/components/ui/ImageModal";
import { Icon } from "~/components/admin/ui/icons";

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
          <span class="text-xs font-medium text-colpsi-text">{p.label}</span>
          <Show when={showExisting()}>
            <span class="text-[10px] font-medium text-emerald-700 bg-emerald-50 border border-emerald-200 px-2 py-0.5 rounded-md">Subida</span>
          </Show>
          <Show when={isPendingDelete() && !newFile()}>
            <span class="text-[10px] font-medium text-red-500 bg-red-50 border border-red-200 px-2 py-0.5 rounded-md">Se eliminará al guardar</span>
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
                class="w-full h-32 object-cover rounded-md border border-colpsi-border transition-colors"
              />
              <div class="absolute inset-0 bg-black/0 group-hover:bg-black/30 rounded-md transition-all flex items-center justify-center opacity-0 group-hover:opacity-100">
                <span class="bg-white text-colpsi-blue p-2 rounded-full shadow-sm">
                  <Icon name="search" class="w-4 h-4" />
                </span>
              </div>
            </div>
            <button
              type="button"
              onClick={() => markForDelete(p.fileKey)}
              class="absolute -top-2 -right-2 bg-red-500 hover:bg-red-600 text-white rounded-full w-7 h-7 flex items-center justify-center shadow-sm transition-colors"
              title="Eliminar imagen"
            >
              <Icon name="trash" class="w-3.5 h-3.5" />
            </button>
          </div>
        </Show>

        {/* Estado: borrado pendiente */}
        <Show when={isPendingDelete() && !newFile()}>
          <div class="w-full h-32 bg-red-50 rounded-md border border-dashed border-red-200 flex flex-col items-center justify-center gap-2">
            <Icon name="trash" class="w-5 h-5 text-red-300" />
            <p class="text-[11px] text-red-400 font-medium text-center">Se eliminará al guardar</p>
            <button
              type="button"
              onClick={() => cancelDelete(p.fileKey)}
              class="text-[11px] underline text-colpsi-muted hover:text-colpsi-blue"
            >
              Cancelar
            </button>
          </div>
        </Show>

        {/* Preview de nuevo archivo seleccionado */}
        <Show when={newFile()}>
          <div class="relative">
            <div
              class="w-full h-32 bg-colpsi-bg rounded-md border border-colpsi-border flex flex-col items-center justify-center cursor-pointer hover:bg-colpsi-bg/70 transition-colors"
              onClick={() => {
                const file = newFile();
                const url = URL.createObjectURL(file);
                setModalImage({ src: url, alt: file.name });
              }}
            >
              <Icon name="fileText" class="w-6 h-6 text-colpsi-blue/50 mb-1.5" />
              <span class="text-[11px] font-medium text-colpsi-blue text-center px-2 truncate max-w-full">
                {newFile()?.name}
              </span>
              <span class="text-[10px] text-colpsi-muted">
                {(newFile()!.size / 1024).toFixed(1)}KB
              </span>
            </div>
            <button
              type="button"
              onClick={() => handleRemoveFile(p.fileKey)}
              class="absolute -top-2 -right-2 bg-red-500 text-white rounded-full w-6 h-6 flex items-center justify-center hover:bg-red-600 transition-colors"
              title="Quitar archivo"
            >
              <Icon name="x" class="w-3.5 h-3.5" />
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



      <div class="space-y-5">
        <div class="bg-colpsi-bg/50 p-5 rounded-md border border-colpsi-border">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-5 mb-5">

            <Show when={props.showUniversity}>
              <div>
                <label class="text-[10px] font-semibold text-colpsi-muted uppercase block mb-1">Universidad</label>
                <p class="text-base font-semibold text-colpsi-text">
                  {props.undergraduateData.university_undergraduate || "No especificada"}
                </p>
              </div>
            </Show>

            <Show when={props.showGraduateDate}>
              <div>
                <label class="text-[10px] font-semibold text-colpsi-muted uppercase block mb-1">Fecha de Egreso</label>
                <p class="text-sm font-semibold text-colpsi-text">
                  {formatDate(props.undergraduateData.graduate_date) || "No especificada"}
                </p>
              </div>
            </Show>

            <Show when={props.showMention}>
              <div>
                <label class="text-[10px] font-semibold text-colpsi-muted uppercase block mb-1">Mención</label>
                <p class="text-sm font-semibold text-colpsi-text">
                  {props.undergraduateData.mention_undergraduate || "No especificada"}
                </p>
              </div>
            </Show>

            <Show when={props.undergraduateData.register_number}>
              <div>
                <label class="text-[10px] font-semibold text-colpsi-muted uppercase block mb-1">N° Registro</label>
                <p class="text-sm font-semibold text-colpsi-text">{props.undergraduateData.register_number}</p>
              </div>
            </Show>

            <Show when={props.undergraduateData.register_folio || props.undergraduateData.register_tome}>
              <div class="md:col-span-2">
                <label class="text-[10px] font-semibold text-colpsi-muted uppercase block mb-1">Folio / Tomo</label>
                <p class="text-sm font-semibold text-colpsi-text">
                  {props.undergraduateData.register_folio && `Folio: ${props.undergraduateData.register_folio}`}
                  {props.undergraduateData.register_folio && props.undergraduateData.register_tome && " • "}
                  {props.undergraduateData.register_tome && `Tomo: ${props.undergraduateData.register_tome}`}
                </p>
              </div>
            </Show>
          </div>

          <div class="space-y-4">
            <label class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide">Documentos del Título</label>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
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
            <p class="text-[11px] text-colpsi-muted mt-4 leading-relaxed">
              Haz clic en cualquier imagen para verla en tamaño completo.
              Para eliminar una imagen existente usa el botón de eliminar — el cambio se aplica al guardar.
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}