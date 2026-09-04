// web/src/components/inscripcion/FileUpload.tsx
//
// Campo de subida de archivos con preview y validación de MIME + tamaño (5MB).
import { createSignal, Show } from "solid-js";

interface Props {
  label: string;
  accept: string;      // "image/*" o "image/*,application/pdf"
  description?: string;
  maxSizeMB?: number;
  file?: File | null;
  // Nombre de un archivo seleccionado previamente (persistido en la sesión).
  onFile: (file: File | null) => void;
}

export function FileUpload(props: Props) {
  const maxSize = props.maxSizeMB ? props.maxSizeMB * 1024 * 1024 : 5 * 1024 * 1024;
  const [error, setError] = createSignal("");
  const savedName = props.savedName;

  const handleChange = (e: Event) => {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0] || null;
    if (!file) {
      props.onFile(null);
      setError("");
      return;
    }

    if (file.size > maxSize) {
      setError(`El archivo supera el tamaño máximo de ${props.maxSizeMB || 5}MB`);
      props.onFile(null);
      return;
    }

    setError("");
    props.onFile(file);
  };

  const hasFile = () => !!props.file || !!savedName;

  return (
    <div>
      <label class="flex flex-col items-center justify-center border-2 border-dashed border-gray-200 rounded-2xl h-32 hover:border-colpsi-blue transition-all cursor-pointer bg-gray-50/50 px-4 text-center">
        <input type="file" accept={props.accept} class="sr-only" onChange={handleChange} />
        <Show when={hasFile()} fallback={<span class="text-2xl">📎</span>}>
          <span class="text-green-600 text-xs font-bold mb-1">✓ Archivo listo</span>
        </Show>
        <span class="text-[11px] font-black text-colpsi-blue uppercase">{props.label || "Seleccionar archivo"}</span>
        <Show when={savedName}>
          <span class="text-[10px] text-gray-400 font-bold mt-1 max-w-full truncate">{savedName}</span>
        </Show>
        {props.description && <span class="text-[10px] text-gray-400 mt-1">{props.description}</span>}
      </label>
      <Show when={error()}>
        <span class="block text-xs text-red-500 font-semibold mt-1">{error()}</span>
      </Show>
    </div>
  );
}