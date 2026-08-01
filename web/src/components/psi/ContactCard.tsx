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

type LocationBlock = {
  title: string;
  icon: string;
  lines: string[];
  phone?: string;
  cell?: string;
};

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

  const blocks = (): LocationBlock[] => {
    const result: LocationBlock[] = [];

    if (props.location?.carabobo) {
      const loc = props.location.carabobo;
      result.push({
        title: "Carabobo",
        icon: "📍",
        lines: [
          loc.municipality ? `${loc.municipality}, Carabobo` : "Carabobo",
          loc.address,
        ].filter(Boolean),
        phone: loc.phone,
        cell: loc.cell_phone,
      });
    }

    if (props.location?.venezuela) {
      const loc = props.location.venezuela;
      result.push({
        title: loc.state || "Venezuela",
        icon: "📍",
        lines: [
          loc.municipality ? `${loc.municipality}, ${loc.state}` : loc.state,
          loc.address,
        ].filter(Boolean),
        phone: loc.phone,
        cell: loc.cell_phone,
      });
    }

    if (props.location?.exterior) {
      const loc = props.location.exterior;
      result.push({
        title: loc.country || "Exterior",
        icon: "🌎",
        lines: [loc.address].filter(Boolean),
        phone: loc.phone,
        cell: loc.cell_phone,
      });
    }

    return result;
  };

  return (
    <Show when={hasContactInfo()}>
      <div class="bg-white rounded-3xl p-6 md:p-8 shadow-premium border border-gray-100">
        <h3 class="text-xs md:text-sm font-black text-colpsi-blue uppercase tracking-widest border-b-2 border-gray-50 pb-4 mb-5 flex items-center gap-2">
          <span class="text-lg">📇</span> Contacto
        </h3>

        {/* ── Contactos principales ──────────────────────────────────── */}
        <div class="flex flex-wrap gap-2.5 mb-6">
          <Show when={props.email}>
            <a
              href={`mailto:${props.email}`}
              class="inline-flex items-center gap-2 bg-gray-50 hover:bg-colpsi-blue/5 text-gray-700 text-xs md:text-sm font-bold px-3 py-2 rounded-xl transition-colors break-all"
            >
              <span class="text-base">✉️</span> {props.email}
            </a>
          </Show>
          <Show when={props.phone}>
            <a
              href={`tel:${props.phone}`}
              class="inline-flex items-center gap-2 bg-gray-50 hover:bg-colpsi-blue/5 text-gray-700 text-xs md:text-sm font-bold px-3 py-2 rounded-xl transition-colors"
            >
              <span class="text-base">📞</span> {props.phone}
            </a>
          </Show>
          <Show when={props.address}>
            <span class="inline-flex items-center gap-2 bg-gray-50 text-gray-700 text-xs md:text-sm font-bold px-3 py-2 rounded-xl">
              <span class="text-base">🏢</span> {props.address}
            </span>
          </Show>
        </div>

        {/* ── Ubicaciones (grid compacto) ────────────────────────────── */}
        <Show when={hasAnyLocation()}>
          <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <For each={blocks()}>
              {(block) => (
                <div class="bg-gray-50/60 border border-gray-50 rounded-2xl p-4 flex flex-col gap-2">
                  <p class="text-[10px] font-black text-gray-400 uppercase tracking-widest flex items-center gap-1.5">
                    <span>{block.icon}</span> {block.title}
                  </p>
                  <div class="space-y-1">
                    <For each={block.lines}>
                      {(line) => (
                        <p class="text-xs md:text-sm text-gray-600 leading-snug">{line}</p>
                      )}
                    </For>
                  </div>
                  {(block.phone || block.cell) && (
                    <div class="flex flex-wrap gap-1.5 mt-auto pt-2">
                      <Show when={block.phone}>
                        <a
                          href={`tel:${block.phone}`}
                          class="inline-flex items-center gap-1 text-[10px] md:text-xs bg-white text-colpsi-blue font-bold px-2 py-1 rounded-lg border border-blue-100 hover:bg-colpsi-blue hover:text-white transition-colors"
                        >
                          📞 {block.phone}
                        </a>
                      </Show>
                      <Show when={block.cell}>
                        <a
                          href={`tel:${block.cell}`}
                          class="inline-flex items-center gap-1 text-[10px] md:text-xs bg-white text-colpsi-blue font-bold px-2 py-1 rounded-lg border border-blue-100 hover:bg-colpsi-blue hover:text-white transition-colors"
                        >
                          📱 {block.cell}
                        </a>
                      </Show>
                    </div>
                  )}
                </div>
              )}
            </For>
          </div>
        </Show>

        {/* ── Redes sociales ─────────────────────────────────────────── */}
        <Show when={props.socialNetworks?.length}>
          <div class="mt-5 pt-4 border-t border-gray-100 flex flex-wrap gap-2">
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
