import { createSignal, Show } from "solid-js";
import { getRequestEvent } from "solid-js/web";
import { useNavigate, useAction, action } from "@solidjs/router";
import { useAuth } from "~/lib/auth";
import { apiPost, ApiError } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { PasswordInputComponent } from "~/components/ui/PasswordInput";

const syncJwtCookie = action(async (token: string) => {
  "use server";
  if (!token) return { error: "Token requerido." };
  const event = getRequestEvent();
  const secure = import.meta.env.PROD === true;
  event?.response?.headers?.set(
    "Set-Cookie",
    `jwt=${token}; HttpOnly; Path=/; Max-Age=86400; SameSite=Strict${secure ? "; Secure" : ""}`,
  );
  return { ok: true };
});

export default function AdminLoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const syncJwt = useAction(syncJwtCookie);

  const [identifier, setIdentifier] = createSignal("");
  const [password, setPassword] = createSignal("");
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      // 1. Petición al endpoint (Solo devuelve message y token)
      const response = await apiPost<{ message: string, token: string }>("/auth/login", {
        identifier: identifier(),
        password: password(),
      });

      // 2. Extraer el Payload del JWT (Nivel Senior)
      // El JWT tiene formato: header.payload.signature
      const payloadBase64 = response.token.split('.')[1];
      const decodedJson = atob(payloadBase64); // Decodifica Base64
      const payload = JSON.parse(decodedJson); // Convierte a objeto JS

      // 3. Guardar en estado global usando el ID real del token 
      // y el identificador que el usuario tipeó en la pantalla.
      login(response.token, {
        id: payload.user_id,    // El UUID real que inyectamos en Go
        username: identifier(), // Lo que escribió en el formulario
        email: identifier(),
        role: "admin",          // El router leerá esto y te dejará pasar
        firstName: "Administrador",
      });

      // Persistir el JWT en la cookie HttpOnly para que SSR/Server Actions autentiquen contra Go
      await syncJwt(response.token);

      // 4. Redirigir al Dashboard de Admin
      navigate("/admin", { replace: true });
      
    } catch (err: any) {
      if (err instanceof ApiError) {
        // 429 = rate-limit: avisar el bloqueo real con el tiempo de espera
        setError(err.status === 429 ? (err.data?.message || err.message) : getUserFacingError(err));
      } else {
        console.error("Error interno:", err); // <-- Útil para ver errores de JS en la consola
        setError("Error interno al procesar la sesión.");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <main class="min-h-screen flex items-center justify-center bg-[#1e293b] px-4 font-sans relative overflow-hidden">
      {/* Fondo técnico/administrativo */}
      <div class="absolute inset-0 z-0 opacity-10">
        <div class="absolute inset-0 bg-[linear-gradient(to_right,#80808012_1px,transparent_1px),linear-gradient(to_bottom,#80808012_1px,transparent_1px)] bg-[size:24px_24px]"></div>
      </div>

      <div class="max-w-md w-full relative z-10">
        <div class="bg-white rounded-3xl shadow-2xl overflow-hidden border border-gray-200">
          
          {/* TOP DECORATION (Rojo Seguridad / Staff) */}
          <div class="bg-gradient-to-b from-gray-900 to-gray-800 p-8 text-center border-b-4 border-colpsi-red">
            <div class="inline-flex items-center justify-center w-16 h-16 bg-white/10 rounded-2xl mb-4 backdrop-blur-md border border-white/20">
              <span class="text-colpsi-red text-3xl font-black">⚙️</span>
            </div>
            <h1 class="text-white text-xl font-bold tracking-widest uppercase">
              Acceso Restringido
            </h1>
            <p class="text-gray-400 text-xs mt-2 uppercase tracking-widest">
              Staff de Administración
            </p>
          </div>

          {/* FORM SECTION */}
          <div class="p-8">
            <form class="space-y-6" onSubmit={handleSubmit}>
              
              <Show when={error()}>
                <div class="bg-red-50 border-l-4 border-red-500 p-4 animate-in shake duration-300">
                  <p class="text-xs text-red-800 font-bold uppercase tracking-wide">Fallo de Seguridad</p>
                  <p class="text-sm text-red-600 mt-1">{error()}</p>
                </div>
              </Show>

              <div class="space-y-5">
                <div class="space-y-2">
                  <label class="block text-xs font-bold text-gray-500 uppercase ml-1">Usuario Administrativo</label>
                  <input
                    type="text"
                    required
                    placeholder="Username o Correo institucional"
                    class="w-full bg-colpsi-surface border-2 border-colpsi-border focus:border-colpsi-red focus:bg-white rounded-xl px-5 py-4 outline-none transition-all text-gray-800 font-medium"
                    onInput={(e) => setIdentifier(e.currentTarget.value)}
                  />
                </div>

                <div class="space-y-2">
                  <label class="block text-xs font-bold text-gray-500 uppercase ml-1">Clave de Acceso</label>
                  <PasswordInputComponent
                    required
                    placeholder="••••••••"
                    class="bg-colpsi-surface border-2 border-colpsi-border focus:border-colpsi-red focus:bg-white rounded-xl px-5 py-4 outline-none transition-all text-gray-800"
                    onInput={(e) => setPassword(e.currentTarget.value)}
                  />
                </div>
              </div>

              <div class="pt-4">
                <button
                  type="submit"
                  disabled={loading()}
                  class="w-full bg-gray-900 hover:bg-black text-white font-black uppercase tracking-widest py-4 rounded-xl shadow-xl hover:shadow-2xl active:scale-[0.98] transition-all disabled:opacity-50 flex items-center justify-center gap-3"
                >
                  <Show when={loading()} fallback={<><span>🔓</span> AUTENTICAR</>}>
                    <svg class="animate-spin h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    VERIFICANDO...
                  </Show>
                </button>
              </div>
            </form>
          </div>
        </div>

        {/* FOOTER WARNING */}
        <div class="mt-8 text-center text-gray-500 text-xs font-mono uppercase tracking-widest bg-gray-900/50 p-4 rounded-xl backdrop-blur-sm border border-gray-700/50">
          <p>⚠️ Intento de acceso registrado</p>
          <p class="mt-1 opacity-50">Toda actividad está sujeta a auditoría legal</p>
        </div>
      </div>
    </main>
  );
}