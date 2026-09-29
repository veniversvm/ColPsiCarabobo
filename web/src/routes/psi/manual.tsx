// web/src/routes/psi/manual.tsx
import { A } from "@solidjs/router";
import ManualPdf from "~/components/manuales/ManualPdf";

export default function ManualDelPortalPage() {
  return (
    <main class="bg-colpsi-bg min-h-screen pb-24 font-sans">
      {/* Cabecera */}
      <div class="bg-heraldic pt-12 pb-20 px-4 md:px-8 shadow-inner">
        <div class="max-w-4xl mx-auto flex items-center justify-between">
          <A
            href="/psi"
            class="bg-colpsi-yellow text-colpsi-blue px-5 py-2.5 rounded-full font-black text-sm shadow-lg hover:bg-colpsi-yellow/90 active:scale-95 transition-all inline-flex items-center gap-2"
          >
            <span>←</span> Volver al Panel
          </A>
        </div>
        <div class="max-w-4xl mx-auto mt-8">
          <h1 class="text-white text-3xl font-black">Manual del Portal</h1>
          <p class="text-blue-200 mt-1 uppercase tracking-widest font-black text-[11px]">
            Guía de uso del agremiado
          </p>
        </div>
      </div>

      <div class="max-w-4xl mx-auto px-4 md:px-8 -mt-10 space-y-6">
        {/* Visor + descarga */}
        <div class="bg-white rounded-3xl p-5 md:p-6 shadow-sm border border-colpsi-border">
          <ManualPdf file="manual-psiuser.pdf" />
        </div>

        {/* Resumen de contenido */}
        <div class="bg-white rounded-3xl p-5 md:p-6 shadow-sm border border-colpsi-border">
          <h2 class="font-black text-colpsi-blue mb-2">¿Qué encontrarás en el manual?</h2>
          <ul class="text-sm text-colpsi-muted space-y-1.5 list-disc list-inside">
            <li>Cómo ingresar al portal y recuperar tu contraseña.</li>
            <li>Tu perfil profesional: identidad, contacto, ubicación y privacidad.</li>
            <li>Postgrados, documentos del expediente y modalidad de servicio.</li>
            <li>Notificaciones gremiales, solicitudes (tickets) y biblioteca virtual.</li>
          </ul>
        </div>
      </div>
    </main>
  );
}