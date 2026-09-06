// web/src/routes/admin/notificaciones/crear/index.tsx
import { createResource, createSignal, For, Show, Suspense } from "solid-js";
import { A, action, useAction, useNavigate } from "@solidjs/router";
import { apiGet, apiPost } from "~/lib/api";
import { PaginatedResponse, PsiAdminListItem } from "~/types/admin";
import {
  CreateNotificationResponse,
  CreateNotificationRequest,
  NotificationTargetType,
  PreviewResponse,
  NotificationFilterDTO,
} from "~/types/notifications";
import { getUserFacingError } from "~/lib/errors";

const createNotification = action(async (payload: { body: CreateNotificationRequest; idem: string }) => {
  "use server";
  return await apiPost<CreateNotificationResponse>("/notifications/admin", payload.body, {
    headers: { "X-Idempotency-Key": payload.idem },
  });
});

const previewNotifications = action(async (body: { target_type: NotificationTargetType; filters?: NotificationFilterDTO; target_user_ids?: string[] }) => {
  "use server";
  return await apiPost<PreviewResponse>("/notifications/admin/preview", body);
});

const TARGET_OPTIONS: { value: NotificationTargetType; label: string; icon: string; desc: string }[] = [
  { value: "global", label: "Global", icon: "🌎", desc: "Todos los agremiados activos" },
  { value: "individual", label: "Individual", icon: "👤", desc: "Psicólogos específicos" },
  { value: "group", label: "Por grupo", icon: "🎯", desc: "Según filtros (zona, género, etc.)" },
];

interface Specialty { id: number; name: string; }

export default function CrearNotificacionPage() {
  const navigate = useNavigate();
  const runCreate = useAction(createNotification);
  const runPreview = useAction(previewNotifications);

  const [title, setTitle] = createSignal("");
  const [message, setMessage] = createSignal("");
  const [targetType, setTargetType] = createSignal<NotificationTargetType>("global");
  const [sendEmail, setSendEmail] = createSignal(false);
  const [scheduledAt, setScheduledAt] = createSignal("");

  // Filtros (grupo)
  const [municipality, setMunicipality] = createSignal("");
  const [state, setState] = createSignal("");
  const [genre, setGenre] = createSignal("");
  const [specialtyId, setSpecialtyId] = createSignal("");
  const [solvent, setSolvent] = createSignal("");

  // Individual
  const [selected, setSelected] = createSignal<Set<string>>(new Set());
  // Detalles (nombre/email) de los seleccionados para mostrarlos como chips
  const [selectedDetails, setSelectedDetails] = createSignal<Record<string, { name: string; email: string }>>({});

  const [preview, setPreview] = createSignal<PreviewResponse | null>(null);
  const [busy, setBusy] = createSignal(false);
  const [error, setError] = createSignal("");
  const [successId, setSuccessId] = createSignal("");

  // Buscador paginado de psicólogos (búsqueda por nombre, CI o FPV)
  const PAGE_SIZE = 20;
  const [search, setSearch] = createSignal("");
  const [debouncedSearch, setDebouncedSearch] = createSignal("");
  const [page, setPage] = createSignal(1);

  let debounceTimer: ReturnType<typeof setTimeout> | undefined;
  const onSearchInput = (e: Event) => {
    const value = e.currentTarget.value;
    setSearch(value);
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      setDebouncedSearch(value.trim());
      setPage(1);
    }, 350);
  };

  const [psiUsers] = createResource(
    () => ({ q: debouncedSearch(), page: page() }),
    ({ q, page }) =>
      apiGet<PaginatedResponse<PsiAdminListItem>>(
        `/admin/psi/list?q=${encodeURIComponent(q)}&page=${page}&limit=${PAGE_SIZE}`,
      ),
  );
  const [specialties] = createResource(
    () => apiGet<Specialty[]>("/specialties")
  );

  const toggleUser = (u: PsiAdminListItem) => {
    const next = new Set(selected());
    const nextDetails = { ...selectedDetails() };
    if (next.has(u.id)) {
      next.delete(u.id);
      delete nextDetails[u.id];
    } else {
      next.add(u.id);
      nextDetails[u.id] = {
        name: `${u.first_name} ${u.last_name}`.trim() || u.email,
        email: u.email,
      };
    }
    setSelected(next);
    setSelectedDetails(nextDetails);
  };

  const removeSelected = (id: string) => {
    const next = new Set(selected());
    next.delete(id);
    const nextDetails = { ...selectedDetails() };
    delete nextDetails[id];
    setSelected(next);
    setSelectedDetails(nextDetails);
  };

  const buildFilters = (): NotificationFilterDTO | undefined => {
    const f: NotificationFilterDTO = {};
    if (municipality()) f.municipality = municipality();
    if (state()) f.state = state();
    if (genre()) f.genre = genre();
    if (specialtyId()) f.specialty_id = Number(specialtyId());
    if (solvent()) f.solvent = solvent() === "true";
    if (Object.keys(f).length === 0) return undefined;
    return f;
  };

  const handlePreview = async () => {
    setBusy(true);
    setError("");
    try {
      const res = await runPreview({
        target_type: targetType(),
        filters: buildFilters(),
        target_user_ids: targetType() === "individual" ? Array.from(selected()) : undefined,
      });
      setPreview(res);
    } catch (e: any) {
      setError(e?.message || "No se pudo previsualizar");
    } finally {
      setBusy(false);
    }
  };

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    if (!title().trim() || !message().trim()) {
      setError("El título y el mensaje son obligatorios");
      return;
    }
    if (targetType() === "individual" && selected().size === 0) {
      setError("Selecciona al menos un destinatario");
      return;
    }
    setBusy(true);
    setError("");
    try {
      const body: CreateNotificationRequest = {
        title: title().trim(),
        message: message().trim(),
        target_type: targetType(),
        send_email: sendEmail(),
        filters: buildFilters(),
        target_user_ids: targetType() === "individual" ? Array.from(selected()) : undefined,
      };
      if (scheduledAt()) body.scheduled_at = new Date(scheduledAt()).toISOString();

      const idem = crypto.randomUUID();
      const res = await runCreate({ body, idem });
      setSuccessId(res.id);
      setTimeout(() => navigate(`/admin/notificaciones/${res.id}`, { replace: true }), 1200);
    } catch (e: any) {
      setError(getUserFacingError(e));
    } finally {
      setBusy(false);
    }
  };

  const inputCls = "w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none transition-all text-gray-800 text-sm";
  const labelCls = "text-[10px] font-black text-gray-500 uppercase tracking-widest mb-1.5 block";

  return (
    <main class="pb-20 animate-in fade-in duration-500 max-w-3xl">
      <A href="/admin/notificaciones" class="inline-flex items-center gap-1 text-sm font-bold text-blue-600 hover:text-blue-800 mb-4">
        ← Volver a Notificaciones
      </A>

      <h1 class="text-2xl md:text-3xl font-black text-gray-800 uppercase tracking-tight mb-1">Nueva Notificación</h1>
      <p class="text-sm text-gray-500 font-medium mb-8">Configura el destino y el contenido del aviso</p>

      <Show when={error()}>
        <div class="bg-red-50 border border-red-200 text-red-700 text-sm font-semibold px-4 py-3 rounded-2xl mb-4">{error()}</div>
      </Show>

      <Show when={successId()}>
        <div class="bg-green-50 border border-green-200 text-green-700 text-sm font-semibold px-4 py-3 rounded-2xl mb-4">
          ✅ Notificación creada. Redirigiendo...
        </div>
      </Show>

      <form onSubmit={handleSubmit} class="space-y-6">
        {/* Contenido */}
        <section class="bg-white rounded-3xl border border-gray-100 shadow-sm p-6 space-y-4">
          <h2 class="text-sm font-black text-blue-800 uppercase tracking-widest">Contenido</h2>
          <div>
            <label class={labelCls}>Título *</label>
            <input value={title()} onInput={(e) => setTitle(e.currentTarget.value)} maxlength={255} class={inputCls} placeholder="Ej: Asamblea general ordinaria" />
          </div>
          <div>
            <label class={labelCls}>Mensaje *</label>
            <textarea value={message()} onInput={(e) => setMessage(e.currentTarget.value)} rows={4} class={inputCls} placeholder="Redacta el comunicado..." />
          </div>
          {/* ⚠️ COMPONENTE DESACTIVADO TEMPORALMENTE (aviso a otras IA: NO ELIMINAR).
              El envío de correos se pidió desactivar (no usar Resend para notificaciones).
              Para reactivar, descomenta el <label> de abajo. Todo el flujo backend
              (send_email, mailSvc.SendEmail) sigue intacto en la API Go. */}
          {/* <label class="flex items-center gap-3 cursor-pointer select-none">
            <input type="checkbox" checked={sendEmail()} onChange={(e) => setSendEmail(e.currentTarget.checked)} class="w-5 h-5 rounded accent-blue-700" />
            <span class="text-sm font-semibold text-gray-700">Enviar también por correo electrónico</span>
          </label> */}
        </section>

        {/* Destino */}
        <section class="bg-white rounded-3xl border border-gray-100 shadow-sm p-6 space-y-4">
          <h2 class="text-sm font-black text-blue-800 uppercase tracking-widest">Destino</h2>
          <div class="grid grid-cols-3 gap-3">
            <For each={TARGET_OPTIONS}>
              {(opt) => (
                <button
                  type="button"
                  onClick={() => setTargetType(opt.value)}
                  class={`rounded-2xl border-2 p-3 text-center transition-all ${
                    targetType() === opt.value
                      ? "border-blue-700 bg-blue-50"
                      : "border-gray-200 hover:border-blue-300 bg-white"
                  }`}
                >
                  <span class="text-2xl block mb-1">{opt.icon}</span>
                  <span class="block text-sm font-black text-gray-800">{opt.label}</span>
                  <span class="block text-[10px] text-gray-400 font-medium mt-0.5">{opt.desc}</span>
                </button>
              )}
            </For>
          </div>

          <Show when={targetType() === "individual"}>
            <div class="space-y-3">
              <input
                type="search"
                value={search()}
                onInput={onSearchInput}
                placeholder="🔎 Buscar por nombre, CI o FPV..."
                class={inputCls}
              />
              <Suspense fallback={<div class="h-24 bg-gray-50 animate-pulse rounded-2xl" />}>
                <div class="bg-gray-50 rounded-2xl p-4 max-h-72 overflow-y-auto space-y-1.5">
                  <Show when={(psiUsers()?.data ?? []).length === 0}>
                    <p class="text-sm text-gray-500 p-3">
                      {debouncedSearch() ? "No hay psicólogos que coincidan con la búsqueda." : "No hay psicólogos registrados."}
                    </p>
                  </Show>
                  <For each={psiUsers()?.data ?? []}>
                    {(u) => (
                      <label class="flex items-center gap-3 bg-white rounded-xl px-3 py-2 cursor-pointer border border-gray-100">
                        <input
                          type="checkbox"
                          checked={selected().has(u.id)}
                          onChange={() => toggleUser(u)}
                          class="w-4 h-4 accent-blue-700 shrink-0"
                        />
                        <span class="text-sm font-semibold text-gray-700 truncate">{u.first_name} {u.last_name}</span>
                        <span class="text-xs text-gray-400 whitespace-nowrap">CI {u.ci} · FPV {u.fpv}</span>
                        <span class="ml-auto text-xs text-gray-400 truncate max-w-[160px]">{u.email}</span>
                      </label>
                    )}
                  </For>
                </div>

                <Show when={(psiUsers()?.total ?? 0) > 0}>
                  <div class="flex items-center justify-between mt-2">
                    <button
                      type="button"
                      disabled={page() <= 1}
                      onClick={() => setPage((p) => Math.max(1, p - 1))}
                      class="px-3 py-1.5 rounded-lg bg-gray-100 hover:bg-gray-200 text-gray-600 font-bold text-xs disabled:opacity-40 transition-colors"
                    >
                      ← Anterior
                    </button>
                    <span class="text-xs text-gray-500 font-semibold">
                      Página {psiUsers()?.page ?? 1} de {psiUsers()?.total_pages ?? 1} · {psiUsers()?.total ?? 0} psicólogos
                    </span>
                    <button
                      type="button"
                      disabled={page() >= (psiUsers()?.total_pages ?? 1)}
                      onClick={() => setPage((p) => p + 1)}
                      class="px-3 py-1.5 rounded-lg bg-gray-100 hover:bg-gray-200 text-gray-600 font-bold text-xs disabled:opacity-40 transition-colors"
                    >
                      Siguiente →
                    </button>
                  </div>
                </Show>
              </Suspense>

              <p class="text-xs text-gray-400 font-semibold">{selected().size} seleccionado(s)</p>

              <Show when={selected().size > 0}>
                <div class="flex flex-wrap gap-1.5">
                  <For each={Object.entries(selectedDetails())}>
                    {([id, d]) => (
                      <span class="inline-flex items-center gap-1.5 bg-blue-50 border border-blue-100 text-blue-700 text-xs font-bold px-2.5 py-1 rounded-full">
                        {d.name}
                        <button
                          type="button"
                          onClick={() => removeSelected(id)}
                          class="text-blue-400 hover:text-blue-700 font-black"
                          aria-label={`Quitar a ${d.name}`}
                        >
                          ✕
                        </button>
                      </span>
                    )}
                  </For>
                </div>
              </Show>
            </div>
          </Show>

          <Show when={targetType() === "group"}>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label class={labelCls}>Municipio</label>
                <input value={municipality()} onInput={(e) => setMunicipality(e.currentTarget.value)} class={inputCls} placeholder="Valencia, Naguanagua..." />
              </div>
              <div>
                <label class={labelCls}>Estado</label>
                <input value={state()} onInput={(e) => setState(e.currentTarget.value)} class={inputCls} placeholder="Carabobo, Miranda..." />
              </div>
              <div>
                <label class={labelCls}>Género</label>
                <select value={genre()} onChange={(e) => setGenre(e.currentTarget.value)} class={inputCls}>
                  <option value="">Todos</option>
                  <option value="F">Femenino</option>
                  <option value="M">Masculino</option>
                  <option value="O">Otro</option>
                </select>
              </div>
              <div>
                <label class={labelCls}>Especialidad</label>
                <select value={specialtyId()} onChange={(e) => setSpecialtyId(e.currentTarget.value)} class={inputCls}>
                  <option value="">Todas</option>
                  <For each={specialties() ?? []}>
                    {(s) => <option value={s.id}>{s.name}</option>}
                  </For>
                </select>
              </div>
              <div>
                <label class={labelCls}>Solvencia</label>
                <select value={solvent()} onChange={(e) => setSolvent(e.currentTarget.value)} class={inputCls}>
                  <option value="">Indistinto</option>
                  <option value="true">Solo solventes</option>
                  <option value="false">Solo insolventes</option>
                </select>
              </div>
            </div>
          </Show>

          <Show when={targetType() !== "global"}>
            <button
              type="button"
              onClick={handlePreview}
              disabled={busy()}
              class="inline-flex items-center gap-2 bg-blue-50 hover:bg-blue-100 text-blue-700 font-black px-4 py-2.5 rounded-xl text-sm transition-colors disabled:opacity-50"
            >
              👁️ Previsualizar destinatarios
            </button>
            <Show when={preview()}>
              <div class="bg-blue-50 border border-blue-100 text-blue-800 text-sm font-semibold px-4 py-3 rounded-2xl">
                {preview()?.total_recipients ?? 0} destinatario(s) potencial(es)
              </div>
            </Show>
          </Show>
        </section>

        {/* Programación */}
        <section class="bg-white rounded-3xl border border-gray-100 shadow-sm p-6">
          <h2 class="text-sm font-black text-blue-800 uppercase tracking-widest mb-4">Programación</h2>
          <div>
            <label class={labelCls}>Enviar en fecha/hora (opcional)</label>
            <input type="datetime-local" value={scheduledAt()} onInput={(e) => setScheduledAt(e.currentTarget.value)} class={inputCls} />
            <p class="text-xs text-gray-400 font-medium mt-1.5">Vacío = se envía de inmediato.</p>
          </div>
        </section>

        <div class="flex gap-3 justify-end">
          <A href="/admin/notificaciones" class="px-6 py-3 rounded-2xl bg-gray-100 hover:bg-gray-200 text-gray-700 font-black transition-colors text-sm">
            Cancelar
          </A>
          <button
            type="submit"
            disabled={busy()}
            class="px-6 py-3 rounded-2xl bg-blue-800 hover:bg-blue-900 text-white font-black transition-all active:scale-95 disabled:opacity-50 text-sm shadow-sm"
          >
            {busy() ? "Enviando..." : targetType() === "global" ? "📨 Enviar" : "Crear notificación"}
          </button>
        </div>
      </form>
    </main>
  );
}
