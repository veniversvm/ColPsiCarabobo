// web/src/components/admin/auditoria/AuditLogDrawer.tsx
// Drawer lateral de detalle de un evento de la bitácora: muestra quién, cuándo,
// desde dónde, y el diff por campo con colores (from → to).

import { createMemo, Show, For } from "solid-js";
import type { ApiChangeLog, AuditChange } from "~/types/audit";
import { actionLabel, entityLabel, fieldLabel, parseAuditJson } from "~/types/audit";
import { Icon } from "~/components/admin/ui/icons";

interface Props {
  log: ApiChangeLog | null;
  onClose: () => void;
}

const fmtBool = (v: unknown): string => (typeof v === "boolean" ? (v ? "✓ Sí" : "✗ No") : String(v ?? "—"));

const fmtDate = (iso: string): string => {
  try {
    return new Date(iso).toLocaleString("es-VE", { day: "2-digit", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit" });
  } catch {
    return iso;
  }
};

export function AuditLogDrawer(props: Props) {
  const log = () => props.log;

  // Parseo defensivo de cambios/metadata: la API los manda como objeto (jsonb)
  // o como string serializado (historias previas al fix del tipo).
  const changesMap = createMemo<Record<string, AuditChange>>(() => {
    const l = log();
    if (!l?.changes) return {};
    return parseAuditJson(l.changes) as Record<string, AuditChange>;
  });

  const metadataMap = createMemo<Record<string, unknown>>(() => {
    const l = log();
    if (!l?.metadata) return {};
    return parseAuditJson(l.metadata);
  });

  return (
    <Show when={log()}>
      {(l) => (
        <div class="fixed inset-0 z-[90] flex justify-end">
          {/* Backdrop */}
          <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" onClick={props.onClose} />

          {/* Panel */}
          <aside class="relative w-full max-w-lg h-full bg-white shadow-2xl overflow-y-auto animate-in slide-in-from-right duration-300">
            {/* Header */}
            <div class="sticky top-0 z-10 bg-white border-b border-colpsi-border px-5 py-3.5 flex items-start justify-between gap-3">
              <div class="min-w-0">
                <p class="text-[10px] font-semibold uppercase tracking-wide text-colpsi-muted">
                  {entityLabel(l().entity)} · {actionLabel(l().action)}
                </p>
                <h3 class="text-lg font-semibold text-colpsi-text truncate mt-0.5">{l().entity_label || l().entity_id}</h3>
              </div>
              <button
                onClick={props.onClose}
                class="flex-shrink-0 inline-flex items-center justify-center h-8 w-8 rounded-md bg-colpsi-bg hover:bg-colpsi-border/60 text-colpsi-muted hover:text-colpsi-blue transition-colors"
              >
                <Icon name="x" class="w-4 h-4" />
              </button>
            </div>

            <div class="px-5 py-5 space-y-5">
              {/* ── Meta ── */}
              <dl class="grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
                <div><dt class="text-[10px] font-semibold uppercase tracking-wide text-colpsi-muted">Fecha</dt><dd class="font-medium text-colpsi-text">{fmtDate(l().created_at)}</dd></div>
                <div><dt class="text-[10px] font-semibold uppercase tracking-wide text-colpsi-muted">Actor</dt><dd class="font-medium text-colpsi-text">{l().actor_username || "—"} <span class="text-colpsi-muted">({l().actor_role || "—"})</span></dd></div>
                <div><dt class="text-[10px] font-semibold uppercase tracking-wide text-colpsi-muted">IP</dt><dd class="font-medium text-colpsi-text">{l().ip || "—"}</dd></div>
                <div><dt class="text-[10px] font-semibold uppercase tracking-wide text-colpsi-muted">Entidad ID</dt><dd class="font-medium text-colpsi-text truncate">{l().entity_id || "—"}</dd></div>
              </dl>
              <Show when={l().user_agent}>
                <div class="bg-colpsi-bg rounded-md px-3 py-2 text-[11px] text-colpsi-muted break-all font-medium border border-colpsi-border">
                  {l().user_agent}
                </div>
              </Show>

              {/* ── Diff ── */}
              <Show when={Object.keys(changesMap()).length > 0}>
                <section>
                  <h4 class="text-xs font-semibold uppercase tracking-wide text-colpsi-muted mb-2.5">Cambios</h4>
                  <div class="space-y-2">
                    <For each={Object.entries(changesMap())}>
                      {([key, chg]) => (
                        <div class="rounded-md border border-colpsi-border overflow-hidden">
                          <p class="bg-colpsi-bg px-3 py-1.5 text-[10px] font-semibold uppercase tracking-wide text-colpsi-muted border-b border-colpsi-border">
                            {fieldLabel(key)}
                          </p>
                          <div class="grid grid-cols-2 divide-x divide-colpsi-border">
                            <div class="px-3 py-2.5 bg-red-50/60 text-colpsi-red text-sm font-medium break-words">
                              <span class="block text-[9px] font-semibold uppercase tracking-wide text-colpsi-red/70 mb-0.5">Antes</span>
                              {fmtBool(chg.from)}
                            </div>
                            <div class="px-3 py-2.5 bg-emerald-50/60 text-emerald-700 text-sm font-medium break-words">
                              <span class="block text-[9px] font-semibold uppercase tracking-wide text-emerald-600/70 mb-0.5">Después</span>
                              {fmtBool(chg.to)}
                            </div>
                          </div>
                        </div>
                      )}
                    </For>
                  </div>
                </section>
              </Show>
              <Show when={Object.keys(changesMap()).length === 0}>
                <p class="text-sm text-colpsi-muted">Sin campos detallados para este suceso.</p>
              </Show>

              {/* ── Metadata ── */}
              <Show when={Object.keys(metadataMap()).length > 0}>
                <section>
                  <h4 class="text-xs font-semibold uppercase tracking-wide text-colpsi-muted mb-2.5">Metadatos</h4>
                  <dl class="space-y-1.5 text-sm">
                    <For each={Object.entries(metadataMap())}>
                      {([k, v]) => (
                        <div class="flex items-start justify-between gap-3">
                          <dt class="text-[10px] font-semibold uppercase tracking-wide text-colpsi-muted mt-0.5">{fieldLabel(k)}</dt>
                          <dd class="text-colpsi-text font-medium text-right break-words max-w-[65%]">{fmtBool(v)}</dd>
                        </div>
                      )}
                    </For>
                  </dl>
                </section>
              </Show>
            </div>
          </aside>
        </div>
      )}
    </Show>
  );
}