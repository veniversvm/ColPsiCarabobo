// web/src/components/admin/noticias/edit/EditImageSection.tsx
import { Show } from "solid-js";
import { Accessor, Setter } from "solid-js";
import { imgUrl } from "./types";
import { Icon } from "~/components/admin/ui/icons";

interface Props {
  currentImageUrl: string | null;
  imageFile: Accessor<File | null>;
  imagePreview: Accessor<string | null>;
  onImageChange: (e: Event) => void;
  onClearImage: () => void;
}

export function EditImageSection(props: Props) {
  return (
    <section class="bg-white rounded-lg p-5 border border-colpsi-border">
      <h2 class="text-base font-semibold text-colpsi-text border-b border-colpsi-border pb-3 mb-4">
        Imagen de Portada
      </h2>

      <Show
        when={props.currentImageUrl}
        fallback={
          <label class="flex flex-col items-center justify-center w-full h-44 border-2 border-dashed border-slate-300 rounded-lg bg-colpsi-bg hover:bg-white hover:border-colpsi-blue/40 transition-colors cursor-pointer group">
            <div class="flex flex-col items-center gap-2 text-colpsi-muted group-hover:text-colpsi-blue transition-colors">
              <svg class="w-10 h-10" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.409a2.25 2.25 0 013.182 0l2.909 2.909M3 9.75h.008M3.375 3h17.25A.375.375 0 0121 3.375v17.25A.375.375 0 0120.625 21H3.375A.375.375 0 013 20.625V3.375A.375.375 0 013.375 3z" />
              </svg>
              <span class="font-semibold text-sm">Haz clic para subir imagen</span>
              <span class="text-[11px]">JPG, PNG, WebP · Máx. 5MB</span>
            </div>
            <input type="file" accept="image/*" class="hidden" onChange={props.onImageChange} />
          </label>
        }
      >
        <div class="relative group rounded-lg overflow-hidden border border-colpsi-border">
          <img src={props.currentImageUrl!} alt="Vista previa" class="w-full max-h-64 object-contain bg-colpsi-bg" />
          <div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
            <label class="inline-flex items-center gap-1.5 h-9 px-3 bg-white text-colpsi-blue font-semibold rounded-md text-xs border border-colpsi-border shadow-sm hover:bg-colpsi-bg transition-colors cursor-pointer">
              <Icon name="refresh" class="w-3.5 h-3.5" />
              Cambiar imagen
              <input type="file" accept="image/*" class="hidden" onChange={props.onImageChange} />
            </label>
            <Show when={props.imagePreview()}>
              <button
                type="button"
                onClick={props.onClearImage}
                class="inline-flex items-center gap-1.5 h-9 px-3 bg-white text-colpsi-red font-semibold rounded-md text-xs border border-colpsi-border shadow-sm hover:bg-red-50 transition-colors"
              >
                <Icon name="refresh" class="w-3.5 h-3.5" />
                Restaurar original
              </button>
            </Show>
          </div>
          <div class="absolute bottom-2 left-2 bg-black/60 text-white text-[10px] font-semibold px-2 py-1 rounded-md">
            {props.imageFile()?.name ?? "Imagen actual"}
          </div>
        </div>
      </Show>
    </section>
  );
}