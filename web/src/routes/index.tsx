// web/src/routes/index.tsx

import { A } from "@solidjs/router";

export default function Home() {
  return (
    <main class="relative min-h-[90vh] flex flex-col items-center justify-center bg-white px-4 text-center overflow-hidden">
      
      {/* SECCIÓN DE IDENTIDAD */}
      <div class="mb-10 relative">
        <div class="bg-colpsi-blue w-20 h-20 rounded-3xl flex items-center justify-center shadow-2xl shadow-blue-900/20">
          <span class="text-white text-5xl font-bold select-none">Ψ</span>
        </div>
        <div class="absolute -top-2 -right-2 w-7 h-7 bg-colpsi-yellow rounded-full border-4 border-white animate-pulse shadow-md" />
      </div>

      <header class="max-w-4xl space-y-4">
        <h1 class="text-4xl md:text-6xl font-black text-colpsi-text leading-tight tracking-tighter">
          Colegio de Psicólogos <br />
          <span class="text-colpsi-blue italic font-black">Estado Carabobo</span>
        </h1>
        <p class="text-colpsi-muted text-lg md:text-xl max-w-2xl mx-auto leading-relaxed font-medium">
          Garantizando la ética, el respaldo profesional y la salud mental de todos los carabobeños.
        </p>
      </header>

      {/* ACCIONES DE ENTRADA */}
      <div class="mt-12 flex flex-col md:flex-row gap-4 w-full md:w-auto z-10 px-4">
        {/* Entrada para el Público General */}
        <A 
          href="/explorar" 
          class="bg-colpsi-yellow text-colpsi-blue px-16 py-5 rounded-2xl font-black shadow-lg shadow-yellow-500/30 hover:scale-105 active:scale-95 transition-all flex items-center justify-center text-lg"
        >
          ENTRAR
        </A>
        
        {/* Entrada para Profesionales */}
        <A 
          href="/login" 
          class="border-2 border-colpsi-blue text-colpsi-blue px-10 py-5 rounded-2xl font-black hover:bg-blue-50 active:scale-95 transition-all text-center text-sm md:text-base"
        >
          INGRESO AGREMIADOS
        </A>
      </div>

      {/* LA FRANJA DINÁMICA */}
      <div class="fixed bottom-0 left-0 w-full h-3 flex overflow-hidden shadow-[0_-4px_15px_rgba(0,0,0,0.1)]">
        <div class="relative flex-1 bg-colpsi-red">
          <div class="absolute inset-0 bg-linear-to-r from-transparent via-white/50 to-transparent animate-flag-flow" />
        </div>
        <div class="relative flex-1 bg-green-700">
          <div class="absolute inset-0 bg-linear-to-r from-transparent via-white/50 to-transparent animate-flag-flow [animation-delay:1s]" />
        </div>
        <div class="relative flex-1 bg-colpsi-blue">
          <div class="absolute inset-0 bg-linear-to-r from-transparent via-white/50 to-transparent animate-flag-flow [animation-delay:2s]" />
        </div>
      </div>

      {/* RESPLANDOR DE FONDO */}
      <div class="absolute bottom-0 left-1/2 -translate-x-1/2 w-[140%] h-[400px] bg-blue-50/50 blur-[120px] rounded-full -z-10" />
    </main>
  );
}