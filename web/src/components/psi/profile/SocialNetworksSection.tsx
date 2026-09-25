// web/src/components/psi/profile/SocialNetworksSection.tsx
import { Show, For, createSignal } from "solid-js";
import { SocialNetwork } from "~/types/psi";
import { Icon } from "~/components/admin/ui/icons";

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
        <div class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 backdrop-blur-sm animate-in fade-in duration-200">
          <div class="bg-white rounded-lg p-5 shadow-lg max-w-sm w-full mx-4 border border-colpsi-border">
            <div class="flex items-center gap-3 mb-4">
              <span class="w-9 h-9 rounded-full bg-colpsi-bg flex items-center justify-center text-colpsi-blue">
                <Icon name="link" class="w-4 h-4" />
              </span>
              <div>
                <h3 class="text-base font-semibold text-colpsi-text">¿Eliminar esta red social?</h3>
                <p class="text-sm text-colpsi-muted">Se quitará de tu perfil público.</p>
              </div>
            </div>
            <div class="flex gap-3">
              <button
                onClick={() => setPendingId(null)}
                class="h-9 flex-1 rounded-md bg-colpsi-bg text-colpsi-text text-sm font-medium hover:bg-slate-100 transition-colors"
              >
                Cancelar
              </button>
              <button
                onClick={handleConfirmDelete}
                class="h-9 flex-1 rounded-md bg-red-500 text-white text-sm font-medium hover:bg-red-600 transition-colors"
              >
                Eliminar
              </button>
            </div>
          </div>
        </div>
      </Show>

      <Show when={props.networks && props.networks.length > 0}>
        <div class="mb-6 space-y-2">
          <For each={props.networks}>
            {(net) => (
              <div class="flex items-center justify-between p-3 rounded-md border border-colpsi-border hover:bg-colpsi-bg/50 transition-colors group">
                <div class="flex items-center gap-3 overflow-hidden">
                  <span class="bg-colpsi-bg px-2.5 py-1 rounded-md text-xs font-semibold text-colpsi-blue border border-colpsi-border">
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
                  class="text-slate-400 hover:text-red-500 hover:bg-red-50 p-1.5 rounded-md transition-colors"
                  title="Eliminar red social"
                >
                  <Icon name="trash" class="w-4 h-4" />
                </button>
              </div>
            )}
          </For>
        </div>
      </Show>

      <form onSubmit={props.onAddNetwork} class="bg-colpsi-bg/50 p-4 rounded-md border border-colpsi-border">
        <div class="flex flex-col md:flex-row gap-3">
          <input
            type="text"
            placeholder="Red (Ej: Instagram)"
            required
            value={props.newNetworkName}
            onInput={(e) => props.onNetworkNameChange(e.currentTarget.value)}
            class="h-9 flex-1 rounded-md border border-slate-300 bg-white px-3 text-sm text-colpsi-text placeholder:text-slate-400 outline-none transition-colors focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15"
          />
          <input
            type="url"
            placeholder="Enlace completo"
            required
            value={props.newNetworkUrl}
            onInput={(e) => props.onNetworkUrlChange(e.currentTarget.value)}
            class="h-9 flex-[2] rounded-md border border-slate-300 bg-white px-3 text-sm text-colpsi-text placeholder:text-slate-400 outline-none transition-colors focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15"
          />
          <button
            type="submit"
            disabled={props.saving}
            class="h-9 px-4 rounded-md bg-colpsi-blue text-white text-sm font-semibold hover:bg-colpsi-blue-light transition-colors disabled:opacity-70 disabled:cursor-not-allowed"
          >
            {props.saving ? "..." : "Agregar"}
          </button>
        </div>
      </form>
    </section>
  );
}