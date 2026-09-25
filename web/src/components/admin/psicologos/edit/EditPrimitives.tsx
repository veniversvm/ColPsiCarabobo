// web/src/components/admin/psicologos/edit/EditPrimitives.tsx

import { Show, createSignal } from "solid-js";

export const IC  = "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text";
export const IC2 = "h-9 w-full rounded-md border border-slate-300 bg-colpsi-bg px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text";

export function Field(props: { label: string; children: any }) {
  return (
    <div>
      <label class="block text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1 mb-1">
        {props.label}
      </label>
      {props.children}
    </div>
  );
}

export function SectionCard(props: { title: string; accent?: string; children: any }) {
  const accent = props.accent ?? "border-colpsi-yellow";
  return (
    <section class="bg-white rounded-lg p-5 border border-colpsi-border">
      <h2 class="text-base font-semibold text-colpsi-text border-b border-colpsi-border pb-3 mb-4">
        {props.title}
      </h2>
      {props.children}
    </section>
  );
}

export function CollapsibleSection(props: {
  title: string;
  accent?: string;
  defaultOpen?: boolean;
  children: any;
}) {
  const accent = props.accent ?? "border-colpsi-yellow";
  const [open, setOpen] = createSignal(props.defaultOpen ?? false);
  return (
    <section class="bg-white rounded-lg border border-colpsi-border overflow-hidden">
      <button
        type="button"
        onClick={() => setOpen(!open())}
        aria-expanded={open()}
        class="w-full flex items-center justify-between gap-4 px-5 py-4 text-left hover:bg-colpsi-bg/60 transition-colors group"
      >
        <h2 class="text-base font-semibold text-colpsi-text">
          {props.title}
        </h2>
        <span class={`text-colpsi-muted transition-transform duration-300 shrink-0 ${open() ? "rotate-180" : ""}`}>
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
          </svg>
        </span>
      </button>
      <Show when={open()}>
        <div class="px-5 pb-5">{props.children}</div>
      </Show>
    </section>
  );
}