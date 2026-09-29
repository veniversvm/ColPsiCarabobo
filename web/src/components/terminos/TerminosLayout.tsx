// web/src/components/terminos/TerminosLayout.tsx
// Layout de los documentos de Términos y Condiciones.
//
// Sirve para los DOS documentos, pero el visitante sin sesión solo llega nunca
// al segundo: el público está en `/terminos` y el del agremiado en
// `/psi/terminos`, que redirige si no hay sesión. Por eso ninguna pieza de este
// layout habla de "partes" ni insinúa que haya un segundo documento: el enlace
// cruzado entre ambos solo se pinta con `mostrarCruce()` (sesión iniciada).
//
// Comparte el lenguaje visual de `components/doc/DocumentLayout.tsx` (hero
// azul, acentos amarillos, tarjetas redondeadas) para que el documento se lea
// como parte del sitio y no como un PDF pegado dentro de una página.
//
// SIN `innerHTML` — a propósito. El texto legal es estático hoy, pero mañana
// alguien va a querer editar una coma y el camino más rápido en Solid es
// `<div innerHTML={...} />`. `Inline` construye nodos, así que aunque el texto
// llegara con `<script>` sería texto y nada más. Es el mismo criterio que
// `lib/sanitize-html.ts` aplica en el resto del proyecto.

import { For, Show, createMemo } from "solid-js";
import { A } from "@solidjs/router";
import { Title, Meta, Link } from "@solidjs/meta";
import { SITE_URL } from "~/lib/site";
import {
  TERMINOS_PARTES,
  getParte,
  agruparSecciones,
  type TerminosParte,
  type TerminosParteClave,
  type TerminosBloque,
} from "~/lib/terminos";
import { COLEGIO } from "~/lib/terminos";
import { useAuth } from "~/lib/auth";

/** Versión que se muestra en la cabecera. La que manda es la de la API. */
import { TERMINOS_VERSION } from "~/lib/terminos";

// -------------------------------------------------------------------------
// Parser de Markdown en línea (subconjunto deliberado y pequeño)
// -------------------------------------------------------------------------

/**
 * Reconoce, y solo esto: **negrita**, *cursiva*, `código` y [enlace](url).
 * Los enlaces solo se aceptan con esquema http/https —Bloquear `javascript:`
 * aquí evita la única vía de XSS que tendría sentido en un documento.
 */
const INLINE_RE =
  /(\*\*[^*]+\*\*)|(\*[^*\n]+\*)|(`[^`]+`)|(\[([^\]]+)\]\((https?:\/\/[^\s)]+)\))/g;

/** Trocea el texto en nodos. Devuelve un arreglo de JSX, no HTML. */
function parseInline(texto: string) {
  const nodos: unknown[] = [];
  let ultimo = 0;
  let m: RegExpExecArray | null;

  // `lastIndex` se reinicia porque INLINE_RE es global y el módulo es singleton.
  INLINE_RE.lastIndex = 0;
  while ((m = INLINE_RE.exec(texto)) !== null) {
    if (m.index > ultimo) nodos.push(texto.slice(ultimo, m.index));
    const [completo, negrita, cursiva, codigo, enlace, textoEnlace, url] = m;

    if (negrita) {
      nodos.push(<strong class="font-black text-colpsi-blue">{negrita.slice(2, -2)}</strong>);
    } else if (cursiva) {
      nodos.push(<em>{cursiva.slice(1, -1)}</em>);
    } else if (codigo) {
      nodos.push(
        <code class="px-1.5 py-0.5 rounded bg-blue-50 text-colpsi-blue text-[0.9em] font-mono">
          {codigo.slice(1, -1)}
        </code>,
      );
    } else if (enlace) {
      nodos.push(
        <a
          href={url}
          target="_blank"
          rel="noopener noreferrer"
          class="text-colpsi-blue font-bold underline underline-offset-2 hover:text-colpsi-red"
        >
          {textoEnlace}
        </a>,
      );
    } else {
      nodos.push(completo);
    }
    ultimo = m.index + completo.length;
  }
  if (ultimo < texto.length) nodos.push(texto.slice(ultimo));

  return nodos;
}

/** Componente que renderiza un fragmento de texto con formato en línea. */
function Inline(props: { texto: string }) {
  return <>{parseInline(props.texto)}</>;
}

// -------------------------------------------------------------------------
// Bloques
// -------------------------------------------------------------------------

function Bloque(props: { bloque: TerminosBloque }) {
  const b = props.bloque;

  return (
    <Show
      when={b.tipo !== "aviso"}
      fallback={
        <Show when={b.tipo === "aviso" && b}>
          {(aviso) => (
            <div
              class={`rounded-2xl border-l-4 p-4 md:p-5 my-5 ${
                aviso().tono === "alerta"
                  ? "bg-red-50 border-colpsi-red text-red-900"
                  : "bg-blue-50 border-colpsi-blue text-blue-950"
              }`}
            >
              <div class="flex gap-3">
                <span aria-hidden="true" class="text-lg leading-none">
                  {aviso().tono === "alerta" ? "⚠️" : "ℹ️"}
                </span>
                <p class="text-sm leading-relaxed text-justify">
                  <Inline texto={aviso().texto} />
                </p>
              </div>
            </div>
          )}
        </Show>
      }
    >
      <Show when={b.tipo === "parrafo" && b}>
        {(p) => (
          <p class="text-gray-700 leading-relaxed text-justify mb-4">
            <Inline texto={p().texto} />
          </p>
        )}
      </Show>

      {/* `ordenada` se respeta: el articulado de los términos es enumerado
          ("1.", "2.", "3."), y viñetarlo perdería la referencia numérica que el
          texto usa al citarse a sí mismo ("lo previsto en el numeral 2"). */}
      <Show when={b.tipo === "lista" && b}>
        {(l) => (
          <Show
            when={l().ordenada}
            fallback={
              <ul class="list-disc list-inside space-y-2 mb-5 text-gray-700 leading-relaxed marker:text-colpsi-yellow">
                <For each={l().items}>
                  {(item) => (
                    <li class="text-justify">
                      <Inline texto={item} />
                    </li>
                  )}
                </For>
              </ul>
            }
          >
            <ol class="list-decimal list-inside space-y-2 mb-5 text-gray-700 leading-relaxed marker:font-black marker:text-colpsi-blue">
              <For each={l().items}>
                {(item) => (
                  <li class="text-justify">
                    <Inline texto={item} />
                  </li>
                )}
              </For>
            </ol>
          </Show>
        )}
      </Show>

      <Show when={b.tipo === "tabla" && b}>
        {(t) => (
          <div class="overflow-x-auto mb-6 rounded-2xl border border-colpsi-border">
            <table class="w-full text-sm border-collapse">
              <thead>
                <tr class="bg-blue-50">
                  <For each={t().cabecera}>
                    {(h) => (
                      <th class="text-left font-black text-colpsi-blue px-4 py-3 border-b border-colpsi-border whitespace-nowrap">
                        {h}
                      </th>
                    )}
                  </For>
                </tr>
              </thead>
              <tbody>
                <For each={t().filas}>
                  {(fila) => (
                    <tr class="even:bg-gray-50">
                      <For each={fila}>
                        {(celda) => (
                          <td class="px-4 py-3 border-b border-colpsi-border text-gray-700 align-top">
                            <Inline texto={celda} />
                          </td>
                        )}
                      </For>
                    </tr>
                  )}
                </For>
              </tbody>
            </table>
          </div>
        )}
      </Show>
    </Show>
  );
}

// -------------------------------------------------------------------------
// Layout
// -------------------------------------------------------------------------

/** Ancla estable por número de sección: "#terminos-10-4". */
function ancla(numero: string) {
  return `terminos-${numero.replace(/\./g, "-")}`;
}

/**
 * `pendiente` marca que a esta parte le falta el texto legal. Si alguien
 * despliega sin pegarlo, la página muestra una banda roja en vez de publicar
 * un documento vacío o —peor— un texto improvisado.
 */
export default function TerminosLayout(props: {
  clave: TerminosParteClave;
  pendiente?: boolean;
  children?: unknown;
}) {
  const parte = createMemo<TerminosParte | undefined>(() => getParte(props.clave));
  const otra = createMemo(() => TERMINOS_PARTES.find((p) => p.clave !== props.clave));
  // El documento se agrupa por nivel para que "5.1" cuelgue de "5".
  const grupos = createMemo(() => agruparSecciones(parte()?.secciones ?? []));

  // El enlace al otro documento (y la nota de que existe) SOLO se muestran con
  // sesión. Para el visitante sin sesión el documento público tiene que leerse
  // como si fuera el único que existe.
  //
  // Se espera a `sessionReady()`: durante SSR y en el primer render del cliente
  // `user()` es null aunque haya sesión, y mostrar el enlace de más no rompe
  // nada, pero el criterio es el mismo que en `routes/psi/terminos.tsx` para no
  // tener dos reglas distintas sobre cuándo la sesión está resuelto.
  const { sessionReady, isAuthenticated } = useAuth();
  const mostrarCruce = () => sessionReady() && isAuthenticated();

  // Cada documento es canónico de SU ruta. El mapeo va explícito porque
  // invertirlo en silencio haría que Google indexara cada página bajo la URL de
  // la otra: el documento público es `/terminos` y el del agremiado
  // `/psi/terminos`.
  const canonical = () =>
    `${SITE_URL}${props.clave === "agremiado" ? "/psi/terminos" : "/terminos"}`;
  // El botón de volver es contextual: desde el documento público no tiene a
  // dónde "volver" dentro del documento (no hay otra parte), así que sale del
  // sitio; desde el del portal sí vuelve al otro texto.
  const volverHref = () => (props.clave === "agremiado" ? "/terminos" : "/explorar");
  const volverTexto = () =>
    props.clave === "agremiado" ? "Volver a los términos" : "Volver al inicio";
  const pageTitle = () =>
    `${parte()?.titulo ?? "Términos y Condiciones"} | ${COLEGIO.nombre}`;
  const descripcion = () =>
    parte()?.resumen ??
    `Términos y Condiciones de uso del sitio del ${COLEGIO.nombre}.`;

  return (
    <>
      <Title>{pageTitle()}</Title>
      <Meta name="description" content={descripcion()} />
      <Meta name="robots" content="index, follow" />
      <Meta property="og:type" content="article" />
      <Meta property="og:url" content={canonical()} />
      <Meta property="og:title" content={pageTitle()} />
      <Meta property="og:description" content={descripcion()} />
      <Meta property="og:image" content={`${SITE_URL}/og-default.jpg`} />
      <Meta property="og:site_name" content={COLEGIO.nombre} />
      <Meta property="og:locale" content="es_VE" />
      <Meta name="twitter:card" content="summary_large_image" />
      <Meta name="twitter:title" content={pageTitle()} />
      <Meta name="twitter:description" content={descripcion()} />
      <Link rel="canonical" href={canonical()} />

      <main class="min-h-screen bg-colpsi-bg pb-24 font-sans">
        <header class="bg-colpsi-blue py-16 px-6 border-b border-blue-900 shadow-inner relative overflow-hidden">
          <div class="absolute left-1/2 top-[-60px] -translate-x-1/2 text-[12rem] opacity-10 font-black select-none pointer-events-none">
            📜
          </div>
          <div class="max-w-4xl mx-auto text-center relative z-10">
            <div class="inline-flex items-center gap-2 px-5 py-2 bg-blue-800/50 text-colpsi-yellow rounded-full text-[10px] font-black tracking-[0.2em] uppercase mb-5 border border-colpsi-yellow/30">
              {/* Etiqueta neutral. Antes imprimía `parte()?.etiqueta`, que para el
                  documento público decía literalmente "Parte I" en el héroe. */}
              {props.clave === "agremiado" ? "Portal del agremiado" : "Condiciones de uso"}
            </div>
            <h1 class="text-3xl md:text-5xl font-black text-white tracking-tight leading-tight mb-4">
              {parte()?.titulo ?? "Términos y Condiciones"}
            </h1>
            <p class="text-blue-200 text-sm font-medium">
              Versión {TERMINOS_VERSION} · {COLEGIO.nombre}
            </p>
          </div>
        </header>

        <section class="max-w-4xl mx-auto px-6 py-12 relative z-10 -mt-8">
          <Show when={props.pendiente}>
            <div class="bg-red-50 border-2 border-colpsi-red text-red-900 rounded-3xl p-6 mb-10">
              <h2 class="font-black text-lg mb-2">Texto legal pendiente</h2>
              <p class="text-sm leading-relaxed">
                Este documento todavía no tiene su texto definitivo. Si estás
                leyendo esto, el Colegio tiene que revisar y publicar la versión
                final antes de que sea visible para el público.
              </p>
            </div>
          </Show>

          <Show when={parte()}>
            {(p) => (
              <>
                <div class="bg-white rounded-3xl shadow-premium border border-colpsi-border p-6 md:p-8 mb-10">
                  <p class="text-gray-600 leading-relaxed text-justify">{p().resumen}</p>

                  <div class="mt-6 flex flex-wrap gap-3">
                    <A
                      href={volverHref()}
                      class="inline-flex items-center gap-2 bg-blue-50 text-colpsi-blue font-bold px-5 py-3 rounded-xl hover:bg-blue-100 transition-colors uppercase text-xs tracking-widest border border-blue-100"
                    >
                      ← {volverTexto()}
                    </A>
                    {/* Enlace al otro documento: solo con sesión. Para el anónimo no
                        se renderiza nada — ni el botón, ni una versión apagada que
                        lo insinúe. */}
                    <Show when={mostrarCruce() && otra()}>
                      {(o) => (
                        <A
                          href={o().clave === "agremiado" ? "/psi/terminos" : "/terminos"}
                          class="inline-flex items-center gap-2 bg-colpsi-blue text-white font-black px-5 py-3 rounded-xl hover:bg-blue-800 transition-colors uppercase text-xs tracking-widest shadow-lg shadow-colpsi-blue/20"
                        >
                          {o().etiqueta} →
                        </A>
                      )}
                    </Show>
                  </div>
                </div>

                <Show when={grupos().length > 1}>
                  <nav class="bg-white rounded-2xl border border-colpsi-border shadow-sm p-5 mb-8">
                    <h3 class="text-xs font-black uppercase tracking-widest text-colpsi-blue mb-3">
                      Contenido
                    </h3>
                    {/* Solo los artículos de primer nivel. Meter las subsecciones
                        duplicaría cada entrada del sumario en un documento que
                        ya tiene quince, y la subsección se alcanza por su padre. */}
                    <ul class="flex flex-wrap gap-2">
                      <For each={grupos()}>
                        {(g) => (
                          <li>
                            <a
                              href={`#${ancla(g.seccion.numero)}`}
                              class="inline-block bg-blue-50 text-gray-700 text-xs font-bold px-3 py-2 rounded-lg hover:bg-colpsi-yellow hover:text-colpsi-blue transition-colors"
                            >
                              {g.seccion.numero} · {g.seccion.titulo}
                            </a>
                          </li>
                        )}
                      </For>
                    </ul>
                  </nav>
                </Show>

                <div class="space-y-10">
                  <For each={grupos()}>
                    {(g) => (
                      <article
                        id={ancla(g.seccion.numero)}
                        class="bg-white rounded-3xl shadow-premium border border-colpsi-border p-6 md:p-8 scroll-mt-24"
                      >
                        <h2 class="text-xl md:text-2xl font-black text-colpsi-blue tracking-tight mb-1">
                          {/* El número va en rojo, no en amarillo: esta tarjeta es
                              blanca y `#facc15` sobre blanco era ilegible. */}
                          <span class="text-colpsi-red">{g.seccion.numero}.</span>{" "}
                          {g.seccion.titulo}
                        </h2>
                        <div class="h-1 w-16 bg-colpsi-yellow rounded-full mb-5" />
                        <For each={g.seccion.bloques}>{(b) => <Bloque bloque={b} />}</For>

                        {/* Subsecciones dentro de la tarjeta de su padre, con un
                            guía lateral y un título menor: la jerarquía del
                            articulado se lee, no se infiere del tamaño. */}
                        <Show when={g.hijos.length > 0}>
                          <div class="mt-7 space-y-6 border-l-2 border-blue-100 pl-5 md:pl-6">
                            <For each={g.hijos}>
                              {(h) => (
                                <section id={ancla(h.numero)} class="scroll-mt-24">
                                  <h3 class="text-base md:text-lg font-black text-colpsi-blue tracking-tight mb-3">
                                    <span class="text-colpsi-red">{h.numero}</span>{" "}
                                    {h.titulo}
                                  </h3>
                                  <For each={h.bloques}>{(b) => <Bloque bloque={b} />}</For>
                                </section>
                              )}
                            </For>
                          </div>
                        </Show>
                      </article>
                    )}
                  </For>
                </div>

                {props.children}
              </>
            )}
          </Show>

          <div class="mt-12 flex justify-center">
            <button
              type="button"
              onClick={() => document.body.scrollIntoView({ behavior: "smooth" })}
              class="inline-flex items-center gap-2 bg-colpsi-blue text-white font-bold px-6 py-3 rounded-xl hover:bg-blue-800 transition-colors uppercase text-xs tracking-widest shadow-lg"
            >
              ↑ Volver al inicio
            </button>
          </div>
        </section>
      </main>
    </>
  );
}
