// web/src/components/directory/PsychologistCard.tsx
import { Show, For } from "solid-js";
import { A } from "@solidjs/router";
import { bucketUrl } from "~/lib/bucket";
import { DirectoryPsychologist } from "~/types/psi";
import { createProfileSlug } from "~/lib/utils";

interface PsychologistCardProps {
  psychologist: DirectoryPsychologist;
}

export function PsychologistCard(props: PsychologistCardProps) {
  const { psychologist } = props;
  
  // Crear slug para la URL usando todos los nombres disponibles
  const profileSlug = createProfileSlug({
    first_name: psychologist.first_name,
    second_name: psychologist.second_name, // Puede ser undefined
    last_name: psychologist.last_name,
    second_last_name: psychologist.second_last_name, // Puede ser undefined
    fpv: psychologist.fpv
  });

  // Debug para ver qué slug se genera
  // console.log("Generando slug:", {
  //   nombres: {
  //     first_name: psychologist.first_name,
  //     second_name: psychologist.second_name,
  //     last_name: psychologist.last_name,
  //     second_last_name: psychologist.second_last_name
  //   },
  //   slug: profileSlug,
  //   fpv: psychologist.fpv
  // });

  // Nombre completo para mostrar
  const fullName = [
    psychologist.first_name,
    psychologist.second_name,
    psychologist.last_name,
    psychologist.second_last_name
  ].filter(Boolean).join(' ');

  return (
    <A 
      href={`/directorio/${profileSlug}`} 
      class="bg-white border-2 border-transparent hover:border-colpsi-yellow p-4 sm:p-5 rounded-[1.75rem] shadow-sm hover:shadow-2xl hover:-translate-y-1 transition-all group flex flex-col"
    >
      <div class="flex items-center gap-3 mb-3">
        <div class="w-14 h-14 sm:w-16 sm:h-16 bg-colpsi-bg rounded-xl overflow-hidden flex-shrink-0 border-2 border-gray-50 group-hover:border-colpsi-yellow transition-colors">
          <Show 
            when={psychologist.profile_picture} 
            fallback={<div class="flex h-full items-center justify-center text-2xl">👤</div>}
          >
            <img 
              src={bucketUrl(psychologist.profile_picture)} 
              alt={fullName}
              class="w-full h-full object-cover" 
              loading="lazy"
              decoding="async"
            />
          </Show>
        </div>
        <div class="min-w-0">
          <h3 class="text-colpsi-blue font-black leading-tight group-hover:underline uppercase text-[13px]">
            {fullName}
          </h3>
          <p class="text-[10px] font-black text-gray-400 uppercase tracking-widest mt-0.5">
            FPV: {psychologist.fpv}
          </p>
        </div>
      </div>

      <div class="text-colpsi-text text-xs leading-relaxed italic mb-3 bg-colpsi-surface p-2.5 rounded-xl line-clamp-3">
        <span class="block whitespace-pre-wrap break-words">
          "{psychologist.mini_bio || 'Profesional federado del estado Carabobo.'}"
        </span>
      </div>

      <Show when={psychologist.specialties.length > 0}>
        <div class="mt-auto flex flex-wrap gap-1">
          <For each={psychologist.specialties}>
            {(s) => (
              <span class="text-[10px] bg-blue-50 text-colpsi-blue font-bold px-2.5 py-1 rounded-lg uppercase tracking-tighter">
                {s}
              </span>
            )}
          </For>
        </div>
      </Show>

      <Show when={psychologist.service_modality}>
        <div class="mt-2 flex flex-wrap gap-1">
          <Show when={psychologist.service_modality!.presencial}>
            <span class="text-[10px] bg-teal-50 text-teal-700 font-bold px-2.5 py-1 rounded-lg uppercase tracking-tighter">
              Presencial
            </span>
          </Show>
          <Show when={psychologist.service_modality!.distance}>
            <span class="text-[10px] bg-teal-50 text-teal-700 font-bold px-2.5 py-1 rounded-lg uppercase tracking-tighter">
              A distancia
            </span>
          </Show>
          <Show when={psychologist.service_modality!.telephone}>
            <span class="text-[10px] bg-teal-50 text-teal-700 font-bold px-2.5 py-1 rounded-lg uppercase tracking-tighter">
              Telefónica
            </span>
          </Show>
        </div>
      </Show>
    </A>
  );
}