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
  tag: string;
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

  // Colores del badge según el tipo de ubicación
  const tagClasses = (tag: string) => {
    switch (tag) {
      case "Carabobo":
        return "bg-blue-50 text-blue-700 border-blue-100";
      case "Venezuela":
        return "bg-indigo-50 text-indigo-700 border-indigo-100";
      default:
        return "bg-emerald-50 text-emerald-700 border-emerald-100";
    }
  };

  const blocks = (): LocationBlock[] => {
    const result: LocationBlock[] = [];

    if (props.location?.carabobo) {
      const loc = props.location.carabobo;
      result.push({
        tag: "Carabobo",
        icon: "📍",
        title: loc.municipality ? `${loc.municipality}, Carabobo` : "Carabobo",
        lines: [loc.address].filter(Boolean),
        phone: loc.phone,
        cell: loc.cell_phone,
      });
    }

    if (props.location?.venezuela) {
      const loc = props.location.venezuela;
      result.push({
        tag: "Venezuela",
        icon: "📍",
        title: loc.municipality ? `${loc.municipality}, ${loc.state}` : loc.state || "Venezuela",
        lines: [loc.address].filter(Boolean),
        phone: loc.phone,
        cell: loc.cell_phone,
      });
    }

    if (props.location?.exterior) {
      const loc = props.location.exterior;
      result.push({
        tag: "Internacional",
        icon: "🌎",
        title: loc.country || "Exterior",
        lines: [loc.address].filter(Boolean),
        phone: loc.phone,
        cell: loc.cell_phone,
      });
    }

    return result;
  };

  return (
    <Show when={hasContactInfo()}>
      <div class="bg-white rounded-3xl p-6 md:p-8 shadow-premium border border-colpsi-border">
        <h3 class="text-sm md:text-lg font-black text-colpsi-blue uppercase tracking-widest border-b-2 border-gray-50 pb-4 mb-5 flex items-center gap-2">
          <span class="text-2xl">📇</span> Contacto
        </h3>

        {/* ── Contactos principales ──────────────────────────────────── */}
        <div class="flex flex-wrap gap-2.5 mb-6">
          <Show when={props.email}>
            <a
              href={`mailto:${props.email}`}
              class="inline-flex items-center gap-2 bg-colpsi-surface hover:bg-colpsi-blue/5 text-gray-700 text-sm md:text-base font-bold px-3 py-2 rounded-xl transition-colors break-all"
            >
              <span class="text-xl">✉️</span> {props.email}
            </a>
          </Show>
          <Show when={props.phone}>
            <a
              href={`tel:${props.phone}`}
              class="inline-flex items-center gap-2 bg-colpsi-surface hover:bg-colpsi-blue/5 text-gray-700 text-sm md:text-base font-bold px-3 py-2 rounded-xl transition-colors"
            >
              <span class="text-xl">📞</span> {props.phone}
            </a>
          </Show>
          <Show when={props.address}>
            <span class="inline-flex items-center gap-2 bg-colpsi-surface text-gray-700 text-sm md:text-base font-bold px-3 py-2 rounded-xl">
              <span class="text-xl">🏢</span> {props.address}
            </span>
          </Show>
        </div>

        {/* ── Ubicaciones (una debajo de la otra con su tipo) ─────────── */}
        <Show when={hasAnyLocation()}>
          <div class="flex flex-col gap-4">
            <For each={blocks()}>
              {(block) => (
                <div class="bg-colpsi-surface/60 border border-gray-50 rounded-2xl p-4 flex flex-col gap-2">
                  <div class="flex items-center justify-between gap-2 flex-wrap">
                    <p class="text-sm md:text-base font-black text-colpsi-blue uppercase tracking-wide flex items-center gap-2">
                      <span class="text-lg">{block.icon}</span> {block.title}
                    </p>
                    <span
                      class={`text-[11px] font-black uppercase tracking-widest px-2.5 py-1 rounded-full border ${tagClasses(block.tag)}`}
                    >
                      {block.tag}
                    </span>
                  </div>
                  <div class="space-y-1">
                    <For each={block.lines}>
                      {(line) => (
                        <p class="text-sm md:text-base text-gray-600 leading-snug">{line}</p>
                      )}
                    </For>
                  </div>
                  {(block.phone || block.cell) && (
                    <div class="flex flex-wrap gap-1.5 mt-auto pt-2">
                      <Show when={block.phone}>
                        <a
                          href={`tel:${block.phone}`}
                          class="inline-flex items-center gap-1 text-xs md:text-sm bg-white text-colpsi-blue font-bold px-2 py-1 rounded-lg border border-blue-100 hover:bg-colpsi-blue hover:text-white transition-colors"
                        >
                          📞 {block.phone}
                        </a>
                      </Show>
                      <Show when={block.cell}>
                        <a
                          href={`tel:${block.cell}`}
                          class="inline-flex items-center gap-1 text-xs md:text-sm bg-white text-colpsi-blue font-bold px-2 py-1 rounded-lg border border-blue-100 hover:bg-colpsi-blue hover:text-white transition-colors"
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
          <div class="mt-5 pt-4 border-t border-colpsi-border flex flex-wrap gap-2">
            <For each={props.socialNetworks}>
              {(net) => (
                <a
                  href={net.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-xs md:text-sm bg-colpsi-surface text-colpsi-blue font-bold px-2 md:px-3 py-1.5 rounded-lg hover:bg-colpsi-yellow transition-colors"
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
