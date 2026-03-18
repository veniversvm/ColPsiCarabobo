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
  return await apiPost("/admin/psi/create", payload, {
    headers: { "X-Idempotency-Key": idempotencyKey },
  });
});

// Genera una key única por intento: adminId se incluye en el servidor,
// aquí usamos crypto.randomUUID() como nonce del lado cliente.
function generateIdempotencyKey(): string {
  return crypto.randomUUID();
}

export default function CreatePsychologistPage() {
  const navigate        = useNavigate();
  const runCreateAction = useAction(createPsiServer);

  const [saving,  setSaving]  = createSignal(false);
  const [message, setMessage] = createSignal<{ type: "success" | "error"; text: string; details?: any } | null>(null);

  // Una key por montaje del componente — se regenera si el admin navega
  // fuera y vuelve, pero se mantiene si solo reintenta el mismo form.
  const [idempotencyKey] = createSignal(generateIdempotencyKey());

  const today = new Date().toISOString().split('T')[0];

  const [form, setForm] = createStore<PsicologoForm>({
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
    date_of_last_solvency: today,
  });

  const updateField = <K extends keyof PsicologoForm>(field: K, value: PsicologoForm[K]) => {
    setForm(field, value);
  };

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    if (saving()) return;
    
    if (!isStrongPassword(form.password)) {
      setMessage({ type: "error", text: "La contraseña no cumple con los requisitos de seguridad (Mín. 8 caracteres, mayúscula, minúscula, número y símbolo especial)." });
      window.scrollTo({ top: 0, behavior: "smooth" });
      return;
    }

    setSaving(true);
    setMessage(null);

    const payload = {
      ...form,
      ci:              parseInt(form.ci)              || 0,
      fpv:             parseInt(form.fpv)             || 0,
      register_number: parseInt(form.register_number) || 0,
    };

    try {
      await runCreateAction(payload, idempotencyKey());
      setMessage({ type: "success", text: "Psicólogo registrado exitosamente. Volviendo al listado..." });
      setTimeout(() => navigate("/admin/psicologos"), 2000);

    } catch (err: any) {
      const errorString = err?.message || String(err);
      const isApiError  = errorString.includes("ApiError") || err?.name === "ApiError";

      if (isApiError) {
        setMessage({ type: "error", text: errorString.replace(/^.*?ApiError:\s*/i, "") });
      } else {
        setMessage({ type: "error", text: "Error crítico de conexión o el servidor está caído." });
      }

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