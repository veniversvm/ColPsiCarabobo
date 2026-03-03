// web/src/routes/login.tsx

import { createSignal, Show } from "solid-js";
import { useNavigate, A } from "@solidjs/router";
import { useAuth } from "~/lib/auth";
import { apiPost, apiGet, ApiError } from "~/lib/api";
import { PasswordInputComponent } from "~/components/ui/PasswordInput";

export default function LoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();

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

      // 4. Redirigir al Dashboard personal
      navigate("/psi", { replace: true });
      
    } catch (err) {
      if (err instanceof ApiError) {
        // Mostrar el mensaje real que viene de Go (ej. "credenciales inválidas")
        setError(err.message);
      } else {
        setError("Ocurrió un error inesperado al intentar conectar.");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <main class="min-h-[calc(100-64px)] flex items-center justify-center bg-[#f8fafc] px-4 py-12">
      <div class="w-full max-w-md">
        {/* CARD CONTAINER */}
        <div class="bg-white rounded-3xl shadow-xl shadow-blue-900/5 overflow-hidden border border-gray-100">
          
          {/* TOP DECORATION (Azul Institucional) */}
          <div class="bg-[#1e3a8a] p-8 text-center">
            <div class="inline-flex items-center justify-center w-16 h-16 bg-white rounded-2xl mb-4 shadow-lg">
              <span class="text-[#1e3a8a] text-4xl font-bold">Ψ</span>
            </div>
            <h1 class="text-white text-xl font-bold tracking-tight">Portal de Agremiados</h1>
            <p class="text-blue-200 text-sm mt-1">Colegio de Psicólogos de Carabobo</p>
          </div>

          {/* FORM SECTION */}
          <div class="p-8">
            <form class="space-y-5" onSubmit={handleSubmit}>
              
              <Show when={error()}>
                <div class="bg-red-50 border-l-4 border-[#991b1b] p-4 animate-shake">
                  <p class="text-xs text-[#991b1b] font-bold uppercase tracking-wide">Error de acceso</p>
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
                    class="w-full bg-gray-50 border-2 border-transparent focus:border-[#facc15] focus:bg-white rounded-xl px-4 py-3.5 outline-none transition-all text-gray-700"
                    onInput={(e) => setIdentifier(e.currentTarget.value)}
                  />
                </div>

                <div>
                  <label class="block text-xs font-bold text-gray-400 uppercase mb-1 ml-1">Contraseña</label>
                  {/* <input
                    type="password"
                    required
                    placeholder="••••••••"
                    class="w-full bg-gray-50 border-2 border-transparent focus:border-[#facc15] focus:bg-white rounded-xl px-4 py-3.5 outline-none transition-all text-gray-700"
                    onInput={(e) => setPassword(e.currentTarget.value)}
                  /> */}
                  <PasswordInputComponent 
                      required 
                      value={password()} 
                      onInput={(e) => setPassword(e.currentTarget.value)}
                      class="w-full bg-gray-50 border-2 border-transparent focus:border-[#facc15] focus:bg-white rounded-xl px-4 py-3.5 outline-none transition-all text-gray-700"
                      placeholder="••••••••" 
                  />
                </div>
              </div>

              <div class="pt-2">
                <button
                  type="submit"
                  disabled={loading()}
                  class="w-full bg-[#facc15] hover:bg-[#eab308] text-[#1e3a8a] font-extrabold py-4 rounded-xl shadow-md shadow-yellow-500/20 active:scale-[0.98] transition-all disabled:opacity-50 disabled:active:scale-100 flex items-center justify-center gap-2"
                >
                  <Show when={loading()} fallback="INGRESAR">
                    <svg class="animate-spin h-5 w-5 text-[#1e3a8a]" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    VERIFICANDO...
                  </Show>
                </button>
              </div>

              <div class="text-center pt-4">
                <A href="/forgot-password" class="text-sm text-gray-400 hover:text-[#1e3a8a] transition-colors">
                  ¿Olvidaste tu contraseña?
                </A>
              </div>
            </form>
          </div>
        </div>

        {/* FOOTER LINKS */}
        <div class="mt-8 text-center text-gray-400 text-sm">
          <p>¿No estás registrado? <A href="/directorio" class="text-[#1e3a8a] font-bold">Consulta el directorio</A></p>
        </div>
      </div>
    </main>
  );
}