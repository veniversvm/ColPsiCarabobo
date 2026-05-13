// web/src/components/psi/ProfileHeader.tsx
// Cabecera del perfil con foto y datos básicos
import { Show, For } from "solid-js";
import QRCodeGenerator from "./profile/QrCode";

interface ProfileHeaderProps {
  firstName: string;
  secondName?: string;
  lastName: string;
  secondLastName?: string;
  fpv: number;
  ci: number;
  profilePicture?: string;
  specialties: string[];
  url: string;
}

export function ProfileHeader(props: ProfileHeaderProps) {
  return (
    <div class="bg-white rounded-3xl p-6 shadow-sm border border-gray-100 text-center relative overflow-hidden">
      {/* <Show when={props.solvent}>
        <div class="absolute top-3 right-3 bg-green-100 text-green-700 p-2 rounded-full shadow-sm text-sm" title="Miembro Solvente">
          ✓
        </div>
      </Show> */}

      <div class="w-24 h-24 md:w-32 md:h-32 mx-auto bg-gray-50 rounded-full overflow-hidden border-4 border-colpsi-yellow shadow-inner mb-4">
        <Show
          when={props.profilePicture}
          fallback={
            <div class="w-full h-full flex items-center justify-center text-4xl bg-blue-50">
              👤
            </div>
          }
        >
          <img
            src={`http://localhost:9000/colpsi-bucket/${props.profilePicture}`}
            alt={`Dr(a). ${props.lastName}`}
            class="w-full h-full object-cover"
          />
        </Show>
      </div>

      <h1 class="text-xl md:text-2xl font-black text-colpsi-blue leading-tight">
        {props.firstName} {props.secondName} {props.lastName}{" "}
        {props.secondLastName}
      </h1>
      <p class="text-gray-500 font-bold tracking-widest uppercase mt-1 text-xs md:text-sm">
        FPV: {props.fpv}
      </p>
      <p class="text-gray-500 font-bold tracking-widest uppercase mt-1 text-xs md:text-sm">
        CI: {props.ci}
      </p>

      <div class="mt-3 flex justify-center gap-2 flex-wrap">
        <For each={props.specialties}>
          {(spec) => (
            <span class="bg-blue-50 text-colpsi-blue text-[10px] md:text-xs font-bold px-2 md:px-3 py-1 rounded-lg">
              {spec}
            </span>
          )}
        </For>
      </div>

      <div class="mt-6 flex justify-center w-full">
        <QRCodeGenerator url={props.url} />
      </div>
    </div>
  );
}
