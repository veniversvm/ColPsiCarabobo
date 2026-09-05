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
import { Presence } from "~/components/ui/Motion";

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
          <span class="flex items-center gap-1 text-amber-600">📝 {props.card.notes!.length}</span>
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
      class="bg-white rounded-2xl border border-gray-200 shadow-sm p-4 cursor-grab active:cursor-grabbing select-none hover:shadow-md hover:border-blue-200 transition-shadow duration-150 group"
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
      class="flex flex-col w-[290px] shrink-0 max-h-full rounded-2xl bg-[#eef1f6] border border-gray-200/60 overflow-hidden transition-[border-color,box-shadow] duration-150"
      classList={{ "border-blue-400 shadow-lg ring-2 ring-blue-300/60": droppable.isActiveDroppable }}
    >
      <div class="flex items-center justify-between px-4 py-3 bg-white border-b border-gray-100 cursor-grab">
        <span class="font-black text-sm text-colpsi-blue flex items-center gap-2">
          {props.column.title}
          <span class="text-[10px] bg-gray-100 text-gray-500 px-2 py-0.5 rounded-full">
            {props.column.cards?.length ?? 0}
          </span>
        </span>
        <div class="flex items-center gap-1">
          <Show when={props.canEdit}>
            <button onClick={props.onNewCard} class="w-7 h-7 rounded-lg text-colpsi-blue font-black hover:bg-gray-100" title="Nueva tarjeta">
              +
            </button>
            <button onClick={props.onEditTitle} class="w-7 h-7 rounded-lg text-gray-400 font-black hover:bg-gray-100" title="Editar">
              ✎
            </button>
            <button onClick={props.onDeleteCol} class="w-7 h-7 rounded-lg text-red-400 font-black hover:bg-red-50" title="Eliminar columna">
              ✕
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
              class="w-full rounded-2xl border-2 border-dashed border-gray-300 text-gray-500 font-black py-3 text-sm hover:border-colpsi-blue hover:text-colpsi-blue transition-colors"
            >
              + Añadir columna
            </button>
          }
        >
          <div class="bg-white rounded-2xl border-2 border-gray-200 p-3">
            <input
              value={title()}
              onInput={(e) => setTitle(e.currentTarget.value)}
              onKeyDown={(e) => e.key === "Enter" && submit()}
              placeholder="Nombre de la columna"
              maxLength={120}
              autofocus
              class="w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-3 py-2 text-sm outline-none"
            />
            <div class="mt-2 flex gap-2">
              <button onClick={submit} disabled={!title().trim() || busy()} class="bg-blue-800 text-white font-black text-xs px-4 py-2 rounded-xl disabled:opacity-40">
                Añadir
              </button>
              <button onClick={() => setOpen(false)} class="text-gray-400 font-black text-xs px-3 py-2 rounded-xl hover:bg-gray-100">
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
  const allCards = () => local()?.cards ?? [];

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
      <div class="flex items-center justify-between mb-6 gap-4 flex-wrap">
        <div>
          <A href="/admin/proyectos" class="text-xs font-bold text-blue-500 hover:text-blue-700">← Volver a proyectos</A>
          <h1 class="text-2xl md:text-3xl font-black text-colpsi-blue mt-1">{project()?.name ?? (loadErr() ? "Proyecto no encontrado" : "Cargando…")}</h1>
          <Show when={project()?.description}>
            <p class="text-sm text-gray-500 mt-1">{project()?.description}</p>
          </Show>
        </div>

        <div class="flex items-center gap-2">
          <button
            onClick={() => setShowMembers(true)}
            class="bg-white border-2 border-gray-200 text-gray-700 font-black px-4 py-2.5 rounded-xl hover:border-colpsi-blue transition-colors"
          >
            👥 Miembros
          </button>
          <div class="relative">
            <button onClick={() => setMenu(!menu())} class="w-11 h-11 rounded-xl bg-white border-2 border-gray-200 text-gray-600 font-black hover:border-colpsi-blue transition-colors">
              ⋯
            </button>
            <Show when={menu()}>
              <div class="absolute right-0 mt-2 w-52 bg-white rounded-2xl border border-gray-100 shadow-xl z-20 p-2" onClick={() => setMenu(false)}>
                <Show when={isManager()}>
                  <button
                    onClick={() => window.confirm("¿Eliminar este proyecto? Esta acción no se puede deshacer.") && deleteProject()}
                    class="w-full text-left px-3 py-2.5 rounded-xl text-sm font-black text-red-600 hover:bg-red-50"
                  >
                    🗑 Eliminar proyecto
                  </button>
                </Show>
                <button onClick={refresh} class="w-full text-left px-3 py-2.5 rounded-xl text-sm font-black text-gray-600 hover:bg-gray-100">
                  🔄 Refrescar
                </button>
              </div>
            </Show>
          </div>
        </div>
      </div>

      <Show when={error()}>
        <div class="mb-5 p-4 rounded-2xl bg-red-50 text-red-800 font-bold text-sm border-l-4 border-red-500">{error()}</div>
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
              <div class="w-[290px] rotate-2 rounded-2xl bg-white border border-blue-200 shadow-2xl p-4 select-none pointer-events-none">
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
          <div class="bg-white rounded-3xl shadow-2xl p-6 w-full max-w-sm">
            <h3 class="font-black text-colpsi-blue text-lg">Renombrar columna</h3>
            <input
              value={renameTitle()}
              onInput={(e) => setRenameTitle(e.currentTarget.value)}
              onKeyDown={(e) => e.key === "Enter" && submitRename()}
              maxLength={120}
              autofocus
              class="mt-4 w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none"
            />
            <div class="mt-5 grid grid-cols-2 gap-3">
              <button onClick={() => setRenameOpen(false)} class="bg-white text-gray-600 border-2 border-gray-200 rounded-xl font-black py-3 hover:bg-gray-50">Cancelar</button>
              <button onClick={submitRename} class="bg-blue-800 text-white rounded-xl font-black py-3 hover:bg-blue-900 disabled:opacity-60" disabled={busy() || !renameTitle().trim()}>Guardar</button>
            </div>
          </div>
        </div>
      </Show>

      <Presence>
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
      </Presence>
    </ErrorBoundary>
  );
}