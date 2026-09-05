// web/src/routes/[...404].tsx

import { A } from "@solidjs/router";
import { HttpStatusCode } from "@solidjs/start";

export default function NotFound() {
  return (
    <main class="min-h-[90vh] flex flex-col items-center justify-center bg-white px-6 text-center">
      {/* Informamos al servidor que el status code debe ser 404 (SEO Friendly) */}
      <HttpStatusCode code={404} />

      {/* IMAGEN TEMÁTICA */}
      <div class="max-w-7xl w-full mb-8 animate-in fade-in zoom-in duration-700">
        <img 
          src="/psi404.png" 
          alt="404 - Página no encontrada" 
          class="w-full h-auto object-contain"
        />
      </div>

      {/* TEXTO INFORMATIVO */}
      <div class="space-y-4 max-w-lg">
        <h1 class="text-3xl md:text-4xl font-black text-colpsi-blue tracking-tighter uppercase">
          ¿Perdido en el inconsciente?
        </h1>
        <p class="text-colpsi-muted text-base md:text-lg leading-relaxed">
          La página que busca no se encuentra en nuestro registro. 
          Tal vez se movió a una nueva ubicación o el enlace es incorrecto.
        </p>
      </div>

      {/* ACCIÓN DE RETORNO */}
      <div class="mt-10 w-full md:w-auto">
        <A 
          href="/" 
          class="inline-flex items-center justify-center w-full md:w-auto bg-colpsi-yellow text-colpsi-blue px-10 py-4 rounded-2xl font-black shadow-lg shadow-yellow-500/20 hover:scale-105 active:scale-95 transition-all gap-2"
        >
          <span>🏠</span> VOLVER AL INICIO
        </A>
      </div>

      {/* LA FRANJA DE LA BANDERA (Para mantener consistencia de marca) */}
      <div class="fixed bottom-0 left-0 w-full h-2 flex overflow-hidden">
        <div class="flex-1 bg-colpsi-red"></div>
        <div class="flex-1 bg-green-700"></div>
        <div class="flex-1 bg-colpsi-blue"></div>
      </div>
    </main>
  );
}