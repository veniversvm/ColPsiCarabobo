// web/src/components/admin/dashboard/AdminLoadingSkeleton.tsx

/**
 * AdminLoadingSkeleton — skeleton de carga para widgets del panel admin.
 *
 * Uso:
 *   <AdminLoadingSkeleton />                        // 1 bloque alto (default)
 *   <AdminLoadingSkeleton rows={4} />               // 4 filas de texto
 *   <AdminLoadingSkeleton variant="cards" count={4} />  // grilla de tarjetas
 *   <AdminLoadingSkeleton variant="chart" />        // bloque de gráfica
 *   <AdminLoadingSkeleton variant="list" rows={6} /> // lista con barras
 */

type SkeletonVariant = "card" | "cards" | "chart" | "list" | "banner";

interface AdminLoadingSkeletonProps {
  variant?: SkeletonVariant;
  /** Número de tarjetas — solo para variant="cards" */
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

      {/* ── Grilla de tarjetas métricas ───────────────────────────────── */}
      {variant() === "cards" && (
        <div class={`grid gap-4 grid-cols-2 md:grid-cols-${count()}`}>
          {Array.from({ length: count() }).map(() => (
            <div class="bg-white rounded-2xl p-5 border border-gray-100 shadow-sm space-y-3">
              <div class="flex justify-between">
                <div class="w-8 h-8 bg-gray-100 rounded-lg" />
                <div class="w-16 h-4 bg-gray-100 rounded" />
              </div>
              <div class="w-20 h-8 bg-gray-100 rounded" />
              <div class="w-24 h-3 bg-gray-100 rounded" />
              <div class="border-t border-gray-100 pt-3 flex gap-4">
                <div class="w-12 h-3 bg-gray-100 rounded" />
                <div class="w-12 h-3 bg-gray-100 rounded" />
              </div>
            </div>
          ))}
        </div>
      )}

      {/* ── Tarjeta única ─────────────────────────────────────────────── */}
      {variant() === "card" && (
        <div class="bg-white rounded-2xl p-5 border border-gray-100 shadow-sm space-y-3">
          <div class="flex justify-between">
            <div class="w-8 h-8 bg-gray-100 rounded-lg" />
            <div class="w-16 h-4 bg-gray-100 rounded" />
          </div>
          <div class="w-20 h-8 bg-gray-100 rounded" />
          <div class="w-24 h-3 bg-gray-100 rounded" />
        </div>
      )}

      {/* ── Gráfica sparkline ─────────────────────────────────────────── */}
      {variant() === "chart" && (
        <div class="bg-white rounded-2xl p-5 border border-gray-100 shadow-sm space-y-3">
          <div class="w-40 h-3 bg-gray-100 rounded" />
          <div class="w-full h-16 bg-gray-100 rounded-xl" />
          <div class="flex justify-between">
            <div class="w-12 h-2 bg-gray-100 rounded" />
            <div class="w-12 h-2 bg-gray-100 rounded" />
          </div>
        </div>
      )}

      {/* ── Lista con barras (ranking) ────────────────────────────────── */}
      {variant() === "list" && (
        <div class="bg-white rounded-2xl p-5 border border-gray-100 shadow-sm space-y-3">
          <div class="w-40 h-3 bg-gray-100 rounded mb-4" />
          {Array.from({ length: rows() }).map(() => (
            <div class="flex items-center gap-3">
              <div class="w-4 h-3 bg-gray-100 rounded flex-shrink-0" />
              <div class="flex-1 space-y-1">
                <div class="flex justify-between">
                  <div class="w-28 h-3 bg-gray-100 rounded" />
                  <div class="w-8 h-3 bg-gray-100 rounded" />
                </div>
                <div class="w-full h-1.5 bg-gray-100 rounded-full" />
              </div>
            </div>
          ))}
        </div>
      )}

      {/* ── Banner de sesiones activas ────────────────────────────────── */}
      {variant() === "banner" && (
        <div class="bg-colpsi-blue/10 rounded-3xl p-6 border border-colpsi-blue/10 flex items-center justify-between">
          <div class="space-y-2">
            <div class="w-32 h-3 bg-colpsi-blue/20 rounded" />
            <div class="w-20 h-10 bg-colpsi-blue/20 rounded" />
            <div class="w-24 h-3 bg-colpsi-blue/20 rounded" />
          </div>
          <div class="w-16 h-16 bg-colpsi-blue/10 rounded-full" />
        </div>
      )}

    </div>
  );
}