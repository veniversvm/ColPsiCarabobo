// web/src/components/admin/psicologos/edit/EditPrimitives.tsx

import { Show, createSignal } from "solid-js";

export const IC  = "w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none transition-all text-gray-800 text-sm";
export const IC2 = "w-full bg-colpsi-surface border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none transition-all text-gray-800 text-sm";

export function Field(props: { label: string; children: any }) {
  return (
    <div>
      <label class="block text-[10px] font-black text-gray-500 uppercase tracking-widest ml-1 mb-1">
        {props.label}
      </label>
      {props.children}
    </div>
  );
}

export function SectionCard(props: { title: string; accent?: string; children: any }) {
  const accent = props.accent ?? "border-colpsi-yellow";
  return (
    <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-colpsi-border">
      <h2 class={`text-lg font-black text-blue-800 mb-6 border-l-4 ${accent} pl-3`}>
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
    <section class="bg-white rounded-3xl shadow-sm border border-colpsi-border overflow-hidden">
      <button
        type="button"
        onClick={() => setOpen(!open())}
        aria-expanded={open()}
        class="w-full flex items-center justify-between gap-4 px-6 md:px-8 py-5 text-left hover:bg-colpsi-surface/60 transition-colors group"
      >
        <h2 class={`text-lg font-black text-blue-800 border-l-4 ${accent} pl-3`}>
          {props.title}
        </h2>
        <span class={`text-gray-400 transition-transform duration-300 shrink-0 ${open() ? "rotate-180" : ""}`}>
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
          </svg>
        </span>
      </button>
      <Show when={open()}>
        <div class="px-6 md:px-8 pb-6 md:pb-8">{props.children}</div>
      </Show>
    </section>
  );
}