// components/admin/staff/RoleSelector.tsx
// Selector de presets de rol (atajos de permisos). Muestra los presets desde
// GET /admin/roles/presets y resalta el que coincide con los permisos actuales.
import { createResource, For, Show } from "solid-js";
import { apiGet } from "~/lib/api";
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
            {() => <div class="h-28 bg-white animate-pulse rounded-2xl border border-gray-100" />}
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
                  class={`text-left rounded-2xl border-2 p-4 transition-all ${
                    isActive()
                      ? "bg-blue-50 border-blue-600 shadow-sm"
                      : "bg-white border-gray-200 hover:border-blue-300"
                  }`}
                >
                  <div class="flex items-center justify-between gap-2 mb-1">
                    <span class={`font-black text-sm ${isActive() ? "text-blue-800" : "text-gray-800"}`}>
                      {preset.name}
                    </span>
                    {isActive() && <span class="text-blue-700 text-xs">●</span>}
                  </div>
                  <p class="text-[11px] text-gray-500 leading-snug mb-2">{preset.description}</p>
                  <span class={`text-[10px] font-black px-2 py-0.5 rounded-lg uppercase tracking-wider ${
                    isActive() ? "bg-blue-100 text-blue-700" : "bg-gray-100 text-gray-500"
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
            class={`text-left rounded-2xl border-2 border-dashed p-4 transition-all ${
              activeSlug() === "personalizado"
                ? "bg-gray-100 border-gray-400"
                : "bg-white border-gray-300 hover:border-gray-400"
            }`}
          >
            <div class="flex items-center justify-between gap-2 mb-1">
              <span class={`font-black text-sm ${activeSlug() === "personalizado" ? "text-gray-800" : "text-gray-500"}`}>
                Personalizado
              </span>
              {activeSlug() === "personalizado" && <span class="text-gray-700 text-xs">●</span>}
            </div>
            <p class="text-[11px] text-gray-500 leading-snug mb-2">
              Sin preset: cada permiso se configura a mano.
            </p>
          </button>
        </div>
      </Show>
    </div>
  );
}