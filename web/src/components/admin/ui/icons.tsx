// web/src/components/admin/ui/icons.tsx
// Iconos SVG inline estilo Lucide para el panel admin. stroke="currentColor"
// → el color lo hereda el texto (activo/inactivo) sin depender de emojis.
import { JSX } from "solid-js";

type IconName =
  | "grid" | "users" | "fileText" | "tag" | "newspaper" | "bell" | "ticket"
  | "kanban" | "shield" | "logout" | "refresh" | "chevronRight" | "menu"
  | "panelLeft" | "sliders" | "search" | "plus" | "pencil" | "trash" | "dots";

const S = {
  rect: (x: number, y: number, w: number, h: number, rx = 2) =>
    <rect x={x} y={y} width={w} height={h} rx={rx} />,
  line: (x1: number, y1: number, x2: number, y2: number) =>
    <line x1={x1} y1={y1} x2={x2} y2={y2} />,
  circle: (cx: number, cy: number, r: number) => <circle cx={cx} cy={cy} r={r} />,
  polyline: (p: string) => <polyline points={p} />,
  path: (d: string) => <path d={d} />,
};

const icons: Record<IconName, () => JSX.Element> = {
  grid: () => (<>
    {S.rect(3, 3, 7, 7)} {S.rect(14, 3, 7, 7)} {S.rect(14, 14, 7, 7)} {S.rect(3, 14, 7, 7)}
  </>),
  users: () => (<>
    {S.path("M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2")}
    {S.circle(9, 7, 4)}
    {S.path("M22 21v-2a4 4 0 0 0-3-3.87")}
    {S.path("M16 3.13a4 4 0 0 1 0 7.75")}
  </>),
  fileText: () => (<>
    {S.path("M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z")}
    {S.polyline("14 2 14 8 20 8")}
    {S.line(16, 13, 8, 13)} {S.line(16, 17, 8, 17)} {S.polyline("10 9 9 9 8 9")}
  </>),
  tag: () => (<>
    {S.path("M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z")}
    {S.line(7, 7, 7.01, 7)}
  </>),
  newspaper: () => (<>
    {S.path("M4 22h16a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2H8a2 2 0 0 0-2 2v16a2 2 0 0 1-2 2Zm0 0a2 2 0 0 1-2-2v-9c0-1.1.9-2 2-2h2")}
    {S.line(18, 14, 8, 14)} {S.line(15, 18, 10, 18)} {S.path("M10 6h8v4h-8V6z")}
  </>),
  bell: () => (<>
    {S.path("M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9")}
    {S.path("M10.3 21a1.94 1.94 0 0 0 3.4 0")}
  </>),
  ticket: () => (<>
    {S.path("M2 9a3 3 0 0 1 0 6v2a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-2a3 3 0 0 1 0-6V7a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2Z")}
    {S.line(13, 5, 13, 7)} {S.line(13, 17, 13, 19)} {S.line(13, 11, 13, 13)}
  </>),
  kanban: () => (<>
    {S.rect(3, 3, 18, 18)}
    {S.line(8, 7, 8, 14)} {S.line(12, 7, 12, 11)} {S.line(16, 7, 16, 16)}
  </>),
  shield: () => (
    <>{S.path("M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z")}</>
  ),
  logout: () => (<>
    {S.path("M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4")}
    {S.polyline("16 17 21 12 16 7")}
    {S.line(21, 12, 9, 12)}
  </>),
  refresh: () => (<>
    {S.path("M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8")}
    {S.path("M21 3v5h-5")}
    {S.path("M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16")}
    {S.path("M8 16H3v5")}
  </>),
  chevronRight: () => <>{S.polyline("9 18 15 12 9 6")}</>,
  menu: () => (<>
    {S.line(4, 6, 20, 6)} {S.line(4, 12, 20, 12)} {S.line(4, 18, 20, 18)}
  </>),
  panelLeft: () => (<>
    {S.rect(3, 3, 18, 18)}
    {S.line(9, 3, 9, 21)}
  </>),
  sliders: () => (<>
    {S.line(4, 21, 4, 14)} {S.line(4, 10, 4, 3)}
    {S.line(12, 21, 12, 12)} {S.line(12, 8, 12, 3)}
    {S.line(20, 21, 20, 16)} {S.line(20, 12, 20, 3)}
    {S.line(1, 14, 7, 14)} {S.line(9, 8, 15, 8)} {S.line(17, 16, 23, 16)}
  </>),
  search: () => (<>
    {S.circle(11, 11, 8)}
    {S.line(21, 21, 16.65, 16.65)}
  </>),
  plus: () => (<>
    {S.line(5, 12, 19, 12)} {S.line(12, 5, 12, 19)}
  </>),
  pencil: () => (<>
    {S.path("M21.174 6.812a1 1 0 0 0-3.986-3.987L3.842 16.174a2 2 0 0 0-.5.83l-1.321 4.352a.5.5 0 0 0 .623.622l4.353-1.32a2 2 0 0 0 .83-.497z")}
    {S.path("m15 5 4 4")}
  </>),
  trash: () => (<>
    {S.line(3, 6, 21, 6)}
    {S.path("M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6")}
    {S.path("M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2")}
    {S.line(10, 11, 10, 17)} {S.line(14, 11, 14, 17)}
  </>),
  dots: () => (<>
    {S.circle(12, 12, 1)} {S.circle(19, 12, 1)} {S.circle(5, 12, 1)}
  </>),
};

export function Icon(props: { name: IconName; class?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
      class={`w-4 h-4 shrink-0 ${props.class ?? ""}`}
      aria-hidden="true"
    >
      {icons[props.name]()}
    </svg>
  );
}