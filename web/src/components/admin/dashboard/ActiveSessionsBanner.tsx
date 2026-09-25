// web/src/components/admin/dashboard/ActiveSessionsBanner.tsx
import { Badge } from "~/components/admin/ui/Badge";

interface ActiveSessionsBannerProps {
  count: number;
}

const fmt = (n?: number) => (n ?? 0).toLocaleString("es-VE");

export function ActiveSessionsBanner(props: ActiveSessionsBannerProps) {
  return (
    <div class="border border-colpsi-border rounded-lg bg-white px-4 py-3.5 flex items-center justify-between gap-4">
      <div class="flex items-center gap-3 min-w-0">
        <span class="relative flex w-2.5 h-2.5 shrink-0">
          <span class="absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75 animate-ping" />
          <span class="relative inline-flex rounded-full w-2.5 h-2.5 bg-emerald-500" />
        </span>
        <div class="min-w-0">
          <p class="text-[11px] font-semibold uppercase tracking-wide text-colpsi-muted">
            En línea ahora mismo
          </p>
          <p class="text-xl font-semibold text-colpsi-text tabular-nums leading-tight">
            {fmt(props.count)}{" "}
            <span class="text-sm font-normal text-colpsi-muted">sesiones activas</span>
          </p>
        </div>
      </div>
      <Badge tone={props.count > 0 ? "success" : "neutral"} class="shrink-0">
        {props.count > 0 ? "Operativo" : "Sin actividad"}
      </Badge>
    </div>
  );
}