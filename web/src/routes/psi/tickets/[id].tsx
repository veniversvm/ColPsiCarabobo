// web/src/routes/psi/tickets/[id].tsx
// Detalle de una solicitud propia: conversación, historial de estados y cierre.
import { createResource, createSignal, For, Show } from "solid-js";
import { useParams } from "@solidjs/router";
import { apiGet, apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import TicketThread from "~/components/tickets/TicketThread";
import type { Ticket, TicketMensaje, TicketStatusLog } from "~/types/tickets";
import {
  MAX_PSI_MENSAJE_CHARS,
  MAX_CLOSE_REASON_CHARS,
  estadoColor,
  formatTicketDateTime,
  formatTicketDate,
} from "~/types/tickets";

export default function PsiTicketDetalle() {
  const params = useParams<{ id: string }>();

  const [ticket, { refetch }] = createResource<Ticket | null, string>(
    () => params.id,
    async (_id) => {
      try {
        return await apiGet<Ticket>(`/psi/tickets/${params.id}`);
      } catch (e: any) {
        return null;
      }
    }
  );

  const [message, setMessage] = createSignal("");
  const [files, setFiles] = createSignal<File[]>([]);
  const [sending, setSending] = createSignal(false);
  const [composerError, setComposerError] = createSignal("");

  const [closeOpen, setCloseOpen] = createSignal(false);
  const [closeReason, setCloseReason] = createSignal("");
  const [closing, setClosing] = createSignal(false);
  const [closeError, setCloseError] = createSignal("");

  // Mensajes enviados en la sesión actual: se añaden al hilo sin recargar la
  // página (la API devuelve el mensaje creado; no se vuelve a consultar el ticket).
  const [mensajesExtra, setMensajesExtra] = createSignal<TicketMensaje[]>([]);

  const t = () => ticket();
  const closed = () => !!t()?.is_closed || !!t()?.closed_at;

  const submitMensaje = async () => {
    const msg = message().trim();
    if (!msg) {
      setComposerError("Escribe un comentario antes de enviar.");
      return;
    }
    if (msg.length > MAX_PSI_MENSAJE_CHARS) {
      setComposerError(`El comentario no puede superar ${MAX_PSI_MENSAJE_CHARS} caracteres.`);
      return;
    }
    setSending(true);
    setComposerError("");
    try {
      const form = new FormData();
      form.set("message", msg);
      for (const f of files()) form.append("files", f);
      const created = await apiPost<TicketMensaje>(`/psi/tickets/${params.id}/mensaje`, form);
      setMessage("");
      setFiles([]);
      setMensajesExtra((prev) => [...prev, created]);
    } catch (e: any) {
      setComposerError(getUserFacingError(e));
    } finally {
      setSending(false);
    }
  };

  const submitClose = async () => {
    const reason = closeReason().trim();
    if (!reason) {
      setCloseError("Indica el motivo por el que cierras tu solicitud.");
      return;
    }
    setClosing(true);
    setCloseError("");
    try {
      await apiPost(`/psi/tickets/${params.id}/cerrar`, { close_reason: reason });
      setCloseOpen(false);
      setCloseReason("");
      refetch();
    } catch (e: any) {
      setCloseError(getUserFacingError(e));
    } finally {
      setClosing(false);
    }
  };

  return (
    <main class="bg-colpsi-bg min-h-screen pb-24">
      <div class="bg-heraldic pt-12 pb-20 px-6">
        <div class="max-w-4xl mx-auto">
          <p class="text-blue-200 text-sm font-bold mb-4">
            <a href="/psi/tickets" class="hover:text-white inline-flex items-center gap-1">← Mis Solicitudes</a>
          </p>
          <h1 class="text-white text-2xl font-bold flex items-center gap-3">
            <span class="text-[11px] text-blue-200 font-black uppercase tracking-widest">Ticket #{params.id}</span>
            <span class={`text-[10px] font-black px-2.5 py-1 rounded-full uppercase tracking-wider ${closed() ? "bg-red-500/20 text-red-200" : "bg-emerald-500/20 text-emerald-200"}`}>
              {closed() ? "Cerrado" : "En curso"}
            </span>
          </h1>
          <p class="text-white text-lg font-bold mt-1">{t()?.title}</p>
        </div>
      </div>

      <div class="max-w-4xl mx-auto px-4 -mt-12 space-y-4">
        <Show when={ticket.loading && !t()}>
          <div class="space-y-3">
            <For each={[1, 2, 3]}>{() => <div class="h-24 bg-white animate-pulse rounded-3xl border border-colpsi-border" />}</For>
          </div>
        </Show>
        <Show when={!ticket.loading && !t()}>
            <div class="bg-white rounded-3xl p-12 text-center shadow-sm border border-colpsi-border">
              <p class="text-5xl mb-4">🔍</p>
              <h3 class="font-black text-gray-700">Solicitud no encontrada</h3>
              <p class="text-sm text-gray-500 mt-1">Puede que no exista o que no te pertenezca.</p>
            </div>
          </Show>

          <Show when={t()}>
            {/* Resumen */}
            <div class="bg-white rounded-3xl p-6 shadow-sm border border-colpsi-border">
              <div class="flex flex-wrap items-center gap-2 text-[11px] font-black uppercase tracking-wider">
                <span class={`px-3 py-1.5 rounded-full ${estadoColor(t()?.estado)}`}>{t()?.estado?.name}</span>
                <span class="px-3 py-1.5 rounded-full bg-gray-100 text-gray-600">{t()?.motivo?.name}</span>
                <span class="px-3 py-1.5 rounded-full bg-gray-100 text-gray-400">Creada: {formatTicketDate(t()?.created_at)}</span>
              </div>
              <p class="text-sm text-gray-600 leading-relaxed mt-4 bg-colpsi-surface rounded-2xl px-4 py-3 whitespace-pre-wrap">
                {t()?.description}
              </p>
              <Show when={closed() && t()?.close_reason}>
                <div class="mt-4 bg-red-50 border border-red-100 rounded-2xl px-4 py-3 text-sm">
                  <span class="font-black text-red-700 text-xs uppercase tracking-widest">Motivo de cierre: </span>
                  <span class="text-red-700 font-semibold">{t()?.close_reason}</span>
                </div>
              </Show>
            </div>

            {/* Conversación */}
            <div class="bg-white rounded-3xl p-6 shadow-sm border border-colpsi-border">
              <h3 class="font-black text-gray-700 text-sm uppercase tracking-widest mb-4">Conversación</h3>
              <TicketThread
                mensajes={[...(t()?.mensajes ?? []), ...mensajesExtra()]}
                adminDisplayName="El Colegio"
                emptyText="Aún no hay respuestas del colegio. Tu descripción inicial ya fue registrada."
              />

              <Show when={closed()} fallback={
                <div class="mt-6 pt-5 border-t border-colpsi-border">
                  <label class="block text-xs font-black text-gray-500 uppercase tracking-widest mb-2">
                    Nuevo comentario
                    <span class="float-right normal-case font-bold text-gray-300">{message().length}/{MAX_PSI_MENSAJE_CHARS}</span>
                  </label>
                  <textarea
                    value={message()}
                    maxLength={MAX_PSI_MENSAJE_CHARS}
                    rows={3}
                    onInput={(e) => setMessage(e.currentTarget.value)}
                    placeholder="Escribe un comentario (máximo 3 seguidos)..."
                    class="w-full px-4 py-3.5 rounded-2xl border-2 border-colpsi-border bg-colpsi-surface outline-none focus:border-colpsi-yellow text-sm font-semibold text-gray-800 transition-all resize-none"
                  />
                  <div class="flex flex-col sm:flex-row gap-3 mt-3">
                    <label class="flex-1 flex items-center justify-center gap-2 px-4 py-3 rounded-2xl border-2 border-dashed border-gray-200 bg-colpsi-surface/60 cursor-pointer hover:border-colpsi-blue transition-all text-sm font-bold text-gray-500">
                      📎 Adjuntar
                      <input type="file" multiple class="hidden"
                        onChange={(e) => { const l = e.currentTarget.files; if (l) setFiles(Array.from(l)); }} />
                    </label>
                    <button
                      onClick={submitMensaje}
                      disabled={sending() || !message().trim()}
                      class="sm:w-48 bg-colpsi-blue hover:bg-colpsi-blue-light text-white font-black py-3 rounded-2xl shadow-lg hover:scale-[1.02] active:scale-95 transition-all disabled:opacity-40 disabled:cursor-not-allowed text-sm uppercase tracking-widest"
                    >
                      {sending() ? "Enviando..." : "Enviar"}
                    </button>
                  </div>
                  <Show when={files().length > 0}>
                    <div class="mt-3 flex flex-wrap gap-2">
                      <For each={files()}>{(f) => (
                        <span class="bg-blue-50 border border-blue-100 text-blue-800 text-[11px] font-bold px-3 py-1.5 rounded-xl">📎 {f.name}</span>
                      )}</For>
                    </div>
                  </Show>
                  <Show when={files().length > 0}>
                    <button
                      onClick={() => setFiles([])}
                      class="mt-2 text-[11px] font-black text-red-500 hover:text-red-700 uppercase tracking-widest"
                    >
                      Quitar anexos
                    </button>
                  </Show>
                  <Show when={composerError()}>
                    <p class="mt-2 text-xs font-bold text-red-600">{composerError()}</p>
                  </Show>
                </div>
              }>
                <div class="mt-6 pt-5 border-t border-colpsi-border text-center text-sm font-bold text-gray-400">
                  Solicitud cerrada — no admite más comentarios.
                </div>
              </Show>
            </div>

            {/* Cierre */}
            <Show when={!closed()}>
              <div class="bg-white rounded-3xl p-6 shadow-sm border border-colpsi-border flex items-center justify-between">
                <div>
                  <h3 class="font-black text-gray-700 text-sm uppercase tracking-widest">Cerrar solicitud</h3>
                  <p class="text-xs text-gray-400 mt-1">Si consideras tu trámite resuelto, puedes cerrarlo tú mismo.</p>
                </div>
                <button
                  onClick={() => { setCloseOpen(true); setCloseReason(""); setCloseError(""); }}
                  class="px-5 py-3 rounded-2xl border-2 border-red-200 text-red-600 hover:bg-red-600 hover:text-white hover:border-red-600 font-black text-xs uppercase tracking-widest transition-all"
                >
                  Cerrar solicitud
                </button>
              </div>
            </Show>

            {/* Historial de estados */}
            <Show when={(t()?.status_logs ?? []).length > 0}>
              <div class="bg-white rounded-3xl p-6 shadow-sm border border-colpsi-border">
                <h3 class="font-black text-gray-700 text-sm uppercase tracking-widest mb-4">Historial de la solicitud</h3>
                <ol class="relative border-l-2 border-gray-200 ml-2 space-y-5">
                  <For each={t()?.status_logs ?? []}>
                    {(log: TicketStatusLog) => (
                      <li class="ml-5">
                        <span class={`absolute -left-[9px] mt-1 w-4 h-4 rounded-full border-4 border-white ${log.new_state?.is_closed ? "bg-red-500" : "bg-colpsi-blue"}`} />
                        <p class="text-sm font-black text-gray-700">
                          Estado: {log.new_state?.name ?? `#${log.new_state_id}`}
                        </p>
                        <Show when={log.reason}>
                          <p class="text-xs text-gray-500 mt-0.5">{log.reason}</p>
                        </Show>
                        <p class="text-[10px] text-gray-400 font-bold uppercase tracking-wider mt-0.5">
                          {formatTicketDateTime(log.created_at)}
                          <Show when={log.changed_by_type === "psi"}>
                            {" · por ti"}
                          </Show>
                          <Show when={log.changed_by_type === "admin"}>
                            {" · por el colegio"}
                          </Show>
                        </p>
                      </li>
                    )}
                  </For>
                </ol>
              </div>
            </Show>
          </Show>
        </div>

      {/* Modal de cierre */}
      <Show when={closeOpen()}>
        <div
          class="fixed inset-0 z-[100] flex items-center justify-center p-6 bg-blue-900/40 backdrop-blur-md"
          onClick={(e) => { if (e.target === e.currentTarget) setCloseOpen(false); }}
        >
          <div class="bg-white rounded-3xl shadow-2xl p-8 w-full max-w-md border border-colpsi-border">
            <h3 class="text-lg font-black text-gray-900 mb-1">Cerrar solicitud</h3>
            <p class="text-sm text-gray-500 mb-4">Indica el motivo por el que cierras este ticket.</p>
            <textarea
              value={closeReason()}
              maxLength={MAX_CLOSE_REASON_CHARS}
              rows={3}
              onInput={(e) => setCloseReason(e.currentTarget.value)}
              placeholder="Ej: Ya resolví mi trámite..."
              class="w-full px-4 py-3 rounded-2xl border-2 border-colpsi-border bg-colpsi-surface outline-none focus:border-red-400 text-sm font-semibold text-gray-800 transition-all resize-none"
            />
            <p class="text-right text-[11px] text-gray-400 font-bold mt-1">{closeReason().length}/{MAX_CLOSE_REASON_CHARS}</p>
            <Show when={closeError()}>
              <p class="text-xs font-bold text-red-600 mt-2">{closeError()}</p>
            </Show>
            <div class="flex gap-3 mt-5">
              <button
                onClick={() => setCloseOpen(false)}
                disabled={closing()}
                class="flex-1 px-5 py-3 rounded-2xl border-2 border-colpsi-border font-black text-gray-400 hover:bg-colpsi-surface transition-all text-xs uppercase tracking-widest"
              >
                Cancelar
              </button>
              <button
                onClick={submitClose}
                disabled={closing() || !closeReason().trim()}
                class="flex-1 px-5 py-3 rounded-2xl bg-red-600 text-white font-black hover:bg-red-700 active:scale-95 transition-all text-xs uppercase tracking-widest shadow-lg shadow-red-200 disabled:opacity-50"
              >
                {closing() ? "Cerrando..." : "Confirmar cierre"}
              </button>
            </div>
          </div>
        </div>
      </Show>
    </main>
  );
}