// web/src/routes/directorio/[slug].tsx
/**
 * Página de perfil público del psicólogo con URLs amigables.
 * Formato: nombre-apellido(s)-fpv-1234
 * El backend busca por FPV
 */
import { createEffect, createResource, createSignal, Show, Suspense } from "solid-js";
import { useParams, A } from "@solidjs/router";
import { apiGet } from "~/lib/api";
import { PsiProfile } from "~/types/psi";
import { ProfileHeader } from "~/components/psi/ProfileHeader";
import { ContactCard } from "~/components/psi/ContactCard";
import { AcademicSection } from "~/components/psi/AcademicSection";
import { PostGradeModal } from "~/components/psi/PostGradeModal";
import { sortPostGradesByYear, extractFpvFromSlug, createProfileSlug } from "~/lib/utils";
import { FullBioModal } from "~/components/directory/FullBioModal";

export default function PsiProfilePage() {
  const params = useParams();
  const[selectedPostGrade, setSelectedPostGrade] = createSignal(null);
  
  // NUEVO: Estado para controlar el modal de la biografía extensa
  const [showBioModal, setShowBioModal] = createSignal(false);
  
  const fpv = () => extractFpvFromSlug(params.slug ?? "");

  const [profile] = createResource(
    () => fpv(),
    async (fpvNumber) => {
      if (!fpvNumber) throw new Error("FPV no válido");
      return await apiGet<PsiProfile>(`/psi/${fpvNumber}`);
    }
  );

  createEffect(() => {
    const p = profile();
    if (p) {
      const correctSlug = createProfileSlug({
        first_name: p.first_name,
        second_name: p.second_name,
        last_name: p.last_name,
        second_last_name: p.second_last_name,
        fpv: p.fpv
      });
      
      if (params.slug !== correctSlug) {
        console.log("Slug actual:", params.slug);
        console.log("Slug correcto:", correctSlug);
      }
    }
  });


  createEffect(() => {
    const p = profile();
    console.log(p)
  });

  return (
    <main class="min-h-screen bg-[#f5f5f5] pb-20 font-sans">
      
      {/* MODALES FLOTANTES */}
      <PostGradeModal 
        postGrade={selectedPostGrade()} 
        onClose={() => setSelectedPostGrade(null)} 
      />
      
      <FullBioModal 
        isOpen={showBioModal()} 
        onClose={() => {
            setShowBioModal(false);
            document.body.style.overflow = "auto"; // Restaurar scroll al cerrar
        }} 
        content={profile()?.full_bio_content}
        psychologistName={`${profile()?.first_name || ''} ${profile()?.last_name || ''}`}
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
                      secondName={psi().second_name}
                      lastName={psi().last_name}
                      secondLastName={psi().second_last_name}
                      fpv={psi().fpv}
                      ci={psi().ci}
                      profilePicture={psi().profile_picture}
                      solvent={psi().solvent}
                      specialties={psi().specialties}
                    />
                    <ContactCard 
                      email={psi().email}
                      phone={psi().phone}
                      location={psi().location}
                      socialNetworks={psi().social_networks}
                      service_address={psi().address}
                    />
                  </div>

                  <div class="lg:col-span-2 space-y-4">
                    {/* SECCIÓN BIO (Mini Bio + Botón para abrir la Full Bio) */}
                    <Show when={psi().mini_bio || psi().full_bio_content}>
                      <div class="bg-white rounded-3xl p-6 md:p-8 shadow-premium border border-gray-100">
                        <h3 class="text-xs md:text-sm font-black text-colpsi-blue uppercase tracking-widest mb-3 border-l-4 border-colpsi-yellow pl-3">
                          Perfil Profesional (Resumen)
                        </h3>
                        
                        <Show when={psi().mini_bio}>
                          <p class="text-gray-700 text-sm md:text-base leading-relaxed whitespace-pre-wrap font-medium">
                            {psi().mini_bio}
                          </p>
                        </Show>

                        {/* BOTÓN MAGICO: Solo aparece si el psicólogo escribió una biografía extensa (Tiptap) */}
                        <Show when={psi().full_bio_content && psi().full_bio_content !== "<p></p>"}>
                          <button 
                            onClick={() => setShowBioModal(true)}
                            class="mt-6 inline-flex items-center gap-2 bg-colpsi-blue/5 hover:bg-colpsi-blue hover:text-white text-colpsi-blue font-bold px-5 py-2.5 rounded-xl transition-all active:scale-95 text-sm"
                          >
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>
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
        </Suspense>
      </div>
    </main>
  );
}

// ... ProfileSkeleton y NotFound se mantienen igual ...
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