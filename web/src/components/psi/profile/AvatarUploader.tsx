// web/src/components/psi/profile/AvatarUploader.tsx
import { Show, createSignal } from "solid-js";
import { Portal } from "solid-js/web"; // Importante para renderizar el modal sobre todo el DOM
import QRCodeGenerator from "./QrCode";
import { bucketUrl } from "~/lib/bucket";
import { Animate, Presence } from "~/components/ui/Motion";

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
      return `La imagen es demasiado grande (${sizeMB}MB). Máximo 5MB.`;
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
      <section class="bg-white rounded-[2.5rem] p-8 shadow-premium border border-gray-100">
        <div class="flex flex-col md:flex-row items-center gap-8">
          {/* Columna izquierda: Avatar */}
          <div class="flex flex-col items-center">
            <div class="relative group">
              {/* Contenedor de la imagen */}
              <div
                class="w-32 h-32 rounded-full overflow-hidden border-4 border-colpsi-yellow bg-gradient-to-br from-blue-50 to-gray-50 shadow-xl cursor-pointer"
                onClick={() => {
                  if (previewUrl()) setIsModalOpen(true);
                }}
                title="Hacer clic para ver en grande"
              >
                <Show
                  when={previewUrl()}
                  fallback={
                    <div class="w-full h-full flex items-center justify-center text-5xl bg-gradient-to-br from-blue-50 to-gray-100 text-colpsi-blue/50">
                      👤
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
                class="absolute bottom-0 right-0 bg-colpsi-blue text-white p-3 rounded-full cursor-pointer shadow-lg hover:bg-colpsi-yellow hover:text-colpsi-blue transition-all group-hover:scale-110 border-2 border-white"
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
                <span class="text-sm">📷</span>
              </label>
            </div>

            <div class="mt-4 text-center">
              <p class="text-xs font-bold text-gray-400 uppercase tracking-wider">
                Foto de Perfil
              </p>
              <p class="text-[10px] text-gray-400 mt-1 bg-gray-50 px-3 py-1 rounded-full">
                JPG, PNG, GIF • Máx 5MB
              </p>
            </div>
          </div>

          {/* Columna central: Información personal */}
          <div class="flex-1 text-center md:text-left">
            <div class="space-y-3">
              <div>
                <p class="text-xs font-bold text-gray-400 uppercase tracking-wider mb-1">
                  Nombre Completo
                </p>
                <h2 class="text-xl md:text-2xl font-black text-colpsi-blue leading-tight">
                  {fullName()}
                </h2>
              </div>

              <div class="flex flex-col sm:flex-row gap-3 justify-center md:justify-start">
                <div class="bg-gradient-to-br from-colpsi-blue/5 to-blue-50 px-4 py-2 rounded-xl border border-blue-100">
                  <p class="text-[10px] font-bold text-colpsi-blue uppercase tracking-wider">
                    FPV
                  </p>
                  <p class="text-lg font-black text-colpsi-blue">{props.FPV}</p>
                </div>

                <div class="bg-gradient-to-br from-gray-50 to-gray-100 px-4 py-2 rounded-xl border border-gray-200">
                  <p class="text-[10px] font-bold text-gray-500 uppercase tracking-wider">
                    Cédula
                  </p>
                  <p class="text-lg font-black text-gray-700">{props.CI}</p>
                </div>
              </div>

              <Show when={props.avatarFile}>
                <Animate variant="slide-top" class="mt-2 text-xs text-green-600 bg-green-50 px-3 py-2 rounded-lg inline-flex items-center gap-2">
                  <span class="text-lg">✅</span>
                  <div>
                    <p class="font-bold">Nueva foto lista para guardar</p>
                  </div>
                  <button
                    type="button"
                    onClick={() => {
                      cleanupPreview();
                      props.onFileChange(null);
                    }}
                    class="ml-2 text-red-500 hover:text-red-700 hover:bg-red-50 p-1 rounded-full"
                  >
                    ✕
                  </button>
                </Animate>
              </Show>
            </div>
          </div>

          {/* Columna derecha: QR Code */}
          <div class="w-full md:w-auto flex justify-center border-t md:border-t-0 md:border-l border-gray-100 pt-6 md:pt-0 md:pl-8">
            <QRCodeGenerator url={props.url} />
          </div>
        </div>
      </section>

      {/* ── MODAL DE IMAGEN AMPLIADA ── */}
      <Presence>
        <Show when={isModalOpen() && previewUrl()}>
          <Portal>
            <Animate
              variant="fade"
              class="fixed inset-0 z-[9999] flex items-center justify-center bg-blue-900/90 backdrop-blur-sm p-4 cursor-zoom-out"
              exit={{ opacity: 0 }}
              onClick={() => setIsModalOpen(false)}
            >
              <div class="relative w-full max-w-2xl flex justify-center items-center">
                {/* Botón de cerrar */}
                <button
                  onClick={() => setIsModalOpen(false)}
                  class="absolute -top-12 right-0 md:-right-12 w-10 h-10 bg-white/20 text-white hover:bg-colpsi-yellow hover:text-colpsi-blue rounded-full flex items-center justify-center font-black text-xl backdrop-blur-md transition-all z-10 border border-white/30"
                  title="Cerrar imagen"
                >
                  ✕
                </button>

                {/* Imagen en detalle (Forzada a ser grande) */}
                <Animate variant="zoom" class="w-full">
                  <img
                    src={previewUrl()!}
                    alt={`Foto de perfil ampliada`}
                    // Clases CLAVE: w-full fuerza a la imagen a ocupar el ancho del max-w-2xl
                    class="w-full max-h-[85vh] min-h-[300px] object-contain rounded-2xl shadow-2xl bg-gray-50 cursor-default"
                    onClick={(e) => e.stopPropagation()}
                  />
                </Animate>
              </div>
            </Animate>
          </Portal>
        </Show>
      </Presence>
    </>
  );
}
