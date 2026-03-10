// web/src/components/psi/ContactCard.tsx
import { Show, For } from "solid-js";
import { SocialNetwork, PsiLocation } from "~/types/psi";

interface ContactCardProps {
  email?: string;
  phone?: string;
  address?: string;
  location: PsiLocation;
  socialNetworks?: SocialNetwork[];
}

export function ContactCard(props: ContactCardProps) {
  const hasAnyLocation = () =>
    props.location?.carabobo ||
    props.location?.venezuela ||
    props.location?.exterior;

  const hasContactInfo = () =>
    props.email ||
    props.phone ||
    props.address ||
    hasAnyLocation() ||
    props.socialNetworks?.length;

  return (
    <Show when={hasContactInfo()}>
      <div class="bg-white rounded-3xl p-6 shadow-sm border border-gray-100 space-y-4">
        <h3 class="text-xs md:text-sm font-black text-colpsi-blue uppercase tracking-widest border-b border-gray-100 pb-2 mb-3">
          Contacto
        </h3>

        {/* ── Contacto principal ──────────────────────────────────────── */}
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

        <Show when={props.address}>
          <div class="flex items-center gap-3 text-xs md:text-sm">
            <span class="text-colpsi-yellow text-base md:text-lg">🏢</span>
            <span class="text-gray-600">{props.address}</span>
          </div>
        </Show>

        {/* ── Ubicación: Carabobo ─────────────────────────────────────── */}
        <Show when={props.location?.carabobo}>
          {(loc) => (
            <div class="pt-3 mt-1 border-t border-gray-50 space-y-2">
              <p class="text-[10px] font-black text-gray-400 uppercase tracking-widest">
                📍 Carabobo
              </p>
              <p class="text-xs md:text-sm text-gray-600 pl-1">
                {loc().municipality}, Carabobo
              </p>
              <Show when={loc().phone}>
                <div class="flex items-center gap-2 text-xs text-gray-500 pl-1">
                  <span>📞</span>
                  <a href={`tel:${loc().phone}`} class="hover:text-colpsi-blue transition-colors">{loc().phone}</a>
                </div>
              </Show>
              <Show when={loc().cell_phone}>
                <div class="flex items-center gap-2 text-xs text-gray-500 pl-1">
                  <span>📱</span>
                  <a href={`tel:${loc().cell_phone}`} class="hover:text-colpsi-blue transition-colors">{loc().cell_phone}</a>
                </div>
              </Show>
              <Show when={loc().address}>
                <p class="text-xs text-gray-500 pl-1">🏢 {loc().address}</p>
              </Show>
            </div>
          )}
        </Show>

        {/* ── Ubicación: Venezuela (fuera de Carabobo) ────────────────── */}
        <Show when={props.location?.venezuela}>
          {(loc) => (
            <div class="pt-3 mt-1 border-t border-gray-50 space-y-2">
              <p class="text-[10px] font-black text-gray-400 uppercase tracking-widest">
                📍 {loc().state}
              </p>
              <Show when={loc().municipality}>
                <p class="text-xs md:text-sm text-gray-600 pl-1">
                  {loc().municipality}, {loc().state}
                </p>
              </Show>
              <Show when={loc().phone}>
                <div class="flex items-center gap-2 text-xs text-gray-500 pl-1">
                  <span>📞</span>
                  <a href={`tel:${loc().phone}`} class="hover:text-colpsi-blue transition-colors">{loc().phone}</a>
                </div>
              </Show>
              <Show when={loc().cell_phone}>
                <div class="flex items-center gap-2 text-xs text-gray-500 pl-1">
                  <span>📱</span>
                  <a href={`tel:${loc().cell_phone}`} class="hover:text-colpsi-blue transition-colors">{loc().cell_phone}</a>
                </div>
              </Show>
              <Show when={loc().address}>
                <p class="text-xs text-gray-500 pl-1">🏢 {loc().address}</p>
              </Show>
            </div>
          )}
        </Show>

        {/* ── Ubicación: Exterior ─────────────────────────────────────── */}
        <Show when={props.location?.exterior}>
          {(loc) => (
            <div class="pt-3 mt-1 border-t border-gray-50 space-y-2">
              <p class="text-[10px] font-black text-gray-400 uppercase tracking-widest">
                🌎 {loc().country}
              </p>
              <Show when={loc().phone}>
                <div class="flex items-center gap-2 text-xs text-gray-500 pl-1">
                  <span>📞</span>
                  <a href={`tel:${loc().phone}`} class="hover:text-colpsi-blue transition-colors">{loc().phone}</a>
                </div>
              </Show>
              <Show when={loc().address}>
                <p class="text-xs text-gray-500 pl-1">🏢 {loc().address}</p>
              </Show>
            </div>
          )}
        </Show>

        {/* ── Redes sociales ──────────────────────────────────────────── */}
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