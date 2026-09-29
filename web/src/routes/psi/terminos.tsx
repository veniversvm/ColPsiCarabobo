// web/src/routes/psi/terminos.tsx → /psi/terminos
//
// Documento de condiciones del portal del agremiado (solo con sesión).. Además del texto,
// muestra el estado de aceptación: es lo que el agremiado necesita para
// confirmar que su registro quedó hecho, y lo que le da trazabilidad cuando
// vuelve a entrar.

import { Show, For, createSignal, createEffect, onMount } from "solid-js";
import { A, useNavigate } from "@solidjs/router";
import { apiGet, apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { useAuth } from "~/lib/auth";
import TerminosLayout from "~/components/terminos/TerminosLayout";
import type { TermsStatus } from "~/components/terminos/AceptarTerminosModal";
import { PENDIENTE } from "~/lib/terminos/parte-ii-agremiado";

/** Formatea una fecha ISO a formato local legible. Cliente-solo (se llama tras montar). */
function fechaLegible(iso: string | null | undefined) {
  if (!iso) return "—";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleDateString("es-VE", {
    day: "2-digit",
    month: "long",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default function TerminosParteIIAgremiado() {
  const { sessionReady, isAuthenticated } = useAuth();
  const navigate = useNavigate();

  const [estado, setEstado] = createSignal<TermsStatus | null>(null);
  const [cargando, setCargando] = createSignal(false);
  const [error, setError] = createSignal("");

  // Guarda: este documento es material interno. Se espera a `sessionReady()` antes
  // de concluir nada — durante SSR y en el primer render `user()` es null
  // aunque haya sesión, y un `!isAuthenticated()` desnudo expulsaría al
  // agremiado en pestaña nueva (el caso que arregla `restoreSessionAction`).
  createEffect(() => {
    if (sessionReady() && !isAuthenticated()) {
      navigate("/terminos", { replace: true });
    }
  });

  const consultar = async () => {
    try {
      setEstado(await apiGet<TermsStatus>("/psi/me/terms"));
    } catch {
      // Sin estado conocido no se muestra panel de aceptación: la lectura del
      // documento no depende de él y el portal sigue igual (misma degradación
      // segura que el modal).
      setEstado(null);
    }
  };

  const aceptar = async () => {
    const version = estado()?.current_version;
    if (!version) return;
    setCargando(true);
    setError("");
    try {
      setEstado(await apiPost<TermsStatus>("/psi/me/terms", { version }));
    } catch (err: unknown) {
      setError(getUserFacingError(err));
    } finally {
      setCargando(false);
    }
  };

  // Solo se consulta con sesión: el endpoint es 401 sin ella y el fetch sería
  // un error garantizado en el camino de redirección.
  onMount(() => {
    if (isAuthenticated()) consultar();
  });

  // Antes de saber si hay sesión no se pinta el documento: si el visitante va a
  // ser redirigido, no debe ver un frame del documento.
  return (
    <Show
      when={sessionReady()}
      fallback={
        <main class="min-h-screen bg-colpsi-bg flex items-center justify-center font-sans">
          <div class="text-center">
            <div class="w-8 h-8 border-4 border-blue-100 border-t-colpsi-blue rounded-full animate-spin mx-auto" />
            <p class="text-xs text-gray-500 font-bold uppercase tracking-widest mt-4">
              Verificando sesión…
            </p>
          </div>
        </main>
      }
    >
      <Show when={isAuthenticated()} fallback={null}>
        <TerminosLayout clave="agremiado" pendiente={PENDIENTE}>
          <div class="bg-white rounded-3xl shadow-premium border border-colpsi-border p-6 md:p-8 mt-10">
            <h2 class="text-lg font-black text-colpsi-blue mb-1">Su registro de aceptación</h2>
            <p class="text-sm text-gray-600 mb-5">
              El Colegio guarda cada aceptación con su fecha y la dirección desde la que
              se realizó. Ese registro no se puede modificar ni borrar.
            </p>

            <Show when={error()}>
              <div class="bg-red-50 border border-red-200 text-red-800 rounded-2xl p-3 mb-4 text-sm">
                {error()}
              </div>
            </Show>

            <Show
              when={estado()}
              fallback={
                <p class="text-sm text-gray-500">
                  No fue posible consultar su registro de aceptación en este momento.
                </p>
              }
            >
              {(e) => (
                <>
                  <Show
                    when={e().accepted}
                    fallback={
                      <div class="bg-blue-50 border-l-4 border-colpsi-blue rounded-2xl p-4">
                        <p class="text-sm text-blue-950 mb-4">
                          Todavía no ha aceptado la versión vigente
                          {" "}<span class="font-black">{e().current_version}</span>.
                        </p>
                        <button
                          type="button"
                          onClick={aceptar}
                          disabled={cargando()}
                          class="inline-flex items-center gap-2 bg-colpsi-blue text-white font-black px-6 py-3 rounded-xl hover:bg-blue-800 transition-colors uppercase text-xs tracking-widest shadow-md disabled:opacity-60"
                        >
                          <Show when={cargando()} fallback={<>Aceptar y continuar</>}>
                            Registrando…
                          </Show>
                        </button>
                      </div>
                    }
                  >
                    <div class="bg-emerald-50 border-l-4 border-emerald-600 rounded-2xl p-4 mb-5">
                      <p class="text-sm text-emerald-900 font-bold mb-1">
                        Aceptó la versión {e().current_version}
                      </p>
                      <p class="text-xs text-emerald-800">
                        {fechaLegible(e().accepted_at)}
                      </p>
                    </div>
                  </Show>

                  <Show when={e().history.length > 1}>
                    <details class="text-sm">
                      <summary class="cursor-pointer font-bold text-colpsi-blue">
                        Historial completo ({e().history.length} aceptaciones)
                      </summary>
                      <ul class="mt-3 space-y-2">
                        <For each={e().history}>
                          {(h) => (
                            <li class="flex justify-between gap-4 text-gray-600 border-b border-gray-100 pb-2">
                              <span class="font-mono text-xs">{h.version}</span>
                              <span class="text-xs">{fechaLegible(h.accepted_at)}</span>
                            </li>
                          )}
                        </For>
                      </ul>
                    </details>
                  </Show>
                </>
              )}
            </Show>

            <div class="mt-6 pt-5 border-t border-colpsi-border">
              <A
                href="/psi"
                class="inline-flex items-center gap-2 text-xs font-black uppercase tracking-widest text-colpsi-blue hover:text-colpsi-red transition-colors"
              >
                ← Volver al portal
              </A>
            </div>
          </div>
        </TerminosLayout>
      </Show>
    </Show>
  );
}
