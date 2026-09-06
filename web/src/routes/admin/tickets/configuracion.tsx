// web/src/routes/admin/tickets/configuracion.tsx
// Configuración del módulo de tickets: CRUD de motivos → estados. Todo ticket
// pertenece a un motivo y cada motivo define su límite de solicitudes abiertas
// por psicólogo (tickets_per_psi). Endpoints bajo /admin/tickets/*.
import { createResource, createSignal, For, Show } from "solid-js";
import { apiDelete, apiGet, apiPatch, apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import type { TicketMotivo, TicketEstado } from "~/types/tickets";

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
    <main class="space-y-5">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 class="text-2xl font-black text-gray-800 flex items-center gap-2">⚙️ Configuración de Tickets</h1>
          <p class="text-sm text-gray-400 mt-1">
            Motivos de atención → estados. Al crear un motivo se siembran los estados por defecto; cada motivo define su límite de solicitudes por psicólogo.
          </p>
        </div>
        <a href="/admin/tickets" class="text-xs font-black text-gray-400 uppercase tracking-widest hover:text-colpsi-blue transition-all">
          ← Volver a la cola
        </a>
      </div>

      <Show when={flash()}>
        <div class={`rounded-2xl px-4 py-3 text-sm font-bold border ${flash()?.kind === "ok" ? "bg-emerald-50 border-emerald-200 text-emerald-700" : "bg-red-50 border-red-200 text-red-700"}`}>
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
          <div class="bg-white rounded-3xl p-12 text-center shadow-sm border border-colpsi-border">
            <p class="text-5xl mb-4">🗂️</p>
            <h3 class="font-black text-gray-700">No hay motivos configurados</h3>
            <p class="text-sm text-gray-500 mt-1">Crea el primer motivo para habilitar las solicitudes de los psicólogos.</p>
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
    <div class="bg-white rounded-3xl p-5 shadow-sm border border-colpsi-border">
      <Show when={!open()} fallback={
        <div class="space-y-3">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <input
              value={name()}
              onInput={(e) => setName(e.currentTarget.value)}
              placeholder="Nombre del motivo (ej: Solvencia)"
              class="w-full px-4 py-3 rounded-2xl border-2 border-colpsi-border bg-colpsi-surface outline-none focus:border-colpsi-yellow text-sm font-semibold text-gray-800 transition-all"
            />
            <input
              value={description()}
              onInput={(e) => setDescription(e.currentTarget.value)}
              placeholder="Descripción (opcional)"
              class="w-full px-4 py-3 rounded-2xl border-2 border-colpsi-border bg-colpsi-surface outline-none focus:border-colpsi-yellow text-sm font-semibold text-gray-800 transition-all"
            />
          </div>
          <div class="flex flex-col sm:flex-row gap-3 items-start sm:items-center sm:justify-between">
            <label class="text-sm font-bold text-gray-500">
              Solicitudes abiertas por psicólogo:
              <input
                type="number" min={1} max={50}
                value={perPsi()}
                onInput={(e) => setPerPsi(Number(e.currentTarget.value) || 1)}
                class="ml-2 w-20 px-3 py-2 rounded-xl border-2 border-colpsi-border bg-colpsi-surface outline-none focus:border-colpsi-yellow text-sm font-semibold text-gray-800 transition-all"
              />
            </label>
            <div class="flex gap-2">
              <button
                onClick={() => setOpen(false)}
                class="px-5 py-3 rounded-2xl border-2 border-colpsi-border font-black text-gray-400 hover:bg-colpsi-surface transition-all text-xs uppercase tracking-widest"
              >
                Cancelar
              </button>
              <button
                onClick={submit}
                disabled={saving() || !name().trim()}
                class="px-5 py-3 rounded-2xl bg-colpsi-blue text-white font-black text-xs uppercase tracking-widest transition-all active:scale-95 disabled:opacity-40"
              >
                {saving() ? "Creando..." : "Crear motivo"}
              </button>
            </div>
          </div>
        </div>
      }>
        <button
          onClick={() => setOpen(true)}
          class="w-full flex items-center justify-center gap-2 border-2 border-dashed border-gray-200 rounded-2xl py-4 text-sm font-black text-gray-400 hover:border-colpsi-blue hover:text-colpsi-blue transition-all"
        >
          ＋ Nuevo motivo de atención
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
    <div class={`bg-white rounded-3xl shadow-sm border border-colpsi-border overflow-hidden ${busy() ? "opacity-40 pointer-events-none" : ""}`}>
      <div class="px-6 py-5 flex flex-wrap items-start justify-between gap-3">
        <div class="min-w-0">
          <div class="flex items-center gap-2">
            <h2 class="font-black text-gray-800 text-lg">{motivo.name}</h2>
            <span class="text-[10px] font-black px-2.5 py-1 rounded-full bg-emerald-50 text-emerald-600 uppercase tracking-wider">
              {motivo.estados?.length ?? 0} estados
            </span>
          </div>
          <Show when={motivo.description}>
            <p class="text-sm text-gray-500 mt-1">{motivo.description}</p>
          </Show>
          <p class="text-[11px] font-bold text-gray-400 mt-2 uppercase tracking-wider">
            Límite: {motivo.tickets_per_psi} solicitudes abiertas por psicólogo
          </p>
        </div>
        <div class="flex gap-2">
          <button
            onClick={() => { setEditing(!editing()); setName(motivo.name); setDescription(motivo.description ?? ""); setPerPsi(motivo.tickets_per_psi); }}
            class="px-4 py-2 rounded-xl border-2 border-colpsi-border text-xs font-black text-gray-500 hover:border-blue-200 hover:text-colpsi-blue transition-all"
          >
            ✏️ Editar
          </button>
          <button
            onClick={remove}
            class="px-4 py-2 rounded-xl border-2 border-red-100 text-xs font-black text-red-500 hover:bg-red-600 hover:text-white hover:border-red-600 transition-all"
          >
            🗑️ Eliminar
          </button>
        </div>
      </div>

      <Show when={editing()}>
        <div class="px-6 pb-5 space-y-3 border-t border-colpsi-border pt-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <input
              value={name()}
              onInput={(e) => setName(e.currentTarget.value)}
              placeholder="Nombre"
              class="w-full px-4 py-2.5 rounded-2xl border-2 border-colpsi-border bg-colpsi-surface outline-none focus:border-colpsi-yellow text-sm font-semibold text-gray-800 transition-all"
            />
            <input
              value={description()}
              onInput={(e) => setDescription(e.currentTarget.value)}
              placeholder="Descripción"
              class="w-full px-4 py-2.5 rounded-2xl border-2 border-colpsi-border bg-colpsi-surface outline-none focus:border-colpsi-yellow text-sm font-semibold text-gray-800 transition-all"
            />
          </div>
          <div class="flex items-center justify-between gap-3">
            <label class="text-sm font-bold text-gray-500">
              Límite de solicitudes por psicólogo:
              <input
                type="number" min={1} max={50}
                value={perPsi()}
                onInput={(e) => setPerPsi(Number(e.currentTarget.value) || 1)}
                class="ml-2 w-20 px-3 py-2 rounded-xl border-2 border-colpsi-border bg-colpsi-surface outline-none focus:border-colpsi-yellow text-sm font-semibold text-gray-800 transition-all"
              />
            </label>
            <div class="flex gap-2">
              <button onClick={() => setEditing(false)} class="px-4 py-2.5 rounded-xl border-2 border-colpsi-border font-black text-gray-400 text-xs uppercase tracking-widest hover:bg-colpsi-surface transition-all">Cancelar</button>
              <button onClick={save} disabled={saving() || !name().trim()} class="px-4 py-2.5 rounded-xl bg-colpsi-blue text-white font-black text-xs uppercase tracking-widest transition-all active:scale-95 disabled:opacity-40">
                {saving() ? "Guardando..." : "Guardar"}
              </button>
            </div>
          </div>
        </div>
      </Show>

      {/* Estados */}
      <div class="px-6 pb-6">
        <p class="text-[10px] font-black text-gray-400 uppercase tracking-widest mb-2">Estados del motivo</p>
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
      <div class="bg-white rounded-2xl border-2 border-blue-200 px-3 py-2 space-y-2 w-full sm:w-auto">
        <div class="flex gap-2">
          <input value={name()} onInput={(e) => setName(e.currentTarget.value)} placeholder="Nombre"
            class="flex-1 px-3 py-1.5 rounded-lg border border-colpsi-border outline-none focus:border-colpsi-yellow text-xs font-semibold text-gray-800 transition-all" />
          <input type="number" min={1} value={order()} onInput={(e) => setOrder(Number(e.currentTarget.value) || 1)}
            class="w-16 px-2 py-1.5 rounded-lg border border-colpsi-border outline-none focus:border-colpsi-yellow text-xs font-semibold text-gray-800 transition-all" title="Orden" />
        </div>
        <label class="flex items-center gap-2 text-[11px] font-black text-gray-500 uppercase tracking-wider cursor-pointer">
          <input type="checkbox" checked={isClosed()} onChange={(e) => setIsClosed(e.currentTarget.checked)}
            class="accent-red-500 w-4 h-4" />
          Estado de cierre
        </label>
        <div class="flex gap-2">
          <button onClick={() => setEditing(false)} class="px-3 py-1.5 rounded-lg border border-gray-200 font-black text-gray-400 text-[10px] uppercase tracking-widest hover:bg-colpsi-surface transition-all">Cancelar</button>
          <button onClick={save} disabled={!name().trim()} class="px-3 py-1.5 rounded-lg bg-colpsi-blue text-white font-black text-[10px] uppercase tracking-widest transition-all active:scale-95 disabled:opacity-40">Guardar</button>
        </div>
      </div>
    }>
      <span class={`inline-flex items-center gap-1.5 px-3 py-2 rounded-2xl ${isClosed() ? "bg-red-100 text-red-700" : "bg-blue-100 text-blue-700"} text-[11px] font-black transition-all`}>
        {estado.name}
        <button onClick={() => { setEditing(true); setName(estado.name); setOrder(estado.order); setIsClosed(estado.is_closed); }}
          class="opacity-60 hover:opacity-100 transition-opacity" title="Editar">✏️</button>
        <button onClick={remove} class="opacity-60 hover:opacity-100 transition-opacity" title="Eliminar">🗑️</button>
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
      <button onClick={() => setOpen(true)} class="px-3 py-2 rounded-2xl border-2 border-dashed border-blue-200 text-[10px] font-black text-blue-500 hover:border-colpsi-blue hover:text-colpsi-blue transition-all">
        ＋ Estado
      </button>
    }>
      <div class="bg-white rounded-2xl border border-blue-200 px-3 py-2 space-y-2">
        <div class="flex gap-2">
          <input value={name()} onInput={(e) => setName(e.currentTarget.value)} placeholder="Nombre"
            class="flex-1 min-w-0 px-3 py-1.5 rounded-lg border border-colpsi-border outline-none focus:border-colpsi-yellow text-xs font-semibold text-gray-800 transition-all" />
          <input type="number" min={1} value={order()} onInput={(e) => setOrder(Number(e.currentTarget.value) || 1)}
            class="w-16 px-2 py-1.5 rounded-lg border border-colpsi-border outline-none focus:border-colpsi-yellow text-xs font-semibold text-gray-800 transition-all" title="Orden" />
        </div>
        <label class="flex items-center gap-2 text-[11px] font-black text-gray-500 uppercase tracking-wider cursor-pointer">
          <input type="checkbox" checked={isClosed()} onChange={(e) => setIsClosed(e.currentTarget.checked)} class="accent-red-500 w-4 h-4" />
          Estado de cierre
        </label>
        <div class="flex gap-2">
          <button onClick={() => setOpen(false)} class="px-3 py-1.5 rounded-lg border border-gray-200 font-black text-gray-400 text-[10px] uppercase tracking-widest hover:bg-colpsi-surface transition-all">Cancelar</button>
          <button onClick={submit} disabled={saving() || !name().trim()} class="px-3 py-1.5 rounded-lg bg-colpsi-blue text-white font-black text-[10px] uppercase tracking-widest transition-all active:scale-95 disabled:opacity-40">
            {saving() ? "Creando..." : "Crear"}
          </button>
        </div>
      </div>
    </Show>
  );
}