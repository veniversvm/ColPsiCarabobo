import { Show } from "solid-js";
import { Icon } from "~/components/admin/ui/icons";

export const FileUploader = (props: { id: string, label: string, onChange: (e: Event) => void, file?: File }) => {
  return (
    <label class="flex flex-col items-center justify-center border border-dashed border-colpsi-border rounded-md h-24 hover:border-colpsi-blue transition-colors cursor-pointer bg-colpsi-bg/40">
      <input type="file" accept="image/*" class="sr-only" onChange={props.onChange} />
      <span class="text-[10px] font-semibold text-colpsi-blue uppercase tracking-wide mb-1">{props.label}</span>
      <Show when={props.file} fallback={<Icon name="camera" class="w-4 h-4 text-colpsi-muted" />}>
        <Icon name="check" class="w-4 h-4 text-emerald-600" />
      </Show>
    </label>
  );
}