// web/src/components/layout/Navbar.tsx
import { useAuth } from "~/lib/auth";
import { Show, createSignal } from "solid-js";
import { A } from "@solidjs/router";

export default function Navbar() {
  const { user, isAuthenticated, logout, role } = useAuth();
  const [isOpen, setIsOpen] = createSignal(false);

  const navLinkClass = "text-[#1e3a8a] hover:bg-blue-50 md:hover:bg-transparent md:hover:text-blue-700 block px-3 py-4 md:py-0 rounded-md text-base font-medium transition-colors";

  return (
    <nav class="bg-white shadow-md sticky top-0 z-50 font-sans">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16 items-center">
          
          {/* LOGO SECTION */}
          <div class="flex-shrink-0 flex items-center gap-2">
            <A href="/explorar" class="flex items-center gap-2 group">
              <div class="bg-[#1e3a8a] text-white w-8 h-8 rounded flex items-center justify-center font-bold text-xl group-hover:scale-110 transition-transform">
                Ψ
              </div>
              <div class="flex flex-col leading-none">
                <span class="text-[#1e3a8a] font-extrabold text-lg tracking-tight">COLPSI</span>
                <span class="text-gray-400 text-[10px] font-bold tracking-widest uppercase">Carabobo</span>
              </div>
            </A>
          </div>

          {/* DESKTOP NAV (Oculto en móvil) */}
          <div class="hidden md:flex items-center space-x-8">
            <A href="/directorio" class="text-gray-600 hover:text-[#1e3a8a] font-medium transition-colors">Directorio</A>
            <A href="/noticias" class="text-gray-600 hover:text-[#1e3a8a] font-medium transition-colors">Noticias</A>
            
            <Show 
              when={isAuthenticated()} 
              fallback={
                <A href="/login" class="bg-[#facc15] text-[#1e3a8a] px-5 py-2 rounded-full font-bold shadow-sm hover:shadow-md hover:bg-[#fde047] transition-all">
                  Iniciar Sesión
                </A>
              }
            >
              <div class="flex items-center gap-4 border-l pl-6 border-gray-100">
                {/* Info de Usuario / Enlace a Perfil */}
                <A href={role() === 'admin' ? '/admin' : '/psi'} class="text-right flex flex-col group">
                  <span class="text-[10px] text-gray-400 font-bold uppercase tracking-tighter">Bienvenido(a)</span>
                  <span class="text-sm font-bold text-[#1e3a8a] group-hover:underline">
                    {user()?.firstName || user()?.username}
                  </span>
                </A>
                
                {/* Botón condicional según Rol */}
                <Show 
                  when={role() === "admin"}
                  fallback={
                    <A href="/psi/me" class="text-xs bg-blue-50 text-colpsi-blue px-3 py-1.5 rounded-lg font-bold hover:bg-[#facc15] transition-colors uppercase">
                      Mi Perfil
                    </A>
                  }
                >
                  <A href="/admin" class="text-xs bg-red-100 text-red-700 px-3 py-1.5 rounded-lg font-bold hover:bg-red-200 uppercase tracking-tighter">
                    Panel Admin
                  </A>
                </Show>

                <button 
                  onClick={logout}
                  class="text-gray-400 hover:text-red-600 transition-colors p-1"
                  title="Cerrar Sesión"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                  </svg>
                </button>
              </div>
            </Show>
          </div>

          {/* MOBILE MENU BUTTON */}
          <div class="flex md:hidden">
            <button 
              onClick={() => setIsOpen(!isOpen())}
              class="inline-flex items-center justify-center p-2 rounded-md text-[#1e3a8a] hover:bg-blue-50 focus:outline-none"
            >
              <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path 
                  stroke-linecap="round" 
                  stroke-linejoin="round" 
                  stroke-width="2" 
                  d={isOpen() ? "M6 18L18 6M6 6l12 12" : "M4 6h16M4 12h16M4 18h16"} 
                />
              </svg>
            </button>
          </div>
        </div>
      </div>

      {/* MOBILE MENU CONTENT (Dropdown) */}
      <Show when={isOpen()}>
        <div class="md:hidden bg-white border-t border-gray-50 shadow-2xl animate-in slide-in-from-top duration-300">
          <div class="px-4 pt-4 pb-6 space-y-2">
            <A href="/directorio" onClick={() => setIsOpen(false)} class={navLinkClass}>Directorio Profesional</A>
            <A href="/noticias" onClick={() => setIsOpen(false)} class={navLinkClass}>Noticias y Avisos</A>
            
            <div class="my-4 border-t border-gray-100 pt-4">
              <Show 
                when={isAuthenticated()} 
                fallback={
                  <A href="/login" onClick={() => setIsOpen(false)} class="block w-full text-center bg-[#facc15] text-[#1e3a8a] px-4 py-4 rounded-2xl font-black shadow-lg shadow-yellow-500/20">
                    INICIAR SESIÓN
                  </A>
                }
              >
                <div class="space-y-4">
                  <div class="flex items-center gap-4 bg-gray-50 p-4 rounded-2xl">
                    <div class="w-12 h-12 bg-colpsi-blue rounded-xl flex items-center justify-center text-white font-black text-xl">
                      {user()?.username.charAt(0).toUpperCase()}
                    </div>
                    <div class="overflow-hidden">
                      <p class="text-sm font-black text-colpsi-blue truncate">
                        {user()?.firstName} {user()?.lastName}
                      </p>
                      <p class="text-xs text-gray-400 truncate">{user()?.email}</p>
                    </div>
                  </div>
                  
                  <Show when={role() === "admin"}>
                    <A href="/admin" onClick={() => setIsOpen(false)} class="block w-full bg-red-50 text-red-700 px-4 py-3 rounded-xl font-bold text-center border border-red-100">
                      Panel Administrativo
                    </A>
                  </Show>
                  
                  <A href="/psi/me" onClick={() => setIsOpen(false)} class="block w-full bg-blue-50 text-[#1e3a8a] px-4 py-3 rounded-xl font-bold text-center border border-blue-100">
                    Gestionar Mi Perfil
                  </A>
                  
                  <button 
                    onClick={() => { logout(); setIsOpen(false); }}
                    class="w-full text-red-500 font-bold py-3 text-sm flex items-center justify-center gap-2"
                  >
                    <span>Cerrar Sesión</span>
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" /></svg>
                  </button>
                </div>
              </Show>
            </div>
          </div>
        </div>
      </Show>
    </nav>
  );
}