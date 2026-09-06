// web/src/routes/psi/documentos.tsx

import { createResource, For, Show, Suspense } from "solid-js";
import { A } from "@solidjs/router";
import { apiGet } from "~/lib/api";
import { bucketUrl } from "~/lib/bucket";
import {
  DOCUMENT_TYPE_LABELS,
  DOCUMENT_TYPE_EMOJI,
  DOCUMENT_TYPE_ORDER,
  isPdf,
} from "~/lib/doc_tipos";
import type { PsiUserDocument } from "~/types/psi";

const formatDate = (dateStr?: string) => {
  if (!dateStr) return "";
  const d = new Date(dateStr);
  if (isNaN(d.getTime())) return "";
  return d.toLocaleDateString("es-VE", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
};

export default function MisDocumentosPage() {
  const [docs] = createResource<PsiUserDocument[]>(
    () => apiGet<PsiUserDocument[]>("/psi/me/documents"),
  );

  return (
    <main class="bg-colpsi-bg min-h-screen pb-24 font-sans">
      {/* ── HEADER ──────────────────────────────────────────────────────────── */}
      <div class="bg-heraldic pt-10 pb-20 px-4 md:px-8 shadow-inner">
        <div class="max-w-4xl mx-auto flex items-center justify-between">
          <A
            href="/psi"
            class="bg-colpsi-yellow text-colpsi-blue px-5 py-2.5 rounded-full font-black text-sm shadow-lg hover:bg-colpsi-yellow/90 active:scale-95 transition-all inline-flex items-center gap-2"
          >
            <span>←</span> Volver al Panel
          </A>
        </div>
        <div class="max-w-4xl mx-auto mt-8">
          <h1 class="text-white text-3xl font-black">Mis Documentos</h1>
          <p class="text-blue-200 mt-1 italic uppercase text-xs tracking-widest font-bold">
            Expediente Digital
          </p>
        </div>
      </div>

      <div class="max-w-4xl mx-auto px-4 md:px-8 -mt-10 space-y-8">
        {/* ── AVISO DE SOLO LECTURA ────────────────────────────────────────── */}
        <div class="bg-blue-50/80 border border-blue-100 rounded-3xl p-5 flex items-start gap-4">
          <span class="text-2xl">🔒</span>
          <div>
            <p class="text-sm font-black text-colpsi-blue">Solo lectura</p>
            <p class="text-xs text-gray-600 mt-0.5 leading-relaxed">
              Estos documentos (cédula, título, RIF, comprobantes y otros) son
              gestionados y verificados por la administración del Colegio. No
              puedes editarlos ni eliminarlos; si necesitas una corrección,
              contacta a la administración.
            </p>
          </div>
        </div>

        {/* ── GALERÍA AGRUPADA ─────────────────────────────────────────────── */}
        <Suspense fallback={<div class="h-48 bg-white animate-pulse rounded-[2.5rem]" />}>
          <Show
            when={(docs()?.length ?? 0) > 0}
            fallback={
              <div class="bg-white rounded-[2.5rem] p-8 md:p-10 shadow-premium border border-gray-100">
                <div class="flex items-start gap-4">
                  <span class="text-4xl">📂</span>
                  <div class="flex-grow space-y-2">
                    <h3 class="text-lg font-black text-colpsi-blue">
                      Aún no tienes documentos en tu expediente
                    </h3>
                    <p class="text-sm text-gray-600">
                      La administración cargará tus soportes (CI, título, RIF,
                      comprobantes de solvencia, etc.) aquí apenas estén
                      verificados.
                    </p>
                  </div>
                </div>
              </div>
            }
          >
            <div class="space-y-8">
              <For each={DOCUMENT_TYPE_ORDER}>
                {(type) => {
                  const group = () => (docs() ?? []).filter((d) => (d.document_type ?? "otro") === type);
                  return (
                    <Show when={group().length > 0}>
                      <section class="space-y-4">
                        <div class="flex items-center gap-2 px-2">
                          <span class="text-xl">{DOCUMENT_TYPE_EMOJI[type]}</span>
                          <h2 class="text-colpsi-blue font-black uppercase tracking-wide text-sm">
                            {DOCUMENT_TYPE_LABELS[type]}
                          </h2>
                          <span class="text-[10px] font-black text-gray-400 bg-gray-100 px-2 py-0.5 rounded-full">
                            {group().length}
                          </span>
                        </div>
                        <div class="space-y-4">
                          <For each={group()}>
                            {(doc: PsiUserDocument) => (
                              <div class="bg-white rounded-3xl p-5 md:p-6 shadow-premium border border-gray-100 flex flex-col md:flex-row gap-5">
                                <a
                                  href={bucketUrl(doc.document_url)}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  class="shrink-0 w-full md:w-28 h-40 md:h-28 bg-gray-50 rounded-2xl border border-gray-200 overflow-hidden flex items-center justify-center hover:ring-2 hover:ring-colpsi-yellow transition-all"
                                >
                                  <Show
                                    when={!isPdf(doc)}
                                    fallback={
                                      <div class="flex flex-col items-center justify-center text-red-500">
                                        <span class="text-3xl">📕</span>
                                        <span class="text-[9px] font-black uppercase mt-1 px-2 text-center">
                                          {doc.filename?.match(/\.\w+$/)?.[0] || "PDF"}
                                        </span>
                                      </div>
                                    }
                                  >
                                    <img
                                      src={bucketUrl(doc.document_url)}
                                      alt={doc.title}
                                      class="w-full h-full object-cover"
                                    />
                                  </Show>
                                </a>
                                <div class="flex-grow min-w-0">
                                  <div class="flex items-baseline justify-between gap-3 flex-wrap">
                                    <h3 class="text-lg font-black text-colpsi-blue uppercase leading-tight">
                                      {doc.title}
                                    </h3>
                                    <span class="text-[10px] font-black text-gray-400 uppercase">
                                      {formatDate(doc.document_date)}
                                    </span>
                                  </div>
                                  <Show when={doc.notes}>
                                    <p class="text-xs text-gray-500 mt-1.5">
                                      {doc.notes}
                                    </p>
                                  </Show>
                                  <div class="flex items-center gap-4 mt-3 flex-wrap">
                                    <Show when={doc.filename}>
                                      <span class="text-[10px] font-bold text-gray-400 truncate max-w-[200px]">
                                        📎 {doc.filename}
                                      </span>
                                    </Show>
                                    <a
                                      href={bucketUrl(doc.document_url)}
                                      target="_blank"
                                      rel="noopener noreferrer"
                                      class="ml-auto bg-colpsi-yellow text-colpsi-blue px-4 py-2 rounded-full font-black text-xs shadow-sm hover:bg-colpsi-yellow/90 active:scale-95 transition-all inline-flex items-center gap-1"
                                    >
                                      {isPdf(doc) ? "Abrir PDF" : "Ver documento"} ↗
                                    </a>
                                  </div>
                                </div>
                              </div>
                            )}
                          </For>
                        </div>
                      </section>
                    </Show>
                  );
                }}
              </For>
            </div>
          </Show>
        </Suspense>
      </div>
    </main>
  );
}