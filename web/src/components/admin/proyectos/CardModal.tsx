// web/src/components/admin/proyectos/CardModal.tsx
import { For, Show, createEffect, createSignal } from "solid-js";
import { apiPost, apiPatch, apiDelete } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { useAuth } from "~/lib/auth";
import { MAX_NOTE_LENGTH_CHARS, MAX_NOTES_PER_CARD, ProjectCard } from "~/types/projects";

export default function CardModal(props: {
  projectId: string;
  columnId: string;
  card: ProjectCard | null;
  canEdit: boolean;
  canManage: boolean;
  onClose: () => void;
  onChange: (card: ProjectCard | null, action: "create" | "update" | "delete") => void;
}) {
  const { user } = useAuth();
  const [title, setTitle] = createSignal(props.card?.title ?? "");
  const [description, setDescription] = createSignal(props.card?.description ?? "");
  const [notes, setNotes] = createSignal<ProjectCard["notes"]>(props.card?.notes ?? []);
  const [noteDraft, setNoteDraft] = createSignal("");
  const [busy, setBusy] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);
  const [saved, setSaved] = createSignal(false);

  createEffect(() => {
    if (props.card) {
      setTitle(props.card.title);
      setDescription(props.card.description);
      setNotes(props.card.notes ?? []);
    } else {
      setTitle("");
      setDescription("");
      setNotes([]);
    }
  });

  const isEditing = () => props.card !== null;

  const save = async () => {
    if (!title().trim()) {
      setError("La tarjeta necesita un título.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      if (isEditing()) {
        await apiPatch(`/admin/projects/cards/${props.card!.id}`, {
          title: title().trim(),
          description: description().trim(),
        });
        props.onChange({ ...props.card!, title: title().trim(), description: description().trim() }, "update");
      } else {
        const card = await apiPost<ProjectCard>(`/admin/projects/${props.projectId}/cards`, {
          column_id: props.columnId,
          title: title().trim(),
          description: description().trim(),
        });
        props.onChange({ ...card, notes: [] }, "create");
      }
      setSaved(true);
      window.setTimeout(() => setSaved(false), 1500);
    } catch (err) {
      setError(getUserFacingError(err));
    } finally {
      setBusy(false);
    }
  };

  const removeCard = async () => {
    if (!props.card) return;
    if (!window.confirm("¿Eliminar esta tarjeta? Se borrarán sus notas de forma definitiva.")) return;
    setBusy(true);
    setError(null);
    try {
      await apiDelete(`/admin/projects/cards/${props.card.id}`);
      props.onChange(props.card, "delete");
    } catch (err) {
      setError(getUserFacingError(err));
    } finally {
      setBusy(false);
    }
  };

  const addNote = async () => {
    const content = noteDraft().trim();
    if (!content || !props.card || busy()) return;
    if (notes().length >= MAX_NOTES_PER_CARD) return;
    setBusy(true);
    setError(null);
    try {
      const note = await apiPost<{ id: string; card_id: string; content: string; created_at: string; create_by: string; create_by_id: string | null }>(
        `/admin/projects/cards/${props.card.id}/notes`,
        { content }
      );
      setNotes((n) => [...n, { ...note, updated_at: note.created_at }]);
      setNoteDraft("");
      props.onChange({ ...props.card, notes: [...notes(), { ...note, updated_at: note.created_at }] }, "update");
    } catch (err) {
      setError(getUserFacingError(err));
    } finally {
      setBusy(false);
    }
  };

  const deleteNote = async (noteId: string) => {
    if (!props.card || busy()) return;
    setBusy(true);
    setError(null);
    try {
      await apiDelete(`/admin/projects/notes/${noteId}`);
      const next = notes().filter((n) => n.id !== noteId);
      setNotes(next);
      props.onChange({ ...props.card, notes: next }, "update");
    } catch (err) {
      setError(getUserFacingError(err));
    } finally {
      setBusy(false);
    }
  };

  const currentUserId = () => user()?.id;

  return (
    <div
      class="fixed inset-0 z-50 flex items-end md:items-center justify-center p-0 md:p-4 bg-black/60 backdrop-blur-sm"
      onClick={(e) => e.target === e.currentTarget && !busy() && props.onClose()}
    >
      <div class="bg-white w-full md:max-w-xl md:rounded-3xl rounded-t-3xl max-h-[92vh] overflow-y-auto" onClick={(e) => e.stopPropagation()}>
        <div class="flex items-center justify-between px-6 py-4 border-b border-colpsi-border sticky top-0 bg-white/95 backdrop-blur">
          <h3 class="font-black text-colpsi-blue text-lg">{isEditing() ? "Detalles de la tarjeta" : "Nueva tarjeta"}</h3>
          <div class="flex items-center gap-2">
            <Show when={saved()}>
              <span class="text-xs font-black text-emerald-600 bg-emerald-50 px-3 py-1.5 rounded-full">Guardado ✓</span>
            </Show>
            <button onClick={props.onClose} class="w-9 h-9 rounded-full bg-gray-100 text-gray-500 font-black hover:bg-gray-200">
              ✕
            </button>
          </div>
        </div>

        <div class="p-6 space-y-5">
          <Show when={error()}>
            <div class="p-3 rounded-xl bg-red-50 text-red-700 text-sm font-bold border-l-4 border-red-500">{error()}</div>
          </Show>

          <div class="space-y-2">
            <label class="block text-[10px] font-black text-gray-500 uppercase tracking-widest ml-1 mb-1">Título</label>
            <input
              value={title()}
              disabled={!props.canEdit}
              maxLength={200}
              onInput={(e) => setTitle(e.currentTarget.value)}
              placeholder="Título de la tarjeta"
              class="w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none transition-all text-gray-800 text-sm disabled:bg-colpsi-surface"
            />
          </div>

          <div class="space-y-2">
            <label class="block text-[10px] font-black text-gray-500 uppercase tracking-widest ml-1 mb-1">Descripción</label>
            <textarea
              value={description()}
              disabled={!props.canEdit}
              maxLength={2000}
              rows={4}
              onInput={(e) => setDescription(e.currentTarget.value)}
              placeholder="Detalles de la tarea…"
              class="w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none transition-all text-gray-800 text-sm resize-none disabled:bg-colpsi-surface"
            />
          </div>

          <Show when={props.canEdit}>
            <div class="flex gap-3">
              <button
                onClick={save}
                disabled={busy()}
                class="flex-grow bg-blue-800 hover:bg-blue-900 text-white font-black py-3 rounded-xl transition-all disabled:opacity-60"
              >
                {busy() ? "Guardando…" : isEditing() ? "Guardar cambios" : "Crear tarjeta"}
              </button>
              <Show when={isEditing()}>
                <button
                  onClick={removeCard}
                  disabled={busy()}
                  class="bg-red-50 text-red-600 font-black px-5 py-3 rounded-xl border-2 border-red-200 hover:bg-red-100 disabled:opacity-60"
                >
                  Eliminar
                </button>
              </Show>
            </div>
          </Show>

          <Show when={isEditing()}>
            <div class="border-t border-colpsi-border pt-5">
              <div class="flex items-center justify-between mb-3">
                <h4 class="font-black text-sm text-gray-700 uppercase tracking-widest text-[11px]">Notas</h4>
                <span class={`text-[11px] font-black ${notes().length >= MAX_NOTES_PER_CARD ? "text-red-500" : "text-gray-400"}`}>
                  {notes().length}/{MAX_NOTES_PER_CARD}
                </span>
              </div>

              <div class="space-y-2.5">
                <For each={notes()}>
                  {(n) => (
                    <div class="group p-3.5 rounded-2xl bg-colpsi-surface border border-colpsi-border">
                      <p class="text-sm text-gray-700 whitespace-pre-wrap break-words">{n.content}</p>
                      <div class="mt-2 flex items-center justify-between">
                        <span class="text-[10px] text-gray-400">
                          {n.create_by || "—"} · {new Date(n.created_at).toLocaleString("es-VE", { day: "2-digit", month: "short", hour: "2-digit", minute: "2-digit" })}
                        </span>
                        <Show when={props.canEdit && (props.canManage || n.create_by_id === currentUserId())}>
                          <button
                            onClick={() => deleteNote(n.id)}
                            class="text-[11px] font-black text-red-400 hover:text-red-600 opacity-0 group-hover:opacity-100 transition-opacity"
                          >
                            Eliminar
                          </button>
                        </Show>
                      </div>
                    </div>
                  )}
                </For>
                <Show when={notes().length === 0}>
                  <p class="text-center text-xs text-gray-400 py-3">Sin notas todavía.</p>
                </Show>
              </div>

              <Show when={props.canEdit}>
                <div class="mt-4 rounded-2xl border-2 border-dashed border-gray-200 p-3">
                  <textarea
                    value={noteDraft()}
                    disabled={notes().length >= MAX_NOTES_PER_CARD || busy()}
                    rows={2}
                    maxLength={MAX_NOTE_LENGTH_CHARS}
                    onInput={(e) => setNoteDraft(e.currentTarget.value)}
                    placeholder={notes().length >= MAX_NOTES_PER_CARD ? "Límite de 10 notas alcanzado" : `Añadir nota (máx. ${MAX_NOTE_LENGTH_CHARS} caracteres)…`}
                    class="w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none text-sm resize-none"
                  />
                  <div class="mt-2 flex items-center justify-between">
                    <span class={`text-[10px] font-black ${noteDraft().length > MAX_NOTE_LENGTH_CHARS - 20 ? "text-red-500" : "text-gray-300"}`}>
                      {noteDraft().length}/{MAX_NOTE_LENGTH_CHARS}
                    </span>
                    <button
                      onClick={addNote}
                      disabled={!noteDraft().trim() || notes().length >= MAX_NOTES_PER_CARD || busy()}
                      class="bg-colpsi-yellow text-colpsi-blue font-black px-4 py-2 rounded-xl text-xs disabled:opacity-40 hover:opacity-90"
                    >
                      Añadir nota
                    </button>
                  </div>
                </div>
              </Show>
            </div>
          </Show>
        </div>
      </div>
    </div>
  );
}