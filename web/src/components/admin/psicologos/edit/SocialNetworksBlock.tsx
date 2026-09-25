// web/src/components/admin/psicologos/edit/SocialNetworksBlock.tsx

import { For, Show, createSignal } from "solid-js";
import { createStore } from "solid-js/store";
import type { PsiProfile, SocialNetwork } from "./types";

interface Props {
  profile: PsiProfile | undefined;
  onAdd: (payload: { name: string; url: string }) => Promise<void>;
  onDelete: (socialId: string) => Promise<void>;
}

export function SocialNetworksBlock(props: Props) {
  const [socialForm, setSocialForm] = createStore({ name: "", url: "" });
  const [saving, setSaving] = createSignal(false);

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    if (!socialForm.name || !socialForm.url) return;
    setSaving(true);
    try {
      await props.onAdd({ name: socialForm.name, url: socialForm.url });
      setSocialForm({ name: "", url: "" });
    } finally {
      setSaving(false);
    }
  };

  return (
    <div>

      {/* Listado existente */}
      <Show when={(props.profile?.social_networks?.length ?? 0) > 0}>
        <div class="mb-6 space-y-3">
          <For each={props.profile?.social_networks}>
            {(net: SocialNetwork) => (
              <div class="flex items-center justify-between bg-colpsi-bg hover:bg-white p-4 rounded-lg border border-colpsi-border hover:border-blue-100 transition-colors group">
                <div class="flex items-center gap-3 overflow-hidden">
                  <span class="bg-white px-3 py-1 rounded-xl text-xs font-semibold text-colpsi-blue shadow-sm border border-colpsi-border">
                    {net.name}
                  </span>
                  <a
                    href={net.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    class="text-sm text-colpsi-muted hover:text-blue-600 truncate max-w-xs transition-colors"
                  >
                    {net.url}
                  </a>
                </div>
                <button
                  onClick={() => net.id && props.onDelete(net.id)}
                  class="text-colpsi-muted hover:text-red-500 hover:bg-red-50 p-2 rounded-xl transition-colors"
                  title="Eliminar"
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

      <Show when={!props.profile?.social_networks?.length}>
        <p class="text-sm italic text-colpsi-muted mb-5">Ninguna red social registrada.</p>
      </Show>

      {/* Formulario de alta */}
      <form onSubmit={handleSubmit} class="bg-blue-50/50 p-5 rounded-lg border border-blue-100">
        <div class="flex flex-col md:flex-row gap-4">
          <input
            type="text"
            placeholder="Red (Ej: Instagram)"
            required
            value={socialForm.name}
            onInput={(e) => setSocialForm("name", e.currentTarget.value)}
            class="flex-1 bg-white border-2 border-transparent focus:border-colpsi-blue rounded-xl px-4 py-3 outline-none text-sm shadow-sm transition-all"
          />
          <input
            type="url"
            placeholder="Enlace completo (https://...)"
            required
            value={socialForm.url}
            onInput={(e) => setSocialForm("url", e.currentTarget.value)}
            class="flex-[2] bg-white border-2 border-transparent focus:border-colpsi-blue rounded-xl px-4 py-3 outline-none text-sm shadow-sm transition-all"
          />
          <button
            type="submit"
            disabled={saving()}
            class="bg-colpsi-blue text-white px-8 py-3 rounded-xl font-bold hover:bg-colpsi-blue active:scale-95 transition-all shadow-md disabled:opacity-70"
          >
            {saving() ? "..." : "AÑADIR"}
          </button>
        </div>
      </form>
    </div>
  );
}