// web/src/components/admin/dashboard/ActiveSessionsBanner.tsx

interface ActiveSessionsBannerProps {
  count: number;
}

const fmt = (n?: number) => (n ?? 0).toLocaleString("es-VE");

export function ActiveSessionsBanner(props: ActiveSessionsBannerProps) {
  return (
    <div class="bg-colpsi-blue rounded-3xl p-6 flex items-center justify-between shadow-lg">
      <div>
        <p class="text-blue-200 text-xs font-black uppercase tracking-widest">En línea ahora mismo</p>
        <p class="text-white text-5xl font-black tabular-nums mt-1">{fmt(props.count)}</p>
        <p class="text-blue-300 text-sm mt-1">sesiones activas</p>
      </div>
      <div class="text-7xl opacity-20">👥</div>
    </div>
  );
}