// web/src/routes/psi/perfil.tsx
import { createResource, createEffect, Show, Suspense, createSignal } from "solid-js";
import { createStore } from "solid-js/store";
import { A, action, useAction } from "@solidjs/router";
import { apiGet, apiPatch, apiPost, apiDelete } from "~/lib/api";
import { sanitizeEmail, sanitizePhone, sanitizeText, enforceMaxLength } from "~/lib/sanitizer";
import { ProfileFormData } from "~/types/psi";

import { AccountSection } from "~/components/psi/profile/AccountSection";
import { ContactSection } from "~/components/psi/profile/ContactSection";
import { LocationSection } from "~/components/psi/profile/LocationSection";
import { ProfessionalSection } from "~/components/psi/profile/ProfessionalSection";
import { PrivacySection } from "~/components/psi/profile/PrivacySection";
import { SecuritySection } from "~/components/psi/profile/SecuritySection";
import { MessageAlert } from "~/components/psi/profile/MessageAlert";
import { AcademicSection } from "~/components/psi/profile/AcademicSection";
import { SocialNetworksSection } from "~/components/psi/profile/SocialNetworksSection";
import { SaveButton } from "~/components/psi/profile/SaveButton";
import { AvatarUploader } from "~/components/psi/profile/vatarUploader";

/**
 * ACCIÓN DE SERVIDOR (BFF)
 * Sanitiza los campos antes de enviarlos a Go.
 *
 * FIX: El orden del loop importa — aplicamos enforceMaxLength solo a mini_bio,
 * y dejamos full_bio completamente intacto antes de cualquier otra transformación.
 */
const updateProfileServer = action(async (formData: FormData) => {
  "use server";
  const { sanitizeEmail, sanitizePhone, sanitizeText, enforceMaxLength } = await import("~/lib/sanitizer");
  const { apiPatch } = await import("~/lib/api");

  const cleanFd = new FormData();

  for (const [key, value] of formData.entries()) {
    if (value instanceof File) {
      cleanFd.append(key, value);
      continue;
    }

    if (key === "full_bio") {
      console.log("full_bio recibido:", value.substring(0, 300)) // ver si llega con style=
      cleanFd.append(key, value);
    }

    if (typeof value === "string") {
      // full_bio: intacto, sin transformaciones
      if (key === "full_bio") {
        cleanFd.append(key, value);
        continue;
      }

      // Campos bool de privacidad: pasan directamente como "0" o "1"
      // NUNCA deben pasar por sanitizeEmail ni ninguna otra transformación
      if (key.startsWith("show_")) {
        cleanFd.append(key, value);
        continue;
      }

      let cleanValue = key === "mini_bio" ? enforceMaxLength(value, 250) : value;

      if (["public_phone", "phone_carabobo", "cel_phone_carabobo", "phone_outside_carabobo", "cel_phone_outside_carabobo"].includes(key)) {
        cleanFd.append(key, sanitizePhone(cleanValue));
      } else if (["municipality_carabobo", "state_outside", "municipality_outside_carabobo", "primary_specialty", "secondary_specialty", "mini_bio", "service_address"].includes(key)) {
        cleanFd.append(key, sanitizeText(cleanValue));
      } else if (key === "contact_email" || key === "email") {
        // FIX: comparación exacta en vez de key.includes("email")
        // Así "show_contact_email" nunca entra aquí
        cleanFd.append(key, sanitizeEmail(cleanValue) ?? "");
      } else {
        cleanFd.append(key, cleanValue);
      }
    }
  }

  return await apiPatch("/psi/me", cleanFd);
});
export default function ProfilePage() {
  const [profile, { refetch }] = createResource(() => apiGet<any>("/psi/me"));

  const [form, setForm] = createStore<ProfileFormData>({} as ProfileFormData);
  const [saving, setSaving] = createSignal(false);
  const [message, setMessage] = createSignal<{ type: "success" | "error"; text: string } | null>(null);

  const [files, setFiles] = createSignal<{ [key: string]: File }>({});
  const [avatarFile, setAvatarFile] = createSignal<File | null>(null);

  const [socialForm, setSocialForm] = createStore({ name: "", url: "" });
  const [savingSocial, setSavingSocial] = createSignal(false);

  const runUpdateAction = useAction(updateProfileServer);

  createEffect(() => {
    const p = profile();
    if (p) {
      setForm({
        username: p.username || "",
        email: p.email || "",
        contact_email: sanitizeEmail(p.contact_email) || "",
        public_phone: sanitizePhone(p.public_phone) || "",
        service_address: sanitizeText(p.service_address) || "",
        municipality_carabobo: sanitizeText(p.municipality_carabobo) || "",
        phone_carabobo: sanitizePhone(p.phone_carabobo) || "",
        cel_phone_carabobo: sanitizePhone(p.cel_phone_carabobo) || "",
        state_outside: sanitizeText(p.state_outside) || "",
        municipality_outside_carabobo: sanitizeText(p.municipality_outside_carabobo || p.municipality_out_side_carabobo) || "",
        phone_outside_carabobo: sanitizePhone(p.phone_outside_carabobo) || "",
        cel_phone_outside_carabobo: sanitizePhone(p.cel_phone_outside_carabobo) || "",
        mini_bio: sanitizeText(p.mini_bio) || "",
        primary_specialty: p.primary_specialty || "",
        secondary_specialty: p.secondary_specialty || "",
        // FIX 3: Protección contra full_bio nulo/undefined
        full_bio: p.full_bio?.content || "",
        // Privacidad — booleans directos desde el perfil
        show_contact_email: p.show_contact_email ?? false,
        show_public_phone: p.show_public_phone ?? false,
        show_public_service_address: p.show_public_service_address ?? false,
        show_university_undergraduate: p.col_data?.show_university_undergraduate ?? false,
        show_graduate_date: p.col_data?.show_graduate_date ?? false,
        show_mention_undergraduate: p.col_data?.show_mention_undergraduate ?? false,
      } as ProfileFormData);
    }
  });

  const handleSaveProfile = async (e: Event) => {
    e.preventDefault();

    if (!form.password) {
      setMessage({ type: "error", text: "Debe ingresar su contraseña actual." });
      return;
    }

    setSaving(true);
    setMessage(null);

    const fd = new FormData();

    // ── CAMPOS DE TEXTO ──────────────────────────────────────────────────
    // Iteramos el store manualmente para tener control total sobre la serialización.
    const textFields: (keyof ProfileFormData)[] = [
      "username", "email", "password", "new_password_1", "new_password_2",
      "contact_email", "public_phone", "service_address",
      "municipality_carabobo", "phone_carabobo", "cel_phone_carabobo",
      "state_outside", "municipality_outside_carabobo", "phone_outside_carabobo", "cel_phone_outside_carabobo",
      "primary_specialty", "secondary_specialty", "mini_bio", "full_bio",
    ];

    for (const key of textFields) {
      const val = form[key];
      if (val !== undefined && val !== null && val !== "") {
        fd.append(key, String(val));
      }
    }

    // ── BOOLEANS — Go/Fiber espera "1" / "0" para parsear *bool desde form tags ──
    // FIX 4: Serializar booleans como "1"/"0", no como "true"/"false"
    const boolFields: (keyof ProfileFormData)[] = [
      "show_contact_email",
      "show_public_phone",
      "show_public_service_address",
      "show_university_undergraduate",
      "show_graduate_date",
      "show_mention_undergraduate",
    ];

    for (const key of boolFields) {
      const val = form[key];
      // Enviamos siempre el valor (incluso false → "0") para que el backend pueda
      // distinguir "no enviado" (nil) de "enviado como false" (false).
      fd.append(key, val ? "1" : "0");
    }

    // ── ARCHIVOS DE TÍTULOS ──────────────────────────────────────────────
    const filesObj = files();
    if (filesObj.title_image_one)   fd.append("title_image_one",   filesObj.title_image_one);
    if (filesObj.title_image_two)   fd.append("title_image_two",   filesObj.title_image_two);
    if (filesObj.title_image_three) fd.append("title_image_three", filesObj.title_image_three);

    // ── AVATAR ───────────────────────────────────────────────────────────
    // FIX 5: El avatar se agrega como File directamente, nunca como String
    const avatar = avatarFile();
    if (avatar) fd.append("profile_picture", avatar);

    try {
      await runUpdateAction(fd);
      setMessage({ type: "success", text: "Perfil actualizado correctamente." });
      setForm("password", "");
      setForm("new_password_1", "");
      setForm("new_password_2", "");
      setAvatarFile(null);
      setFiles({});
      refetch();
    } catch (err: any) {
      setMessage({ type: "error", text: err.message || "Error al actualizar." });
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
      setSocialForm({ name: "", url: "" });
      refetch();
    } finally {
      setSavingSocial(false);
    }
  };

  const handleDeleteSocial = async (id: string) => {
    if (!confirm("¿Eliminar?")) return;
    await apiDelete(`/psi/me/social/${id}`);
    refetch();
  };


  return (
    <main class="bg-[#f8fafc] min-h-screen pb-24 font-sans">
      <div class="bg-colpsi-blue pt-10 pb-24 px-4 md:px-8 shadow-inner">
        <div class="max-w-4xl mx-auto flex items-center justify-between">
          <A href="/psi" class="text-white hover:text-colpsi-yellow font-bold flex items-center gap-2 transition-colors">
            <span>←</span> Volver al Panel
          </A>
          <span class="text-blue-200 text-sm font-black tracking-widest uppercase hidden sm:block">Ajustes de Perfil</span>
        </div>
        <div class="max-w-4xl mx-auto mt-8">
          <h1 class="text-white text-3xl md:text-4xl font-black">Tu Identidad Digital</h1>
          <p class="text-blue-200 mt-2 text-sm md:text-base">
            Actualiza tus datos, fotografía y gestiona qué información pueden ver los pacientes.
          </p>
        </div>
      </div>

      <div class="max-w-4xl mx-auto px-4 md:px-8 -mt-16 relative z-10 space-y-8">
        <Suspense fallback={<div class="h-96 bg-white animate-pulse rounded-[2.5rem] shadow-premium border border-gray-100" />}>

          <AvatarUploader
            currentAvatarUrl={profile()?.profile_picture_url}
            avatarFile={avatarFile()}
            onFileChange={setAvatarFile}
            firstName={profile()?.first_name || ""}
            secondName={profile()?.second_name || ""}
            lastName={profile()?.last_name || ""}
            secondLastName={profile()?.second_last_name || ""}
            FPV={profile()?.fpv || 0}
            CI={profile()?.ci || 0}
          />

          <Show when={message()}>
            <MessageAlert type={message()!.type} text={message()!.text} />
          </Show>

          <form onSubmit={handleSaveProfile} class="space-y-8">
            <AccountSection
              username={form.username}
              email={form.email}
              newPassword1={form.new_password_1}
              newPassword2={form.new_password_2}
              onUsernameChange={(v) => setForm("username", v)}
              onEmailChange={(v) => setForm("email", v)}
              onNewPassword1Change={(v) => setForm("new_password_1", v)}
              onNewPassword2Change={(v) => setForm("new_password_2", v)}
            />

            <ContactSection
              contactEmail={form.contact_email ?? ""}
              publicPhone={form.public_phone ?? ""}
              serviceAddress={form.service_address ?? ""}
              onContactEmailChange={(v) => setForm("contact_email", v)}
              onPublicPhoneChange={(v) => setForm("public_phone", v)}
              onServiceAddressChange={(v) => setForm("service_address", v)}
            />

            <AcademicSection
              undergraduateData={{
                university_undergraduate: profile()?.col_data?.university_undergraduate,
                graduate_date: profile()?.col_data?.graduate_date,
                mention_undergraduate: profile()?.col_data?.mention_undergraduate,
                title_image_one_url: profile()?.col_data?.title_image_one_url,
                title_image_two_url: profile()?.col_data?.title_image_two_url,
                title_image_three_url: profile()?.col_data?.title_image_three_url,
                register_number: profile()?.col_data?.register_number,
                register_folio: profile()?.col_data?.register_folio,
                register_tome: profile()?.col_data?.register_tome,
                register_title_date: profile()?.col_data?.register_title_date,
                register_title_state: profile()?.col_data?.register_title_state,
              }}
              showUniversity={form.show_university_undergraduate}
              showGraduateDate={form.show_graduate_date}
              showMention={form.show_mention_undergraduate}
              files={files()}
              setFiles={setFiles}
            />

            <LocationSection
              municipalityCarabobo={form.municipality_carabobo ?? ""}
              phoneCarabobo={form.phone_carabobo ?? ""}
              celPhoneCarabobo={form.cel_phone_carabobo ?? ""}
              stateOutside={form.state_outside ?? ""}
              municipalityOutside={form.municipality_outside_carabobo ?? ""}
              phoneOutside={form.phone_outside_carabobo ?? ""}
              celPhoneOutside={form.cel_phone_outside_carabobo ?? ""}
              onMunicipalityCaraboboChange={(v) => setForm("municipality_carabobo", v)}
              onPhoneCaraboboChange={(v) => setForm("phone_carabobo", v)}
              onCelPhoneCaraboboChange={(v) => setForm("cel_phone_carabobo", v)}
              onStateOutsideChange={(v) => setForm("state_outside", v)}
              onMunicipalityOutsideChange={(v) => setForm("municipality_outside_carabobo", v)}
              onPhoneOutsideChange={(v) => setForm("phone_outside_carabobo", v)}
              onCelPhoneOutsideChange={(v) => setForm("cel_phone_outside_carabobo", v)}
            />

            <ProfessionalSection
              primarySpecialty={form.primary_specialty ?? ""}
              secondarySpecialty={form.secondary_specialty ?? ""}
              miniBio={form.mini_bio ?? ""}
              fullBio={form.full_bio ?? ""}
              onFullBioChange={(v) => setForm("full_bio", v)}
              onPrimarySpecialtyChange={(v) => setForm("primary_specialty", v)}
              onSecondarySpecialtyChange={(v) => setForm("secondary_specialty", v)}
              onMiniBioChange={(v) => setForm("mini_bio", v)}
            />

            <PrivacySection
              showContactEmail={form.show_contact_email}
              showPublicPhone={form.show_public_phone}
              showServiceAddress={form.show_public_service_address}
              showUniversity={form.show_university_undergraduate}
              showGraduateDate={form.show_graduate_date}
              showMention={form.show_mention_undergraduate}
              onShowContactEmailChange={(v) => setForm("show_contact_email", v)}
              onShowPublicPhoneChange={(v) => setForm("show_public_phone", v)}
              onShowServiceAddressChange={(v) => setForm("show_public_service_address", v)}
              onShowUniversityChange={(v) => setForm("show_university_undergraduate", v)}
              onShowGraduateDateChange={(v) => setForm("show_graduate_date", v)}
              onShowMentionChange={(v) => setForm("show_mention_undergraduate", v)}
            />

            <SecuritySection
              password={form.password}
              onPasswordChange={(v) => setForm("password", v)}
              message={message()}
            />

            <SaveButton saving={saving()} />
          </form>

          <SocialNetworksSection
            networks={profile()?.social_networks}
            newNetworkName={socialForm.name}
            newNetworkUrl={socialForm.url}
            saving={savingSocial()}
            onNetworkNameChange={(v) => setSocialForm("name", v)}
            onNetworkUrlChange={(v) => setSocialForm("url", v)}
            onAddNetwork={handleAddSocial}
            onDeleteNetwork={handleDeleteSocial}
          />

        </Suspense>
      </div>
    </main>
  );
}