// web/src/routes/admin/inscripciones/[id].tsx
import { createResource, createSignal, createEffect, Show, For } from "solid-js";
import { createStore, unwrap } from "solid-js/store";
import { useParams, useNavigate } from "@solidjs/router";
import { apiGet, apiPost, apiPatch, apiDelete, ApiError } from "~/lib/api";
import { bucketUrl } from "~/lib/bucket";
import { ImageModal } from "~/components/ui/ImageModal";
import { Notebook, NotebookPage } from "~/components/ui/Notebook";
import { Field, IC } from "~/components/admin/psicologos/edit/EditPrimitives";
import { Icon } from "~/components/admin/ui/icons";
import { MUNICIPIOS_CARABOBO, ESTADOS_VENEZUELA, municipiosDe } from "~/lib/geo";
import type {
  InscriptionDetail,
  InscriptionDocument,
  InscriptionDocumentType,
  UpdateInscriptionRequest,
  WorkArea,
} from "~/types/inscription";

const formatDate = (dateStr?: string | null) => (dateStr ? dateStr.split("T")[0] : "");

const isImageUrl = (url?: string) => !!url && /\.(png|jpe?g|gif|webp|bmp|svg)(\?|#|$)/i.test(url);

const DOC_SPECS: { type: InscriptionDocumentType; label: string; required: boolean }[] = [
  { type: "cedula", label: "Cédula de Identidad", required: true },
  { type: "titulo", label: "Título de Psicólogo", required: true },
  { type: "rif", label: "RIF", required: true },
  { type: "otro", label: "Otro documento", required: false },
];

export default function AdminInscriptionDetail() {
  const params = useParams();
  const navigate = useNavigate();
  const [detail, { refetch }] = createResource<InscriptionDetail>(() =>
    apiGet(`/admin/inscripciones/${params.id}`)
  );
  const [workAreas] = createResource<WorkArea[]>(() => apiGet("/specialties"));

  const [fichaMsg, setFichaMsg] = createSignal<{ type: "ok" | "err"; text: string } | null>(null);
  const [savingFicha, setSavingFicha] = createSignal(false);
  const [busy, setBusy] = createSignal(false);
  const [feedback, setFeedback] = createSignal<{ type: "ok" | "err"; text: string } | null>(null);
  const [modalImage, setModalImage] = createSignal<{ src: string; alt: string } | null>(null);
  const [confirmApprove, setConfirmApprove] = createSignal(false);
  const [confirmReject, setConfirmReject] = createSignal(false);

  const closeModal = () => setModalImage(null);

  // ── Form store (ficha completa, reemplazo al guardar) ────────────────────
  const [form, setForm] = createStore<UpdateInscriptionRequest>({} as UpdateInscriptionRequest);

  const syncForm = (d: InscriptionDetail) => {
    setForm({
      cedula: d.cedula,
      nacionalidad: d.nacionalidad || "V",
      nombres: d.nombres ?? "",
      apellidos: d.apellidos ?? "",
      segundo_nombre: d.segundo_nombre ?? "",
      segundo_apellido: d.segundo_apellido ?? "",
      genero: d.genero ?? "",
      fpv: d.fpv ?? 0,
      telefono: d.telefono ?? "",
      correo: d.correo ?? "",
      fecha_nacimiento: formatDate(d.fecha_nacimiento) || null,
      titulo_universidad: d.titulo_universidad ?? "",
      titulo_fecha_graduacion: formatDate(d.titulo_fecha_graduacion) || null,
      titulo_mencion: d.titulo_mencion ?? "",
      titulo_registro_numero: d.titulo_registro_numero ?? "",
      titulo_registro_estado: d.titulo_registro_estado ?? "",
      titulo_registro_tomo: d.titulo_registro_tomo ?? "",
      titulo_registro_folio: d.titulo_registro_folio ?? "",
      rif: d.rif ?? "",
      service_address: d.service_address ?? "",
      municipality_carabobo: d.municipality_carabobo ?? "",
      state_outside: d.state_outside ?? "",
      municipality_outside_carabobo: d.municipality_outside_carabobo ?? "",
      country: d.country ?? "",
      service_modality_presencial: d.service_modality_presencial ?? false,
      service_modality_distance: d.service_modality_distance ?? false,
      service_modality_telephone: d.service_modality_telephone ?? false,
      primary_specialty_id: d.primary_specialty_id ?? null,
      secondary_specialty_id: d.secondary_specialty_id ?? null,
    });
  };

  // Solo resincronizamos desde el servidor en la primera carga: así las ediciones
  // pendientes no se pierden cuando un upload de foto/documento dispara un refetch.
  let formSynced = false;
  createEffect(() => {
    const d = detail();
    if (!d || formSynced) return;
    formSynced = true;
    syncForm(d);
  });

  const set = (key: keyof UpdateInscriptionRequest, value: any) => setForm(key as any, value);

  // Regla de campos obligatorios de la ficha (misma que el backend): personales,
  // académicos y al menos un bloque de ubicación completo.
  const fichaIncompleta = (): string => {
    const f = unwrap(form);
    const s = (v: any) => String(v ?? "").trim();
    if (!s(f.segundo_apellido)) return "El segundo apellido es obligatorio";
    if (!s(f.genero)) return "El género es obligatorio";
    if (!s(f.telefono)) return "El teléfono de contacto es obligatorio";
    if (!s(f.fecha_nacimiento)) return "La fecha de nacimiento es obligatoria";
    if (!s(f.titulo_universidad)) return "La universidad es obligatoria";
    if (!s(f.titulo_fecha_graduacion)) return "La fecha de graduación es obligatoria";
    if (!s(f.titulo_registro_estado)) return "El estado del registro es obligatorio";
    const cedulaNum = parseInt(s(f.cedula), 10);
    if (!s(f.cedula) || !Number.isFinite(cedulaNum) || cedulaNum <= 0) {
      return "La cédula es obligatoria y debe ser un número positivo";
    }
    const fpvNum = Number(f.fpv);
    if (f.fpv !== "" && f.fpv !== null && (Number.isNaN(fpvNum) || fpvNum < 0)) {
      return "El N° FPV debe ser un número positivo";
    }
    const carabobo = s(f.municipality_carabobo) !== "" && s(f.service_address) !== "";
    const otroEstado = s(f.state_outside) !== "" && s(f.municipality_outside_carabobo) !== "";
    const exterior = s(f.country) !== "";
    if (!carabobo && !otroEstado && !exterior) {
      return "Debes completar al menos una ubicación completa (Carabobo, otro estado o exterior)";
    }
    return "";
  };

  const saveFicha = async () => {
    if (savingFicha()) return;
    const incompleta = fichaIncompleta();
    if (incompleta) {
      setFichaMsg({ type: "err", text: incompleta });
      return;
    }
    setSavingFicha(true);
    setFichaMsg(null);
    const toInt = (v: any) => {
      const n = parseInt(String(v ?? ""), 10);
      return Number.isFinite(n) ? n : 0;
    };
    const toOptId = (v: any) => {
      const n = parseInt(String(v ?? ""), 10);
      return v !== null && v !== "" && Number.isFinite(n) && n > 0 ? n : null;
    };
    try {
      const raw = unwrap(form);
      const payload: UpdateInscriptionRequest = {
        ...raw,
        cedula: toInt(raw.cedula),
        fpv: toInt(raw.fpv),
        primary_specialty_id: toOptId(raw.primary_specialty_id),
        secondary_specialty_id: toOptId(raw.secondary_specialty_id),
        fecha_nacimiento: raw.fecha_nacimiento || null,
        titulo_fecha_graduacion: raw.titulo_fecha_graduacion || null,
      };
      await apiPatch(`/admin/inscripciones/${params.id}`, payload);
      setFichaMsg({ type: "ok", text: "Ficha guardada correctamente." });
      const updated = await apiGet<InscriptionDetail>(`/admin/inscripciones/${params.id}`);
      syncForm(updated);
      refetch();
      window.scrollTo({ top: 0, behavior: "smooth" });
    } catch (err) {
      setFichaMsg({ type: "err", text: err instanceof ApiError ? err.message : "Error al guardar la ficha." });
    } finally {
      setSavingFicha(false);
    }
  };

  // ── Reemplazo de foto / comprobante ──────────────────────────────────────
  const replacePhoto = async (kind: "foto" | "comprobante", file: File) => {
    setFichaMsg(null);
    try {
      const fd = new FormData();
      fd.set("kind", kind);
      fd.set("file", file);
      await apiPost(`/admin/inscripciones/${params.id}/photo`, fd);
      setFichaMsg({ type: "ok", text: kind === "foto" ? "Foto reemplazada." : "Comprobante reemplazado." });
      refetch();
    } catch (err) {
      setFichaMsg({ type: "err", text: err instanceof ApiError ? err.message : "Error al reemplazar el archivo." });
    }
  };

  // ── CRUD de documentos de la ficha ───────────────────────────────────────
  const docByType = (t: InscriptionDocumentType) =>
    detail()?.documents?.find((d) => d.document_type === t);

  const addDoc = async (docType: InscriptionDocumentType, file: File) => {
    setFichaMsg(null);
    try {
      const fd = new FormData();
      fd.set("document_type", docType);
      fd.set("file", file);
      await apiPost(`/admin/inscripciones/${params.id}/documents`, fd);
      setFichaMsg({ type: "ok", text: "Documento guardado." });
      refetch();
    } catch (err) {
      setFichaMsg({ type: "err", text: err instanceof ApiError ? err.message : "Error al guardar el documento." });
    }
  };

  const deleteDoc = async (doc: InscriptionDocument) => {
    if (!confirm(`¿Eliminar la foto de "${doc.original_filename}" de esta solicitud?`)) return;
    setFichaMsg(null);
    try {
      await apiDelete(`/admin/inscripciones/${params.id}/documents/${doc.id}`);
      setFichaMsg({ type: "ok", text: "Documento eliminado." });
      refetch();
    } catch (err) {
      setFichaMsg({ type: "err", text: err instanceof ApiError ? err.message : "Error al eliminar el documento." });
    }
  };

  // ── Notas administrativas ────────────────────────────────────────────────
  const [notesDraft, setNotesDraft] = createSignal("");
  const [savingNotes, setSavingNotes] = createSignal(false);
  const [notesFeedback, setNotesFeedback] = createSignal<{ type: "ok" | "err"; text: string } | null>(null);
  let loadedID: string | null = null;
  createEffect(() => {
    const d = detail();
    if (d && d.id !== loadedID) {
      loadedID = d.id;
      setNotesDraft(d.notes || "");
    }
  });

  const saveNotes = async () => {
    setSavingNotes(true);
    setNotesFeedback(null);
    try {
      await apiPatch(`/admin/inscripciones/${params.id}/notes`, { notes: notesDraft() });
      setNotesFeedback({ type: "ok", text: "Notas guardadas" });
    } catch (err) {
      setNotesFeedback({ type: "err", text: err instanceof ApiError ? err.message : "Error al guardar las notas" });
    } finally { setSavingNotes(false); }
  };

  // ── Enviar correo al solicitante ─────────────────────────────────────────
  const [emailSubject, setEmailSubject] = createSignal("");
  const [emailMessage, setEmailMessage] = createSignal("");
  const [sendingEmail, setSendingEmail] = createSignal(false);
  const [emailFeedback, setEmailFeedback] = createSignal<{ type: "ok" | "err"; text: string } | null>(null);

  const sendEmail = async () => {
    setSendingEmail(true);
    setEmailFeedback(null);
    try {
      const res = await apiPost<{ email_sent: boolean }>(`/admin/inscripciones/${params.id}/email`, {
        subject: emailSubject(),
        message: emailMessage(),
      });
      setEmailFeedback({ type: "ok", text: `Correo ${res.email_sent ? "enviado" : "encolado"} a ${detail()?.correo || "la solicitante"}` });
      setEmailSubject("");
      setEmailMessage("");
    } catch (err) {
      setEmailFeedback({ type: "err", text: err instanceof ApiError ? err.message : "Error al enviar el correo" });
    } finally { setSendingEmail(false); }
  };

  const status = () => detail()?.status;

  const doApprove = async () => {
    setBusy(true);
    setFeedback(null);
    try {
      const res = await apiPost<{ control_number: string; email_sent: boolean }>(`/admin/inscripciones/${params.id}/approve`, {});
      setFeedback({ type: "ok", text: `Aprobada · N° de control ${res.control_number}${res.email_sent ? "" : " (email no enviado)"}` });
      setConfirmApprove(false);
    } catch (err) {
      setFeedback({ type: "err", text: err instanceof ApiError ? err.message : "Error al aprobar" });
    } finally { setBusy(false); }
  };

  const doReject = async () => {
    setBusy(true);
    setFeedback(null);
    try {
      await apiDelete(`/admin/inscripciones/${params.id}`);
      navigate("/admin/inscripciones");
    } catch (err) {
      setFeedback({ type: "err", text: err instanceof ApiError ? err.message : "Error al rechazar" });
      setBusy(false);
    }
  };

  const ModalityChip = (props: { label: string; on: boolean; onChange: (v: boolean) => void }) => (
    <label class="inline-flex items-center gap-2 text-sm font-medium text-colpsi-text cursor-pointer">
      <input type="checkbox" checked={props.on} onChange={(e) => props.onChange(e.currentTarget.checked)} class="accent-colpsi-blue" />
      {props.label}
    </label>
  );

  const DocSlot = (props: { spec: { type: InscriptionDocumentType; label: string; required: boolean } }) => {
    const doc = docByType(props.spec.type);
    return (
      <div class="bg-colpsi-bg rounded-lg border border-colpsi-border p-4 space-y-3">
        <div class="flex items-center justify-between gap-2">
          <span class="text-sm font-semibold text-colpsi-text">
            {props.spec.label} {props.spec.required && <span class="text-colpsi-red">*</span>}
          </span>
          <Show when={doc && doc.original_filename}>
            <span class="text-[10px] text-colpsi-muted font-medium max-w-[40%] truncate">{doc!.original_filename}</span>
          </Show>
        </div>

        <Show
          when={doc}
          fallback={
            <div class="border border-dashed border-colpsi-border rounded-md p-5 flex items-center justify-center">
              <span class="text-[11px] font-medium text-colpsi-muted uppercase">Sin documento</span>
            </div>
          }
        >
          <Show
            when={isImageUrl(doc!.url)}
            fallback={
              <a href={doc!.url} target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-2 h-8 px-3 rounded-md bg-white text-xs font-semibold text-colpsi-blue border border-colpsi-border hover:bg-colpsi-bg transition-colors">
                Ver documento ↗
              </a>
            }
          >
            <button
              onClick={() => setModalImage({ src: doc!.url, alt: props.spec.label })}
              class="block group relative w-full h-40 overflow-hidden rounded-xl border border-gray-200 cursor-pointer hover:border-colpsi-blue transition-all"
              title="Ampliar documento"
            >
              <img src={doc!.url} alt={props.spec.label} class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
              <span class="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-colors flex items-center justify-center opacity-0 group-hover:opacity-100">
                <span class="bg-white text-colpsi-blue p-2 rounded-full shadow-lg"><Icon name="search" class="w-4 h-4" /></span>
              </span>
            </button>
          </Show>
        </Show>

        <div class="flex flex-wrap gap-2">
          <label class="inline-flex items-center gap-2 h-9 px-3.5 rounded-md bg-colpsi-blue text-white text-xs font-semibold cursor-pointer hover:bg-colpsi-blue-light transition-colors">
            <input type="file" accept="image/*,application/pdf" class="sr-only" onChange={(e) => {
              const f = e.currentTarget.files?.[0];
              if (f) addDoc(props.spec.type, f);
              e.currentTarget.value = "";
            }} />
            {doc ? "Reemplazar" : "Adjuntar"}
          </label>
          <Show when={doc}>
            <button
              onClick={() => deleteDoc(doc!)}
              class="h-9 px-3.5 rounded-md border border-red-200 bg-white text-colpsi-red text-xs font-semibold hover:bg-red-50 transition-colors"
            >
              Eliminar
            </button>
          </Show>
        </div>
      </div>
    );
  };

  return (
    <main class="space-y-4 pb-12">
      <button onClick={() => navigate("/admin/inscripciones")} class="inline-flex items-center gap-1.5 text-xs font-semibold text-colpsi-muted uppercase tracking-wide hover:text-colpsi-blue transition-colors">
        <Icon name="chevronRight" class="w-3.5 h-3.5 rotate-180" />
        Volver a solicitudes
      </button>

      <Show when={detail()} fallback={<div class="p-20 text-center"><div class="w-10 h-10 border-4 border-colpsi-blue border-t-transparent rounded-full animate-spin mx-auto" /></div>}>
        {(d) => (
          <>
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-4 border-b border-colpsi-border">
              <div>
                <h1 class="text-lg font-semibold text-colpsi-text">{d().nombres} {d().apellidos}</h1>
                <p class="text-sm text-colpsi-muted font-medium mt-0.5">
                  C.I. {d().cedula} {d().nacionalidad} · Solicitada el {new Date(d().created_at).toLocaleDateString()}
                </p>
              </div>
              <StatusPill status={d().status} />
            </div>

            <Show when={d().control_number}>
              <div class="bg-blue-50 border border-blue-200 rounded-md p-4 flex flex-wrap items-center gap-x-6 gap-y-1">
                <span class="text-sm font-semibold text-colpsi-blue">
                  N° de control asignado: {d().control_number}
                </span>
                <span class="text-xs font-medium text-colpsi-blue/70">
                  Solvencias pagadas: {d().solvency_count}
                </span>
              </div>
            </Show>

            <Show when={feedback()}>
              <div class={`rounded-md p-3 text-sm font-medium ${feedback()!.type === "ok" ? "bg-emerald-50 text-emerald-700 border border-emerald-200" : "bg-red-50 text-colpsi-red border border-red-200"}`}>
                {feedback()!.text}
              </div>
            </Show>

            {/* ── Ficha de inscripción (editable) ───────────────────────────── */}
            <Notebook
              pages={[
                { id: "personales", label: "Datos Personales" },
                { id: "academicos", label: "Datos Académicos" },
                { id: "ubicacion", label: "Ubicación y Modalidad" },
                { id: "documentos", label: "Fotografía y Documentos" },
              ]}
            >
              <NotebookPage id="personales">
                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
                  <Field label="Cédula"><input type="number" value={form.cedula} onInput={(e) => set("cedula", e.currentTarget.value)} class={IC} /></Field>
                  <Field label="Nacionalidad">
                    <select value={form.nacionalidad} onChange={(e) => set("nacionalidad", e.currentTarget.value)} class={IC}>
                      <option value="V">V - Venezolano</option>
                      <option value="E">E - Extranjero</option>
                    </select>
                  </Field>
                  <Field label="N° FPV"><input type="number" value={form.fpv} onInput={(e) => set("fpv", e.currentTarget.value)} class={IC} /></Field>
                  <Field label="Nombres"><input type="text" value={form.nombres} onInput={(e) => set("nombres", e.currentTarget.value)} class={IC} /></Field>
                  <Field label="Segundo nombre"><input type="text" value={form.segundo_nombre} onInput={(e) => set("segundo_nombre", e.currentTarget.value)} class={IC} /></Field>
                  <Field label="Apellidos"><input type="text" value={form.apellidos} onInput={(e) => set("apellidos", e.currentTarget.value)} class={IC} /></Field>
                  <Field label="Segundo apellido"><input type="text" value={form.segundo_apellido} onInput={(e) => set("segundo_apellido", e.currentTarget.value)} class={IC} /></Field>
                  <Field label="Género">
                    <select value={form.genero} onChange={(e) => set("genero", e.currentTarget.value)} class={IC}>
                      <option value="">Seleccionar</option>
                      <option value="M">Masculino</option>
                      <option value="F">Femenino</option>
                    </select>
                  </Field>
                  <Field label="Fecha de nacimiento"><input type="date" value={form.fecha_nacimiento ?? ""} onInput={(e) => set("fecha_nacimiento", e.currentTarget.value)} class={IC} /></Field>
                  <Field label="Teléfono"><input type="tel" value={form.telefono} onInput={(e) => set("telefono", e.currentTarget.value)} class={IC} /></Field>
                  <Field label="Correo electrónico"><input type="email" value={form.correo} onInput={(e) => set("correo", e.currentTarget.value)} class={IC} /></Field>
                  <Field label="RIF"><input type="text" value={form.rif} onInput={(e) => set("rif", e.currentTarget.value)} class={IC} /></Field>
                </div>
              </NotebookPage>

              <NotebookPage id="academicos">
                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
                  <Field label="Universidad"><input type="text" value={form.titulo_universidad} onInput={(e) => set("titulo_universidad", e.currentTarget.value)} class={IC} /></Field>
                  <Field label="Fecha de graduación"><input type="date" value={form.titulo_fecha_graduacion ?? ""} onInput={(e) => set("titulo_fecha_graduacion", e.currentTarget.value)} class={IC} /></Field>
                  <Field label="Mención"><input type="text" value={form.titulo_mencion} onInput={(e) => set("titulo_mencion", e.currentTarget.value)} class={IC} /></Field>
                  <Field label="N° Registro del título"><input type="text" value={form.titulo_registro_numero} onInput={(e) => set("titulo_registro_numero", e.currentTarget.value)} class={IC} /></Field>
                  <Field label="Estado del registro"><input type="text" value={form.titulo_registro_estado} onInput={(e) => set("titulo_registro_estado", e.currentTarget.value)} class={IC} /></Field>
                  <Field label="Tomo del registro"><input type="text" value={form.titulo_registro_tomo} onInput={(e) => set("titulo_registro_tomo", e.currentTarget.value)} class={IC} /></Field>
                  <Field label="Folio del registro"><input type="text" value={form.titulo_registro_folio} onInput={(e) => set("titulo_registro_folio", e.currentTarget.value)} class={IC} /></Field>
                </div>
              </NotebookPage>

              <NotebookPage id="ubicacion">
                <div class="space-y-8">
                  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
                    <Field label="Municipio (Carabobo)">
                      <select value={form.municipality_carabobo} onChange={(e) => set("municipality_carabobo", e.currentTarget.value)} class={IC}>
                        <option value="">Seleccionar municipio…</option>
                        <For each={MUNICIPIOS_CARABOBO}>{(m) => <option value={m}>{m}</option>}</For>
                        {form.municipality_carabobo && !MUNICIPIOS_CARABOBO.includes(form.municipality_carabobo) && (
                          <option value={form.municipality_carabobo}>{form.municipality_carabobo} (no estándar)</option>
                        )}
                      </select>
                    </Field>
                    <Field label="Dirección del consultorio"><input type="text" value={form.service_address} onInput={(e) => set("service_address", e.currentTarget.value)} class={IC} placeholder="Av. Principal, edificio…" /></Field>
                    <Field label="Otro estado (fuera de Carabobo)">
                      <select
                        value={form.state_outside}
                        onChange={(e) => {
                          const v = e.currentTarget.value;
                          set("state_outside", v);
                          const cur = form.municipality_outside_carabobo;
                          if (cur && !municipiosDe(v).includes(cur)) set("municipality_outside_carabobo", "");
                        }}
                        class={IC}
                      >
                        <option value="">Seleccionar estado…</option>
                        <For each={ESTADOS_VENEZUELA}>{(s) => <option value={s}>{s}</option>}</For>
                        {form.state_outside && !ESTADOS_VENEZUELA.includes(form.state_outside) && (
                          <option value={form.state_outside}>{form.state_outside} (no estándar)</option>
                        )}
                      </select>
                    </Field>
                    <Field label="Municipio / ciudad (fuera de Carabobo)">
                      <select
                        value={form.municipality_outside_carabobo}
                        onChange={(e) => set("municipality_outside_carabobo", e.currentTarget.value)}
                        disabled={!form.state_outside}
                        class={IC}
                      >
                        <option value="">
                          {!form.state_outside
                            ? "Primero selecciona un estado"
                            : municipiosDe(form.state_outside).length === 0
                              ? "Este estado no tiene municipios"
                              : "Seleccionar municipio…"}
                        </option>
                        <For each={municipiosDe(form.state_outside)}>{(m) => <option value={m}>{m}</option>}</For>
                        {form.municipality_outside_carabobo &&
                          !municipiosDe(form.state_outside).includes(form.municipality_outside_carabobo) && (
                            <option value={form.municipality_outside_carabobo}>
                              {form.municipality_outside_carabobo} (no estándar)
                            </option>
                          )}
                      </select>
                    </Field>
                    <Field label="País (fuera de Venezuela)"><input type="text" value={form.country} onInput={(e) => set("country", e.currentTarget.value)} class={IC} placeholder="Ej: España" /></Field>
                    <Field label="Área de trabajo principal">
                      <select value={form.primary_specialty_id ?? ""} onChange={(e) => set("primary_specialty_id", e.currentTarget.value)} class={IC}>
                        <option value="">— Sin área —</option>
                        <For each={workAreas() ?? []}>{(wa) => <option value={wa.id}>{wa.name}</option>}</For>
                      </select>
                    </Field>
                    <Field label="Área de trabajo secundaria">
                      <select value={form.secondary_specialty_id ?? ""} onChange={(e) => set("secondary_specialty_id", e.currentTarget.value)} class={IC}>
                        <option value="">— Sin área —</option>
                        <For each={workAreas() ?? []}>
                          {(wa) => (
                            <option value={wa.id} disabled={String(wa.id) === String(form.primary_specialty_id)}>{wa.name}</option>
                          )}
                        </For>
                      </select>
                    </Field>
                  </div>

                  <div class="bg-colpsi-bg p-5 rounded-lg border border-colpsi-border">
                    <p class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide mb-3 ml-1">Modalidad de servicio</p>
                    <div class="flex flex-wrap gap-6">
                      <ModalityChip label="Presencial" on={form.service_modality_presencial} onChange={(v) => set("service_modality_presencial", v)} />
                      <ModalityChip label="A distancia" on={form.service_modality_distance} onChange={(v) => set("service_modality_distance", v)} />
                      <ModalityChip label="Telefónica" on={form.service_modality_telephone} onChange={(v) => set("service_modality_telephone", v)} />
                    </div>
                  </div>
                </div>
              </NotebookPage>

              <NotebookPage id="documentos">
                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div class="bg-colpsi-bg rounded-lg border border-colpsi-border p-4 space-y-3">
                    <span class="text-sm font-semibold text-colpsi-text">Foto tipo carnet</span>
                    <Show when={d().foto_url} fallback={<p class="text-sm text-colpsi-muted">Sin foto</p>}>
                      <button
                        onClick={() => setModalImage({ src: bucketUrl(d().foto_url), alt: "Foto tipo carnet del solicitante" })}
                        class="block group relative w-full h-44 overflow-hidden rounded-xl border border-gray-200 cursor-pointer hover:border-colpsi-blue transition-all"
                        title="Ampliar foto"
                      >
                        <img src={bucketUrl(d().foto_url)} alt="Foto del solicitante" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
                        <span class="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-colors flex items-center justify-center opacity-0 group-hover:opacity-100">
                          <span class="bg-white text-colpsi-blue p-2 rounded-full shadow-lg"><Icon name="search" class="w-4 h-4" /></span>
                        </span>
                      </button>
                    </Show>
                    <label class="inline-flex items-center gap-2 h-9 px-3.5 rounded-md bg-colpsi-blue text-white text-xs font-semibold cursor-pointer hover:bg-colpsi-blue-light transition-colors">
                      <input type="file" accept="image/*" class="sr-only" onChange={(e) => {
                        const f = e.currentTarget.files?.[0];
                        if (f) replacePhoto("foto", f);
                        e.currentTarget.value = "";
                      }} />
                      Reemplazar foto
                    </label>
                  </div>

                  <div class="bg-colpsi-bg rounded-lg border border-colpsi-border p-4 space-y-3">
                    <span class="text-sm font-semibold text-colpsi-text">Comprobante de pago</span>
                    <Show when={d().comprobante_url} fallback={<p class="text-sm text-colpsi-muted">Sin comprobante</p>}>
                      <Show
                        when={isImageUrl(d().comprobante_url)}
                        fallback={
                          <a href={d().comprobante_url} target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-2 h-8 px-3 rounded-md bg-white text-xs font-semibold text-colpsi-blue border border-colpsi-border hover:bg-colpsi-bg transition-colors">
                            Ver comprobante ↗
                          </a>
                        }
                      >
                        <button
                          onClick={() => setModalImage({ src: bucketUrl(d().comprobante_url), alt: "Comprobante de pago" })}
                          class="block group relative w-full h-44 overflow-hidden rounded-xl border border-gray-200 cursor-pointer hover:border-colpsi-blue transition-all"
                          title="Ampliar comprobante"
                        >
                          <img src={bucketUrl(d().comprobante_url)} alt="Comprobante de pago" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
                          <span class="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-colors flex items-center justify-center opacity-0 group-hover:opacity-100">
                            <span class="bg-white text-colpsi-blue p-2 rounded-full shadow-lg"><Icon name="search" class="w-4 h-4" /></span>
                          </span>
                        </button>
                      </Show>
                    </Show>
                    <label class="inline-flex items-center gap-2 h-9 px-3.5 rounded-md bg-colpsi-blue text-white text-xs font-semibold cursor-pointer hover:bg-colpsi-blue-light transition-colors">
                      <input type="file" accept="image/*,application/pdf" class="sr-only" onChange={(e) => {
                        const f = e.currentTarget.files?.[0];
                        if (f) replacePhoto("comprobante", f);
                        e.currentTarget.value = "";
                      }} />
                      Reemplazar comprobante
                    </label>
                  </div>

                  <For each={DOC_SPECS}>{(spec) => <DocSlot spec={spec} />}</For>
                </div>
              </NotebookPage>
            </Notebook>

            <Show when={fichaMsg()}>
              <div class={`rounded-md p-3 text-sm font-medium ${fichaMsg()!.type === "ok" ? "bg-emerald-50 text-emerald-700 border border-emerald-200" : "bg-red-50 text-colpsi-red border border-red-200"}`}>
                {fichaMsg()!.text}
              </div>
            </Show>

            <div class="flex justify-end">
              <button
                onClick={saveFicha}
                disabled={savingFicha()}
                class="inline-flex items-center gap-2 h-11 px-6 rounded-md bg-colpsi-blue hover:bg-colpsi-blue-light text-white font-semibold transition-all disabled:opacity-50"
              >
                <Show when={savingFicha()} fallback={<><Icon name="check" /><span>Guardar ficha</span></>}>
                  <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  <span>Guardando...</span>
                </Show>
              </button>
            </div>

            {/* Notas administrativas */}
            <div class="bg-white rounded-lg border border-colpsi-border p-5 space-y-3">
              <h2 class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide">Notas administrativas</h2>
              <textarea
                value={notesDraft()}
                onInput={(e) => setNotesDraft(e.currentTarget.value)}
                rows={4}
                placeholder="Escribe notas internas sobre esta solicitud..."
                class="w-full rounded-md border border-slate-300 bg-white px-3 py-2.5 text-sm text-colpsi-text outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors resize-y"
              />
              <Show when={notesFeedback()}>
                <div class={`rounded-md p-3 text-sm font-medium ${notesFeedback()!.type === "ok" ? "bg-emerald-50 text-emerald-700 border border-emerald-200" : "bg-red-50 text-colpsi-red border border-red-200"}`}>
                  {notesFeedback()!.text}
                </div>
              </Show>
              <div class="flex justify-end">
                <button
                  onClick={saveNotes}
                  disabled={savingNotes()}
                  class="h-9 px-4 rounded-md bg-colpsi-blue text-white font-semibold text-sm hover:bg-colpsi-blue-light transition-colors disabled:opacity-50"
                >
                  {savingNotes() ? "Guardando..." : "Guardar notas"}
                </button>
              </div>
            </div>

            {/* Enviar correo al solicitante */}
            <div class="bg-white rounded-lg border border-colpsi-border p-5 space-y-3">
              <h2 class="text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide">Enviar correo al solicitante</h2>
              <p class="text-sm text-colpsi-muted">
                Para: <span class="font-semibold text-colpsi-text">{d().correo}</span>
              </p>
              <input
                value={emailSubject()}
                onInput={(e) => setEmailSubject(e.currentTarget.value)}
                placeholder="Asunto"
                class="w-full h-9 rounded-md border border-slate-300 bg-white px-3 text-sm text-colpsi-text outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors"
              />
              <textarea
                value={emailMessage()}
                onInput={(e) => setEmailMessage(e.currentTarget.value)}
                rows={4}
                placeholder="Mensaje para el solicitante..."
                class="w-full rounded-md border border-slate-300 bg-white px-3 py-2.5 text-sm text-colpsi-text outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors resize-y"
              />
              <Show when={emailFeedback()}>
                <div class={`rounded-md p-3 text-sm font-medium ${emailFeedback()!.type === "ok" ? "bg-emerald-50 text-emerald-700 border border-emerald-200" : "bg-red-50 text-colpsi-red border border-red-200"}`}>
                  {emailFeedback()!.text}
                </div>
              </Show>
              <div class="flex justify-end">
                <button
                  onClick={sendEmail}
                  disabled={sendingEmail() || !emailSubject().trim() || !emailMessage().trim()}
                  class="h-9 px-4 rounded-md bg-colpsi-blue text-white font-semibold text-sm hover:bg-colpsi-blue-light transition-colors disabled:opacity-50"
                >
                  {sendingEmail() ? "Enviando..." : "Enviar correo"}
                </button>
              </div>
            </div>

            <Show when={status() === "pending"}>
              <div class="flex flex-col sm:flex-row gap-3 pt-2">
                <button
                  onClick={() => setConfirmApprove(true)}
                  disabled={busy()}
                  class="flex-1 h-11 rounded-md bg-emerald-600 text-white px-6 font-semibold hover:bg-emerald-700 transition-colors disabled:opacity-50"
                >
                  Aprobar inscripción
                </button>
                <button
                  onClick={() => setConfirmReject(true)}
                  disabled={busy()}
                  class="flex-1 h-11 rounded-md bg-colpsi-red text-white px-6 font-semibold hover:opacity-90 transition-colors disabled:opacity-50"
                >
                  Rechazar solicitud
                </button>
              </div>
            </Show>
          </>
        )}
      </Show>

      {/* Modal aprobar */}
      <Show when={confirmApprove()}>
        <Modal title="Aprobar inscripción" onClose={() => setConfirmApprove(false)}>
          <p class="text-sm text-gray-600 leading-relaxed">
            Se creará la cuenta del psicólogo <strong>activa, solvente y con fe de vida</strong> (la foto
            tipo carnet pasará a ser su foto de perfil). Los documentos digitales del formulario, incluido
            el comprobante de pago, se migrarán a su expediente. Se le asignará un número de control
            secuencial y se enviará un correo con las credenciales. ¿Confirmar?
          </p>
          <ModalActions onCancel={() => setConfirmApprove(false)} onConfirm={doApprove} busy={busy()} confirmLabel="Aprobar" />
        </Modal>
      </Show>

      {/* Modal rechazar */}
      <Show when={confirmReject()}>
        <Modal title="Rechazar solicitud" onClose={() => setConfirmReject(false)}>
          <p class="text-sm text-gray-600 leading-relaxed">
            Se eliminará permanentemente esta solicitud y los archivos adjuntos. ¿Está seguro?
          </p>
          <ModalActions onCancel={() => setConfirmReject(false)} onConfirm={doReject} busy={busy()} confirmLabel="Rechazar y eliminar" danger />
        </Modal>
      </Show>

      {/* Modal de imagen (expansión de foto / comprobante / documentos) */}
      <ImageModal
        src={modalImage()?.src || ""}
        alt={modalImage()?.alt || ""}
        isOpen={!!modalImage()}
        onClose={closeModal}
      />
    </main>
  );
}

function StatusPill(props: { status: string }) {
  const map: Record<string, string> = {
    pending: "bg-amber-50 text-amber-700 border-amber-200",
    approved: "bg-green-50 text-green-700 border-green-200",
    rejected: "bg-red-50 text-red-700 border-red-200",
  };
  const label: Record<string, string> = { pending: "Pendiente", approved: "Aprobada", rejected: "Rechazada" };
  return <span class={`inline-flex items-center px-2.5 py-0.5 rounded-md text-xs font-semibold border ${map[props.status] || ""}`}>{label[props.status] || props.status}</span>;
}

function Modal(props: { title: string; onClose: () => void; children: any }) {
  return (
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm" onClick={props.onClose}>
      <div class="bg-white rounded-lg shadow-lg border border-colpsi-border max-w-md w-full p-5" onClick={(e) => e.stopPropagation()}>
        <h3 class="text-base font-semibold text-colpsi-text mb-4">{props.title}</h3>
        {props.children}
      </div>
    </div>
  );
}

function ModalActions(props: { onCancel: () => void; onConfirm: () => void; busy: boolean; confirmLabel: string; danger?: boolean }) {
  return (
    <div class="flex gap-2 mt-5">
      <button onClick={props.onCancel} disabled={props.busy} class="flex-1 h-9 rounded-md border border-colpsi-border bg-white font-medium text-colpsi-text hover:bg-colpsi-bg transition-colors disabled:opacity-50 text-sm">Cancelar</button>
      <button
        onClick={props.onConfirm}
        disabled={props.busy}
        class={`flex-1 h-9 rounded-md font-semibold text-white transition-colors disabled:opacity-50 text-sm ${props.danger ? "bg-colpsi-red hover:opacity-90" : "bg-emerald-600 hover:bg-emerald-700"}`}
      >
        {props.busy ? "Procesando..." : props.confirmLabel}
      </button>
    </div>
  );
}