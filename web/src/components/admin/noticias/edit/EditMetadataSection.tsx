// web/src/components/admin/noticias/edit/EditMetadataSection.tsx
import { Show } from "solid-js";
import { Accessor, Setter } from "solid-js";
import { PostStatus, STATUS_OPTIONS, IC, labelClass } from "./types";

interface Props {
  title: Accessor<string>;
  setTitle: Setter<string>;
  shortDescription: Accessor<string>;
  setShortDescription: Setter<string>;
  type: Accessor<"public" | "psi">;
  setType: Setter<"public" | "psi">;
  status: Accessor<PostStatus>;
  setStatus: Setter<PostStatus>;
  publishAt: Accessor<string>;
  setPublishAt: Setter<string>;
}

export function EditMetadataSection(props: Props) {
  return (
    <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-colpsi-border space-y-5">
      <h2 class="text-sm font-black text-blue-800 uppercase tracking-widest border-b border-colpsi-border pb-3">
        Información General
      </h2>

      {/* Título */}
      <div>
        <label class={labelClass}>Título <span class="text-red-400">*</span></label>
        <input
          type="text"
          required
          maxLength={100}
          value={props.title()}
          onInput={(e) => props.setTitle(e.currentTarget.value)}
          class={IC}
        />
        <p class="text-[10px] text-gray-400 mt-1 text-right">{props.title().length}/100</p>
      </div>

      {/* Resumen */}
      <div>
        <label class={labelClass}>Resumen (snippet del feed)</label>
        <textarea
          rows={2}
          maxLength={250}
          value={props.shortDescription()}
          onInput={(e) => props.setShortDescription(e.currentTarget.value)}
          class={`${IC} resize-none`}
        />
        <p class="text-[10px] text-gray-400 mt-1 text-right">{props.shortDescription().length}/250</p>
      </div>

      {/* Audiencia y Estado */}
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
        {/* Audiencia */}
        <div>
          <label class={labelClass}>Audiencia</label>
          <div class="flex gap-3 mt-1">
            {(["public", "psi"] as const).map((t) => (
              <button
                type="button"
                onClick={() => props.setType(t)}
                class={`flex-1 py-2.5 rounded-xl text-xs font-black uppercase tracking-wide border-2 transition-all ${
                  props.type() === t
                    ? t === "public"
                      ? "bg-emerald-600 text-white border-emerald-600"
                      : "bg-blue-700 text-white border-blue-700"
                    : "bg-white text-gray-500 border-gray-200 hover:border-gray-300"
                }`}
              >
                {t === "public" ? "🌐 Público" : "🔒 Colegiados"}
              </button>
            ))}
          </div>
          <p class="text-[10px] text-gray-400 mt-1 ml-1">
            {props.type() === "public" 
              ? "Visible para cualquier visitante." 
              : "Solo psicólogos con sesión iniciada."}
          </p>
        </div>

        {/* Estado */}
        <div>
          <label class={labelClass}>Estado</label>
          <div class="grid grid-cols-2 gap-2 mt-1">
            {STATUS_OPTIONS.map((opt) => (
              <button
                type="button"
                onClick={() => props.setStatus(opt.value)}
                class={`py-2 rounded-xl text-xs font-black uppercase tracking-wide border-2 transition-all ${
                  props.status() === opt.value
                    ? "bg-blue-800 text-white border-blue-800"
                    : "bg-white text-gray-500 border-gray-200 hover:border-gray-300"
                }`}
              >
                {opt.icon} {opt.label}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Fecha de publicación — solo si status === scheduled */}
      <Show when={props.status() === "scheduled"}>
        <div>
          <label class={labelClass}>
            Fecha de publicación <span class="text-red-400">*</span>
          </label>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class={labelClass}>Fecha</label>
              <input
                type="date"
                value={props.publishAt().split("T")[0] ?? ""}
                onInput={(e) => {
                  const time = props.publishAt().split("T")[1] || "08:00";
                  props.setPublishAt(`${e.currentTarget.value}T${time}`);
                }}
                class={IC}
              />
            </div>
            <div>
              <label class={labelClass}>Hora</label>
              <select
                value={props.publishAt().split("T")[1]?.slice(0, 5) ?? "08:00"}
                onChange={(e) => {
                  const date = props.publishAt().split("T")[0] || new Date().toISOString().split("T")[0];
                  props.setPublishAt(`${date}T${e.currentTarget.value}`);
                }}
                class={IC}
              >
                {Array.from({ length: 24 }, (_, h) =>
                  ["00", "30"].map((m) => {
                    const val = `${String(h).padStart(2, "0")}:${m}`;
                    return <option value={val}>{val}</option>;
                  })
                )}
              </select>
            </div>
          </div>
          <p class="text-[10px] text-gray-400 mt-1 ml-1">
            El post se publicará automáticamente en esta fecha y hora.
          </p>
        </div>
      </Show>
    </section>
  );
}