// web/src/components/manuales/ManualPdf.tsx
//
// Vista previa + descarga de un manual PDF servido por la API (endpoint
// autenticado GET /api/v1/manuales/:file, incrustado en el binario Go). El
// archivo se baja como blob con el token JWT en la cabecera Authorization
// (apiDownloadBlob) y se previsualiza en un <iframe> desde un objectURL — el
// token jamás viaja en la URL ni en la barra del navegador. Solo-cliente: usa
// DOM (URL.createObjectURL) y sessionStorage; en SSR no monta nada.
import { createEffect, createSignal, onCleanup, Show } from "solid-js";
import { isServer } from "solid-js/web";
import { apiDownloadBlob } from "~/lib/api";
import { Icon } from "~/components/admin/ui/icons";

export type ManualFile = "manual-admin.pdf" | "manual-psiuser.pdf";

interface Props {
  file: ManualFile;
  /** Alto del iframe de previsualización. */
  height?: string;
}

export default function ManualPdf(props: Props) {
  const [objectUrl, setObjectUrl] = createSignal("");
  const [loading, setLoading] = createSignal(true);
  const [error, setError] = createSignal("");

  // URL actual del blob, libre de reactividad: se revoca al cambiar de archivo
  // o al desmontar el componente (evita fugas de memoria).
  let currentUrl = "";
  let disposed = false;
  onCleanup(() => {
    disposed = true;
    if (currentUrl) URL.revokeObjectURL(currentUrl);
  });

  createEffect(() => {
    if (isServer || disposed) return;
    const file = props.file;

    setLoading(true);
    setError("");
    if (currentUrl) {
      URL.revokeObjectURL(currentUrl);
      currentUrl = "";
    }
    setObjectUrl("");

    apiDownloadBlob(`/manuales/${file}`)
      .then((blob) => {
        if (disposed) return;
        currentUrl = URL.createObjectURL(blob);
        setObjectUrl(currentUrl);
      })
      .catch((e: any) => {
        if (disposed) return;
        setError(
          e?.status === 401
            ? "Tu sesión ya no es válida. Vuelve a iniciar sesión para ver el manual."
            : "No se pudo cargar el manual. Intenta de nuevo en unos segundos."
        );
      })
      .finally(() => {
        if (!disposed) setLoading(false);
      });
  });

  return (
    <Show when={!isServer}>
      <div>
        <Show when={loading()}>
          <div class="flex items-center justify-center h-40 rounded-lg border border-slate-200 bg-slate-50 text-sm text-colpsi-muted">
            Cargando manual…
          </div>
        </Show>

        <Show when={error()}>
          <div class="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700">
            {error()}
          </div>
        </Show>

        <Show when={objectUrl() && !error()}>
          <iframe
            src={objectUrl()}
            title="Manual en PDF"
            class="w-full rounded-lg border border-colpsi-border bg-white"
            style={{ height: props.height || "32rem" }}
          />
          <div class="mt-3">
            <a
              href={objectUrl()}
              download={props.file}
              rel="noopener"
              class="inline-flex items-center gap-2 rounded-lg bg-colpsi-blue px-4 py-2.5 text-sm font-bold text-white transition-colors hover:bg-colpsi-blue/90 active:scale-[0.98]"
            >
              <Icon name="download" class="w-4 h-4" />
              Descargar PDF
            </a>
          </div>
        </Show>
      </div>
    </Show>
  );
}