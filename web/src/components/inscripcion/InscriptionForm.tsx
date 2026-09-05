// web/src/components/inscripcion/InscriptionForm.tsx
import { createSignal, createEffect, createResource, For, Show } from "solid-js";
import { isServer } from "solid-js/web";
import { apiGet, apiPost, ApiError } from "~/lib/api";
import { CheckField } from "~/components/inscripcion/CheckField";
import { FileUpload } from "~/components/inscripcion/FileUpload";
import { SuccessMessage } from "~/components/inscripcion/SuccessMessage";
import { MUNICIPIOS_CARABOBO, ESTADOS_VENEZUELA } from "~/lib/geo";
import FlatDatePicker from "~/components/ui/FlatDatePicker";
import type { WorkArea } from "~/types/inscription";

const STORAGE_KEY = "inscripcion_draft";

// Persiste cada campo de texto en sessionStorage: los datos se conservan mientras
// la pestaña/ventana del navegador siga abierta (incluso al cambiar a la pestaña
// "Instrucciones" o recargar la página). Se limpian al cerrar la ventana o enviar.
function createSessionField(key: string, initial: string): [() => string, (v: string) => void] {
  const [value, setValue] = createSignal(load(key, initial));
  createEffect(() => {
    if (!isServer) save(key, value());
  });
  return [value, setValue];
}

function load(key: string, initial: string): string {
  if (isServer) return initial;
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    if (!raw) return initial;
    const parsed = JSON.parse(raw);
    return typeof parsed[key] === "string" ? parsed[key] : initial;
  } catch {
    return initial;
  }
}

function save(key: string, value: string) {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    const parsed = raw ? JSON.parse(raw) : {};
    parsed[key] = value;
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(parsed));
  } catch {
    /* sessionStorage no disponible */
  }
}

// Guarda el estado de archivos seleccionados (solo metadatos; los bytes no son
// recuperables del storage, pero la UI conserva el aspecto de "seleccionado").
function loadFileMeta(key: string): { name: string; size: number; type: string; lastModified: number } | null {
  if (isServer) return null;
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw);
    const m = parsed[`file_${key}`];
    return m && typeof m.name === "string" ? m : null;
  } catch {
    return null;
  }
}
function saveFileMeta(key: string, f: File | null) {
  if (isServer) return;
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    const parsed = raw ? JSON.parse(raw) : {};
    if (f) {
      parsed[`file_${key}`] = {
        name: f.name,
        size: f.size,
        type: f.type,
        lastModified: f.lastModified,
      };
    } else {
      delete parsed[`file_${key}`];
    }
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(parsed));
  } catch {
    /* ignore */
  }
}
function clearDraft() {
  if (isServer) return;
  try {
    sessionStorage.removeItem(STORAGE_KEY);
  } catch {
    /* ignore */
  }
}

export function InscriptionForm() {
  const [submitted, setSubmitted] = createSignal(false);
  const [submitting, setSubmitting] = createSignal(false);
  const [error, setError] = createSignal("");
  const [ciInvalid, setCiInvalid] = createSignal("");
  const [fpvInvalid, setFpvInvalid] = createSignal("");
  const [emailInvalid, setEmailInvalid] = createSignal("");

  // Datos del formulario (persistidos en sessionStorage)
  const [cedula, setCedula] = createSessionField("cedula", "");
  const [nacionalidad, setNacionalidad] = createSessionField("nacionalidad", "V");
  const [nombres, setNombres] = createSessionField("nombres", "");
  const [apellidos, setApellidos] = createSessionField("apellidos", "");
  const [segundoNombre, setSegundoNombre] = createSessionField("segundo_nombre", "");
  const [segundoApellido, setSegundoApellido] = createSessionField("segundo_apellido", "");
  const [genero, setGenero] = createSessionField("genero", "");
  const [fpv, setFpv] = createSessionField("fpv", "");
  const [telefono, setTelefono] = createSessionField("telefono", "");
  const [correo, setCorreo] = createSessionField("correo", "");
  const [fechaNacimiento, setFechaNacimiento] = createSessionField("fecha_nacimiento", "");
  const [universidad, setUniversidad] = createSessionField("universidad", "");
  const [fechaGraduacion, setFechaGraduacion] = createSessionField("fecha_graduacion", "");
  const [mencion, setMencion] = createSessionField("mencion", "");
  const [regNumero, setRegNumero] = createSessionField("reg_numero", "");
  const [regEstado, setRegEstado] = createSessionField("reg_estado", "");
  const [regTomo, setRegTomo] = createSessionField("reg_tomo", "");
  const [regFolio, setRegFolio] = createSessionField("reg_folio", "");
  const [rif, setRif] = createSessionField("rif", "");

  // ── Ubicación y áreas de la ficha ──────────────────────────────────────
  const [serviceAddress, setServiceAddress] = createSessionField("service_address", "");
  const [municipalityCarabobo, setMunicipalityCarabobo] = createSessionField("municipality_carabobo", "");
  const [stateOutside, setStateOutside] = createSessionField("state_outside", "");
  const [municipalityOutside, setMunicipalityOutside] = createSessionField("municipality_outside_carabobo", "");
  const [country, setCountry] = createSessionField("country", "");
  // Modalidades (no se persisten; se mantienen en señal local)
  const [modPresencial, setModPresencial] = createSignal(false);
  const [modDistance, setModDistance] = createSignal(false);
  const [modTelephone, setModTelephone] = createSignal(false);
  // Áreas de trabajo (ids del catálogo)
  const [primarySpecialtyId, setPrimarySpecialtyId] = createSignal("");
  const [secondarySpecialtyId, setSecondarySpecialtyId] = createSignal("");
  const [workAreas] = createResource<WorkArea[]>(() => apiGet("/specialties"));

  // Archivos: conservamos el File en memoria y su metadato en sessionStorage
  const [foto, setFotoState] = createSignal<File | null>(null);
  const [comprobante, setComprobanteState] = createSignal<File | null>(null);
  const [docCedula, setDocCedulaState] = createSignal<File | null>(null);
  const [docTitulo, setDocTituloState] = createSignal<File | null>(null);
  const [docRif, setDocRifState] = createSignal<File | null>(null);
  const [docOtro, setDocOtroState] = createSignal<File | null>(null);
  // metadato restaurado que simula el estado "archivo seleccionado" tras recargar
  const fotoMeta = loadFileMeta("foto");
  const comprobanteMeta = loadFileMeta("comprobante");
  const docCedulaMeta = loadFileMeta("doc_cedula");
  const docTituloMeta = loadFileMeta("doc_titulo");
  const docRifMeta = loadFileMeta("doc_rif");
  const docOtroMeta = loadFileMeta("doc_otro");

  const setFoto = (f: File | null) => { setFotoState(f); saveFileMeta("foto", f); };
  const setComprobante = (f: File | null) => { setComprobanteState(f); saveFileMeta("comprobante", f); };
  const setDocCedula = (f: File | null) => { setDocCedulaState(f); saveFileMeta("doc_cedula", f); };
  const setDocTitulo = (f: File | null) => { setDocTituloState(f); saveFileMeta("doc_titulo", f); };
  const setDocRif = (f: File | null) => { setDocRifState(f); saveFileMeta("doc_rif", f); };
  const setDocOtro = (f: File | null) => { setDocOtroState(f); saveFileMeta("doc_otro", f); };

  const validate = (): string => {
    if (ciInvalid()) return "La cédula ingresada no está disponible";
    if (fpvInvalid()) return "El número de FPV ingresado no está disponible";
    if (emailInvalid()) return "El correo ingresado no está disponible";
    if (!cedula().trim()) return "La cédula es obligatoria";
    if (!nombres().trim()) return "Los nombres son obligatorios";
    if (!apellidos().trim()) return "Los apellidos son obligatorios";
    if (!segundoApellido().trim()) return "El segundo apellido es obligatorio";
    if (!genero().trim()) return "El género es obligatorio";
    if (!telefono().trim()) return "El teléfono de contacto es obligatorio";
    if (!fechaNacimiento().trim()) return "La fecha de nacimiento es obligatoria";
    if (!universidad().trim()) return "La universidad es obligatoria";
    if (!fechaGraduacion().trim()) return "La fecha de graduación es obligatoria";
    if (!regEstado().trim()) return "El estado del registro es obligatorio";
    const carabobo = municipalityCarabobo().trim() !== "" && serviceAddress().trim() !== "";
    const otroEstado = stateOutside().trim() !== "" && municipalityOutside().trim() !== "";
    const exterior = country().trim() !== "";
    if (!carabobo && !otroEstado && !exterior) {
      return "Debes completar al menos una ubicación completa (Carabobo, otro estado o exterior)";
    }
    if (!correo().trim()) return "El correo es obligatorio";
    if (!/(.+)@(.+)\.(.+)/.test(correo().trim())) return "Ingresa un correo válido";
    if (!foto() && !fotoMeta) return "Debes adjuntar la foto tipo carnet";
    if (!comprobante() && !comprobanteMeta) return "Debes adjuntar el comprobante de pago";
    if (!docCedula() && !docCedulaMeta) return "Debes adjuntar la copia de la cédula";
    if (!docTitulo() && !docTituloMeta) return "Debes adjuntar la copia del título";
    if (!docRif() && !docRifMeta) return "Debes adjuntar la copia del RIF";
    return "";
  };

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    const v = validate();
    if (v) { setError(v); return; }
    setError("");
    setSubmitting(true);

    try {
      const fd = new FormData();
      fd.set("cedula", cedula());
      fd.set("nacionalidad", nacionalidad());
      fd.set("nombres", nombres());
      fd.set("apellidos", apellidos());
      if (segundoNombre()) fd.set("segundo_nombre", segundoNombre());
      if (segundoApellido()) fd.set("segundo_apellido", segundoApellido());
      if (genero()) fd.set("genero", genero());
      if (fpv()) fd.set("fpv", fpv());
      if (telefono()) fd.set("telefono", telefono());
      fd.set("correo", correo());
      if (fechaNacimiento()) fd.set("fecha_nacimiento", fechaNacimiento());
      if (universidad()) fd.set("titulo_universidad", universidad());
      if (fechaGraduacion()) fd.set("titulo_fecha_graduacion", fechaGraduacion());
      if (mencion()) fd.set("titulo_mencion", mencion());
      if (regNumero()) fd.set("titulo_registro_numero", regNumero());
      if (regEstado()) fd.set("titulo_registro_estado", regEstado());
      if (regTomo()) fd.set("titulo_registro_tomo", regTomo());
      if (regFolio()) fd.set("titulo_registro_folio", regFolio());
      if (rif()) fd.set("rif", rif());
      if (serviceAddress()) fd.set("service_address", serviceAddress());
      if (municipalityCarabobo()) fd.set("municipality_carabobo", municipalityCarabobo());
      if (stateOutside()) fd.set("state_outside", stateOutside());
      if (municipalityOutside()) fd.set("municipality_outside_carabobo", municipalityOutside());
      if (country()) fd.set("country", country());
      if (modPresencial()) fd.set("service_modality_presencial", "1");
      if (modDistance()) fd.set("service_modality_distance", "1");
      if (modTelephone()) fd.set("service_modality_telephone", "1");
      if (primarySpecialtyId()) fd.set("primary_specialty_id", primarySpecialtyId());
      if (secondarySpecialtyId()) fd.set("secondary_specialty_id", secondarySpecialtyId());
      if (foto()) fd.set("foto", foto()!);
      if (comprobante()) fd.set("comprobante", comprobante()!);
      if (docCedula()) fd.set("doc_cedula", docCedula()!);
      if (docTitulo()) fd.set("doc_titulo", docTitulo()!);
      if (docRif()) fd.set("doc_rif", docRif()!);
      if (docOtro()) fd.set("doc_otro", docOtro()!);

      await apiPost("/inscripcion/submit", fd);
      clearDraft();
      setSubmitted(true);
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 409) {
          setError(err.message || "La cédula o el FPV ya se encuentran registrados");
        } else if (err.status === 400) {
          setError(err.message || "Verifica los datos del formulario e inténtalo de nuevo");
        } else if (err.status === 503 || err.status === 0) {
          setError("No se pudo conectar con el servidor. Verifica tu conexión a internet e inténtalo de nuevo.");
        } else {
          setError(err.message || "Ocurrió un error al enviar la solicitud, inténtalo de nuevo.");
        }
        console.error("Error al enviar la inscripción:", err);
      } else {
        setError("Ocurrió un error inesperado al enviar la solicitud, inténtalo de nuevo.");
        console.error("Error inesperado al enviar la inscripción:", err);
      }
    } finally {
      setSubmitting(false);
    }
  };

  const Field = (props: { label: string; required?: boolean; value: () => string; onChange: (v: string) => void; type?: string; placeholder?: string }) => {
    if (props.type === "date") {
      return (
        <label class="block">
          <span class="block text-sm font-bold text-gray-700 mb-1.5">
            {props.label} {props.required && <span class="text-red-500">*</span>}
          </span>
          <FlatDatePicker
            value={props.value()}
            onChange={props.onChange}
            placeholder={props.placeholder}
            class="w-full px-4 py-3 bg-white border border-gray-200 rounded-xl outline-none focus:ring-2 focus:ring-blue-100 transition-all"
          />
        </label>
      );
    }
    return (
      <label class="block">
        <span class="block text-sm font-bold text-gray-700 mb-1.5">
          {props.label} {props.required && <span class="text-red-500">*</span>}
        </span>
        <input
          type={props.type || "text"}
          value={props.value()}
          onInput={(e) => props.onChange((e.target as HTMLInputElement).value)}
          placeholder={props.placeholder}
          class="w-full px-4 py-3 bg-white border border-gray-200 rounded-xl outline-none focus:ring-2 focus:ring-blue-100 transition-all"
        />
      </label>
    );
  };

  const Section = (props: { n: number; title: string; children: any }) => {
    return (
      <div class="bg-white p-8 md:p-10 rounded-[2rem] shadow-sm border border-gray-100 space-y-6">
        <div class="flex items-center gap-4 border-b border-gray-100 pb-5">
          <div class="w-10 h-10 bg-colpsi-blue text-white rounded-xl flex items-center justify-center text-lg font-black shrink-0">
            {props.n}
          </div>
          <h3 class="text-2xl font-black text-colpsi-blue tracking-tight">{props.title}</h3>
        </div>
        {props.children}
      </div>
    );
  };

  return (
    <Show
      when={submitted()}
      fallback={(
        <form onSubmit={handleSubmit} class="space-y-8">
          <Section n={1} title="Datos Personales">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
              <CheckField
                label="Cédula de Identidad"
                required
                endpoint="/inscripcion/check-ci"
                param="ci"
                initialValue={cedula()}
                onChange={(v) => { setCedula(v); setCiInvalid(""); }}
                onValid={(v) => { setCedula(v); setCiInvalid(""); }}
                onInvalid={(m) => { if (m) setCiInvalid(m); }}
              />
              <div>
                <span class="block text-sm font-bold text-gray-700 mb-1.5">Nacionalidad <span class="text-red-500">*</span></span>
                <select
                  value={nacionalidad()}
                  onInput={(e) => setNacionalidad((e.target as HTMLSelectElement).value)}
                  class="w-full px-4 py-3 bg-white border border-gray-200 rounded-xl outline-none focus:ring-2 focus:ring-blue-100 transition-all"
                >
                  <option value="V">V - Venezolano</option>
                  <option value="E">E - Extranjero</option>
                </select>
              </div>
              <Field label="Nombres" required value={nombres} onChange={setNombres} />
              <Field label="Apellidos" required value={apellidos} onChange={setApellidos} />
              <Field label="Segundo nombre" value={segundoNombre} onChange={setSegundoNombre} />
              <Field label="Segundo apellido" required value={segundoApellido} onChange={setSegundoApellido} />
              <div>
                <span class="block text-sm font-bold text-gray-700 mb-1.5">Género <span class="text-red-500">*</span></span>
                <select
                  value={genero()}
                  onInput={(e) => setGenero((e.target as HTMLSelectElement).value)}
                  class="w-full px-4 py-3 bg-white border border-gray-200 rounded-xl outline-none focus:ring-2 focus:ring-blue-100 transition-all"
                >
                  <option value="">Seleccionar</option>
                  <option value="M">Masculino</option>
                  <option value="F">Femenino</option>
                </select>
              </div>
              <CheckField
                label="N° FPV (si lo posee)"
                endpoint="/inscripcion/check-fpv"
                param="fpv"
                initialValue={fpv()}
                onChange={(v) => { setFpv(v); setFpvInvalid(""); }}
                onValid={(v) => { setFpv(v); setFpvInvalid(""); }}
                onInvalid={(m) => { if (m) setFpvInvalid(m); }}
              />
              <Field label="Teléfono de contacto" required value={telefono} onChange={setTelefono} />
              <CheckField
                label="Correo electrónico"
                required
                type="email"
                endpoint="/inscripcion/check-email"
                param="correo"
                initialValue={correo()}
                onChange={(v) => { setCorreo(v); setEmailInvalid(""); }}
                onValid={(v) => { setCorreo(v); setEmailInvalid(""); }}
                onInvalid={(m) => { if (m) setEmailInvalid(m); }}
              />
              <Field label="Fecha de nacimiento" required type="date" value={fechaNacimiento} onChange={setFechaNacimiento} />
            </div>
          </Section>

          <Section n={2} title="Datos Académicos y Registro del Título">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
              <Field label="Universidad" required value={universidad} onChange={setUniversidad} />
              <Field label="Fecha de graduación" required type="date" value={fechaGraduacion} onChange={setFechaGraduacion} />
              <Field label="Mención" value={mencion} onChange={setMencion} />
              <Field label="N° Registro del título" value={regNumero} onChange={setRegNumero} />
              <Field label="Estado del registro" required value={regEstado} onChange={setRegEstado} />
              <Field label="Tomo del registro" value={regTomo} onChange={setRegTomo} />
              <Field label="Folio del registro" value={regFolio} onChange={setRegFolio} />
              <Field label="RIF" value={rif} onChange={setRif} />
            </div>
          </Section>

          <Section n={3} title="Ubicación y Modalidad de Servicio">
            <div class="space-y-8">
              <p class="text-xs text-gray-500">Se requiere al menos un bloque de ubicación completo: Carabobo (municipio + dirección), otro estado (estado + municipio/ciudad) o país (exterior).</p>
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
                <div>
                  <span class="block text-sm font-bold text-gray-700 mb-1.5">Municipio (Carabobo)</span>
                  <select
                    value={municipalityCarabobo()}
                    onChange={(e) => setMunicipalityCarabobo(e.currentTarget.value)}
                    class="w-full px-4 py-3 bg-white border border-gray-200 rounded-xl outline-none focus:ring-2 focus:ring-blue-100 transition-all"
                  >
                    <option value="">Seleccionar municipio…</option>
                    <For each={MUNICIPIOS_CARABOBO}>{(m) => <option value={m}>{m}</option>}</For>
                  </select>
                </div>
                <Field label="Dirección del consultorio" value={serviceAddress} onChange={setServiceAddress} />
                <div>
                  <span class="block text-sm font-bold text-gray-700 mb-1.5">Otro estado (fuera de Carabobo)</span>
                  <select
                    value={stateOutside()}
                    onChange={(e) => setStateOutside(e.currentTarget.value)}
                    class="w-full px-4 py-3 bg-white border border-gray-200 rounded-xl outline-none focus:ring-2 focus:ring-blue-100 transition-all"
                  >
                    <option value="">Seleccionar estado…</option>
                    <For each={ESTADOS_VENEZUELA}>{(e) => <option value={e}>{e}</option>}</For>
                  </select>
                </div>
                <Field label="Municipio / ciudad (fuera de Carabobo)" value={municipalityOutside} onChange={setMunicipalityOutside} />
                <Field label="País (fuera de Venezuela)" value={country} onChange={setCountry} placeholder="Ej: España" />
              </div>

              <div class="bg-gray-50 p-5 rounded-2xl border border-gray-100">
                <span class="block text-sm font-bold text-gray-700 mb-3">Modalidad de servicio</span>
                <div class="flex flex-wrap gap-4">
                  <label class="inline-flex items-center gap-2 text-sm font-semibold text-gray-700">
                    <input type="checkbox" checked={modPresencial()} onChange={(e) => setModPresencial(e.currentTarget.checked)} class="accent-colpsi-blue" />
                    Presencial
                  </label>
                  <label class="inline-flex items-center gap-2 text-sm font-semibold text-gray-700">
                    <input type="checkbox" checked={modDistance()} onChange={(e) => setModDistance(e.currentTarget.checked)} class="accent-colpsi-blue" />
                    A distancia
                  </label>
                  <label class="inline-flex items-center gap-2 text-sm font-semibold text-gray-700">
                    <input type="checkbox" checked={modTelephone()} onChange={(e) => setModTelephone(e.currentTarget.checked)} class="accent-colpsi-blue" />
                    Telefónica
                  </label>
                </div>
              </div>
            </div>
          </Section>

          <Section n={4} title="Áreas de Trabajo">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
              <div>
                <span class="block text-sm font-bold text-gray-700 mb-1.5">Área principal</span>
                <select
                  value={primarySpecialtyId()}
                  onChange={(e) => setPrimarySpecialtyId(e.currentTarget.value)}
                  class="w-full px-4 py-3 bg-white border border-gray-200 rounded-xl outline-none focus:ring-2 focus:ring-blue-100 transition-all"
                >
                  <option value="">— Sin área —</option>
                  <For each={workAreas() ?? []}>
                    {(wa) => <option value={String(wa.id)}>{wa.name}</option>}
                  </For>
                </select>
              </div>
              <div>
                <span class="block text-sm font-bold text-gray-700 mb-1.5">Área secundaria</span>
                <select
                  value={secondarySpecialtyId()}
                  onChange={(e) => setSecondarySpecialtyId(e.currentTarget.value)}
                  class="w-full px-4 py-3 bg-white border border-gray-200 rounded-xl outline-none focus:ring-2 focus:ring-blue-100 transition-all"
                >
                  <option value="">— Sin área —</option>
                  <For each={workAreas() ?? []}>
                    {(wa) => <option value={String(wa.id)} disabled={String(wa.id) === primarySpecialtyId()}>{wa.name}</option>}
                  </For>
                </select>
              </div>
            </div>
          </Section>

          <Section n={5} title="Documentos Requeridos">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
              <FileUpload
                label="Foto de perfil (tipo carnet)"
                accept="image/*"
                description="Imagen (máx. 5MB)"
                file={foto()}
                savedName={fotoMeta ? fotoMeta.name : undefined}
                onFile={setFoto}
              />
              <FileUpload
                label="Comprobante de pago"
                accept="image/*,application/pdf"
                description="Imagen o PDF (máx. 5MB)"
                file={comprobante()}
                savedName={comprobanteMeta ? comprobanteMeta.name : undefined}
                onFile={setComprobante}
              />
              <FileUpload
                label="Copia de la cédula de identidad"
                accept="image/*,application/pdf"
                description="Obligatorio · imagen o PDF (máx. 5MB)"
                file={docCedula()}
                savedName={docCedulaMeta ? docCedulaMeta.name : undefined}
                onFile={setDocCedula}
              />
              <FileUpload
                label="Copia del título de psicólogo"
                accept="image/*,application/pdf"
                description="Obligatorio · imagen o PDF (máx. 5MB)"
                file={docTitulo()}
                savedName={docTituloMeta ? docTituloMeta.name : undefined}
                onFile={setDocTitulo}
              />
              <FileUpload
                label="Copia del RIF vigente"
                accept="image/*,application/pdf"
                description="Obligatorio · imagen o PDF (máx. 5MB)"
                file={docRif()}
                savedName={docRifMeta ? docRifMeta.name : undefined}
                onFile={setDocRif}
              />
              <FileUpload
                label="Otro documento (opcional)"
                accept="image/*,application/pdf"
                description="Opcional · imagen o PDF (máx. 5MB)"
                file={docOtro()}
                savedName={docOtroMeta ? docOtroMeta.name : undefined}
                onFile={setDocOtro}
              />
            </div>
          </Section>

          <Show when={error()}>
            <div class="bg-red-50 border border-red-200 text-red-700 text-sm font-semibold rounded-2xl p-4">
              {error()}
            </div>
          </Show>

          <button
            type="submit"
            disabled={submitting()}
            class="w-full bg-colpsi-yellow text-colpsi-blue py-4 px-6 rounded-2xl font-black text-lg shadow-premium hover:scale-[1.01] active:scale-[0.99] transition-transform disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-3"
          >
            <Show when={submitting()} fallback={<span>Enviar solicitud</span>}>
              <span class="animate-spin inline-block h-5 w-5 border-2 border-colpsi-blue border-t-transparent rounded-full" />
              Enviando...
            </Show>
          </button>
        </form>
      )}
    >
      <SuccessMessage />
    </Show>
  );
}