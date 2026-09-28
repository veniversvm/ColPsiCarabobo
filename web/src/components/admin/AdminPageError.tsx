// web/src/components/admin/AdminPageError.tsx
//
// Fallback inline del ErrorBoundary del layout admin (admin.tsx). Reemplaza al
// OfflineAlert a pantalla completa: mantiene el menú lateral y el topbar visibles
// para que un error puntual (p. ej. un 500 transitorio de la API) NO saque al
// admin del panel. Si el error es de red (503 OFFLINE_SERVICE) se muestra copy de
// conexión; en otro caso se expone el mensaje real del ApiError.
import { useNavigate } from "@solidjs/router";
import { Icon } from "~/components/admin/ui/icons";

export default function AdminPageError(props: { error: any; reset: () => void }) {
  const navigate = useNavigate();
  const status = props.error?.status ?? 500;
  const rawMessage = String(props.error?.message ?? "");
  const isOffline =
    status === 503 ||
    props.error?.data === "OFFLINE_SERVICE" ||
    /offline|fetch|network/i.test(rawMessage);

  return (
    <div class="bg-white border border-colpsi-border rounded-lg p-10 text-center">
      <span class="inline-flex h-12 w-12 items-center justify-center rounded-md bg-colpsi-bg text-colpsi-blue mb-4">
        <Icon name="alertTriangle" class="w-6 h-6" />
      </span>
      <h2 class="text-lg font-semibold text-colpsi-text mb-1">
        {isOffline ? "Conexión en pausa" : "No se pudo cargar esta sección"}
      </h2>
      <p class="text-sm text-colpsi-muted max-w-lg mx-auto mb-6">
        {isOffline
          ? "Estamos teniendo dificultades para conectar con el servidor del Colegio. Verifique su internet o intente de nuevo."
          : rawMessage || "Ocurrió un error inesperado. Intente de nuevo."}
      </p>
      <div class="flex flex-wrap items-center justify-center gap-2">
        <button
          onClick={props.reset}
          class="inline-flex items-center gap-2 h-9 px-4 rounded-md bg-colpsi-blue text-white text-sm font-semibold hover:bg-colpsi-blue-light transition-all"
        >
          <Icon name="refresh" class="w-4 h-4" /> Reintentar
        </button>
        <button
          onClick={() => navigate("/admin")}
          class="inline-flex items-center gap-2 h-9 px-4 rounded-md border border-colpsi-border bg-white text-sm font-medium text-colpsi-text hover:bg-colpsi-bg transition-all"
        >
          Volver al panel
        </button>
      </div>
      <p class="text-[9px] text-gray-300 font-mono tracking-widest uppercase mt-6">ID Error: {status}</p>
    </div>
  );
}