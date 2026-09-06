/**
 * Página de perfil público del psicólogo con URLs amigables y SSR.
 * Formato: nombre-apellido(s)-fpv-1234
 */
import { createSignal, Show, Suspense } from "solid-js";
import { createResource } from "solid-js";
import { useParams, A } from "@solidjs/router";
import { apiGet, ApiError } from "~/lib/api";
import { Meta, Title, Link } from "@solidjs/meta";
import { PsiProfile } from "~/types/psi";
import { bucketUrl } from "~/lib/bucket";
import { ProfileHeader } from "~/components/psi/ProfileHeader";
import { ContactCard } from "~/components/psi/ContactCard";
import { AcademicSection } from "~/components/psi/AcademicSection";
import { PostGradeModal } from "~/components/psi/PostGradeModal";
import { sortPostGradesByYear, extractFpvFromSlug } from "~/lib/utils";
import { FullBioModal } from "~/components/directory/FullBioModal";

export const ssr = true;

const SITE_URL = import.meta.env.VITE_SITE_URL || "http://localhost:3000";
const imgUrl = bucketUrl;

// ✅ Server function para obtener el perfil desde la API
async function fetchProfile(fpv: string): Promise<PsiProfile | null> {
  "use server";
  if (!fpv) return null;
  try {
    return await apiGet<PsiProfile>(`/psi/${fpv}`);
  } catch (err: any) {
    if (err instanceof ApiError) {
      console.warn(`[fetchProfile] ${err.status} – fpv=${fpv}: ${err.message}`);
      return null;
    }
    return null;
  }
}

export default function PsiProfilePage() {
  const params = useParams();
  const [selectedPostGrade, setSelectedPostGrade] = createSignal(null);
  const [showBioModal, setShowBioModal] = createSignal(false);

  const [profile] = createResource(async () => {
    try {
      const fpv = extractFpvFromSlug(params.slug ?? "");
      if (!fpv) return null;
      return await fetchProfile(String(fpv));
    } catch {
      return null;
    }
  });

  const profileData = () => profile();

  // SEO Helpers
  const fullName = () => {
    const p = profileData();
    if (!p) return "Psicólogo";
    return [p.first_name, p.second_name, p.last_name, p.second_last_name]
      .filter(Boolean)
      .join(" ");
  };

  const canonicalUrl = `${SITE_URL}/directorio/${params.slug}`;
  
  const ogImage = () => {
    const p = profileData();
    return p?.profile_picture ? imgUrl(p.profile_picture) : `${SITE_URL}/og-default.jpg`;
  };

  const description = () => {
    const p = profileData();
    return p?.mini_bio || `Perfil oficial de el/la Psicólogo(a) ${fullName()} en el Colegio de Psicólogos del Estado Carabobo. Información de contacto y formación académica.`;
  };

  return (
    <>
      {/* ── SEO AVANZADO (SSR) ────────────────────────────────────────────── */}
      <Title>{`${fullName()} | COLPSI Carabobo`}</Title>
      <Meta name="description" content={description()} />
      <Meta name="keywords" content={`psicólogo, salud mental, Carabobo, Valencia, ${fullName()}, FPV ${profileData()?.fpv}, directorio profesional`} />
      <Meta name="robots" content="index, follow" />
      <Link rel="canonical" href={canonicalUrl} />

      {/* Open Graph / Facebook */}
      <Meta property="og:type" content="profile" />
      <Meta property="og:url" content={canonicalUrl} />
      <Meta property="og:title" content={`${fullName()} | COLPSI Carabobo`} />
      <Meta property="og:description" content={description()} />
      <Meta property="og:image" content={ogImage()} />
      
      {/* Etiquetas específicas de Perfil */}
      <Show when={profileData()}>
        {(p) => (
          <>
            <Meta property="profile:first_name" content={p().first_name} />
            <Meta property="profile:last_name" content={p().last_name} />
            <Meta property="profile:username" content={`fpv-${p().fpv}`} />
            <Meta property="profile:gender" content={p().gender === 'M' ? 'male' : 'female'} />
          </>
        )}
      </Show>

      {/* Twitter */}
      <Meta name="twitter:card" content="summary_large_image" />
      <Meta name="twitter:title" content={`${fullName()} | COLPSI Carabobo`} />
      <Meta name="twitter:description" content={description()} />
      <Meta name="twitter:image" content={ogImage()} />

      {/* ── Página ───────────────────────────────────────────────────────── */}
      <main class="min-h-screen bg-colpsi-bg pb-20 font-sans">

        {/* Modales */}
        <PostGradeModal
          postGrade={selectedPostGrade()}
          onClose={() => setSelectedPostGrade(null)}
        />
        
        <FullBioModal
          isOpen={showBioModal()}
          onClose={() => {
            setShowBioModal(false);
            document.body.style.overflow = "auto";
          }}
          content={profileData()?.full_bio_content}
          psychologistName={fullName()}
        />

        {/* Barra superior navegable */}
        <div class="bg-colpsi-blue py-6 px-4 shadow-md sticky top-0 z-40">
          <div class="max-w-5xl mx-auto flex items-center justify-between">
            <A href="/directorio" class="text-white hover:text-colpsi-yellow font-bold flex items-center gap-2 transition-colors text-sm">
              <span>←</span> Volver al Directorio
            </A>
            <span class="text-blue-200 text-xs md:text-sm font-black tracking-widest uppercase">
              Ficha Profesional
            </span>
          </div>
        </div>

        <div class="max-w-5xl mx-auto px-4 mt-4 md:mt-8">
          <Show when={profile.loading}>
            <ProfileSkeleton />
          </Show>

          <Show when={!profile.loading && profileData() === null}>
            <NotFound />
          </Show>

          <Show when={profileData()}>
            {(psi) => {
              const sortedPostGrades = sortPostGradesByYear(psi().post_grades);
              return (
                <div class="flex flex-col lg:grid lg:grid-cols-3 gap-6">
                  <div class="space-y-4">
                    <ProfileHeader
                      url={canonicalUrl}
                      firstName={psi().first_name}
                      secondName={psi().second_name}
                      lastName={psi().last_name}
                      secondLastName={psi().second_last_name}
                      fpv={psi().fpv}
                      ci={psi().ci}
                      profilePicture={psi().profile_picture}
                      specialties={psi().specialties}
                    />
                  </div>

                  <div class="lg:col-span-2 space-y-4">
                    <ContactCard
                      email={psi().email}
                      location={psi().location}
                      socialNetworks={psi().social_networks}
                    />
                    <Show when={psi().mini_bio || psi().full_bio_content}>
                      <div class="bg-white rounded-3xl p-6 md:p-8 shadow-premium border border-gray-100">
                        <h3 class="text-xs font-black text-colpsi-blue uppercase tracking-widest mb-3 border-l-4 border-colpsi-yellow pl-3">
                          Perfil Profesional
                        </h3>
                        <Show when={psi().mini_bio}>
                          <p class="text-gray-700 text-sm md:text-base leading-relaxed whitespace-pre-wrap font-medium">
                            {psi().mini_bio}
                          </p>
                        </Show>
                        <Show when={psi().full_bio_content && psi().full_bio_content !== "<p></p>"}>
                          <button
                            onClick={() => setShowBioModal(true)}
                            class="mt-6 inline-flex items-center gap-2 bg-colpsi-blue/5 hover:bg-colpsi-blue hover:text-white text-colpsi-blue font-bold px-5 py-2.5 rounded-xl transition-all active:scale-95 text-sm"
                          >
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                            </svg>
                            Leer biografía completa
                          </button>
                        </Show>
                      </div>
                    </Show>

                    <AcademicSection
                      undergraduate={psi().undergraduate}
                      postGrades={sortedPostGrades}
                      onPostGradeClick={setSelectedPostGrade}
                    />
                  </div>
                </div>
              );
            }}
          </Show>
        </div>
      </main>
    </>
  );
}

// ── COMPONENTES INTERNOS ─────────────────────────────────────────────

function ProfileSkeleton() {
  return (
    <div class="flex flex-col lg:grid lg:grid-cols-3 gap-6 animate-pulse">
      <div class="bg-white rounded-3xl h-96 shadow-sm border border-gray-100" />
      <div class="lg:col-span-2 space-y-4">
        <div class="bg-white rounded-3xl h-32 shadow-sm border border-gray-100" />
        <div class="bg-white rounded-3xl h-32 shadow-sm border border-gray-100" />
        <div class="bg-white rounded-3xl h-64 shadow-sm border border-gray-100" />
      </div>
    </div>
  );
}

function NotFound() {
  return (
    <div class="text-center py-20 bg-white rounded-3xl shadow-sm border border-gray-100">
      <Meta name="robots" content="noindex, nofollow" />
      <div class="text-6xl mb-4">🛡️</div>
      <h2 class="text-2xl font-black text-colpsi-blue uppercase tracking-tight">Perfil Protegido</h2>
      <p class="text-colpsi-muted mt-2 max-w-md mx-auto font-medium px-6">
        El perfil solicitado es privado, el profesional no se encuentra solvente o el registro no existe en nuestra base de datos oficial.
      </p>
      <A href="/directorio" class="mt-8 inline-flex items-center gap-2 bg-colpsi-blue text-white px-6 py-3 rounded-xl font-bold hover:bg-blue-800 transition-all shadow-lg">
        Volver al Directorio
      </A>
    </div>
  );
}