// web/src/components/admin/dashboard/Sparkline.tsx
import { Show } from "solid-js";

interface DailyCount { date: string; count: number }

interface SparklineProps {
  data: DailyCount[];
  color: string;
  label: string;
}

function sparkMax(data: DailyCount[]) {
  return Math.max(...data.map(d => d.count), 1);
}

export function Sparkline(props: SparklineProps) {
  const max = () => sparkMax(props.data);
  const W = 320; const H = 60; const pad = 4;

  const points = () =>
    props.data.map((d, i) => {
      const x = pad + (i / Math.max(props.data.length - 1, 1)) * (W - pad * 2);
      const y = H - pad - ((d.count / max()) * (H - pad * 2));
      return `${x},${y}`;
    }).join(" ");

  const area = () => {
    if (props.data.length === 0) return "";
    const pts = props.data.map((d, i) => {
      const x = pad + (i / Math.max(props.data.length - 1, 1)) * (W - pad * 2);
      const y = H - pad - ((d.count / max()) * (H - pad * 2));
      return `${x},${y}`;
    });
    return `${pad},${H - pad} ${pts.join(" ")} ${W - pad},${H - pad}`;
  };

  return (
    <div class="bg-white rounded-2xl p-5 shadow-sm border border-colpsi-border">
      <p class="text-xs font-black text-gray-400 uppercase tracking-widest mb-3">
        {props.label} — últimos 14 días
      </p>
      <Show
        when={props.data.length > 0}
        fallback={
          <div class="h-16 flex items-center justify-center text-xs text-gray-300 italic">
            Sin datos aún
          </div>
        }
      >
        <svg viewBox={`0 0 ${W} ${H}`} class="w-full" style="height:64px">
          <defs>
            <linearGradient id={`grad-${props.color.replace("#", "")}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%"   stop-color={props.color} stop-opacity="0.25" />
              <stop offset="100%" stop-color={props.color} stop-opacity="0.02" />
            </linearGradient>
          </defs>
          <polygon points={area()} fill={`url(#grad-${props.color.replace("#", "")})`} />
          <polyline
            points={points()}
            fill="none"
            stroke={props.color}
            stroke-width="2"
            stroke-linejoin="round"
            stroke-linecap="round"
          />
          <Show when={props.data.length > 0}>
            {(() => {
              const last = props.data[props.data.length - 1];
              const i    = props.data.length - 1;
              const x    = pad + (i / Math.max(props.data.length - 1, 1)) * (W - pad * 2);
              const y    = H - pad - ((last.count / max()) * (H - pad * 2));
              return <circle cx={x} cy={y} r="4" fill={props.color} />;
            })()}
          </Show>
        </svg>
        <div class="flex justify-between mt-1">
          <span class="text-[9px] text-gray-300">{props.data[0]?.date}</span>
          <span class="text-[9px] text-gray-300">{props.data[props.data.length - 1]?.date}</span>
        </div>
      </Show>
    </div>
  );
}