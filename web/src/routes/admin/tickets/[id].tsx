// web/src/routes/admin/tickets/[id].tsx
// Detalle administrativo de un ticket: conversación, cambio de estado,
// cierre e historial. El menú de estados proviene de la config del motivo.
import { createResource, createMemo, createSignal, For, Show } from "solid-js";
import { useParams } from "@solidjs/router";
import { apiGet, apiPatch, apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import TicketThread from "~/components/tickets/TicketThread";
import type { Ticket, TicketMensaje, TicketMotivo, TicketStatusLog } from "~/types/tickets";
import {
  MAX_ADMIN_MENSAJE_CHARS,
  MAX_CLOSE_REASON_CHARS,
  estadoColor,
  formatTicketDateTime,
  formatTicketDate,
} from "~/types/tickets";
import { Icon } from "~/components/admin/ui/icons";

const IC = "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text";
const labelClass = "block text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1 mb-1";

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

  // Mensajes enviados en la sesión actual: se añaden al hilo sin recargar la
  // página (la API devuelve el mensaje creado; no se vuelve a consultar el ticket).
  const [mensajesExtra, setMensajesExtra] = createSignal<TicketMensaje[]>([]);

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
      const created = await apiPost<TicketMensaje>(`/admin/tickets/${params.id}/mensaje`, form);
      setMessage("");
      setFiles([]);
      setMensajesExtra((prev) => [...prev, created]);
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
    <main class="space-y-4 pb-12">
      <a href="/admin/tickets" class="inline-flex items-center gap-1.5 text-xs font-semibold text-colpsi-muted uppercase tracking-wide hover:text-colpsi-blue transition-all">
        <Icon name="chevronRight" class="w-3.5 h-3.5 rotate-180" />
        Cola de tickets
      </a>

      <Show when={ticket.loading && !t()}>
        <div class="space-y-4">
          <div class="h-32 bg-white animate-pulse rounded-lg border border-colpsi-border" />
          <div class="h-64 bg-white animate-pulse rounded-lg border border-colpsi-border" />
        </div>
      </Show>
      <Show when={!ticket.loading && !t()}>
          <div class="bg-white rounded-lg p-12 text-center border border-colpsi-border">
            <span class="inline-flex h-12 w-12 items-center justify-center rounded-md bg-colpsi-bg text-colpsi-muted mb-4">
              <Icon name="search" class="w-6 h-6" />
            </span>
            <h3 class="text-base font-semibold text-colpsi-text">Ticket no encontrado</h3>
          </div>
        </Show>

        <Show when={t()}>
          {/* Encabezado del ticket */}
          <div class="bg-white rounded-lg p-5 border border-colpsi-border">
            <div class="flex flex-wrap items-start justify-between gap-4">
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="text-xs text-colpsi-muted font-medium">#{t()?.id}</span>
                  <span class={`inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-semibold uppercase tracking-wide ${estadoColor(t()?.estado)}`}>
                    {t()?.estado?.name ?? "Sin estado"}
                  </span>
                  <Show when={closed()}>
                    <span class="inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-semibold uppercase tracking-wide bg-red-100 text-colpsi-red">Cerrado</span>
                  </Show>
                </div>
                <h1 class="text-xl font-semibold text-colpsi-text mt-2">{t()?.title}</h1>
                <div class="flex flex-wrap items-center gap-x-3 gap-y-1 mt-2 text-[11px] font-medium text-colpsi-muted">
                  <span class="inline-flex items-center gap-1.5"><Icon name="user" class="w-3.5 h-3.5" /> {[t()?.psi_first_name, t()?.psi_last_name].filter(Boolean).join(" ") || "Psicólogo/a"}</span>
                  <span class="w-1 h-1 bg-colpsi-border rounded-full" />
                  <span class="inline-flex items-center gap-1.5"><Icon name="tag" class="w-3.5 h-3.5" /> {t()?.motivo?.name}</span>
                  <span class="w-1 h-1 bg-colpsi-border rounded-full" />
                  <span>{formatTicketDate(t()?.created_at)}</span>
                </div>
              </div>

              {/* Acciones: cambio de estado */}
              <div class="w-full md:w-72 bg-colpsi-bg rounded-lg p-4 space-y-3 border border-colpsi-border">
                <p class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide">Cambiar estado</p>
                <select
                  value={estadoId()}
                  onChange={(e) => setEstadoId(e.currentTarget.value)}
                  class={IC}
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
                  class={IC}
                />
                <Show when={estadoError()}>
                  <p class="text-xs font-medium text-colpsi-red">{estadoError()}</p>
                </Show>
                <button
                  onClick={submitEstado}
                  disabled={estadoSaving() || !estadoId()}
                  class="w-full h-9 rounded-md bg-colpsi-blue hover:bg-colpsi-blue-light text-white font-semibold transition-all disabled:opacity-40 text-sm"
                >
                  {estadoSaving() ? "Guardando..." : "Actualizar estado"}
                </button>

                <Show when={!closed()}>
                  <div class="border-t border-colpsi-border pt-3 space-y-2">
                    <p class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide">Cerrar solicitud</p>
                    <input
                      value={closeReason()}
                      onInput={(e) => setCloseReason(e.currentTarget.value)}
                      maxLength={MAX_CLOSE_REASON_CHARS}
                      placeholder="Motivo de cierre (obligatorio)"
                      class={IC}
                    />
                    <Show when={closeError()}>
                      <p class="text-xs font-medium text-colpsi-red">{closeError()}</p>
                    </Show>
                    <button
                      onClick={submitClose}
                      disabled={closing() || !closeReason().trim()}
                      class="w-full h-9 rounded-md border border-red-200 text-colpsi-red hover:bg-red-600 hover:text-white hover:border-red-600 font-semibold transition-all disabled:opacity-40 text-sm bg-white"
                    >
                      {closing() ? "Cerrando..." : "Cerrar solicitud"}
                    </button>
                  </div>
                </Show>
              </div>
            </div>

            <Show when={t()?.description}>
              <p class="text-sm text-colpsi-text leading-relaxed mt-4 bg-colpsi-bg rounded-md px-4 py-3 whitespace-pre-wrap border border-colpsi-border">
                {t()?.description}
              </p>
            </Show>
            <Show when={closed() && t()?.close_reason}>
              <div class="mt-4 bg-red-50 border border-red-200 rounded-md px-4 py-3 text-sm">
                <span class="font-semibold text-colpsi-red text-xs uppercase tracking-wide">Motivo de cierre: </span>
                <span class="text-colpsi-red font-medium">{t()?.close_reason}</span>
              </div>
            </Show>
          </div>

          <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
            {/* Conversación */}
            <div class="lg:col-span-2 bg-white rounded-lg p-5 border border-colpsi-border">
              <h3 class="text-sm font-semibold text-colpsi-text border-b border-colpsi-border pb-3 mb-4">Conversación</h3>
              <TicketThread
                mensajes={[...(t()?.mensajes ?? []), ...mensajesExtra()]}
                adminDisplayName="El Colegio"
                emptyText="Aún no hay mensajes en esta conversación."
              />

              <Show when={closed()} fallback={
                <div class="mt-5 pt-4 border-t border-colpsi-border">
                  <div class="flex items-center justify-between">
                    <label class={labelClass}>Responder al psicólogo</label>
                    <span class="text-xs font-medium text-colpsi-muted">{message().length}/{MAX_ADMIN_MENSAJE_CHARS}</span>
                  </div>
                  <textarea
                    value={message()}
                    maxLength={MAX_ADMIN_MENSAJE_CHARS}
                    rows={3}
                    onInput={(e) => setMessage(e.currentTarget.value)}
                    placeholder="Escribe tu respuesta..."
                    class={`${IC} resize-none py-2.5 leading-relaxed`}
                  />
                  <div class="flex flex-col sm:flex-row gap-2 mt-3">
                    <label class="flex-1 flex items-center justify-center gap-2 h-10 rounded-md border border-dashed border-colpsi-border bg-white cursor-pointer hover:border-colpsi-blue transition-all text-sm font-medium text-colpsi-muted">
                      <Icon name="fileText" class="w-4 h-4" />
                      Adjuntar
                      <input type="file" multiple class="hidden"
                        onChange={(e) => { const l = e.currentTarget.files; if (l) setFiles(Array.from(l)); }} />
                    </label>
                    <button
                      onClick={submitMensaje}
                      disabled={sending() || !message().trim()}
                      class="sm:w-52 h-10 rounded-md bg-colpsi-blue hover:bg-colpsi-blue-light text-white font-semibold transition-all disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                    >
                      {sending() ? "Enviando..." : "Responder"}
                    </button>
                  </div>
                  <Show when={files().length > 0}>
                    <div class="mt-3 flex flex-wrap gap-2 items-center">
                      <For each={files()}>{(f) => (
                        <span class="inline-flex items-center gap-1.5 bg-blue-50 border border-blue-200 text-colpsi-blue text-[11px] font-medium px-2.5 py-1 rounded-md"><Icon name="fileText" class="w-3 h-3" /> {f.name}</span>
                      )}</For>
                      <button onClick={() => setFiles([])} class="text-[11px] font-semibold text-colpsi-red hover:underline">
                        Quitar anexos
                      </button>
                    </div>
                  </Show>
                  <Show when={composerError()}>
                    <p class="mt-2 text-xs font-medium text-colpsi-red">{composerError()}</p>
                  </Show>
                </div>
              }>
                <div class="mt-5 pt-4 border-t border-colpsi-border text-center text-sm font-medium text-colpsi-muted">
                  Solicitud cerrada — no admite más respuestas.
                </div>
              </Show>
            </div>

            {/* Historial de estados */}
            <div class="bg-white rounded-lg p-5 border border-colpsi-border h-fit">
              <h3 class="text-sm font-semibold text-colpsi-text border-b border-colpsi-border pb-3 mb-4">Historial de la solicitud</h3>
              <Show when={(t()?.status_logs ?? []).length > 0} fallback={
                <p class="text-sm text-colpsi-muted">Sin cambios registrados.</p>
              }>
                <ol class="relative border-l-2 border-colpsi-border ml-1 space-y-5">
                  <For each={t()?.status_logs ?? []}>
                    {(log: TicketStatusLog) => (
                      <li class="ml-4">
                        <span class={`absolute -left-[9px] mt-1 w-4 h-4 rounded-full border-4 border-white shadow-sm ${log.new_state?.is_closed ? "bg-colpsi-red" : "bg-colpsi-blue"}`} />
                        <p class="text-sm font-semibold text-colpsi-text">
                          {log.new_state?.name ?? `#${log.new_state_id}`}
                        </p>
                        <Show when={log.reason}>
                          <p class="text-xs text-colpsi-muted mt-0.5">{log.reason}</p>
                        </Show>
                        <p class="text-[10px] text-colpsi-muted font-medium mt-0.5">
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
    </main>
  );
}