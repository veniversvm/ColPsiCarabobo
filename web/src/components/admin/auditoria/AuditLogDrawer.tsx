// web/src/components/admin/auditoria/AuditLogDrawer.tsx
// Drawer lateral de detalle de un evento de la bitácora: muestra quién, cuándo,
// desde dónde, y el diff por campo con colores (from → to).

import { createMemo, Show, For } from "solid-js";
import type { ApiChangeLog, AuditChange } from "~/types/audit";
import { actionLabel, entityLabel, fieldLabel, parseAuditJson } from "~/types/audit";

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
            <div class="sticky top-0 z-10 bg-white border-b border-gray-100 px-6 py-4 flex items-start justify-between gap-3">
              <div class="min-w-0">
                <p class="text-[10px] font-black uppercase tracking-widest text-gray-400">
                  {entityLabel(l().entity)} · {actionLabel(l().action)}
                </p>
                <h3 class="text-lg font-black text-gray-900 truncate mt-0.5">{l().entity_label || l().entity_id}</h3>
              </div>
              <button
                onClick={props.onClose}
                class="flex-shrink-0 w-8 h-8 rounded-full bg-gray-100 hover:bg-gray-200 text-gray-500 font-bold flex items-center justify-center transition-colors"
              >
                ✕
              </button>
            </div>

            <div class="px-6 py-5 space-y-6">
              {/* ── Meta ── */}
              <dl class="grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
                <div><dt class="text-[10px] font-black uppercase tracking-widest text-gray-400">Fecha</dt><dd class="font-bold text-gray-800">{fmtDate(l().created_at)}</dd></div>
                <div><dt class="text-[10px] font-black uppercase tracking-widest text-gray-400">Actor</dt><dd class="font-bold text-gray-800">{l().actor_username || "—"} <span class="text-gray-400 font-medium">({l().actor_role || "—"})</span></dd></div>
                <div><dt class="text-[10px] font-black uppercase tracking-widest text-gray-400">IP</dt><dd class="font-bold text-gray-800">{l().ip || "—"}</dd></div>
                <div><dt class="text-[10px] font-black uppercase tracking-widest text-gray-400">Entidad ID</dt><dd class="font-bold text-gray-800 truncate">{l().entity_id || "—"}</dd></div>
              </dl>
              <Show when={l().user_agent}>
                <div class="bg-gray-50 rounded-xl px-3 py-2 text-[11px] text-gray-500 break-all font-medium">
                  {l().user_agent}
                </div>
              </Show>

              {/* ── Diff ── */}
              <Show when={Object.keys(changesMap()).length > 0}>
                <section>
                  <h4 class="text-xs font-black uppercase tracking-widest text-gray-400 mb-3">Cambios</h4>
                  <div class="space-y-2">
                    <For each={Object.entries(changesMap())}>
                      {([key, chg]) => (
                        <div class="rounded-xl border border-gray-200 overflow-hidden">
                          <p class="bg-gray-50 px-3 py-1.5 text-[10px] font-black uppercase tracking-widest text-gray-500 border-b border-gray-200">
                            {fieldLabel(key)}
                          </p>
                          <div class="grid grid-cols-2 divide-x divide-gray-200">
                            <div class="px-3 py-2.5 bg-red-50/60 text-red-700 text-sm font-bold break-words">
                              <span class="block text-[9px] font-black uppercase tracking-widest text-red-400 mb-0.5">Antes</span>
                              {fmtBool(chg.from)}
                            </div>
                            <div class="px-3 py-2.5 bg-emerald-50/60 text-emerald-700 text-sm font-bold break-words">
                              <span class="block text-[9px] font-black uppercase tracking-widest text-emerald-500 mb-0.5">Después</span>
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
                <p class="text-sm text-gray-400 font-medium">Sin campos detallados para este suceso.</p>
              </Show>

              {/* ── Metadata ── */}
              <Show when={Object.keys(metadataMap()).length > 0}>
                <section>
                  <h4 class="text-xs font-black uppercase tracking-widest text-gray-400 mb-3">Metadatos</h4>
                  <dl class="space-y-1.5 text-sm">
                    <For each={Object.entries(metadataMap())}>
                      {([k, v]) => (
                        <div class="flex items-start justify-between gap-3">
                          <dt class="text-[10px] font-black uppercase tracking-widest text-gray-400 mt-0.5">{fieldLabel(k)}</dt>
                          <dd class="text-gray-800 font-bold text-right break-words max-w-[65%]">{fmtBool(v)}</dd>
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