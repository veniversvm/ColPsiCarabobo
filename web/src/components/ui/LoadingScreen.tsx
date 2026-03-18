// web/src/components/ui/LoadingScreen.tsx
import { Show } from "solid-js";

interface LoadingScreenProps {
  /** Imagen opcional — logo, ilustración, etc. Si no se pasa usa el símbolo Ψ */
  image?: string;
  /** Alt text para la imagen */
  imageAlt?: string;
  /** Mensaje principal. Default: "Cargando..." */
  message?: string;
  /** Mensaje secundario opcional */
  submessage?: string;
  /** Ocupa toda la pantalla (fixed inset-0). Default: false — ocupa el contenedor padre */
  fullscreen?: boolean;
  /** Tamaño del círculo de imagen en px. Default: 140 */
  size?: number;
  /** Tamaño interno de la imagen como porcentaje del círculo. Default: 75 */
  imageScale?: number;
}

/**
 * LoadingScreen — componente de carga institucional ColPsi.
 *
 * Uso mínimo:
 *   <LoadingScreen />
 *
 * Con imagen:
 *   <LoadingScreen image="/logo.png" message="Verificando sesión..." />
 *
 * Pantalla completa:
 *   <LoadingScreen fullscreen message="Iniciando sistema..." />
 */
export function LoadingScreen(props: LoadingScreenProps) {
  const message    = () => props.message    ?? "Cargando...";
  const submessage = () => props.submessage ?? "";
  const fullscreen = () => props.fullscreen ?? false;
  const scale      = () => props.imageScale ?? 75;

  return (
    <div
      class={`flex flex-col items-center justify-center bg-white ${
        fullscreen()
          ? "fixed inset-0 z-50"
          : "w-full min-h-[320px]"
      }`}
    >
      {/* ── Contenedor central ──────────────────────────────────────────── */}
      <div class="flex flex-col items-center gap-6 select-none">

        {/* ── Logo / Imagen ──────────────────────────────────────────────── */}
        <div class="relative">
          {/* Halo animado exterior */}
          <div class="absolute inset-0 rounded-full bg-colpsi-blue/8 animate-ping" style="animation-duration:2s" />
          {/* Anillo giratorio */}
          <div
            class="absolute rounded-full border-2 border-transparent"
            style={`
              inset: -8px;
              border-top-color: #1e3a8a;
              border-right-color: #facc15;
              animation: colpsi-spin 1.1s linear infinite;
            `}
          />

          {/* Imagen o símbolo Ψ */}
          <div
            class="relative rounded-full bg-white flex items-center justify-center shadow-lg overflow-hidden"
            style={`width:${props.size ?? 140}px; height:${props.size ?? 140}px`}
          >
            <Show
              when={props.image}
              fallback={
                <span class="text-colpsi-blue font-black text-4xl leading-none" style="font-family: serif">
                  Ψ
                </span>
              }
            >
              <img
                src={props.image}
                alt={props.imageAlt ?? "Cargando"}
                style={`width:${scale()}%; height:${scale()}%`}
                class="object-contain"
              />
            </Show>
          </div>
        </div>

        {/* ── Texto ──────────────────────────────────────────────────────── */}
        <div class="text-center space-y-1.5">
          <p class="text-colpsi-blue font-black text-base tracking-wide uppercase">
            {message()}
          </p>
          <Show when={submessage()}>
            <p class="text-colpsi-muted text-xs font-medium">{submessage()}</p>
          </Show>
        </div>

        {/* ── Barra de progreso indeterminada ────────────────────────────── */}
        <div class="w-48 h-1 bg-gray-100 rounded-full overflow-hidden">
          <div
            class="h-full rounded-full"
            style="
              width: 40%;
              background: linear-gradient(90deg, #1e3a8a, #facc15);
              animation: colpsi-bar 1.4s ease-in-out infinite;
            "
          />
        </div>

      </div>

      {/* ── Keyframes inline — no requieren tailwind plugin ────────────── */}
      <style>{`
        @keyframes colpsi-spin {
          to { transform: rotate(360deg); }
        }
        @keyframes colpsi-bar {
          0%   { transform: translateX(-160%); }
          50%  { transform: translateX(60%);   }
          100% { transform: translateX(260%);  }
        }
      `}</style>
    </div>
  );
}