// web/src/routes/admin/psicologos/[id]/editar.tsx

import {
  createResource,
  createEffect,
  Suspense,
  createSignal,
  Show,
} from "solid-js";
import { createStore, unwrap } from "solid-js/store";
import { useParams, action, useAction } from "@solidjs/router";
import { apiGet } from "~/lib/api";

import {
  EditPageHeader,
  EditAlert,
  SocialNetworksBlock,
  AuditBlock,
  AccountSection,
  AdminStatusSection,
  LegalIdentitySection,
  ContactVisibilitySection,
  LocationSection,
  ProfessionalSection,
  AcademicSection,
} from "~/components/admin/psicologos/edit";
import type { EditFormState } from "~/components/admin/psicologos/edit";
import { SolvenciesSection } from "~/components/admin/psicologos/edit/SolvenciesSection";

// ─── Server Actions ───────────────────────────────────────────────────────────

const updateAdminPsiServer = action(
  async (params: { id: string; payload: any }) => {
    "use server";
    const { apiPatch } = await import("~/lib/api");
    const cleanPayload = { ...params.payload };

    // Convertimos strings vacíos a null para que Go los procese correctamente
    Object.keys(cleanPayload).forEach((key) => {
      if (cleanPayload[key] === "") cleanPayload[key] = null;
    });

    return await apiPatch(`/admin/psi/${params.id}`, cleanPayload);
  },
);

const addSocialServer = action(
  async (params: { id: string; payload: { name: string; url: string } }) => {
    "use server";
    const { apiPost } = await import("~/lib/api");
    return await apiPost(`/admin/psi/${params.id}/social`, params.payload);
  },
);

const deleteSocialServer = action(
  async (params: { psiId: string; socialId: string }) => {
    "use server";
    const { apiDelete } = await import("~/lib/api");
    return await apiDelete(
      `/admin/psi/${params.psiId}/social/${params.socialId}`,
    );
  },
);

// ─── Helpers ──────────────────────────────────────────────────────────────────

const formatDate = (dateStr?: string) => (dateStr ? dateStr.split("T")[0] : "");

// Campos que el backend de Go espera recibir como "1" o "0" (multipart-form)
const RAW_BOOL_FIELDS = [
  "show_contact_email",
  "show_public_service_address",
  // Privacidad Carabobo
  "show_municipality_carabobo",
  "show_phone_carabobo",
  "show_cel_phone_carabobo",
  // Privacidad Fuera de Carabobo
  "show_state_outside",
  "show_municipality_outside_carabobo",
  "show_phone_outside_carabobo",
  "show_cel_phone_outside_carabobo",
  "show_public_service_address_outside_carabobo",
  // Privacidad Exterior
  "show_phone_outside_venezuela",
  "show_cel_phone_outside_venezuela",
  "show_public_service_address_outside_venezuela",
  // Privacidad Académica
  "show_university_undergraduate",
  "show_graduate_date",
  "show_mention_undergraduate",
];

// ─── Page Component ───────────────────────────────────────────────────────────

export default function AdminEditPsiPage() {
  const params = useParams();

  const runUpdateAction = useAction(updateAdminPsiServer);
  const runAddSocial = useAction(addSocialServer);
  const runDeleteSocial = useAction(deleteSocialServer);

  const [profile, { refetch }] = createResource(() =>
    apiGet<any>(`/admin/psi/${params.id}`),
  );
  const [workAreas] = createResource(() => apiGet<any[]>("/specialties"));

  const [form, setForm] = createStore<EditFormState>({} as EditFormState);
  const [saving, setSaving] = createSignal(false);
  const [message, setMessage] = createSignal<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const SITE_URL = import.meta.env.VITE_SITE_URL || "http://localhost:3000";

  const canonicalUrl = () => {
    const p = profile();
    if (!p) return "";
    return `${SITE_URL}/directorio/${p.first_name}-${p.last_name}-fpv${p.fpv}`;
  };

  // ── Sync DB → Store (Mapeo completo al nuevo modelo) ───────────────────────
  createEffect(() => {
    const p = profile();
    if (!p) return;

    setForm({
      // 1. Identidad y Acceso
      username: p.username ?? "",
      email: p.email ?? "",
      first_name: p.first_name ?? "",
      second_name: p.second_name ?? "",
      last_name: p.last_name ?? "",
      second_last_name: p.second_last_name ?? "",
      ci: p.ci ?? "",
      fpv: p.fpv ?? "",
      nationality: p.nationality ?? "V",
      genre: p.genre ?? "M",
      born_date: formatDate(p.born_date),

      // 2. Estatus Administrativo
      is_active: p.is_active ?? true,
      solvent: p.solvent ?? false,
      proof_of_life: p.proof_of_life ?? false,

      // 3. Contacto y Privacidad General
      contact_email: p.contact_email ?? "",
      show_contact_email: p.show_contact_email ?? false,
      contact_phone: p.contact_phone ?? "",
      contact_cell_phone: p.contact_cell_phone ?? "",
      service_address: p.service_address ?? "",
      show_public_service_address: p.show_public_service_address ?? false,

      // 4. Ubicación: Carabobo
      municipality_carabobo: p.municipality_carabobo ?? "",
      show_municipality_carabobo: p.show_municipality_carabobo ?? false,
      phone_carabobo: p.phone_carabobo ?? "",
      show_phone_carabobo: p.show_phone_carabobo ?? false,
      cel_phone_carabobo: p.cel_phone_carabobo ?? "",
      show_cel_phone_carabobo: p.show_cel_phone_carabobo ?? false,

      // 5. Ubicación: Fuera de Carabobo (Venezuela)
      state_outside: p.state_outside ?? "",
      show_state_outside: p.show_state_outside ?? false,
      municipality_outside_carabobo: p.municipality_outside_carabobo ?? "",
      show_municipality_outside_carabobo:
        p.show_municipality_outside_carabobo ?? false,
      phone_outside_carabobo: p.phone_outside_carabobo ?? "",
      show_phone_outside_carabobo: p.show_phone_outside_carabobo ?? false,
      cel_phone_outside_carabobo: p.cel_phone_outside_carabobo ?? "",
      show_cel_phone_outside_carabobo:
        p.show_cel_phone_outside_carabobo ?? false,
      service_address_outside_carabobo:
        p.service_address_outside_carabobo ?? "",
      show_public_service_address_outside_carabobo:
        p.show_public_service_address_outside_carabobo ?? false,

      // 6. Ubicación: Exterior
      country: p.country ?? "",
      phone_outside_venezuela: p.phone_outside_venezuela ?? "",
      show_phone_outside_venezuela: p.show_phone_outside_venezuela ?? false,
      cell_phone_outside_venezuela: p.cell_phone_outside_venezuela ?? "",
      show_cel_phone_outside_venezuela:
        p.show_cel_phone_outside_venezuela ?? false,
      service_address_outside_venezuela:
        p.service_address_outside_venezuela ?? "",
      show_public_service_address_outside_venezuela:
        p.show_public_service_address_outside_venezuela ?? false,

      // 7. Profesional (Áreas de Desempeño)
      primary_work_area: p.primary_work_area ?? "",
      secondary_work_area: p.secondary_work_area ?? "",
      mini_bio: p.mini_bio ?? "",
      full_bio: p.full_bio?.content ?? "",

      // 8. Datos Académicos
      guild_inscription_date: formatDate(p.col_data?.guild_inscription_date),
      university_undergraduate: p.col_data?.university_undergraduate ?? "",
      graduate_date: formatDate(p.col_data?.graduate_date),
      mention_undergraduate: p.col_data?.mention_undergraduate ?? "",
      register_number: p.col_data?.register_number ?? "",
      register_title_state: p.col_data?.register_title_state ?? "",
      register_title_date: formatDate(p.col_data?.register_title_date),
      register_folio: p.col_data?.register_folio ?? "",
      register_tome: p.col_data?.register_tome ?? "",

      // 9. Privacidad Académica
      show_university_undergraduate:
        p.col_data?.show_university_undergraduate ?? false,
      show_graduate_date: p.col_data?.show_graduate_date ?? false,
      show_mention_undergraduate:
        p.col_data?.show_mention_undergraduate ?? false,

      // 10. Banderas Institucionales
      guild_director: p.col_data?.guild_director ?? false,
      sixty_five_or_plus: p.col_data?.sixty_five_or_plus ?? false,
      guild_collaborator: p.col_data?.guild_collaborator ?? false,
      public_employee: p.col_data?.public_employee ?? false,
      discapacity: p.col_data?.discapacity ?? false,
      university_professor: p.col_data?.university_professor ?? false,
      double_guild: p.col_data?.double_guild ?? false,
      double_guild_location: p.col_data?.double_guild_location ?? "",
      date_of_last_solvency: formatDate(p.col_data?.date_of_last_solvency),

      solvencies: p.solvencies ?? [],
    });
  });

  // ── Submit ───────────────────────────────────────────────────────────────────
  const handleSave = async (e: Event) => {
    e.preventDefault();
    if (saving()) return;
    setSaving(true);
    setMessage(null);

    const rawForm = unwrap(form);
    const payload: Record<string, any> = {};

    // Saneamiento del payload para la API de Go
    for (const [key, value] of Object.entries(rawForm)) {
      if (RAW_BOOL_FIELDS.includes(key)) {
        payload[key] = value === true ? "1" : "0";
      } else if (value === "") {
        payload[key] = null;
      } else {
        payload[key] = value;
      }
    }

    // Conversión de tipos numéricos
    payload.ci = parseInt(String(rawForm.ci)) || null;
    payload.fpv = parseInt(String(rawForm.fpv)) || null;
    payload.register_number = parseInt(String(rawForm.register_number)) || null;

    try {
      await runUpdateAction({ id: params.id ?? "", payload });
      setMessage({
        type: "success",
        text: "Expediente del psicólogo actualizado correctamente.",
      });
      refetch();
      window.scrollTo({ top: 0, behavior: "smooth" });
    } catch (err: any) {
      const msg = err?.message || String(err);
      setMessage({
        type: "error",
        text:
          msg.replace(/^.*?ApiError:\s*/i, "") ||
          "Error en el servidor al guardar.",
      });
      window.scrollTo({ top: 0, behavior: "smooth" });
    } finally {
      setSaving(false);
    }
  };

  const set = (key: keyof EditFormState, value: any) =>
    setForm(key as any, value);

  return (
    <main class="pb-28 animate-in fade-in duration-500 font-sans">
      <EditPageHeader profile={profile()} />

      <Suspense
        fallback={
          <div class="h-96 bg-white animate-pulse rounded-[2.5rem] shadow-sm border border-gray-100" />
        }
      >
        <EditAlert message={message()} />

        <form onSubmit={handleSave} class="space-y-8">
          <AccountSection form={form} setForm={set} url={canonicalUrl()} />
          <AdminStatusSection form={form} setForm={set} />

          {/* Historial de Solvencias */}
          <SolvenciesSection
            solvencies={form.solvencies}
            onAddLocalSolvency={(year) => {
              const newSolv = { date: `${year}-12-31T00:00:00Z` };
              set("solvencies", [...form.solvencies, newSolv]);
            }}
          />

          <LegalIdentitySection form={form} setForm={set} />
          <ContactVisibilitySection form={form} setForm={set} />
          <LocationSection form={form} setForm={set} />

          {/* Se pasa 'workAreas' siguiendo la nueva terminología */}
          <ProfessionalSection
            form={form}
            setForm={set}
            workAreas={workAreas()}
          />

          <AcademicSection form={form} setForm={set} />

          {/* Botón de guardado flotante */}
          <div class="sticky bottom-10 z-50 flex justify-end max-w-5xl mx-auto px-4">
            <button
              type="submit"
              disabled={saving()}
              class="bg-blue-900 text-white px-12 py-5 rounded-2xl font-black shadow-2xl hover:bg-blue-800 active:scale-95 transition-all disabled:opacity-50 flex items-center gap-3 border-2 border-white/20 uppercase tracking-tight"
            >
              <Show
                when={saving()}
                fallback={<span>💾 Guardar Expediente Maestro</span>}
              >
                <div class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                <span>Procesando...</span>
              </Show>
            </button>
          </div>
        </form>

        <SocialNetworksBlock
          profile={profile()}
          onAdd={async (p) => {
            // Aseguramos que params.id sea string para TypeScript
            const id = params.id ?? "";
            if (!id) return;

            await runAddSocial({ id: id, payload: p });
            refetch();
          }}
          onDelete={async (sid) => {
            const id = params.id ?? "";
            if (!id) return;

            if (confirm("¿Eliminar esta red social?")) {
              await runDeleteSocial({ psiId: id, socialId: sid });
              refetch();
            }
          }}
        />

        <AuditBlock profile={profile()} />
      </Suspense>
    </main>
  );
}
