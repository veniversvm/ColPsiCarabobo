// Layout compartido para las páginas de detalle de un documento normativo.
// Incluye hero, tabla de contenidos (anclas a títulos) y botón de descarga.

import { For, createSignal, Show } from "solid-js";
import { A } from "@solidjs/router";
import { Title, Meta, Link } from "@solidjs/meta";
import type { ControlSecciones, DocModulo } from "../../lib/documentos";
import DocumentSectionCard from "./DocumentSectionCard";

const SITE_URL = import.meta.env.VITE_SITE_URL || "http://localhost:3000";

function slugId(i: number) {
  return `seccion-${i}`;
}

export default function DocumentLayout(props: { doc: DocModulo }) {
  const doc = props.doc;
  const [control, setControl] = createSignal<ControlSecciones>({
    accion: "abrir",
    tick: 0,
  });

  const abrirSeccion = new Map<number, () => void>();

  const registrar = (indice: number, abrir: () => void) => {
    abrirSeccion.set(indice, abrir);
  };

  const irACapitulo = (indice: number) => {
    const abrir = abrirSeccion.get(indice);
    if (abrir) abrir();
    requestAnimationFrame(() => {
      const el = document.getElementById(slugId(indice));
      if (el) el.scrollIntoView({ behavior: "smooth", block: "start" });
    });
  };

  const canonical = `${SITE_URL}/documentos/${doc.slug}`;
  const ogImage = `${SITE_URL}/og-default.jpg`;
  const pageTitle = `${doc.titulo} | Colegio de Psicólogos del Estado Carabobo`;

  return (
    <>
      <Title>{pageTitle}</Title>
      <Meta name="description" content={doc.descripcion} />
      <Meta
        name="keywords"
        content={`${doc.categoria}, ${doc.titulo}, Colegio de Psicólogos, Carabobo, Venezuela, normativa, legislación`}
      />
      <Meta name="robots" content="index, follow" />
      <Meta property="og:type" content="article" />
      <Meta property="og:url" content={canonical} />
      <Meta property="og:title" content={pageTitle} />
      <Meta property="og:description" content={doc.descripcion} />
      <Meta property="og:image" content={ogImage} />
      <Meta property="og:site_name" content="Colegio de Psicólogos del Estado Carabobo" />
      <Meta property="og:locale" content="es_VE" />
      <Meta name="twitter:card" content="summary_large_image" />
      <Meta name="twitter:title" content={pageTitle} />
      <Meta name="twitter:description" content={doc.descripcion} />
      <Link rel="canonical" href={canonical} />

      <main class="min-h-screen bg-colpsi-bg pb-24 font-sans">
        <header class="bg-colpsi-blue py-16 px-6 border-b border-blue-900 shadow-inner relative overflow-hidden">
          <div class="absolute left-1/2 top-[-60px] -translate-x-1/2 text-[12rem] opacity-10 font-black select-none pointer-events-none">
            ⚖️
          </div>
          <div class="max-w-4xl mx-auto text-center relative z-10">
            <div class="inline-flex items-center gap-2 px-5 py-2 bg-blue-800/50 text-colpsi-yellow rounded-full text-[10px] font-black tracking-[0.2em] uppercase mb-5 border border-colpsi-yellow/30">
              {doc.categoria}
            </div>
            <h1 class="text-3xl md:text-5xl font-black text-white tracking-tight leading-tight mb-4">
              {doc.titulo}
            </h1>
            <p class="text-blue-200 text-sm md:text-base font-medium max-w-2xl mx-auto">
              {doc.fuente}
            </p>
          </div>
        </header>

        <section class="max-w-4xl mx-auto px-6 py-12 relative z-10 -mt-8">
          <div class="bg-white rounded-3xl shadow-premium border border-colpsi-border p-6 md:p-8 mb-10">
            <p class="text-gray-600 leading-relaxed text-justify">{doc.descripcion}</p>
            <div class="mt-6 flex flex-wrap gap-3">
              <a
                href={doc.archivo}
                download
                class="inline-flex items-center gap-2 bg-colpsi-yellow text-colpsi-blue font-black px-5 py-3 rounded-xl hover:scale-[1.03] active:scale-95 transition-all uppercase text-xs tracking-widest shadow-lg shadow-yellow-500/20"
              >
                Descargar {doc.tipoArchivo} <span>↓</span>
              </a>
              <A
                href="/documentos"
                class="inline-flex items-center gap-2 bg-blue-50 text-colpsi-blue font-bold px-5 py-3 rounded-xl hover:bg-blue-100 transition-colors uppercase text-xs tracking-widest border border-blue-100"
              >
                ← Todos los documentos
              </A>
            </div>
          </div>

          {doc.secciones.length > 1 && (
            <nav class="bg-white rounded-2xl border border-colpsi-border shadow-sm p-5 mb-8">
              <div class="flex items-center justify-between flex-wrap gap-2 mb-3">
                <h3 class="text-xs font-black uppercase tracking-widest text-colpsi-blue">
                  Contenido
                </h3>
                <div class="flex gap-2">
                  <button
                    type="button"
                    onClick={() =>
                      setControl((c) => ({ accion: "abrir", tick: c.tick + 1 }))
                    }
                    class="text-[11px] font-bold uppercase tracking-widest text-colpsi-blue bg-blue-50 hover:bg-blue-100 px-3 py-1.5 rounded-lg transition-colors"
                  >
                    Abrir todas
                  </button>
                  <button
                    type="button"
                    onClick={() =>
                      setControl((c) => ({ accion: "colapsar", tick: c.tick + 1 }))
                    }
                    class="text-[11px] font-bold uppercase tracking-widest text-colpsi-blue bg-blue-50 hover:bg-blue-100 px-3 py-1.5 rounded-lg transition-colors"
                  >
                    Colapsar todas
                  </button>
                </div>
              </div>
              <ul class="flex flex-wrap gap-2">
                <For each={doc.secciones}>
                  {(s, i) => (
                    <li>
                      <a
                        href={`#${slugId(i())}`}
                        onClick={(ev) => {
                          ev.preventDefault();
                          irACapitulo(i());
                        }}
                        class="inline-block bg-blue-50 text-gray-700 text-xs font-bold px-3 py-2 rounded-lg hover:bg-colpsi-yellow hover:text-colpsi-blue transition-colors"
                      >
                        {s.titulo}
                      </a>
                    </li>
                  )}
                </For>
              </ul>
            </nav>
          )}

          <div class="space-y-10">
            <For each={doc.secciones}>
              {(s, i) => (
                <DocumentSectionCard
                  indice={i()}
                  seccion={s}
                  control={control()}
                  registrar={registrar}
                />
              )}
            </For>
          </div>

          <Show when={doc.secciones.length > 1}>
            <div class="mt-12 flex justify-center">
              <button
                type="button"
                onClick={() => document.body.scrollIntoView({ behavior: "smooth" })}
                class="inline-flex items-center gap-2 bg-colpsi-blue text-white font-bold px-6 py-3 rounded-xl hover:bg-blue-800 transition-colors uppercase text-xs tracking-widest shadow-lg"
              >
                ↑ Volver al inicio
              </button>
            </div>
          </Show>
        </section>
      </main>
    </>
  );
}
