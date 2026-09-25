// web/src/components/admin/settings/ReceptionSwitchesCard.tsx
// Interruptores globales de recepción (solo SUDO): controlan si se pueden
// abrir tickets de solicitudes y si la pre-inscripción está habilitada.
import { createResource, createSignal, For, Show } from "solid-js";
import { apiGet, apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { Panel } from "~/components/admin/ui/Panel";
import { Button } from "~/components/admin/ui/Button";
import { Badge } from "~/components/admin/ui/Badge";
import { Icon } from "~/components/admin/ui/icons";

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
    <Panel class={props.class ?? ""}>
      <div class="flex items-center gap-3 mb-4">
        <span class="h-8 w-8 rounded-md bg-colpsi-bg text-colpsi-blue flex items-center justify-center">
          <Icon name="sliders" />
        </span>
        <div>
          <h2 class="text-base font-semibold text-colpsi-text">Recepción global</h2>
          <p class="text-xs text-colpsi-muted mt-0.5">
            Activa o pausa la entrada de solicitudes al instante. Solo visible para SUDO.
          </p>
        </div>
      </div>

      <Show when={!loaded()}>
        <div class="space-y-3">
          <div class="h-16 bg-colpsi-surface animate-pulse rounded-md" />
          <div class="h-16 bg-colpsi-surface animate-pulse rounded-md" />
        </div>
      </Show>

      <Show when={loaded()}>
        <Show when={me()?.sudo}>
          <div class="divide-y divide-colpsi-border">
            <For each={rows}>
              {(row) => (
                <div class="py-4 first:pt-0 last:pb-0 space-y-3">
                  <div class="flex items-start justify-between gap-3">
                    <div>
                      <p class="font-semibold text-colpsi-text text-sm">{row.title}</p>
                      <p class="text-colpsi-muted text-xs mt-0.5">{row.desc}</p>
                    </div>
                    <button
                      onClick={() => patch(row.key, !switches()[row.key].enabled, switches()[row.key].message)}
                      class={`relative inline-flex shrink-0 w-10 h-6 rounded-full transition-colors ${
                        switches()[row.key].enabled ? "bg-emerald-500" : "bg-slate-300"
                      }`}
                      title={switches()[row.key].enabled ? "Desactivar" : "Activar"}
                    >
                      <span
                        class={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-sm transition-transform ${
                          switches()[row.key].enabled ? "translate-x-4" : ""
                        }`}
                      />
                    </button>
                  </div>

                  <Show when={!switches()[row.key].enabled}>
                    <div>
                      <label class="block text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide mb-1">
                        Mensaje público (máx. 500)
                      </label>
                      <textarea
                        rows={2}
                        maxLength={500}
                        value={switches()[row.key].message}
                        onInput={(e) => patch(row.key, switches()[row.key].enabled, e.currentTarget.value)}
                        placeholder="Ej: Reanudamos la recepción el 20 de este mes."
                        class="w-full px-3 py-2 rounded-md border border-slate-300 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 outline-none text-sm text-colpsi-text"
                      />
                    </div>
                  </Show>

                  <div class="flex items-center justify-between gap-3">
                    <Badge tone={switches()[row.key].enabled ? "success" : "warning"}>
                      {switches()[row.key].enabled ? "Habilitado" : "Pausado"}
                    </Badge>
                    <div class="flex items-center gap-2">
                      <Show when={savedKey() === row.key}>
                        <span class="text-[11px] font-semibold text-emerald-600">✓ Guardado</span>
                      </Show>
                      <Button
                        variant="primary"
                        size="sm"
                        onClick={() => save(row.key)}
                        disabled={saving() !== ""}
                      >
                        {saving() === row.key ? "Guardando..." : "Guardar"}
                      </Button>
                    </div>
                  </div>
                </div>
              )}
            </For>
          </div>
        </Show>
        <Show when={!me()?.sudo}>
          <p class="text-colpsi-muted text-xs font-medium bg-colpsi-surface rounded-md p-3">
            Solo el Super Usuario puede gestionar los interruptores de recepción.
          </p>
        </Show>
      </Show>

      <Show when={error()}>
        <div class="bg-red-50 border border-red-200 text-red-700 text-sm font-medium rounded-md p-3 mt-3">{error()}</div>
      </Show>
    </Panel>
  );
}