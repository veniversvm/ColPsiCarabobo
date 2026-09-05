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
  DeontologiaBlock,
  ObservationsBlock,
  DocumentsBlock,
  AuditBlock,
  AccountSection,
  AdminStatusSection,
  LegalIdentitySection,
  ContactVisibilitySection,
  LocationSection,
  ProfessionalSection,
  AcademicSection,
} from "~/components/admin/psicologos/edit";
import type { EditFormState, DeontologiaEntry, ObservacionesEntry } from "~/components/admin/psicologos/edit";
import type { PsiUserDocument } from "~/types/psi";
import { SolvenciesSection } from "~/components/admin/psicologos/edit/SolvenciesSection";
import { Panel, PanelSection } from "~/components/ui/Panel";

// ─── Server Actions ───────────────────────────────────────────────────────────

const updateAdminPsiServer = action(
  async (params: { id: string; payload: FormData }) => {
    "use server";
    const { apiPatch } = await import("~/lib/api");
    // Enviamos el FormData directamente a la API
    return await apiPatch(`/admin/psi/${params.id}`, params.payload);
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

const deleteProfilePictureServer = action(
  async (id: string) => {
    "use server";
    const { apiDelete } = await import("~/lib/api");
    return await apiDelete(`/admin/psi/${id}/picture`);
  },
);

const resetPasswordServer = action(async (id: string) => {
  "use server";
  const { apiPost } = await import("~/lib/api");
  return await apiPost(`/admin/psi/${id}/reset-password`, {});
});

const addDeontologiaServer = action(
  async (params: { id: string; content: string }) => {
    "use server";
    const { apiPost } = await import("~/lib/api");
    return await apiPost(`/admin/psi/${params.id}/deontologia`, {
      content: params.content,
    });
  },
);

const updateDeontologiaServer = action(
  async (params: { psiId: string; entryId: string; content: string }) => {
    "use server";
    const { apiPatch } = await import("~/lib/api");
    return await apiPatch(
      `/admin/psi/${params.psiId}/deontologia/${params.entryId}`,
      { content: params.content },
    );
  },
);

const addObservacionesServer = action(
  async (params: { id: string; content: string }) => {
    "use server";
    const { apiPost } = await import("~/lib/api");
    return await apiPost(`/admin/psi/${params.id}/observaciones`, {
      content: params.content,
    });
  },
);

const updateObservacionesServer = action(
  async (params: { psiId: string; entryId: string; content: string }) => {
    "use server";
    const { apiPatch } = await import("~/lib/api");
    return await apiPatch(
      `/admin/psi/${params.psiId}/observaciones/${params.entryId}`,
      { content: params.content },
    );
  },
);

const addDocumentServer = action(
  async (params: { id: string; payload: FormData }) => {
    "use server";
    const { apiPost } = await import("~/lib/api");
    return await apiPost(`/admin/psi/${params.id}/documents`, params.payload);
  },
);

const updateDocumentServer = action(
  async (params: { psiId: string; docId: string; payload: FormData }) => {
    "use server";
    const { apiPatch } = await import("~/lib/api");
    return await apiPatch(
      `/admin/psi/${params.psiId}/documents/${params.docId}`,
      params.payload,
    );
  },
);

const deleteDocumentServer = action(
  async (params: { psiId: string; docId: string }) => {
    "use server";
    const { apiDelete } = await import("~/lib/api");
    return await apiDelete(
      `/admin/psi/${params.psiId}/documents/${params.docId}`,
    );
  },
);

// ─── Helpers ──────────────────────────────────────────────────────────────────

const formatDate = (dateStr?: string) => (dateStr ? dateStr.split("T")[0] : "");

const calculateAge = (dateStr?: string) => {
  if (!dateStr) return 0;
  const born = new Date(dateStr);
  if (isNaN(born.getTime())) return 0;
  const now = new Date();
  let years = now.getFullYear() - born.getFullYear();
  if (
    now.getMonth() < born.getMonth() ||
    (now.getMonth() === born.getMonth() && now.getDate() < born.getDate())
  ) {
    years--;
  }
  return years < 0 ? 0 : years;
};

// Campos que el backend de Go espera recibir como "1" o "0" desde el FormData
const RAW_BOOL_FIELDS = [
  "is_active",
  "solvent",
  "proof_of_life",
  "ministry_registration_confirmed",
  "show_contact_email",
  "show_public_service_address",
  "show_municipality_carabobo",
  "show_phone_carabobo",
  "show_cel_phone_carabobo",
  "show_state_outside",
  "show_municipality_outside_carabobo",
  "show_phone_outside_carabobo",
  "show_cel_phone_outside_carabobo",
  "show_public_service_address_outside_carabobo",
  "show_phone_outside_venezuela",
  "show_cel_phone_outside_venezuela",
  "show_public_service_address_outside_venezuela",
  "show_university_undergraduate",
  "show_graduate_date",
  "show_mention_undergraduate",
  "guild_director",
  "sixty_five_or_plus",
  "guild_collaborator",
  "public_employee",
  "discapacity",
  "university_professor",
  "double_guild",
  "cpsm",
];

export default function AdminEditPsiPage() {
  const params = useParams();

  const runUpdateAction = useAction(updateAdminPsiServer);
  const runDeletePicture = useAction(deleteProfilePictureServer);
  const runResetPassword = useAction(resetPasswordServer);
  const [profile, { refetch }] = createResource(() =>
    apiGet<any>(`/admin/psi/${params.id}`),
  );
  const [workAreas] = createResource(() => apiGet<any[]>("/specialties"));
  const [deontologia, { refetch: refetchDeontologia }] = createResource(
    () => apiGet<DeontologiaEntry[]>(`/admin/psi/${params.id}/deontologia`),
  );
  const [observaciones, { refetch: refetchObservaciones }] = createResource(
    () => apiGet<ObservacionesEntry[]>(`/admin/psi/${params.id}/observaciones`),
  );
  const [documentos, { refetch: refetchDocumentos }] = createResource(
    () => apiGet<PsiUserDocument[]>(`/admin/psi/${params.id}/documents`),
  );

  const [form, setForm] = createStore<EditFormState>({} as EditFormState);
  const [saving, setSaving] = createSignal(false);
  const [deletingPicture, setDeletingPicture] = createSignal(false);
  const [resettingPassword, setResettingPassword] = createSignal(false);
  const [message, setMessage] = createSignal<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  // Señales para manejar archivos (Imágenes de títulos y perfil)
  const [files, setFiles] = createSignal<{ [key: string]: File }>({});
  const [avatarFile, setAvatarFile] = createSignal<File | null>(null);

  const SITE_URL = import.meta.env.VITE_SITE_URL || "http://localhost:3000";
  const canonicalUrl = () => {
    const p = profile();
    if (!p) return "";
    return `${SITE_URL}/directorio/${p.first_name}-${p.last_name}-fpv${p.fpv}`;
  };

  // ── Sync DB → Store (Mapeo completo) ───────────────────────
  createEffect(() => {
    const p = profile();
    if (!p) return;
    setForm({
      // ... (tu lógica de mapeo actual está bien, asegúrate de incluir todos los campos)
      username: p.username ?? "",
      email: p.email ?? "",
      first_name: p.first_name ?? "",
      second_name: p.second_name ?? "",
      last_name: p.last_name ?? "",
      second_last_name: p.second_last_name ?? "",
      ci: p.ci ?? "",
      fpv: p.fpv ?? "",
      nationality: p.nationality || "V",
      control_number: p.control_number ?? "",
      genre: p.genre ?? "M",
      born_date: formatDate(p.born_date),
      is_active: p.is_active ?? true,
      solvent: p.solvent ?? false,
      proof_of_life: p.proof_of_life ?? false,
      ministry_registration_confirmed: p.col_data?.ministry_registration_confirmed ?? false,
      contact_email: p.contact_email ?? "",
      show_contact_email: p.show_contact_email ?? false,
      contact_phone: p.contact_phone ?? "",
      contact_cell_phone: p.contact_cell_phone ?? "",
      service_address: p.service_address ?? "",
      show_public_service_address: p.show_public_service_address ?? false,
      municipality_carabobo: p.municipality_carabobo ?? "",
      show_municipality_carabobo: p.show_municipality_carabobo ?? false,
      phone_carabobo: p.phone_carabobo ?? "",
      show_phone_carabobo: p.show_phone_carabobo ?? false,
      cel_phone_carabobo: p.cel_phone_carabobo ?? "",
      show_cel_phone_carabobo: p.show_cel_phone_carabobo ?? false,
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
      primary_work_area: p.primary_work_area ?? "",
      secondary_work_area: p.secondary_work_area ?? "",
      mini_bio: p.mini_bio ?? "",
      full_bio: p.full_bio?.content ?? "",
      guild_inscription_date: formatDate(p.col_data?.guild_inscription_date),
      university_undergraduate: p.col_data?.university_undergraduate ?? "",
      graduate_date: formatDate(p.col_data?.graduate_date),
      mention_undergraduate: p.col_data?.mention_undergraduate ?? "",
      register_number: p.col_data?.register_number ?? "",
      register_title_state: p.col_data?.register_title_state ?? "",
      register_title_date: formatDate(p.col_data?.register_title_date),
      register_folio: p.col_data?.register_folio ?? "",
      register_tome: p.col_data?.register_tome ?? "",
      show_university_undergraduate:
        p.col_data?.show_university_undergraduate ?? false,
      show_graduate_date: p.col_data?.show_graduate_date ?? false,
      show_mention_undergraduate:
        p.col_data?.show_mention_undergraduate ?? false,
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

  // ── Submit con FormData ──────────────────────────────────────────────────
  const handleSave = async (e: Event) => {
    e.preventDefault();
    if (saving()) return;
    setSaving(true);
    setMessage(null);

    const rawForm = unwrap(form);
    const fd = new FormData(); // 👈 Usamos FormData para soportar imágenes y tipos de Go

    // 1. Procesar campos del formulario
    // 1. Procesar campos del formulario
    for (const [key, value] of Object.entries(rawForm)) {
      // Failsafe: Si el valor es null, undefined o un string vacío, NO lo enviamos.
      // Esto garantiza que en Go el puntero llegue como 'nil'.
      if (value === "" || value === null || value === undefined) {
        continue;
      }

      // Caso especial: Historial de Solvencias (Array -> JSON String)
      if (key === "solvencies") {
        fd.append("solvencies", JSON.stringify(value));
        continue;
      }

      // Caso especial: Booleanos de Privacidad y Flags
      if (RAW_BOOL_FIELDS.includes(key)) {
        // Enviamos "1" o "0" que es lo más compatible con parsers de formularios
        fd.append(key, value === true ? "1" : "0");
        continue;
      }

      // Resto de campos (Strings, Números, etc.)
      fd.append(key, String(value));
    }
    // 2. Adjuntar Archivos (Imágenes de títulos)
    const filesObj = files();
    if (filesObj.title_image_one)
      fd.append("title_image_one", filesObj.title_image_one);
    if (filesObj.title_image_two)
      fd.append("title_image_two", filesObj.title_image_two);
    if (filesObj.title_image_three)
      fd.append("title_image_three", filesObj.title_image_three);

    // 3. Adjuntar Foto de Perfil
    const avatar = avatarFile();
    if (avatar) fd.append("profile_picture", avatar);

    try {
      await runUpdateAction({ id: params.id ?? "", payload: fd });
      setMessage({
        type: "success",
        text: "Expediente actualizado exitosamente.",
      });
      refetch();
      window.scrollTo({ top: 0, behavior: "smooth" });
    } catch (err: any) {
      const msg = err?.message || String(err);
      setMessage({
        type: "error",
        text: msg.replace(/^.*?ApiError:\s*/i, "") || "Error al guardar.",
      });
    } finally {
      setSaving(false);
    }
  };

  const set = (key: keyof EditFormState, value: any) =>
    setForm(key as any, value);

  const handleDeletePicture = async () => {
    const id = params.id ?? "";
    if (!id || deletingPicture() || !confirm("¿Eliminar la foto de perfil del psicólogo?")) {
      return;
    }
    setDeletingPicture(true);
    setMessage(null);
    try {
      await runDeletePicture(id);
      setAvatarFile(null);
      refetch();
      setMessage({ type: "success", text: "Foto de perfil eliminada." });
    } catch (err: any) {
      const msg = err?.message || String(err);
      setMessage({
        type: "error",
        text: msg.replace(/^.*?ApiError:\s*/i, "") || "Error al eliminar la foto.",
      });
    } finally {
      setDeletingPicture(false);
    }
  };

  const handleResetPassword = async () => {
    const id = params.id ?? "";
    if (
      !id ||
      resettingPassword() ||
      !confirm(
        "¿Reiniciar la clave de acceso de este psicólogo?\n\n" +
          "Se cerrarán sus sesiones activas y recibirá una contraseña temporal por correo. Deberá cambiarla al iniciar sesión.",
      )
    ) {
      return;
    }
    setResettingPassword(true);
    setMessage(null);
    try {
      await runResetPassword(id);
      setMessage({
        type: "success",
        text: "Clave reiniciada. La contraseña temporal fue enviada al correo del psicólogo.",
      });
    } catch (err: any) {
      const msg = err?.message || String(err);
      setMessage({
        type: "error",
        text: msg.replace(/^.*?ApiError:\s*/i, "") || "Error al reiniciar la clave.",
      });
    } finally {
      setResettingPassword(false);
    }
  };

  return (
    <main class="pb-28 animate-in fade-in duration-500 font-sans">
      <EditPageHeader profile={profile()} />
      <Suspense
        fallback={
          <div class="h-96 bg-white animate-pulse rounded-[2.5rem] border border-gray-100" />
        }
      >
        <EditAlert message={message()} />

        <form onSubmit={handleSave}>
          <Panel>
            {/* Se pasa avatarFile y onFileChange para que AccountSection pueda capturar la foto */}
            <PanelSection title="Cuenta y Perfil Visual" accent="border-colpsi-yellow" defaultOpen>
              <AccountSection
                form={form}
                setForm={set}
                url={canonicalUrl()}
                avatarFile={avatarFile()}
                onAvatarChange={setAvatarFile}
                pictureUrl={profile()?.profile_picture_url || ""}
                onDeletePicture={handleDeletePicture}
                onResetPassword={handleResetPassword}
                resettingPassword={resettingPassword()}
              />
            </PanelSection>
            <PanelSection title="Estatus Administrativo" accent="border-yellow-400">
              <AdminStatusSection form={form} setForm={set} />
            </PanelSection>

            <PanelSection title="Historial de Solvencias" accent="border-emerald-400">
              <SolvenciesSection
                solvencies={form.solvencies}
                onAddLocalSolvency={(year) => {
                  const newSolv = { date: `${year}-12-31T00:00:00Z` };
                  set("solvencies", [...form.solvencies, newSolv]);
                }}
              />
            </PanelSection>

            <PanelSection title="Identidad Legal">
              <LegalIdentitySection form={form} setForm={set} age={calculateAge(form.born_date)} />
            </PanelSection>
            <PanelSection title="Gestión de Contacto y Privacidad" accent="border-colpsi-blue">
              <ContactVisibilitySection form={form} setForm={set} />
            </PanelSection>
            <PanelSection title="Ubicación Geográfica y Privacidad" accent="border-indigo-400">
              <LocationSection form={form} setForm={set} />
            </PanelSection>
            <PanelSection title="Perfil Profesional">
              <ProfessionalSection
                form={form}
                setForm={set}
                workAreas={workAreas()}
              />
            </PanelSection>

            {/* Se pasan los estados de archivos a AcademicSection */}
            <PanelSection title="Expediente Académico y Gremial" accent="border-colpsi-blue">
              <AcademicSection
                form={form}
                setForm={set}
                files={files()}
                setFiles={setFiles}
              />
            </PanelSection>
          </Panel>

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

        <div class="mt-8">
          <Panel>
            <PanelSection title="Presencia Digital / Redes Sociales" accent="border-gray-300">
              <SocialNetworksBlock
                profile={profile()}
                onAdd={async (p) => {
                  const id = params.id ?? "";
                  if (!id) return;
                  await addSocialServer({ id, payload: p });
                  refetch();
                }}
                onDelete={async (sid) => {
                  const id = params.id ?? "";
                  if (!id || !confirm("¿Eliminar?")) return;
                  await deleteSocialServer({ psiId: id, socialId: sid });
                  refetch();
                }}
              />
            </PanelSection>

            <PanelSection title="Expediente Deontológico" accent="border-gray-300">
              <DeontologiaBlock
                entries={deontologia()}
                onAdd={async (content) => {
                  const id = params.id ?? "";
                  if (!id) return;
                  await addDeontologiaServer({ id, content });
                  refetchDeontologia();
                }}
                onUpdate={async (entryId, content) => {
                  const id = params.id ?? "";
                  if (!id) return;
                  await updateDeontologiaServer({ psiId: id, entryId, content });
                  refetchDeontologia();
                }}
              />
            </PanelSection>

            <PanelSection title="Observaciones Internas" accent="border-gray-300">
              <ObservationsBlock
                entries={observaciones()}
                onAdd={async (content) => {
                  const id = params.id ?? "";
                  if (!id) return;
                  await addObservacionesServer({ id, content });
                  refetchObservaciones();
                }}
                onUpdate={async (entryId, content) => {
                  const id = params.id ?? "";
                  if (!id) return;
                  await updateObservacionesServer({ psiId: id, entryId, content });
                  refetchObservaciones();
                }}
              />
            </PanelSection>

            <PanelSection title="Registro Digital de Documentos" accent="border-gray-300">
              <DocumentsBlock
                entries={documentos()}
                onAdd={async (payload) => {
                  const id = params.id ?? "";
                  if (!id) return;
                  await addDocumentServer({ id, payload });
                  refetchDocumentos();
                }}
                onUpdate={async (docId, payload) => {
                  const id = params.id ?? "";
                  if (!id) return;
                  await updateDocumentServer({ psiId: id, docId, payload });
                  refetchDocumentos();
                }}
                onDelete={async (docId) => {
                  const id = params.id ?? "";
                  if (!id) return;
                  await deleteDocumentServer({ psiId: id, docId });
                  refetchDocumentos();
                }}
              />
            </PanelSection>

            <PanelSection title="Información de Auditoría (Solo Lectura)" accent="border-gray-400">
              <AuditBlock profile={profile()} />
            </PanelSection>
          </Panel>
        </div>
      </Suspense>
    </main>
  );
}
