// web/src/components/psi/ContactCard.tsx
// Tarjeta de contacto
import { Show, For } from "solid-js";
import { SocialNetwork, Location } from "~/types/psi";

interface ContactCardProps {
  email?: string;
  phone?: string;
  location: Location;
  socialNetworks?: SocialNetwork[];
}

export function ContactCard(props: ContactCardProps) {
  const hasContactInfo = () => props.email || props.phone || props.socialNetworks?.length;

  return (
    <Show when={hasContactInfo()}>
      <div class="bg-white rounded-3xl p-6 shadow-sm border border-gray-100 space-y-4">
        <h3 class="text-xs md:text-sm font-black text-colpsi-blue uppercase tracking-widest border-b border-gray-100 pb-2 mb-3">
          Contacto
        </h3>
        
        <Show when={props.email}>
          <div class="flex items-center gap-3 text-xs md:text-sm">
            <span class="text-colpsi-yellow text-base md:text-lg">✉️</span>
            <a href={`mailto:${props.email}`} class="text-gray-600 hover:text-colpsi-blue transition-colors break-all">
              {props.email}
            </a>
          </div>
        </Show>
        
        <Show when={props.phone}>
          <div class="flex items-center gap-3 text-xs md:text-sm">
            <span class="text-colpsi-yellow text-base md:text-lg">📞</span>
            <a href={`tel:${props.phone}`} class="text-gray-600 hover:text-colpsi-blue transition-colors">
              {props.phone}
            </a>
          </div>
        </Show>

        <Show when={props.location.municipality}>
          <div class="flex items-center gap-3 text-xs md:text-sm">
            <span class="text-colpsi-yellow text-base md:text-lg">📍</span>
            <span class="text-gray-600">{props.location.municipality}, {props.location.state}</span>
          </div>
        </Show>

        <Show when={props.socialNetworks?.length}>
          <div class="pt-3 mt-3 border-t border-gray-100 flex flex-wrap gap-2">
            <For each={props.socialNetworks}>
              {(net) => (
                <a 
                  href={net.url} 
                  target="_blank" 
                  rel="noopener noreferrer" 
                  class="text-[10px] md:text-xs bg-gray-50 text-colpsi-blue font-bold px-2 md:px-3 py-1.5 rounded-lg hover:bg-colpsi-yellow transition-colors"
                >
                  {net.name}
                </a>
              )}
            </For>
          </div>
        </Show>
      </div>
    </Show>
  );
}