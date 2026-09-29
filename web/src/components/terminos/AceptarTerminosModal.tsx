// web/src/components/terminos/AceptarTerminosModal.tsx
// Aviso de aceptación de las condiciones del portal. Aparece en TODO el portal
// /psi/* hasta
// que el agremiado acepta la versión vigente.
//
// ── LA REGLA MÁS IMPORTANTE DE ESTE ARCHIVO ──
//
// Si `GET /psi/me/terms` falla, este componente NO muestra el modal y el portal
// funciona por completo. Nunca bloquea por un fallo del backend.
//
// La razón es que este aviso es un requisito de exquisito legal, no una función
// de negocio. Comprometer el acceso de varioscientos agremiados porque una tabla
// no estaba migrada, el endpoint tardó, o una noche el service worker sirvió una
// versión vieja sería un daño real a cambio de un beneficio nulo. El backend
// además NUNCA bloquea (`psi_service_terms.go`), para que esta sea una decisión
// puramente del cliente y no una defensa a medias.

import { Show, createEffect, createSignal, onMount } from "solid-js";
import { A } from "@solidjs/router";
import { apiGet, apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { TERMINOS_VERSION, COLEGIO } from "~/lib/terminos";

export interface TermsStatus {
  current_version: string;
  accepted: boolean;
  accepted_at: string | null;
  history: Array<{ version: string; accepted_at: string }>;
}

export default function AceptarTerminosModal(props: { activo: boolean }) {
  const [visible, setVisible] = createSignal(false);
  const [estado, setEstado] = createSignal<TerminosStatus | null>(null);
  const [cargando, setCargando] = createSignal(false);
  const [error, setError] = createSignal("");
  const [yaConsultado, setYaConsultado] = createSignal(false);

  /**
   * Versión contra la que se acepta. Se toma de la API cuando llega; la
   * constante del frontend solo es el valor de arranque. Si difieren, gana la
   * API: aceptar el texto que el servidor no reconoce devuelve 409.
   */
  const version = () => estado()?.current_version || TERMINOS_VERSION;

  const consultar = async () => {
    try {
      const data = await apiGet<TermsStatus>("/psi/me/terms");
      setEstado(data);
      // El modal aparece solo si hay sesión y la versión vigente NO está
      // aceptada. Un 401 no abre nada (todavía no sabemos si hay sesión).
      setVisible(props.activo && !data.accepted);
    } catch {
      // Degradación segura: sin estado conocido, no se molesta al usuario.
      setVisible(false);
    } finally {
      setYaConsultado(true);
    }
  };

  const aceptar = async () => {
    setCargando(true);
    setError("");
    try {
      const data = await apiPost<TermsStatus>("/psi/me/terms", { version: version() });
      setEstado(data);
      setVisible(false);
    } catch (err: unknown) {
      const e = err as { status?: number; current_version?: string };
      if (e?.status === 409 && e.current_version) {
        // El texto cambió mientras el modal estaba abierto: se reintenta con la
        // versión que el servidor dice tener, sin dejar al agremiado trapped.
        setEstado((prev) =>
          prev ? { ...prev, current_version: e.current_version as string, accepted: false } : prev,
        );
        setError("El Colegio publicó una versión nueva de los términos. Revísala y vuelve a aceptar.");
      } else {
        setError(getUserFacingError(err));
      }
    } finally {
      setCargando(false);
    }
  };

  onMount(() => {
    if (props.activo) consultar();
  });

  // Si la ruta cambia dentro del portal, el modal no se re-consulta: el estado
  // es el mismo para todas las páginas y una petición por navegación sería
  // tráfico puro. Solo se re-consulta si `activo` pasa de falso a verdadero.
  createEffect(() => {
    const a = props.activo;
    if (a && !yaConsultado()) consultar();
  });

  return (
    <Show when={visible()}>
      <div class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
        <div
          role="dialog"
          aria-modal="true"
          aria-labelledby="titulo-aceptar-terminos"
          class="bg-white rounded-3xl shadow-2xl border border-colpsi-border w-full max-w-lg overflow-hidden max-h-[90vh] flex flex-col"
        >
          <div class="bg-colpsi-blue px-6 py-5 text-white shrink-0">
            <div class="flex items-center gap-3">
              <span aria-hidden="true" class="text-2xl">📜</span>
              <h2 id="titulo-aceptar-terminos" class="font-black text-lg tracking-tight">
                Términos y Condiciones
              </h2>
            </div>
            <p class="text-blue-200 text-xs mt-1">
              Versión {version()} · Colegio de Psicólogos del Estado Carabobo
            </p>
          </div>

          <div class="p-6 overflow-y-auto flex-1">
            <p class="text-gray-700 text-sm leading-relaxed text-justify mb-4">
              El Colegio de Psicólogos del Estado Carabobo actualizó sus Términos y
              Condiciones para el uso del portal del agremiado. Antes de seguir
              usando el portal, debe leer y aceptar las condiciones del portal.
            </p>

            <div class="bg-blue-50 border-l-4 border-colpsi-blue rounded-2xl p-4 mb-5">
              <p class="text-sm text-blue-950 leading-relaxed text-justify">
                Al aceptar, el Colegio registra su aceptación junto con la fecha y
                la dirección desde la que la realizó. Ese registro no se puede
                modificar ni borrar.
              </p>
            </div>

            <Show when={error()}>
              <div class="bg-red-50 border border-red-200 text-red-800 rounded-2xl p-3 mb-5 text-sm">
                {error()}
              </div>
            </Show>

            <p class="text-xs text-gray-500 leading-relaxed">
              Si tiene una duda sobre el contenido, escríbanos antes de aceptar a{" "}
              <span class="font-bold text-gray-700">{COLEGIO.correoReportes}</span>.
            </p>
          </div>

          <div class="border-t border-colpsi-border p-5 flex flex-col gap-3 shrink-0">
            <A
              href="/psi/terminos"
              target="_blank"
              class="text-center text-xs font-bold text-colpsi-blue underline underline-offset-2 hover:text-colpsi-red"
            >
              Leer las condiciones del portal antes de aceptar
            </A>
            <button
              type="button"
              onClick={aceptar}
              disabled={cargando()}
              class="inline-flex items-center justify-center gap-2 bg-colpsi-blue text-white font-black px-6 py-3.5 rounded-xl hover:bg-blue-800 transition-colors uppercase text-xs tracking-widest shadow-lg shadow-colpsi-blue/20 disabled:opacity-60 disabled:cursor-not-allowed"
            >
              <Show when={cargando()} fallback={<>Aceptar y continuar</>}>
                Registrando aceptación…
              </Show>
            </button>
          </div>
        </div>
      </div>
    </Show>
  );
}
