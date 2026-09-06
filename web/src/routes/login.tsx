// web/src/routes/login.tsx

import { createSignal, Show } from "solid-js";
import { getRequestEvent } from "solid-js/web";
import { useNavigate, A, useAction, action } from "@solidjs/router";
import { useAuth } from "~/lib/auth";
import { apiPost, apiGet, ApiError } from "~/lib/api";
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

export default function LoginPage() {
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
      // 1. Login en Go (Obtenemos el Token)
      const authResponse = await apiPost<{ token: string }>("/psi/login", {
        identifier: identifier(),
        password: password(),
      });

      // 2. Recuperar el perfil completo
      // FIX SENIOR: Como la cookie aún no existe, inyectamos el token recién recibido 
      // directamente en las cabeceras de esta petición específica.
      const userProfile = await apiGet<any>("/psi/me", {
        headers: {
          "Authorization": `Bearer ${authResponse.token}`
        }
      });

      // 3. Inicializar estado global 
      // (Aquí es donde la función useAuth guarda las cookies permanentemente)
      login(authResponse.token, {
        id: userProfile.id,
        username: userProfile.username,
        email: userProfile.email,
        role: "psi",
        firstName: userProfile.first_name,
        lastName: userProfile.last_name,
      });

      // Persistir el JWT en la cookie HttpOnly para que SSR/Server Actions autentiquen contra Go
      await syncJwt(authResponse.token);

      // 4. Redirigir al Dashboard personal
      navigate("/psi", { replace: true });
      
    } catch (err) {
      if (err instanceof ApiError) {
        // 429 = rate-limit: avisar el bloqueo real con el tiempo de espera
        setError(err.status === 429 ? (err.data?.message || err.message) : getUserFacingError(err));
      } else {
        setError("Ocurrió un error inesperado al intentar conectar.");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <main class="min-h-[calc(100-64px)] flex items-center justify-center bg-colpsi-bg px-4 py-12">
      <div class="w-full max-w-md">
        {/* CARD CONTAINER */}
        <div class="bg-white rounded-3xl shadow-xl shadow-blue-900/5 overflow-hidden border border-colpsi-border">
          
          {/* TOP DECORATION (Azul Institucional) */}
          <div class="bg-colpsi-blue p-8 text-center">
            <div class="inline-flex items-center justify-center w-16 h-16 bg-white rounded-2xl mb-4 shadow-lg overflow-hidden">
              <img
                src="/emblema.png"
                alt="Emblema del Colegio de Psicólogos del Estado Carabobo"
                class="w-full h-full object-cover"
              />
            </div>
            <h1 class="text-white text-xl font-bold tracking-tight">Portal de Agremiados</h1>
            <p class="text-blue-200 text-sm mt-1">Colegio de Psicólogos de Carabobo</p>
          </div>

          {/* FORM SECTION */}
          <div class="p-8">
            <form class="space-y-5" onSubmit={handleSubmit}>
              
              <Show when={error()}>
                <div class="bg-red-50 border-l-4 border-colpsi-red p-4 animate-shake">
                  <p class="text-xs text-colpsi-red font-bold uppercase tracking-wide">Error de acceso</p>
                  <p class="text-sm text-red-700">{error()}</p>
                </div>
              </Show>

              <div class="space-y-4">
                <div>
                  <label class="block text-xs font-bold text-gray-400 uppercase mb-1 ml-1">Identificación</label>
                  <input
                    type="text"
                    required
                    placeholder="Usuario o correo electrónico"
                    class="w-full bg-colpsi-surface border-2 border-transparent focus:border-colpsi-yellow focus:bg-white rounded-xl px-4 py-3.5 outline-none transition-all text-gray-700"
                    onInput={(e) => setIdentifier(e.currentTarget.value)}
                  />
                </div>

                <div>
                  <label class="block text-xs font-bold text-gray-400 uppercase mb-1 ml-1">Contraseña</label>
                  {/* <input
                    type="password"
                    required
                    placeholder="••••••••"
                    class="w-full bg-colpsi-surface border-2 border-transparent focus:border-colpsi-yellow focus:bg-white rounded-xl px-4 py-3.5 outline-none transition-all text-gray-700"
                    onInput={(e) => setPassword(e.currentTarget.value)}
                  /> */}
                  <PasswordInputComponent 
                      required 
                      value={password()} 
                      onInput={(e) => setPassword(e.currentTarget.value)}
                      class="w-full bg-colpsi-surface border-2 border-transparent focus:border-colpsi-yellow focus:bg-white rounded-xl px-4 py-3.5 outline-none transition-all text-gray-700"
                      placeholder="••••••••" 
                  />
                </div>
              </div>

              <div class="pt-2">
                <button
                  type="submit"
                  disabled={loading()}
                  class="w-full bg-colpsi-yellow hover:bg-colpsi-yellow-dark text-colpsi-blue font-extrabold py-4 rounded-xl shadow-md shadow-yellow-500/20 active:scale-[0.98] transition-all disabled:opacity-50 disabled:active:scale-100 flex items-center justify-center gap-2"
                >
                  <Show when={loading()} fallback="INGRESAR">
                    <svg class="animate-spin h-5 w-5 text-colpsi-blue" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    VERIFICANDO...
                  </Show>
                </button>
              </div>

              <div class="text-center pt-4">
                <A href="/forgot-password" class="text-sm text-gray-400 hover:text-colpsi-blue transition-colors">
                  ¿Olvidaste tu contraseña?
                </A>
              </div>
            </form>
          </div>
        </div>

        {/* FOOTER LINKS */}
        <div class="mt-8 text-center text-gray-400 text-sm">
          <p>¿No estás registrado? <A href="/directorio" class="text-colpsi-blue font-bold">Consulta el directorio</A></p>
        </div>
      </div>
    </main>
  );
}