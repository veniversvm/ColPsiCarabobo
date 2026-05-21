// web/src/components/psi/ProfileHeader.tsx
// Cabecera del perfil con foto y datos básicos
import { Show, For, createSignal } from "solid-js";
import { Portal } from "solid-js/web";
import QRCodeGenerator from "./profile/QrCode";

// ── Variable de entorno con fallback al entorno de desarrollo ──
const BUCKET_URL = import.meta.env.VITE_BUCKET_URL || "http://localhost:9000/colpsi-bucket";

interface ProfileHeaderProps {
  firstName: string;
  secondName?: string;
  lastName: string;
  secondLastName?: string;
  fpv: number;
  ci: number;
  profilePicture?: string;
  specialties: string[];
  url: string;
}

export function ProfileHeader(props: ProfileHeaderProps) {
  // Estado para controlar el modal de la foto ampliada
  const [isModalOpen, setIsModalOpen] = createSignal(false);

  // Helper para construir el nombre completo
  const fullName = () => [props.firstName, props.secondName, props.lastName, props.secondLastName]
    .filter(Boolean)
    .join(" ");

  // Helper para la URL de la imagen
  const imageUrl = () => props.profilePicture ? `${BUCKET_URL}/${props.profilePicture}` : null;

  return (
    <>
      <div class="bg-white rounded-[2.5rem] p-6 shadow-premium border border-gray-100 text-center relative overflow-hidden">
        
        {/* ── AVATAR ────────────────────────────────────────────────────── */}
        <div 
          class={`w-24 h-24 md:w-32 md:h-32 mx-auto bg-gray-50 rounded-full overflow-hidden border-4 border-colpsi-yellow shadow-inner mb-5 relative group ${props.profilePicture ? 'cursor-pointer' : ''}`}
          onClick={() => {
            if (props.profilePicture) setIsModalOpen(true);
          }}
          title={props.profilePicture ? "Hacer clic para ampliar foto" : ""}
        >
          <Show
            when={props.profilePicture && imageUrl()}
            fallback={
              <div class="w-full h-full flex items-center justify-center text-4xl bg-blue-50">
                👤
              </div>
            }
          >
            <img
              src={imageUrl()!}
              alt={`Dr(a). ${props.lastName}`}
              class="w-full h-full object-cover transition-transform duration-300 group-hover:scale-110"
            />
            
            {/* Overlay sutil al hacer hover para indicar que es clickeable */}
            <div class="absolute inset-0 bg-blue-900/20 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center text-white text-2xl backdrop-blur-[1px]">
              🔍
            </div>
          </Show>
        </div>

        {/* ── DATOS BÁSICOS ──────────────────────────────────────────────── */}
        <h1 class="text-xl md:text-2xl font-black text-colpsi-blue leading-tight mb-2">
          {fullName()}
        </h1>
        
        <div class="flex flex-col sm:flex-row justify-center items-center gap-1 sm:gap-4 mb-4">
          <p class="text-gray-500 font-bold tracking-widest uppercase text-xs">
            FPV: <span class="text-colpsi-blue">{props.fpv}</span>
          </p>
          <Show when={props.ci}>
            <span class="hidden sm:inline text-gray-300">•</span>
            <p class="text-gray-500 font-bold tracking-widest uppercase text-xs">
              CI: <span class="text-gray-700">{props.ci}</span>
            </p>
          </Show>
        </div>

        {/* ── ÁREAS DE DESEMPEÑO (Antes Specialties) ────────────────────── */}
        <div class="mt-4 flex justify-center gap-2 flex-wrap">
          <For each={props.specialties}>
            {(spec) => (
              <span class="bg-blue-50 border border-blue-100 text-colpsi-blue text-[10px] md:text-xs font-black uppercase tracking-wider px-3 py-1.5 rounded-xl">
                {spec}
              </span>
            )}
          </For>
        </div>

        {/* ── QR CODE ───────────────────────────────────────────────────── */}
        <div class="mt-8 flex flex-col items-center justify-center w-full border-t border-gray-100 pt-6">
          <p class="text-[10px] text-gray-400 font-bold uppercase tracking-[0.2em] mb-3">
            Ficha Digital
          </p>
          <QRCodeGenerator url={props.url} />
        </div>
      </div>

      {/* ── MODAL DE IMAGEN AMPLIADA ────────────────────────────────────── */}
      <Show when={isModalOpen() && imageUrl()}>
        <Portal>
          <div 
            class="fixed inset-0 z-[100] flex items-center justify-center bg-blue-900/90 backdrop-blur-sm p-4 animate-in fade-in duration-200 cursor-zoom-out"
            onClick={() => setIsModalOpen(false)} // Cierra al hacer clic afuera
          >
            <div class="relative w-full max-w-xl flex justify-center items-center">
              
              {/* Botón de cerrar */}
              <button 
                onClick={() => setIsModalOpen(false)}
                class="absolute -top-12 right-0 md:-right-12 w-10 h-10 bg-white/20 text-white hover:bg-colpsi-yellow hover:text-colpsi-blue rounded-full flex items-center justify-center font-black text-xl backdrop-blur-md transition-all z-10 border border-white/30 shadow-lg"
                title="Cerrar imagen"
              >
                ✕
              </button>
              
              {/* Imagen forzada a verse grande */}
              <img 
                src={imageUrl()!} 
                alt={`Foto ampliada de ${fullName()}`}
                class="w-full max-h-[85vh] min-h-[300px] object-contain rounded-2xl shadow-[0_20px_50px_rgba(0,0,0,0.5)] bg-gray-50 cursor-default" 
                onClick={(e) => e.stopPropagation()} // Previene que se cierre al hacer clic en la imagen
              />
            </div>
          </div>
        </Portal>
      </Show>
    </>
  );
}