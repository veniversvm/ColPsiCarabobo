// web/src/routes/admin/layout.tsx
// web/src/routes/admin/index.tsx
import { createResource, createSignal, For, Show, onMount, onCleanup, createEffect } from "solid-js";
import { apiGet } from "~/lib/api";

// ── Tipos ─────────────────────────────────────────────────────────────────────
interface TopItem  { value: string; count: number; name: string }
interface DailyCount { date: string; count: number }
interface TopProfile {
    psi_id:     string
    first_name: string
    last_name:  string
    fpv:        number
    count:      number
}

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

// ── Helpers ───────────────────────────────────────────────────────────────────
const fmt = (n?: number) => (n ?? 0).toLocaleString("es-VE")

function sparkMax(data: DailyCount[]) {
  return Math.max(...data.map(d => d.count), 1)
}

// ── Sub-componentes ───────────────────────────────────────────────────────────

/** Tarjeta de métrica con valor principal y desglose */
function StatCard(props: {
  icon: string
  label: string
  value: number
  sub?: { label: string; value: number }[]
  accent?: string
  pulse?: boolean
}) {
  const accent = props.accent ?? "border-colpsi-yellow"
  return (
    <div class={`bg-white rounded-2xl p-5 shadow-sm border border-gray-100 border-l-4 ${accent} flex flex-col gap-3`}>
      <div class="flex items-center justify-between">
        <span class="text-2xl">{props.icon}</span>
        <Show when={props.pulse}>
          <span class="flex items-center gap-1.5 text-[10px] font-black text-green-600 uppercase tracking-widest">
            <span class="w-2 h-2 bg-green-500 rounded-full animate-pulse inline-block" />
            En vivo
          </span>
        </Show>
      </div>
      <div>
        <p class="text-3xl font-black text-colpsi-blue tabular-nums">{fmt(props.value)}</p>
        <p class="text-xs font-bold text-gray-400 uppercase tracking-widest mt-0.5">{props.label}</p>
      </div>
      <Show when={props.sub && props.sub.length > 0}>
        <div class="flex flex-wrap gap-3 border-t border-gray-100 pt-3">
          <For each={props.sub}>
            {(s) => (
              <div>
                <p class="text-sm font-black text-gray-700 tabular-nums">{fmt(s.value)}</p>
                <p class="text-[10px] text-gray-400 uppercase tracking-widest">{s.label}</p>
              </div>
            )}
          </For>
        </div>
      </Show>
    </div>
  )
}

/** Sparkline SVG simple — sin dependencias */
function Sparkline(props: { data: DailyCount[]; color: string; label: string }) {
  const max = () => sparkMax(props.data)
  const W = 320; const H = 60; const pad = 4

  const points = () =>
    props.data.map((d, i) => {
      const x = pad + (i / Math.max(props.data.length - 1, 1)) * (W - pad * 2)
      const y = H - pad - ((d.count / max()) * (H - pad * 2))
      return `${x},${y}`
    }).join(" ")

  const area = () => {
    if (props.data.length === 0) return ""
    const pts = props.data.map((d, i) => {
      const x = pad + (i / Math.max(props.data.length - 1, 1)) * (W - pad * 2)
      const y = H - pad - ((d.count / max()) * (H - pad * 2))
      return `${x},${y}`
    })
    return `${pad},${H - pad} ${pts.join(" ")} ${W - pad},${H - pad}`
  }

  return (
    <div class="bg-white rounded-2xl p-5 shadow-sm border border-gray-100">
      <p class="text-xs font-black text-gray-400 uppercase tracking-widest mb-3">{props.label} — últimos 14 días</p>
      <Show when={props.data.length > 0} fallback={
        <div class="h-16 flex items-center justify-center text-xs text-gray-300 italic">Sin datos aún</div>
      }>
        <svg viewBox={`0 0 ${W} ${H}`} class="w-full" style="height:64px">
          <defs>
            <linearGradient id={`grad-${props.color}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color={props.color} stop-opacity="0.25" />
              <stop offset="100%" stop-color={props.color} stop-opacity="0.02" />
            </linearGradient>
          </defs>
          <polygon points={area()} fill={`url(#grad-${props.color})`} />
          <polyline
            points={points()}
            fill="none"
            stroke={props.color}
            stroke-width="2"
            stroke-linejoin="round"
            stroke-linecap="round"
          />
          {/* Punto final destacado */}
          <Show when={props.data.length > 0}>
            {(() => {
              const last = props.data[props.data.length - 1]
              const i = props.data.length - 1
              const x = pad + (i / Math.max(props.data.length - 1, 1)) * (W - pad * 2)
              const y = H - pad - ((last.count / max()) * (H - pad * 2))
              return <circle cx={x} cy={y} r="4" fill={props.color} />
            })()}
          </Show>
        </svg>
        {/* Eje X: primer y último día */}
        <div class="flex justify-between mt-1">
          <span class="text-[9px] text-gray-300">{props.data[0]?.date}</span>
          <span class="text-[9px] text-gray-300">{props.data[props.data.length - 1]?.date}</span>
        </div>
      </Show>
    </div>
  )
}

/** Ranking de items con barra proporcional */
function RankingList(props: { title: string; icon: string; items: TopItem[] }) {
  const max = () => Math.max(...(props.items ?? []).map(i => i.count), 1)
  return (
    <div class="bg-white rounded-2xl p-5 shadow-sm border border-gray-100">
      <p class="text-xs font-black text-gray-400 uppercase tracking-widest mb-4 flex items-center gap-2">
        <span>{props.icon}</span>{props.title}
        <span class="ml-auto text-gray-300">últimos 30 días</span>
      </p>
      <Show when={props.items?.length > 0} fallback={
        <p class="text-xs text-gray-300 italic text-center py-4">Sin datos aún</p>
      }>
        <div class="space-y-2.5">
          <For each={props.items?.slice(0, 8)}>
            {(item, i) => (
              <div class="flex items-center gap-3">
                <span class="text-[10px] font-black text-gray-300 w-4 text-right">{i() + 1}</span>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center justify-between mb-0.5">
                    <span class="text-xs font-bold text-gray-700 truncate max-w-[160px]">
                      {item?.name || item.value || "—"}
                    </span>
                    <span class="text-xs font-black text-colpsi-blue tabular-nums ml-2">{fmt(item.count)}</span>
                  </div>
                  <div class="h-1.5 bg-gray-100 rounded-full overflow-hidden">
                    <div
                      class="h-full bg-colpsi-blue/60 rounded-full transition-all duration-700"
                      style={{ width: `${(item.count / max()) * 100}%` }}
                    />
                  </div>
                </div>
              </div>
            )}
          </For>
        </div>
      </Show>
    </div>
  )
}

// ── Página principal ──────────────────────────────────────────────────────────
export default function AdminDashboard() {
  const [stats, { refetch }] = createResource<DashboardStats>(() =>
    apiGet("/admin/dashboard/stats")
  )


  createEffect(() => {
    if (stats) {
      console.log(stats())
    }
  })


  // Auto-refresh cada 60 segundos para las sesiones activas
  let interval: ReturnType<typeof setInterval>
  onMount(() => { interval = setInterval(refetch, 60_000 * 15) })
  onCleanup(() => clearInterval(interval))

  const [lastRefresh, setLastRefresh] = createSignal(new Date())
  createResource(() => stats() && setLastRefresh(new Date()))

  return (
    <div class="space-y-8 animate-in fade-in duration-500 pb-24">

      {/* ── HEADER ──────────────────────────────────────────────────────── */}
      <div class="flex items-start justify-between">
        <div>
          <h1 class="text-2xl font-black text-colpsi-blue">Panel de Control</h1>
          <p class="text-gray-400 text-sm mt-1">
            Métricas del portal en tiempo real
            <span class="ml-2 text-[10px] font-bold text-gray-300 uppercase tracking-widest">
              · Actualiza cada 60s
            </span>
          </p>
        </div>
        <button
          onClick={() => { refetch(); setLastRefresh(new Date()) }}
          class="bg-white border-2 border-gray-100 text-gray-500 px-4 py-2 rounded-xl font-bold text-sm hover:border-colpsi-blue hover:text-colpsi-blue transition-all flex items-center gap-2"
        >
          <span class={stats.loading ? "animate-spin inline-block" : ""}>↻</span>
          Actualizar
        </button>
      </div>

      <Show when={stats.error}>
        <div class="bg-red-50 border-l-4 border-red-500 p-4 rounded-2xl text-sm text-red-700 font-bold">
          ⚠️ Error al cargar estadísticas — verifica que el servidor esté activo.
        </div>
      </Show>

      <Show when={stats.loading && !stats()}>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <For each={Array(8).fill(0)}>
            {() => <div class="h-32 bg-white rounded-2xl animate-pulse border border-gray-100" />}
          </For>
        </div>
      </Show>

      <Show when={stats()}>
        {(s) => (
          <div class="space-y-8">

            {/* ── SESIONES ACTIVAS (destacado) ────────────────────────────── */}
            <div class="bg-colpsi-blue rounded-3xl p-6 flex items-center justify-between shadow-lg">
              <div>
                <p class="text-blue-200 text-xs font-black uppercase tracking-widest">En línea ahora mismo</p>
                <p class="text-white text-5xl font-black tabular-nums mt-1">{fmt(s().active_sessions_now)}</p>
                <p class="text-blue-300 text-sm mt-1">sesiones activas</p>
              </div>
              <div class="text-7xl opacity-20">👥</div>
            </div>

            {/* ── CONTADORES: LOGINS ──────────────────────────────────────── */}
            <div>
              <h2 class="text-xs font-black text-gray-400 uppercase tracking-widest mb-3 pl-1">
                Inicios de Sesión
              </h2>
              <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                <StatCard
                  icon="🔑"
                  label="Hoy"
                  value={s().logins_today}
                  accent="border-colpsi-yellow"
                  sub={[{ label: "Únicos", value: s().unique_users_today }]}
                />
                <StatCard icon="📅" label="Esta semana"  value={s().logins_this_week}  accent="border-blue-300" />
                <StatCard icon="🗓️" label="Este mes"     value={s().logins_this_month} accent="border-blue-300" />
                <StatCard icon="📊" label="Total histórico" value={s().logins_total}   accent="border-gray-200" />
              </div>
            </div>

            {/* ── CONTADORES: VISITAS ─────────────────────────────────────── */}
            <div>
              <h2 class="text-xs font-black text-gray-400 uppercase tracking-widest mb-3 pl-1">
                Visitas al Portal
              </h2>
              <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                <StatCard
                  icon="👁️"
                  label="Hoy"
                  value={s().page_views_today}
                  accent="border-green-400"
                  sub={[{ label: "Visitantes únicos", value: s().unique_visitors_today }]}
                />
                <StatCard
                  icon="📈"
                  label="Esta semana"
                  value={s().page_views_this_week}
                  accent="border-green-300"
                  sub={[{ label: "Únicos semana", value: s().unique_visitors_week }]}
                />
                <StatCard icon="🌐" label="Total páginas vistas" value={s().page_views_total} accent="border-gray-200" />
                <StatCard
                  icon="🔍"
                  label="Búsquedas hoy"
                  value={s().searches_today}
                  accent="border-purple-300"
                  sub={[
                    { label: "Esta semana", value: s().searches_this_week },
                    { label: "Total",       value: s().searches_total },
                  ]}
                />
              </div>
            </div>

            {/* ── CONTADORES: PERFILES ────────────────────────────────────── */}
            <div>
              <h2 class="text-xs font-black text-gray-400 uppercase tracking-widest mb-3 pl-1">
                Visitas a Perfiles
              </h2>
              <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
                <StatCard icon="👤" label="Hoy"          value={s().profile_views_today} accent="border-orange-300" />
                <StatCard icon="📆" label="Esta semana"  value={s().profile_views_week}  accent="border-orange-300" />
                <StatCard icon="🏆" label="Total"        value={s().profile_views_total} accent="border-gray-200"   />
              </div>
            </div>

            {/* ── TENDENCIAS (sparklines) ──────────────────────────────────── */}
            <div>
              <h2 class="text-xs font-black text-gray-400 uppercase tracking-widest mb-3 pl-1">
                Tendencia Diaria
              </h2>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Sparkline
                  data={s().login_trend ?? []}
                  color="#1e40af"
                  label="Logins"
                />
                <Sparkline
                  data={s().view_trend ?? []}
                  color="#16a34a"
                  label="Visitas al portal"
                />
              </div>
            </div>

            {/* ── TOPS ────────────────────────────────────────────────────── */}
            <div>
              <h2 class="text-xs font-black text-gray-400 uppercase tracking-widest mb-3 pl-1">
                Análisis de Búsquedas
              </h2>
              <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                <RankingList
                  title="Especialidades más buscadas"
                  icon="🧠"
                  items={s().top_specialties ?? []}
                />
                <RankingList
                  title="Municipios más buscados"
                  icon="📍"
                  items={s().top_municipios ?? []}
                />
                <RankingList
                  title="Términos de búsqueda"
                  icon="🔤"
                  items={s().top_search_terms ?? []}
                />
              </div>
            </div>

            {/* ── TOP PERFILES ────────────────────────────────────────────── */}
            <Show when={s().top_profiles?.length > 0}>
              <div>
                <h2 class="text-xs font-black text-gray-400 uppercase tracking-widest mb-3 pl-1">
                  Perfiles más Visitados (últimos 30 días)
                </h2>
                <div class="bg-white rounded-2xl p-5 shadow-sm border border-gray-100">
                  <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-5 gap-3">
                    <For each={s().top_profiles?.slice(0, 10)}>
                      {(p, i) => (
                        <div class="bg-gray-50 rounded-xl p-3 border border-gray-100 flex flex-col gap-1">
                          <span class="text-[9px] font-black text-gray-300 uppercase">#{i() + 1}</span>
                          <span class="text-lg font-black text-colpsi-blue tabular-nums">{fmt(p.count)}</span>
                          <span class="text-xs font-bold text-gray-700 truncate">
                              {p.first_name} {p.last_name}
                          </span>
                          <span class="text-[9px] text-gray-400">FPV {p.fpv} · {fmt(p.count)} visitas</span>
                        </div>
                      )}
                    </For>
                  </div>
                </div>
              </div>
            </Show>

          </div>
        )}
      </Show>

    </div>
  )
}