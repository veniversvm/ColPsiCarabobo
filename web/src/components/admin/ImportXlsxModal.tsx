// web/src/components/admin/ImportXlsxModal.tsx
import { createSignal, Show, For } from "solid-js";
import { apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";

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
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-blue-900/40 backdrop-blur-md"
      onClick={(e) => { if (e.target === e.currentTarget) props.onClose(); }}
    >
      <div class="bg-white rounded-[2.5rem] shadow-premium w-full max-w-2xl border border-colpsi-border overflow-hidden animate-in zoom-in-95 duration-200">

        {/* ── HEADER ────────────────────────────────────────────────────── */}
        <div class="flex items-center justify-between px-10 py-8 border-b border-gray-50">
          <div>
            <h2 class="text-2xl font-black text-blue-900 uppercase tracking-tight">
              Carga Masiva Excel
            </h2>
            <p class="text-gray-400 text-sm mt-1 font-medium">
              {step() === "select" && "Sube el archivo de agremiados 2026"}
              {step() === "confirm" && "Confirma los datos del archivo"}
              {step() === "uploading" && "Procesando registros en el servidor..."}
              {step() === "result" && "Resultados de la importación"}
            </p>
          </div>
          <button
            onClick={props.onClose}
            class="w-10 h-10 rounded-2xl bg-colpsi-surface hover:bg-red-50 hover:text-red-500 text-gray-400 font-bold flex items-center justify-center transition-all text-xl shadow-sm"
          >
            ✕
          </button>
        </div>

        {/* ── BODY ──────────────────────────────────────────────────────── */}
        <div class="px-10 py-8 max-h-[60vh] overflow-y-auto">

          <Show when={error()}>
            <div class="mb-6 p-5 rounded-2xl bg-red-50 text-red-800 font-bold text-sm border-2 border-red-100 animate-in slide-in-from-top-2 duration-300">
              ⚠️ {error()}
            </div>
          </Show>

          {/* ── STEP: SELECT ──────────────────────────────────────────── */}
          <Show when={step() === "select"}>
            <label class="flex flex-col items-center justify-center w-full h-64 border-3 border-dashed border-gray-200 rounded-3xl bg-colpsi-surface/50 hover:bg-blue-50 hover:border-blue-300 transition-all cursor-pointer group">
              <div class="flex flex-col items-center gap-4 text-gray-400 group-hover:text-blue-600 transition-colors">
                <div class="w-20 h-20 bg-white rounded-3xl shadow-sm flex items-center justify-center text-4xl group-hover:scale-110 transition-transform">
                  📊
                </div>
                <div class="text-center">
                  <span class="block font-black text-lg text-gray-600 group-hover:text-blue-700">Arrastra o selecciona el archivo</span>
                  <span class="text-sm font-medium tracking-wide">Formato soportado: .xlsx / .xls</span>
                </div>
              </div>
              <input type="file" accept=".xlsx,.xls" class="hidden" onChange={handleFileChange} />
            </label>

            <div class="mt-8 grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="p-5 bg-emerald-50 rounded-2xl border border-emerald-100">
                 <p class="text-emerald-800 text-[10px] font-black uppercase tracking-widest mb-2">Requisito de Formato</p>
                 <p class="text-emerald-700 text-xs leading-relaxed font-medium">
                   El sistema espera que los datos comiencen en la <span class="font-black">fila 3</span>. La fila 2 se asume como encabezado.
                 </p>
              </div>
              <div class="p-5 bg-blue-50 rounded-2xl border border-blue-100">
                 <p class="text-blue-800 text-[10px] font-black uppercase tracking-widest mb-2">Campos Clave</p>
                 <p class="text-blue-700 text-xs leading-relaxed font-medium">
                   Asegúrate de incluir FPV, Cédula, Email y las <span class="font-black">Áreas de Desempeño</span> correctamente.
                 </p>
              </div>
            </div>
          </Show>

          {/* ── STEP: CONFIRM ─────────────────────────────────────────── */}
          <Show when={step() === "confirm" && file()}>
            <div class="space-y-6">
              <div class="flex items-center gap-6 p-6 bg-colpsi-surface rounded-3xl border border-colpsi-border shadow-inner">
                <div class="text-5xl">📄</div>
                <div class="flex-1 min-w-0">
                   <p class="text-blue-900 font-black text-lg truncate">{file()?.name}</p>
                   <p class="text-gray-400 text-sm font-bold uppercase tracking-widest">
                     Tamaño: {((file()?.size ?? 0) / 1024).toFixed(1)} KB
                   </p>
                </div>
              </div>

              <div class="bg-amber-50 rounded-3xl p-6 border-2 border-amber-100 flex items-start gap-4">
                <span class="text-2xl">⚡</span>
                <p class="text-amber-900 text-sm leading-relaxed font-medium">
                  Al procesar este archivo, el sistema validará cada fila individualmente. Los psicólogos nuevos recibirán automáticamente sus credenciales temporales vía email.
                </p>
              </div>
            </div>
          </Show>

          {/* ── STEP: UPLOADING ───────────────────────────────────────── */}
          <Show when={step() === "uploading"}>
            <div class="flex flex-col items-center justify-center py-16 gap-8">
              <div class="relative">
                <div class="w-24 h-24 border-4 border-blue-50 border-t-blue-600 rounded-full animate-spin" />
                <div class="absolute inset-0 flex items-center justify-center text-2xl animate-pulse">⚙️</div>
              </div>
              <div class="text-center">
                <p class="font-black text-blue-900 text-xl uppercase tracking-tighter">Sincronizando Base de Datos</p>
                <p class="text-gray-400 text-sm mt-2 font-medium">Leyendo celdas y validando credenciales gremiales...</p>
              </div>
            </div>
          </Show>

          {/* ── STEP: RESULT ──────────────────────────────────────────── */}
          <Show when={step() === "result" && result()}>
            {(res) => (
              <div class="space-y-6">
                <div class="grid grid-cols-2 gap-4">
                  <div class="bg-emerald-50 rounded-3xl p-6 text-center border-2 border-emerald-100 shadow-sm">
                    <p class="text-5xl font-black text-emerald-700">{res().imported}</p>
                    <p class="text-[10px] font-black text-emerald-600 uppercase tracking-[0.2em] mt-2">Éxito</p>
                  </div>
                  <div class={`rounded-3xl p-6 text-center border-2 shadow-sm ${res().failed > 0 ? "bg-red-50 border-red-100" : "bg-colpsi-surface border-colpsi-border"}`}>
                    <p class={`text-5xl font-black ${res().failed > 0 ? "text-red-700" : "text-gray-400"}`}>{res().failed}</p>
                    <p class={`text-[10px] font-black uppercase tracking-[0.2em] mt-2 ${res().failed > 0 ? "text-red-600" : "text-gray-400"}`}>
                      Errores
                    </p>
                  </div>
                </div>

                <Show when={res().errors && res().errors.length > 0}>
                  <div class="space-y-3">
                    <p class="text-[10px] font-black text-gray-400 uppercase tracking-[0.2em] ml-2">Detalle de incidencias</p>
                    <div class="space-y-2 max-h-60 overflow-y-auto pr-2 custom-scrollbar">
                      <For each={res().errors}>
                        {(err) => (
                          <div class="bg-white rounded-2xl p-4 border border-red-100 shadow-sm">
                            <div class="flex justify-between items-start mb-1">
                              <p class="font-black text-gray-800 text-sm">{err.nombre || `Fila ${err.fila}`}</p>
                              <span class="text-[9px] font-black px-2 py-0.5 bg-red-100 text-red-700 rounded-md uppercase">Fallo</span>
                            </div>
                            <p class="text-xs text-red-600 leading-relaxed font-medium">{err.error}</p>
                            <div class="flex gap-4 mt-2 text-[9px] font-bold text-gray-400 uppercase tracking-widest">
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
                  <div class="bg-emerald-600 rounded-3xl p-6 text-center shadow-lg shadow-emerald-200">
                    <p class="text-white font-black uppercase tracking-widest text-sm">🎉 Importación completada con éxito</p>
                  </div>
                </Show>
              </div>
            )}
          </Show>

        </div>

        {/* ── FOOTER ────────────────────────────────────────────────────── */}
        <div class="px-10 py-8 border-t border-gray-50 bg-colpsi-surface/50 flex justify-between items-center">
          
          <div>
            <Show when={step() === "confirm" || step() === "result"}>
              <button
                onClick={handleReset}
                class="text-xs font-black text-blue-600 hover:text-blue-800 uppercase tracking-widest transition-colors"
              >
                {step() === "result" ? "↩ Cargar otro" : "← Cambiar archivo"}
              </button>
            </Show>
          </div>

          <div class="flex gap-4">
            <button
              onClick={props.onClose}
              class="px-8 py-3.5 rounded-2xl border-2 border-gray-200 font-black text-gray-500 hover:bg-gray-100 transition-all text-xs uppercase tracking-widest"
            >
              {step() === "result" ? "Finalizar" : "Cancelar"}
            </button>

            <Show when={step() === "confirm"}>
              <button
                onClick={handleUpload}
                class="px-10 py-3.5 rounded-2xl bg-blue-900 text-white font-black hover:bg-blue-800 active:scale-95 transition-all text-xs uppercase tracking-widest shadow-xl shadow-blue-900/20"
              >
                🚀 Iniciar Importación
              </button>
            </Show>
          </div>
        </div>

      </div>
    </div>
  );
}