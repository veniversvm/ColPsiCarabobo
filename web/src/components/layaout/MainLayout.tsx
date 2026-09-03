// web/src/components/layaout/MainLayout.tsx

import { JSX, createSignal, Show } from "solid-js";
import { useAuth } from "~/lib/auth";

export default function MainLayout(props: { children: JSX.Element }) {
  const [isMenuOpen, setIsMenuOpen] = createSignal(false);
  const { isAuthenticated, logout, user } = useAuth();

  return (
    <div class="min-h-screen bg-white flex flex-col font-sans">
      {/* HEADER / NAVBAR */}
      <header class="bg-white border-b border-gray-100 sticky top-0 z-50">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div class="flex justify-between h-16 items-center">
            {/* Logo e Identidad */}
            <div class="flex items-center space-x-2">
              <div class="w-8 h-8 bg-colpsi-blue rounded-full flex items-center justify-center text-white font-bold">
                Ψ
              </div>
              <span class="text-colpsi-blue font-bold text-lg leading-tight hidden sm:block">
                Colegio de Psicólogos
                <span class="text-gray-400">del Estado Carabobo</span>
              </span>
            </div>

            {/* Desktop Navigation */}
            <nav class="hidden md:flex space-x-8 items-center">
              <a
                href="/directorio"
                class="text-gray-600 hover:text-colpsi-blue font-medium"
              >
                Directorio
              </a>
              <a
                href="/noticias"
                class="text-gray-600 hover:text-colpsi-blue font-medium"
              >
                Noticias
              </a>
              <Show when={!isAuthenticated()}>
                <a
                  href="/login"
                  class="bg-colpsi-blue text-white px-4 py-2 rounded-lg font-semibold hover:bg-opacity-90 transition-all"
                >
                  Ingresar
                </a>
              </Show>
              <Show when={isAuthenticated()}>
                <button onClick={logout} class="text-colpsi-red font-medium">
                  Salir
                </button>
              </Show>
            </nav>

            {/* Mobile Menu Button */}
            <div class="md:hidden flex items-center">
              <button
                onClick={() => setIsMenuOpen(!isMenuOpen())}
                class="text-colpsi-blue p-2"
              >
                <svg
                  class="w-6 h-6"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d={
                      isMenuOpen()
                        ? "M6 18L18 6M6 6l12 12"
                        : "M4 6h16M4 12h16M4 18h16"
                    }
                  />
                </svg>
              </button>
            </div>
          </div>
        </div>

        {/* Mobile Menu Overlay */}
        <Show when={isMenuOpen()}>
          <div class="md:hidden bg-white border-b border-gray-100 px-4 pt-2 pb-6 space-y-2 shadow-xl">
            <a
              href="/directorio"
              class="block py-3 text-gray-700 font-medium border-b border-gray-50"
            >
              Directorio
            </a>
            <a
              href="/noticias"
              class="block py-3 text-gray-700 font-medium border-b border-gray-50"
            >
              Noticias
            </a>
            <a
              href="/nosotros"
              class="block py-3 text-gray-700 font-medium border-b border-gray-50"
            >
              Nosotros
            </a>
            <a
              href="/documentos"
              class="block py-3 text-gray-700 font-medium border-b border-gray-50"
            >
              Marco Legal
            </a>
            <Show when={!isAuthenticated()}>
              <a href="/login" class="block py-3 text-colpsi-blue font-bold">
                Iniciar Sesión
              </a>
            </Show>
            <Show when={isAuthenticated()}>
              <a href="/psi/me" class="block py-3 text-colpsi-blue font-bold">
                Mi Perfil
              </a>
              <button
                onClick={logout}
                class="block py-3 text-colpsi-red font-bold"
              >
                Cerrar Sesión
              </button>
            </Show>
          </div>
        </Show>
      </header>

      {/* MAIN CONTENT AREA */}
      <main class="flex-grow">{props.children}</main>

      {/* FOOTER SIMPLE */}
      <footer class="bg-gray-50 py-8 border-t border-gray-100">
        <div class="text-center text-gray-400 text-xs px-4">
          <p>© 2026 Colegio de Psicólogos del Estado Carabobo</p>
          <p class="mt-2">Valencia, Venezuela</p>
        </div>
      </footer>
    </div>
  );
}
