// web/src/routes/inscripcion.tsx

import { createSignal, Show } from "solid-js";
import { Title, Meta, Link } from "@solidjs/meta";
import { InscriptionForm } from "~/components/inscripcion/InscriptionForm";

const SITE_URL = import.meta.env.VITE_SITE_URL || "https://colpsicarabobo.org";

export default function InscriptionPage() {
  const canonicalUrl = `${SITE_URL}/inscripcion`;
  const ogImage = `${SITE_URL}/og-default.jpg`;
  const pageTitle = "Procedimiento de Inscripción | Colegio de Psicólogos del Estado Carabobo";
  const pageDescription =
    "Conoce los pasos y requisitos legales exactos para formalizar tu colegiatura en el Colegio de Psicólogos del Estado Carabobo.";
  const [tab, setTab] = createSignal<"instrucciones" | "formulario">("instrucciones");

  return (
    <>
      {/* ── SEO & METADATA (SSR) ───────────────────────────────────────────── */}
      <Title>{pageTitle}</Title>
      <Meta name="description" content={pageDescription} />
      <Meta
        name="keywords"
        content="Colegio de Psicólogos, Carabobo, inscripción, colegiatura, FPV, requisitos, psicólogos Venezuela"
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

      {/* ── CONTENIDO DE LA PÁGINA ─────────────────────────────────────────── */}
      <main class="bg-colpsi-bg min-h-screen pb-24 font-sans">

        {/* ── HERO SUTIL ─────────────────────────────────────────────────── */}
        <header class="bg-heraldic py-20 px-6 border-b border-blue-900 shadow-inner">
          <div class="max-w-4xl mx-auto text-center">
            <div class="inline-block px-5 py-2 bg-blue-800/50 text-colpsi-yellow rounded-full text-[10px] font-black tracking-[0.2em] uppercase mb-6 border border-colpsi-yellow/30">
              Trámites Legales
            </div>
            <h1 class="text-4xl md:text-6xl font-black text-white tracking-tight leading-tight">
              Procedimiento de <br />
              <span class="text-colpsi-yellow italic">Inscripción</span>
            </h1>
          </div>
        </header>

        <section class="max-w-4xl mx-auto px-6 py-16 space-y-12 -mt-10 relative z-10">

          {/* ── ADVERTENCIA LEGAL INICIAL ────────────────────────────────── */}
          <div class="bg-blue-50/90 rounded-[2.5rem] shadow-premium p-8 md:p-10 border border-blue-100 flex items-start gap-6">
            <div class="text-5xl drop-shadow-sm">⚖️</div>
            <div>
              <h2 class="text-sm font-black text-blue-900 uppercase tracking-widest mb-2">
                Aviso Importante
              </h2>
              <p class="text-blue-800 text-sm leading-relaxed font-medium text-justify">
                Este procedimiento es exclusivamente para profesionales debidamente graduados en psicología o licenciatura en psicología en Venezuela, o con título extranjero equivalente debidamente validado.
              </p>
            </div>
          </div>

          {/* ── SELECTOR DE SECCIÓN: INSTRUCCIONES | FORMULARIO ──────────── */}
          <div class="bg-white p-3 rounded-3xl shadow-premium border border-gray-100 flex flex-col sm:flex-row gap-2">
            <button
              onClick={() => setTab("instrucciones")}
              class={`flex-1 py-4 px-6 rounded-2xl font-black text-base flex items-center justify-center gap-3 transition-all ${
                tab() === "instrucciones"
                  ? "bg-colpsi-blue text-white shadow"
                  : "bg-gray-50 text-gray-500 hover:bg-gray-100"
              }`}
            >
              <span>📋</span> Instrucciones
            </button>
            <button
              onClick={() => setTab("formulario")}
              class={`flex-1 py-4 px-6 rounded-2xl font-black text-base flex items-center justify-center gap-3 transition-all ${
                tab() === "formulario"
                  ? "bg-colpsi-blue text-white shadow"
                  : "bg-gray-50 text-gray-500 hover:bg-gray-100"
              }`}
            >
              <span>📝</span> Formulario de Inscripción
            </button>
          </div>

          {/* ── VISTA: INSTRUCCIONES (pasos del procedimiento) ───────────── */}
          <Show when={tab() === "instrucciones"}>
            <div class="space-y-12">
              {/* ── PASO 1 ─────────────────────────────────────────────────────── */}
              <div class="bg-white p-8 md:p-12 rounded-[2.5rem] shadow-sm border border-gray-100 group hover:border-colpsi-yellow transition-colors">
                <div class="flex flex-col sm:flex-row sm:items-center gap-4 mb-8 border-b border-gray-100 pb-6">
                  <div class="w-14 h-14 bg-amber-50 text-amber-600 rounded-2xl flex items-center justify-center text-3xl font-black group-hover:scale-110 transition-transform border border-amber-100 flex-shrink-0">
                    1
                  </div>
                  <h3 class="text-2xl font-black text-colpsi-blue tracking-tight leading-tight">
                    Determinación de la Deuda y Cálculo del Monto en Bolívares
                  </h3>
                </div>

                <div class="space-y-6 text-gray-600 text-sm leading-relaxed font-medium text-justify">
                  <p>
                    El profesional que aspira a la colegiación debe realizar el cálculo de los aranceles correspondientes según las directrices vigentes emanadas de la Asamblea Extraordinaria del 18 de abril de 2026.
                  </p>

                  <div class="bg-gray-50 border-l-4 border-colpsi-blue p-5 rounded-r-2xl">
                    <p>
                      <strong class="text-colpsi-blue">Política de Exoneración:</strong> Todos los años de graduación anteriores al 2024 se consideran completamente exentos de cualquier tipo de deuda por anualidad. Los cálculos se inician de forma obligatoria a partir del año 2024.
                    </p>
                  </div>

                  <div>
                    <p class="font-bold text-gray-900 mb-3">Aranceles Base (Fijados en Divisas):</p>
                    <ul class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
                      <li class="bg-gray-50 p-4 rounded-xl border border-gray-100"><strong class="text-gray-900">Derecho de Inscripción:</strong> $30 (pago único de ingreso).</li>
                      <li class="bg-gray-50 p-4 rounded-xl border border-gray-100"><strong class="text-gray-900">Anualidad 2024:</strong> $40</li>
                      <li class="bg-gray-50 p-4 rounded-xl border border-gray-100"><strong class="text-gray-900">Anualidad 2025:</strong> $40</li>
                      <li class="bg-gray-50 p-4 rounded-xl border border-gray-100"><strong class="text-gray-900">Anualidad 2026:</strong> $40</li>
                    </ul>
                  </div>

                  <p>
                    <strong class="text-colpsi-blue">Mecanismo de Conversión Legal:</strong> La sumatoria total obtenida en dólares (inscripción más los años de solventación correspondientes desde su graduación) debe ser cancelada estrictamente en bolívares. Para ello, se aplicará de forma obligatoria la tasa de cambio oficial del día de la transacción emitida por el Banco Central de Venezuela (BCV).
                  </p>
                </div>
              </div>

              {/* ── PASO 2 ─────────────────────────────────────────────────────── */}
              <div class="bg-white p-8 md:p-12 rounded-[2.5rem] shadow-sm border border-gray-100 group hover:border-colpsi-yellow transition-colors">
                <div class="flex flex-col sm:flex-row sm:items-center gap-4 mb-8 border-b border-gray-100 pb-6">
                  <div class="w-14 h-14 bg-emerald-50 text-emerald-600 rounded-2xl flex items-center justify-center text-3xl font-black group-hover:scale-110 transition-transform border border-emerald-100 flex-shrink-0">
                    2
                  </div>
                  <h3 class="text-2xl font-black text-colpsi-blue tracking-tight leading-tight">
                    Ejecución del Pago Bancario
                  </h3>
                </div>

                <div class="space-y-6 text-gray-600 text-sm leading-relaxed font-medium text-justify">
                  <p>
                    Una vez obtenido el monto exacto en bolívares, se debe realizar un depósito o transferencia bancaria en la siguiente cuenta institucional:
                  </p>

                  <div class="bg-emerald-50/50 rounded-2xl border border-emerald-100 p-6 my-6">
                    <ul class="grid grid-cols-1 md:grid-cols-2 gap-y-4 gap-x-8 text-sm">
                      <li class="flex flex-col border-b border-emerald-100 pb-2 gap-1">
                        <span class="text-[10px] text-emerald-700 uppercase tracking-widest font-black">Banco</span>
                        <span class="font-black text-emerald-900">Provincial</span>
                      </li>
                      <li class="flex flex-col border-b border-emerald-100 pb-2 gap-1">
                        <span class="text-[10px] text-emerald-700 uppercase tracking-widest font-black">Tipo de Cuenta</span>
                        <span class="font-black text-emerald-900">Corriente</span>
                      </li>
                      <li class="flex flex-col border-b border-emerald-100 pb-2 gap-1">
                        <span class="text-[10px] text-emerald-700 uppercase tracking-widest font-black">Número de Cuenta</span>
                        <span class="font-black text-emerald-900">0108-0558-94-0100208134</span>
                      </li>
                      <li class="flex flex-col border-b border-emerald-100 pb-2 gap-1">
                        <span class="text-[10px] text-emerald-700 uppercase tracking-widest font-black">A nombre de</span>
                        <span class="font-black text-emerald-900">Colegio de Psicólogos del Estado Carabobo</span>
                      </li>
                      <li class="flex flex-col border-b border-emerald-100 pb-2 gap-1">
                        <span class="text-[10px] text-emerald-700 uppercase tracking-widest font-black">RIF</span>
                        <span class="font-black text-emerald-900">J-508172418</span>
                      </li>
                    </ul>
                  </div>

                  <p>
                    <strong class="text-colpsi-blue">Requisito del Soporte:</strong> El comprobante digital o físico emitido por la plataforma bancaria debe guardarse e imprimirse garantizando que sean perfectamente visibles cuatro datos fundamentales: el banco emisor, el número de referencia o transferencia, el monto exacto debitado y la fecha de la operación.
                  </p>
                </div>
              </div>

              {/* ── PASO 3 ─────────────────────────────────────────────────────── */}
              <div class="bg-white p-8 md:p-12 rounded-[2.5rem] shadow-sm border border-gray-100 group hover:border-colpsi-yellow transition-colors">
                <div class="flex flex-col sm:flex-row sm:items-center gap-4 mb-8 border-b border-gray-100 pb-6">
                  <div class="w-14 h-14 bg-purple-50 text-purple-600 rounded-2xl flex items-center justify-center text-3xl font-black group-hover:scale-110 transition-transform border border-purple-100 flex-shrink-0">
                    3
                  </div>
                  <h3 class="text-2xl font-black text-colpsi-blue tracking-tight leading-tight">
                    Postulación Digital y Envío de Correo Electrónico
                  </h3>
                </div>

                <div class="space-y-6 text-gray-600 text-sm leading-relaxed font-medium text-justify">
                  <p>
                    El interesado debe escanear y enviar sus requisitos por vía digital. Toda la documentación adjunta debe estar dispuesta en formato tamaño carta (con excepción de la fotografía). El correo debe dirigirse estrictamente a la siguiente dirección electrónica:
                  </p>

                  <div class="bg-purple-50 text-center p-6 rounded-2xl border border-purple-100">
                    <p class="text-xs font-black text-purple-600 uppercase tracking-widest mb-1">Correo Electrónico Oficial</p>
                    <p class="text-lg font-black text-purple-900">admon.colpsicarabobo@gmail.com</p>
                  </div>

                  <p class="font-bold text-gray-900 mt-6">Los documentos a adjuntar en formato digital son:</p>

                  <ul class="space-y-4">
                    <li class="flex items-start gap-3 bg-gray-50 p-5 rounded-2xl border border-gray-100">
                      <span class="text-purple-500 font-black text-lg mt-0.5">•</span>
                      <span>Una (1) foto tipo carnet reciente, con fondo blanco.</span>
                    </li>
                    <li class="flex items-start gap-3 bg-gray-50 p-5 rounded-2xl border border-gray-100">
                      <span class="text-purple-500 font-black text-lg mt-0.5">•</span>
                      <span>Una (1) copia del título de psicólogo, el cual debe haber sido registrado previamente en el Registro Principal correspondiente a su jurisdicción.</span>
                    </li>
                    <li class="flex items-start gap-3 bg-gray-50 p-5 rounded-2xl border border-gray-100">
                      <span class="text-purple-500 font-black text-lg mt-0.5">•</span>
                      <span>Una (1) copia legible de la cédula de identidad o del pasaporte vigente.</span>
                    </li>
                    <li class="flex items-start gap-3 bg-gray-50 p-5 rounded-2xl border border-gray-100">
                      <span class="text-purple-500 font-black text-lg mt-0.5">•</span>
                      <span>Una (1) copia del Registro de Información Fiscal (RIF) vigente, el cual debe estar actualizado y reflejar su domicilio fiscal en la región determinada dentro de la circunscripción del Estado Carabobo.</span>
                    </li>
                    <li class="flex items-start gap-3 bg-gray-50 p-5 rounded-2xl border border-gray-100">
                      <span class="text-purple-500 font-black text-lg mt-0.5">•</span>
                      <span>El comprobante de transferencia o depósito bancario para la verificación de los fondos y posterior emisión del recibo oficial.</span>
                    </li>
                  </ul>
                </div>
              </div>

              {/* ── DISCLOSURE: tras la aprobación ───────────────────────────── */}
              <div class="bg-teal-50 p-6 md:p-8 rounded-[2rem] border border-teal-200">
                <div class="flex flex-col sm:flex-row sm:items-center gap-4 mb-4">
                  <div class="w-11 h-11 bg-teal-600 text-white rounded-2xl flex items-center justify-center text-xl font-black flex-shrink-0">
                    ✓
                  </div>
                  <h3 class="text-xl font-black text-colpsi-blue tracking-tight leading-tight">
                    Aprobación y registro del profesional
                  </h3>
                </div>
                <ul class="space-y-3 text-teal-800 text-sm leading-relaxed font-medium">
                  <li>
                    Una vez aprobada la solicitud, su cuenta queda creada en el sistema como psicólogo activo,
                    solvente y con fe de vida vigente; la solvencia y la fe de vida se renuevan periódicamente.
                  </li>
                  <li>
                    La administración confirma la inscripción en el Ministerio del Poder Popular para la Educación
                    Universitaria, la cual se tramita a través de la Federación de Psicólogos de Venezuela (FPV),
                    y el N° de FPV lo asigna la propia Federación al procesar su expediente.
                  </li>
                  <li>
                    Los documentos digitales que adjuntó en el formulario pasan a integrar su expediente
                    profesional (cédula, título, RIF y comprobante de pago).
                  </li>
                </ul>
              </div>

              {/* ── PASO 4 ─────────────────────────────────────────────────────── */}
              <div class="bg-white p-8 md:p-12 rounded-[2.5rem] shadow-sm border border-gray-100 group hover:border-colpsi-yellow transition-colors">
                <div class="flex flex-col sm:flex-row sm:items-center gap-4 mb-8 border-b border-gray-100 pb-6">
                  <div class="w-14 h-14 bg-blue-50 text-blue-600 rounded-2xl flex items-center justify-center text-3xl font-black group-hover:scale-110 transition-transform border border-blue-100 flex-shrink-0">
                    4
                  </div>
                  <h3 class="text-2xl font-black text-colpsi-blue tracking-tight leading-tight">
                    Fase de Confirmación, Lapso de Espera y Cita
                  </h3>
                </div>

                <div class="space-y-6 text-gray-600 text-sm leading-relaxed font-medium text-justify">
                  <p>
                    A partir del momento en que el profesional envía el correo electrónico con todos los datos y recaudos completos, se inicia un lapso de cinco (5) días hábiles.
                  </p>
                  <p>
                    Durante este período, el departamento administrativo del Colegio de Psicólogos del Estado Carabobo realizará la revisión técnica, la validación del pago y la preparación preliminar del expediente. Transcurridos los cinco días hábiles, la institución contactará al solicitante para asignarle la fecha y hora de su cita presencial obligatoria.
                  </p>
                </div>
              </div>

              {/* ── PASO 5 ─────────────────────────────────────────────────────── */}
              <div class="bg-white p-8 md:p-12 rounded-[2.5rem] shadow-sm border border-gray-100 group hover:border-colpsi-yellow transition-colors">
                <div class="flex flex-col sm:flex-row sm:items-center gap-4 mb-8 border-b border-gray-100 pb-6">
                  <div class="w-14 h-14 bg-red-50 text-red-600 rounded-2xl flex items-center justify-center text-3xl font-black group-hover:scale-110 transition-transform border border-red-100 flex-shrink-0">
                    5
                  </div>
                  <h3 class="text-2xl font-black text-colpsi-blue tracking-tight leading-tight">
                    Fase Presencial, Entrega de Expedientes Físicos y Protocolo de Resguardo
                  </h3>
                </div>

                <div class="space-y-8 text-gray-600 text-sm leading-relaxed font-medium text-justify">
                  <p>
                    El día de la cita asignada, el profesional debe acudir a la sede del colegio con sus recaudos organizados de forma estricta. Toda la documentación física (fotocopias, comprobantes, etc.) debe estar impresa obligatoriamente en hojas tamaño carta.
                  </p>
                  <p>
                    Se deben consignar dos expedientes idénticos en su contenido, pero con distinta presentación física y rotulación, organizados de la siguiente manera:
                  </p>

                  {/* Carpeta 1 */}
                  <div>
                    <h4 class="text-base font-black text-colpsi-blue mb-4">
                      A. Carpeta Nº 1 (Destinada a la Federación de Psicólogos de Venezuela - FPV)
                    </h4>
                    <p class="mb-4">
                      Debe ser una carpeta física tipo manila tamaño carta, dirigida a la FPV. Contiene exactamente los siguientes documentos con sus respectivas especificaciones de rotulación en el reverso:
                    </p>
                    <ul class="space-y-4">
                      <li class="bg-gray-50 p-5 rounded-xl border border-gray-100">
                        <strong class="text-gray-900 block mb-1">Una (1) fotografía tipo carnet:</strong> Reciente y en fondo blanco. Por la parte posterior (reverso), debe tener escrito de forma legible a bolígrafo: los nombres completos, los apellidos completos, el número de cédula de identidad y la frase "Estado Carabobo" o "Carabobo".
                      </li>
                      <li class="bg-gray-50 p-5 rounded-xl border border-gray-100">
                        <strong class="text-gray-900 block mb-1">Una (1) planilla de inscripción de la Federación de Psicólogos de Venezuela</strong>
                      </li>
                      <li class="bg-gray-50 p-5 rounded-xl border border-gray-100">
                        <strong class="text-gray-900 block mb-1">Una (1) copia del Título de Psicólogo:</strong> Por la parte posterior (reverso) de la hoja, se deben transcribir de forma clara todos los datos registrales del título (Número de Tomo, Folio y Número de Registro asignado por el Registro Principal).
                      </li>
                      <li class="bg-gray-50 p-5 rounded-xl border border-gray-100">
                        <strong class="text-gray-900 block mb-1">Una (1) copia de la Cédula de Identidad o Pasaporte vigente.</strong>
                      </li>
                      <li class="bg-gray-50 p-5 rounded-xl border border-gray-100">
                        <strong class="text-gray-900 block mb-1">Una (1) copia del RIF actualizado</strong> (con dirección en el Estado Carabobo).
                      </li>
                      <li class="bg-gray-50 p-5 rounded-xl border border-gray-100">
                        <strong class="text-gray-900 block mb-1">El Comprobante de Pago Bancario:</strong> Por la parte posterior (reverso) del papel, debe estar plenamente identificado con los datos de la persona que se está colegiando (Nombres, Apellidos y Cédula).
                      </li>
                    </ul>
                  </div>

                  {/* Expediente 2 */}
                  <div>
                    <h4 class="text-base font-black text-colpsi-blue mb-4">
                      B. Expediente Nº 2 (Destinado al Archivo Regional del Colegio de Psicólogos de Carabobo)
                    </h4>
                    <p>
                      Consiste en un juego de copias exactamente igual al de la carpeta anterior. Contiene los mismos seis recaudos (Foto rotulada al reverso, planilla de inscripción en la Federación de Psicólogos de Venezuela, copia del título con datos registrales al reverso, cédula, RIF de Carabobo y soporte de pago identificado al reverso por el solicitante). La diferencia radica en que todos estos documentos deben introducirse obligatoriamente dentro de un sobre plástico transparente tamaño carta.
                    </p>
                  </div>

                  {/* Protocolo */}
                  <div>
                    <h4 class="text-base font-black text-colpsi-blue mb-4">
                      C. Protocolo de Verificación e Identidad Biométrica (Pasos Finales de la Cita)
                    </h4>
                    <p class="mb-4">
                      Para dar por concluido el trámite de manera conforme, se ejecutarán en el sitio los siguientes dos pasos de seguridad:
                    </p>
                    <ul class="space-y-4">
                      <li class="bg-gray-50 p-5 rounded-xl border border-gray-100">
                        <strong class="text-gray-900 block mb-1">Confrontación del Título Original:</strong> El solicitante debe presentar de forma obligatoria su Título de Psicólogo Original. El personal técnico del colegio lo revisará para verificar su autenticidad y confrontarlo con las copias suministradas.
                      </li>
                      <li class="bg-gray-50 p-5 rounded-xl border border-gray-100">
                        <strong class="text-gray-900 block mb-1">Registro Fotográfico de Seguridad:</strong> Como medida de máxima seguridad, resguardo de la identidad profesional y validación del proceso de colegiatura, el personal institucional tomará una fotografía presencial de la persona que se está colegiando en ese mismo instante.
                      </li>
                    </ul>
                  </div>

                </div>
              </div>
            </div>
          </Show>

          {/* ── VISTA: FORMULARIO DE PRE-INSCRIPCIÓN ─────────────────────── */}
          <Show when={tab() === "formulario"}>
            <InscriptionForm />
          </Show>

          {/* CTA para ir de instrucciones a formulario */}
          <Show when={tab() === "instrucciones"}>
            <div class="text-center pb-4">
              <button
                onClick={() => setTab("formulario")}
                class="inline-flex items-center gap-3 bg-colpsi-yellow text-colpsi-blue px-10 py-5 rounded-2xl font-black text-lg shadow-premium hover:scale-[1.02] active:scale-[0.98] transition-transform"
              >
                <span>📝</span> Ir al Formulario de Inscripción
              </button>
            </div>
          </Show>

        </section>
      </main>
    </>
  );
}