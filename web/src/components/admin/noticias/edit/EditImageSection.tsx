// web/src/components/admin/noticias/edit/EditImageSection.tsx
import { Show } from "solid-js";
import { Accessor, Setter } from "solid-js";
import { imgUrl } from "./types";

interface Props {
  currentImageUrl: string | null;
  imageFile: Accessor<File | null>;
  imagePreview: Accessor<string | null>;
  onImageChange: (e: Event) => void;
  onClearImage: () => void;
}

export function EditImageSection(props: Props) {
  return (
    <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100">
      <h2 class="text-sm font-black text-blue-800 uppercase tracking-widest border-b border-gray-100 pb-3 mb-5">
        Imagen de Portada
      </h2>

      <Show
        when={props.currentImageUrl}
        fallback={
          <label class="flex flex-col items-center justify-center w-full h-44 border-2 border-dashed border-gray-300 rounded-2xl bg-gray-50 hover:bg-blue-50 hover:border-blue-300 transition-all cursor-pointer group">
            <div class="flex flex-col items-center gap-2 text-gray-400 group-hover:text-blue-500 transition-colors">
              <svg class="w-10 h-10" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.409a2.25 2.25 0 013.182 0l2.909 2.909M3 9.75h.008M3.375 3h17.25A.375.375 0 0121 3.375v17.25A.375.375 0 0120.625 21H3.375A.375.375 0 013 20.625V3.375A.375.375 0 013.375 3z" />
              </svg>
              <span class="font-bold text-sm">Haz clic para subir imagen</span>
              <span class="text-[11px]">JPG, PNG, WebP · Máx. 5MB</span>
            </div>
            <input type="file" accept="image/*" class="hidden" onChange={props.onImageChange} />
          </label>
        }
      >
        <div class="relative group rounded-2xl overflow-hidden border-2 border-gray-200">
          <img src={props.currentImageUrl!} alt="Vista previa" class="w-full max-h-64 object-contain bg-gray-50" />
          <div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-3">
            <label class="bg-white text-blue-700 font-black px-4 py-2 rounded-xl text-sm hover:bg-blue-50 transition-all shadow cursor-pointer">
              🖼 Cambiar imagen
              <input type="file" accept="image/*" class="hidden" onChange={props.onImageChange} />
            </label>
            <Show when={props.imagePreview()}>
              <button
                type="button"
                onClick={props.onClearImage}
                class="bg-white text-red-600 font-black px-4 py-2 rounded-xl text-sm hover:bg-red-50 transition-all shadow"
              >↩ Restaurar original</button>
            </Show>
          </div>
          <div class="absolute bottom-2 left-2 bg-black/60 text-white text-[10px] font-bold px-2 py-1 rounded-lg">
            {props.imageFile()?.name ?? "Imagen actual"}
          </div>
        </div>
      </Show>
    </section>
  );
}