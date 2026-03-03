// web/src/routes/directorio/[id].tsx
/**
 * Página de perfil público del psicólogo. Aquí se muestra la información básica, especialidades, mini biografía y datos de contacto autorizados por el profesional.
 * 
 * FIX SENIOR: Esta página es accesible para cualquier usuario que haga clic en un resultado del directorio. No requiere autenticación, pero sí debe mostrar solo la información que el psicólogo ha autorizado como pública.
 * El endpoint /psi/{id} en Go se encarga de devolver solo los campos públicos del perfil para garantizar la privacidad.
 */
// web/src/routes/directorio/[id].tsx
// Página principal simplificada
import { createResource, createSignal, Show, Suspense } from "solid-js";
import { useParams, A } from "@solidjs/router";
import { apiGet } from "~/lib/api";
import { PsiProfile } from "~/types/psi";
import { ProfileHeader } from "~/components/psi/ProfileHeader";
import { ContactCard } from "~/components/psi/ContactCard";
import { AcademicSection } from "~/components/psi/AcademicSection";
import { PostGradeModal } from "~/components/psi/PostGradeModal";
import { sortPostGradesByYear } from "~/lib/utils";

export default function PsiProfilePage() {
  const params = useParams();
  const [selectedPostGrade, setSelectedPostGrade] = createSignal(null);

  const [profile] = createResource(() => 
    apiGet<PsiProfile>(`/psi/${params.id}`)
  );

  return (
    <main class="min-h-screen bg-[#f5f5f5] pb-20 font-sans">
      <PostGradeModal 
        postGrade={selectedPostGrade()} 
        onClose={() => setSelectedPostGrade(null)} 
      />

      <div class="bg-colpsi-blue py-6 px-4 shadow-md sticky top-0 z-40">
        <div class="max-w-5xl mx-auto flex items-center justify-between">
          <A href="/directorio" class="text-white hover:text-colpsi-yellow font-bold flex items-center gap-2 transition-colors text-sm md:text-base">
            <span>←</span> Volver al Directorio
          </A>
          <span class="text-blue-200 text-xs md:text-sm font-medium tracking-widest uppercase">
            Ficha Técnica
          </span>
        </div>
      </div>

      <div class="max-w-5xl mx-auto px-4 mt-4 md:mt-8">
        <Suspense fallback={<ProfileSkeleton />}>
          <Show when={profile()} fallback={<NotFound />}>
            {(psi) => {
              const sortedPostGrades = sortPostGradesByYear(psi().post_grades);
              
              return (
                <div class="flex flex-col lg:grid lg:grid-cols-3 gap-6">
                  <div class="space-y-4">
                    <ProfileHeader 
                      firstName={psi().first_name}
                      lastName={psi().last_name}
                      fpv={psi().fpv}
                      profilePicture={psi().profile_picture}
                      solvent={psi().solvent}
                      specialties={psi().specialties}
                    />
                    <ContactCard 
                      email={psi().email}
                      phone={psi().phone}
                      location={psi().location}
                      socialNetworks={psi().social_networks}
                    />
                  </div>

                  <div class="lg:col-span-2 space-y-4">
                    <Show when={psi().mini_bio}>
                      <div class="bg-white rounded-3xl p-6 shadow-sm border border-gray-100">
                        <h3 class="text-xs md:text-sm font-black text-colpsi-blue uppercase tracking-widest mb-3">
                          Perfil Profesional
                        </h3>
                        <p class="text-gray-700 text-sm md:text-base leading-relaxed whitespace-pre-wrap">
                          {psi().mini_bio}
                        </p>
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
        </Suspense>
      </div>
    </main>
  );
}

function ProfileSkeleton() {
  return (
    <div class="flex flex-col lg:grid lg:grid-cols-3 gap-6 animate-pulse">
      <div class="bg-white rounded-3xl h-96 shadow-sm border border-gray-100"></div>
      <div class="lg:col-span-2 space-y-4">
        <div class="bg-white rounded-3xl h-32 shadow-sm border border-gray-100"></div>
        <div class="bg-white rounded-3xl h-64 shadow-sm border border-gray-100"></div>
      </div>
    </div>
  );
}

function NotFound() {
  return (
    <div class="text-center py-20">
      <h2 class="text-2xl font-black text-colpsi-blue">Perfil no encontrado</h2>
      <p class="text-colpsi-muted mt-2">El profesional no existe o su perfil es privado.</p>
    </div>
  );
}