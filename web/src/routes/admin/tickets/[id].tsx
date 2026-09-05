// web/src/routes/admin/tickets/[id].tsx
// Detalle administrativo de un ticket: conversación, cambio de estado,
// cierre e historial. El menú de estados proviene de la config del motivo.
import { createResource, createMemo, createSignal, For, Show, Suspense } from "solid-js";
import { useParams } from "@solidjs/router";
import { apiGet, apiPatch, apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import TicketThread from "~/components/tickets/TicketThread";
import type { Ticket, TicketStatusLog, TicketMotivo } from "~/types/tickets";
import {
  MAX_ADMIN_MENSAJE_CHARS,
  MAX_CLOSE_REASON_CHARS,
  estadoColor,
  formatTicketDateTime,
  formatTicketDate,
} from "~/types/tickets";

export default function AdminTicketDetalle() {
  const params = useParams<{ id: string }>();

  const [ticket, { refetch }] = createResource<Ticket | null, string>(
    () => params.id,
    async (_k) => {
      try {
        return await apiGet<Ticket>(`/admin/tickets/${params.id}`);
      } catch {
        return null;
      }
    }
  );

  const [motivosConfig] = createResource(() => apiGet<{ data: TicketMotivo[] }>("/admin/tickets/motivos"), {
    initialValue: { data: [] },
  });

  const t = () => ticket();
  const closed = () => !!t()?.is_closed || !!t()?.closed_at;

  const motivoEstados = createMemo(() => {
    const m = (motivosConfig()?.data ?? []).find((mo) => mo.id === t()?.motivo_id);
    return m?.estados ?? [];
  });

  // ── Composer ────────────────────────────────────────────────────────────
  const [message, setMessage] = createSignal("");
  const [files, setFiles] = createSignal<File[]>([]);
  const [sending, setSending] = createSignal(false);
  const [composerError, setComposerError] = createSignal("");

  // ── Cambio de estado ────────────────────────────────────────────────────
  const [estadoId, setEstadoId] = createSignal("");
  const [estadoReason, setEstadoReason] = createSignal("");
  const [estadoSaving, setEstadoSaving] = createSignal(false);
  const [estadoError, setEstadoError] = createSignal("");

  // ── Cierre ──────────────────────────────────────────────────────────────
  const [closeReason, setCloseReason] = createSignal("");
  const [closing, setClosing] = createSignal(false);
  const [closeError, setCloseError] = createSignal("");

  const submitMensaje = async () => {
    const msg = message().trim();
    if (!msg) {
      setComposerError("Escribe una respuesta antes de enviar.");
      return;
    }
    if (msg.length > MAX_ADMIN_MENSAJE_CHARS) {
      setComposerError(`La respuesta no puede superar ${MAX_ADMIN_MENSAJE_CHARS} caracteres.`);
      return;
    }
    setSending(true);
    setComposerError("");
    try {
      const form = new FormData();
      form.set("message", msg);
      for (const f of files()) form.append("files", f);
      await apiPost(`/admin/tickets/${params.id}/mensaje`, form);
      setMessage("");
      setFiles([]);
      refetch();
    } catch (e: any) {
      setComposerError(getUserFacingError(e));
    } finally {
      setSending(false);
    }
  };

  const submitEstado = async () => {
    if (!estadoId()) {
      setEstadoError("Selecciona el nuevo estado.");
      return;
    }
    setEstadoSaving(true);
    setEstadoError("");
    try {
      const body: Record<string, unknown> = { estado_id: Number(estadoId()) };
      if (estadoReason().trim()) body.reason = estadoReason().trim();
      await apiPatch(`/admin/tickets/${params.id}/estado`, body);
      setEstadoReason("");
      refetch();
    } catch (e: any) {
      setEstadoError(getUserFacingError(e));
    } finally {
      setEstadoSaving(false);
    }
  };

  const submitClose = async () => {
    const reason = closeReason().trim();
    if (!reason) {
      setCloseError("Indica el motivo por el que se cierra la solicitud.");
      return;
    }
    setClosing(true);
    setCloseError("");
    try {
      await apiPost(`/admin/tickets/${params.id}/cerrar`, { close_reason: reason });
      setCloseReason("");
      refetch();
    } catch (e: any) {
      setCloseError(getUserFacingError(e));
    } finally {
      setClosing(false);
    }
  };

  return (
    <main class="space-y-5">
      <a href="/admin/tickets" class="inline-flex items-center gap-1 text-xs font-black text-gray-400 uppercase tracking-widest hover:text-[#1e3a8a] transition-all">
        ← Cola de tickets
      </a>

      <Suspense fallback={
        <div class="space-y-4">
          <div class="h-32 bg-white animate-pulse rounded-3xl border border-gray-100" />
          <div class="h-64 bg-white animate-pulse rounded-3xl border border-gray-100" />
        </div>
      }>
        <Show when={!ticket.loading && !t()}>
          <div class="bg-white rounded-3xl p-12 text-center shadow-sm border border-gray-100">
            <p class="text-5xl mb-4">🔍</p>
            <h3 class="font-black text-gray-700">Ticket no encontrado</h3>
          </div>
        </Show>

        <Show when={t()}>
          {/* Encabezado del ticket */}
          <div class="bg-white rounded-3xl p-6 shadow-sm border border-gray-100">
            <div class="flex flex-wrap items-start justify-between gap-4">
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="text-[10px] text-gray-400 font-black">#{t()?.id}</span>
                  <span class={`text-[10px] font-black px-2.5 py-1 rounded-full uppercase tracking-wider ${estadoColor(t()?.estado)}`}>
                    {t()?.estado?.name ?? "Sin estado"}
                  </span>
                  <Show when={closed()}>
                    <span class="text-[10px] font-black px-2.5 py-1 rounded-full bg-red-100 text-red-700 uppercase tracking-wider">Cerrado</span>
                  </Show>
                </div>
                <h1 class="text-xl font-black text-gray-800 mt-2">{t()?.title}</h1>
                <div class="flex flex-wrap items-center gap-x-3 gap-y-1 mt-2 text-[11px] text-gray-400 font-bold uppercase tracking-wider">
                  <span>👤 {[t()?.psi_first_name, t()?.psi_last_name].filter(Boolean).join(" ") || "Psicólogo/a"}</span>
                  <span class="w-1 h-1 bg-gray-200 rounded-full" />
                  <span>🗂️ {t()?.motivo?.name}</span>
                  <span class="w-1 h-1 bg-gray-200 rounded-full" />
                  <span>{formatTicketDate(t()?.created_at)}</span>
                </div>
              </div>

              {/* Acciones: cambio de estado */}
              <div class="w-full md:w-72 bg-gray-50 rounded-2xl p-4 space-y-3">
                <p class="text-[10px] font-black text-gray-500 uppercase tracking-widest">Cambiar estado</p>
                <select
                  value={estadoId()}
                  onChange={(e) => setEstadoId(e.currentTarget.value)}
                  class="w-full px-3 py-2.5 rounded-xl border-2 border-gray-100 bg-white outline-none focus:border-[#1e3a8a] text-sm font-semibold text-gray-700 transition-all"
                >
                  <option value="">Selecciona...</option>
                  <For each={motivoEstados()}>
                    {(e) => <option value={e.id} disabled={e.id === t()?.estado_id}>{e.name}</option>}
                  </For>
                </select>
                <input
                  value={estadoReason()}
                  onInput={(e) => setEstadoReason(e.currentTarget.value)}
                  placeholder="Comentario (opcional)"
                  class="w-full px-3 py-2.5 rounded-xl border-2 border-gray-100 bg-white outline-none focus:border-[#1e3a8a] text-sm font-semibold text-gray-700 transition-all"
                />
                <Show when={estadoError()}>
                  <p class="text-xs font-bold text-red-600">{estadoError()}</p>
                </Show>
                <button
                  onClick={submitEstado}
                  disabled={estadoSaving() || !estadoId()}
                  class="w-full bg-[#1e3a8a] hover:bg-[#1e40af] text-white font-black py-2.5 rounded-xl text-xs uppercase tracking-widest transition-all active:scale-95 disabled:opacity-40"
                >
                  {estadoSaving() ? "Guardando..." : "Actualizar estado"}
                </button>

                <Show when={!closed()}>
                  <div class="border-t border-gray-200 pt-3 space-y-2">
                    <p class="text-[10px] font-black text-gray-500 uppercase tracking-widest">Cerrar solicitud</p>
                    <input
                      value={closeReason()}
                      onInput={(e) => setCloseReason(e.currentTarget.value)}
                      maxLength={MAX_CLOSE_REASON_CHARS}
                      placeholder="Motivo de cierre (obligatorio)"
                      class="w-full px-3 py-2.5 rounded-xl border-2 border-gray-100 bg-white outline-none focus:border-red-400 text-sm font-semibold text-gray-700 transition-all"
                    />
                    <Show when={closeError()}>
                      <p class="text-xs font-bold text-red-600">{closeError()}</p>
                    </Show>
                    <button
                      onClick={submitClose}
                      disabled={closing() || !closeReason().trim()}
                      class="w-full border-2 border-red-200 text-red-600 hover:bg-red-600 hover:text-white hover:border-red-600 font-black py-2.5 rounded-xl text-xs uppercase tracking-widest transition-all active:scale-95 disabled:opacity-40"
                    >
                      {closing() ? "Cerrando..." : "Cerrar solicitud"}
                    </button>
                  </div>
                </Show>
              </div>
            </div>

            <Show when={t()?.description}>
              <p class="text-sm text-gray-600 leading-relaxed mt-4 bg-gray-50 rounded-2xl px-4 py-3 whitespace-pre-wrap">
                {t()?.description}
              </p>
            </Show>
            <Show when={closed() && t()?.close_reason}>
              <div class="mt-4 bg-red-50 border border-red-100 rounded-2xl px-4 py-3 text-sm">
                <span class="font-black text-red-700 text-xs uppercase tracking-widest">Motivo de cierre: </span>
                <span class="text-red-700 font-semibold">{t()?.close_reason}</span>
              </div>
            </Show>
          </div>

          <div class="grid grid-cols-1 lg:grid-cols-3 gap-5">
            {/* Conversación */}
            <div class="lg:col-span-2 bg-white rounded-3xl p-6 shadow-sm border border-gray-100">
              <h3 class="font-black text-gray-700 text-sm uppercase tracking-widest mb-4">Conversación</h3>
              <TicketThread
                mensajes={t()?.mensajes ?? []}
                adminDisplayName="El Colegio"
                emptyText="Aún no hay mensajes en esta conversación."
              />

              <Show when={closed()} fallback={
                <div class="mt-6 pt-5 border-t border-gray-100">
                  <label class="block text-xs font-black text-gray-500 uppercase tracking-widest mb-2">
                    Responder al psicólogo
                    <span class="float-right normal-case font-bold text-gray-300">{message().length}/{MAX_ADMIN_MENSAJE_CHARS}</span>
                  </label>
                  <textarea
                    value={message()}
                    maxLength={MAX_ADMIN_MENSAJE_CHARS}
                    rows={3}
                    onInput={(e) => setMessage(e.currentTarget.value)}
                    placeholder="Escribe tu respuesta..."
                    class="w-full px-4 py-3.5 rounded-2xl border-2 border-gray-100 bg-gray-50 outline-none focus:border-[#1e3a8a] text-sm font-semibold text-gray-800 transition-all resize-none"
                  />
                  <div class="flex flex-col sm:flex-row gap-3 mt-3">
                    <label class="flex-1 flex items-center justify-center gap-2 px-4 py-3 rounded-2xl border-2 border-dashed border-gray-200 bg-gray-50/60 cursor-pointer hover:border-[#1e3a8a] transition-all text-sm font-bold text-gray-500">
                      📎 Adjuntar
                      <input type="file" multiple class="hidden"
                        onChange={(e) => { const l = e.currentTarget.files; if (l) setFiles(Array.from(l)); }} />
                    </label>
                    <button
                      onClick={submitMensaje}
                      disabled={sending() || !message().trim()}
                      class="sm:w-52 bg-[#1e3a8a] hover:bg-[#1e40af] text-white font-black py-3 rounded-2xl shadow-lg hover:scale-[1.02] active:scale-95 transition-all disabled:opacity-40 disabled:cursor-not-allowed text-sm uppercase tracking-widest"
                    >
                      {sending() ? "Enviando..." : "Responder"}
                    </button>
                  </div>
                  <Show when={files().length > 0}>
                    <div class="mt-3 flex flex-wrap gap-2">
                      <For each={files()}>{(f) => (
                        <span class="bg-blue-50 border border-blue-100 text-blue-800 text-[11px] font-bold px-3 py-1.5 rounded-xl">📎 {f.name}</span>
                      )}</For>
                      <button onClick={() => setFiles([])} class="text-[11px] font-black text-red-500 hover:text-red-700 uppercase tracking-widest">
                        Quitar anexos
                      </button>
                    </div>
                  </Show>
                  <Show when={composerError()}>
                    <p class="mt-2 text-xs font-bold text-red-600">{composerError()}</p>
                  </Show>
                </div>
              }>
                <div class="mt-6 pt-5 border-t border-gray-100 text-center text-sm font-bold text-gray-400">
                  Solicitud cerrada — no admite más respuestas.
                </div>
              </Show>
            </div>

            {/* Historial de estados */}
            <div class="bg-white rounded-3xl p-6 shadow-sm border border-gray-100 h-fit">
              <h3 class="font-black text-gray-700 text-sm uppercase tracking-widest mb-4">Historial de la solicitud</h3>
              <Show when={(t()?.status_logs ?? []).length > 0} fallback={
                <p class="text-sm text-gray-400 font-bold">Sin cambios registrados.</p>
              }>
                <ol class="relative border-l-2 border-gray-200 ml-1 space-y-5">
                  <For each={t()?.status_logs ?? []}>
                    {(log: TicketStatusLog) => (
                      <li class="ml-4">
                        <span class={`absolute -left-[9px] mt-1 w-4 h-4 rounded-full border-4 border-white ${log.new_state?.is_closed ? "bg-red-500" : "bg-[#1e3a8a]"}`} />
                        <p class="text-sm font-black text-gray-700">
                          {log.new_state?.name ?? `#${log.new_state_id}`}
                        </p>
                        <Show when={log.reason}>
                          <p class="text-xs text-gray-500 mt-0.5">{log.reason}</p>
                        </Show>
                        <p class="text-[10px] text-gray-400 font-bold uppercase tracking-wider mt-0.5">
                          {formatTicketDateTime(log.created_at)}
                          <Show when={log.changed_by_type === "psi"}> · psicólogo</Show>
                          <Show when={log.changed_by_type === "admin"}> · admin</Show>
                          <Show when={log.changed_by_type === "system"}> · sistema</Show>
                        </p>
                      </li>
                    )}
                  </For>
                </ol>
              </Show>
            </div>
          </div>
        </Show>
      </Suspense>
    </main>
  );
}