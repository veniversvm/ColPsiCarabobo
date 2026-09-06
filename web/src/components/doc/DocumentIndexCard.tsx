// Tarjeta de documento para el índice /documentos.

import { A } from "@solidjs/router";
import type { DocModulo } from "../../lib/documentos";

export default function DocumentIndexCard(props: { doc: DocModulo }) {
  const d = props.doc;
  const secciones = d.secciones.length;
  const artCount = d.secciones.reduce(
    (acc, s) => acc + s.bloques.filter((b) => b.tipo === "articulo").length,
    0,
  );

  return (
    <A
      href={`/documentos/${d.slug}`}
      class="group bg-white rounded-3xl shadow-premium border border-colpsi-border p-7 md:p-8 hover:shadow-xl hover:-translate-y-1 transition-all flex flex-col gap-5"
    >
      <div class="flex items-start justify-between gap-4">
        <span class="inline-block px-3 py-1.5 bg-colpsi-blue text-colpsi-yellow rounded-full text-[10px] font-black uppercase tracking-[0.15em]">
          {d.categoria}
        </span>
        <span class="text-3xl opacity-20 group-hover:opacity-40 transition-opacity">📄</span>
      </div>

      <div class="space-y-3">
        <h3 class="text-xl font-black text-colpsi-blue uppercase tracking-tight leading-tight group-hover:text-blue-800 transition-colors">
          {d.titulo}
        </h3>
        <p class="text-sm text-gray-600 leading-relaxed text-justify">{d.descripcion}</p>
      </div>

      <div class="mt-auto pt-2 flex items-center justify-between border-t border-colpsi-border">
        <div class="text-[11px] text-gray-400 font-bold uppercase tracking-wide">
          {secciones} título(s) · {artCount} artículos
        </div>
        <span class="inline-flex items-center gap-1 text-colpsi-blue font-black text-sm">
          Leer <span class="group-hover:translate-x-0.5 transition-transform">→</span>
        </span>
      </div>
    </A>
  );
}
