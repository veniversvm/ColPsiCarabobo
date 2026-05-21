// web/src/routes/admin/psicologos/crear.tsx
import { createSignal } from "solid-js";
import { createStore } from "solid-js/store";
import { useNavigate, action, useAction } from "@solidjs/router";
import { isStrongPassword } from "~/lib/sanitizer";
import {
  FormHeader,
  FormMessage,
  FormActions,
  AccountSection,
  LegalIdentitySection,
  AcademicRegistrationSection,
  ContactSection,
  InstitutionalStatusSection
} from "~/components/admin/psicologos/create";
import { PsicologoForm } from "~/types/admin";

const createPsiServer = action(async (payload: any, idempotencyKey: string) => {
  "use server";
  const { apiPost } = await import("~/lib/api");
  // El endpoint ahora espera el nuevo esquema (contact_phone, work_areas, etc.)
  return await apiPost("/admin/psi/create", payload, {
    headers: { "X-Idempotency-Key": idempotencyKey },
  });
});

function generateIdempotencyKey(): string {
  return crypto.randomUUID();
}

export default function CreatePsychologistPage() {
  const navigate        = useNavigate();
  const runCreateAction = useAction(createPsiServer);

  const [saving,  setSaving]  = createSignal(false);
  const [message, setMessage] = createSignal<{ type: "success" | "error"; text: string; details?: any } | null>(null);
  const [idempotencyKey] = createSignal(generateIdempotencyKey());

  const today = new Date().toISOString().split('T')[0];

  // ── ESTADO INICIAL ACTUALIZADO AL MODELO 2026 ──────────────────────────
  const [form, setForm] = createStore<PsicologoForm>({
    // Auth
    username: "", email: "", password: "",
    
    // Identidad
    first_name: "", second_name: "", last_name: "", second_last_name: "",
    ci: "", fpv: "", nationality: "V", genre: "M", born_date: "",
    
    // Estatus
    is_active: true, solvent: true, proof_of_life: true,
    
    // Contacto Principal
    contact_email: "", 
    contact_phone: "",      // Reemplaza a public_phone
    contact_cell_phone: "",  // Nuevo
    service_address: "",
    
    // Ubicación: Carabobo
    municipality_carabobo: "", phone_carabobo: "", cel_phone_carabobo: "",
    
    // Ubicación: Fuera de Carabobo (Venezuela)
    state_outside: "", 
    municipality_outside_carabobo: "", 
    phone_outside_carabobo: "", 
    cel_phone_outside_carabobo: "",
    service_address_outside_carabobo: "", // Nuevo

    // Ubicación: Exterior
    country: "",
    phone_outside_venezuela: "",
    cell_phone_outside_venezuela: "", // Nuevo
    service_address_outside_venezuela: "", // Nuevo
    
    // Profesional (Áreas de Desempeño)
    primary_work_area: "", 
    secondary_work_area: "",
    
    // Académico y Registro
    guild_inscription_date: today, // Nuevo
    university_undergraduate: "", graduate_date: "", mention_undergraduate: "",
    register_number: "", register_title_state: "", register_title_date: "", register_folio: "", register_tome: "",
    
    // Flags Gremiales
    guild_director: false, 
    sixty_five_or_plus: false, 
    guild_collaborator: false, 
    public_employee: false, 
    discapacity: false,      // Nuevo
    university_professor: false, 
    double_guild: false, 
    double_guild_location: "", // Nuevo
    cpsm: false,
    date_of_last_solvency: today,
  });

  const updateField = <K extends keyof PsicologoForm>(field: K, value: PsicologoForm[K]) => {
    setForm(field, value);
  };

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    if (saving()) return;
    
    if (!isStrongPassword(form.password || "")) {
      setMessage({ type: "error", text: "Seguridad insuficiente: La contraseña requiere mayúscula, número y símbolo." });
      window.scrollTo({ top: 0, behavior: "smooth" });
      return;
    }

    setSaving(true);
    setMessage(null);

    // Tipado correcto para el envío
    const payload = {
      ...form,
      ci:              parseInt(String(form.ci))              || 0,
      fpv:             parseInt(String(form.fpv))             || 0,
      register_number: parseInt(String(form.register_number)) || 0,
    };

    try {
      await runCreateAction(payload, idempotencyKey());
      setMessage({ type: "success", text: "Agremiado creado con éxito. Redirigiendo..." });
      setTimeout(() => navigate("/admin/psicologos"), 1500);

    } catch (err: any) {
      const errorString = err?.message || String(err);
      setMessage({ type: "error", text: errorString.includes("ApiError") ? errorString.replace(/.*ApiError:\s*/i, "") : "Error de red o servidor." });
      window.scrollTo({ top: 0, behavior: "smooth" });
    } finally {
      setSaving(false);
    }
  };

  return (
    <div class="space-y-6 animate-in fade-in duration-500 pb-24 max-w-5xl mx-auto">
      <FormHeader />
      <FormMessage type={message()?.type} text={message()?.text || ""} details={message()?.details} />
      
      <form onSubmit={handleSubmit} class="space-y-8">
        {/* Cada sección debe actualizarse para manejar los nuevos nombres de campos */}
        <AccountSection            form={form} setForm={updateField} />
        <LegalIdentitySection      form={form} setForm={updateField} />
        <AcademicRegistrationSection form={form} setForm={updateField} />
        <ContactSection            form={form} setForm={updateField} />
        <InstitutionalStatusSection form={form} setForm={updateField} today={today} />
        
        <FormActions saving={saving()} />
      </form>
    </div>
  );
}