// Tarjeta de sección colapsable dentro de un documento normativo.
// Muestra el título como botón (abre/cierra) y un enlace para volver al inicio
// de la sección, al principio y al final de la sección expandida.

import { createEffect, createSignal, Show } from "solid-js";
import type { DocumentoSeccion, ControlSecciones } from "../../lib/documentos";
import DocumentSectionBody from "./DocumentSectionBody";

type Props = {
  indice: number;
  seccion: DocumentoSeccion;
  control: ControlSecciones;
  registrar: (indice: number, abrir: () => void) => void;
};

const BtnVolverAlInicio = (callback: () => void) => (
  <button
    type="button"
    onClick={callback}
    class="inline-flex items-center gap-1.5 text-xs font-bold uppercase tracking-widest text-colpsi-blue bg-blue-50 hover:bg-colpsi-yellow hover:text-colpsi-blue px-3 py-2 rounded-lg transition-colors"
  >
    ↑ Volver al inicio
  </button>
);

export default function DocumentSectionCard(props: Props) {
  const [abierta, setAbierta] = createSignal(props.indice === 0);

  createEffect(() => {
    props.registrar(props.indice, () => setAbierta(true));
  });

  createEffect(() => {
    const c = props.control;
    if (c.tick === 0) return;
    setAbierta(c.accion === "abrir");
  });

  const id = () => `seccion-${props.indice}`;

  const irAlInicio = () => {
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  return (
    <section
      id={id()}
      class="bg-white rounded-3xl shadow-premium border border-colpsi-border overflow-hidden scroll-mt-24"
    >
      <button
        type="button"
        onClick={() => setAbierta((v) => !v)}
        aria-expanded={abierta()}
        class="w-full flex items-center justify-between gap-4 p-6 md:p-8 text-left hover:bg-blue-50/50 transition-colors cursor-pointer"
      >
        <span class="text-xl md:text-2xl font-black text-colpsi-blue uppercase tracking-tight border-l-4 border-colpsi-yellow pl-4">
          {props.seccion.titulo}
        </span>
        <span
          class={`shrink-0 flex items-center justify-center w-9 h-9 rounded-full bg-blue-50 text-colpsi-blue text-sm font-black transition-transform ${
            abierta() ? "rotate-180" : ""
          }`}
        >
          ▾
        </span>
      </button>

      <Show when={abierta()}>
        <div class="px-6 md:px-8 pb-8 pt-2 border-t border-colpsi-border">
          <div class="flex justify-center mb-6">
            {BtnVolverAlInicio(irAlInicio)}
          </div>
          <DocumentSectionBody bloques={props.seccion.bloques} />
          <div class="flex justify-center mt-6">
            {BtnVolverAlInicio(irAlInicio)}
          </div>
        </div>
      </Show>
    </section>
  );
}
