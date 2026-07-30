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
      class="bg-white border-2 border-transparent hover:border-colpsi-yellow p-6 rounded-[2.5rem] shadow-sm hover:shadow-2xl hover:-translate-y-1 transition-all group flex flex-col"
    >
      <div class="flex items-center gap-4 mb-4">
        <div class="w-16 h-16 bg-colpsi-bg rounded-2xl overflow-hidden flex-shrink-0 border-2 border-gray-50 group-hover:border-colpsi-yellow transition-colors">
          <Show 
            when={psychologist.profile_picture} 
            fallback={<div class="flex h-full items-center justify-center text-3xl">👤</div>}
          >
            <img 
              src={bucketUrl(psychologist.profile_picture)} 
              alt={fullName}
              class="w-full h-full object-cover" 
            />
          </Show>
        </div>
        <div>
          <h3 class="text-colpsi-blue font-black leading-tight group-hover:underline uppercase text-sm">
            {fullName}
          </h3>
          <p class="text-[10px] font-black text-gray-400 uppercase tracking-widest mt-0.5">
            FPV: {psychologist.fpv}
          </p>
        </div>
      </div>

      <div class="text-colpsi-text text-xs leading-relaxed italic mb-4 bg-gray-50 p-3 rounded-2xl">
        <span class="block whitespace-pre-wrap break-words">
          "{psychologist.mini_bio || 'Profesional federado del estado Carabobo.'}"
        </span>
        <Show when={psychologist.mini_bio && psychologist.mini_bio.length > 200}>
          <span class="text-[8px] text-gray-400 mt-1 block text-right">
            {psychologist.mini_bio.length}/250
          </span>
        </Show>
      </div>

      <div class="mt-auto flex flex-wrap gap-1">
        <For each={psychologist.specialties}>
          {(s) => (
            <span class="text-[9px] bg-blue-50 text-colpsi-blue font-bold px-2.5 py-1 rounded-lg uppercase tracking-tighter">
              {s}
            </span>
          )}
        </For>
      </div>
    </A>
  );
}