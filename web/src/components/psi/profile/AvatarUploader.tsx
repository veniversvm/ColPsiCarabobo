// web/src/components/psi/profile/AvatarUploader.tsx
import { Show, createSignal } from "solid-js";
import { Portal } from "solid-js/web"; // Importante para renderizar el modal sobre todo el DOM
import QRCodeGenerator from "./QrCode";
import { bucketUrl } from "~/lib/bucket";
import { Icon } from "~/components/admin/ui/icons";

interface AvatarUploaderProps {
  url: string;
  currentAvatarUrl?: string;
  avatarFile: File | null;
  firstName: string;
  secondName?: string;
  lastName: string;
  secondLastName?: string;
  FPV: number;
  CI: number;
  onFileChange: (file: File | null) => void;
  onError?: (message: string) => void;
}

const MAX_FILE_SIZE = 5 * 1024 * 1024;
const ALLOWED_TYPES = ["image/jpeg", "image/png", "image/jpg", "image/gif"];

export function AvatarUploader(props: AvatarUploaderProps) {
  // ── ESTADO DEL MODAL ──
  const [isModalOpen, setIsModalOpen] = createSignal(false);

  const fullName = () => {
    return [
      props.firstName,
      props.secondName,
      props.lastName,
      props.secondLastName,
    ]
      .filter(Boolean)
      .join(" ");
  };

  const previewUrl = () => {
    if (props.avatarFile) {
      return URL.createObjectURL(props.avatarFile);
    }
    return props.currentAvatarUrl
      ? bucketUrl(props.currentAvatarUrl)
      : null;
  };

  const cleanupPreview = () => {
    if (props.avatarFile && previewUrl()) {
      URL.revokeObjectURL(previewUrl()!);
    }
  };

  const validateFile = (file: File): string | null => {
    if (!ALLOWED_TYPES.includes(file.type))
      return "Formato no permitido. Usa JPG, PNG o GIF.";
    if (file.size > MAX_FILE_SIZE) {
      const sizeMB = (file.size / (1024 * 1024)).toFixed(2);
      return `El archivo pesa ${sizeMB} MB. El máximo es 5 MB.`;
    }
    return null;
  };

  const handleFileChange = (file: File | null) => {
    cleanupPreview();
    if (!file) {
      props.onFileChange(null);
      return;
    }
    const error = validateFile(file);
    if (error) {
      if (props.onError) props.onError(error);
      else alert(error);
      props.onFileChange(null);
      return;
    }
    props.onFileChange(file);
  };

  return (
    <>
      <section class="bg-white rounded-lg border border-colpsi-border p-5">
        <div class="flex flex-col md:flex-row items-center gap-6">
          {/* Columna izquierda: Avatar */}
          <div class="flex flex-col items-center">
            <div class="relative group">
              {/* Contenedor de la imagen */}
              <div
                class="w-32 h-32 rounded-full overflow-hidden border-4 border-white ring-1 ring-colpsi-border bg-colpsi-bg shadow-sm cursor-pointer"
                onClick={() => {
                  if (previewUrl()) setIsModalOpen(true);
                }}
                title="Hacer clic para ver en grande"
              >
                <Show
                  when={previewUrl()}
                  fallback={
                    <div class="w-full h-full flex items-center justify-center text-colpsi-blue/40">
                      <Icon name="user" class="w-12 h-12" />
                    </div>
                  }
                >
                  <img
                    src={previewUrl()!}
                    class="w-full h-full object-cover transition-transform group-hover:scale-110"
                    alt={fullName()}
                    loading="lazy"
                    decoding="async"
                  />
                </Show>
              </div>

              {/* Botón flotante para subir foto */}
              <label
                class="absolute bottom-0 right-0 bg-colpsi-blue text-white p-2.5 rounded-full cursor-pointer shadow-sm hover:bg-colpsi-blue-light transition-colors border-2 border-white"
                title="Cambiar foto de perfil (máx 5MB)"
              >
                <input
                  type="file"
                  class="sr-only"
                  accept="image/jpeg, image/png, image/jpg, image/gif"
                  onChange={(e) => {
                    const file = e.currentTarget.files?.[0];
                    handleFileChange(file || null);
                    e.currentTarget.value = "";
                  }}
                />
                <Icon name="camera" class="w-4 h-4" />
              </label>
            </div>

            <div class="mt-3 text-center">
              <p class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide">
                Foto de perfil
              </p>
              <p class="text-[10px] text-colpsi-muted mt-0.5">
                JPG, PNG, GIF • Máx 5MB
              </p>
            </div>
          </div>

          {/* Columna central: Información personal */}
          <div class="flex-1 text-center md:text-left">
            <div class="space-y-3">
              <div>
                <p class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide mb-1">
                  Nombre completo
                </p>
                <h2 class="text-xl font-semibold text-colpsi-text leading-tight">
                  {fullName()}
                </h2>
              </div>

              <div class="flex flex-col sm:flex-row gap-3 justify-center md:justify-start">
                <div class="rounded-md border border-colpsi-border bg-colpsi-bg px-3 py-2">
                  <p class="text-[10px] font-semibold text-colpsi-muted uppercase tracking-wide">
                    FPV
                  </p>
                  <p class="text-base font-semibold text-colpsi-text">{props.FPV}</p>
                </div>

                <div class="rounded-md border border-colpsi-border bg-colpsi-bg px-3 py-2">
                  <p class="text-[10px] font-semibold text-colpsi-muted uppercase tracking-wide">
                    Cédula
                  </p>
                  <p class="text-base font-semibold text-colpsi-text">{props.CI}</p>
                </div>
              </div>

              <Show when={props.avatarFile}>
                <div class="mt-2 text-xs text-emerald-700 bg-emerald-50 border border-emerald-200 px-3 py-2 rounded-md inline-flex items-center gap-2 animate-in fade-in slide-in-from-top-2">
                  <Icon name="check" class="w-3.5 h-3.5" />
                  <span class="font-medium">Nueva foto lista para guardar</span>
                  <button
                    type="button"
                    onClick={() => {
                      cleanupPreview();
                      props.onFileChange(null);
                    }}
                    class="ml-1 text-red-500 hover:text-red-700 hover:bg-red-50 p-1 rounded-md"
                    title="Descartar foto"
                  >
                    <Icon name="x" class="w-3.5 h-3.5" />
                  </button>
                </div>
              </Show>
            </div>
          </div>

          {/* Columna derecha: QR Code */}
          <div class="w-full md:w-auto flex justify-center border-t md:border-t-0 md:border-l border-colpsi-border pt-4 md:pt-0 md:pl-6">
            <QRCodeGenerator url={props.url} />
          </div>
        </div>
      </section>

      {/* ── MODAL DE IMAGEN AMPLIADA ── */}
      <Show when={isModalOpen() && previewUrl()}>
        <Portal>
          <div
            class="fixed inset-0 z-[9999] flex items-center justify-center bg-slate-900/80 backdrop-blur-sm p-4 animate-in fade-in duration-200 cursor-zoom-out"
            onClick={() => setIsModalOpen(false)}
          >
            <div class="relative w-full max-w-2xl flex justify-center items-center">
              {/* Botón de cerrar */}
              <button
                onClick={() => setIsModalOpen(false)}
                class="absolute -top-12 right-0 md:-right-12 w-9 h-9 bg-white/20 text-white hover:bg-white/30 rounded-full flex items-center justify-center backdrop-blur-md transition-all z-10 border border-white/30"
                title="Cerrar imagen"
              >
                <Icon name="x" class="w-4 h-4" />
              </button>

              {/* Imagen en detalle (Forzada a ser grande) */}
              <img
                src={previewUrl()!}
                alt={`Foto de perfil ampliada`}
                // Clases CLAVE: w-full fuerza a la imagen a ocupar el ancho del max-w-2xl
                class="w-full max-h-[85vh] min-h-[300px] object-contain rounded-lg shadow-2xl bg-white cursor-default"
                onClick={(e) => e.stopPropagation()}
              />
            </div>
          </div>
        </Portal>
      </Show>
    </>
  );
}