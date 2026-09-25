// web/src/components/admin/dashboard/AdminLoadingSkeleton.tsx

/**
 * AdminLoadingSkeleton — skeleton de carga para widgets del panel admin.
 *
 * Uso:
 *   <AdminLoadingSkeleton />                        // 1 bloque alto (default)
 *   <AdminLoadingSkeleton rows={4} />               // 4 filas de texto
 *   <AdminLoadingSkeleton variant="kpi" count={4} /> // faja de celdas métricas
 *   <AdminLoadingSkeleton variant="cards" count={4} />  // grilla de tarjetas
 *   <AdminLoadingSkeleton variant="chart" />        // bloque de gráfica
 *   <AdminLoadingSkeleton variant="list" rows={6} /> // lista con barras
 */

type SkeletonVariant = "card" | "cards" | "kpi" | "chart" | "list" | "banner";

interface AdminLoadingSkeletonProps {
  variant?: SkeletonVariant;
  /** Número de tarjetas/celdas — para variant="cards" o "kpi" */
  count?: number;
  /** Número de filas — para variant="list" */
  rows?: number;
  /** Clase extra para el contenedor */
  class?: string;
}

export function AdminLoadingSkeleton(props: AdminLoadingSkeletonProps) {
  const variant = () => props.variant ?? "card";
  const count   = () => props.count  ?? 4;
  const rows    = () => props.rows   ?? 5;

  return (
    <div class={`animate-pulse ${props.class ?? ""}`}>

      {/* ── Faja de celdas métricas (sustituye a cards) ─────────────────── */}
      {variant() === "kpi" && (
        <div class="grid grid-cols-2 lg:grid-cols-4 border-t border-l border-colpsi-border">
          {Array.from({ length: count() }).map(() => (
            <div class="border-b border-r border-colpsi-border bg-white p-4 space-y-2.5">
              <div class="w-20 h-2.5 bg-slate-100 rounded" />
              <div class="w-14 h-6 bg-slate-100 rounded" />
              <div class="w-24 h-2 bg-slate-100 rounded" />
            </div>
          ))}
        </div>
      )}

      {/* ── Grilla de tarjetas ──────────────────────────────────────────── */}
      {variant() === "cards" && (
        <div class="grid gap-3 grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: count() }).map(() => (
            <div class="rounded-lg border border-colpsi-border bg-white p-4 space-y-2.5">
              <div class="w-20 h-2.5 bg-slate-100 rounded" />
              <div class="w-14 h-6 bg-slate-100 rounded" />
              <div class="w-24 h-2 bg-slate-100 rounded" />
            </div>
          ))}
        </div>
      )}

      {/* ── Tarjeta única ───────────────────────────────────────────────── */}
      {variant() === "card" && (
        <div class="rounded-lg border border-colpsi-border bg-white p-4 space-y-2.5">
          <div class="w-20 h-2.5 bg-slate-100 rounded" />
          <div class="w-14 h-6 bg-slate-100 rounded" />
          <div class="w-24 h-2 bg-slate-100 rounded" />
        </div>
      )}

      {/* ── Gráfica sparkline ───────────────────────────────────────────── */}
      {variant() === "chart" && (
        <div class="rounded-lg border border-colpsi-border bg-white p-4 space-y-3">
          <div class="w-40 h-2.5 bg-slate-100 rounded" />
          <div class="w-full h-16 bg-slate-100 rounded-md" />
          <div class="flex justify-between">
            <div class="w-12 h-2 bg-slate-100 rounded" />
            <div class="w-12 h-2 bg-slate-100 rounded" />
          </div>
        </div>
      )}

      {/* ── Lista con barras (ranking) ──────────────────────────────────── */}
      {variant() === "list" && (
        <div class="rounded-lg border border-colpsi-border bg-white p-4 space-y-3">
          <div class="w-40 h-2.5 bg-slate-100 rounded mb-2" />
          {Array.from({ length: rows() }).map(() => (
            <div class="flex items-center gap-3">
              <div class="w-4 h-2.5 bg-slate-100 rounded flex-shrink-0" />
              <div class="flex-1 space-y-1">
                <div class="flex justify-between">
                  <div class="w-28 h-2.5 bg-slate-100 rounded" />
                  <div class="w-8 h-2.5 bg-slate-100 rounded" />
                </div>
                <div class="w-full h-1.5 bg-slate-100 rounded-full" />
              </div>
            </div>
          ))}
        </div>
      )}

      {/* ── Banner de sesiones activas ──────────────────────────────────── */}
      {variant() === "banner" && (
        <div class="rounded-lg border border-colpsi-border bg-white px-4 py-3.5 h-16 flex items-center">
          <div class="flex items-center gap-3 w-full">
            <div class="w-2.5 h-2.5 rounded-full bg-slate-200" />
            <div class="space-y-2 flex-1">
              <div class="w-32 h-2.5 bg-slate-100 rounded" />
              <div class="w-20 h-4 bg-slate-100 rounded" />
            </div>
          </div>
        </div>
      )}

    </div>
  );
}