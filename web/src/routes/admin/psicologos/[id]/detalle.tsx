// web/src/routes/admin/psicologos/[id]/editar.tsx

import { createResource, createEffect, Suspense } from "solid-js";
import { createStore, unwrap } from "solid-js/store";
import { useParams, action, useAction } from "@solidjs/router";
import { createSignal } from "solid-js";
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

// ─── Server Actions ───────────────────────────────────────────────────────────

const updateAdminPsiServer = action(async (params: { id: string; payload: any }) => {
  "use server";
  const { apiPatch } = await import("~/lib/api");
  const cleanPayload = { ...params.payload };
  Object.keys(cleanPayload).forEach((key) => {
    if (cleanPayload[key] === "") cleanPayload[key] = null;
  });
  return await apiPatch(`/admin/psi/${params.id}`, cleanPayload);
});

const addSocialServer = action(async (params: { id: string; payload: { name: string; url: string } }) => {
  "use server";
  const { apiPost } = await import("~/lib/api");
  return await apiPost(`/admin/psi/${params.id}/social`, params.payload);
});

const deleteSocialServer = action(async (params: { psiId: string; socialId: string }) => {
  "use server";
  const { apiDelete } = await import("~/lib/api");
  return await apiDelete(`/admin/psi/${params.psiId}/social/${params.socialId}`);
});

// ─── Helpers ──────────────────────────────────────────────────────────────────

const formatDate = (dateStr?: string) => (dateStr ? dateStr.split("T")[0] : "");

const RAW_BOOL_FIELDS = [
  "show_contact_email",
  "show_public_phone",
  "show_public_service_address",
  "show_phone_outside_carabobo",
  "show_cel_phone_outside_carabobo",
  "show_public_service_address_outside_carabobo",
  "show_phone_outside_venezuela",
  "show_cel_phone_outside_venezuela",
  "show_public_service_address_outside_venezuela",
  "show_university_undergraduate",
  "show_graduate_date",
  "show_mention_undergraduate",
];

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function AdminEditPsiPage() {
  const params = useParams();

  const runUpdateAction = useAction(updateAdminPsiServer);
  const runAddSocial    = useAction(addSocialServer);
  const runDeleteSocial = useAction(deleteSocialServer);

  const [profile, { refetch }] = createResource(() => apiGet<any>(`/admin/psi/${params.id}`));
  const [specialties]          = createResource(() => apiGet<any[]>("/specialties"));

  const [form, setForm] = createStore<EditFormState>({} as EditFormState);
  const [saving,  setSaving]  = createSignal(false);
  const [message, setMessage] = createSignal<{ type: "success" | "error"; text: string } | null>(null);

  // ── Sync DB → Store ─────────────────────────────────────────────────────────
  createEffect(() => {
    const p = profile();
    if (!p) return;
    setForm({
      username:  p.username  ?? "",
      email:     p.email     ?? "",

      first_name:        p.first_name        ?? "",
      second_name:       p.second_name       ?? "",
      last_name:         p.last_name         ?? "",
      second_last_name:  p.second_last_name  ?? "",
      ci:                p.ci                ?? "",
      fpv:               p.fpv               ?? "",
      nationality:       p.nationality       ?? "V",
      genre:             p.genre             ?? "M",
      born_date:         formatDate(p.born_date),

      is_active:     p.is_active     ?? true,
      solvent:       p.solvent       ?? false,
      proof_of_life: p.proof_of_life ?? false,

      contact_email:               p.contact_email               ?? "",
      show_contact_email:          p.show_contact_email          ?? false,
      public_phone:                p.public_phone                ?? "",
      show_public_phone:           p.show_public_phone           ?? false,
      service_address:             p.service_address             ?? "",
      show_public_service_address: p.show_public_service_address ?? false,

      municipality_carabobo:  p.municipality_carabobo  ?? "",
      phone_carabobo:         p.phone_carabobo          ?? "",
      cel_phone_carabobo:     p.cel_phone_carabobo      ?? "",

      state_outside:                            p.state_outside                             ?? "",
      municipality_outside_carabobo:            p.municipality_outside_carabobo            ?? p.municipality_out_side_carabobo  ?? "",
      phone_outside_carabobo:                   p.phone_outside_carabobo                   ?? p.phone_out_side_carabobo         ?? "",
      cel_phone_outside_carabobo:               p.cel_phone_outside_carabobo               ?? p.cel_phone_out_side_carabobo     ?? "",
      service_address_outside_carabobo:         p.service_address_outside_carabobo         ?? "",
      show_phone_outside_carabobo:              p.show_phone_outside_carabobo              ?? false,
      show_cel_phone_outside_carabobo:          p.show_cel_phone_outside_carabobo          ?? false,
      show_public_service_address_outside_carabobo: p.show_public_service_address_outside_carabobo ?? false,

      country:                                      p.country                                       ?? "",
      phone_outside_venezuela:                      p.phone_outside_venezuela                       ?? "",
      service_address_outside_venezuela:            p.service_address_outside_venezuela             ?? "",
      show_phone_outside_venezuela:                 p.show_phone_outside_venezuela                  ?? false,
      show_cel_phone_outside_venezuela:             p.show_cel_phone_outside_venezuela              ?? false,
      show_public_service_address_outside_venezuela: p.show_public_service_address_outside_venezuela ?? false,

      primary_specialty:   p.primary_specialty   ?? "",
      secondary_specialty: p.secondary_specialty ?? "",
      mini_bio:            p.mini_bio            ?? "",
      full_bio:            p.full_bio?.content   ?? "",

      university_undergraduate: p.col_data?.university_undergraduate ?? "",
      graduate_date:            formatDate(p.col_data?.graduate_date),
      mention_undergraduate:    p.col_data?.mention_undergraduate    ?? "",
      register_number:          p.col_data?.register_number          ?? "",
      register_title_state:     p.col_data?.register_title_state     ?? "",
      register_title_date:      formatDate(p.col_data?.register_title_date),
      register_folio:           p.col_data?.register_folio           ?? "",
      register_tome:            p.col_data?.register_tome            ?? "",

      show_university_undergraduate: p.col_data?.show_university_undergraduate ?? false,
      show_graduate_date:            p.col_data?.show_graduate_date            ?? false,
      show_mention_undergraduate:    p.col_data?.show_mention_undergraduate    ?? false,

      guild_director:       p.col_data?.guild_director       ?? false,
      sixty_five_or_plus:   p.col_data?.sixty_five_or_plus   ?? false,
      guild_collaborator:   p.col_data?.guild_collaborator   ?? false,
      public_employee:      p.col_data?.public_employee      ?? false,
      university_professor: p.col_data?.university_professor ?? false,
      double_guild:         p.col_data?.double_guild         ?? false,
      cpsm:                 p.col_data?.cpsm                 ?? false,
      date_of_last_solvency: formatDate(p.col_data?.date_of_last_solvency),
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

    for (const [key, value] of Object.entries(rawForm)) {
      if (RAW_BOOL_FIELDS.includes(key)) {
        payload[key] = value === true ? "1" : "0";
      } else if (value === "") {
        payload[key] = null;
      } else {
        payload[key] = value;
      }
    }

    payload.ci              = parseInt(rawForm.ci)              || null;
    payload.fpv             = parseInt(rawForm.fpv)             || null;
    payload.register_number = parseInt(rawForm.register_number) || null;

    try {
      await runUpdateAction({ id: params.id ?? "", payload });
      setMessage({ type: "success", text: "Expediente actualizado exitosamente." });
      refetch();
      window.scrollTo({ top: 0, behavior: "smooth" });
    } catch (err: any) {
      const msg = err?.message || String(err);
      setMessage({ type: "error", text: msg.replace(/^.*?ApiError:\s*/i, "") || "Error al actualizar." });
      window.scrollTo({ top: 0, behavior: "smooth" });
    } finally {
      setSaving(false);
    }
  };

  // ── Social networks ──────────────────────────────────────────────────────────
  const handleAddSocial = async (payload: { name: string; url: string }) => {
    await runAddSocial({ id: params.id ?? "", payload });
    refetch();
  };

  const handleDeleteSocial = async (socialId: string) => {
    if (!confirm("¿Eliminar esta red social?")) return;
    await runDeleteSocial({ psiId: params.id ?? "", socialId });
    refetch();
  };

  // ─── Convenience setter typed to keyof EditFormState ────────────────────────
  const set = (key: keyof EditFormState, value: any) => setForm(key as any, value);

  // ────────────────────────────────────────────────────────────────────────────
  return (
    <main class="pb-28 animate-in fade-in duration-500">

      <EditPageHeader profile={profile()} />

      <Suspense fallback={<div class="h-96 bg-white animate-pulse rounded-3xl" />}>

        <EditAlert message={message()} />

        <form onSubmit={handleSave} class="space-y-8">
          <AccountSection          form={form} setForm={set} />
          <AdminStatusSection      form={form} setForm={set} />
          <LegalIdentitySection    form={form} setForm={set} />
          <ContactVisibilitySection form={form} setForm={set} />
          <LocationSection         form={form} setForm={set} />
          <ProfessionalSection     form={form} setForm={set} specialties={specialties()} />
          <AcademicSection         form={form} setForm={set} />

          {/* Floating save button */}
          <div class="sticky bottom-16 z-50 flex justify-end">
            <button
              type="submit"
              disabled={saving()}
              class="bg-blue-800 text-white px-10 py-4 rounded-2xl font-black shadow-2xl hover:scale-105 active:scale-95 transition-all disabled:opacity-70 disabled:pointer-events-none flex items-center gap-3 border-2 border-white"
            >
              {saving()
                ? <><span class="animate-spin inline-block">⏳</span> GUARDANDO...</>
                : "💾 GUARDAR EXPEDIENTE"
              }
            </button>
          </div>
        </form>

        <SocialNetworksBlock
          profile={profile()}
          onAdd={handleAddSocial}
          onDelete={handleDeleteSocial}
        />

        <AuditBlock profile={profile()} />

      </Suspense>
    </main>
  );
}