// Página índice del Marco Legal y Normativo del gremio.

import { For } from "solid-js";
import { Title, Meta, Link } from "@solidjs/meta";
import { documentos } from "../../lib/documentos";
import DocumentIndexCard from "../../components/doc/DocumentIndexCard";

const SITE_URL = import.meta.env.VITE_SITE_URL || "http://localhost:3000";

export default function DocumentosIndex() {
  const canonicalUrl = `${SITE_URL}/documentos`;
  const ogImage = `${SITE_URL}/og-default.jpg`;
  const pageTitle = "Marco Legal y Normativo | Colegio de Psicólogos del Estado Carabobo";
  const pageDescription =
    "Consulta los estatutos, la Ley de Ejercicio de la Psicología, el Código de Ética y el Reglamento Interno del Colegio de Psicólogos del Estado Carabobo (CPEC).";

  return (
    <>
      <Title>{pageTitle}</Title>
      <Meta name="description" content={pageDescription} />
      <Meta
        name="keywords"
        content="Marco legal, normativa, estatutos, ley de ejercicio de la psicología, código de ética, reglamento interno, CPEC, Colegio de Psicólogos, Carabobo"
      />
      <Meta name="robots" content="index, follow" />

      <Meta property="og:type" content="website" />
      <Meta property="og:url" content={canonicalUrl} />
      <Meta property="og:title" content={pageTitle} />
      <Meta property="og:description" content={pageDescription} />
      <Meta property="og:image" content={ogImage} />
      <Meta property="og:site_name" content="Colegio de Psicólogos del Estado Carabobo" />
      <Meta property="og:locale" content="es_VE" />

      <Meta name="twitter:card" content="summary_large_image" />
      <Meta name="twitter:title" content={pageTitle} />
      <Meta name="twitter:description" content={pageDescription} />
      <Meta name="twitter:image" content={ogImage} />

      <Link rel="canonical" href={canonicalUrl} />

      <main class="min-h-screen bg-[#f8fafc] pb-24 font-sans">
        <header class="bg-colpsi-blue py-20 px-6 border-b border-blue-900 shadow-inner relative overflow-hidden">
          <div class="absolute left-1/2 top-[-60px] -translate-x-1/2 text-[14rem] opacity-10 font-black select-none pointer-events-none">
            ⚖️
          </div>
          <div class="max-w-4xl mx-auto text-center relative z-10">
            <div class="inline-block px-5 py-2 bg-blue-800/50 text-colpsi-yellow rounded-full text-[10px] font-black tracking-[0.2em] uppercase mb-6 border border-colpsi-yellow/30">
              Transparencia institucional
            </div>
            <h1 class="text-4xl md:text-6xl font-black text-white tracking-tight leading-tight">
              Marco{" "}
              <span class="text-colpsi-yellow italic">Legal y Normativo</span>
            </h1>
            <p class="text-blue-200 text-lg max-w-2xl mx-auto font-medium leading-relaxed mt-4">
              Estatutos, ley de ejercicio profesional, código de ética y reglamento
              que rigen nuestra corporación y la profesión del Psicólogo en Venezuela.
            </p>
          </div>
        </header>

        <section class="max-w-5xl mx-auto px-6 py-14">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <For each={documentos}>
              {(d) => <DocumentIndexCard doc={d} />}
            </For>
          </div>

          <div class="mt-12 bg-blue-50 border border-blue-100 rounded-2xl p-6 text-center">
            <p class="text-sm text-gray-600 leading-relaxed">
              Los documentos aquí presentados son versiones de consulta de la normativa
              del Colegio de Psicólogos del Estado Carabobo y de la Federación de
              Psicólogos de Venezuela. Para fines legales, rige el texto oficial
              publicado en los organismos competentes.
            </p>
          </div>
        </section>
      </main>
    </>
  );
}
