// web/src/routes/admin/tickets/configuracion.tsx
// Configuración del módulo de tickets: CRUD de motivos → estados. Todo ticket
// pertenece a un motivo y cada motivo define su límite de solicitudes abiertas
// por psicólogo (tickets_per_psi). Endpoints bajo /admin/tickets/*.
import { createResource, createSignal, For, Show } from "solid-js";
import { apiDelete, apiGet, apiPatch, apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import type { TicketMotivo, TicketEstado } from "~/types/tickets";
import { Icon } from "~/components/admin/ui/icons";

const IC = "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text";
const IC_SM = "h-8 w-full rounded-md border border-slate-300 bg-white px-2.5 text-xs outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text";
const labelClass = "block text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1 mb-1";

export default function AdminTicketsConfiguracion() {
  const [motivosResource, { refetch }] = createResource(
    () => apiGet<{ data: TicketMotivo[] }>("/admin/tickets/motivos"),
    { initialValue: { data: [] } }
  );
  const motivos = () => motivosResource()?.data ?? [];

  const [flash, setFlash] = createSignal<{ kind: "ok" | "err"; text: string } | null>(null);
  const showFlash = (kind: "ok" | "err", text: string) => {
    setFlash({ kind, text });
    setTimeout(() => setFlash(null), 3500);
  };
  const errMsg = (e: any): string => getUserFacingError(e);
  const reload = () => { refetch(); };

  return (
    <main class="space-y-4 pb-12">
      <div class="flex flex-wrap items-center justify-between gap-3 pb-4 border-b border-colpsi-border">
        <div>
          <h1 class="text-lg font-semibold text-colpsi-text">Configuración de Tickets</h1>
          <p class="text-sm text-colpsi-muted mt-0.5">
            Motivos de atención → estados. Al crear un motivo se siembran los estados por defecto; cada motivo define su límite de solicitudes por psicólogo.
          </p>
        </div>
        <a href="/admin/tickets" class="inline-flex items-center gap-1.5 text-xs font-semibold text-colpsi-muted hover:text-colpsi-blue transition-colors uppercase tracking-wide">
          <Icon name="chevronRight" class="w-3.5 h-3.5 rotate-180" />
          Volver a la cola
        </a>
      </div>

      <Show when={flash()}>
        <div class={`rounded-md px-3 py-2.5 text-sm font-medium border ${flash()?.kind === "ok" ? "bg-emerald-50 border-emerald-200 text-emerald-700" : "bg-red-50 border-red-200 text-red-700"}`}>
          {flash()?.text}
        </div>
      </Show>

      <MotivoCreateForm onDone={(msg) => { showFlash("ok", msg); reload(); }} onError={(e) => showFlash("err", errMsg(e))} />

      <div class="space-y-4">
        <For each={motivos()}>
          {(motivo) => (
            <MotivoCard
              motivo={motivo}
              onDone={(m) => { showFlash("ok", m); reload(); }}
              onError={(e) => showFlash("err", errMsg(e))}
            />
          )}
        </For>
        <Show when={motivos().length === 0 && !motivosResource.loading}>
          <div class="bg-white rounded-lg p-12 text-center border border-colpsi-border">
            <span class="inline-flex h-12 w-12 items-center justify-center rounded-md bg-colpsi-bg text-colpsi-muted mb-4">
              <Icon name="inbox" class="w-6 h-6" />
            </span>
            <h3 class="text-base font-semibold text-colpsi-text">No hay motivos configurados</h3>
            <p class="text-sm text-colpsi-muted mt-1">Crea el primer motivo para habilitar las solicitudes de los psicólogos.</p>
          </div>
        </Show>
      </div>
    </main>
  );
}

// ── Motivo: crear ────────────────────────────────────────────────────────────
function MotivoCreateForm(props: { onDone: (m: string) => void; onError: (e: any) => void }) {
  const [open, setOpen] = createSignal(false);
  const [name, setName] = createSignal("");
  const [description, setDescription] = createSignal("");
  const [perPsi, setPerPsi] = createSignal(3);
  const [saving, setSaving] = createSignal(false);

  const submit = async () => {
    setSaving(true);
    try {
      const m = await apiPost<TicketMotivo>("/admin/tickets/motivos", {
        name: name().trim(),
        description: description().trim() || undefined,
        tickets_per_psi: perPsi(),
      });
      props.onDone(`Motivo "${m.name}" creado con sus estados por defecto.`);
      setOpen(false);
      setName(""); setDescription(""); setPerPsi(3);
    } catch (e) {
      props.onError(e);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div class="bg-white rounded-lg p-5 border border-colpsi-border">
      <Show when={!open()} fallback={
        <div class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class={labelClass}>Nombre del motivo</label>
              <input
                value={name()}
                onInput={(e) => setName(e.currentTarget.value)}
                placeholder="Ej: Solvencia"
                class={IC}
              />
            </div>
            <div>
              <label class={labelClass}>Descripción (opcional)</label>
              <input
                value={description()}
                onInput={(e) => setDescription(e.currentTarget.value)}
                placeholder="Descripción del motivo"
                class={IC}
              />
            </div>
          </div>
          <div class="flex flex-col sm:flex-row gap-3 items-start sm:items-center sm:justify-between">
            <div>
              <label class={labelClass}>Solicitudes abiertas por psicólogo</label>
              <input
                type="number" min={1} max={50}
                value={perPsi()}
                onInput={(e) => setPerPsi(Number(e.currentTarget.value) || 1)}
                class="h-9 w-24 rounded-md border border-slate-300 bg-white px-3 text-sm outline-none focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text"
              />
            </div>
            <div class="flex gap-2">
              <button
                onClick={() => setOpen(false)}
                class="h-9 px-4 rounded-md border border-colpsi-border bg-white text-sm font-medium text-colpsi-text hover:bg-colpsi-bg transition-colors"
              >
                Cancelar
              </button>
              <button
                onClick={submit}
                disabled={saving() || !name().trim()}
                class="inline-flex items-center gap-2 h-9 px-4 rounded-md bg-colpsi-blue text-white text-sm font-semibold hover:bg-colpsi-blue-light transition-all disabled:opacity-40"
              >
                {saving() ? "Creando..." : "Crear motivo"}
              </button>
            </div>
          </div>
        </div>
      }>
        <button
          onClick={() => setOpen(true)}
          class="w-full flex items-center justify-center gap-2 border border-dashed border-colpsi-border rounded-md py-3.5 text-sm font-medium text-colpsi-muted hover:border-colpsi-blue hover:text-colpsi-blue transition-all"
        >
          <Icon name="plus" class="w-4 h-4" />
          Nuevo motivo de atención
        </button>
      </Show>
    </div>
  );
}

// ── Motivo: tarjeta con estados ──────────────────────────────────────────────
function MotivoCard(props: { motivo: TicketMotivo; onDone: (m: string) => void; onError: (e: any) => void }) {
  const { motivo } = props;
  const [editing, setEditing] = createSignal(false);
  const [name, setName] = createSignal(motivo.name);
  const [description, setDescription] = createSignal(motivo.description ?? "");
  const [perPsi, setPerPsi] = createSignal(motivo.tickets_per_psi);
  const [saving, setSaving] = createSignal(false);
  const [busy, setBusy] = createSignal(false);

  const save = async () => {
    setSaving(true);
    try {
      await apiPatch(`/admin/tickets/motivos/${motivo.id}`, {
        name: name().trim() || undefined,
        description: description().trim() || undefined,
        tickets_per_psi: perPsi(),
      });
      setEditing(false);
      props.onDone(`Motivo #${motivo.id} actualizado.`);
    } catch (e) {
      props.onError(e);
    } finally {
      setSaving(false);
    }
  };

  const remove = async () => {
    if (!window.confirm(`¿Eliminar el motivo "${motivo.name}"? Solo se permite si no tiene solicitudes.`)) return;
    setBusy(true);
    try {
      await apiDelete(`/admin/tickets/motivos/${motivo.id}`);
      props.onDone(`Motivo "${motivo.name}" eliminado.`);
    } catch (e) {
      props.onError(e);
    } finally {
      setBusy(false);
    }
  };

  return (
    <div class={`bg-white rounded-lg border border-colpsi-border overflow-hidden ${busy() ? "opacity-40 pointer-events-none" : ""}`}>
      <div class="px-5 py-4 flex flex-wrap items-start justify-between gap-3">
        <div class="min-w-0">
          <div class="flex items-center gap-2">
            <h2 class="text-base font-semibold text-colpsi-text">{motivo.name}</h2>
            <span class="inline-flex items-center px-2 py-0.5 rounded border border-emerald-200 bg-emerald-50 text-emerald-700 text-[10px] font-semibold uppercase tracking-wide">
              {motivo.estados?.length ?? 0} estados
            </span>
          </div>
          <Show when={motivo.description}>
            <p class="text-sm text-colpsi-muted mt-1">{motivo.description}</p>
          </Show>
          <p class="text-[11px] font-medium text-colpsi-muted mt-2 uppercase tracking-wide">
            Límite: {motivo.tickets_per_psi} solicitudes abiertas por psicólogo
          </p>
        </div>
        <div class="flex gap-2">
          <button
            onClick={() => { setEditing(!editing()); setName(motivo.name); setDescription(motivo.description ?? ""); setPerPsi(motivo.tickets_per_psi); }}
            class="inline-flex items-center gap-1.5 h-9 px-3.5 rounded-md border border-colpsi-border text-xs font-medium text-colpsi-muted hover:border-colpsi-blue/40 hover:text-colpsi-blue transition-all bg-white"
          >
            <Icon name="pencil" class="w-3.5 h-3.5" />
            Editar
          </button>
          <button
            onClick={remove}
            class="inline-flex items-center gap-1.5 h-9 px-3.5 rounded-md border border-red-200 text-xs font-medium text-colpsi-red hover:bg-red-50 transition-all bg-white"
          >
            <Icon name="trash" class="w-3.5 h-3.5" />
            Eliminar
          </button>
        </div>
      </div>

      <Show when={editing()}>
        <div class="px-5 pb-5 space-y-4 border-t border-colpsi-border pt-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class={labelClass}>Nombre</label>
              <input
                value={name()}
                onInput={(e) => setName(e.currentTarget.value)}
                placeholder="Nombre del motivo"
                class={IC}
              />
            </div>
            <div>
              <label class={labelClass}>Descripción</label>
              <input
                value={description()}
                onInput={(e) => setDescription(e.currentTarget.value)}
                placeholder="Descripción"
                class={IC}
              />
            </div>
          </div>
          <div class="flex items-end justify-between gap-3">
            <div>
              <label class={labelClass}>Límite de solicitudes por psicólogo</label>
              <input
                type="number" min={1} max={50}
                value={perPsi()}
                onInput={(e) => setPerPsi(Number(e.currentTarget.value) || 1)}
                class="h-9 w-24 rounded-md border border-slate-300 bg-white px-3 text-sm outline-none focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text"
              />
            </div>
            <div class="flex gap-2">
              <button onClick={() => setEditing(false)} class="h-9 px-4 rounded-md border border-colpsi-border bg-white font-medium text-colpsi-muted hover:bg-colpsi-bg transition-all text-sm">Cancelar</button>
              <button onClick={save} disabled={saving() || !name().trim()} class="h-9 px-4 rounded-md bg-colpsi-blue text-white font-semibold hover:bg-colpsi-blue-light transition-all disabled:opacity-40 text-sm">
                {saving() ? "Guardando..." : "Guardar"}
              </button>
            </div>
          </div>
        </div>
      </Show>

      {/* Estados */}
      <div class="px-5 pb-5">
        <p class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide mb-2">Estados del motivo</p>
        <div class="flex flex-wrap gap-2 items-center">
          <For each={motivo.estados ?? []}>
            {(estado) => (
              <EstadoChip estado={estado} onDone={props.onDone} onError={props.onError} />
            )}
          </For>
          <EstadoCreateForm motivoId={motivo.id} onDone={props.onDone} onError={props.onError} />
        </div>
      </div>
    </div>
  );
}

// ── Estado: chip editable ───────────────────────────────────────────────────
function EstadoChip(props: { estado: TicketEstado; onDone: (m: string) => void; onError: (e: any) => void }) {
  const { estado } = props;
  const [editing, setEditing] = createSignal(false);
  const [name, setName] = createSignal(estado.name);
  const [order, setOrder] = createSignal(estado.order);
  const [isClosed, setIsClosed] = createSignal(estado.is_closed);
  const [busy, setBusy] = createSignal(false);

  const save = async () => {
    try {
      await apiPatch(`/admin/tickets/estados/${estado.id}`, {
        name: name().trim() || undefined,
        order: order(),
        is_closed: isClosed(),
      });
      setEditing(false);
      props.onDone(`Estado actualizado.`);
    } catch (e) {
      props.onError(e);
    }
  };

  const remove = async () => {
    if (!window.confirm(`¿Eliminar el estado "${estado.name}"?`)) return;
    setBusy(true);
    try {
      await apiDelete(`/admin/tickets/estados/${estado.id}`);
      props.onDone(`Estado "${estado.name}" eliminado.`);
    } catch (e) {
      props.onError(e);
    } finally {
      setBusy(false);
    }
  };

  return (
    <Show when={!editing()} fallback={
      <div class="bg-white rounded-md border border-colpsi-border px-3 py-2.5 space-y-2 w-full sm:w-auto">
        <div class="flex gap-2">
          <input value={name()} onInput={(e) => setName(e.currentTarget.value)} placeholder="Nombre" class={IC_SM} />
          <input type="number" min={1} value={order()} onInput={(e) => setOrder(Number(e.currentTarget.value) || 1)}
            class="w-16 h-8 rounded-md border border-slate-300 bg-white px-2 text-xs outline-none focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text" title="Orden" />
        </div>
        <label class="flex items-center gap-2 text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide cursor-pointer">
          <input type="checkbox" checked={isClosed()} onChange={(e) => setIsClosed(e.currentTarget.checked)}
            class="accent-colpsi-red w-4 h-4" />
          Estado de cierre
        </label>
        <div class="flex gap-2">
          <button onClick={() => setEditing(false)} class="h-8 px-3 rounded-md border border-colpsi-border bg-white font-medium text-colpsi-muted hover:bg-colpsi-bg transition-all text-xs">Cancelar</button>
          <button onClick={save} disabled={!name().trim()} class="h-8 px-3 rounded-md bg-colpsi-blue text-white font-semibold hover:bg-colpsi-blue-light transition-all disabled:opacity-40 text-xs">Guardar</button>
        </div>
      </div>
    }>
      <span class={`inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-md border ${isClosed() ? "bg-red-50 text-colpsi-red border-red-200" : "bg-blue-50 text-colpsi-blue border-blue-200"} text-[11px] font-medium transition-all`}>
        {estado.name}
        <button onClick={() => { setEditing(true); setName(estado.name); setOrder(estado.order); setIsClosed(estado.is_closed); }}
          class="opacity-60 hover:opacity-100 transition-opacity inline-flex" title="Editar"><Icon name="pencil" class="w-3 h-3" /></button>
        <button onClick={remove} class="opacity-60 hover:opacity-100 transition-opacity inline-flex" title="Eliminar"><Icon name="trash" class="w-3 h-3" /></button>
        {busy() && <span class="animate-pulse">…</span>}
      </span>
    </Show>
  );
}

// ── Formulario de creación de estados ────────────────────────────────────────
function EstadoCreateForm(props: { motivoId: number; onDone: (m: string) => void; onError: (e: any) => void }) {
  const [open, setOpen] = createSignal(false);
  const [name, setName] = createSignal("");
  const [order, setOrder] = createSignal(1);
  const [isClosed, setIsClosed] = createSignal(false);
  const [saving, setSaving] = createSignal(false);

  const submit = async () => {
    setSaving(true);
    try {
      const e = await apiPost<TicketEstado>("/admin/tickets/estados", {
        motivo_id: props.motivoId,
        name: name().trim(),
        order: order(),
        is_closed: isClosed(),
      });
      props.onDone(`Estado "${e.name}" creado.`);
      setOpen(false); setName(""); setOrder(1); setIsClosed(false);
    } catch (er) {
      props.onError(er);
    } finally {
      setSaving(false);
    }
  };

  return (
    <Show when={open()} fallback={
      <button onClick={() => setOpen(true)} class="inline-flex items-center gap-1.5 px-3 h-8 rounded-md border border-dashed border-colpsi-border text-[11px] font-medium text-colpsi-muted hover:border-colpsi-blue hover:text-colpsi-blue transition-all">
        <Icon name="plus" class="w-3.5 h-3.5" />
        Estado
      </button>
    }>
      <div class="bg-white rounded-md border border-colpsi-border px-3 py-2.5 space-y-2">
        <div class="flex gap-2">
          <input value={name()} onInput={(e) => setName(e.currentTarget.value)} placeholder="Nombre" class={`${IC_SM} min-w-0 flex-1`} />
          <input type="number" min={1} value={order()} onInput={(e) => setOrder(Number(e.currentTarget.value) || 1)}
            class="w-16 h-8 rounded-md border border-slate-300 bg-white px-2 text-xs outline-none focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text" title="Orden" />
        </div>
        <label class="flex items-center gap-2 text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide cursor-pointer">
          <input type="checkbox" checked={isClosed()} onChange={(e) => setIsClosed(e.currentTarget.checked)} class="accent-colpsi-red w-4 h-4" />
          Estado de cierre
        </label>
        <div class="flex gap-2">
          <button onClick={() => setOpen(false)} class="h-8 px-3 rounded-md border border-colpsi-border bg-white font-medium text-colpsi-muted hover:bg-colpsi-bg transition-all text-xs">Cancelar</button>
          <button onClick={submit} disabled={saving() || !name().trim()} class="h-8 px-3 rounded-md bg-colpsi-blue text-white font-semibold hover:bg-colpsi-blue-light transition-all disabled:opacity-40 text-xs">
            {saving() ? "Creando..." : "Crear"}
          </button>
        </div>
      </div>
    </Show>
  );
}