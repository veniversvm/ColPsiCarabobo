// web/src/routes/admin/proyectos/[id].tsx
import { For, Show, createSignal, createResource, ErrorBoundary } from "solid-js";
import { isServer } from "solid-js/web";
import { A, useParams } from "@solidjs/router";
import {
  DragDropProvider,
  DragDropSensors,
  DragOverlay,
  createDraggable,
  createDroppable,
} from "~/vendor/thisbeyond-solid-dnd";
import { apiGet, apiPost, apiPatch, apiDelete } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import {
  BoardColumn, BoardCard, BoardMember, ProjectBoard,
} from "~/types/projects";
import { canEditProject, canManageProject } from "~/types/projects";
import CardModal from "~/components/admin/proyectos/CardModal";
import MembersModal from "~/components/admin/proyectos/MembersModal";
import ConfirmModal from "~/components/admin/proyectos/ConfirmModal";
import { Icon } from "~/components/admin/ui/icons";

interface BoardChunk {
  columns: BoardColumn[];
  cards: BoardCard[];
}

function CardBody(props: { card: BoardCard }) {
  return (
    <>
      <h4 class="font-bold text-sm text-gray-800 leading-snug break-words">{props.card.title}</h4>
      <Show when={props.card.description}>
        <p class="mt-1 text-xs text-gray-400 line-clamp-2">{props.card.description}</p>
      </Show>
      <div class="mt-3 flex items-center gap-2 text-[11px] font-bold text-gray-400">
        <Show when={props.card.notes && props.card.notes!.length > 0}>
          <span class="flex items-center gap-1 text-amber-600"><Icon name="fileText" class="w-3 h-3" /> {props.card.notes!.length}</span>
        </Show>
        <span class="ml-auto text-[10px]">{props.card.create_by || "—"}</span>
      </div>
    </>
  );
}

function Card(props: { card: BoardCard; canEdit: boolean; onOpen: (c: BoardCard) => void }) {
  const draggable = createDraggable(`card:${props.card.id}`, { type: "card", card: props.card });
  return (
    <div
      ref={(el) => draggable(el, () => ({ skipTransform: true }))}
      onClick={() => props.onOpen(props.card)}
      class="bg-white rounded-lg border border-colpsi-border p-3.5 cursor-grab active:cursor-grabbing select-none hover:border-colpsi-blue/40 transition-shadow duration-150 group"
      classList={{ "opacity-40": draggable.isActiveDraggable }}
    >
      <CardBody card={props.card} />
    </div>
  );
}

function Column(props: {
  column: BoardColumn;
  canEdit: boolean;
  onOpenCard: (c: BoardCard) => void;
  onNewCard: () => void;
  onEditTitle: () => void;
  onDeleteCol: () => void;
}) {
  const droppable = createDroppable(`column:${props.column.id}`, {
    type: "column",
    columnId: props.column.id,
  });
  return (
    <div
      ref={droppable.ref}
      class="flex flex-col w-[290px] shrink-0 max-h-full rounded-lg bg-[#f4f6f9] border border-colpsi-border overflow-hidden transition-[border-color,box-shadow] duration-150"
      classList={{ "border-blue-400 shadow-lg ring-2 ring-blue-300/60": droppable.isActiveDroppable }}
    >
      <div class="flex items-center justify-between px-3 py-2.5 bg-white border-b border-colpsi-border cursor-grab">
        <span class="font-semibold text-sm text-colpsi-text flex items-center gap-2">
          {props.column.title}
          <span class="text-[10px] font-medium bg-colpsi-bg text-colpsi-muted px-1.5 py-0.5 rounded">
            {props.column.cards?.length ?? 0}
          </span>
        </span>
        <div class="flex items-center gap-1">
          <Show when={props.canEdit}>
            <button onClick={props.onNewCard} class="inline-flex items-center justify-center w-7 h-7 rounded-md text-colpsi-blue font-bold hover:bg-colpsi-bg" title="Nueva tarjeta">
              <Icon name="plus" class="w-4 h-4" />
            </button>
            <button onClick={props.onEditTitle} class="inline-flex items-center justify-center w-7 h-7 rounded-md text-colpsi-muted hover:bg-colpsi-bg hover:text-colpsi-blue" title="Editar">
              <Icon name="pencil" class="w-3.5 h-3.5" />
            </button>
            <button onClick={props.onDeleteCol} class="inline-flex items-center justify-center w-7 h-7 rounded-md text-colpsi-red/70 hover:bg-red-50 hover:text-colpsi-red" title="Eliminar columna">
              <Icon name="trash" class="w-3.5 h-3.5" />
            </button>
          </Show>
        </div>
      </div>
      <div class="flex flex-col gap-2.5 p-3 overflow-y-auto flex-grow">
        <For each={props.column.cards ?? []}>
          {(card) => <Card card={card} canEdit={props.canEdit} onOpen={(c) => props.onOpenCard(c)} />}
        </For>
      </div>
    </div>
  );
}

function AddColumn(props: { canEdit: boolean; onAdd: (title: string) => Promise<void> }) {
  const [open, setOpen] = createSignal(false);
  const [title, setTitle] = createSignal("");
  const [busy, setBusy] = createSignal(false);

  const submit = async () => {
    if (!title().trim() || busy()) return;
    setBusy(true);
    try {
      await props.onAdd(title().trim());
      setTitle("");
      setOpen(false);
    } catch {
      // error mostrado por el padre
    } finally {
      setBusy(false);
    }
  };

  return (
    <Show when={props.canEdit}>
      <div class="w-[290px] shrink-0">
        <Show
          when={open()}
          fallback={
            <button
              onClick={() => setOpen(true)}
              class="w-full rounded-lg border border-dashed border-colpsi-border text-colpsi-muted font-medium py-3 text-sm hover:border-colpsi-blue hover:text-colpsi-blue transition-colors"
            >
              + Añadir columna
            </button>
          }
        >
          <div class="bg-white rounded-lg border border-colpsi-border p-3">
            <input
              value={title()}
              onInput={(e) => setTitle(e.currentTarget.value)}
              onKeyDown={(e) => e.key === "Enter" && submit()}
              placeholder="Nombre de la columna"
              maxLength={120}
              autofocus
              class="h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors"
            />
            <div class="mt-2 flex gap-2">
              <button onClick={submit} disabled={!title().trim() || busy()} class="h-8 px-4 rounded-md bg-colpsi-blue text-white font-semibold text-xs disabled:opacity-40 hover:bg-colpsi-blue-light transition-colors">
                Añadir
              </button>
              <button onClick={() => setOpen(false)} class="h-8 px-3 rounded-md text-colpsi-muted text-xs font-medium hover:bg-colpsi-bg transition-colors">
                Cancelar
              </button>
            </div>
          </div>
        </Show>
      </div>
    </Show>
  );
}

export default function ProjectBoardPage() {
  const params = useParams();
  const projectId = params.id;

  const [error, setError] = createSignal<string | null>(null);
  const [openedCard, setOpenedCard] = createSignal<{ columnId: string; card: BoardCard | null } | null>(null);
  const [showMembers, setShowMembers] = createSignal(false);
  const [menu, setMenu] = createSignal(false);
  const [renameOpen, setRenameOpen] = createSignal(false);
  const [renameCol, setRenameCol] = createSignal<BoardColumn | null>(null);
  const [renameTitle, setRenameTitle] = createSignal("");
  const [deleteColumn, setDeleteColumn] = createSignal<BoardColumn | null>(null);
  const [busy, setBusy] = createSignal(false);
  const [loadErr, setLoadErr] = createSignal(false);
  const [activeCard, setActiveCard] = createSignal<BoardCard | null>(null);

  const [board, { refetch }] = createResource<ProjectBoard | null>(
    () => {
      setLoadErr(false);
      return apiGet<ProjectBoard>(`/admin/projects/${projectId}`).catch(() => {
        setLoadErr(true);
        return null;
      });
    }
  );

  // Estado local del tablero para interoperar con el DnD sin re-fetches.
  const [local, setLocal] = createSignal<BoardChunk | null>(null);

  const project = () => board()?.project;
  const isEditor = () => (project() ? canEditProject(project()!) : false);
  const isManager = () => (project() ? canManageProject(project()!) : false);

  const columns = () => local()?.columns ?? board()?.columns ?? [];
  const allCards = () => (local()?.columns ?? board()?.columns ?? []).flatMap((c) => c.cards ?? []);

  const [members, { refetch: refetchMembers }] = createResource<BoardMember[]>(
    () => apiGet<{ data: BoardMember[] }>(`/admin/projects/${projectId}/members`).then((r) => r.data)
  );

  const findCard = (cardId: string) => allCards().find((c) => c.id === cardId);

  const refresh = () => {
    refetch();
    refetchMembers();
  };

  // ── DnD ────────────────────────────────────────────────────────────────
  const findColumnById = (id: string) => columns().find((c) => c.id === id);

  const onDragEnd = async (event: any) => {
    if (!isEditor()) return;
    const { draggable, droppable } = event;
    setActiveCard(null);
    if (!draggable || !droppable) return;

    const dragId: string = String(draggable.id);
    const dropId: string = String(droppable.id);
    const cardId = dragId.startsWith("card:") ? dragId.slice("card:".length) : null;
    const targetColumnId = dropId.startsWith("column:") ? dropId.slice("column:".length) : null; 
    if (!cardId || !targetColumnId) return;

    const targetCol = findColumnById(targetColumnId);
    if (!targetCol) return;

    const card = findCard(cardId);
    if (!card || card.column_id === targetColumnId) return;

    // Optimistic update
    const nextCards = allCards().map((c) => (c.id === cardId ? { ...c, column_id: targetColumnId } : c));
    const nextColumns = columns().map((c) => ({
      ...c,
      cards: nextCards.filter((x) => x.column_id === c.id),
    }));
    setLocal({ columns: nextColumns, cards: nextCards });

    try {
      await apiPatch(`/admin/projects/cards/${cardId}`, { column_id: targetColumnId });
    } catch (err) {
      setError(getUserFacingError(err));
      refresh();
    }
  };

  const openCard = (card: BoardCard) => setOpenedCard({ columnId: card.column_id, card });

  const saveCard = (card: BoardCard | null, action: "create" | "update" | "delete") => {
    if (action === "create" && card) {
      const nextColumns = columns().map((c) =>
        c.id === card.column_id ? { ...c, cards: [...(c.cards ?? []), { ...card, notes: card.notes ?? [] }] } : c
      );
      const nextCards = [...allCards(), card];
      setLocal({ columns: nextColumns, cards: nextCards });
    } else if (action === "update" && card) {
      const nextCards = allCards().map((c) => (c.id === card.id ? { ...card, notes: card.notes ?? [] } : c));
      const nextColumns = columns().map((c) => ({ ...c, cards: nextCards.filter((x) => x.column_id === c.id) }));
      setLocal({ columns: nextColumns, cards: nextCards });
    } else if (action === "delete" && card) {
      const nextCards = allCards().filter((c) => c.id !== card.id);
      const nextColumns = columns().map((c) => ({ ...c, cards: nextCards.filter((x) => x.column_id === c.id) }));
      setLocal({ columns: nextColumns, cards: nextCards });
    }
    setOpenedCard(null);
    refresh();
  };

  const addColumn = async (title: string) => {
    const col = await apiPost<BoardColumn>(`/admin/projects/${projectId}/columns`, { title });
    const nextColumns = [...columns(), { ...col, cards: [] }];
    setLocal({ columns: nextColumns, cards: allCards() });
    setError(null);
  };

  const startRename = (col: BoardColumn) => {
    setRenameTitle(col.title);
    setRenameCol(col);
    setRenameOpen(true);
  };

  const submitRename = async () => {
    const col = renameCol();
    if (!col || !renameTitle().trim() || busy()) return;
    setBusy(true);
    try {
      await apiPatch(`/admin/projects/columns/${col.id}`, { title: renameTitle().trim() });
      const nextColumns = columns().map((c) => (c.id === col.id ? { ...c, title: renameTitle().trim() } : c));
      setLocal({ columns: nextColumns, cards: allCards() });
      setRenameOpen(false);
      setRenameCol(null);
    } catch (err) {
      setError(getUserFacingError(err));
    } finally {
      setBusy(false);
    }
  };

  const confirmDeleteColumn = async () => {
    const col = deleteColumn();
    if (!col || busy()) return;
    setBusy(true);
    try {
      await apiDelete(`/admin/projects/columns/${col.id}`);
      const nextCards = allCards().filter((c) => c.column_id !== col.id);
      const nextColumns = columns().filter((c) => c.id !== col.id);
      setLocal({ columns: nextColumns, cards: nextCards });
      setDeleteColumn(null);
    } catch (err) {
      setError(getUserFacingError(err));
    } finally {
      setBusy(false);
    }
  };

  const deleteProject = async () => {
    if (busy()) return;
    setBusy(true);
    try {
      await apiDelete(`/admin/projects/${projectId}`);
      window.location.href = "/admin/proyectos";
    } catch (err) {
      setError(getUserFacingError(err));
      setBusy(false);
    }
  };

  return (
    <ErrorBoundary fallback={<p class="text-sm text-red-500">No se pudo cargar el tablero.</p>}>
      <div class="flex items-center justify-between gap-3 pb-4 border-b border-colpsi-border flex-wrap">
        <div class="flex items-center gap-3 min-w-0">
          <A href="/admin/proyectos" class="inline-flex items-center justify-center h-8 w-8 rounded-md border border-colpsi-border bg-white text-colpsi-muted hover:text-colpsi-blue hover:bg-colpsi-bg transition-colors shrink-0" title="Volver a proyectos">
            <Icon name="chevronRight" class="w-4 h-4 rotate-180" />
          </A>
          <div class="min-w-0">
            <h1 class="text-lg font-semibold text-colpsi-text truncate">{project()?.name ?? (loadErr() ? "Proyecto no encontrado" : "Cargando...")}</h1>
            <Show when={project()?.description}>
              <p class="text-sm text-colpsi-muted mt-0.5 truncate">{project()?.description}</p>
            </Show>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <button
            onClick={() => setShowMembers(true)}
            class="inline-flex items-center gap-2 h-9 px-3.5 rounded-md bg-white border border-colpsi-border text-sm font-medium text-colpsi-text hover:border-colpsi-blue/40 hover:bg-colpsi-bg transition-colors"
          >
            <Icon name="users" class="w-4 h-4" />
            Miembros
          </button>
          <div class="relative">
            <button onClick={() => setMenu(!menu())} class="inline-flex items-center justify-center h-9 w-9 rounded-md bg-white border border-colpsi-border text-colpsi-muted hover:border-colpsi-blue/40 hover:text-colpsi-blue transition-colors">
              <Icon name="dots" class="w-4 h-4" />
            </button>
            <Show when={menu()}>
              <div class="absolute right-0 mt-1 w-52 bg-white rounded-md border border-colpsi-border shadow-lg z-20 p-1" onClick={() => setMenu(false)}>
                <Show when={isManager()}>
                  <button
                    onClick={() => window.confirm("¿Eliminar este proyecto? Esta acción no se puede deshacer.") && deleteProject()}
                    class="w-full flex items-center gap-2 text-left px-3 py-2 rounded text-sm font-medium text-colpsi-red hover:bg-red-50 transition-colors"
                  >
                    <Icon name="trash" class="w-4 h-4" />
                    Eliminar proyecto
                  </button>
                </Show>
                <button onClick={refresh} class="w-full flex items-center gap-2 text-left px-3 py-2 rounded text-sm font-medium text-colpsi-muted hover:bg-colpsi-bg transition-colors">
                  <Icon name="refresh" class="w-4 h-4" />
                  Refrescar
                </button>
              </div>
            </Show>
          </div>
        </div>
      </div>

      <Show when={error()}>
        <div class="p-3 rounded-md bg-red-50 text-red-700 border border-red-200 text-sm font-medium">{error()}</div>
      </Show>

      <Show when={!isServer}>
        <DragDropProvider
          onDragStart={({ draggable }: any) => {
            if (draggable?.data?.type === "card") setActiveCard(draggable.data.card);
          }}
          onDragEnd={onDragEnd}
        >
          <DragDropSensors />
          <div class="flex gap-4 items-start overflow-x-auto pb-6 -mx-2 px-2">
            <For each={columns()}>
              {(col) => (
                <Column
                  column={col}
                  canEdit={isEditor()}
                  onOpenCard={(c) => openCard(c)}
                  onNewCard={() => setOpenedCard({ columnId: col.id, card: null })}
                  onEditTitle={() => startRename(col)}
                  onDeleteCol={() => setDeleteColumn(col)}
                />
              )}
            </For>
            <AddColumn canEdit={isEditor()} onAdd={addColumn} />
          </div>
          <DragOverlay>
            <Show when={activeCard()}>
              <div class="w-[290px] rotate-2 rounded-lg bg-white border border-colpsi-blue/40 shadow-2xl p-3.5 select-none pointer-events-none">
                <CardBody card={activeCard()!} />
              </div>
            </Show>
          </DragOverlay>
        </DragDropProvider>
      </Show>

      <Show when={openedCard()}>
        <CardModal
          projectId={projectId}
          columnId={openedCard()!.columnId}
          card={openedCard()!.card}
          canEdit={isEditor()}
          canManage={isManager()}
          onClose={() => setOpenedCard(null)}
          onChange={saveCard}
        />
      </Show>

      <Show when={showMembers()}>
        <MembersModal
          project={project()!}
          members={members() ?? []}
          canManage={isManager()}
          onClose={() => setShowMembers(false)}
          reload={refetchMembers}
        />
      </Show>

      <Show when={renameOpen() && renameCol()}>
        <div
          class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
          onClick={(e) => e.target === e.currentTarget && !busy() && setRenameOpen(false)}
        >
          <div class="bg-white rounded-lg shadow-lg p-5 w-full max-w-sm border border-colpsi-border">
            <h3 class="text-base font-semibold text-colpsi-text">Renombrar columna</h3>
            <input
              value={renameTitle()}
              onInput={(e) => setRenameTitle(e.currentTarget.value)}
              onKeyDown={(e) => e.key === "Enter" && submitRename()}
              maxLength={120}
              autofocus
              class="mt-3 h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors"
            />
            <div class="mt-4 grid grid-cols-2 gap-2">
              <button onClick={() => setRenameOpen(false)} class="h-9 rounded-md bg-white text-colpsi-text border border-colpsi-border font-medium hover:bg-colpsi-bg transition-colors text-sm">Cancelar</button>
              <button onClick={submitRename} class="h-9 rounded-md bg-colpsi-blue text-white font-semibold hover:bg-colpsi-blue-light transition-colors disabled:opacity-60 text-sm" disabled={busy() || !renameTitle().trim()}>Guardar</button>
            </div>
          </div>
        </div>
      </Show>

      <Show when={deleteColumn()}>
        <ConfirmModal
          title="Eliminar columna"
          message={`¿Eliminar la columna «${deleteColumn()!.title}» y todas sus tarjetas?`}
          confirmLabel="Eliminar"
          danger
          busy={busy()}
          onConfirm={confirmDeleteColumn}
          onClose={() => !busy() && setDeleteColumn(null)}
        />
      </Show>
    </ErrorBoundary>
  );
}