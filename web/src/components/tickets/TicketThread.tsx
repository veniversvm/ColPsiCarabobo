// web/src/components/tickets/TicketThread.tsx
// Hilo de conversación de un ticket: burbujas de mensajes (psi a la izquierda,
// admin a la derecha) con sus anexos. Es compartido por el portal psi y el
// panel admin. Anexos: los adjuntos ya llegan con `url` resuelta por la API;
// `bucketUrl` es idempotente por si algún día entra una key cruda.
import { For, Show } from "solid-js";
import { bucketUrl } from "~/lib/bucket";
import type { TicketMensaje } from "~/types/tickets";
import { formatTicketDateTime, formatFileSize } from "~/types/tickets";

interface Props {
  mensajes: TicketMensaje[];
  /** Nombre mostrado cuando el autor es el psicólogo dueño (ej. portal psi). */
  psiDisplayName?: string;
  /** Nombre del admin mostrado en el portal psi ("El colegio" por defecto). */
  adminDisplayName?: string;
  emptyText?: string;
}

export default function TicketThread(props: Props) {
  return (
    <div class="space-y-4">
      <Show
        when={props.mensajes.length > 0}
        fallback={
          <div class="bg-blue-50/60 border border-dashed border-blue-200 rounded-3xl p-10 text-center">
            <p class="text-4xl mb-3">💬</p>
            <p class="text-sm font-bold text-colpsi-muted">
              {props.emptyText || "Sin comentarios todavía"}
            </p>
          </div>
        }
      >
        <For each={props.mensajes}>
          {(m) => {
            const isAdmin = m.author_type === "admin";
            const adjuntos = m.adjuntos ?? [];
            return (
              <div class={`flex ${isAdmin ? "justify-end" : "justify-start"}`}>
                <div
                  class={`max-w-[85%] md:max-w-[75%] rounded-3xl px-5 py-4 shadow-sm border ${
                    isAdmin
                      ? "bg-blue-700 text-white border-blue-800 rounded-tr-none"
                      : "bg-white text-gray-800 border-colpsi-border rounded-tl-none"
                  }`}
                >
                  <div class={`flex items-baseline justify-between gap-3 mb-1.5`}>
                    <span class={`text-[10px] font-black uppercase tracking-widest ${isAdmin ? "text-blue-200" : "text-blue-600"}`}>
                      {isAdmin ? (props.adminDisplayName || "El Colegio") : (props.psiDisplayName || m.author_name || "Yo")}
                    </span>
                    <span class={`text-[9px] font-bold whitespace-nowrap ${isAdmin ? "text-blue-300" : "text-gray-400"}`}>
                      {formatTicketDateTime(m.created_at)}
                    </span>
                  </div>
                  <p class="text-sm leading-relaxed whitespace-pre-wrap break-words">{m.message}</p>

                  <Show when={adjuntos.length > 0}>
                    <div class={`mt-3 pt-3 border-t ${isAdmin ? "border-blue-600" : "border-colpsi-border"} flex flex-wrap gap-2`}>
                      <For each={adjuntos}>
                        {(adj) => (
                          <a
                            href={bucketUrl(adj.url)}
                            target="_blank"
                            rel="noopener noreferrer"
                            class={`inline-flex items-center gap-2 px-3 py-1.5 rounded-xl text-[11px] font-bold transition-all active:scale-95 ${
                              isAdmin
                                ? "bg-blue-800/60 text-blue-100 hover:bg-blue-900"
                                : "bg-blue-50 text-blue-700 hover:bg-blue-100"
                            }`}
                          >
                            <span class="text-sm">📎</span>
                            <span class="max-w-[160px] truncate">{adj.original_name || adj.mime_type}</span>
                            <Show when={adj.size_bytes}>
                              <span class={`opacity-60 ${isAdmin ? "text-blue-200" : "text-gray-400"}`}>
                                {formatFileSize(adj.size_bytes)}
                              </span>
                            </Show>
                          </a>
                        )}
                      </For>
                    </div>
                  </Show>
                </div>
              </div>
            );
          }}
        </For>
      </Show>
    </div>
  );
}