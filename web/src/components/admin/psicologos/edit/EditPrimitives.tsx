// web/src/components/admin/psicologos/edit/EditPrimitives.tsx

export const IC  = "w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none transition-all text-gray-800 text-sm";
export const IC2 = "w-full bg-gray-50 border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none transition-all text-gray-800 text-sm";

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
    <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100">
      <h2 class={`text-lg font-black text-blue-800 mb-6 border-l-4 ${accent} pl-3`}>
        {props.title}
      </h2>
      {props.children}
    </section>
  );
}