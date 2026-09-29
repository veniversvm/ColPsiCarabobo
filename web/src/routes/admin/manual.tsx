// web/src/routes/admin/manual.tsx
import { createSignal, For } from "solid-js";
import { A } from "@solidjs/router";
import ManualPdf, { ManualFile } from "~/components/manuales/ManualPdf";

// Ambos manuales accesibles desde el panel (el del psicólogo sirve para
// orientar a los agremiados). El PDF se sirve embebido en la API y solo con
// sesión válida — ver components/manuales/ManualPdf.tsx.
const MANUALES: { key: ManualFile; label: string; desc: string }[] = [
  {
    key: "manual-admin.pdf",
    label: "Panel de Administración",
    desc: "Módulos, permisos y operación diaria del panel administrativo.",
  },
  {
    key: "manual-psiuser.pdf",
    label: "Portal del Psicólogo",
    desc: "Vista del agremiado: útil para orientar a los colegiados.",
  },
];

export default function AdminManualPage() {
  const [file, setFile] = createSignal<ManualFile>("manual-admin.pdf");
  const activeManual = () => MANUALES.find((m) => m.key === file()) ?? MANUALES[0];

  return (
    <main class="space-y-4 pb-12 max-w-4xl mx-auto">
      {/* Cabecera */}
      <div>
        <A
          href="/admin"
          class="inline-flex items-center gap-1.5 text-sm font-medium text-colpsi-muted hover:text-colpsi-blue transition-colors"
        >
          ← Volver al panel
        </A>
        <h1 class="text-xl font-bold text-colpsi-text mt-2">Manuales</h1>
        <p class="text-sm text-colpsi-muted mt-1">
          Documentación oficial del sistema. Elige el manual y usa los botones para navegar o descargar
          el PDF.
        </p>
      </div>

      {/* Selector de manual */}
      <div class="flex flex-wrap gap-2">
        <For each={MANUALES}>
          {(m) => (
            <button
              type="button"
              onClick={() => setFile(m.key)}
              class={`rounded-lg px-4 py-2 text-sm font-semibold transition-colors border ${
                file() === m.key
                  ? "bg-colpsi-blue text-white border-colpsi-blue"
                  : "bg-white text-colpsi-muted border-colpsi-border hover:border-colpsi-blue hover:text-colpsi-blue"
              }`}
            >
              {m.label}
            </button>
          )}
        </For>
      </div>

      {/* Vista previa + descarga */}
      <section class="bg-white rounded-lg p-5 border border-colpsi-border">
        <p class="text-xs font-semibold text-colpsi-muted uppercase tracking-wide mb-3">
          {activeManual().desc}
        </p>
        <ManualPdf file={file()} height="min(72rem, calc(100vh - 9rem))" />
      </section>
    </main>
  );
}