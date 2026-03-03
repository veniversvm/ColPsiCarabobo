// web/src/components/psi/profile/SocialNetworksSection.tsx
import { Show, For } from "solid-js";
import { SocialNetwork } from "~/types/psi";

interface SocialNetworksSectionProps {
  networks?: SocialNetwork[];
  newNetworkName: string;
  newNetworkUrl: string;
  saving: boolean;
  onNetworkNameChange: (value: string) => void;
  onNetworkUrlChange: (value: string) => void;
  onAddNetwork: (e: Event) => void;
  onDeleteNetwork: (id: string) => void;
}

export function SocialNetworksSection(props: SocialNetworksSectionProps) {
  return (
    <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100 mt-12 mb-20">
      <div class="mb-6 border-l-4 border-colpsi-yellow pl-3">
        <h2 class="text-xl font-black text-colpsi-blue leading-tight">Presencia Digital</h2>
      </div>
      
      <Show when={props.networks && props.networks.length > 0}>
        <div class="mb-8 space-y-3">
          <For each={props.networks}>
            {(net) => (
              <div class="flex items-center justify-between bg-gray-50 hover:bg-white p-4 rounded-2xl border border-gray-100 hover:border-blue-100 transition-colors group">
                <div class="flex items-center gap-3 overflow-hidden">
                  <span class="bg-white px-3 py-1 rounded-xl text-xs font-black text-colpsi-blue shadow-sm border border-gray-100">
                    {net.name}
                  </span>
                  <a 
                    href={net.url} 
                    target="_blank" 
                    rel="noopener noreferrer" 
                    class="text-sm text-colpsi-muted hover:text-colpsi-blue truncate max-w-[150px] sm:max-w-md transition-colors"
                  >
                    {net.url}
                  </a>
                </div>
                <button 
                  onClick={() => net.id && props.onDeleteNetwork(net.id)} 
                  class="text-gray-400 hover:text-red-500 hover:bg-red-50 p-2 rounded-xl transition-colors"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
                  </svg>
                </button>
              </div>
            )}
          </For>
        </div>
      </Show>

      <form onSubmit={props.onAddNetwork} class="bg-blue-50/50 p-6 md:p-8 rounded-[2rem] border border-blue-100 shadow-inner">
        <div class="flex flex-col md:flex-row gap-4">
          <input 
            type="text" 
            placeholder="Red (Ej: Instagram)" 
            required 
            value={props.newNetworkName} 
            onInput={(e) => props.onNetworkNameChange(e.currentTarget.value)} 
            class="flex-1 bg-white border-2 border-transparent focus:border-colpsi-blue rounded-xl px-5 py-3 outline-none text-sm text-colpsi-text shadow-sm transition-all" 
          />
          <input 
            type="url" 
            placeholder="Enlace completo" 
            required 
            value={props.newNetworkUrl} 
            onInput={(e) => props.onNetworkUrlChange(e.currentTarget.value)} 
            class="flex-[2] bg-white border-2 border-transparent focus:border-colpsi-blue rounded-xl px-5 py-3 outline-none text-sm text-colpsi-text shadow-sm transition-all" 
          />
          <button 
            type="submit" 
            disabled={props.saving} 
            class="bg-colpsi-blue text-white px-8 py-3 rounded-xl font-bold hover:bg-blue-800 active:scale-95 transition-all shadow-md disabled:opacity-70"
          >
            {props.saving ? "..." : "AÑADIR"}
          </button>
        </div>
      </form>
    </section>
  );
}