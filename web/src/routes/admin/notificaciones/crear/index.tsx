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
import { MUNICIPIOS_CARABOBO, ESTADOS_VENEZUELA_INCL_CARABOBO } from "~/lib/geo";
import { Icon, IconName } from "~/components/admin/ui/icons";

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

const TARGET_OPTIONS: { value: NotificationTargetType; label: string; icon: IconName; desc: string }[] = [
  { value: "global", label: "Global", icon: "globe", desc: "Todos los agremiados activos" },
  { value: "individual", label: "Individual", icon: "user", desc: "Psicólogos específicos" },
  { value: "group", label: "Por grupo", icon: "sliders", desc: "Según filtros (zona, género, etc.)" },
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

  const inputCls = "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text";
  const labelCls = "text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide mb-1.5 block";

  return (
    <main class="space-y-4 pb-12 max-w-3xl">
      <div class="flex items-center gap-3 pb-4 border-b border-colpsi-border">
        <A href="/admin/notificaciones" class="inline-flex items-center justify-center h-8 w-8 rounded-md border border-colpsi-border bg-white text-colpsi-muted hover:text-colpsi-blue hover:bg-colpsi-bg transition-colors" title="Volver">
          <Icon name="chevronRight" class="w-4 h-4 rotate-180" />
        </A>
        <div>
          <h1 class="text-lg font-semibold text-colpsi-text">Nueva Notificación</h1>
          <p class="text-sm text-colpsi-muted mt-0.5">Configura el destino y el contenido del aviso</p>
        </div>
      </div>

      <Show when={error()}>
        <div class="bg-red-50 border border-red-200 text-red-700 text-sm font-medium px-3 py-2.5 rounded-md">{error()}</div>
      </Show>

      <Show when={successId()}>
        <div class="bg-emerald-50 border border-emerald-200 text-emerald-700 text-sm font-medium px-3 py-2.5 rounded-md">
          Notificación creada. Redirigiendo...
        </div>
      </Show>

      <form onSubmit={handleSubmit} class="space-y-4">
        {/* Contenido */}
        <section class="bg-white rounded-lg border border-colpsi-border p-5 space-y-4">
          <h2 class="text-base font-semibold text-colpsi-text border-b border-colpsi-border pb-3">Contenido</h2>
          <div>
            <label class={labelCls}>Título *</label>
            <input value={title()} onInput={(e) => setTitle(e.currentTarget.value)} maxlength={255} class={inputCls} placeholder="Ej: Asamblea general ordinaria" />
          </div>
          <div>
            <label class={labelCls}>Mensaje *</label>
            <textarea value={message()} onInput={(e) => setMessage(e.currentTarget.value)} rows={4} class={`${inputCls} resize-none min-h-28`} placeholder="Redacta el comunicado..." />
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
        <section class="bg-white rounded-lg border border-colpsi-border p-5 space-y-4">
          <h2 class="text-base font-semibold text-colpsi-text border-b border-colpsi-border pb-3">Destino</h2>
          <div class="grid grid-cols-3 gap-2">
            <For each={TARGET_OPTIONS}>
              {(opt) => (
                <button
                  type="button"
                  onClick={() => setTargetType(opt.value)}
                  class={`rounded-md border p-3 text-center transition-colors ${
                    targetType() === opt.value
                      ? "border-colpsi-blue bg-colpsi-blue/5"
                      : "border-colpsi-border hover:border-colpsi-blue/40 bg-white"
                  }`}
                >
                  <Icon name={opt.icon} class={`w-5 h-5 mx-auto block mb-1.5 ${targetType() === opt.value ? "text-colpsi-blue" : "text-colpsi-muted"}`} />
                  <span class={`block text-sm font-semibold ${targetType() === opt.value ? "text-colpsi-blue" : "text-colpsi-text"}`}>{opt.label}</span>
                  <span class="block text-[10px] text-colpsi-muted font-medium mt-0.5">{opt.desc}</span>
                </button>
              )}
            </For>
          </div>

          <Show when={targetType() === "individual"}>
            <div class="space-y-3">
              <div class="relative">
                <Icon name="search" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                <input
                  type="search"
                  value={search()}
                  onInput={onSearchInput}
                  placeholder="Buscar por nombre, CI o FPV..."
                  class={`${inputCls} pl-9`}
                />
              </div>
              <Suspense fallback={<div class="h-24 bg-white animate-pulse rounded-md border border-colpsi-border" />}>
                <div class="bg-colpsi-bg rounded-md border border-colpsi-border p-3 max-h-72 overflow-y-auto space-y-1.5">
                  <Show when={(psiUsers()?.data ?? []).length === 0}>
                    <p class="text-sm text-colpsi-muted p-3">
                      {debouncedSearch() ? "No hay psicólogos que coincidan con la búsqueda." : "No hay psicólogos registrados."}
                    </p>
                  </Show>
                  <For each={psiUsers()?.data ?? []}>
                    {(u) => (
                      <label class="flex items-center gap-3 bg-white rounded-md px-3 py-2 cursor-pointer border border-colpsi-border">
                        <input
                          type="checkbox"
                          checked={selected().has(u.id)}
                          onChange={() => toggleUser(u)}
                          class="w-4 h-4 accent-colpsi-blue shrink-0"
                        />
                        <span class="text-sm font-medium text-colpsi-text truncate">{u.first_name} {u.last_name}</span>
                        <span class="text-xs text-colpsi-muted whitespace-nowrap">CI {u.ci} · FPV {u.fpv}</span>
                        <span class="ml-auto text-xs text-colpsi-muted truncate max-w-[160px]">{u.email}</span>
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
                      class="h-7 px-3 rounded-md bg-white border border-colpsi-border text-xs font-medium text-colpsi-muted hover:text-colpsi-blue hover:bg-colpsi-bg disabled:opacity-40 transition-colors"
                    >
                      ← Anterior
                    </button>
                    <span class="text-xs text-colpsi-muted font-medium">
                      Página {psiUsers()?.page ?? 1} de {psiUsers()?.total_pages ?? 1} · {psiUsers()?.total ?? 0} psicólogos
                    </span>
                    <button
                      type="button"
                      disabled={page() >= (psiUsers()?.total_pages ?? 1)}
                      onClick={() => setPage((p) => p + 1)}
                      class="h-7 px-3 rounded-md bg-white border border-colpsi-border text-xs font-medium text-colpsi-muted hover:text-colpsi-blue hover:bg-colpsi-bg disabled:opacity-40 transition-colors"
                    >
                      Siguiente →
                    </button>
                  </div>
                </Show>
              </Suspense>

              <p class="text-xs text-colpsi-muted font-medium">{selected().size} seleccionado(s)</p>

              <Show when={selected().size > 0}>
                <div class="flex flex-wrap gap-1.5">
                  <For each={Object.entries(selectedDetails())}>
                    {([id, d]) => (
                      <span class="inline-flex items-center gap-1.5 bg-colpsi-blue/5 border border-colpsi-blue/20 text-colpsi-blue text-xs font-medium px-2 py-1 rounded-md">
                        {d.name}
                        <button
                          type="button"
                          onClick={() => removeSelected(id)}
                          class="text-colpsi-blue/60 hover:text-colpsi-blue"
                          aria-label={`Quitar a ${d.name}`}
                        >
                          <Icon name="x" class="w-3 h-3" />
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
                <select value={municipality()} onChange={(e) => setMunicipality(e.currentTarget.value)} class={inputCls}>
                  <option value="">Todos</option>
                  {MUNICIPIOS_CARABOBO.map((m) => (
                    <option value={m}>{m}</option>
                  ))}
                </select>
              </div>
              <div>
                <label class={labelCls}>Estado</label>
                <select value={state()} onChange={(e) => setState(e.currentTarget.value)} class={inputCls}>
                  <option value="">Todos</option>
                  {ESTADOS_VENEZUELA_INCL_CARABOBO.map((s) => (
                    <option value={s}>{s}</option>
                  ))}
                </select>
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
              class="inline-flex items-center gap-1.5 h-9 px-3 bg-colpsi-blue/5 hover:bg-colpsi-blue/10 text-colpsi-blue font-semibold rounded-md text-xs transition-colors disabled:opacity-50"
            >
              <Icon name="eye" class="w-3.5 h-3.5" />
              Previsualizar destinatarios
            </button>
            <Show when={preview()}>
              <div class="bg-colpsi-blue/5 border border-colpsi-blue/20 text-colpsi-blue text-sm font-medium px-3 py-2.5 rounded-md">
                {preview()?.total_recipients ?? 0} destinatario(s) potencial(es)
              </div>
            </Show>
          </Show>
        </section>

        {/* Programación */}
        <section class="bg-white rounded-lg border border-colpsi-border p-5">
          <h2 class="text-base font-semibold text-colpsi-text border-b border-colpsi-border pb-3 mb-4">Programación</h2>
          <div>
            <label class={labelCls}>Enviar en fecha/hora (opcional)</label>
            <input type="datetime-local" value={scheduledAt()} onInput={(e) => setScheduledAt(e.currentTarget.value)} class={inputCls} />
            <p class="text-xs text-colpsi-muted font-medium mt-1.5">Vacío = se envía de inmediato.</p>
          </div>
        </section>

        <div class="flex justify-end gap-2">
          <A href="/admin/notificaciones" class="h-10 px-4 rounded-md bg-white border border-colpsi-border hover:bg-colpsi-bg text-colpsi-text font-medium transition-colors text-sm inline-flex items-center">
            Cancelar
          </A>
          <button
            type="submit"
            disabled={busy()}
            class="inline-flex items-center justify-center gap-2 h-10 px-5 rounded-md bg-colpsi-blue hover:bg-colpsi-blue-light text-white font-semibold transition-colors disabled:opacity-50 text-sm"
          >
            <Icon name="send" class="w-4 h-4" />
            {busy() ? "Enviando..." : targetType() === "global" ? "Enviar" : "Crear notificación"}
          </button>
        </div>
      </form>
    </main>
  );
}