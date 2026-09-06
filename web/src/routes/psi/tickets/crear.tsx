// web/src/routes/psi/tickets/crear.tsx
// Nueva solicitud de ticket: elegir motivo, título, descripción y anexos
// opcionales. Se envía como multipart/form-data al backend Go.
import { createResource, createMemo, createSignal, For, Show, Suspense } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { apiGet, apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import type { MotivoConfigQueryResponse, TicketMotivo, Ticket } from "~/types/tickets";
import { MAX_TICKET_TITLE_CHARS, MAX_TICKET_DESC_CHARS } from "~/types/tickets";

export default function PsiCrearTicket() {
  const navigate = useNavigate();

  const [config] = createResource(() => apiGet<MotivoConfigQueryResponse>("/psi/tickets/config"), {
    initialValue: { data: [] },
  });

  const [motivoId, setMotivoId] = createSignal("");
  const [title, setTitle] = createSignal("");
  const [description, setDescription] = createSignal("");
  const [files, setFiles] = createSignal<File[]>([]);
  const [sending, setSending] = createSignal(false);
  const [error, setError] = createSignal("");

  // Interruptor global de recepción de tickets: si está desactivado, se bloquea
  // la creación y se muestra el motivo público.
  const [reception] = createResource<{ enabled: boolean; message?: string }>(async () => {
    try {
      const res = await apiGet<{ enabled: boolean; message?: string }>("/psi/tickets/status");
      return res ?? { enabled: true };
    } catch {
      return { enabled: true };
    }
  });
  const receptionDisabled = () => reception() ? reception()!.enabled === false : false;

  const motivos = () => (config()?.data ?? []) as TicketMotivo[];

  const canSubmit = createMemo(
    () => !receptionDisabled() && motivoId() !== "" && title().trim().length > 0 && description().trim().length > 0 && !sending()
  );

  const handleFiles = (e: Event) => {
    const list = (e.currentTarget as HTMLInputElement).files;
    if (!list) return;
    setFiles(Array.from(list));
  };

  const submit = async () => {
    if (receptionDisabled()) {
      setError(reception()?.message || "La recepción de solicitudes se encuentra temporalmente desactivada. Intenta más tarde.");
      return;
    }
    if (!canSubmit()) {
      setError("Completa el motivo, título y descripción antes de enviar.");
      return;
    }
    setSending(true);
    setError("");
    try {
      const form = new FormData();
      form.set("motivo_id", motivoId());
      form.set("title", title().trim());
      form.set("description", description().trim());
      for (const f of files()) form.append("files", f);
      const created = await apiPost<Ticket>("/psi/tickets", form);
      navigate(`/psi/tickets/${created.id}`); // replaced a destinos
    } catch (e: any) {
      setError(getUserFacingError(e));
      setSending(false);
    }
  };

  return (
    <main class="bg-colpsi-bg min-h-screen pb-24">
      <div class="bg-heraldic pt-10 pb-16 px-6 mb-4">
        <div class="max-w-2xl mx-auto">
          <p class="text-blue-200 text-sm font-bold mb-3">
            <a href="/psi/tickets" class="hover:text-white inline-flex items-center gap-1">← Mis Solicitudes</a>
          </p>
          <h1 class="text-white text-2xl font-bold">🎫 Nueva Solicitud</h1>
          <p class="text-blue-200 text-sm mt-1">Describe tu trámite. El colegio te atenderá por este canal.</p>
        </div>
      </div>

      <div class="max-w-2xl mx-auto px-4">
        <div class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100 space-y-5">
          <Show when={receptionDisabled()}>
            <div class="bg-amber-50 border-2 border-amber-200 text-amber-800 rounded-2xl p-5 flex items-start gap-4">
              <span class="text-2xl">⏸️</span>
              <div>
                <p class="font-black">Recepción de solicitudes temporalmente desactivada</p>
                <p class="text-sm mt-1">
                  {reception()?.message || "Por favor intenta nuevamente en los próximos días."}
                </p>
              </div>
            </div>
          </Show>
          <Suspense fallback={<div class="h-16 bg-gray-50 animate-pulse rounded-2xl" />}>
            <Show when={motivos().length === 0 && !config.loading}>
              <div class="bg-blue-50 border border-blue-200 text-blue-800 text-sm font-semibold px-4 py-3 rounded-2xl">
                El colegio aún no ha configurado motivos de atención. Intenta más tarde.
              </div>
            </Show>

            {/* Motivo */}
            <div>
              <label class="block text-xs font-black text-gray-500 uppercase tracking-widest mb-2">Motivo *</label>
              <select
                value={motivoId()}
                onChange={(e) => setMotivoId(e.currentTarget.value)}
                class="w-full px-4 py-3.5 rounded-2xl border-2 border-gray-100 bg-gray-50 outline-none focus:border-colpsi-yellow text-sm font-semibold text-gray-800 transition-all"
              >
                <option value="">Selecciona un motivo...</option>
                <For each={motivos()}>
                  {(m) => <option value={m.id}>{m.name}</option>}
                </For>
              </select>
              <p class="text-[11px] text-gray-400 mt-1.5">
                Por este motivo puedes tener hasta {motivos().find((m) => String(m.id) === motivoId())?.tickets_per_psi ?? "—"} solicitudes abiertas a la vez.
              </p>
            </div>
          </Suspense>

          {/* Título */}
          <div>
            <label class="block text-xs font-black text-gray-500 uppercase tracking-widest mb-2">
              Título *
              <span class="float-right normal-case font-bold text-gray-300">{title().length}/{MAX_TICKET_TITLE_CHARS}</span>
            </label>
            <input
              type="text"
              value={title()}
              maxLength={MAX_TICKET_TITLE_CHARS}
              onInput={(e) => setTitle(e.currentTarget.value)}
              placeholder="Ej: Solicitud de constancia de solvencia"
              class="w-full px-4 py-3.5 rounded-2xl border-2 border-gray-100 bg-gray-50 outline-none focus:border-colpsi-yellow text-sm font-semibold text-gray-800 transition-all"
            />
          </div>

          {/* Descripción */}
          <div>
            <label class="block text-xs font-black text-gray-500 uppercase tracking-widest mb-2">
              Descripción *
              <span class="float-right normal-case font-bold text-gray-300">{description().length}/{MAX_TICKET_DESC_CHARS}</span>
            </label>
            <textarea
              value={description()}
              maxLength={MAX_TICKET_DESC_CHARS}
              rows={5}
              onInput={(e) => setDescription(e.currentTarget.value)}
              placeholder="Explica tu solicitud con el detalle que consideres necesario..."
              class="w-full px-4 py-3.5 rounded-2xl border-2 border-gray-100 bg-gray-50 outline-none focus:border-colpsi-yellow text-sm font-semibold text-gray-800 transition-all resize-none"
            />
          </div>

          {/* Anexos opcionales */}
          <div>
            <label class="block text-xs font-black text-gray-500 uppercase tracking-widest mb-2">
              Anexos <span class="normal-case font-bold text-gray-300">(opcional)</span>
            </label>
            <label class="flex flex-col items-center justify-center gap-2 px-6 py-8 rounded-2xl border-2 border-dashed border-gray-200 bg-gray-50/60 cursor-pointer hover:border-colpsi-blue hover:bg-blue-50/40 transition-all">
              <span class="text-3xl">📎</span>
              <span class="text-sm font-bold text-gray-600">Agregar archivos</span>
              <span class="text-[11px] text-gray-400">PDF, imágenes u otros (máx. 4 MB por archivo)</span>
              <input type="file" multiple onChange={handleFiles} class="hidden" />
            </label>
            <Show when={files().length > 0}>
              <ul class="mt-3 space-y-2">
                <For each={files()}>
                  {(f, i) => (
                    <li class="flex items-center justify-between bg-blue-50 border border-blue-100 rounded-xl px-4 py-2.5 text-sm">
                      <span class="font-semibold text-blue-800 truncate">📎 {f.name}</span>
                      <button
                        onClick={() => setFiles(files().filter((_, idx) => idx !== i()))}
                        class="text-red-500 font-black hover:text-red-700 ml-3"
                      >
                        ✕
                      </button>
                    </li>
                  )}
                </For>
              </ul>
            </Show>
          </div>

          <Show when={error()}>
            <div class="bg-red-50 border border-red-200 text-red-700 text-sm font-semibold px-4 py-3 rounded-2xl">{error()}</div>
          </Show>

          <button
            onClick={submit}
            disabled={!canSubmit()}
            class="w-full bg-colpsi-blue hover:bg-colpsi-blue-light text-white font-black py-4 rounded-2xl shadow-lg hover:shadow-xl active:scale-[0.98] transition-all disabled:opacity-40 disabled:cursor-not-allowed text-sm uppercase tracking-widest"
          >
            {sending() ? "Enviando solicitud..." : receptionDisabled() ? "Recepción desactivada" : "Enviar Solicitud"}
          </button>
        </div>
      </div>
    </main>
  );
}