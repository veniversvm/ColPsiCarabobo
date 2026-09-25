// web/src/routes/admin/auditoria/index.tsx
// Bitácora de cambios de la API (audit logs): búsqueda, histórico por psicólogo
// o por staff, estadísticas y exportación CSV. Requiere can_view_logs (menú) y
// can_export_logs solo para el botón CSV; el backend es la barrera real.

import { createResource, createSignal, createEffect, Show, For, Suspense } from "solid-js";
import { useSearchParams } from "@solidjs/router";
import { apiGet, apiDownloadBlob } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { AuditLogDrawer } from "~/components/admin/auditoria/AuditLogDrawer";
import {
  actionLabel,
  entityLabel,
  auditChangesKeys,
  fieldLabel,
  type ApiChangeLog,
  type AuditListResponse,
  type AuditStatsResponse,
  type AuditStat,
} from "~/types/audit";
import { ACTION_LABELS, ENTITY_LABELS } from "~/types/audit";
import { Icon } from "~/components/admin/ui/icons";

type Tab = "general" | "psi" | "staff" | "stats";

interface AdminMeAudit {
  sudo: boolean;
  can_view_logs?: boolean;
  can_export_logs?: boolean;
}

const IC =
  "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text";
const LBL = "block text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1 mb-1";

const fmtDate = (iso: string): string => {
  try {
    return new Date(iso).toLocaleString("es-VE", { day: "2-digit", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit" });
  } catch {
    return iso;
  }
};

export default function AdminAuditoriaPage() {
  const [searchParams] = useSearchParams();

  // ── Permisos del operador (solo cosmético; el backend enmascara con 404) ──
  const [me] = createResource<AdminMeAudit | null>(async () => {
    try {
      return await apiGet<AdminMeAudit>("/admin/me");
    } catch {
      return null;
    }
  });
  const canExport = () => me()?.sudo || me()?.can_export_logs || false;

  // ── Estado de pestañas y filtros ──────────────────────────────────────────
  const [tab, setTab] = createSignal<Tab>("general");
  const [q, setQ] = createSignal("");
  const [suceso, setSuceso] = createSignal("");
  const [entidad, setEntidad] = createSignal("");
  const [actorID, setActorID] = createSignal("");
  const [entityID, setEntityID] = createSignal("");
  const [desde, setDesde] = createSignal("");
  const [hasta, setHasta] = createSignal("");
  const [page, setPage] = createSignal(1);

  // Deep-links: ?psi_id= (desde la ficha) y ?actor_id= (desde staff).
  const [hydrated, setHydrated] = createSignal(false);
  createEffect(() => {
    if (hydrated()) return;
    const psi = searchParams.psi_id as string | undefined;
    const actor = searchParams.actor_id as string | undefined;
    if (psi) {
      setEntityID(psi);
      setTab("psi");
    } else if (actor) {
      setActorID(actor);
      setTab("staff");
    }
    setHydrated(true);
  });

  // Etiqueta legible del psicólogo enlazado (best-effort).
  const [psiLabel] = createResource(
    () => (hydrated() && tab() === "psi" && entityID() ? entityID() : ""),
    async (id: string) => {
      if (!id) return "";
      try {
        const p = await apiGet<{ id: string; first_name?: string; last_name?: string; fpv?: number }>(`/admin/psi/${id}`);
        return `${p?.first_name ?? ""} ${p?.last_name ?? ""}`.trim() || `FPV ${p?.fpv ?? ""}`.trim() || id;
      } catch {
        return id;
      }
    }
  );

  const buildQuery = (includePage: boolean) => {
    const p = new URLSearchParams();
    if (q().trim()) p.set("q", q().trim());
    if (suceso()) p.set("suceso", suceso());
    if (tab() === "general" && entidad()) p.set("entidad", entidad());
    if (tab() === "psi" && !entityID()) p.set("entidad", "psi");
    if (tab() === "staff") p.set("entidad", "staff");
    if (actorID().trim()) p.set("actor_id", actorID().trim());
    if (desde()) p.set("desde", desde());
    if (hasta()) p.set("hasta", hasta());
    if (includePage && page() > 1) p.set("page", String(page()));
    return p.toString();
  };

  const key = () => `${tab()}|${entityID()}|${buildQuery(true)}`;

  // ── Carga de datos por pestaña ────────────────────────────────────────────
  const [result] = createResource<AuditListResponse | null>(
    key,
    async (): Promise<AuditListResponse | null> => {
      try {
        const qs = buildQuery(true);
        const base =
          tab() === "psi" && entityID()
            ? `/admin/audit-logs/psi/${entityID()}`
            : "/admin/audit-logs";
        return await apiGet<AuditListResponse>(`${base}${qs ? `?${qs}` : ""}`);
      } catch {
        return null;
      }
    }
  );

  const [stats] = createResource<AuditStatsResponse | null>(
    () => (tab() === "stats" ? desde() : "__off__"),
    async (d) => {
      try {
        return await apiGet<AuditStatsResponse>(`/admin/audit-logs/stats?desde=${d || ""}`);
      } catch {
        return null;
      }
    }
  );

  const logs = () => result()?.data ?? [];
  const total = () => result()?.total ?? 0;

  const [selected, setSelected] = createSignal<ApiChangeLog | null>(null);

  const resetPage = () => setPage(1);
  const changeTab = (t: Tab) => { setTab(t); resetPage(); setSuceso(""); };

  // ── Exportación CSV (fetch + blob: el auth viaja por cabecera) ────────────
  const [exportBusy, setExportBusy] = createSignal(false);
  const [exportError, setExportError] = createSignal<string | null>(null);
  const handleExport = async () => {
    setExportBusy(true);
    setExportError(null);
    try {
      const blob = await apiDownloadBlob(`/admin/audit-logs/export?${buildQuery(false)}`);
      const a = document.createElement("a");
      a.href = URL.createObjectURL(blob);
      a.download = `audit-logs-${new Date().toISOString().slice(0, 10)}.csv`;
      a.click();
      URL.revokeObjectURL(a.href);
    } catch (err: any) {
      setExportError(getUserFacingError(err));
    } finally {
      setExportBusy(false);
    }
  };

  const statTotal = (list: AuditStat[] | undefined) => (list ?? []).reduce((s, x) => s + x.count, 0);

  return (
    <main class="pb-12 space-y-4">
      {/* ── HEADER ────────────────────────────────────────────────────────── */}
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-4 border-b border-colpsi-border">
        <div class="flex items-center gap-3">
          <span class="inline-flex h-9 w-9 items-center justify-center rounded-md bg-colpsi-bg border border-colpsi-border text-colpsi-blue">
            <Icon name="sliders" class="w-4.5 h-4.5" />
          </span>
          <div>
            <h1 class="text-lg font-semibold text-colpsi-text">Auditoría</h1>
            <p class="text-sm text-colpsi-muted mt-0.5">
              Bitácora de cambios del sistema · quién, qué y cuándo
            </p>
          </div>
        </div>
        <Show when={canExport()}>
          <button
            onClick={handleExport}
            disabled={exportBusy}
            class="inline-flex items-center gap-2 h-9 px-4 rounded-md bg-colpsi-blue hover:bg-colpsi-blue-light text-white font-semibold transition-all text-sm disabled:opacity-60"
          >
            <Show when={exportBusy()} fallback={<Icon name="arrowRight" class="w-4 h-4 rotate-90" />}>
              <span class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            </Show>
            {exportBusy() ? "Generando..." : "Exportar CSV"}
          </button>
        </Show>
      </div>
      <Show when={exportError()}>
        <div class="p-3 rounded-md bg-red-50 text-red-700 border border-red-200 text-sm font-medium inline-flex items-center gap-2 w-full">
          <Icon name="alertTriangle" class="w-4 h-4" />
          {exportError()}
        </div>
      </Show>

      {/* ── TABS ───────────────────────────────────────────────────────────── */}
      <div class="flex flex-wrap gap-2">
        {([
          ["general", "General"],
          ["psi", "Por psicólogo"],
          ["staff", "Por staff"],
          ["stats", "Estadísticas"],
        ] as [Tab, string][]).map(([t, label]) => (
          <button
            onClick={() => changeTab(t)}
            class={`h-9 px-4 rounded-md text-xs font-semibold transition-all border ${
              tab() === t
                ? "bg-colpsi-blue text-white border-colpsi-blue"
                : "bg-white text-colpsi-muted border-colpsi-border hover:border-colpsi-blue/40"
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {/* ── BANNER DE ENLACE PROFUNDO ─────────────────────────────────────── */}
      <Show when={tab() === "psi" && entityID()}>
        <div class="flex items-center gap-3 bg-indigo-50 border border-indigo-200 rounded-md px-4 py-2.5 text-sm">
          <Icon name="user" class="w-4 h-4 text-indigo-600 shrink-0" />
          <p class="font-medium text-indigo-800">
            Historial completo de <span class="font-semibold underline">{psiLabel()}</span>
          </p>
          <button
            onClick={() => { setEntityID(""); setTab("general"); }}
            class="ml-auto inline-flex items-center gap-1 text-xs font-semibold text-indigo-600 hover:underline"
          >
            Quitar filtro
            <Icon name="x" class="w-3.5 h-3.5" />
          </button>
        </div>
      </Show>

      {/* ── FILTROS (todas menos stats) ───────────────────────────────────── */}
      <Show when={tab() !== "stats"}>
        <div class="bg-white rounded-lg border border-colpsi-border p-5">
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div class="col-span-2 md:col-span-1">
              <label class={LBL}>Buscar texto</label>
              <input
                type="text" placeholder="Etiqueta, nombre, FPV..." value={q()}
                onInput={(e) => { setQ(e.currentTarget.value); resetPage(); }}
                class={IC}
              />
            </div>
            <div>
              <label class={LBL}>Suceso</label>
              <select value={suceso()} onChange={(e) => { setSuceso(e.currentTarget.value); resetPage(); }} class={IC}>
                <option value="">Todos</option>
                <For each={Object.entries(ACTION_LABELS)}>
                  {([k, v]) => <option value={k}>{v}</option>}
                </For>
              </select>
            </div>
            <Show when={tab() === "general"}>
              <div>
                <label class={LBL}>Entidad</label>
                <select value={entidad()} onChange={(e) => { setEntidad(e.currentTarget.value); resetPage(); }} class={IC}>
                  <option value="">Todas</option>
                  <For each={Object.entries(ENTITY_LABELS)}>
                    {([k, v]) => <option value={k}>{v}</option>}
                  </For>
                </select>
              </div>
            </Show>
            <Show when={tab() === "staff"}>
              <div>
                <label class={LBL}>Actor (ID del staff)</label>
                <input
                  type="text" placeholder="UUID..." value={actorID()}
                  onInput={(e) => { setActorID(e.currentTarget.value); resetPage(); }}
                  class={IC}
                />
              </div>
            </Show>
            <div>
              <label class={LBL}>Desde</label>
              <input type="date" value={desde()} onInput={(e) => { setDesde(e.currentTarget.value); resetPage(); }} class={IC} />
            </div>
            <div>
              <label class={LBL}>Hasta</label>
              <input type="date" value={hasta()} onInput={(e) => { setHasta(e.currentTarget.value); resetPage(); }} class={IC} />
            </div>
          </div>
        </div>
      </Show>

      {/* ── ESTADÍSTICAS ───────────────────────────────────────────────────── */}
      <Show when={tab() === "stats"}>
        <div class="bg-white rounded-lg border border-colpsi-border p-5">
          <div class="flex items-end gap-4 flex-wrap">
            <div class="w-48">
              <label class={LBL}>Desde (por defecto: últimos 30 días)</label>
              <input type="date" value={desde()} onInput={(e) => setDesde(e.currentTarget.value)} class={IC} />
            </div>
            <p class="text-xs text-colpsi-muted font-medium pb-2.5">
              {statTotal(stats()?.stats)} sucesos registrados
            </p>
          </div>
        </div>

        <Show when={stats.loading && !stats()}>
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <For each={Array(8).fill(0)}>
              {() => <div class="h-24 bg-white rounded-lg border border-colpsi-border animate-pulse" />}
            </For>
          </div>
        </Show>
        <Show when={!stats.loading && stats() && (stats()?.stats ?? []).length === 0}>
          <div class="text-center py-16 bg-white rounded-lg border border-colpsi-border">
            <span class="inline-flex h-12 w-12 items-center justify-center rounded-md bg-colpsi-bg text-colpsi-muted mb-4">
              <Icon name="inbox" class="w-6 h-6" />
            </span>
            <p class="font-semibold text-colpsi-text">Sin sucesos en el período</p>
          </div>
        </Show>
        <Show when={!stats.loading && (stats()?.stats ?? []).length > 0}>
          <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            <For each={stats()?.stats ?? []}>
              {(s) => (
                <div class="bg-white rounded-lg border border-colpsi-border p-5">
                  <p class="text-2xl font-semibold text-colpsi-blue">{s.count}</p>
                  <p class="text-xs font-semibold uppercase tracking-wide text-colpsi-muted mt-1">
                    {entityLabel(s.entity)}
                  </p>
                  <p class="text-[11px] text-colpsi-muted/80 font-medium">{actionLabel(s.action)}</p>
                </div>
              )}
            </For>
          </div>
        </Show>
      </Show>

      {/* ── TABLA DE LOGS ──────────────────────────────────────────────────── */}
      <Show when={tab() !== "stats"}>
        <Suspense fallback={<div class="space-y-3"><For each={Array(6).fill(0)}>{() => <div class="h-20 bg-white animate-pulse rounded-lg border border-colpsi-border" />}</For></div>}>
          <Show when={!result.loading && logs().length === 0}>
            <div class="text-center py-16 bg-white rounded-lg border border-colpsi-border">
              <span class="inline-flex h-12 w-12 items-center justify-center rounded-md bg-colpsi-bg text-colpsi-muted mb-4">
                <Icon name="fileText" class="w-6 h-6" />
              </span>
              <p class="font-semibold text-colpsi-text">Sin registros para los filtros aplicados</p>
            </div>
          </Show>

          <Show when={logs().length > 0}>
            <PaginationBar page={page} total={total} setPage={setPage} />
            <div class="space-y-2">
              <For each={logs()}>
                {(log) => (
                  <button
                    onClick={() => setSelected(log)}
                    class="w-full text-left bg-white rounded-md border border-colpsi-border hover:border-colpsi-blue/40 hover:bg-colpsi-bg/40 transition-all p-4"
                  >
                    <div class="flex flex-wrap items-center gap-x-3 gap-y-1 mb-1">
                      <span class={`inline-flex items-center px-2 py-0.5 rounded text-[10px] font-semibold uppercase tracking-wide ${
                        log.action === "create" ? "bg-emerald-50 text-emerald-700 border border-emerald-200"
                        : log.action === "delete" ? "bg-red-50 text-colpsi-red border border-red-200"
                        : log.action === "update" || log.action === "update_permissions" || log.action === "role_change" ? "bg-amber-50 text-amber-700 border border-amber-200"
                        : log.action === "login" || log.action === "logout" ? "bg-blue-50 text-colpsi-blue border border-blue-200"
                        : "bg-slate-100 text-slate-600 border border-slate-200"
                      }`}>
                        {actionLabel(log.action)}
                      </span>
                      <span class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-semibold uppercase tracking-wide bg-slate-100 text-slate-600 border border-slate-200">
                        {entityLabel(log.entity)}
                      </span>
                      <span class="text-sm font-semibold text-colpsi-text truncate max-w-[45%]">{log.entity_label || log.entity_id}</span>
                    </div>
                    <div class="flex flex-wrap items-center gap-x-4 gap-y-0.5 text-[11px] text-colpsi-muted font-medium">
                      <span class="inline-flex items-center gap-1"><Icon name="user" class="w-3.5 h-3.5" /> <span class="font-semibold text-colpsi-text">{log.actor_username || "—"}</span> <span class="text-colpsi-muted/70">({log.actor_role || "—"})</span></span>
                      <span class="inline-flex items-center gap-1"><Icon name="clock" class="w-3.5 h-3.5" /> {fmtDate(log.created_at)}</span>
                      <Show when={log.ip}><span class="inline-flex items-center gap-1"><Icon name="globe" class="w-3.5 h-3.5" /> {log.ip}</span></Show>
                    </div>
                    <Show when={auditChangesKeys(log.changes).length > 0}>
                      <p class="inline-flex items-center gap-1 text-[11px] font-medium text-colpsi-muted truncate mt-1.5">
                        <Icon name="pencil" class="w-3 h-3" /> {auditChangesKeys(log.changes).slice(0, 4).map(fieldLabel).join(" · ")}
                        <Show when={auditChangesKeys(log.changes).length > 4}> …</Show>
                      </p>
                    </Show>
                  </button>
                )}
              </For>
            </div>

            {/* ── Paginación ── */}
            <PaginationBar page={page} total={total} setPage={setPage} />
          </Show>
        </Suspense>
      </Show>

      {/* ── DRAWER DE DETALLE ──────────────────────────────────────────────── */}
      <AuditLogDrawer log={selected()} onClose={() => setSelected(null)} />
    </main>
  );
}

// Barra de paginación reutilizable: se renderiza arriba y abajo del listado.
// Solo aparece cuando hay más de una página (total > 20) y comparte el estado
// `page` del listado; cada cambio de filtro resetea a la página 1.
function PaginationBar(props: {
  page: () => number;
  total: () => number;
  setPage: (updater: (prev: number) => number) => void;
}) {
  return (
    <Show when={props.total() > 20}>
      <div class="flex items-center justify-center gap-4 py-4">
        <button
          onClick={() => props.setPage((p) => Math.max(1, p - 1))}
          disabled={props.page() <= 1}
          class="inline-flex items-center gap-1.5 h-9 px-4 rounded-md bg-white border border-colpsi-border text-colpsi-muted font-medium text-xs disabled:opacity-40 hover:border-colpsi-blue/40 transition-colors"
        >
          <Icon name="chevronRight" class="w-3.5 h-3.5 rotate-180" />
          Anterior
        </button>
        <span class="text-xs font-medium text-colpsi-muted">
          Página {props.page()} · {props.total()} registros
        </span>
        <button
          onClick={() => props.setPage((p) => p + 1)}
          disabled={props.page() * 20 >= props.total()}
          class="inline-flex items-center gap-1.5 h-9 px-4 rounded-md bg-white border border-colpsi-border text-colpsi-muted font-medium text-xs disabled:opacity-40 hover:border-colpsi-blue/40 transition-colors"
        >
          Siguiente
          <Icon name="chevronRight" class="w-3.5 h-3.5" />
        </button>
      </div>
    </Show>
  );
}