import { createResource, createEffect, Show, Suspense, createSignal, For } from "solid-js";
import { createStore } from "solid-js/store";
import { A } from "@solidjs/router";
import { apiGet, apiPatch, apiPost, apiDelete, ApiError } from "~/lib/api";
import { ToggleSwitch } from "~/components/ui/ToggleSwitch";
import { sanitizeEmail, sanitizePhone, sanitizeText } from "~/lib/sanitizer";

export default function ProfilePage() {
  // 1. Cargar datos del usuario desde la API (Se usa refetch para actualizar la lista de RRSS al añadir/borrar)
  const [profile, { refetch }] = createResource(() => apiGet<any>("/psi/me"));

  // 2. ESTADO: Formulario Principal de Perfil (PATCH)
  const[form, setForm] = createStore<any>({});
  const [saving, setSaving] = createSignal(false);
  const [message, setMessage] = createSignal<{ type: "success" | "error"; text: string } | null>(null);

  // 3. ESTADO: Formulario de Redes Sociales (POST)
  const [socialForm, setSocialForm] = createStore({ name: "", url: "" });
  const [savingSocial, setSavingSocial] = createSignal(false);


  // 4. Sincronizar el Store principal cuando llegan los datos del backend
  createEffect(() => {
    const p = profile();
    if (p) {
      setForm({
        // Contacto
        contact_email: sanitizeEmail(p.contact_email) || "",
        public_phone: sanitizePhone(p.public_phone) || "",
        service_address: sanitizeText(p.service_address) || "",

        // Ubicación Carabobo
        municipality_carabobo: sanitizeText(p.municipality_carabobo) || "",
        phone_carabobo: sanitizePhone(p.phone_carabobo) || "",
        cel_phone_carabobo: sanitizePhone(p.cel_phone_carabobo) || "",

        // Ubicación Exterior / Otros Estados
        state_outside: sanitizeText(p.state_outside) || "",
        municipality_outside_carabobo: sanitizeText(p.municipality_outside_carabobo || p.municipality_out_side_carabobo) || "",
        phone_outside_carabobo: sanitizePhone(p.phone_outside_carabobo) || "",
        cel_phone_outside_carabobo: sanitizePhone(p.cel_phone_outside_carabobo) || "",

        // Profesional
        mini_bio: sanitizeText(p.mini_bio) || "",
        primary_specialty: p.primary_specialty || "",
        secondary_specialty: p.secondary_specialty || "",

        // PRIVACIDAD (Se usan punteros booleanos en el Backend, aquí mapeamos a booleanos estrictos)
        show_contact_email: p.show_contact_email ?? false,
        show_public_phone: p.show_public_phone ?? false,
        show_public_service_address: p.show_public_service_address ?? false,
        show_university_undergraduate: p.col_data?.show_university_undergraduate ?? false,
        show_graduate_date: p.col_data?.show_graduate_date ?? false,
        show_mention_undergraduate: p.col_data?.show_mention_undergraduate ?? false,
      });
    }
  });

  // =========================================================================
  // HANDLERS DE FORMULARIO
  // =========================================================================

  const handleSaveProfile = async (e: Event) => {
    e.preventDefault();
    setSaving(true);
    setMessage(null);
    try {
      await apiPatch("/psi/me", form);
      setMessage({ type: "success", text: "Perfil y configuración de privacidad actualizados correctamente." });
      setTimeout(() => setMessage(null), 4000);
    } catch (err: any) {
      if (err instanceof ApiError) {
        setMessage({ type: "error", text: err.message });
      } else {
        setMessage({ type: "error", text: "Error de red al guardar el perfil." });
      }
    } finally {
      setSaving(false);
    }
  };

  const handleAddSocial = async (e: Event) => {
    e.preventDefault();
    if (!socialForm.name || !socialForm.url) return;
    setSavingSocial(true);
    try {
      await apiPost("/psi/me/social", socialForm);
      setSocialForm({ name: "", url: "" }); // Limpiar formulario de red social
      refetch(); // Recargar el perfil desde Go para ver la nueva red en la lista
    } catch (err: any) {
      alert(err instanceof ApiError ? err.message : "Error de red al añadir la red social.");
    } finally {
      setSavingSocial(false);
    }
  };

  const handleDeleteSocial = async (id: string) => {
    if (!confirm("¿Estás seguro de que deseas eliminar esta red social de tu perfil público?")) return;
    try {
      await apiDelete(`/psi/me/social/${id}`);
      refetch(); // Recargar el perfil para quitar la red de la lista visual
    } catch (err: any) {
      alert("Error al eliminar el registro.");
    }
  };

  // =========================================================================
  // RENDERIZADO UI (Mobile-First & Clean Premium)
  // =========================================================================

  return (
    <main class="bg-[#f8fafc] min-h-screen pb-24 font-sans">
      {/* --- HEADER INSTITUCIONAL --- */}
      <div class="bg-colpsi-blue pt-10 pb-24 px-4 md:px-8 shadow-inner">
        <div class="max-w-4xl mx-auto flex items-center justify-between">
          <A href="/psi" class="text-white hover:text-colpsi-yellow font-bold flex items-center gap-2 transition-colors">
            <span>←</span> Volver al Panel
          </A>
          <span class="text-blue-200 text-sm font-black tracking-widest uppercase hidden sm:block">
            Ajustes de Perfil
          </span>
        </div>
        <div class="max-w-4xl mx-auto mt-8">
          <h1 class="text-white text-3xl md:text-4xl font-black">Tu Identidad Digital</h1>
          <p class="text-blue-200 mt-2 text-sm md:text-base">
            Actualiza tus datos y gestiona qué información pueden ver los pacientes en el directorio público.
          </p>
        </div>
      </div>

      {/* --- ÁREA PRINCIPAL --- */}
      <div class="max-w-4xl mx-auto px-4 md:px-8 -mt-12 relative z-10 space-y-8">
        <Suspense fallback={<div class="h-96 bg-white animate-pulse rounded-[2.5rem] shadow-premium border border-gray-100" />}>
          
          <Show when={message()}>
            <div class={`p-4 rounded-2xl font-bold text-sm shadow-sm animate-in fade-in slide-in-from-top-2 ${message()?.type === 'success' ? 'bg-green-50 text-green-700 border border-green-200' : 'bg-red-50 text-red-700 border border-red-200'}`}>
              {message()?.text}
            </div>
          </Show>

          {/* ======================================================== */}
          {/* FORMULARIO PRINCIPAL DE PERFIL (PATCH)                     */}
          {/* ======================================================== */}
          <form onSubmit={handleSaveProfile} class="space-y-8">
            
            {/* SECCIÓN 1: DATOS DE CONTACTO (Inputs Limpios) */}
            <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100">
              <div class="mb-6 border-l-4 border-colpsi-yellow pl-3">
                <h2 class="text-xl font-black text-colpsi-blue leading-tight">Contacto de Consulta</h2>
                <p class="text-xs text-colpsi-muted mt-1">Esta información es la que usarán los pacientes para contactarte.</p>
              </div>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div class="space-y-2">
                  <label class="text-xs font-bold text-gray-500 uppercase ml-2">Email de Consulta</label>
                  <input type="email" value={form.contact_email} onInput={(e) => setForm("contact_email", e.currentTarget.value)} class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all" placeholder="ejemplo@correo.com" />
                </div>
                <div class="space-y-2">
                  <label class="text-xs font-bold text-gray-500 uppercase ml-2">Teléfono Principal</label>
                  <input type="tel" value={form.public_phone} onInput={(e) => setForm("public_phone", e.currentTarget.value)} class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all" placeholder="Ej: 0414-1234567" />
                </div>
                <div class="md:col-span-2 space-y-2">
                  <label class="text-xs font-bold text-gray-500 uppercase ml-2">Dirección de Consultorio</label>
                  <input type="text" value={form.service_address} onInput={(e) => setForm("service_address", e.currentTarget.value)} class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all" placeholder="Ej: Torre Banaven, Piso 4, Ofic 42" />
                </div>
              </div>
            </section>

            {/* SECCIÓN 2: UBICACIÓN EN CARABOBO */}
            <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100">
              <h2 class="text-xl font-black text-colpsi-blue mb-6 border-l-4 border-colpsi-yellow pl-3">Ubicación en Carabobo</h2>
              <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div class="space-y-2">
                  <label class="text-xs font-bold text-gray-500 uppercase ml-2">Municipio</label>
                  <input type="text" value={form.municipality_carabobo} onInput={(e) => setForm("municipality_carabobo", e.currentTarget.value)} class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all" placeholder="Ej: Valencia"/>
                </div>
                <div class="space-y-2">
                  <label class="text-xs font-bold text-gray-500 uppercase ml-2">Teléfono Fijo</label>
                  <input type="tel" value={form.phone_carabobo} onInput={(e) => setForm("phone_carabobo", e.currentTarget.value)} class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all" />
                </div>
                <div class="space-y-2">
                  <label class="text-xs font-bold text-gray-500 uppercase ml-2">Celular Secundario</label>
                  <input type="tel" value={form.cel_phone_carabobo} onInput={(e) => setForm("cel_phone_carabobo", e.currentTarget.value)} class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all" />
                </div>
              </div>
            </section>

            {/* SECCIÓN 3: UBICACIÓN FUERA DE CARABOBO */}
            <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100">
              <h2 class="text-xl font-black text-colpsi-blue mb-6 border-l-4 border-gray-300 pl-3">Ubicación en el Exterior u Otros Estados</h2>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div class="space-y-2">
                  <label class="text-xs font-bold text-gray-500 uppercase ml-2">Estado / Región / País</label>
                  <input type="text" value={form.state_outside} onInput={(e) => setForm("state_outside", e.currentTarget.value)} class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all" placeholder="Ej: Madrid, España"/>
                </div>
                <div class="space-y-2">
                  <label class="text-xs font-bold text-gray-500 uppercase ml-2">Ciudad / Municipio</label>
                  <input type="text" value={form.municipality_outside_carabobo} onInput={(e) => setForm("municipality_outside_carabobo", e.currentTarget.value)} class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all" />
                </div>
                <div class="space-y-2">
                  <label class="text-xs font-bold text-gray-500 uppercase ml-2">Teléfono Fijo (Internacional)</label>
                  <input type="tel" value={form.phone_outside_carabobo} onInput={(e) => setForm("phone_outside_carabobo", e.currentTarget.value)} class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all" />
                </div>
                <div class="space-y-2">
                  <label class="text-xs font-bold text-gray-500 uppercase ml-2">Celular (Internacional)</label>
                  <input type="tel" value={form.cel_phone_outside_carabobo} onInput={(e) => setForm("cel_phone_outside_carabobo", e.currentTarget.value)} class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all" />
                </div>
              </div>
            </section>

            {/* SECCIÓN 4: RESEÑA PROFESIONAL */}
            <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100">
              <h2 class="text-xl font-black text-colpsi-blue mb-6 border-l-4 border-colpsi-yellow pl-3">Perfil Profesional</h2>
              <div class="space-y-6">
                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                   <div class="space-y-2">
                    <label class="text-xs font-bold text-gray-500 uppercase ml-2">Especialidad Principal</label>
                    <input type="text" value={form.primary_specialty} onInput={(e) => setForm("primary_specialty", e.currentTarget.value)} class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all" placeholder="Ej: Psicología Clínica"/>
                  </div>
                  <div class="space-y-2">
                    <label class="text-xs font-bold text-gray-500 uppercase ml-2">Especialidad Secundaria</label>
                    <input type="text" value={form.secondary_specialty} onInput={(e) => setForm("secondary_specialty", e.currentTarget.value)} class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-5 py-3 outline-none text-colpsi-text transition-all" placeholder="Ej: Orientación Vocacional"/>
                  </div>
                </div>
                <div class="space-y-2">
                  <label class="text-xs font-bold text-gray-500 uppercase ml-2">Mini Biografía (Descripción Pública)</label>
                  <textarea value={form.mini_bio} onInput={(e) => setForm("mini_bio", e.currentTarget.value)} class="w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-yellow rounded-2xl px-5 py-4 outline-none text-colpsi-text transition-all min-h-[120px] resize-y" placeholder="Describe brevemente tu enfoque terapéutico, experiencia y a quién va dirigida tu consulta..."></textarea>
                </div>
              </div>
            </section>

            {/* SECCIÓN 5: PRIVACIDAD Y VISIBILIDAD (CENTRAL DE CONTROL) */}
            <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100">
              <div class="mb-6 border-l-4 border-colpsi-yellow pl-3">
                <h2 class="text-xl font-black text-colpsi-blue leading-tight">Centro de Privacidad</h2>
                <p class="text-sm text-colpsi-muted mt-1">Elige exactamente qué información deseas hacer pública en tu tarjeta del directorio.</p>
              </div>
              
              <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
                {/* Bloque: Información de Contacto */}
                <div class="space-y-4 bg-blue-50/50 p-6 rounded-3xl border border-blue-100">
                  <h3 class="text-sm font-bold text-colpsi-blue uppercase tracking-widest border-b border-blue-100 pb-2">Información de Contacto</h3>
                  <ToggleSwitch label="Mostrar Email de Contacto" checked={!!form.show_contact_email} onChange={(v) => setForm("show_contact_email", v)} />
                  <ToggleSwitch label="Mostrar Teléfono Público" checked={!!form.show_public_phone} onChange={(v) => setForm("show_public_phone", v)} />
                  <ToggleSwitch label="Mostrar Dirección de Consulta" checked={!!form.show_public_service_address} onChange={(v) => setForm("show_public_service_address", v)} />
                </div>

                {/* Bloque: Información Académica */}
                <div class="space-y-4 bg-gray-50 p-6 rounded-3xl border border-gray-100">
                  <h3 class="text-sm font-bold text-colpsi-text uppercase tracking-widest border-b border-gray-200 pb-2">Datos Académicos (Pregrado)</h3>
                  <ToggleSwitch label="Mostrar Universidad de Egreso" checked={!!form.show_university_undergraduate} onChange={(v) => setForm("show_university_undergraduate", v)} />
                  <ToggleSwitch label="Mostrar Fecha de Egreso" checked={!!form.show_graduate_date} onChange={(v) => setForm("show_graduate_date", v)} />
                  <ToggleSwitch label="Mostrar Mención de Grado" checked={!!form.show_mention_undergraduate} onChange={(v) => setForm("show_mention_undergraduate", v)} />
                  <p class="text-[10px] text-gray-400 italic pt-2 mt-2 leading-tight">
                    * Nota Gremial: Sus postgrados registrados siempre serán públicos siempre y cuando usted se encuentre solvente.
                  </p>
                </div>
              </div>
            </section>

            {/* BOTÓN FLOTANTE GUARDAR (Sticky) */}
            <div class="sticky bottom-6 z-50 flex justify-end px-2">
               <button 
                  type="submit" 
                  disabled={saving()}
                  class="bg-colpsi-yellow text-colpsi-blue px-8 py-4 md:px-12 rounded-full font-black shadow-2xl shadow-yellow-500/30 hover:scale-105 active:scale-95 transition-all disabled:opacity-70 flex items-center gap-3 border-2 border-transparent focus:border-white focus:ring-4 focus:ring-yellow-500/50"
                >
                  <Show when={saving()} fallback={<><span class="text-xl">💾</span> GUARDAR CAMBIOS</>}>
                    <svg class="animate-spin h-5 w-5 text-colpsi-blue" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                    GUARDANDO...
                  </Show>
               </button>
            </div>
          </form>

          {/* ======================================================== */}
          {/* SECCIÓN 6: REDES SOCIALES (FORMULARIO INDEPENDIENTE)     */}
          {/* ======================================================== */}
          <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100 mt-12">
            <div class="mb-6 border-l-4 border-colpsi-yellow pl-3">
              <h2 class="text-xl font-black text-colpsi-blue leading-tight">Presencia Digital</h2>
              <p class="text-sm text-colpsi-muted mt-1">Añade enlaces a tu Instagram, LinkedIn, Web Personal, etc.</p>
            </div>
            
            {/* Lista de Redes Guardadas */}
            <Show when={profile()?.social_networks?.length > 0}>
              <div class="mb-8 space-y-3">
                <For each={profile().social_networks}>
                  {(net: any) => (
                    <div class="flex items-center justify-between bg-gray-50 hover:bg-white p-4 rounded-2xl border border-gray-100 hover:border-blue-100 transition-colors group">
                      <div class="flex items-center gap-3 overflow-hidden">
                        <span class="bg-white px-3 py-1 rounded-xl text-xs font-black text-colpsi-blue shadow-sm border border-gray-100">{net.name}</span>
                        <a href={net.url} target="_blank" rel="noopener noreferrer" class="text-sm text-colpsi-muted hover:text-colpsi-blue truncate max-w-[150px] sm:max-w-md transition-colors">{net.url}</a>
                      </div>
                      <button onClick={() => handleDeleteSocial(net.id)} class="text-gray-400 hover:text-red-500 hover:bg-red-50 p-2 rounded-xl transition-colors" title="Eliminar enlace">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                          <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
                        </svg>
                      </button>
                    </div>
                  )}
                </For>
              </div>
            </Show>

            {/* Componente para Agregar Red (POST) */}
            <form onSubmit={handleAddSocial} class="bg-blue-50/50 p-6 md:p-8 rounded-[2rem] border border-blue-100 shadow-inner">
              <h3 class="text-sm font-bold text-colpsi-blue mb-4 uppercase tracking-widest">Vincular Nueva Cuenta</h3>
              <div class="flex flex-col md:flex-row gap-4">
                <input 
                  type="text" 
                  placeholder="Red (Ej: Instagram, LinkedIn)" 
                  required
                  value={socialForm.name} 
                  onInput={(e) => setSocialForm("name", e.currentTarget.value)} 
                  class="flex-1 bg-white border-2 border-transparent focus:border-colpsi-blue rounded-xl px-5 py-3 outline-none text-sm text-colpsi-text shadow-sm transition-all" 
                />
                <input 
                  type="url" 
                  placeholder="Enlace completo (Ej: https://instagram.com/tu_perfil)" 
                  required
                  value={socialForm.url} 
                  onInput={(e) => setSocialForm("url", e.currentTarget.value)} 
                  class="flex-[2] bg-white border-2 border-transparent focus:border-colpsi-blue rounded-xl px-5 py-3 outline-none text-sm text-colpsi-text shadow-sm transition-all" 
                />
                <button 
                  type="submit" 
                  disabled={savingSocial()}
                  class="bg-colpsi-blue text-white px-8 py-3 rounded-xl font-bold hover:bg-blue-800 active:scale-95 transition-all disabled:opacity-50 shadow-md flex justify-center min-w-[120px]"
                >
                  {savingSocial() ? "..." : "AÑADIR"}
                </button>
              </div>
            </form>
          </section>

        </Suspense>
      </div>
    </main>
  );
}

