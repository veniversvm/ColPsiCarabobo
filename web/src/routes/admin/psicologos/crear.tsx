import { createSignal, Show } from "solid-js";
import { createStore } from "solid-js/store";
import { A, useNavigate, action, useAction } from "@solidjs/router";
import { ApiError } from "~/lib/api";
import { ToggleSwitch } from "~/components/ui/ToggleSwitch";
import { isStrongPassword } from "~/lib/sanitizer";
import { PasswordInputComponent } from "~/components/ui/PasswordInput";

const createPsiServer = action(async (payload: any) => {
  "use server";
  const { apiPost } = await import("~/lib/api");
  // Se lanza a la capa Go, donde GORM validará restricciones únicas (CI, FPV, Email)
  return await apiPost("/admin/psi/create", payload);
});

export default function CreatePsychologistPage() {
  const navigate = useNavigate();
  const runCreateAction = useAction(createPsiServer);

  const [saving, setSaving] = createSignal(false);
  const [message, setMessage] = createSignal<{ type: "success" | "error"; text: string; details?: any } | null>(null);

  // Generamos la fecha de hoy en formato YYYY-MM-DD para el input type="date"
  const today = new Date().toISOString().split('T')[0];

  // Store alineado con el DTO. Valores iniciales en blanco para obligar al llenado
  const [form, setForm] = createStore({
    username: "", email: "", password: "",
    first_name: "", second_name: "", last_name: "", second_last_name: "",
    ci: "", fpv: "", nationality: "V", genre: "M", born_date: "",
    
    is_active: true, solvent: true, proof_of_life: true,
    
    public_phone: "", service_address: "",
    
    municipality_carabobo: "", phone_carabobo: "", cel_phone_carabobo: "",
    state_outside: "", municipality_outside_carabobo: "", phone_outside_carabobo: "", cel_phone_outside_carabobo: "",
    
    primary_specialty: "", secondary_specialty: "",
    
    university_undergraduate: "", graduate_date: "", mention_undergraduate: "",
    register_number: "", register_title_state: "", register_title_date: "", register_folio: "", register_tome: "",
    
    guild_director: false, sixty_five_or_plus: false, guild_collaborator: false, 
    public_employee: false, university_professor: false, double_guild: false, cpsm: false,
    date_of_last_solvency: today, // Se toma el día de hoy automáticamente
  });

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    
    // 1. Validaciones Críticas en Frontend
    if (!isStrongPassword(form.password)) {
      setMessage({ type: "error", text: "La contraseña no cumple con los requisitos de seguridad (Mín. 8 caracteres, mayúscula, minúscula, número y símbolo especial)." });
      window.scrollTo({ top: 0, behavior: 'smooth' });
      return;
    }

    setSaving(true);
    setMessage(null);

    // 2. Formateo de Payload para el tipado estricto de Go
    const payload = {
      ...form,
      ci: parseInt(form.ci) || 0,
      fpv: parseInt(form.fpv) || 0,
      register_number: parseInt(form.register_number) || 0,
    };

    try {
      await runCreateAction(payload);
      setMessage({ type: "success", text: "Psicólogo registrado exitosamente. Volviendo al listado..." });
      
      setTimeout(() => {
        navigate("/admin/psicologos");
      }, 2000);

    } catch (err: any) {
      console.log("---- ERR: ", err);

      // 1. Detectamos si el error contiene la marca de nuestra clase ApiError
      const errorString = err?.message || String(err);
      const isApiError = errorString.includes("ApiError") || err?.name === "ApiError";

      if (isApiError) {
        // 2. Limpiamos el mensaje: quitamos "ApiError: " del principio si existe
        // Esto transforma "ApiError: el número de FPV..." en "el número de FPV..."
        const cleanMessage = errorString.replace(/^.*?ApiError:\s*/i, "");

        setMessage({ 
          type: "error", 
          text: cleanMessage,
        //   details: err.data 
        });
      } else {
        // 3. Fallback para errores reales de red o 500 sin formato
        setMessage({ 
          type: "error", 
          text: "Error crítico de conexión o el servidor está caído." 
        });
      }
      
      window.scrollTo({ top: 0, behavior: 'smooth' });
    } finally {
      setSaving(false);
    }
  };

  // Helper de UI para campos obligatorios
  const inputClass = "w-full bg-gray-50 border-2 border-transparent focus:border-colpsi-blue rounded-xl px-4 py-2.5 outline-none";

  return (
    <div class="space-y-6 animate-in fade-in duration-500 pb-24 max-w-5xl mx-auto">
      
      {/* HEADER */}
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-black text-colpsi-blue">Alta de Colegiado</h1>
          <p class="text-gray-500 text-sm mt-1">Apertura de nuevo expediente institucional (* Campos obligatorios)</p>
        </div>
        <A href="/admin/psicologos" class="bg-white border-2 border-gray-200 text-gray-700 px-4 py-2.5 rounded-xl font-bold hover:bg-gray-50 transition-colors text-sm">
          Cancelar
        </A>
      </div>

      {/* ALERTAS DE ESTADO Y ERRORES DEL BACKEND */}
      <Show when={message()}>
        <div class={`p-5 rounded-2xl shadow-md border-l-4 animate-in slide-in-from-top-4 ${message()?.type === 'success' ? 'bg-green-50 border-green-500' : 'bg-red-50 border-red-500'}`}>
          <div class="flex items-start gap-3">
            <span class="text-2xl">{message()?.type === 'success' ? '✅' : '⚠️'}</span>
            <div>
              <p class={`font-black uppercase tracking-wide text-xs ${message()?.type === 'success' ? 'text-green-800' : 'text-red-800'}`}>
                {message()?.type === 'success' ? 'Operación Exitosa' : 'Alerta del Sistema'}
              </p>
              <p class={`text-sm font-medium mt-1 ${message()?.type === 'success' ? 'text-green-700' : 'text-red-700'}`}>
                {message()?.text}
              </p>
              {/* Si Go manda una lista de errores (Fiber Validation Errors) */}
              <Show when={message()?.details?.error}>
                 <pre class="mt-3 p-3 bg-red-100/50 rounded-lg text-xs font-mono text-red-900 overflow-x-auto">
                   {JSON.stringify(message()?.details, null, 2)}
                 </pre>
              </Show>
            </div>
          </div>
        </div>
      </Show>

      {/* FORMULARIO MAESTRO */}
      <form onSubmit={handleSubmit} class="space-y-8">
        
        {/* BLOQUE 1: CUENTA Y ACCESO */}
        <section class="bg-white rounded-[2rem] p-6 shadow-sm border border-gray-100">
          <h2 class="text-lg font-black text-colpsi-blue mb-4 border-l-4 border-colpsi-yellow pl-3">Cuenta y Acceso</h2>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div class="space-y-1">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Usuario <span class="text-red-500">*</span></label>
              <input type="text" required value={form.username} onInput={(e) => setForm("username", e.currentTarget.value)} class={inputClass} />
            </div>
            <div class="space-y-1">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Email <span class="text-red-500">*</span></label>
              <input type="email" required value={form.email} onInput={(e) => setForm("email", e.currentTarget.value)} class={inputClass} />
            </div>
            <div class="space-y-1">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Contraseña Inicial <span class="text-red-500">*</span></label>
              <PasswordInputComponent required value={form.password} onInput={(e: any) => setForm("password", e.currentTarget.value)} class={inputClass} />
            </div>
          </div>
        </section>

        {/* BLOQUE 2: IDENTIDAD LEGAL */}
        <section class="bg-white rounded-[2rem] p-6 shadow-sm border border-gray-100">
          <h2 class="text-lg font-black text-colpsi-blue mb-4 border-l-4 border-colpsi-yellow pl-3">Identidad Legal</h2>
          <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div class="space-y-1 md:col-span-2">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Primer Nombre <span class="text-red-500">*</span></label>
              <input type="text" required value={form.first_name} onInput={(e) => setForm("first_name", e.currentTarget.value)} class={inputClass} />
            </div>
            {/* Opcional */}
            <div class="space-y-1 md:col-span-2">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Segundo Nombre</label>
              <input type="text" value={form.second_name} onInput={(e) => setForm("second_name", e.currentTarget.value)} class={inputClass} />
            </div>
            <div class="space-y-1 md:col-span-2">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Primer Apellido <span class="text-red-500">*</span></label>
              <input type="text" required value={form.last_name} onInput={(e) => setForm("last_name", e.currentTarget.value)} class={inputClass} />
            </div>
            {/* Opcional */}
            <div class="space-y-1 md:col-span-2">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Segundo Apellido</label>
              <input type="text" value={form.second_last_name} onInput={(e) => setForm("second_last_name", e.currentTarget.value)} class={inputClass} />
            </div>

            <div class="space-y-1">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Nacionalidad <span class="text-red-500">*</span></label>
              <select required value={form.nationality} onChange={(e) => setForm("nationality", e.currentTarget.value)} class={inputClass}>
                <option value="V">V - Venezolano</option>
                <option value="E">E - Extranjero</option>
              </select>
            </div>
            <div class="space-y-1">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Cédula <span class="text-red-500">*</span></label>
              <input type="number" required value={form.ci} onInput={(e) => setForm("ci", e.currentTarget.value)} class={inputClass} />
            </div>
            <div class="space-y-1">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Nro. FPV <span class="text-red-500">*</span></label>
              <input type="number" required value={form.fpv} onInput={(e) => setForm("fpv", e.currentTarget.value)} class="w-full bg-colpsi-yellow/20 border-2 border-transparent focus:border-colpsi-yellow rounded-xl px-4 py-2.5 outline-none font-bold text-colpsi-blue" />
            </div>
            <div class="space-y-1">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Género <span class="text-red-500">*</span></label>
              <select required value={form.genre} onChange={(e) => setForm("genre", e.currentTarget.value)} class={inputClass}>
                <option value="M">Masculino</option>
                <option value="F">Femenino</option>
              </select>
            </div>
            <div class="space-y-1 md:col-span-2">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Fecha de Nacimiento <span class="text-red-500">*</span></label>
              <input type="date" required value={form.born_date} onInput={(e) => setForm("born_date", e.currentTarget.value)} class={inputClass} />
            </div>
          </div>
        </section>

        {/* BLOQUE 3: DATOS COLEGIALES */}
        <section class="bg-white rounded-[2rem] p-6 shadow-sm border border-gray-100">
          <h2 class="text-lg font-black text-colpsi-blue mb-4 border-l-4 border-colpsi-yellow pl-3">Registro Académico Base</h2>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div class="space-y-1 md:col-span-2">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Universidad de Egreso <span class="text-red-500">*</span></label>
              <input type="text" required value={form.university_undergraduate} onInput={(e) => setForm("university_undergraduate", e.currentTarget.value)} class={inputClass} />
            </div>
            <div class="space-y-1">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Fecha de Egreso <span class="text-red-500">*</span></label>
              <input type="date" required value={form.graduate_date} onInput={(e) => setForm("graduate_date", e.currentTarget.value)} class={inputClass} />
            </div>
            {/* Opcional */}
            <div class="space-y-1 md:col-span-3">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Mención</label>
              <input type="text" value={form.mention_undergraduate} onInput={(e) => setForm("mention_undergraduate", e.currentTarget.value)} class={inputClass} />
            </div>

            <div class="col-span-full mt-4"><hr class="border-gray-100"/></div>
            
            <div class="space-y-1">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Estado de Registro <span class="text-red-500">*</span></label>
              <input type="text" required value={form.register_title_state} onInput={(e) => setForm("register_title_state", e.currentTarget.value)} class={inputClass} />
            </div>
            <div class="space-y-1">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Fecha de Registro <span class="text-red-500">*</span></label>
              <input type="date" required value={form.register_title_date} onInput={(e) => setForm("register_title_date", e.currentTarget.value)} class={inputClass} />
            </div>
            <div class="space-y-1">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Número de Registro <span class="text-red-500">*</span></label>
              <input type="number" required value={form.register_number} onInput={(e) => setForm("register_number", e.currentTarget.value)} class={inputClass} />
            </div>
            <div class="space-y-1">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Folio <span class="text-red-500">*</span></label>
              <input type="text" required value={form.register_folio} onInput={(e) => setForm("register_folio", e.currentTarget.value)} class={inputClass} />
            </div>
            <div class="space-y-1">
              <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Tomo <span class="text-red-500">*</span></label>
              <input type="text" required value={form.register_tome} onInput={(e) => setForm("register_tome", e.currentTarget.value)} class={inputClass} />
            </div>
          </div>
        </section>

        {/* BLOQUE 4: DATOS DE CONTACTO */}
        <section class="bg-white rounded-[2rem] p-6 shadow-sm border border-gray-100">
           <h2 class="text-lg font-black text-colpsi-blue mb-4 border-l-4 border-gray-300 pl-3">Datos de Contacto</h2>
           <p class="text-xs text-gray-500 mb-4 bg-gray-50 p-3 rounded-xl border border-gray-100">Información requerida para mantener comunicación oficial con el agremiado.</p>
           <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="space-y-1 md:col-span-2">
                <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Teléfono Fijo <span class="text-red-500">*</span></label>
                <input type="tel" required value={form.public_phone} onInput={(e) => setForm("public_phone", e.currentTarget.value)} class={inputClass} />
              </div>
              <div class="space-y-1 md:col-span-2">
                <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Dirección Exacta <span class="text-red-500">*</span></label>
                <input type="text" required value={form.service_address} onInput={(e) => setForm("service_address", e.currentTarget.value)} class={inputClass} />
              </div>
           </div>
        </section>

        {/* BLOQUE 5: ESTATUS Y BANDERAS INSTITUCIONALES */}
        <section class="bg-white rounded-[2rem] p-6 shadow-sm border border-gray-100">
          <h2 class="text-lg font-black text-colpsi-blue mb-4 border-l-4 border-colpsi-yellow pl-3">Estatus Institucional</h2>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
            <div class="bg-gray-50 p-5 rounded-2xl space-y-3 border border-gray-200">
              <h3 class="text-xs font-bold text-gray-500 uppercase tracking-widest border-b pb-2">Estado Principal</h3>
              <ToggleSwitch label="Cuenta Activa (Acceso al sistema)" checked={form.is_active} onChange={(v) => setForm("is_active", v)} />
              <ToggleSwitch label="Miembro Solvente (Al día con pagos)" checked={form.solvent} onChange={(v) => setForm("solvent", v)} />
              <ToggleSwitch label="Fe de Vida Activa" checked={form.proof_of_life} onChange={(v) => setForm("proof_of_life", v)} />
              
              <div class="pt-2">
                <label class="text-[10px] font-black text-gray-400 uppercase ml-1">Fecha Última Solvencia <span class="text-red-500">*</span></label>
                <input type="date" required value={form.date_of_last_solvency} onInput={(e) => setForm("date_of_last_solvency", e.currentTarget.value)} class={inputClass} />
              </div>
            </div>

            <div class="bg-blue-50/50 p-5 rounded-2xl space-y-3 border border-blue-100">
              <h3 class="text-xs font-bold text-colpsi-blue uppercase tracking-widest border-b border-blue-100 pb-2">Roles Gremiales</h3>
              <ToggleSwitch label="Director del Gremio" checked={form.guild_director} onChange={(v) => setForm("guild_director", v)} />
              <ToggleSwitch label="Colaborador del Gremio" checked={form.guild_collaborator} onChange={(v) => setForm("guild_collaborator", v)} />
              <ToggleSwitch label="Profesor Universitario" checked={form.university_professor} onChange={(v) => setForm("university_professor", v)} />
              <ToggleSwitch label="Empleado Público" checked={form.public_employee} onChange={(v) => setForm("public_employee", v)} />
              <ToggleSwitch label="Doble Gremio" checked={form.double_guild} onChange={(v) => setForm("double_guild", v)} />
              <ToggleSwitch label="Beneficio 65+ Años" checked={form.sixty_five_or_plus} onChange={(v) => setForm("sixty_five_or_plus", v)} />
            </div>
          </div>
        </section>

        {/* BOTÓN FLOTANTE */}
        <div class="sticky bottom-6 z-50 flex justify-end">
          <button 
            type="submit" 
            disabled={saving()}
            class="bg-colpsi-blue text-white px-10 py-4 rounded-2xl font-black shadow-2xl hover:scale-105 active:scale-95 transition-all disabled:opacity-70 flex items-center gap-3 border-2 border-white"
          >
            {saving() ? "PROCESANDO EN SERVIDOR..." : "✔️ REGISTRAR EXPEDIENTE"}
          </button>
        </div>

      </form>
    </div>
  );
}