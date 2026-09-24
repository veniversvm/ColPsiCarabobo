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

type Tab = "general" | "psi" | "staff" | "stats";

interface AdminMeAudit {
  sudo: boolean;
  can_view_logs?: boolean;
  can_export_logs?: boolean;
}

const IC =
  "w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-3 py-2 outline-none transition-all text-gray-800 text-sm";
const LBL = "block text-[10px] font-black text-gray-500 uppercase tracking-widest ml-1 mb-1";

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
    <main class="pb-20 animate-in fade-in duration-500">
      {/* ── HEADER ────────────────────────────────────────────────────────── */}
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6 bg-white p-6 rounded-3xl shadow-sm border border-colpsi-border">
        <div>
          <h1 class="text-2xl font-black text-blue-900 uppercase tracking-tight flex items-center gap-2">
            <span>🧾</span> Auditoría
          </h1>
          <p class="text-gray-400 text-sm mt-0.5 font-medium">
            Bitácora de cambios del sistema · quién, qué y cuándo
          </p>
        </div>
        <Show when={canExport()}>
          <button
            onClick={handleExport}
            disabled={exportBusy}
            class="inline-flex items-center gap-2 bg-blue-800 hover:bg-blue-900 text-white font-black px-6 py-3 rounded-2xl shadow-lg hover:scale-105 active:scale-95 transition-all text-sm disabled:opacity-60"
          >
            <span>{exportBusy() ? "⏳" : "⬇"}</span> {exportBusy() ? "Generando..." : "Exportar CSV"}
          </button>
        </Show>
      </div>
      <Show when={exportError()}>
        <div class="mb-4 p-4 rounded-2xl bg-red-50 text-red-800 font-bold text-sm border-l-4 border-red-500 shadow-sm">
          ⚠️ {exportError()}
        </div>
      </Show>

      {/* ── TABS ───────────────────────────────────────────────────────────── */}
      <div class="flex flex-wrap gap-2 mb-6">
        {([
          ["general", "General"],
          ["psi", "Por psicólogo"],
          ["staff", "Por staff"],
          ["stats", "Estadísticas"],
        ] as [Tab, string][]).map(([t, label]) => (
          <button
            onClick={() => changeTab(t)}
            class={`px-4 py-2.5 rounded-xl text-xs font-black uppercase tracking-wide transition-all border-2 ${
              tab() === t
                ? "bg-blue-800 text-white border-blue-800"
                : "bg-white text-gray-500 border-gray-200 hover:border-gray-300"
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {/* ── BANNER DE ENLACE PROFUNDO ─────────────────────────────────────── */}
      <Show when={tab() === "psi" && entityID()}>
        <div class="mb-6 flex items-center gap-3 bg-indigo-50 border-l-4 border-indigo-500 rounded-2xl px-4 py-3 text-sm">
          <span class="text-lg">👤</span>
          <p class="font-bold text-indigo-800">
            Historial completo de <span class="underline">{psiLabel()}</span>
          </p>
          <button
            onClick={() => { setEntityID(""); setTab("general"); }}
            class="ml-auto text-xs font-black text-indigo-600 hover:underline"
          >
            Quitar filtro ✕
          </button>
        </div>
      </Show>

      {/* ── FILTROS (todas menos stats) ───────────────────────────────────── */}
      <Show when={tab() !== "stats"}>
        <div class="bg-white rounded-3xl border border-colpsi-border shadow-sm p-5 mb-6">
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
        <div class="bg-white rounded-3xl border border-colpsi-border shadow-sm p-5 mb-6">
          <div class="flex items-end gap-4 flex-wrap">
            <div class="w-48">
              <label class={LBL}>Desde (por defecto: últimos 30 días)</label>
              <input type="date" value={desde()} onInput={(e) => setDesde(e.currentTarget.value)} class={IC} />
            </div>
            <p class="text-xs text-gray-400 font-bold pb-2">
              {statTotal(stats()?.stats)} sucesos registrados
            </p>
          </div>
        </div>

        <Show when={stats.loading && !stats()}>
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <For each={Array(8).fill(0)}>
              {() => <div class="h-24 bg-white rounded-2xl border border-colpsi-border animate-pulse" />}
            </For>
          </div>
        </Show>
        <Show when={!stats.loading && stats() && (stats()?.stats ?? []).length === 0}>
          <div class="text-center py-20 bg-white rounded-3xl border border-colpsi-border">
            <p class="text-4xl mb-3">📭</p>
            <p class="text-gray-400 font-bold">Sin sucesos en el período</p>
          </div>
        </Show>
        <Show when={!stats.loading && (stats()?.stats ?? []).length > 0}>
          <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            <For each={stats()?.stats ?? []}>
              {(s) => (
                <div class="bg-white rounded-2xl border border-colpsi-border p-5 shadow-sm">
                  <p class="text-3xl font-black text-blue-800">{s.count}</p>
                  <p class="text-xs font-black uppercase tracking-widest text-gray-500 mt-1">
                    {entityLabel(s.entity)}
                  </p>
                  <p class="text-[11px] text-gray-400 font-bold">{actionLabel(s.action)}</p>
                </div>
              )}
            </For>
          </div>
        </Show>
      </Show>

      {/* ── TABLA DE LOGS ──────────────────────────────────────────────────── */}
      <Show when={tab() !== "stats"}>
        <Suspense fallback={<div class="space-y-3"><For each={Array(6).fill(0)}>{() => <div class="h-20 bg-white animate-pulse rounded-2xl border border-colpsi-border" />}</For></div>}>
          <Show when={!result.loading && logs().length === 0}>
            <div class="text-center py-20 bg-white rounded-3xl border border-colpsi-border">
              <p class="text-5xl mb-4">🧾</p>
              <p class="text-gray-400 font-bold">Sin registros para los filtros aplicados</p>
            </div>
          </Show>

          <Show when={logs().length > 0}>
            <div class="space-y-3">
              <For each={logs()}>
                {(log) => (
                  <button
                    onClick={() => setSelected(log)}
                    class="w-full text-left bg-white rounded-2xl border border-colpsi-border hover:border-blue-200 hover:shadow-md transition-all p-4"
                  >
                    <div class="flex flex-wrap items-center gap-x-3 gap-y-1 mb-1">
                      <span class={`text-[10px] font-black px-2 py-0.5 rounded-lg uppercase tracking-wider ${
                        log.action === "create" ? "bg-emerald-100 text-emerald-700"
                        : log.action === "delete" ? "bg-red-100 text-red-700"
                        : log.action === "update" || log.action === "update_permissions" || log.action === "role_change" ? "bg-amber-100 text-amber-700"
                        : log.action === "login" || log.action === "logout" ? "bg-blue-50 text-blue-600"
                        : "bg-gray-100 text-gray-600"
                      }`}>
                        {actionLabel(log.action)}
                      </span>
                      <span class="text-[10px] font-black px-2 py-0.5 rounded-lg uppercase tracking-wider bg-slate-100 text-slate-600">
                        {entityLabel(log.entity)}
                      </span>
                      <span class="text-sm font-black text-gray-900 truncate max-w-[45%]">{log.entity_label || log.entity_id}</span>
                    </div>
                    <div class="flex flex-wrap items-center gap-x-4 gap-y-0.5 text-[11px] text-gray-400 font-medium">
                      <span>👤 <span class="font-bold text-gray-600">{log.actor_username || "—"}</span> <span class="text-gray-300">({log.actor_role || "—"})</span></span>
                      <span>🕒 {fmtDate(log.created_at)}</span>
                      <Show when={log.ip}><span>🌐 {log.ip}</span></Show>
                    </div>
                    <Show when={auditChangesKeys(log.changes).length > 0}>
                      <p class="text-[11px] font-black text-gray-500 truncate mt-1.5">
                        ✏️ {auditChangesKeys(log.changes).slice(0, 4).map(fieldLabel).join(" · ")}
                        <Show when={auditChangesKeys(log.changes).length > 4}> …</Show>
                      </p>
                    </Show>
                  </button>
                )}
              </For>
            </div>

            {/* ── Paginación ── */}
            <Show when={total() > 20}>
              <div class="flex items-center justify-center gap-4 mt-6">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page() <= 1}
                  class="px-4 py-2 rounded-xl bg-white border-2 border-gray-200 text-gray-600 font-black text-xs disabled:opacity-40"
                >
                  ← Anterior
                </button>
                <span class="text-xs font-black text-gray-500">
                  Página {page()} · {total()} registros
                </span>
                <button
                  onClick={() => setPage((p) => p + 1)}
                  disabled={page() * 20 >= total()}
                  class="px-4 py-2 rounded-xl bg-white border-2 border-gray-200 text-gray-600 font-black text-xs disabled:opacity-40"
                >
                  Siguiente →
                </button>
              </div>
            </Show>
          </Show>
        </Suspense>
      </Show>

      {/* ── DRAWER DE DETALLE ──────────────────────────────────────────────── */}
      <AuditLogDrawer log={selected()} onClose={() => setSelected(null)} />
    </main>
  );
}