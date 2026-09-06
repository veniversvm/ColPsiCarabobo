// Render de un bloque normativo dentro de una sección: capítulos, artículos y texto.

import { Show, For } from "solid-js";
import type { DocumentoBloque } from "../../lib/documentos";
import DocumentText from "./DocumentText";

type Props = {
  bloques: DocumentoBloque[];
};

export default function DocumentSectionBody(props: Props) {
  return (
    <div class="space-y-6">
      <For each={props.bloques}>
        {(b) => (
          <Show
            when={b.tipo === "capitulo"}
            fallback={
              <Show
                when={b.tipo === "articulo"}
                fallback={
                  <p class="text-sm font-bold text-gray-500 uppercase tracking-widest">
                    {b.texto}
                  </p>
                }
              >
                <article class="flex gap-4">
                  <div class="shrink-0 flex flex-col items-center pt-1">
                    <span class="flex items-center justify-center min-w-[44px] h-[44px] px-2 bg-colpsi-blue text-white text-xs font-black rounded-xl shadow-sm">
                      Art. {b.numero}
                    </span>
                  </div>
                  <div class="flex-1 bg-white border border-colpsi-border rounded-2xl p-5 shadow-sm">
                    <DocumentText text={b.texto ?? ""} />
                  </div>
                </article>
              </Show>
            }
          >
            <h4 class="text-lg font-black text-colpsi-blue uppercase tracking-tight border-l-4 border-colpsi-yellow pl-4">
              {b.texto}
            </h4>
          </Show>
        )}
      </For>
    </div>
  );
}
