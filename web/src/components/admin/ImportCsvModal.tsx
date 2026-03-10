// web/src/components/admin/ImportCsvModal.tsx
import { createSignal, Show, For } from "solid-js";
import { apiPost } from "~/lib/api";

interface FailedRecord {
  fila: string;
  nombre: string;
  ci: string;
  fpv: string;
  error: string;
}

interface ImportResult {
  imported: number;
  failed: number;
  errors: FailedRecord[];
}

interface CsvPreview {
  rowCount: number;
  headers: string[];
  sample: string[][]; // primeras 3 filas
}

interface ImportCsvModalProps {
  onClose: () => void;
  onSuccess: () => void; // para refetch del listado
}

// Parsea el CSV en el cliente solo para preview — no envía datos aún
function parseCsvPreview(text: string): CsvPreview {
  const lines = text.trim().split("\n").filter(Boolean);
  if (lines.length === 0) return { rowCount: 0, headers: [], sample: [] };

  const parse = (line: string) =>
    line.split(",").map((v) => v.trim().replace(/^"|"$/g, ""));

  const headers = parse(lines[0]);
  const dataLines = lines.slice(1);
  const sample = dataLines.slice(0, 3).map(parse);

  return { rowCount: dataLines.length, headers, sample };
}

export function ImportCsvModal(props: ImportCsvModalProps) {
  // ── Estado ──────────────────────────────────────────────────────────────
  type Step = "select" | "preview" | "uploading" | "result";
  const [step, setStep] = createSignal<Step>("select");
  const [file, setFile] = createSignal<File | null>(null);
  const [preview, setPreview] = createSignal<CsvPreview | null>(null);
  const [result, setResult] = createSignal<ImportResult | null>(null);
  const [error, setError] = createSignal<string | null>(null);

  // ── Handlers ─────────────────────────────────────────────────────────────
  const handleFileChange = async (e: Event) => {
    const f = (e.currentTarget as HTMLInputElement).files?.[0] ?? null;
    if (!f) return;

    if (!f.name.endsWith(".csv")) {
      setError("El archivo debe ser un CSV (.csv).");
      return;
    }

    setFile(f);
    setError(null);

    // Leer y parsear localmente para el preview
    const text = await f.text();
    setPreview(parseCsvPreview(text));
    setStep("preview");
  };

  const handleUpload = async () => {
    const f = file();
    if (!f) return;

    setStep("uploading");
    setError(null);

    try {
      const fd = new FormData();
      fd.append("csv", f);

      const res = await apiPost<ImportResult>("/admin/psi/upload-csv", fd);
      setResult(res);
      setStep("result");

      // Si todo fue exitoso, notificar al padre para refetch
      if (res.failed === 0) props.onSuccess();
    } catch (err: any) {
      setError(err.message || "Error al procesar el archivo.");
      setStep("preview"); // volver a preview para reintentar
    }
  };

  const handleReset = () => {
    setFile(null);
    setPreview(null);
    setResult(null);
    setError(null);
    setStep("select");
  };

  // ── Render ────────────────────────────────────────────────────────────────
  return (
    <div
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
      onClick={(e) => { if (e.target === e.currentTarget) props.onClose(); }}
    >
      <div class="bg-white rounded-3xl shadow-2xl w-full max-w-2xl border border-gray-100 overflow-hidden animate-in zoom-in-95 duration-200">

        {/* ── HEADER ────────────────────────────────────────────────────── */}
        <div class="flex items-center justify-between px-8 py-6 border-b border-gray-100">
          <div>
            <h2 class="text-xl font-black text-blue-900 uppercase tracking-tight">
              Importar desde CSV
            </h2>
            <p class="text-gray-400 text-sm mt-0.5">
              {step() === "select" && "Selecciona el archivo para comenzar"}
              {step() === "preview" && `${preview()?.rowCount ?? 0} registros detectados`}
              {step() === "uploading" && "Procesando en el servidor..."}
              {step() === "result" && `Importación completada`}
            </p>
          </div>
          <button
            onClick={props.onClose}
            class="w-9 h-9 rounded-full bg-gray-100 hover:bg-gray-200 text-gray-500 font-bold flex items-center justify-center transition-colors text-lg leading-none"
          >
            ×
          </button>
        </div>

        {/* ── BODY ──────────────────────────────────────────────────────── */}
        <div class="px-8 py-6 max-h-[65vh] overflow-y-auto">

          {/* Error global */}
          <Show when={error()}>
            <div class="mb-5 p-4 rounded-2xl bg-red-50 text-red-800 font-bold text-sm border-l-4 border-red-500">
              {error()}
            </div>
          </Show>

          {/* ── STEP: SELECT ──────────────────────────────────────────── */}
          <Show when={step() === "select"}>
            <label class="flex flex-col items-center justify-center w-full h-52 border-2 border-dashed border-gray-300 rounded-2xl bg-gray-50 hover:bg-blue-50 hover:border-blue-300 transition-all cursor-pointer group">
              <div class="flex flex-col items-center gap-3 text-gray-400 group-hover:text-blue-500 transition-colors">
                <svg class="w-14 h-14" fill="none" stroke="currentColor" stroke-width="1.2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <span class="font-black text-base">Haz clic para seleccionar el archivo</span>
                <span class="text-sm">Solo archivos .csv</span>
              </div>
              <input type="file" accept=".csv" class="hidden" onChange={handleFileChange} />
            </label>

            <div class="mt-6 bg-blue-50 rounded-2xl p-4 border border-blue-100">
              <p class="text-blue-800 text-xs font-black mb-2">📋 Formato esperado del CSV:</p>
              <p class="text-blue-700 text-xs font-mono leading-relaxed">
                username, email, password, first_name, second_name, last_name, second_last_name, fpv, ci, ..., nationality, born_date, genre, contact_email, ...
              </p>
              <p class="text-blue-600 text-xs mt-2 font-medium">
                La primera fila debe ser la cabecera y será ignorada.
              </p>
            </div>
          </Show>

          {/* ── STEP: PREVIEW ─────────────────────────────────────────── */}
          <Show when={step() === "preview" && preview()}>
            {(p) => (
              <div class="space-y-5">

                {/* Resumen */}
                <div class="grid grid-cols-3 gap-3">
                  <div class="bg-blue-50 rounded-2xl p-4 text-center border border-blue-100">
                    <p class="text-3xl font-black text-blue-800">{p().rowCount}</p>
                    <p class="text-xs font-black text-blue-600 uppercase tracking-wide mt-1">Registros</p>
                  </div>
                  <div class="bg-gray-50 rounded-2xl p-4 text-center border border-gray-100">
                    <p class="text-3xl font-black text-gray-700">{p().headers.length}</p>
                    <p class="text-xs font-black text-gray-500 uppercase tracking-wide mt-1">Columnas</p>
                  </div>
                  <div class="bg-emerald-50 rounded-2xl p-4 text-center border border-emerald-100">
                    <p class="text-sm font-black text-emerald-700 truncate mt-1">{file()?.name}</p>
                    <p class="text-xs font-black text-emerald-600 uppercase tracking-wide mt-1">
                      {((file()?.size ?? 0) / 1024).toFixed(1)} KB
                    </p>
                  </div>
                </div>

                {/* Preview tabla */}
                <div>
                  <p class="text-xs font-black text-gray-500 uppercase tracking-widest mb-2">
                    Vista previa (primeras 3 filas)
                  </p>
                  <div class="overflow-x-auto rounded-xl border border-gray-200">
                    <table class="w-full text-xs">
                      <thead>
                        <tr class="bg-gray-50 border-b border-gray-200">
                          <For each={p().headers.slice(0, 8)}>
                            {(h) => (
                              <th class="px-3 py-2 text-left font-black text-gray-500 uppercase tracking-wide whitespace-nowrap">
                                {h}
                              </th>
                            )}
                          </For>
                          <Show when={p().headers.length > 8}>
                            <th class="px-3 py-2 text-gray-400 font-bold">+{p().headers.length - 8} más</th>
                          </Show>
                        </tr>
                      </thead>
                      <tbody class="divide-y divide-gray-100">
                        <For each={p().sample}>
                          {(row) => (
                            <tr class="hover:bg-gray-50">
                              <For each={row.slice(0, 8)}>
                                {(cell) => (
                                  <td class="px-3 py-2 text-gray-700 whitespace-nowrap max-w-[120px] truncate">
                                    {cell || <span class="text-gray-300 italic">vacío</span>}
                                  </td>
                                )}
                              </For>
                              <Show when={row.length > 8}>
                                <td class="px-3 py-2 text-gray-400">...</td>
                              </Show>
                            </tr>
                          )}
                        </For>
                      </tbody>
                    </table>
                  </div>
                </div>

                <div class="bg-amber-50 rounded-2xl p-4 border border-amber-100">
                  <p class="text-amber-800 text-xs font-bold">
                    ⚠️ Se crearán <span class="font-black">{p().rowCount} registros</span> en la base de datos. Los registros duplicados (mismo FPV o CI) fallarán individualmente sin afectar el resto.
                  </p>
                </div>
              </div>
            )}
          </Show>

          {/* ── STEP: UPLOADING ───────────────────────────────────────── */}
          <Show when={step() === "uploading"}>
            <div class="flex flex-col items-center justify-center py-16 gap-6">
              <div class="w-16 h-16 border-4 border-blue-200 border-t-blue-700 rounded-full animate-spin" />
              <div class="text-center">
                <p class="font-black text-gray-800 text-lg">Importando registros...</p>
                <p class="text-gray-500 text-sm mt-1">Esto puede tomar unos momentos según el tamaño del archivo.</p>
              </div>
            </div>
          </Show>

          {/* ── STEP: RESULT ──────────────────────────────────────────── */}
          <Show when={step() === "result" && result()}>
            {(r) => (
              <div class="space-y-5">

                {/* Resumen resultado */}
                <div class="grid grid-cols-2 gap-4">
                  <div class="bg-emerald-50 rounded-2xl p-5 text-center border border-emerald-100">
                    <p class="text-4xl font-black text-emerald-700">{r().imported}</p>
                    <p class="text-xs font-black text-emerald-600 uppercase tracking-wide mt-1">✓ Importados</p>
                  </div>
                  <div class={`rounded-2xl p-5 text-center border ${r().failed > 0 ? "bg-red-50 border-red-100" : "bg-gray-50 border-gray-100"}`}>
                    <p class={`text-4xl font-black ${r().failed > 0 ? "text-red-700" : "text-gray-400"}`}>{r().failed}</p>
                    <p class={`text-xs font-black uppercase tracking-wide mt-1 ${r().failed > 0 ? "text-red-600" : "text-gray-400"}`}>
                      {r().failed > 0 ? "✗ Fallidos" : "Sin errores"}
                    </p>
                  </div>
                </div>

                {/* Lista de errores */}
                <Show when={r().errors && r().errors.length > 0}>
                  <div>
                    <p class="text-xs font-black text-gray-500 uppercase tracking-widest mb-3">
                      Registros con error ({r().errors.length})
                    </p>
                    <div class="space-y-2 max-h-64 overflow-y-auto pr-1">
                      <For each={r().errors}>
                        {(err) => (
                          <div class="bg-red-50 rounded-xl p-4 border border-red-100">
                            <div class="flex items-start justify-between gap-3">
                              <div class="flex-1 min-w-0">
                                <p class="font-black text-red-900 text-sm truncate">{err.nombre || err.fila}</p>
                                <div class="flex gap-3 mt-0.5 text-[11px] text-red-600 font-medium">
                                  <Show when={err.ci}>
                                    <span>CI: {err.ci}</span>
                                  </Show>
                                  <Show when={err.fpv}>
                                    <span>FPV: {err.fpv}</span>
                                  </Show>
                                </div>
                              </div>
                              <span class="text-[10px] font-black px-2 py-1 bg-red-100 text-red-700 rounded-lg whitespace-nowrap flex-shrink-0">
                                Error
                              </span>
                            </div>
                            <p class="text-xs text-red-700 mt-2 font-medium leading-relaxed">
                              {err.error}
                            </p>
                          </div>
                        )}
                      </For>
                    </div>
                  </div>
                </Show>

                <Show when={r().failed === 0}>
                  <div class="bg-emerald-50 rounded-2xl p-4 border border-emerald-100 text-center">
                    <p class="text-emerald-800 font-black">🎉 Todos los registros importados correctamente</p>
                  </div>
                </Show>
              </div>
            )}
          </Show>

        </div>

        {/* ── FOOTER ────────────────────────────────────────────────────── */}
        <div class="px-8 py-5 border-t border-gray-100 bg-gray-50 flex justify-between items-center gap-3">

          {/* Botón izquierdo */}
          <div>
            <Show when={step() === "preview"}>
              <button
                onClick={handleReset}
                class="text-sm font-black text-gray-500 hover:text-gray-700 transition-colors"
              >
                ← Cambiar archivo
              </button>
            </Show>
            <Show when={step() === "result"}>
              <button
                onClick={handleReset}
                class="text-sm font-black text-blue-600 hover:text-blue-800 transition-colors"
              >
                ↩ Importar otro archivo
              </button>
            </Show>
          </div>

          {/* Botones derecha */}
          <div class="flex gap-3">
            <button
              onClick={props.onClose}
              class="px-5 py-2.5 rounded-xl border-2 border-gray-200 font-black text-gray-600 hover:bg-gray-100 transition-all text-sm"
            >
              {step() === "result" ? "Cerrar" : "Cancelar"}
            </button>

            <Show when={step() === "preview"}>
              <button
                onClick={handleUpload}
                class="px-8 py-2.5 rounded-xl bg-blue-800 text-white font-black hover:bg-blue-900 active:scale-95 transition-all text-sm shadow-lg"
              >
                📥 Importar {preview()?.rowCount} registros
              </button>
            </Show>
          </div>
        </div>

      </div>
    </div>
  );
}