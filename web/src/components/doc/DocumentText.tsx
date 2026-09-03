// Renderiza el texto de un artículo/bloque separando los puntos (•) y párrafos
// para una lectura más limpia. Es contenido propio y estático (sin innerHTML).

import { For } from "solid-js";

type Props = {
  text: string;
};

export default function DocumentText(props: Props) {
  const paragraphs = () =>
    props.text
      .split(/\n+/)
      .map((p) => p.trim())
      .filter(Boolean);

  return (
    <div class="space-y-3">
      <For each={paragraphs()}>
        {(p) => {
          if (p.includes("•")) {
            const items = p
              .split(/•/)
              .map((s) => s.trim())
              .filter(Boolean);
            return (
              <ul class="space-y-2 pl-1 list-none">
                <For each={items}>
                  {(it) => (
                    <li class="flex gap-3 text-justify leading-relaxed">
                      <span class="text-colpsi-yellow font-black">•</span>
                      <span class="flex-1 text-gray-700">{it}</span>
                    </li>
                  )}
                </For>
              </ul>
            );
          }
          return <p class="text-justify text-gray-700 leading-relaxed">{p}</p>;
        }}
      </For>
    </div>
  );
}
