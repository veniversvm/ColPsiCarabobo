import {
  createResource,
  createEffect,
  Show,
  Suspense,
  createSignal,
} from "solid-js";
import { createStore } from "solid-js/store";
import { A, action, useAction } from "@solidjs/router";
import { apiGet, apiPatch, apiPost, apiDelete } from "~/lib/api";
import {
  sanitizeEmail,
  sanitizePhone,
  sanitizeText,
  enforceMaxLength,
} from "~/lib/sanitizer";
import { ProfileFormData } from "~/types/psi";

import { AccountSection } from "~/components/psi/profile/AccountSection";
import { ContactSection } from "~/components/psi/profile/ContactSection";
import { LocationSection } from "~/components/psi/profile/LocationSection";
import { ProfessionalSection } from "~/components/psi/profile/ProfessionalSection";
import { PrivacySection } from "~/components/psi/profile/PrivacySection";
import { ServicePreferencesSection } from "~/components/psi/profile/ServicePreferencesSection";
import { SecuritySection } from "~/components/psi/profile/SecuritySection";
import { MessageAlert } from "~/components/psi/profile/MessageAlert";
import { AcademicSection } from "~/components/psi/profile/AcademicSection";
import { SocialNetworksSection } from "~/components/psi/profile/SocialNetworksSection";
import { SaveButton } from "~/components/psi/profile/SaveButton";
import { AvatarUploader } from "~/components/psi/profile/AvatarUploader";
import { Panel, PanelSection } from "~/components/ui/Panel";

const updateProfileServer = action(async (formData: FormData) => {
  "use server";
  const { sanitizeEmail, sanitizePhone, sanitizeText, enforceMaxLength } =
    await import("~/lib/sanitizer");
  const { apiPatch } = await import("~/lib/api");

  const cleanFd = new FormData();

  for (const [key, value] of formData.entries()) {
    if (value instanceof File) {
      cleanFd.append(key, value);
      continue;
    }

    if (typeof value === "string") {
      if (key === "full_bio") {
        const { sanitizeHtml } = await import("~/lib/sanitize-html");
        cleanFd.append(key, sanitizeHtml(value));
        continue;
      }

      if (key.startsWith("show_")) {
        cleanFd.append(key, value);
        continue;
      }

      let cleanValue =
        key === "mini_bio" ? enforceMaxLength(value, 250) : value;

      const phoneFields = [
        "contact_phone",
        "contact_cell_phone", // Reemplazan a public_phone
        "phone_carabobo",
        "cel_phone_carabobo",
        "phone_outside_carabobo",
        "cel_phone_outside_carabobo",
        "phone_outside_venezuela",
        "cell_phone_outside_venezuela", // Nuevo cell
      ];

      const textFields = [
        "municipality_carabobo",
        "state_outside",
        "municipality_outside_carabobo",
        "service_address",
        "service_address_outside_carabobo",
        "service_address_outside_venezuela",
        "country",
        "primary_work_area",
        "secondary_work_area",
        "mini_bio", // WorkArea en vez de specialty
      ];

      if (phoneFields.includes(key)) {
        cleanFd.append(key, sanitizePhone(cleanValue));
      } else if (textFields.includes(key)) {
        cleanFd.append(key, sanitizeText(cleanValue));
      } else if (key === "contact_email" || key === "email") {
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
  const [specialties] = createResource(() => apiGet<any[]>("/specialties")); // Asumiendo que esta API devuelve áreas de trabajo ahora

  const [form, setForm] = createStore<ProfileFormData>({} as ProfileFormData);
  const [saving, setSaving] = createSignal(false);
  const [message, setMessage] = createSignal<{
    type: "success" | "error";
    text: string;
  } | null>(null);

  const [files, setFiles] = createSignal<{ [key: string]: File }>({});
  const [avatarFile, setAvatarFile] = createSignal<File | null>(null);

  const [socialForm, setSocialForm] = createStore({ name: "", url: "" });
  const [savingSocial, setSavingSocial] = createSignal(false);

  const runUpdateAction = useAction(updateProfileServer);

  const SITE_URL =
    import.meta.env.VITE_SITE_URL || "http://localhost:28080/api/v1";

  const canonicalUrl = `${SITE_URL}/directorio/${profile()?.first_name}-${profile()?.last_name}-fpv${profile()?.fpv}`;

  createEffect(() => {
    const p = profile();
    if (p) {
      setForm({
        username: p.username || "",
        email: p.email || "",
        contact_email: sanitizeEmail(p.contact_email) || "",
        contact_phone: sanitizePhone(p.contact_phone) || "", // NUEVO
        contact_cell_phone: sanitizePhone(p.contact_cell_phone) || "", // NUEVO
        service_address: sanitizeText(p.service_address) || "",

        // Carabobo
        municipality_carabobo: sanitizeText(p.municipality_carabobo) || "",
        phone_carabobo: sanitizePhone(p.phone_carabobo) || "",
        cel_phone_carabobo: sanitizePhone(p.cel_phone_carabobo) || "",

        // Fuera de Carabobo (Venezuela)
        state_outside: sanitizeText(p.state_outside) || "",
        municipality_outside_carabobo:
          sanitizeText(
            p.municipality_outside_carabobo || p.municipality_out_side_carabobo,
          ) || "",
        phone_outside_carabobo: sanitizePhone(p.phone_outside_carabobo) || "",
        cel_phone_outside_carabobo:
          sanitizePhone(p.cel_phone_outside_carabobo) || "",
        service_address_outside_carabobo:
          sanitizeText(p.service_address_outside_carabobo) || "",

        // Exterior
        country: sanitizeText(p.country) || "",
        phone_outside_venezuela: sanitizePhone(p.phone_outside_venezuela) || "",
        cell_phone_outside_venezuela:
          sanitizePhone(p.cell_phone_outside_venezuela) || "", // NUEVO
        service_address_outside_venezuela:
          sanitizeText(p.service_address_outside_venezuela) || "",

        mini_bio: sanitizeText(p.mini_bio) || "",
        primary_work_area: p.primary_work_area || "", // NUEVO
        secondary_work_area: p.secondary_work_area || "", // NUEVO
        full_bio: p.full_bio?.content || "",

        // Privacidad: Contacto principal
        show_contact_email: p.show_contact_email ?? false,
        show_public_service_address: p.show_public_service_address ?? false,

        // Privacidad: Carabobo (NUEVOS)
        show_municipality_carabobo: p.show_municipality_carabobo ?? false,
        show_phone_carabobo: p.show_phone_carabobo ?? false,
        show_cel_phone_carabobo: p.show_cel_phone_carabobo ?? false,

        // Privacidad: Fuera de Carabobo
        show_state_outside: p.show_state_outside ?? false, // NUEVO
        show_municipality_outside_carabobo:
          p.show_municipality_outside_carabobo ?? false, // NUEVO
        show_phone_outside_carabobo: p.show_phone_outside_carabobo ?? false,
        show_cel_phone_outside_carabobo:
          p.show_cel_phone_outside_carabobo ?? false,
        show_public_service_address_outside_carabobo:
          p.show_public_service_address_outside_carabobo ?? false,

        // Privacidad: Exterior
        show_phone_outside_venezuela: p.show_phone_outside_venezuela ?? false,
        show_cel_phone_outside_venezuela:
          p.show_cel_phone_outside_venezuela ?? false,
        show_public_service_address_outside_venezuela:
          p.show_public_service_address_outside_venezuela ?? false,

        // Privacidad: Académicos
        show_university_undergraduate:
          p.col_data?.show_university_undergraduate ?? false,
        show_graduate_date: p.col_data?.show_graduate_date ?? false,
        show_mention_undergraduate:
          p.col_data?.show_mention_undergraduate ?? false,

        // Modalidad de servicio (auto-gestión)
        service_modality_presencial: p.service_modality_presencial ?? false,
        service_modality_distance: p.service_modality_distance ?? false,
        service_modality_telephone: p.service_modality_telephone ?? false,
        show_service_modality: p.show_service_modality ?? false,

        // Aviso de cumpleaños (opt-in)
        birthday_notification: p.col_data?.birthday_notification ?? false,
      } as ProfileFormData);
    }
  });

  const handleSaveProfile = async (e: Event) => {
    e.preventDefault();

    if (!form.password) {
      setMessage({
        type: "error",
        text: "Debe ingresar su contraseña actual.",
      });
      return;
    }

    setSaving(true);
    setMessage(null);

    const fd = new FormData();

    const textFields: (keyof ProfileFormData)[] = [
      "username",
      "email",
      "password",
      "new_password_1",
      "new_password_2",
      "contact_email",
      "contact_phone",
      "contact_cell_phone",
      "service_address",
      // Carabobo
      "municipality_carabobo",
      "phone_carabobo",
      "cel_phone_carabobo",
      // Venezuela
      "state_outside",
      "municipality_outside_carabobo",
      "phone_outside_carabobo",
      "cel_phone_outside_carabobo",
      "service_address_outside_carabobo",
      // Exterior
      "country",
      "phone_outside_venezuela",
      "cell_phone_outside_venezuela",
      "service_address_outside_venezuela",
      // Profesional
      "primary_work_area",
      "secondary_work_area",
      "mini_bio",
      "full_bio",
    ];

    for (const key of textFields) {
      const val = form[key];
      if (val !== undefined && val !== null && val !== "") {
        fd.append(key, String(val));
      }
    }

    const boolFields: (keyof ProfileFormData)[] = [
      // Contacto principal
      "show_contact_email",
      "show_public_service_address",
      // Carabobo
      "show_municipality_carabobo",
      "show_phone_carabobo",
      "show_cel_phone_carabobo",
      // Fuera de Carabobo
      "show_state_outside",
      "show_municipality_outside_carabobo",
      "show_phone_outside_carabobo",
      "show_cel_phone_outside_carabobo",
      "show_public_service_address_outside_carabobo",
      // Exterior
      "show_phone_outside_venezuela",
      "show_cel_phone_outside_venezuela",
      "show_public_service_address_outside_venezuela",
      // Académicos
      "show_university_undergraduate",
      "show_graduate_date",
      "show_mention_undergraduate",
      // Modalidad de servicio
      "service_modality_presencial",
      "service_modality_distance",
      "service_modality_telephone",
      "show_service_modality",
      // Aviso de cumpleaños
      "birthday_notification",
    ];

    for (const key of boolFields) {
      fd.append(key, form[key] ? "1" : "0");
    }

    // Archivos de títulos
    const filesObj = files();
    if (filesObj.title_image_one)
      fd.append("title_image_one", filesObj.title_image_one);
    if (filesObj.title_image_two)
      fd.append("title_image_two", filesObj.title_image_two);
    if (filesObj.title_image_three)
      fd.append("title_image_three", filesObj.title_image_three);

    const avatar = avatarFile();
    if (avatar) fd.append("profile_picture", avatar);

    try {
      await runUpdateAction(fd);
      setMessage({
        type: "success",
        text: "Perfil actualizado correctamente.",
      });
      setForm("password", "");
      setForm("new_password_1", "");
      setForm("new_password_2", "");
      setAvatarFile(null);
      setFiles({});
      refetch();
    } catch (err: any) {
      setMessage({
        type: "error",
        text: err.message || "Error al actualizar.",
      });
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
    <main class="bg-colpsi-bg min-h-screen pb-24 font-sans">
      <div class="bg-heraldic pt-12 pb-20 px-4 md:px-8 shadow-inner">
        <div class="max-w-6xl mx-auto flex items-center justify-between">
          <A
            href="/psi"
            class="bg-colpsi-yellow text-colpsi-blue px-5 py-2.5 rounded-full font-black text-sm shadow-lg hover:bg-colpsi-yellow/90 active:scale-95 transition-all inline-flex items-center gap-2"
          >
            <span>←</span> Volver al Panel
          </A>
          <span class="text-blue-200 text-sm font-black tracking-widest uppercase hidden sm:block">
            Ajustes de Perfil
          </span>
        </div>
        <div class="max-w-6xl mx-auto mt-8">
          <h1 class="text-white text-3xl md:text-4xl font-black">
            Tu Identidad Digital
          </h1>
          <p class="text-blue-200 mt-2 text-sm md:text-base">
            Actualiza tus datos, fotografía y gestiona qué información pueden
            ver los pacientes.
          </p>
        </div>
      </div>

      <div class="max-w-6xl mx-auto px-4 md:px-8 -mt-16 relative z-10 space-y-8">
        <Suspense
          fallback={
            <div class="h-96 bg-white animate-pulse rounded-[2.5rem] shadow-premium border border-colpsi-border" />
          }
        >
          <AvatarUploader
            url={canonicalUrl}
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

          <form onSubmit={handleSaveProfile}>
            <Panel>
              <PanelSection title="Cuenta y Seguridad" accent="border-colpsi-yellow" defaultOpen>
                <AccountSection
                  username={form.username}
                  email={form.email}
                  newPassword1={form.new_password_1 || ""}
                  newPassword2={form.new_password_2 || ""}
                  onUsernameChange={(v) => setForm("username", v)}
                  onEmailChange={(v) => setForm("email", v)}
                  onNewPassword1Change={(v) => setForm("new_password_1", v)}
                  onNewPassword2Change={(v) => setForm("new_password_2", v)}
                />

                <SecuritySection
                  password={form.password ?? ""}
                  onPasswordChange={(v) => setForm("password", v)}
                  message={message()}
                />
              </PanelSection>

              <PanelSection title="Información de Contacto" accent="border-colpsi-blue">
                <ContactSection
                  contactEmail={form.contact_email ?? ""}
                  contactPhone={form.contact_phone ?? ""}
                  contactCellPhone={form.contact_cell_phone ?? ""}
                  onContactEmailChange={(v) => setForm("contact_email", v)}
                  onContactPhoneChange={(v) => setForm("contact_phone", v)}
                  onContactCellPhoneChange={(v) => setForm("contact_cell_phone", v)}
                />
              </PanelSection>

              <PanelSection title="Expediente Académico" accent="border-colpsi-blue">
                <AcademicSection
                  undergraduateData={{
                    university_undergraduate:
                      profile()?.col_data?.university_undergraduate,
                    graduate_date: profile()?.col_data?.graduate_date,
                    mention_undergraduate:
                      profile()?.col_data?.mention_undergraduate,
                    title_image_one_url: profile()?.col_data?.title_image_one_url,
                    title_image_two_url: profile()?.col_data?.title_image_two_url,
                    title_image_three_url:
                      profile()?.col_data?.title_image_three_url,
                    register_number: profile()?.col_data?.register_number,
                    register_folio: profile()?.col_data?.register_folio,
                    register_tome: profile()?.col_data?.register_tome,
                    register_title_date: profile()?.col_data?.register_title_date,
                    register_title_state: profile()?.col_data?.register_title_state,
                  }}
                  showUniversity={true}
                  showGraduateDate={true}
                  showMention={true}
                  files={files()}
                  setFiles={setFiles}
                />
              </PanelSection>

              <PanelSection title="Ubicación Geográfica" accent="border-indigo-400">
                <LocationSection
                  serviceAddress={form.service_address ?? ""}
                  municipalityCarabobo={form.municipality_carabobo ?? ""}
                  phoneCarabobo={form.phone_carabobo ?? ""}
                  celPhoneCarabobo={form.cel_phone_carabobo ?? ""}
                  stateOutside={form.state_outside ?? ""}
                  municipalityOutside={form.municipality_outside_carabobo ?? ""}
                  phoneOutside={form.phone_outside_carabobo ?? ""}
                  celPhoneOutside={form.cel_phone_outside_carabobo ?? ""}
                  serviceAddressOutsideCarabobo={
                    form.service_address_outside_carabobo ?? ""
                  }
                  country={form.country ?? ""}
                  phoneOutsideVenezuela={form.phone_outside_venezuela ?? ""}
                  cellPhoneOutsideVenezuela={
                    form.cell_phone_outside_venezuela ?? ""
                  }
                  serviceAddressOutsideVenezuela={
                    form.service_address_outside_venezuela ?? ""
                  }
                  onServiceAddressChange={(v) => setForm("service_address", v)}
                  onMunicipalityCaraboboChange={(v) =>
                    setForm("municipality_carabobo", v)
                  }
                  onPhoneCaraboboChange={(v) => setForm("phone_carabobo", v)}
                  onCelPhoneCaraboboChange={(v) => setForm("cel_phone_carabobo", v)}
                  onStateOutsideChange={(v) => setForm("state_outside", v)}
                  onMunicipalityOutsideChange={(v) =>
                    setForm("municipality_outside_carabobo", v)
                  }
                  onPhoneOutsideChange={(v) => setForm("phone_outside_carabobo", v)}
                  onCelPhoneOutsideChange={(v) =>
                    setForm("cel_phone_outside_carabobo", v)
                  }
                  onServiceAddressOutsideCaraboboChange={(v) =>
                    setForm("service_address_outside_carabobo", v)
                  }
                  onCountryChange={(v) => setForm("country", v)}
                  onPhoneOutsideVenezuelaChange={(v) =>
                    setForm("phone_outside_venezuela", v)
                  }
                  onCellPhoneOutsideVenezuelaChange={(v) =>
                    setForm("cell_phone_outside_venezuela", v)
                  }
                  onServiceAddressOutsideVenezuelaChange={(v) =>
                    setForm("service_address_outside_venezuela", v)
                  }
                />
              </PanelSection>

              <PanelSection title="Perfil Profesional">
                <ProfessionalSection
                  primaryWorkArea={form.primary_work_area ?? ""}
                  secondaryWorkArea={form.secondary_work_area ?? ""}
                  miniBio={form.mini_bio ?? ""}
                  fullBio={form.full_bio ?? ""}
                  specialties={specialties()}
                  onFullBioChange={(v) => setForm("full_bio", v)}
                  onPrimaryWorkAreaChange={(v) => setForm("primary_work_area", v)}
                  onSecondaryWorkAreaChange={(v) =>
                    setForm("secondary_work_area", v)
                  }
                  onMiniBioChange={(v) => setForm("mini_bio", v)}
                />
              </PanelSection>

              <PanelSection title="Servicio y Preferencias" accent="border-teal-400">
                <ServicePreferencesSection
                  serviceModalityPresencial={
                    form.service_modality_presencial ?? false
                  }
                  serviceModalityDistance={form.service_modality_distance ?? false}
                  serviceModalityTelephone={
                    form.service_modality_telephone ?? false
                  }
                  showServiceModality={form.show_service_modality ?? false}
                  birthdayNotification={form.birthday_notification ?? false}
                  onServiceModalityPresencialChange={(v) =>
                    setForm("service_modality_presencial", v)
                  }
                  onServiceModalityDistanceChange={(v) =>
                    setForm("service_modality_distance", v)
                  }
                  onServiceModalityTelephoneChange={(v) =>
                    setForm("service_modality_telephone", v)
                  }
                  onShowServiceModalityChange={(v) =>
                    setForm("show_service_modality", v)
                  }
                  onBirthdayNotificationChange={(v) =>
                    setForm("birthday_notification", v)
                  }
                />
              </PanelSection>

              <PanelSection title="Privacidad y Visibilidad" accent="border-colpsi-yellow">
                <PrivacySection
                  showContactEmail={form.show_contact_email}
                  showServiceAddress={form.show_public_service_address}
                  showMunicipalityCarabobo={form.show_municipality_carabobo}
                  showPhoneCarabobo={form.show_phone_carabobo}
                  showCelPhoneCarabobo={form.show_cel_phone_carabobo}
                  showStateOutside={form.show_state_outside}
                  showMunicipalityOutsideCarabobo={
                    form.show_municipality_outside_carabobo
                  }
                  showPhoneOutsideCarabobo={form.show_phone_outside_carabobo}
                  showCelPhoneOutsideCarabobo={form.show_cel_phone_outside_carabobo}
                  showServiceAddressOutsideCarabobo={
                    form.show_public_service_address_outside_carabobo
                  }
                  showPhoneOutsideVenezuela={form.show_phone_outside_venezuela}
                  showCelPhoneOutsideVenezuela={
                    form.show_cel_phone_outside_venezuela
                  }
                  showServiceAddressOutsideVenezuela={
                    form.show_public_service_address_outside_venezuela
                  }
                  showUniversity={form.show_university_undergraduate}
                  showGraduateDate={form.show_graduate_date}
                  showMention={form.show_mention_undergraduate}
                  onShowContactEmailChange={(v) => setForm("show_contact_email", v)}
                  onShowServiceAddressChange={(v) =>
                    setForm("show_public_service_address", v)
                  }
                  onShowMunicipalityCaraboboChange={(v) =>
                    setForm("show_municipality_carabobo", v)
                  }
                  onShowPhoneCaraboboChange={(v) =>
                    setForm("show_phone_carabobo", v)
                  }
                  onShowCelPhoneCaraboboChange={(v) =>
                    setForm("show_cel_phone_carabobo", v)
                  }
                  onShowStateOutsideChange={(v) => setForm("show_state_outside", v)}
                  onShowMunicipalityOutsideCaraboboChange={(v) =>
                    setForm("show_municipality_outside_carabobo", v)
                  }
                  onShowPhoneOutsideCaraboboChange={(v) =>
                    setForm("show_phone_outside_carabobo", v)
                  }
                  onShowCelPhoneOutsideCaraboboChange={(v) =>
                    setForm("show_cel_phone_outside_carabobo", v)
                  }
                  onShowServiceAddressOutsideCaraboboChange={(v) =>
                    setForm("show_public_service_address_outside_carabobo", v)
                  }
                  onShowPhoneOutsideVenezuelaChange={(v) =>
                    setForm("show_phone_outside_venezuela", v)
                  }
                  onShowCelPhoneOutsideVenezuelaChange={(v) =>
                    setForm("show_cel_phone_outside_venezuela", v)
                  }
                  onShowServiceAddressOutsideVenezuelaChange={(v) =>
                    setForm("show_public_service_address_outside_venezuela", v)
                  }
                  onShowUniversity={(v) =>
                    setForm("show_university_undergraduate", v)
                  }
                  onShowGraduateDateChange={(v) => setForm("show_graduate_date", v)}
                  onShowMentionChange={(v) =>
                    setForm("show_mention_undergraduate", v)
                  }
                />
              </PanelSection>

              <PanelSection title="Redes Sociales">
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
              </PanelSection>
            </Panel>

            <SaveButton saving={saving()} />
          </form>
        </Suspense>
      </div>
    </main>
  );
}
