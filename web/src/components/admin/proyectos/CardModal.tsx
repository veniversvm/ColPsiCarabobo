// web/src/components/admin/proyectos/CardModal.tsx
import { For, Show, createEffect, createSignal } from "solid-js";
import { apiPost, apiPatch, apiDelete } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { useAuth } from "~/lib/auth";
import { MAX_NOTE_LENGTH_CHARS, MAX_NOTES_PER_CARD, ProjectCard } from "~/types/projects";
import { Icon } from "~/components/admin/ui/icons";

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
      <div class="bg-white w-full md:max-w-xl md:rounded-lg rounded-t-lg max-h-[92vh] overflow-y-auto" onClick={(e) => e.stopPropagation()}>
        <div class="flex items-center justify-between px-5 py-3.5 border-b border-colpsi-border sticky top-0 bg-white/95 backdrop-blur">
          <h3 class="font-semibold text-colpsi-text">{isEditing() ? "Detalles de la tarjeta" : "Nueva tarjeta"}</h3>
          <div class="flex items-center gap-2">
            <Show when={saved()}>
              <span class="text-xs font-medium text-emerald-700 bg-emerald-50 border border-emerald-200 px-2 py-0.5 rounded inline-flex items-center gap-1"><Icon name="check" class="w-3 h-3" /> Guardado</span>
            </Show>
            <button onClick={props.onClose} class="inline-flex items-center justify-center h-8 w-8 rounded-md bg-colpsi-bg text-colpsi-muted hover:text-colpsi-blue hover:bg-colpsi-border/60 transition-colors">
              <Icon name="x" class="w-4 h-4" />
            </button>
          </div>
        </div>

        <div class="p-5 space-y-4">
          <Show when={error()}>
            <div class="p-3 rounded-md bg-red-50 text-red-700 text-sm font-medium border border-red-200">{error()}</div>
          </Show>

          <div class="space-y-1.5">
            <label class="block text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1 mb-1">Título</label>
            <input
              value={title()}
              disabled={!props.canEdit}
              maxLength={200}
              onInput={(e) => setTitle(e.currentTarget.value)}
              placeholder="Título de la tarjeta"
              class="h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors disabled:bg-colpsi-bg"
            />
          </div>

          <div class="space-y-1.5">
            <label class="block text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1 mb-1">Descripción</label>
            <textarea
              value={description()}
              disabled={!props.canEdit}
              maxLength={2000}
              rows={4}
              onInput={(e) => setDescription(e.currentTarget.value)}
              placeholder="Detalles de la tarea..."
              class="w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors resize-none py-2 disabled:bg-colpsi-bg"
            />
          </div>

          <Show when={props.canEdit}>
            <div class="flex gap-2">
              <button
                onClick={save}
                disabled={busy()}
                class="flex-grow h-10 rounded-md bg-colpsi-blue hover:bg-colpsi-blue-light text-white font-semibold transition-colors disabled:opacity-60 text-sm"
              >
                {busy() ? "Guardando..." : isEditing() ? "Guardar cambios" : "Crear tarjeta"}
              </button>
              <Show when={isEditing()}>
                <button
                  onClick={removeCard}
                  disabled={busy()}
                  class="inline-flex items-center gap-1.5 h-10 px-4 rounded-md bg-white text-colpsi-red font-medium border border-red-200 hover:bg-red-50 disabled:opacity-60 transition-colors text-sm"
                >
                  <Icon name="trash" class="w-4 h-4" />
                  Eliminar
                </button>
              </Show>
            </div>
          </Show>

          <Show when={isEditing()}>
            <div class="border-t border-colpsi-border pt-4">
              <div class="flex items-center justify-between mb-2.5">
                <h4 class="text-xs font-semibold text-colpsi-muted uppercase tracking-wide">Notas</h4>
                <span class={`text-[11px] font-medium ${notes().length >= MAX_NOTES_PER_CARD ? "text-colpsi-red" : "text-colpsi-muted"}`}>
                  {notes().length}/{MAX_NOTES_PER_CARD}
                </span>
              </div>

              <div class="space-y-2">
                <For each={notes()}>
                  {(n) => (
                    <div class="group p-3 rounded-md bg-colpsi-bg border border-colpsi-border">
                      <p class="text-sm text-colpsi-text whitespace-pre-wrap break-words">{n.content}</p>
                      <div class="mt-2 flex items-center justify-between">
                        <span class="text-[11px] text-colpsi-muted">
                          {n.create_by || "—"} · {new Date(n.created_at).toLocaleString("es-VE", { day: "2-digit", month: "short", hour: "2-digit", minute: "2-digit" })}
                        </span>
                        <Show when={props.canEdit && (props.canManage || n.create_by_id === currentUserId())}>
                          <button
                            onClick={() => deleteNote(n.id)}
                            class="text-[11px] font-medium text-colpsi-red/70 hover:text-colpsi-red opacity-0 group-hover:opacity-100 transition-opacity"
                          >
                            Eliminar
                          </button>
                        </Show>
                      </div>
                    </div>
                  )}
                </For>
                <Show when={notes().length === 0}>
                  <p class="text-center text-xs text-colpsi-muted py-2.5">Sin notas todavía.</p>
                </Show>
              </div>

              <Show when={props.canEdit}>
                <div class="mt-3 rounded-md border border-dashed border-colpsi-border p-3">
                  <textarea
                    value={noteDraft()}
                    disabled={notes().length >= MAX_NOTES_PER_CARD || busy()}
                    rows={2}
                    maxLength={MAX_NOTE_LENGTH_CHARS}
                    onInput={(e) => setNoteDraft(e.currentTarget.value)}
                    placeholder={notes().length >= MAX_NOTES_PER_CARD ? "Límite de 10 notas alcanzado" : `Añadir nota (máx. ${MAX_NOTE_LENGTH_CHARS} caracteres)...`}
                    class="w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors resize-none"
                  />
                  <div class="mt-2 flex items-center justify-between">
                    <span class={`text-[10px] font-medium ${noteDraft().length > MAX_NOTE_LENGTH_CHARS - 20 ? "text-colpsi-red" : "text-colpsi-muted"}`}>
                      {noteDraft().length}/{MAX_NOTE_LENGTH_CHARS}
                    </span>
                    <button
                      onClick={addNote}
                      disabled={!noteDraft().trim() || notes().length >= MAX_NOTES_PER_CARD || busy()}
                      class="h-8 px-4 rounded-md bg-colpsi-blue text-white font-semibold text-xs disabled:opacity-40 hover:bg-colpsi-blue-light transition-colors"
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