// web/src/routes/reset-password.tsx
//
// Recuperación de contraseña — paso 2: el psicólogo llega con el token de la
// URL (?token=...) y fija su nueva contraseña. El token es de un solo uso y
// expira en 1 hora.
import { createSignal, Show } from "solid-js";
import { getRequestEvent } from "solid-js/web";
import { useSearchParams, useNavigate, A, action, useAction } from "@solidjs/router";
import { apiPost, ApiError } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import { useAuth } from "~/lib/auth";
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

export default function ResetPasswordPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const syncJwt = useAction(syncJwtCookie);
  const [searchParams] = useSearchParams();

  const [password, setPassword] = createSignal("");
  const [confirm, setConfirm] = createSignal("");
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);
  const [done, setDone] = createSignal(false);

  // useSearchParams() devuelve [location.query, setter]: query es un memo
  // reactivo (acceso por propiedad), NO una función.
  const token = () => (searchParams.token as string) || "";

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    setError(null);

    if (password() !== confirm()) {
      setError("Las contraseñas no coinciden.");
      return;
    }
    if (password().length < 8) {
      setError("La contraseña debe tener al menos 8 caracteres.");
      return;
    }

    setLoading(true);
    try {
      const res = await apiPost<{
        token?: string;
        id?: string;
        username?: string;
        email?: string;
        first_name?: string;
        last_name?: string;
      }>("/psi/reset-password", {
        token: token(),
        new_password: password(),
        confirm_password: confirm(),
      });

      // La API emite un JWT fresco tras el reset → auto-login sin volver a
      // digitar. Se persiste la cookie HttpOnly y se inicia la sesión local.
      if (res?.token) {
        await syncJwt(res.token);
        login(res.token, {
          id: res.id || "",
          username: res.username || "",
          email: res.email || "",
          role: "psi",
          firstName: res.first_name || "",
          lastName: res.last_name || "",
        });
        navigate("/psi", { replace: true });
        return;
      }

      setDone(true);
      setTimeout(() => navigate("/login", { replace: true }), 2500);
    } catch (err) {
      if (err instanceof ApiError) {
        // 429 = rate-limit; 400 = enlace inválido/expirado/usado
        setError(err.status === 429 ? (err.data?.message || err.message) : getUserFacingError(err));
      } else {
        setError("Ocurrió un error inesperado al intentar actualizar la contraseña.");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <main class="min-h-[calc(100vh-64px)] flex items-center justify-center bg-white px-4 py-12">
      <div class="w-full max-w-md">
        {/* CARD CONTAINER */}
        <div class="bg-white rounded-3xl shadow-xl shadow-blue-900/5 overflow-hidden border border-colpsi-border">

          {/* TOP DECORATION (Azul Institucional) */}
          <div class="bg-colpsi-blue p-8 text-center">
            <div class="inline-flex items-center justify-center bg-white rounded-2xl px-6 py-3 mb-4 shadow-lg">
              <img
                src="/logo-horizontal.png"
                alt="Colegio de Psicólogos del Estado Carabobo"
                class="h-9 w-auto"
              />
            </div>
            <h1 class="text-white text-xl font-bold tracking-tight">Nueva contraseña</h1>
            <p class="text-blue-200 text-sm mt-1">Colegio de Psicólogos de Carabobo</p>
          </div>

          {/* SECTION */}
          <div class="p-8">

            {/* Sin token en la URL — enlace inválido */}
            <Show when={!token() && !done()}>
              <div class="text-center">
                <div class="bg-red-50 border-l-4 border-colpsi-red p-4 mb-6 text-left">
                  <p class="text-xs text-colpsi-red font-bold uppercase tracking-wide">Enlace inválido</p>
                  <p class="text-sm text-red-700">Este enlace no tiene un token válido. Solicite uno nuevo.</p>
                </div>
                <A href="/forgot-password" class="inline-block bg-colpsi-yellow hover:bg-colpsi-yellow-dark text-colpsi-blue font-extrabold py-3 px-6 rounded-xl shadow-md shadow-yellow-500/20 transition-all">
                  Solicitar enlace
                </A>
              </div>
            </Show>

            {/* Formulario con token presente */}
            <Show when={token() && !done()}>
              <form class="space-y-5" onSubmit={handleSubmit}>

                <Show when={error()}>
                  <div class="bg-red-50 border-l-4 border-colpsi-red p-4 animate-shake">
                    <p class="text-xs text-colpsi-red font-bold uppercase tracking-wide">No se pudo restablecer</p>
                    <p class="text-sm text-red-700">{error()}</p>
                  </div>
                  <Show when={error()?.includes("inválido") || error()?.includes("expirado")}>
                    <div class="text-center">
                      <A href="/forgot-password" class="text-sm text-colpsi-blue font-bold hover:underline">
                        Solicitar un nuevo enlace
                      </A>
                    </div>
                  </Show>
                </Show>

                <div>
                  <label class="block text-xs font-bold text-gray-400 uppercase mb-1 ml-1">Nueva contraseña</label>
                  <PasswordInputComponent
                    required
                    value={password()}
                    onInput={(e) => setPassword(e.currentTarget.value)}
                    placeholder="Mínimo 8 caracteres"
                    class="w-full bg-colpsi-surface border-2 border-transparent focus:border-colpsi-yellow focus:bg-white rounded-xl px-4 py-3.5 outline-none transition-all text-gray-700"
                  />
                </div>

                <div>
                  <label class="block text-xs font-bold text-gray-400 uppercase mb-1 ml-1">Confirmar contraseña</label>
                  <PasswordInputComponent
                    required
                    value={confirm()}
                    onInput={(e) => setConfirm(e.currentTarget.value)}
                    placeholder="Repite la contraseña"
                    class="w-full bg-colpsi-surface border-2 border-transparent focus:border-colpsi-yellow focus:bg-white rounded-xl px-4 py-3.5 outline-none transition-all text-gray-700"
                  />
                </div>

                <div class="pt-2">
                  <button
                    type="submit"
                    disabled={loading()}
                    class="w-full bg-colpsi-yellow hover:bg-colpsi-yellow-dark text-colpsi-blue font-extrabold py-4 rounded-xl shadow-md shadow-yellow-500/20 active:scale-[0.98] transition-all disabled:opacity-50 disabled:active:scale-100 flex items-center justify-center gap-2"
                  >
                    <Show when={loading()} fallback="GUARDAR CONTRASEÑA">
                      <svg class="animate-spin h-5 w-5 text-colpsi-blue" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      GUARDANDO...
                    </Show>
                  </button>
                </div>
              </form>
            </Show>

            {/* Éxito */}
            <Show when={done()}>
              <div class="text-center">
                <div class="inline-flex items-center justify-center w-16 h-16 bg-green-50 rounded-full mb-4">
                  <svg class="w-8 h-8 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <h2 class="text-lg font-bold text-colpsi-blue mb-2">Contraseña actualizada</h2>
                <p class="text-sm text-gray-600 leading-relaxed">
                  Ya puedes iniciar sesión con tu nueva contraseña. Te redirigimos al portal...
                </p>
              </div>
            </Show>
          </div>
        </div>

        {/* FOOTER LINKS */}
        <div class="mt-8 text-center text-gray-400 text-sm">
          <p>¿No estás registrado? <A href="/directorio" class="text-colpsi-blue font-bold">Consulta el directorio</A></p>
        </div>
      </div>

      {/* LA FRANJA DINÁMICA (tira venezolana) */}
      <div class="fixed bottom-0 left-0 w-full h-3 flex overflow-hidden shadow-[0_-4px_15px_rgba(0,0,0,0.1)]">
        <div class="relative flex-1 bg-colpsi-red">
          <div class="absolute inset-0 bg-linear-to-r from-transparent via-white/50 to-transparent animate-flag-flow" />
        </div>
        <div class="relative flex-1 bg-green-700">
          <div class="absolute inset-0 bg-linear-to-r from-transparent via-white/50 to-transparent animate-flag-flow [animation-delay:1s]" />
        </div>
        <div class="relative flex-1 bg-colpsi-blue">
          <div class="absolute inset-0 bg-linear-to-r from-transparent via-white/50 to-transparent animate-flag-flow [animation-delay:2s]" />
        </div>
      </div>
    </main>
  );
}