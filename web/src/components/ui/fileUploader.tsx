import { Show } from "solid-js";

export const FileUploader = (props: { id: string, label: string, onChange: (e: Event) => void, file?: File }) => {
  return (
    <label class="flex flex-col items-center justify-center border-2 border-dashed border-gray-200 rounded-2xl h-28 hover:border-colpsi-blue transition-all cursor-pointer bg-gray-50/50">
      <input type="file" accept="image/*" class="sr-only" onChange={props.onChange} />
      <span class="text-[10px] font-black text-colpsi-blue uppercase mb-1">{props.label}</span>
      <Show when={props.file} fallback={<span class="text-xl">📷</span>}>
        <span class="text-green-600 text-xs font-bold">✓</span>
      </Show>
    </label>
  );
}