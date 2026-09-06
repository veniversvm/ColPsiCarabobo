// web/src/components/psi/profile/SocialNetworksSection.tsx
import { Show, For, createSignal } from "solid-js";
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
  // Estado local del modal — no sube a perfil.tsx
  const [pendingId, setPendingId] = createSignal<string | null>(null);

  const handleConfirmDelete = () => {
    const id = pendingId();
    if (!id) return;
    props.onDeleteNetwork(id);
    setPendingId(null);
  };

  return (
    <section>

      {/* ── Modal de confirmación ────────────────────────────────────────── */}
      <Show when={pendingId()}>
        <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm animate-in fade-in duration-200">
          <div class="bg-white rounded-3xl p-8 shadow-2xl max-w-sm w-full mx-4 border border-colpsi-border">
            <div class="text-center mb-6">
              <span class="text-4xl">🔗</span>
              <h3 class="text-lg font-black text-gray-800 mt-3">¿Eliminar esta red social?</h3>
              <p class="text-sm text-colpsi-muted mt-1">Se quitará de tu perfil público.</p>
            </div>
            <div class="flex gap-3">
              <button
                onClick={() => setPendingId(null)}
                class="flex-1 bg-gray-100 text-gray-700 py-3 rounded-2xl font-black hover:bg-gray-200 transition-colors"
              >
                Cancelar
              </button>
              <button
                onClick={handleConfirmDelete}
                class="flex-1 bg-red-500 text-white py-3 rounded-2xl font-black hover:bg-red-600 active:scale-95 transition-all"
              >
                Eliminar
              </button>
            </div>
          </div>
        </div>
      </Show>



      <Show when={props.networks && props.networks.length > 0}>
        <div class="mb-8 space-y-3">
          <For each={props.networks}>
            {(net) => (
              <div class="flex items-center justify-between bg-colpsi-surface hover:bg-white p-4 rounded-2xl border border-colpsi-border hover:border-blue-100 transition-colors group">
                <div class="flex items-center gap-3 overflow-hidden">
                  <span class="bg-white px-3 py-1 rounded-xl text-xs font-black text-colpsi-blue shadow-sm border border-colpsi-border">
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
                  onClick={() => net.id && setPendingId(net.id)}
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

      <form onSubmit={props.onAddNetwork} class="bg-blue-50/50 p-6 md:p-8 rounded-3xl border border-blue-100 shadow-inner">
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