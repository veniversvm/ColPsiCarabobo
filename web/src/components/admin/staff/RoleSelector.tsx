// components/admin/staff/RoleSelector.tsx
// Selector de presets de rol (atajos de permisos). Muestra los presets desde
// GET /admin/roles/presets y resalta el que coincide con los permisos actuales.
import { createResource, For, Show } from "solid-js";
import { apiGet } from "~/lib/api";
import { Icon } from "~/components/admin/ui/icons";
import {
  findRoleForPerms,
  type PermissionState,
  type RolePreset,
} from "~/lib/staff-permissions";

interface Props {
  perms: PermissionState;
  storedRole?: string | null;
  onSelect: (preset: RolePreset) => void;
  onClear: () => void;
}

export default function RoleSelector(props: Props) {
  const [presets] = createResource<RolePreset[]>(async () => {
    try {
      return await apiGet<RolePreset[]>("/admin/roles/presets");
    } catch {
      return [];
    }
  });

  const activeSlug = () => {
    const ps = presets();
    if (!ps || ps.length === 0) return props.storedRole ?? null;
    return findRoleForPerms(props.perms, ps) ?? "personalizado";
  };

  const permCount = (p: RolePreset) =>
    Object.values(p.permissions).filter(Boolean).length;

  return (
    <div class="space-y-3">
      <Show when={presets.loading}>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          <For each={[1, 2, 3]}>
            {() => <div class="h-28 bg-white animate-pulse rounded-lg border border-colpsi-border" />}
          </For>
        </div>
      </Show>

      <Show when={!presets.loading && presets()?.length}>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          <For each={presets()}>
            {(preset) => {
              const isActive = () => activeSlug() === preset.slug;
              return (
                <button
                  type="button"
                  onClick={() => props.onSelect(preset)}
                  class={`text-left rounded-lg border p-4 transition-all ${
                    isActive()
                      ? "bg-blue-50 border-colpsi-blue"
                      : "bg-white border-colpsi-border hover:border-colpsi-blue/50"
                  }`}
                >
                  <div class="flex items-center justify-between gap-2 mb-1">
                    <span class={`font-semibold text-sm ${isActive() ? "text-colpsi-blue" : "text-colpsi-text"}`}>
                      {preset.name}
                    </span>
                    <Show when={isActive()}><Icon name="check" class="w-4 h-4 text-colpsi-blue shrink-0" /></Show>
                  </div>
                  <p class="text-[11px] text-colpsi-muted leading-snug mb-2">{preset.description}</p>
                  <span class={`text-[11px] font-semibold px-2 py-0.5 rounded-md uppercase tracking-wide ${
                    isActive() ? "bg-blue-100 text-colpsi-blue" : "bg-colpsi-bg text-colpsi-muted"
                  }`}>
                    {permCount(preset)}/{Object.keys(preset.permissions).length} permisos
                  </span>
                </button>
              );
            }}
          </For>

          <button
            type="button"
            onClick={props.onClear}
            class={`text-left rounded-lg border border-dashed p-4 transition-all ${
              activeSlug() === "personalizado"
                ? "bg-colpsi-bg border-colpsi-blue/60"
                : "bg-white border-colpsi-border hover:border-colpsi-blue/50"
            }`}
          >
            <div class="flex items-center justify-between gap-2 mb-1">
              <span class={`font-semibold text-sm ${activeSlug() === "personalizado" ? "text-colpsi-text" : "text-colpsi-muted"}`}>
                Personalizado
              </span>
              <Show when={activeSlug() === "personalizado"}><Icon name="check" class="w-4 h-4 text-colpsi-blue shrink-0" /></Show>
            </div>
            <p class="text-[11px] text-colpsi-muted leading-snug mb-2">
              Sin preset: cada permiso se configura a mano.
            </p>
          </button>
        </div>
      </Show>
    </div>
  );
}