// web/src/components/psi/profile/EmergencyContactsSection.tsx
//
// Gestión de las personas de contacto para emergencias. Estos datos son de un
// TERCERO: los usa el Colegio para localizar al profesional ante un accidente o
// cuando no puede ser localizado, y NUNCA se publican en el directorio.
//
// Es la MISMA sección en los dos bandos: la usa el agremiado en su perfil
// (`audience="psi"`, dentro del `<form>` del perfil, como pestaña del Notebook) y
// el admin en la ficha del agremiado (`audience="admin"`, en el Notebook de
// gestión, fuera del formulario del expediente). Solo cambian los textos del
// aviso. En la versión psi:
//
//   - los botones son `type="button"` para NO disparar el guardado del perfil;
//   - no se anida otro `<form>` (HTML inválido que el navegador descartaría al
//     parsear el SSR): el envío se dispara con el botón o con Enter, que se
//     intercepta para evitar el submit implícito del formulario padre.

import { For, Show, createEffect, createSignal } from "solid-js";
import type { EmergencyContact } from "~/types/psi";
import { Icon } from "~/components/admin/ui/icons";
import {
  MAX_EMERGENCY_CONTACTS,
  OTHER_RELATIONSHIP,
  RELATIONSHIPS,
  relationshipOption,
  resolveRelationship,
  toEmergencyPayload,
  validateEmergencyContact,
  type EmergencyContactDraft,
  type EmergencyContactError,
} from "~/lib/emergency-contact";

type Payload = ReturnType<typeof toEmergencyPayload>;

interface Props {
  contacts: EmergencyContact[];
  onAdd: (payload: Payload) => Promise<void>;
  onUpdate: (id: string, payload: Payload) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
  /** Ajusta los textos del aviso de privacidad (autogestión vs. ficha admin). */
  audience?: "psi" | "admin";
}

const EMPTY: EmergencyContactDraft = { name: "", relationship: "", phone: "", email: "" };

const inputClass =
  "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm text-colpsi-text placeholder:text-slate-400 outline-none transition-colors focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15";
const labelClass = "block text-xs font-medium text-colpsi-muted mb-1";

export function EmergencyContactsSection(props: Props) {
  const [draft, setDraft] = createSignal<EmergencyContactDraft>({ ...EMPTY });
  // "" (sin elegir) | un valor del catálogo | OTHER_RELATIONSHIP (texto libre)
  const [selected, setSelected] = createSignal("");
  const [custom, setCustom] = createSignal("");
  const [editingId, setEditingId] = createSignal<string | null>(null);
  const [pendingDelete, setPendingDelete] = createSignal<string | null>(null);
  const [error, setError] = createSignal<EmergencyContactError | null>(null);
  const [message, setMessage] = createSignal<string | null>(null);
  const [saving, setSaving] = createSignal(false);

  const full = () => props.contacts.length >= MAX_EMERGENCY_CONTACTS;

  // El setter de un SIGNAL solo recibe el valor completo (la firma con ruta
  // `set("campo", v)` es de createStore), así que el borrador se actualiza
  // spreads sobre el objeto anterior.
  const patch = (part: Partial<EmergencyContactDraft>) =>
    setDraft((prev) => ({ ...prev, ...part }));

  const reset = () => {
    setDraft({ ...EMPTY });
    setSelected("");
    setCustom("");
    setEditingId(null);
    setError(null);
  };

  const loadIntoForm = (contact: EmergencyContact) => {
    setDraft({
      name: contact.name,
      relationship: contact.relationship,
      phone: contact.phone ?? "",
      email: contact.email ?? "",
    });
    setSelected(relationshipOption(contact.relationship));
    setCustom(contact.relationship);
    setError(null);
  };

  // Si el contacto que se estaba editando desaparece (borrado en otra pestaña o
  // refetch), se sale del modo edición en vez de dejar un formulario fantasma.
  createEffect(() => {
    const id = editingId();
    if (id && !props.contacts.some((c) => c.id === id)) setEditingId(null);
  });

  const submit = async () => {
    setMessage(null);
    const relationship = resolveRelationship(selected(), custom());
    const invalid = validateEmergencyContact({ ...draft(), relationship });
    if (invalid) {
      setError(invalid);
      return;
    }
    setError(null);
    setSaving(true);
    try {
      const payload = toEmergencyPayload({ ...draft(), relationship });
      const id = editingId();
      if (id) {
        await props.onUpdate(id, payload);
        setMessage("Contacto actualizado.");
      } else {
        await props.onAdd(payload);
        setMessage("Contacto registrado.");
      }
      reset();
    } catch (err: any) {
      setError({ field: "general", message: err?.message || "No se pudo guardar." });
    } finally {
      setSaving(false);
    }
  };

  // Enter envía el formulario de esta sección en vez del perfil completo.
  const onEnter = (e: KeyboardEvent) => {
    if (e.key !== "Enter") return;
    e.preventDefault();
    void submit();
  };

  const confirmDelete = async () => {
    const id = pendingDelete();
    if (!id) return;
    setSaving(true);
    try {
      await props.onDelete(id);
      if (editingId() === id) reset();
      setMessage("Contacto eliminado.");
    } catch (err: any) {
      setError({ field: "general", message: err?.message || "No se pudo eliminar." });
    } finally {
      setPendingDelete(null);
      setSaving(false);
    }
  };

  const isOther = () => selected() === OTHER_RELATIONSHIP;
  const errorFor = (field: string) => (error()?.field === field ? error()!.message : null);

  return (
    <section class="space-y-4">
      {/* ── Aviso de privacidad ────────────────────────────────────────── */}
      <div class="rounded-md border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900 leading-relaxed flex gap-2.5">
        <Icon name="shield" class="w-4 h-4 shrink-0 mt-0.5" />
        <p>
          {props.audience === "admin" ? (
            <>
              Datos de un <strong>tercero</strong>, declarados por el agremiado.
              El Colegio los consulta únicamente cuando no puede localizarlo o
              ante un accidente; <strong>no se publican</strong> en el directorio.
              Máximo {MAX_EMERGENCY_CONTACTS} personas.
            </>
          ) : (
            <>
              Estos datos <strong>no se publican</strong> en tu perfil del
              directorio. Solo el Colegio los consulta, y únicamente cuando no
              pueda localizarte o ante un accidente. Registra hasta{" "}
              {MAX_EMERGENCY_CONTACTS} personas.
            </>
          )}
        </p>
      </div>

      {/* ── Listado ────────────────────────────────────────────────────── */}
      <Show
        when={props.contacts.length > 0}
        fallback={
          <p class="text-sm text-colpsi-muted">
            {props.audience === "admin"
              ? "El agremiado no ha registrado ninguna persona de contacto."
              : "Todavía no registras ninguna persona de contacto."}
          </p>
        }
      >
        <ul class="space-y-2">
          <For each={props.contacts}>
            {(contact) => (
              <li
                class="flex flex-wrap items-center justify-between gap-3 p-3 rounded-md border border-colpsi-border bg-colpsi-bg/40"
                classList={{ "ring-2 ring-colpsi-blue/25": editingId() === contact.id }}
              >
                <div class="min-w-0">
                  <p class="text-sm font-semibold text-colpsi-text">
                    {contact.name}
                    <span class="ml-2 text-xs font-normal text-colpsi-muted">
                      {contact.relationship}
                    </span>
                  </p>
                  <p class="text-xs text-colpsi-muted mt-0.5 truncate">
                    {[contact.phone, contact.email].filter(Boolean).join(" · ")}
                  </p>
                </div>
                <div class="flex items-center gap-1">
                  <button
                    type="button"
                    onClick={() => {
                      if (editingId() === contact.id) reset();
                      else {
                        loadIntoForm(contact);
                        setEditingId(contact.id!);
                      }
                    }}
                    class="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-medium text-colpsi-blue hover:bg-colpsi-blue/10 transition-colors"
                  >
                    <Icon name={editingId() === contact.id ? "x" : "pencil"} class="w-3.5 h-3.5" />
                    {editingId() === contact.id ? "Cancelar" : "Editar"}
                  </button>
                  <button
                    type="button"
                    onClick={() => setPendingDelete(contact.id!)}
                    class="p-1.5 rounded-md text-slate-400 hover:text-red-500 hover:bg-red-50 transition-colors"
                    title="Eliminar persona de contacto"
                  >
                    <Icon name="trash" class="w-4 h-4" />
                  </button>
                </div>
              </li>
            )}
          </For>
        </ul>
      </Show>

      {/* ── Alta / edición ─────────────────────────────────────────────── */}
      {/* El mismo formulario sirve para alta y edición: solo se oculta cuando la
          cuota está llena Y no hay una edición en curso (si se escondiera al
          editar, el usuario perdería los datos que acaba de cargar). */}
      <Show
        when={!full() || !!editingId()}
        fallback={
          <p class="text-sm text-colpsi-muted">
            {props.audience === "admin"
              ? `El agremiado alcanzó el máximo de ${MAX_EMERGENCY_CONTACTS} personas. Edite o elimine una para registrar otra.`
              : `Alcanzaste el máximo de ${MAX_EMERGENCY_CONTACTS} personas. Edita o elimina una para registrar otra.`}
          </p>
        }
      >
        <div class="bg-colpsi-bg/50 p-4 rounded-md border border-colpsi-border space-y-3">
          <Show when={editingId()}>
            <div class="flex items-center justify-between gap-3 border-b border-colpsi-border pb-2">
              <p class="text-sm font-semibold text-colpsi-text">
                Editando a {draft().name}
              </p>
              <button
                type="button"
                onClick={reset}
                class="inline-flex items-center gap-1 text-xs font-medium text-colpsi-muted hover:text-colpsi-blue transition-colors"
              >
                <Icon name="x" class="w-3.5 h-3.5" />
                Cancelar
              </button>
            </div>
          </Show>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label class={labelClass} for="emerg-name">
                Nombre de la persona *
              </label>
              <input
                id="emerg-name"
                type="text"
                class={inputClass}
                placeholder="Ej: María Rodríguez"
                value={draft().name}
                onInput={(e) => patch({ name: e.currentTarget.value })}
                onKeyDown={onEnter}
              />
              <Show when={errorFor("name")}>
                <p class="mt-1 text-xs text-red-600">{errorFor("name")}</p>
              </Show>
            </div>

            <div>
              <label class={labelClass} for="emerg-rel">
                Parentesco o vínculo *
              </label>
              <select
                id="emerg-rel"
                class={inputClass}
                value={selected()}
                onChange={(e) => setSelected(e.currentTarget.value)}
                onKeyDown={onEnter}
              >
                <option value="" disabled>
                  Selecciona el vínculo…
                </option>
                <For each={RELATIONSHIPS}>
                  {(r) => <option value={r.value}>{r.label}</option>}
                </For>
                <option value={OTHER_RELATIONSHIP}>Otro…</option>
              </select>
              <Show when={isOther()}>
                <input
                  type="text"
                  class={`${inputClass} mt-2`}
                  placeholder="Ej: tía, pareja, supervisor"
                  value={custom()}
                  onInput={(e) => setCustom(e.currentTarget.value)}
                  onKeyDown={onEnter}
                />
              </Show>
              <Show when={errorFor("relationship")}>
                <p class="mt-1 text-xs text-red-600">{errorFor("relationship")}</p>
              </Show>
            </div>

            <div>
              <label class={labelClass} for="emerg-phone">
                Teléfono
              </label>
              <input
                id="emerg-phone"
                type="tel"
                class={inputClass}
                placeholder="Ej: 0412 1234567"
                value={draft().phone}
                onInput={(e) => patch({ phone: e.currentTarget.value })}
                onKeyDown={onEnter}
              />
            </div>

            <div>
              <label class={labelClass} for="emerg-email">
                Correo electrónico
              </label>
              <input
                id="emerg-email"
                type="email"
                class={inputClass}
                placeholder="Ej: maria@correo.com"
                value={draft().email}
                onInput={(e) => patch({ email: e.currentTarget.value })}
                onKeyDown={onEnter}
              />
            </div>
          </div>

          <p class="text-xs text-colpsi-muted">
            <span class="text-red-500">*</span> Nombre y parentesco son obligatorios.
            Indica al menos un teléfono <strong>o</strong> un correo.
          </p>
          <Show when={errorFor("phone")}>
            <p class="text-xs text-red-600">{errorFor("phone")}</p>
          </Show>
          <Show when={errorFor("email")}>
            <p class="text-xs text-red-600">{errorFor("email")}</p>
          </Show>
          <Show when={errorFor("general")}>
            <p class="text-xs text-red-600">{errorFor("general")}</p>
          </Show>
          <Show when={message()}>
            <p class="text-xs text-emerald-700">{message()}</p>
          </Show>

          <button
            type="button"
            onClick={() => void submit()}
            disabled={saving()}
            class="h-9 px-4 rounded-md bg-colpsi-blue text-white text-sm font-semibold hover:bg-colpsi-blue-light transition-colors disabled:opacity-70 disabled:cursor-not-allowed"
          >
            {saving() ? "Guardando..." : editingId() ? "Guardar cambios" : "Agregar contacto"}
          </button>
        </div>
      </Show>

      {/* ── Confirmación de borrado ────────────────────────────────────── */}
      <Show when={pendingDelete()}>
        <div class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 backdrop-blur-sm">
          <div class="bg-white rounded-lg p-5 shadow-lg max-w-sm w-full mx-4 border border-colpsi-border">
            <div class="flex items-center gap-3 mb-4">
              <span class="w-9 h-9 rounded-full bg-red-50 flex items-center justify-center text-red-600">
                <Icon name="alertTriangle" class="w-4 h-4" />
              </span>
              <div>
                <h3 class="text-base font-semibold text-colpsi-text">
                  ¿Eliminar este contacto de emergencia?
                </h3>
                <p class="text-sm text-colpsi-muted">
                  El Colegio ya no podrá avisarle en una emergencia.
                </p>
              </div>
            </div>
            <div class="flex justify-end gap-2">
              <button
                type="button"
                onClick={() => setPendingDelete(null)}
                class="px-4 py-2 rounded-md text-sm font-medium text-colpsi-muted hover:bg-colpsi-bg transition-colors"
              >
                Cancelar
              </button>
              <button
                type="button"
                onClick={() => void confirmDelete()}
                disabled={saving()}
                class="px-4 py-2 rounded-md bg-red-600 text-white text-sm font-semibold hover:bg-red-700 transition-colors disabled:opacity-70"
              >
                Eliminar
              </button>
            </div>
          </div>
        </div>
      </Show>
    </section>
  );
}
