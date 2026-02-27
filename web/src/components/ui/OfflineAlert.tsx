export default function OfflineAlert(props: { error: any, reset: () => void }) {
  return (
    <div class="fixed inset-0 z-[100] flex items-center justify-center bg-white/80 backdrop-blur-sm px-6">
      <div class="w-full max-w-sm bg-white rounded-[2.5rem] p-8 shadow-2xl border border-gray-100 text-center space-y-6 animate-in fade-in zoom-in duration-300">
        <div class="w-20 h-20 bg-colpsi-bg rounded-3xl mx-auto flex items-center justify-center text-4xl shadow-inner border border-blue-50">
          <span class="animate-pulse">📡</span>
        </div>
        
        <div class="space-y-2">
          <h2 class="text-colpsi-blue font-black text-xl uppercase tracking-tighter">
            Conexión en pausa
          </h2>
          <p class="text-colpsi-muted text-sm leading-relaxed">
            Estamos teniendo dificultades para conectar con el servidor del Colegio. 
            Por favor, verifique su internet o intente de nuevo.
          </p>
        </div>

        <button 
          onClick={() => props.reset()}
          class="w-full bg-colpsi-yellow text-colpsi-blue py-4 rounded-2xl font-black text-sm shadow-lg shadow-yellow-500/20 active:scale-95 transition-all"
        >
          REINTENTAR AHORA
        </button>

        <p class="text-[9px] text-gray-300 font-mono tracking-widest uppercase">
          ID Error: {props.error?.status || 500}
        </p>
      </div>
    </div>
  );
}