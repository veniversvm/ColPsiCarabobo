// web/src/routes/forgot-password.tsx
//
// Recuperación de contraseña — paso 1: el psicólogo introduce su correo y
// recibe un enlace de un solo uso (token). Respuesta genérica por diseño
// (no revela si el correo existe en el sistema).
import { createSignal, Show } from "solid-js";
import { A } from "@solidjs/router";
import { apiPost, ApiError } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";

export default function ForgotPasswordPage() {
  const [email, setEmail] = createSignal("");
  const [loading, setLoading] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);
  const [done, setDone] = createSignal(false);

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      await apiPost("/psi/forgot-password", { email: email() });
      setDone(true);
    } catch (err) {
      if (err instanceof ApiError) {
        // 429 = rate-limit: avisar el bloqueo real con el tiempo de espera
        setError(err.status === 429 ? (err.data?.message || err.message) : getUserFacingError(err));
      } else {
        setError("Ocurrió un error inesperado al intentar enviar la solicitud.");
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
            <h1 class="text-white text-xl font-bold tracking-tight">Recuperar contraseña</h1>
            <p class="text-blue-200 text-sm mt-1">Colegio de Psicólogos de Carabobo</p>
          </div>

          {/* FORM / SUCCESS SECTION */}
          <div class="p-8">
            <Show
              when={done()}
              fallback={
                <form class="space-y-5" onSubmit={handleSubmit}>

                  <Show when={error()}>
                    <div class="bg-red-50 border-l-4 border-colpsi-red p-4 animate-shake">
                      <p class="text-xs text-colpsi-red font-bold uppercase tracking-wide">Error</p>
                      <p class="text-sm text-red-700">{error()}</p>
                    </div>
                  </Show>

                  <div class="space-y-4">
                    <div>
                      <label class="block text-xs font-bold text-gray-400 uppercase mb-1 ml-1">Correo registrado</label>
                      <input
                        type="email"
                        required
                        placeholder="tu.correo@ejemplo.com"
                        class="w-full bg-colpsi-surface border-2 border-transparent focus:border-colpsi-yellow focus:bg-white rounded-xl px-4 py-3.5 outline-none transition-all text-gray-700"
                        onInput={(e) => setEmail(e.currentTarget.value)}
                      />
                    </div>
                  </div>

                  <div class="pt-2">
                    <button
                      type="submit"
                      disabled={loading()}
                      class="w-full bg-colpsi-yellow hover:bg-colpsi-yellow-dark text-colpsi-blue font-extrabold py-4 rounded-xl shadow-md shadow-yellow-500/20 active:scale-[0.98] transition-all disabled:opacity-50 disabled:active:scale-100 flex items-center justify-center gap-2"
                    >
                      <Show when={loading()} fallback="ENVIAR ENLACE">
                        <svg class="animate-spin h-5 w-5 text-colpsi-blue" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        ENVIANDO...
                      </Show>
                    </button>
                  </div>

                  <div class="text-center pt-4">
                    <A href="/login" class="text-sm text-gray-400 hover:text-colpsi-blue transition-colors">
                      Volver al inicio de sesión
                    </A>
                  </div>
                </form>
              }
            >
              <div class="text-center">
                <div class="inline-flex items-center justify-center w-16 h-16 bg-green-50 rounded-full mb-4">
                  <svg class="w-8 h-8 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <h2 class="text-lg font-bold text-colpsi-blue mb-2">Revisa tu correo</h2>
                <p class="text-sm text-gray-600 leading-relaxed">
                  Si el correo <strong class="text-gray-800">{email()}</strong> está registrado,
                  recibirás un enlace para restablecer tu contraseña.
                  El enlace expira en 1 hora y solo puede usarse una vez.
                </p>
                <div class="pt-6">
                  <A href="/login" class="text-sm text-colpsi-blue font-bold hover:underline">
                    Volver al inicio de sesión
                  </A>
                </div>
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