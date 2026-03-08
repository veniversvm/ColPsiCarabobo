import { createResource, createEffect, Show, Suspense, createSignal, For } from "solid-js";
import { createStore, unwrap } from "solid-js/store";
import { useParams, useNavigate, action, useAction } from "@solidjs/router";
import { apiGet } from "~/lib/api";
import { ToggleSwitch } from "~/components/ui/ToggleSwitch";
import { RichTextEditor } from "~/components/ui/RichTextEditor";

/**
 * ACCIÓN DE SERVIDOR (BFF)
 * Envía la actualización parcial (PATCH) en formato JSON puro.
 */
const updateAdminPsiServer = action(async (params: { id: string; payload: any }) => {
  "use server";
  const { apiPatch } = await import("~/lib/api");

  // Limpieza: strings vacíos → null para que Go los trate como nil pointer
  const cleanPayload = { ...params.payload };
  Object.keys(cleanPayload).forEach((key) => {
    if (cleanPayload[key] === "") cleanPayload[key] = null;
  });

  return await apiPatch(`/admin/psi/${params.id}`, cleanPayload);
});

// Parsea ISO → YYYY-MM-DD para inputs type="date"
const formatDate = (dateStr?: string) => (dateStr ? dateStr.split("T")[0] : "");

// ─── Sub-componente reutilizable ────────────────────────────────────────────
function Field(props: {
  label: string;
  children: any;
}) {
  return (
    <div>
      <label class="block text-[10px] font-black text-gray-500 uppercase tracking-widest ml-1 mb-1">
        {props.label}
      </label>
      {props.children}
    </div>
  );
}

// ────────────────────────────────────────────────────────────────────────────

export default function AdminEditPsiPage() {
  const params = useParams();
  const navigate = useNavigate();
  const runUpdateAction = useAction(updateAdminPsiServer);

  const IC = "w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none transition-all text-gray-800 text-sm";

  // ── Carga de datos ──────────────────────────────────────────────────────
  const [profile, { refetch }] = createResource(() =>
    apiGet<any>(`/admin/psi/${params.id}`)
  );

  // ── Store del formulario ────────────────────────────────────────────────
  const [form, setForm] = createStore<any>({});
  const [saving, setSaving] = createSignal(false);
  const [message, setMessage] = createSignal<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  // ── Sincronización BD → Store ───────────────────────────────────────────
  createEffect(() => {
    const p = profile();
    if (!p) return;

    setForm({
      // ── Identidad Legal ───────────────────────────────────────────────
      first_name: p.first_name ?? "",
      second_name: p.second_name ?? "",
      last_name: p.last_name ?? "",
      second_last_name: p.second_last_name ?? "",
      ci: p.ci ?? "",
      fpv: p.fpv ?? "",
      nationality: p.nationality ?? "V",
      genre: p.genre ?? "M",
      born_date: formatDate(p.born_date),
      email: p.email ?? "",

      // ── Estatus Institucional ─────────────────────────────────────────
      is_active: p.is_active ?? true,
      solvent: p.solvent ?? false,
      proof_of_life: p.proof_of_life ?? false,

      // ── Contacto & Visibilidad ────────────────────────────────────────
      contact_email: p.contact_email ?? "",
      show_contact_email: p.show_contact_email ?? false,
      public_phone: p.public_phone ?? "",
      show_public_phone: p.show_public_phone ?? false,
      service_address: p.service_address ?? "",
      show_service_address: p.show_public_service_address ?? false,

      // ── Ubicación Carabobo ────────────────────────────────────────────
      municipality_carabobo: p.municipality_carabobo ?? "",
      phone_carabobo: p.phone_carabobo ?? "",
      cel_phone_carabobo: p.cel_phone_carabobo ?? "",

      // ── Ubicación Exterior ────────────────────────────────────────────
      state_outside: p.state_outside ?? "",
      municipality_outside_carabobo:
        p.municipality_outside_carabobo ?? p.municipality_out_side_carabobo ?? "",
      phone_outside_carabobo:
        p.phone_outside_carabobo ?? p.phone_out_side_carabobo ?? "",
      cel_phone_outside_carabobo:
        p.cel_phone_outside_carabobo ?? p.cel_phone_out_side_carabobo ?? "",

      // ── Perfil Profesional ────────────────────────────────────────────
      primary_specialty: p.primary_specialty ?? "",
      secondary_specialty: p.secondary_specialty ?? "",
      mini_bio: p.mini_bio ?? "",
      // full_bio viene como objeto { id, content, ... } — extraemos solo el HTML
      full_bio: p.full_bio?.content ?? "",

      // ── Datos Colegiales (ColData) ────────────────────────────────────
      university_undergraduate: p.col_data?.university_undergraduate ?? "",
      graduate_date: formatDate(p.col_data?.graduate_date),
      mention_undergraduate: p.col_data?.mention_undergraduate ?? "",
      register_number: p.col_data?.register_number ?? "",
      register_title_state: p.col_data?.register_title_state ?? "",
      register_title_date: formatDate(p.col_data?.register_title_date),
      register_folio: p.col_data?.register_folio ?? "",
      register_tome: p.col_data?.register_tome ?? "",

      // ── Flags Gremiales (ColData) ─────────────────────────────────────
      guild_director: p.col_data?.guild_director ?? false,
      sixty_five_or_plus: p.col_data?.sixty_five_or_plus ?? false,
      guild_collaborator: p.col_data?.guild_collaborator ?? false,
      public_employee: p.col_data?.public_employee ?? false,
      university_professor: p.col_data?.university_professor ?? false,
      double_guild: p.col_data?.double_guild ?? false,
      cpsm: p.col_data?.cpsm ?? false,
      date_of_last_solvency: formatDate(p.col_data?.date_of_last_solvency),
    });
  });

  // ── Submit ──────────────────────────────────────────────────────────────
  const handleSave = async (e: Event) => {
    e.preventDefault();
    setSaving(true);
    setMessage(null);

    const payload = {
      ...unwrap(form),
      ci: parseInt(form.ci) || null,
      fpv: parseInt(form.fpv) || null,
      register_number: parseInt(form.register_number) || null,
    };

    try {
      await runUpdateAction({ id: params.id ?? "", payload });
      setMessage({ type: "success", text: "Expediente actualizado exitosamente." });
      refetch();
      window.scrollTo({ top: 0, behavior: "smooth" });
    } catch (err: any) {
      setMessage({ type: "error", text: err.message || "Error al actualizar el expediente." });
      window.scrollTo({ top: 0, behavior: "smooth" });
    } finally {
      setSaving(false);
    }
  };

  // ────────────────────────────────────────────────────────────────────────
  return (
    <main class="pb-28 animate-in fade-in duration-500">

      {/* ── HEADER ──────────────────────────────────────────────────────── */}
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8 bg-white p-6 rounded-3xl shadow-sm border border-gray-100">
        <div class="flex items-center gap-4">
          <button
            onClick={() => navigate(-1)}
            class="w-10 h-10 bg-gray-50 hover:bg-gray-100 text-gray-600 rounded-full font-bold flex items-center justify-center transition-colors"
          >
            ←
          </button>

          {/* Avatar */}
          <Show when={profile()?.profile_picture_url}>
            <img
              src={`http://localhost:9000/colpsi-bucket/${profile()?.profile_picture_url}`}
              class="w-14 h-14 rounded-2xl object-cover border-2 border-gray-100 shadow"
              alt="Foto de perfil"
            />
          </Show>

          <div>
            <h1 class="text-2xl font-black text-blue-800 uppercase">Expediente de Colegiado</h1>
            <p class="text-gray-500 text-sm font-bold tracking-widest mt-0.5">
              FPV: {profile()?.fpv || "—"} · USUARIO: {profile()?.username || "—"}
            </p>
          </div>
        </div>

        <div class="flex items-center gap-3 flex-wrap">
          <Show when={profile()?.solvent}>
            <span class="bg-green-100 text-green-700 px-3 py-1.5 rounded-lg text-xs font-black uppercase">
              Solvente
            </span>
          </Show>
          <Show when={!profile()?.is_active}>
            <span class="bg-red-100 text-red-700 px-3 py-1.5 rounded-lg text-xs font-black uppercase">
              Suspendido
            </span>
          </Show>
          <Show when={profile()?.proof_of_life}>
            <span class="bg-blue-100 text-blue-700 px-3 py-1.5 rounded-lg text-xs font-black uppercase">
              Fe de Vida ✓
            </span>
          </Show>
        </div>
      </div>

      <Suspense fallback={<div class="h-96 bg-white animate-pulse rounded-3xl" />}>

        {/* ── MENSAJE ────────────────────────────────────────────────────── */}
        <Show when={message()}>
          <div
            class={`mb-8 p-4 rounded-2xl font-bold text-sm shadow-sm border-l-4 ${
              message()?.type === "success"
                ? "bg-green-50 text-green-800 border-green-500"
                : "bg-red-50 text-red-800 border-red-500"
            }`}
          >
            {message()?.text}
          </div>
        </Show>

        <form onSubmit={handleSave} class="space-y-8">

          {/* ══ BLOQUE 1: ESTATUS ADMINISTRATIVO ════════════════════════════ */}
          <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-blue-100 relative overflow-hidden">
            <div class="absolute top-0 left-0 w-2 h-full bg-yellow-400" />
            <h2 class="text-lg font-black text-blue-800 mb-6 ml-2">Estatus Administrativo</h2>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6 ml-2">
              {/* Switches de estado */}
              <div class="space-y-4 bg-gray-50 p-6 rounded-2xl border border-gray-100">
                <ToggleSwitch label="Cuenta Activa en Sistema" checked={form.is_active} onChange={(v) => setForm("is_active", v)} />
                <ToggleSwitch label="Estado: Solvente" checked={form.solvent} onChange={(v) => setForm("solvent", v)} />
                <ToggleSwitch label="Fe de Vida Activa" checked={form.proof_of_life} onChange={(v) => setForm("proof_of_life", v)} />
                <div class="pt-3">
                  <Field label="Fecha Última Solvencia">
                    <input type="date" value={form.date_of_last_solvency} onInput={(e) => setForm("date_of_last_solvency", e.currentTarget.value)} class={IC} />
                  </Field>
                </div>
              </div>

              {/* Roles gremiales */}
              <div class="space-y-3 bg-blue-50/50 p-6 rounded-2xl border border-blue-100">
                <h3 class="text-xs font-bold text-blue-800 uppercase tracking-widest border-b border-blue-100 pb-2 mb-3">
                  Roles Gremiales
                </h3>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <ToggleSwitch label="Director" checked={form.guild_director} onChange={(v) => setForm("guild_director", v)} />
                  <ToggleSwitch label="Colaborador" checked={form.guild_collaborator} onChange={(v) => setForm("guild_collaborator", v)} />
                  <ToggleSwitch label="Prof. Universitario" checked={form.university_professor} onChange={(v) => setForm("university_professor", v)} />
                  <ToggleSwitch label="Empleado Público" checked={form.public_employee} onChange={(v) => setForm("public_employee", v)} />
                  <ToggleSwitch label="Doble Gremio" checked={form.double_guild} onChange={(v) => setForm("double_guild", v)} />
                  <ToggleSwitch label="65+ Años" checked={form.sixty_five_or_plus} onChange={(v) => setForm("sixty_five_or_plus", v)} />
                  <ToggleSwitch label="CPSM" checked={form.cpsm} onChange={(v) => setForm("cpsm", v)} />
                </div>
              </div>
            </div>
          </section>

          {/* ══ BLOQUE 2: IDENTIDAD LEGAL ════════════════════════════════════ */}
          <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100">
            <h2 class="text-lg font-black text-blue-800 mb-6">Identidad Legal</h2>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-5">
              <div class="md:col-span-2">
                <Field label="Nombres">
                  <div class="flex gap-2">
                    <input type="text" required placeholder="Primer Nombre" value={form.first_name} onInput={(e) => setForm("first_name", e.currentTarget.value)} class={IC} />
                    <input type="text" placeholder="Segundo Nombre" value={form.second_name} onInput={(e) => setForm("second_name", e.currentTarget.value)} class={IC} />
                  </div>
                </Field>
              </div>

              <div class="md:col-span-2">
                <Field label="Apellidos">
                  <div class="flex gap-2">
                    <input type="text" required placeholder="Primer Apellido" value={form.last_name} onInput={(e) => setForm("last_name", e.currentTarget.value)} class={IC} />
                    <input type="text" placeholder="Segundo Apellido" value={form.second_last_name} onInput={(e) => setForm("second_last_name", e.currentTarget.value)} class={IC} />
                  </div>
                </Field>
              </div>

              <div>
                <Field label="Cédula de Identidad">
                  <div class="flex gap-2">
                    <select value={form.nationality} onChange={(e) => setForm("nationality", e.currentTarget.value)} class={`${IC} w-24`}>
                      <option value="V">V</option>
                      <option value="E">E</option>
                    </select>
                    <input type="number" required value={form.ci} onInput={(e) => setForm("ci", e.currentTarget.value)} class={IC} />
                  </div>
                </Field>
              </div>

              <div>
                <Field label="Nro. FPV">
                  <input type="number" required value={form.fpv} onInput={(e) => setForm("fpv", e.currentTarget.value)} class={`${IC} bg-yellow-50 border-yellow-300 font-bold`} />
                </Field>
              </div>

              <div>
                <Field label="Género">
                  <select value={form.genre} onChange={(e) => setForm("genre", e.currentTarget.value)} class={IC}>
                    <option value="M">Masculino</option>
                    <option value="F">Femenino</option>
                  </select>
                </Field>
              </div>

              <div>
                <Field label="Nacimiento">
                  <input type="date" required value={form.born_date} onInput={(e) => setForm("born_date", e.currentTarget.value)} class={IC} />
                </Field>
              </div>

              <div class="md:col-span-2">
                <Field label="Email Institucional (login)">
                  <input type="email" value={form.email} onInput={(e) => setForm("email", e.currentTarget.value)} class={IC} />
                </Field>
              </div>
            </div>
          </section>

          {/* ══ BLOQUE 3: CONTACTO & VISIBILIDAD ════════════════════════════ */}
          <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100">
            <h2 class="text-lg font-black text-blue-800 mb-6">Contacto Público & Visibilidad</h2>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Email de contacto */}
              <div class="space-y-3 bg-gray-50 rounded-2xl p-5 border border-gray-100">
                <Field label="Email de Contacto Público">
                  <input type="email" value={form.contact_email} onInput={(e) => setForm("contact_email", e.currentTarget.value)} class={IC} />
                </Field>
                <ToggleSwitch label="Mostrar email de contacto" checked={form.show_contact_email} onChange={(v) => setForm("show_contact_email", v)} />
              </div>

              {/* Teléfono público */}
              <div class="space-y-3 bg-gray-50 rounded-2xl p-5 border border-gray-100">
                <Field label="Teléfono Público">
                  <input type="tel" value={form.public_phone} onInput={(e) => setForm("public_phone", e.currentTarget.value)} class={IC} />
                </Field>
                <ToggleSwitch label="Mostrar teléfono público" checked={form.show_public_phone} onChange={(v) => setForm("show_public_phone", v)} />
              </div>

              {/* Dirección */}
              <div class="md:col-span-2 space-y-3 bg-gray-50 rounded-2xl p-5 border border-gray-100">
                <Field label="Dirección de Servicio">
                  <input type="text" value={form.service_address} onInput={(e) => setForm("service_address", e.currentTarget.value)} class={IC} />
                </Field>
                <ToggleSwitch label="Mostrar dirección de servicio" checked={form.show_service_address} onChange={(v) => setForm("show_service_address", v)} />
              </div>
            </div>
          </section>

          {/* ══ BLOQUE 4: UBICACIÓN ═══════════════════════════════════════════ */}
          <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100">
            <h2 class="text-lg font-black text-blue-800 mb-6">Ubicación</h2>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
              {/* Carabobo */}
              <div>
                <h3 class="text-xs font-black text-blue-700 uppercase tracking-widest mb-4 pb-2 border-b border-blue-50">
                  📍 Carabobo
                </h3>
                <div class="space-y-4">
                  <Field label="Municipio">
                    <input type="text" value={form.municipality_carabobo} onInput={(e) => setForm("municipality_carabobo", e.currentTarget.value)} class={IC} />
                  </Field>
                  <Field label="Teléfono Fijo">
                    <input type="tel" value={form.phone_carabobo} onInput={(e) => setForm("phone_carabobo", e.currentTarget.value)} class={IC} />
                  </Field>
                  <Field label="Celular">
                    <input type="tel" value={form.cel_phone_carabobo} onInput={(e) => setForm("cel_phone_carabobo", e.currentTarget.value)} class={IC} />
                  </Field>
                </div>
              </div>

              {/* Exterior */}
              <div>
                <h3 class="text-xs font-black text-purple-700 uppercase tracking-widest mb-4 pb-2 border-b border-purple-50">
                  🌐 Fuera de Carabobo
                </h3>
                <div class="space-y-4">
                  <Field label="Estado">
                    <input type="text" value={form.state_outside} onInput={(e) => setForm("state_outside", e.currentTarget.value)} class={IC} />
                  </Field>
                  <Field label="Municipio">
                    <input type="text" value={form.municipality_outside_carabobo} onInput={(e) => setForm("municipality_outside_carabobo", e.currentTarget.value)} class={IC} />
                  </Field>
                  <Field label="Teléfono Fijo">
                    <input type="tel" value={form.phone_outside_carabobo} onInput={(e) => setForm("phone_outside_carabobo", e.currentTarget.value)} class={IC} />
                  </Field>
                  <Field label="Celular">
                    <input type="tel" value={form.cel_phone_outside_carabobo} onInput={(e) => setForm("cel_phone_outside_carabobo", e.currentTarget.value)} class={IC} />
                  </Field>
                </div>
              </div>
            </div>
          </section>

          {/* ══ BLOQUE 5: PERFIL PROFESIONAL ════════════════════════════════ */}
          <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100">
            <h2 class="text-lg font-black text-blue-800 mb-6">Perfil Profesional</h2>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
              <Field label="Especialidad Principal">
                <input type="text" value={form.primary_specialty} onInput={(e) => setForm("primary_specialty", e.currentTarget.value)} class={IC} />
              </Field>
              <Field label="Especialidad Secundaria">
                <input type="text" value={form.secondary_specialty} onInput={(e) => setForm("secondary_specialty", e.currentTarget.value)} class={IC} />
              </Field>
              <div class="md:col-span-2">
                <Field label="Mini Bio (texto corto público)">
                  <textarea
                    rows={4}
                    value={form.mini_bio}
                    onInput={(e) => setForm("mini_bio", e.currentTarget.value)}
                    class={`${IC} resize-none`}
                    maxLength={500}
                  />
                  <p class="text-[10px] text-gray-400 mt-1 text-right">{(form.mini_bio || "").length}/500</p>
                </Field>
              </div>

              {/* Bio extensa con editor enriquecido */}
              <div class="md:col-span-2">
                <RichTextEditor
                  label="Biografía Extensa (Bio Pública Completa)"
                  content={form.full_bio}
                  onUpdate={(html) => setForm("full_bio", html)}
                />
              </div>
            </div>
          </section>

          {/* ══ BLOQUE 6: REGISTRO ACADÉMICO ════════════════════════════════ */}
          <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100">
            <h2 class="text-lg font-black text-blue-800 mb-6">Registro de Título (Pregrado)</h2>

            <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
              <div class="md:col-span-2">
                <Field label="Universidad de Egreso">
                  <input type="text" value={form.university_undergraduate} onInput={(e) => setForm("university_undergraduate", e.currentTarget.value)} class={IC} />
                </Field>
              </div>
              <Field label="Fecha de Egreso">
                <input type="date" value={form.graduate_date} onInput={(e) => setForm("graduate_date", e.currentTarget.value)} class={IC} />
              </Field>
              <div class="md:col-span-3">
                <Field label="Mención">
                  <input type="text" value={form.mention_undergraduate} onInput={(e) => setForm("mention_undergraduate", e.currentTarget.value)} class={IC} />
                </Field>
              </div>

              <div class="col-span-full border-t border-gray-100 my-1 pt-4">
                <p class="text-[10px] font-black text-gray-400 uppercase tracking-widest">Datos de Registro de Título en Estado</p>
              </div>

              <Field label="Nro. Registro">
                <input type="number" value={form.register_number} onInput={(e) => setForm("register_number", e.currentTarget.value)} class={IC} />
              </Field>
              <Field label="Estado de Registro">
                <input type="text" value={form.register_title_state} onInput={(e) => setForm("register_title_state", e.currentTarget.value)} class={IC} />
              </Field>
              <Field label="Fecha Registro">
                <input type="date" value={form.register_title_date} onInput={(e) => setForm("register_title_date", e.currentTarget.value)} class={IC} />
              </Field>
              <Field label="Folio">
                <input type="text" value={form.register_folio} onInput={(e) => setForm("register_folio", e.currentTarget.value)} class={IC} />
              </Field>
              <Field label="Tomo">
                <input type="text" value={form.register_tome} onInput={(e) => setForm("register_tome", e.currentTarget.value)} class={IC} />
              </Field>
            </div>
          </section>

          {/* ══ BLOQUE 7: AUDITORÍA / SOLO LECTURA ══════════════════════════ */}
          <section class="bg-gray-50 rounded-3xl p-6 shadow-inner border border-gray-200">
            <h2 class="text-sm font-black text-gray-500 mb-5 uppercase tracking-widest">
              Información Adjunta (Solo Lectura)
            </h2>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Redes sociales */}
              <div class="bg-white p-5 rounded-2xl border border-gray-100">
                <h3 class="text-xs font-bold text-blue-800 mb-3">
                  Redes Sociales ({profile()?.social_networks?.length || 0})
                </h3>
                <ul class="text-sm text-gray-600 space-y-2">
                  <Show when={!profile()?.social_networks?.length}>
                    <li class="italic text-gray-400">Ninguna registrada</li>
                  </Show>
                  <For each={profile()?.social_networks}>
                    {(net: any) => (
                      <li class="flex items-center gap-2">
                        <span class="font-semibold text-gray-700">{net.name}:</span>
                        <a href={net.url} target="_blank" rel="noopener noreferrer" class="text-blue-500 hover:underline truncate">
                          {net.url}
                        </a>
                      </li>
                    )}
                  </For>
                </ul>
              </div>

              {/* Postgrados */}
              <div class="bg-white p-5 rounded-2xl border border-gray-100">
                <h3 class="text-xs font-bold text-blue-800 mb-3">
                  Postgrados ({profile()?.post_grades?.length || 0})
                </h3>
                <ul class="text-sm text-gray-600 space-y-2">
                  <Show when={!profile()?.post_grades?.length}>
                    <li class="italic text-gray-400">Ninguno registrado</li>
                  </Show>
                  <For each={profile()?.post_grades}>
                    {(pg: any) => (
                      <li class="flex flex-col">
                        <span class="font-semibold text-gray-800">{pg.post_grade_title}</span>
                        <span class="text-xs text-gray-400">{pg.post_grade_graduation_year}</span>
                      </li>
                    )}
                  </For>
                </ul>
              </div>
            </div>

            {/* Metadatos */}
            <div class="mt-5 bg-white rounded-2xl border border-gray-100 p-4 grid grid-cols-2 md:grid-cols-4 gap-4 text-[11px] text-gray-500">
              <div>
                <p class="font-black uppercase tracking-widest text-[9px] text-gray-400 mb-0.5">ID Interno</p>
                <p class="font-mono truncate">{profile()?.id}</p>
              </div>
              <div>
                <p class="font-black uppercase tracking-widest text-[9px] text-gray-400 mb-0.5">Creado por</p>
                <p>{profile()?.create_by}</p>
              </div>
              <div>
                <p class="font-black uppercase tracking-widest text-[9px] text-gray-400 mb-0.5">Último update</p>
                <p>{profile()?.update_by}</p>
              </div>
              <div>
                <p class="font-black uppercase tracking-widest text-[9px] text-gray-400 mb-0.5">Actualizado</p>
                <p>{profile()?.updated_at ? new Date(profile()!.updated_at).toLocaleDateString("es-VE") : "—"}</p>
              </div>
            </div>
          </section>

          {/* ── BOTÓN FLOTANTE ──────────────────────────────────────────── */}
          <div class="sticky bottom-6 z-50 flex justify-end">
            <button
              type="submit"
              disabled={saving()}
              class="bg-blue-800 text-white px-10 py-4 rounded-2xl font-black shadow-2xl hover:scale-105 active:scale-95 transition-all disabled:opacity-70 flex items-center gap-3 border-2 border-white"
            >
              {saving() ? "GUARDANDO..." : "💾 GUARDAR EXPEDIENTE"}
            </button>
          </div>

        </form>
      </Suspense>
    </main>
  );
}