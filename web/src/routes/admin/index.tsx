// web/src/routes/admin/index.tsx
import { createResource, createSignal, For, Show, onMount, onCleanup, createEffect } from "solid-js";
import { apiGet } from "~/lib/api";

import { AdminLoadingSkeleton } from "~/components/admin/dashboard/AdminLoadingSkeleton";
import { Sparkline }            from "~/components/admin/dashboard/Sparkline";
import { RankingList }          from "~/components/admin/dashboard/RankingList";
import { TopProfiles }          from "~/components/admin/dashboard/TopProfiles";
import { ActiveSessionsBanner } from "~/components/admin/dashboard/ActiveSessionsBanner";
import { BirthdayBanner } from "~/components/admin/dashboard/BirthdayBanner";
import { ReceptionSwitchesCard } from "~/components/admin/settings/ReceptionSwitchesCard";

import { Panel }     from "~/components/admin/ui/Panel";
import { KpiStrip }  from "~/components/admin/ui/KpiStrip";
import { PageHeader } from "~/components/admin/ui/PageHeader";
import { Button }    from "~/components/admin/ui/Button";
import { Icon }      from "~/components/admin/ui/icons";

// ── Tipos ─────────────────────────────────────────────────────────────────────
interface TopItem     { value: string; count: number; name: string }
interface DailyCount  { date: string;  count: number }
interface TopProfile  { psi_id: string; first_name: string; last_name: string; fpv: number; count: number }

interface DashboardStats {
  logins_total: number;       logins_today: number
  logins_this_week: number;   logins_this_month: number
  unique_users_today: number

  page_views_total: number;   page_views_today: number
  page_views_this_week: number
  unique_visitors_today: number; unique_visitors_week: number

  searches_total: number;     searches_today: number
  searches_this_week: number

  profile_views_total: number; profile_views_today: number
  profile_views_week: number

  active_sessions_now: number

  top_specialties:  TopItem[]
  top_municipios:   TopItem[]
  top_search_terms: TopItem[]
  top_profiles:     TopProfile[]

  login_trend: DailyCount[]
  view_trend:  DailyCount[]
}

// ─────────────────────────────────────────────────────────────────────────────
export default function AdminDashboard() {
  const [stats, { refetch }] = createResource<DashboardStats>(() =>
    apiGet("/admin/dashboard/stats")
  );

  const [cachedStats, setCachedStats] = createSignal<DashboardStats | undefined>(undefined);
  const [lastRefresh, setLastRefresh] = createSignal(new Date());

  // Actualizar caché solo cuando hay datos nuevos
  createEffect(() => {
    const data = stats();
    if (data) {
      setCachedStats(data);
      setLastRefresh(new Date());
    }
  });

  let interval: ReturnType<typeof setInterval>;
  onMount(() => { interval = setInterval(refetch, 60_000 * 15); });
  onCleanup(() => clearInterval(interval));

  // Loading SOLO en carga inicial
  const initialLoading = () => stats.loading && !cachedStats();

  const fmt = (n?: number) => (n ?? 0).toLocaleString("es-VE");

  const sectionTitle = (t: string) => (
    <h2 class="text-[11px] font-semibold uppercase tracking-wide text-colpsi-muted">{t}</h2>
  );

  return (
    <div class="space-y-5">

      {/* ── HEADER ──────────────────────────────────────────────────────── */}
      <PageHeader
        title="Panel de Control"
        description={(() => {
          const d = lastRefresh();
          return `Métricas del portal en tiempo real · Actualiza cada 15 min · Última actualización: ${d.toLocaleTimeString("es-VE")}`;
        })()}
        actions={
          <Button
            variant="secondary"
            onClick={() => { refetch(); }}
            disabled={stats.loading}
          >
            <Icon name="refresh" class={stats.loading ? "animate-spin" : ""} />
            {stats.loading && cachedStats() ? "Actualizando..." : "Actualizar"}
          </Button>
        }
      />

      {/* ── ERROR ───────────────────────────────────────────────────────── */}
      <Show when={stats.error}>
        <div class="border border-red-200 bg-red-50 rounded-md p-3 text-sm text-red-700 font-medium">
          ⚠️ Error al cargar estadísticas — verifica que el servidor esté activo.
        </div>
      </Show>

      {/* ── BANNER: SESIONES ACTIVAS ─────────────────────────────────────── */}
      <Show when={initialLoading()}>
        <AdminLoadingSkeleton variant="banner" />
      </Show>
      <Show when={cachedStats()}>
        {(s) => <ActiveSessionsBanner count={s().active_sessions_now} />}
      </Show>

      {/* ── BANNER: CUMPLEAÑOS DEL AGREMIADO (opt-in) ────────────────────── */}
      <BirthdayBanner />

      {/* ── RECEPCIÓN GLOBAL (solo SUDO gestiona) ───────────────────────── */}
      <ReceptionSwitchesCard />

      {/* ── SECCIÓN: LOGINS ──────────────────────────────────────────────── */}
      <section class="space-y-2">
        {sectionTitle("Inicios de sesión")}
        <Show when={initialLoading()}>
          <AdminLoadingSkeleton variant="kpi" count={4} />
        </Show>
        <Show when={cachedStats()}>
          {(s) => (
            <Panel flush>
              <KpiStrip
                columns={4}
                cells={[
                  { label: "Hoy", value: s().logins_today, sub: `Únicos: ${fmt(s().unique_users_today)}`, dot: "bg-colpsi-yellow" },
                  { label: "Esta semana", value: s().logins_this_week, dot: "bg-blue-400" },
                  { label: "Este mes", value: s().logins_this_month, dot: "bg-blue-400" },
                  { label: "Total histórico", value: s().logins_total, dot: "bg-slate-300" },
                ]}
              />
            </Panel>
          )}
        </Show>
      </section>

      {/* ── SECCIÓN: VISITAS ─────────────────────────────────────────────── */}
      <section class="space-y-2">
        {sectionTitle("Visitas al portal")}
        <Show when={initialLoading()}>
          <AdminLoadingSkeleton variant="kpi" count={4} />
        </Show>
        <Show when={cachedStats()}>
          {(s) => (
            <Panel flush>
              <KpiStrip
                columns={4}
                cells={[
                  {
                    label: "Hoy",
                    value: s().page_views_today,
                    sub: `Visitantes únicos: ${fmt(s().unique_visitors_today)}`,
                    dot: "bg-emerald-500",
                  },
                  { label: "Esta semana", value: s().page_views_this_week, sub: `Únicos: ${fmt(s().unique_visitors_week)}`, dot: "bg-emerald-400" },
                  { label: "Total páginas vistas", value: s().page_views_total, dot: "bg-slate-300" },
                  { label: "Búsquedas hoy", value: s().searches_today, sub: `Semana: ${fmt(s().searches_this_week)} · Total: ${fmt(s().searches_total)}`, dot: "bg-blue-400" },
                ]}
              />
            </Panel>
          )}
        </Show>
      </section>

      {/* ── SECCIÓN: PERFILES ────────────────────────────────────────────── */}
      <section class="space-y-2">
        {sectionTitle("Visitas a perfiles")}
        <Show when={initialLoading()}>
          <AdminLoadingSkeleton variant="kpi" count={3} />
        </Show>
        <Show when={cachedStats()}>
          {(s) => (
            <Panel flush>
              <KpiStrip
                columns={3}
                cells={[
                  { label: "Hoy", value: s().profile_views_today, dot: "bg-amber-400" },
                  { label: "Esta semana", value: s().profile_views_week, dot: "bg-amber-400" },
                  { label: "Total", value: s().profile_views_total, dot: "bg-slate-300" },
                ]}
              />
            </Panel>
          )}
        </Show>
      </section>

      {/* ── SECCIÓN: TENDENCIAS ──────────────────────────────────────────── */}
      <section class="space-y-2">
        {sectionTitle("Tendencia diaria")}
        <Show when={initialLoading()}>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <AdminLoadingSkeleton variant="chart" />
            <AdminLoadingSkeleton variant="chart" />
          </div>
        </Show>
        <Show when={cachedStats()}>
          {(s) => (
            <Panel flush>
              <div class="grid grid-cols-1 md:grid-cols-2 divide-y md:divide-y-0 md:divide-x divide-colpsi-border">
                <Sparkline data={s().login_trend ?? []} color="#1e40af" label="Logins" />
                <Sparkline data={s().view_trend  ?? []} color="#16a34a" label="Visitas al portal" />
              </div>
            </Panel>
          )}
        </Show>
      </section>

      {/* ── SECCIÓN: RANKINGS ────────────────────────────────────────────── */}
      <section class="space-y-2">
        {sectionTitle("Análisis de búsquedas")}
        <Show when={initialLoading()}>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <AdminLoadingSkeleton variant="list" rows={6} />
            <AdminLoadingSkeleton variant="list" rows={6} />
            <AdminLoadingSkeleton variant="list" rows={6} />
          </div>
        </Show>
        <Show when={cachedStats()}>
          {(s) => (
            <Panel flush>
              <div class="grid grid-cols-1 md:grid-cols-3 divide-y md:divide-y-0 md:divide-x divide-colpsi-border">
                <RankingList title="Especialidades más buscadas" items={s().top_specialties  ?? []} />
                <RankingList title="Municipios más buscados"     items={s().top_municipios   ?? []} />
                <RankingList title="Términos de búsqueda"        items={s().top_search_terms ?? []} />
              </div>
            </Panel>
          )}
        </Show>
      </section>

      {/* ── SECCIÓN: TOP PERFILES ────────────────────────────────────────── */}
      <Show when={initialLoading()}>
        <div class="space-y-2">
          <div class="w-64 h-2.5 bg-slate-100 rounded animate-pulse" />
          <div class="grid grid-cols-2 md:grid-cols-5 gap-3">
            <For each={Array(5).fill(0)}>
              {() => <div class="h-24 bg-white rounded-lg border border-colpsi-border animate-pulse" />}
            </For>
          </div>
        </div>
      </Show>
      <Show when={cachedStats()}>
        {(s) => <TopProfiles profiles={s().top_profiles ?? []} />}
      </Show>

    </div>
  );
}