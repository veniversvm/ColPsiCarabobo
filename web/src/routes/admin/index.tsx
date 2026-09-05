// web/src/routes/admin/index.tsx
import { createResource, createSignal, For, Show, onMount, onCleanup, createEffect } from "solid-js";
import { apiGet } from "~/lib/api";

import { AdminLoadingSkeleton } from "~/components/admin/dashboard/AdminLoadingSkeleton";
import { StatCard }             from "~/components/admin/dashboard/StatCard";
import { Sparkline }            from "~/components/admin/dashboard/Sparkline";
import { RankingList }          from "~/components/admin/dashboard/RankingList";
import { TopProfiles }          from "~/components/admin/dashboard/TopProfiles";
import { ActiveSessionsBanner } from "~/components/admin/dashboard/ActiveSessionsBanner";
import { BirthdayBanner } from "~/components/admin/dashboard/BirthdayBanner";

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

  return (
    <div class="space-y-8 animate-in fade-in duration-500 pb-24">

      {/* ── HEADER ──────────────────────────────────────────────────────── */}
      <div class="flex items-start justify-between">
        <div>
          <h1 class="text-2xl font-black text-colpsi-blue">Panel de Control</h1>
          <div class="flex items-center gap-2 mt-1">
            <p class="text-gray-400 text-sm">
              Métricas del portal en tiempo real
              <span class="ml-2 text-[10px] font-bold text-gray-300 uppercase tracking-widest">
                · Actualiza cada 15 min
              </span>
            </p>
            <Show when={cachedStats()}>
              <p class="text-[10px] text-gray-300">
                · Última actualización: {lastRefresh().toLocaleTimeString("es-VE")}
              </p>
            </Show>
          </div>
        </div>
        <button
          onClick={() => { refetch(); }}
          disabled={stats.loading}
          class="bg-white border-2 border-gray-100 text-gray-500 px-4 py-2 rounded-xl font-bold text-sm hover:border-colpsi-blue hover:text-colpsi-blue transition-all flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <span class={stats.loading ? "animate-spin inline-block" : ""}>↻</span>
          <span>{stats.loading && cachedStats() ? "Actualizando..." : "Actualizar"}</span>
        </button>
      </div>

      {/* ── ERROR ───────────────────────────────────────────────────────── */}
      <Show when={stats.error}>
        <div class="bg-red-50 border-l-4 border-red-500 p-4 rounded-2xl text-sm text-red-700 font-bold">
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

      {/* ── SECCIÓN: LOGINS ──────────────────────────────────────────────── */}
      <section>
        <h2 class="text-xs font-black text-gray-400 uppercase tracking-widest mb-3 pl-1">
          Inicios de Sesión
        </h2>
        <Show when={initialLoading()}>
          <AdminLoadingSkeleton variant="cards" count={4} />
        </Show>
        <Show when={cachedStats()}>
          {(s) => (
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
              <StatCard
                icon="🔑" label="Hoy"
                value={s().logins_today}
                accent="border-colpsi-yellow"
                sub={[{ label: "Únicos", value: s().unique_users_today }]}
              />
              <StatCard icon="📅" label="Esta semana"     value={s().logins_this_week}  accent="border-blue-300" />
              <StatCard icon="🗓️" label="Este mes"        value={s().logins_this_month} accent="border-blue-300" />
              <StatCard icon="📊" label="Total histórico" value={s().logins_total}       accent="border-gray-200" />
            </div>
          )}
        </Show>
      </section>

      {/* ── SECCIÓN: VISITAS ─────────────────────────────────────────────── */}
      <section>
        <h2 class="text-xs font-black text-gray-400 uppercase tracking-widest mb-3 pl-1">
          Visitas al Portal
        </h2>
        <Show when={initialLoading()}>
          <AdminLoadingSkeleton variant="cards" count={4} />
        </Show>
        <Show when={cachedStats()}>
          {(s) => (
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
              <StatCard
                icon="👁️" label="Hoy"
                value={s().page_views_today}
                accent="border-green-400"
                sub={[{ label: "Visitantes únicos", value: s().unique_visitors_today }]}
              />
              <StatCard
                icon="📈" label="Esta semana"
                value={s().page_views_this_week}
                accent="border-green-300"
                sub={[{ label: "Únicos semana", value: s().unique_visitors_week }]}
              />
              <StatCard icon="🌐" label="Total páginas vistas" value={s().page_views_total} accent="border-gray-200" />
              <StatCard
                icon="🔍" label="Búsquedas hoy"
                value={s().searches_today}
                accent="border-purple-300"
                sub={[
                  { label: "Esta semana", value: s().searches_this_week },
                  { label: "Total",       value: s().searches_total },
                ]}
              />
            </div>
          )}
        </Show>
      </section>

      {/* ── SECCIÓN: PERFILES ────────────────────────────────────────────── */}
      <section>
        <h2 class="text-xs font-black text-gray-400 uppercase tracking-widest mb-3 pl-1">
          Visitas a Perfiles
        </h2>
        <Show when={initialLoading()}>
          <AdminLoadingSkeleton variant="cards" count={3} />
        </Show>
        <Show when={cachedStats()}>
          {(s) => (
            <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
              <StatCard icon="👤" label="Hoy"          value={s().profile_views_today} accent="border-orange-300" />
              <StatCard icon="📆" label="Esta semana"  value={s().profile_views_week}  accent="border-orange-300" />
              <StatCard icon="🏆" label="Total"        value={s().profile_views_total} accent="border-gray-200"   />
            </div>
          )}
        </Show>
      </section>

      {/* ── SECCIÓN: TENDENCIAS ──────────────────────────────────────────── */}
      <section>
        <h2 class="text-xs font-black text-gray-400 uppercase tracking-widest mb-3 pl-1">
          Tendencia Diaria
        </h2>
        <Show when={initialLoading()}>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <AdminLoadingSkeleton variant="chart" />
            <AdminLoadingSkeleton variant="chart" />
          </div>
        </Show>
        <Show when={cachedStats()}>
          {(s) => (
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Sparkline data={s().login_trend ?? []} color="#1e40af" label="Logins" />
              <Sparkline data={s().view_trend  ?? []} color="#16a34a" label="Visitas al portal" />
            </div>
          )}
        </Show>
      </section>

      {/* ── SECCIÓN: RANKINGS ────────────────────────────────────────────── */}
      <section>
        <h2 class="text-xs font-black text-gray-400 uppercase tracking-widest mb-3 pl-1">
          Análisis de Búsquedas
        </h2>
        <Show when={initialLoading()}>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <AdminLoadingSkeleton variant="list" rows={6} />
            <AdminLoadingSkeleton variant="list" rows={6} />
            <AdminLoadingSkeleton variant="list" rows={6} />
          </div>
        </Show>
        <Show when={cachedStats()}>
          {(s) => (
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
              <RankingList title="Especialidades más buscadas" icon="🧠" items={s().top_specialties  ?? []} />
              <RankingList title="Municipios más buscados"     icon="📍" items={s().top_municipios   ?? []} />
              <RankingList title="Términos de búsqueda"        icon="🔤" items={s().top_search_terms ?? []} />
            </div>
          )}
        </Show>
      </section>

      {/* ── SECCIÓN: TOP PERFILES ────────────────────────────────────────── */}
      <section>
        <Show when={initialLoading()}>
          <div class="space-y-3">
            <div class="w-64 h-3 bg-gray-100 rounded animate-pulse" />
            <div class="grid grid-cols-2 md:grid-cols-5 gap-3">
              <For each={Array(5).fill(0)}>
                {() => <div class="h-24 bg-white rounded-xl border border-gray-100 animate-pulse" />}
              </For>
            </div>
          </div>
        </Show>
        <Show when={cachedStats()}>
          {(s) => <TopProfiles profiles={s().top_profiles ?? []} />}
        </Show>
      </section>

    </div>
  );
}