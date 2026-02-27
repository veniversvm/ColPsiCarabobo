// web/src/routes/nosotros.tsx

import { A } from "@solidjs/router";
import { Show } from "solid-js";

export default function AboutUs() {
  return (
    <main class="bg-white min-h-screen pb-24">
      {/* Hero sutil de la página */}
      <header class="bg-colpsi-bg py-16 px-4 border-b border-gray-100">
        <div class="max-w-4xl mx-auto text-center">
          <div class="inline-block px-4 py-1.5 bg-blue-100 text-colpsi-blue rounded-full text-xs font-black tracking-widest uppercase mb-4">
            Institución
          </div>
          <h1 class="text-4xl md:text-5xl font-black text-colpsi-text tracking-tighter">
            Nuestra <span class="text-colpsi-blue">Misión y Visión</span>
          </h1>
        </div>
      </header>

      <section class="max-w-4xl mx-auto px-6 py-16 space-y-16">
        {/* Bloque: Quiénes Somos */}
        <div class="grid grid-cols-1 md:grid-cols-2 gap-12 items-center">
          <div class="space-y-6">
            <h2 class="text-2xl font-black text-colpsi-text border-l-4 border-colpsi-yellow pl-4">
              ¿Quiénes Somos?
            </h2>
            <p class="text-colpsi-text text-lg leading-relaxed">
              El Colegio de Psicólogos del Estado Carabobo es una corporación gremial sin fines de lucro, dedicada a la vigilancia del ejercicio profesional ético y la protección de la salud mental de nuestra comunidad.
            </p>
          </div>
          <div class="bg-colpsi-blue p-8 rounded-3xl shadow-2xl shadow-blue-900/20 rotate-1 md:rotate-2">
            <p class="text-blue-100 italic text-lg font-medium">
              "Velamos por la excelencia científica y el compromiso humano de cada profesional colegiado en nuestra región."
            </p>
          </div>
        </div>

        {/* Bloque: Misión y Visión */}
        <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
          <div class="bg-gray-50 p-10 rounded-[2.5rem] border border-gray-100 hover:border-colpsi-blue transition-colors group">
            <div class="w-12 h-12 bg-white rounded-2xl shadow-sm flex items-center justify-center text-2xl mb-6 group-hover:scale-110 transition-transform">🎯</div>
            <h3 class="text-xl font-black text-colpsi-blue mb-4 uppercase">Misión</h3>
            <p class="text-colpsi-muted leading-relaxed">
              Organizar y representar a los psicólogos del estado, garantizando que el ejercicio de la profesión se realice bajo estrictas normas bioéticas y legales.
            </p>
          </div>

          <div class="bg-gray-50 p-10 rounded-[2.5rem] border border-gray-100 hover:border-colpsi-yellow transition-colors group">
            <div class="w-12 h-12 bg-white rounded-2xl shadow-sm flex items-center justify-center text-2xl mb-6 group-hover:scale-110 transition-transform">👁️</div>
            <h3 class="text-xl font-black text-colpsi-blue mb-4 uppercase">Visión</h3>
            <p class="text-colpsi-muted leading-relaxed">
              Ser el referente gremial líder en Venezuela por su transparencia, innovación tecnológica y aporte al bienestar emocional de la sociedad carabobeña.
            </p>
          </div>
        </div>
      </section>

      {/* Footer sutil de la página para llamar a la acción */}
      <div class="max-w-4xl mx-auto px-6 text-center mt-12">
        <A href="/directorio" class="inline-flex items-center gap-2 text-colpsi-blue font-black hover:gap-4 transition-all uppercase text-sm tracking-widest">
          Consultar Directorio de Profesionales <span>→</span>
        </A>
      </div>
    </main>
  );
}