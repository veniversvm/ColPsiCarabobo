// web/src/routes/admin/inscripciones/[id].tsx
import { createResource, createSignal, createEffect, Show } from "solid-js";
import { useParams, useNavigate } from "@solidjs/router";
import { apiGet, apiPost, apiPatch, apiDelete, ApiError } from "~/lib/api";
import { bucketUrl } from "~/lib/bucket";
import { ImageModal } from "~/components/ui/ImageModal";
import type { InscriptionDetail, ApproveInscriptionResponse, SendEmailToApplicantResponse } from "~/types/inscription";

export default function AdminInscriptionDetail() {
  const params = useParams();
  const navigate = useNavigate();
  const [detail] = createResource<InscriptionDetail>(() =>
    apiGet(`/admin/inscripciones/${params.id}`)
  );

  const [confirmApprove, setConfirmApprove] = createSignal(false);
  const [confirmReject, setConfirmReject] = createSignal(false);
  const [busy, setBusy] = createSignal(false);
  const [feedback, setFeedback] = createSignal<{ type: "ok" | "err"; text: string } | null>(null);
  const [modalImage, setModalImage] = createSignal<{ src: string; alt: string } | null>(null);

  const closeModal = () => setModalImage(null);

  // El comprobante puede ser imagen o PDF: solo las imágenes se expanden en el modal.
  const isImageUrl = (url?: string) => !!url && /\.(png|jpe?g|gif|webp|bmp|svg)(\?|#|$)/i.test(url);

  // ── Notas administrativas ──────────────────────────────────────────────
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

  // ── Enviar correo al solicitante ───────────────────────────────────────
  const [emailSubject, setEmailSubject] = createSignal("");
  const [emailMessage, setEmailMessage] = createSignal("");
  const [sendingEmail, setSendingEmail] = createSignal(false);
  const [emailFeedback, setEmailFeedback] = createSignal<{ type: "ok" | "err"; text: string } | null>(null);

  const sendEmail = async () => {
    setSendingEmail(true);
    setEmailFeedback(null);
    try {
      const res = await apiPost<SendEmailToApplicantResponse>(`/admin/inscripciones/${params.id}/email`, {
        subject: emailSubject(),
        message: emailMessage(),
      });
      setEmailFeedback({ type: "ok", text: `Correo ${res.email_sent ? "enviado" : "encolado"} a ${d()?.correo || "la solicitante"}` });
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
      const res = await apiPost<ApproveInscriptionResponse>(`/admin/inscripciones/${params.id}/approve`, {});
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

  const row = (label: string, value?: string | number | null) => (
    <div class="flex justify-between gap-4 py-2.5 border-b border-gray-50 last:border-0">
      <span class="text-xs font-bold text-gray-400 uppercase tracking-widest">{label}</span>
      <span class="text-sm font-semibold text-gray-700 text-right">{value ?? "—"}</span>
    </div>
  );

  return (
    <div class="space-y-6 font-sans">
      <button onClick={() => navigate("/admin/inscripciones")} class="text-sm font-bold text-gray-400 hover:text-colpsi-blue transition-colors">
        ← Volver a solicitudes
      </button>

      <Show when={detail()} fallback={<div class="p-20 text-center"><div class="w-10 h-10 border-4 border-colpsi-blue border-t-transparent rounded-full animate-spin mx-auto" /></div>}>
        {(d) => (
          <>
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div>
                <h1 class="text-2xl font-black text-colpsi-blue">{d().nombres} {d().apellidos}</h1>
                <p class="text-sm text-gray-500 font-medium">
                  C.I. {d().cedula} {d().nacionalidad} · Solicitada el {new Date(d().created_at).toLocaleDateString()}
                </p>
              </div>
              <StatusPill status={d().status} />
            </div>

            <Show when={d().control_number}>
              <div class="bg-blue-50 border border-blue-100 rounded-2xl p-4 flex flex-wrap items-center gap-x-6 gap-y-1">
                <span class="text-sm font-black text-colpsi-blue">
                  N° de control asignado: {d().control_number}
                </span>
                <span class="text-xs font-bold text-blue-600/70">
                  Solvencias pagadas: {d().solvency_count}
                </span>
              </div>
            </Show>

            <Show when={feedback()}>
              <div class={`rounded-2xl p-4 text-sm font-bold ${feedback()!.type === "ok" ? "bg-green-50 text-green-700 border border-green-200" : "bg-red-50 text-red-700 border border-red-200"}`}>
                {feedback()!.text}
              </div>
            </Show>

            <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Datos */}
              <div class="bg-white rounded-3xl shadow-sm border border-gray-100 p-6 space-y-6">
                <h2 class="text-sm font-black text-gray-400 uppercase tracking-widest">Datos Personales</h2>
                <div>
                  {row("Cédula", `${d().cedula} ${d().nacionalidad}`)}
                  {row("Nombres", [d().nombres, d().segundo_nombre].filter(Boolean).join(" ") || null)}
                  {row("Apellidos", [d().apellidos, d().segundo_apellido].filter(Boolean).join(" ") || null)}
                  {row("Género", d().genero === "M" ? "Masculino" : d().genero === "F" ? "Femenino" : null)}
                  {row("N° FPV", d().fpv || null)}
                  {row("Teléfono", d().telefono)}
                  {row("Correo", d().correo)}
                  {row("Fecha de nacimiento", d().fecha_nacimiento ? new Date(d().fecha_nacimiento).toLocaleDateString() : null)}
                  {row("RIF", d().rif)}
                </div>

                <h2 class="text-sm font-black text-gray-400 uppercase tracking-widest pt-2">Datos Académicos</h2>
                <div>
                  {row("Universidad", d().titulo_universidad)}
                  {row("Fecha de graduación", d().titulo_fecha_graduacion ? new Date(d().titulo_fecha_graduacion).toLocaleDateString() : null)}
                  {row("Mención", d().titulo_mencion)}
                  {row("N° Registro título", d().titulo_registro_numero)}
                  {row("Tomo del registro", d().titulo_registro_tomo)}
                  {row("Folio del registro", d().titulo_registro_folio)}
                  {row("Estado del registro", d().titulo_registro_estado)}
                </div>
              </div>

              {/* Archivos */}
              <div class="bg-white rounded-3xl shadow-sm border border-gray-100 p-6 space-y-8">
                <div>
                  <h2 class="text-sm font-black text-gray-400 uppercase tracking-widest mb-3">Foto tipo carnet</h2>
                  <Show when={d().foto_url} fallback={<p class="text-sm text-gray-400">Sin foto</p>}>
                    <button
                      onClick={() => setModalImage({ src: bucketUrl(d().foto_url), alt: "Foto tipo carnet del solicitante" })}
                      class="block group relative w-48 h-48 overflow-hidden rounded-2xl border border-gray-100 cursor-pointer hover:border-colpsi-blue transition-all"
                      title="Ampliar foto"
                    >
                      <img
                        src={bucketUrl(d().foto_url)}
                        alt="Foto del solicitante"
                        class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                      />
                      <span class="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-colors flex items-center justify-center opacity-0 group-hover:opacity-100">
                        <span class="bg-white text-colpsi-blue p-2 rounded-full shadow-lg">🔍</span>
                      </span>
                    </button>
                  </Show>
                </div>
                <div>
                  <h2 class="text-sm font-black text-gray-400 uppercase tracking-widest mb-3">Comprobante de pago</h2>
                  <Show when={d().comprobante_url} fallback={<p class="text-sm text-gray-400">Sin comprobante</p>}>
                    <Show
                      when={isImageUrl(d().comprobante_url)}
                      fallback={
                        <a href={d().comprobante_url} target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-2 px-4 py-2.5 bg-gray-100 rounded-xl text-sm font-black text-colpsi-blue hover:bg-gray-200 transition-colors">
                          Ver comprobante ↗
                        </a>
                      }
                    >
                      <button
                        onClick={() => setModalImage({ src: bucketUrl(d().comprobante_url), alt: "Comprobante de pago" })}
                        class="block group relative w-full h-56 overflow-hidden rounded-2xl border border-gray-100 cursor-pointer hover:border-colpsi-blue transition-all"
                        title="Ampliar comprobante"
                      >
                        <img
                          src={bucketUrl(d().comprobante_url)}
                          alt="Comprobante de pago"
                          class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                        />
                        <span class="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-colors flex items-center justify-center opacity-0 group-hover:opacity-100">
                          <span class="bg-white text-colpsi-blue p-2 rounded-full shadow-lg">🔍</span>
                        </span>
                      </button>
                    </Show>
                  </Show>
                </div>
              </div>
            </div>

            {/* Notas administrativas */}
            <div class="bg-white rounded-3xl shadow-sm border border-gray-100 p-6 space-y-4">
              <h2 class="text-sm font-black text-gray-400 uppercase tracking-widest">Notas administrativas</h2>
              <textarea
                value={notesDraft()}
                onInput={(e) => setNotesDraft(e.currentTarget.value)}
                rows={4}
                placeholder="Escribe notas internas sobre esta solicitud..."
                class="w-full rounded-2xl border border-gray-200 p-3 text-sm text-gray-700 focus:outline-none focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/20 resize-y"
              />
              <Show when={notesFeedback()}>
                <div class={`rounded-xl p-3 text-sm font-bold ${notesFeedback()!.type === "ok" ? "bg-green-50 text-green-700 border border-green-200" : "bg-red-50 text-red-700 border border-red-200"}`}>
                  {notesFeedback()!.text}
                </div>
              </Show>
              <div class="flex justify-end">
                <button
                  onClick={saveNotes}
                  disabled={savingNotes()}
                  class="px-5 py-2.5 rounded-xl bg-colpsi-blue text-white font-black text-sm hover:bg-colpsi-blue/90 transition-colors disabled:opacity-50"
                >
                  {savingNotes() ? "Guardando..." : "Guardar notas"}
                </button>
              </div>
            </div>

            {/* Enviar correo al solicitante */}
            <div class="bg-white rounded-3xl shadow-sm border border-gray-100 p-6 space-y-4">
              <h2 class="text-sm font-black text-gray-400 uppercase tracking-widest">Enviar correo al solicitante</h2>
              <p class="text-sm text-gray-500">
                Para: <span class="font-bold text-gray-700">{d().correo}</span>
              </p>
              <input
                value={emailSubject()}
                onInput={(e) => setEmailSubject(e.currentTarget.value)}
                placeholder="Asunto"
                class="w-full rounded-2xl border border-gray-200 p-3 text-sm text-gray-700 focus:outline-none focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/20"
              />
              <textarea
                value={emailMessage()}
                onInput={(e) => setEmailMessage(e.currentTarget.value)}
                rows={4}
                placeholder="Mensaje para el solicitante..."
                class="w-full rounded-2xl border border-gray-200 p-3 text-sm text-gray-700 focus:outline-none focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/20 resize-y"
              />
              <Show when={emailFeedback()}>
                <div class={`rounded-xl p-3 text-sm font-bold ${emailFeedback()!.type === "ok" ? "bg-green-50 text-green-700 border border-green-200" : "bg-red-50 text-red-700 border border-red-200"}`}>
                  {emailFeedback()!.text}
                </div>
              </Show>
              <div class="flex justify-end">
                <button
                  onClick={sendEmail}
                  disabled={sendingEmail() || !emailSubject().trim() || !emailMessage().trim()}
                  class="px-5 py-2.5 rounded-xl bg-indigo-600 text-white font-black text-sm hover:bg-indigo-700 transition-colors disabled:opacity-50"
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
                  class="flex-1 bg-green-600 text-white py-3.5 px-6 rounded-2xl font-black hover:bg-green-700 transition-colors disabled:opacity-50"
                >
                  Aprobar inscripción
                </button>
                <button
                  onClick={() => setConfirmReject(true)}
                  disabled={busy()}
                  class="flex-1 bg-red-500 text-white py-3.5 px-6 rounded-2xl font-black hover:bg-red-600 transition-colors disabled:opacity-50"
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
            Se creará la cuenta del psicólogo <strong>activa y solvente</strong> (la foto
            tipo carnet pasará a ser su foto de perfil). Se le asignará un número de
            control secuencial y se enviará un correo con las credenciales. ¿Confirmar?
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

      {/* Modal de imagen (expansión de foto / comprobante) */}
      <ImageModal
        src={modalImage()?.src || ""}
        alt={modalImage()?.alt || ""}
        isOpen={!!modalImage()}
        onClose={closeModal}
      />
    </div>
  );
}

function StatusPill(props: { status: string }) {
  const map: Record<string, string> = {
    pending: "bg-amber-50 text-amber-700 border-amber-200",
    approved: "bg-green-50 text-green-700 border-green-200",
    rejected: "bg-red-50 text-red-700 border-red-200",
  };
  const label: Record<string, string> = { pending: "Pendiente", approved: "Aprobada", rejected: "Rechazada" };
  return <span class={`px-3 py-1 rounded-full text-xs font-black border ${map[props.status] || ""}`}>{label[props.status] || props.status}</span>;
}

function Modal(props: { title: string; onClose: () => void; children: any }) {
  return (
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm" onClick={props.onClose}>
      <div class="bg-white rounded-3xl shadow-2xl max-w-md w-full p-6" onClick={(e) => e.stopPropagation()}>
        <h3 class="text-lg font-black text-colpsi-blue mb-4">{props.title}</h3>
        {props.children}
      </div>
    </div>
  );
}

function ModalActions(props: { onCancel: () => void; onConfirm: () => void; busy: boolean; confirmLabel: string; danger?: boolean }) {
  return (
    <div class="flex gap-3 mt-6">
      <button onClick={props.onCancel} disabled={props.busy} class="flex-1 py-2.5 rounded-xl font-bold text-gray-500 bg-gray-100 hover:bg-gray-200 transition-colors disabled:opacity-50">Cancelar</button>
      <button
        onClick={props.onConfirm}
        disabled={props.busy}
        class={`flex-1 py-2.5 rounded-xl font-black text-white transition-colors disabled:opacity-50 ${props.danger ? "bg-red-500 hover:bg-red-600" : "bg-green-600 hover:bg-green-700"}`}
      >
        {props.busy ? "Procesando..." : props.confirmLabel}
      </button>
    </div>
  );
}