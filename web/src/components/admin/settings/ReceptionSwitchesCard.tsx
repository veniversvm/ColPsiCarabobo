// web/src/components/admin/settings/ReceptionSwitchesCard.tsx
// Interruptores globales de recepción (solo SUDO): controlan si se pueden
// abrir tickets de solicitudes y si la pre-inscripción está habilitada.
import { createResource, createSignal, For, Show } from "solid-js";
import { apiGet, apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";

type ReceptionSetting = { enabled: boolean; message: string };
type Switches = { tickets: ReceptionSetting; inscriptions: ReceptionSetting };

interface Props {
  class?: string;
}

export function ReceptionSwitchesCard(props: Props) {
  const [me] = createResource<{ sudo: boolean }>(async () => {
    try {
      return (await apiGet("/admin/me")) ?? { sudo: false };
    } catch {
      return { sudo: false };
    }
  });

  const [switches, setSwitches] = createSignal<Switches>({
    tickets: { enabled: true, message: "" },
    inscriptions: { enabled: true, message: "" },
  });
  const [loaded, setLoaded] = createSignal(false);
  const [saving, setSaving] = createSignal<"" | "tickets" | "inscriptions">("");
  const [error, setError] = createSignal("");
  const [savedKey, setSavedKey] = createSignal<"" | "tickets" | "inscriptions">("");

  const load = async () => {
    try {
      const res = await apiGet<Switches>("/admin/settings/reception");
      if (res) setSwitches(res);
    } catch {
      /* el GET es best-effort */
    } finally {
      setLoaded(true);
    }
  };
  void load();

  const patch = (key: "tickets" | "inscriptions", enabled: boolean, message: string) => {
    setSwitches((s) => ({ ...s, [key]: { enabled, message } }));
  };

  const save = async (key: "tickets" | "inscriptions") => {
    setSaving(key);
    setError("");
    setSavedKey("");
    const target = switches()[key];
    try {
      const res = await apiPost<Switches>("/admin/settings/reception", {
        key: key === "tickets" ? "tickets.reception_enabled" : "inscriptions.reception_enabled",
        enabled: target.enabled,
        message: target.message,
      });
      if (res) setSwitches(res);
      setSavedKey(key);
    } catch (e: any) {
      if (e?.status === 403) {
        setError("Solo el Super Usuario puede modificar esta configuración.");
      } else {
        setError(getUserFacingError(e));
      }
    } finally {
      setSaving("");
    }
  };

  const rows: { key: "tickets" | "inscriptions"; title: string; desc: string }[] = [
    { key: "tickets", title: "Recepción de solicitudes (portal psi)", desc: "Permite a los psicólogos abrir nuevos tickets de trámite." },
    { key: "inscriptions", title: "Recepción de inscripciones", desc: "Permite a nuevos colegiados enviar la pre-inscripción en línea." },
  ];

  return (
    <section class={`bg-white rounded-3xl shadow-sm border border-gray-100 p-6 md:p-8 space-y-5 ${props.class ?? ""}`}>
      <div class="flex items-center gap-3">
        <span class="text-2xl">🎛️</span>
        <div>
          <h2 class="text-lg font-black text-colpsi-blue">Recepción global</h2>
          <p class="text-gray-400 text-xs mt-0.5">
            Activa o pausa la entrada de solicitudes al instante. Solo visible para SUDO.
          </p>
        </div>
      </div>

      <Show when={!loaded()}>
        <div class="space-y-3">
          <div class="h-16 bg-gray-50 animate-pulse rounded-2xl" />
          <div class="h-16 bg-gray-50 animate-pulse rounded-2xl" />
        </div>
      </Show>

      <Show when={loaded()}>
        <Show when={me()?.sudo}>
          <For each={rows}>
            {(row) => (
              <div class="border-2 border-gray-100 rounded-2xl p-4 space-y-3">
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <p class="font-black text-gray-800 text-sm">{row.title}</p>
                    <p class="text-gray-400 text-xs mt-0.5">{row.desc}</p>
                  </div>
                  <button
                    onClick={() => patch(row.key, !switches()[row.key].enabled, switches()[row.key].message)}
                    class={`relative inline-flex shrink-0 w-12 h-7 rounded-full transition-colors ${
                      switches()[row.key].enabled ? "bg-emerald-500" : "bg-gray-300"
                    }`}
                    title={switches()[row.key].enabled ? "Desactivar" : "Activar"}
                  >
                    <span
                      class={`absolute top-0.5 left-0.5 w-6 h-6 bg-white rounded-full shadow transition-transform ${
                        switches()[row.key].enabled ? "translate-x-5" : ""
                      }`}
                    />
                  </button>
                </div>

                <Show when={!switches()[row.key].enabled}>
                  <div>
                    <label class="block text-[10px] font-black text-gray-500 uppercase tracking-widest mb-1">
                      Mensaje público (máx. 500)
                    </label>
                    <textarea
                      rows={2}
                      maxLength={500}
                      value={switches()[row.key].message}
                      onInput={(e) => patch(row.key, switches()[row.key].enabled, e.currentTarget.value)}
                      placeholder="Ej: Reanudamos la recepción el 20 de este mes."
                      class="w-full px-3 py-2 rounded-xl border-2 border-gray-200 focus:border-amber-500 outline-none text-sm text-gray-700"
                    />
                  </div>
                </Show>

                <div class="flex items-center justify-between gap-3">
                  <span
                    class={`inline-flex items-center gap-1.5 text-[10px] font-black uppercase tracking-widest px-3 py-1 rounded-full ${
                      switches()[row.key].enabled
                        ? "bg-emerald-50 text-emerald-700"
                        : "bg-amber-50 text-amber-700"
                    }`}
                  >
                    <span class={`w-1.5 h-1.5 rounded-full ${switches()[row.key].enabled ? "bg-emerald-500" : "bg-amber-500"}`} />
                    {switches()[row.key].enabled ? "Habilitado" : "Pausado"}
                  </span>
                  <div class="flex items-center gap-2">
                    <Show when={savedKey() === row.key}>
                      <span class="text-[11px] font-bold text-emerald-600">✓ Guardado</span>
                    </Show>
                    <button
                      onClick={() => save(row.key)}
                      disabled={saving() !== ""}
                      class="px-4 py-2 rounded-xl bg-colpsi-blue hover:bg-blue-900 disabled:opacity-50 text-white font-black text-xs transition-all"
                    >
                      {saving() === row.key ? "Guardando..." : "Guardar"}
                    </button>
                  </div>
                </div>
              </div>
            )}
          </For>
        </Show>
        <Show when={!me()?.sudo}>
          <p class="text-gray-400 text-xs font-semibold bg-gray-50 rounded-2xl p-4">
            Solo el Super Usuario puede gestionar los interruptores de recepción.
          </p>
        </Show>
      </Show>

      <Show when={error()}>
        <div class="bg-red-50 border border-red-200 text-red-700 text-sm font-semibold rounded-2xl p-3">{error()}</div>
      </Show>
    </section>
  );
}