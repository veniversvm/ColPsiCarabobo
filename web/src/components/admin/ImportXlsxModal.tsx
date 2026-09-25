// web/src/components/admin/ImportXlsxModal.tsx
import { createSignal, Show, For } from "solid-js";
import { apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { Icon } from "~/components/admin/ui/icons";

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

interface FileMetadata {
  name: string;
  size: number;
  type: string;
}

interface ImportXlsxModalProps {
  onClose: () => void;
  onSuccess: () => void;
}

export function ImportXlsxModal(props: ImportXlsxModalProps) {
  // ── Estado ──────────────────────────────────────────────────────────────
  type Step = "select" | "confirm" | "uploading" | "result";
  const [step, setStep] = createSignal<Step>("select");
  const [file, setFile] = createSignal<File | null>(null);
  const [result, setResult] = createSignal<ImportResult | null>(null);
  const [error, setError] = createSignal<string | null>(null);

  // ── Handlers ─────────────────────────────────────────────────────────────
  const handleFileChange = (e: Event) => {
    const f = (e.currentTarget as HTMLInputElement).files?.[0] ?? null;
    if (!f) return;

    // Validación de extensión para Excel
    const validExtensions = [".xlsx", ".xls"];
    const fileName = f.name.toLowerCase();
    const isValid = validExtensions.some(ext => fileName.endsWith(ext));

    if (!isValid) {
      setError("El archivo debe ser un documento de Excel (.xlsx o .xls).");
      return;
    }

    // Validar tipo MIME
    const allowedMimes = [
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
      "application/vnd.ms-excel",
    ];
    if (!allowedMimes.includes(f.type)) {
      setError("El archivo no parece ser un documento de Excel válido.");
      return;
    }

    // Validar tamaño (5MB max)
    const MAX_SIZE = 5 * 1024 * 1024;
    if (f.size > MAX_SIZE) {
      setError("El archivo excede el límite de 5MB.");
      return;
    }

    setFile(f);
    setError(null);
    setStep("confirm");
  };

  const handleUpload = async () => {
    const f = file();
    if (!f) return;

    setStep("uploading");
    setError(null);

    try {
      const fd = new FormData();
      // El backend de Go ahora procesa este archivo XLSX
      fd.append("xlsx", f);

      const res = await apiPost<ImportResult>("/admin/psi/upload-csv", fd);
      setResult(res);
      setStep("result");

      // Si se importaron registros con éxito, notificar al padre
      if (res.imported > 0) props.onSuccess();
    } catch (err: any) {
      setError(getUserFacingError(err));
      setStep("confirm");
    }
  };

  const handleReset = () => {
    setFile(null);
    setResult(null);
    setError(null);
    setStep("select");
  };

  // ── Render ────────────────────────────────────────────────────────────────
  return (
    <div
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
      onClick={(e) => { if (e.target === e.currentTarget) props.onClose(); }}
    >
      <div class="bg-white rounded-lg shadow-lg w-full max-w-2xl border border-colpsi-border overflow-hidden animate-in zoom-in-95 duration-200">

        {/* ── HEADER ────────────────────────────────────────────────────── */}
        <div class="flex items-center justify-between px-6 py-5 border-b border-colpsi-border">
          <div>
            <h2 class="text-lg font-semibold text-colpsi-text">Carga Masiva Excel</h2>
            <p class="text-colpsi-muted text-sm mt-0.5 font-medium">
              {step() === "select" && "Sube el archivo de agremiados 2026"}
              {step() === "confirm" && "Confirma los datos del archivo"}
              {step() === "uploading" && "Procesando registros en el servidor..."}
              {step() === "result" && "Resultados de la importación"}
            </p>
          </div>
          <button
            onClick={props.onClose}
            class="inline-flex items-center justify-center h-8 w-8 rounded-md bg-colpsi-bg hover:bg-red-50 hover:text-colpsi-red text-colpsi-muted transition-all"
            title="Cerrar"
          >
            <Icon name="x" class="w-4 h-4" />
          </button>
        </div>

        {/* ── BODY ──────────────────────────────────────────────────────── */}
        <div class="px-6 py-5 max-h-[60vh] overflow-y-auto">

          <Show when={error()}>
            <div class="mb-4 p-3 rounded-md bg-red-50 text-colpsi-red font-medium text-sm border border-red-200 animate-in slide-in-from-top-2 duration-300 flex items-start gap-2">
              <Icon name="alertTriangle" class="w-4 h-4 shrink-0 mt-0.5" />
              {error()}
            </div>
          </Show>

          {/* ── STEP: SELECT ──────────────────────────────────────────── */}
          <Show when={step() === "select"}>
            <label class="flex flex-col items-center justify-center w-full h-64 border border-dashed border-slate-300 rounded-lg bg-colpsi-bg/60 hover:border-colpsi-blue hover:bg-blue-50/40 transition-all cursor-pointer group">
              <div class="flex flex-col items-center gap-4 text-colpsi-muted group-hover:text-colpsi-blue transition-colors">
                <div class="w-16 h-16 bg-white rounded-lg border border-colpsi-border flex items-center justify-center group-hover:scale-105 transition-transform text-colpsi-blue">
                  <Icon name="barChart" class="w-8 h-8" />
                </div>
                <div class="text-center">
                  <span class="block font-semibold text-base text-colpsi-text group-hover:text-colpsi-blue">Arrastra o selecciona el archivo</span>
                  <span class="text-sm text-colpsi-muted font-medium tracking-wide">Formato soportado: .xlsx / .xls</span>
                </div>
              </div>
              <input type="file" accept=".xlsx,.xls" class="hidden" onChange={handleFileChange} />
            </label>

            <div class="mt-6 grid grid-cols-1 md:grid-cols-2 gap-3">
              <div class="p-4 bg-emerald-50 rounded-md border border-emerald-200">
                 <p class="text-emerald-800 text-[10px] font-semibold uppercase tracking-wide mb-1.5">Requisito de Formato</p>
                 <p class="text-emerald-700 text-xs leading-relaxed font-medium">
                   El sistema espera que los datos comiencen en la <span class="font-semibold">fila 3</span>. La fila 2 se asume como encabezado.
                 </p>
              </div>
              <div class="p-4 bg-blue-50 rounded-md border border-blue-200">
                 <p class="text-colpsi-blue text-[10px] font-semibold uppercase tracking-wide mb-1.5">Campos Clave</p>
                 <p class="text-colpsi-blue/80 text-xs leading-relaxed font-medium">
                   Asegúrate de incluir FPV, Cédula, Email y las <span class="font-semibold">Áreas de Desempeño</span> correctamente.
                 </p>
              </div>
            </div>
          </Show>

          {/* ── STEP: CONFIRM ─────────────────────────────────────────── */}
          <Show when={step() === "confirm" && file()}>
            <div class="space-y-4">
              <div class="flex items-center gap-4 p-4 bg-colpsi-bg rounded-lg border border-colpsi-border">
                <span class="inline-flex h-11 w-11 items-center justify-center rounded-md bg-white border border-colpsi-border text-colpsi-blue shrink-0">
                  <Icon name="fileText" class="w-5 h-5" />
                </span>
                <div class="flex-1 min-w-0">
                   <p class="text-colpsi-text font-semibold text-sm truncate">{file()?.name}</p>
                   <p class="text-colpsi-muted text-xs font-medium uppercase tracking-wide">
                     Tamaño: {((file()?.size ?? 0) / 1024).toFixed(1)} KB
                   </p>
                </div>
              </div>

              <div class="bg-amber-50 rounded-md p-4 border border-amber-200 flex items-start gap-2.5">
                <Icon name="info" class="w-4 h-4 shrink-0 mt-0.5 text-amber-600" />
                <p class="text-amber-900 text-sm leading-relaxed font-medium">
                  Al procesar este archivo, el sistema validará cada fila individualmente. Los psicólogos nuevos recibirán automáticamente sus credenciales temporales vía email.
                </p>
              </div>
            </div>
          </Show>

          {/* ── STEP: UPLOADING ───────────────────────────────────────── */}
          <Show when={step() === "uploading"}>
            <div class="flex flex-col items-center justify-center py-16 gap-6">
              <div class="relative">
                <div class="w-20 h-20 border-4 border-slate-100 border-t-colpsi-blue rounded-full animate-spin" />
                <div class="absolute inset-0 flex items-center justify-center text-colpsi-blue">
                  <Icon name="sliders" class="w-7 h-7 animate-pulse" />
                </div>
              </div>
              <div class="text-center">
                <p class="font-semibold text-colpsi-text text-lg">Sincronizando Base de Datos</p>
                <p class="text-colpsi-muted text-sm mt-1 font-medium">Leyendo celdas y validando credenciales gremiales...</p>
              </div>
            </div>
          </Show>

          {/* ── STEP: RESULT ──────────────────────────────────────────── */}
          <Show when={step() === "result" && result()}>
            {(res) => (
              <div class="space-y-4">
                <div class="grid grid-cols-2 gap-3">
                  <div class="bg-emerald-50 rounded-lg p-5 text-center border border-emerald-200">
                    <p class="text-3xl font-semibold text-emerald-700">{res().imported}</p>
                    <p class="text-[10px] font-semibold text-emerald-600 uppercase tracking-[0.15em] mt-1.5">Registros importados</p>
                  </div>
                  <div class={`rounded-lg p-5 text-center border ${res().failed > 0 ? "bg-red-50 border-red-200" : "bg-colpsi-bg border-colpsi-border"}`}>
                    <p class={`text-3xl font-semibold ${res().failed > 0 ? "text-colpsi-red" : "text-colpsi-muted"}`}>{res().failed}</p>
                    <p class={`text-[10px] font-semibold uppercase tracking-[0.15em] mt-1.5 ${res().failed > 0 ? "text-colpsi-red" : "text-colpsi-muted"}`}>
                      Incidencias
                    </p>
                  </div>
                </div>

                <Show when={res().errors && res().errors.length > 0}>
                  <div class="space-y-2.5">
                    <p class="text-[10px] font-semibold text-colpsi-muted uppercase tracking-[0.15em] ml-1">Detalle de incidencias</p>
                    <div class="space-y-2 max-h-60 overflow-y-auto pr-1 custom-scrollbar">
                      <For each={res().errors}>
                        {(err) => (
                          <div class="bg-white rounded-md p-3.5 border border-red-200">
                            <div class="flex justify-between items-start mb-1">
                              <p class="font-semibold text-colpsi-text text-sm">{err.nombre || `Fila ${err.fila}`}</p>
                              <span class="text-[9px] font-semibold px-2 py-0.5 bg-red-100 text-colpsi-red rounded-md uppercase">Fallo</span>
                            </div>
                            <p class="text-xs text-colpsi-red leading-relaxed font-medium">{err.error}</p>
                            <div class="flex gap-4 mt-2 text-[9px] font-semibold text-colpsi-muted uppercase tracking-wide">
                               <Show when={err.fpv}><span>FPV: {err.fpv}</span></Show>
                               <Show when={err.ci}><span>CI: {err.ci}</span></Show>
                            </div>
                          </div>
                        )}
                      </For>
                    </div>
                  </div>
                </Show>

                <Show when={res().failed === 0}>
                  <div class="bg-emerald-600 rounded-md p-4 text-center inline-flex items-center justify-center gap-2 w-full">
                    <Icon name="checkCircle" class="w-4 h-4 text-white" />
                    <p class="text-white font-semibold uppercase tracking-wide text-sm">Importación completada con éxito</p>
                  </div>
                </Show>
              </div>
            )}
          </Show>

        </div>

        {/* ── FOOTER ────────────────────────────────────────────────────── */}
        <div class="px-6 py-4 border-t border-colpsi-border bg-colpsi-bg/50 flex justify-between items-center">

          <div>
            <Show when={step() === "confirm" || step() === "result"}>
              <button
                onClick={handleReset}
                class="inline-flex items-center gap-1.5 text-xs font-semibold text-colpsi-blue hover:underline uppercase tracking-wide transition-colors"
              >
                {step() === "result"
                  ? <><Icon name="refresh" class="w-3.5 h-3.5" /> Cargar otro</>
                  : <><Icon name="chevronRight" class="w-3.5 h-3.5 rotate-180" /> Cambiar archivo</>}
              </button>
            </Show>
          </div>

          <div class="flex gap-2">
            <button
              onClick={props.onClose}
              class="h-11 px-6 rounded-md border border-colpsi-border bg-white font-medium text-colpsi-text hover:bg-colpsi-bg transition-all text-sm"
            >
              {step() === "result" ? "Finalizar" : "Cancelar"}
            </button>

            <Show when={step() === "confirm"}>
              <button
                onClick={handleUpload}
                class="inline-flex items-center gap-2 h-11 px-6 rounded-md bg-colpsi-blue hover:bg-colpsi-blue-light text-white font-semibold transition-all text-sm shadow-sm"
              >
                <Icon name="rocket" class="w-4 h-4" />
                Iniciar Importación
              </button>
            </Show>
          </div>
        </div>

      </div>
    </div>
  );
}