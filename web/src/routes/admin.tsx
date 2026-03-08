// web/src/routes/admin.tsx

import { JSX, createSignal, Show, createEffect } from "solid-js";
import { A, useNavigate } from "@solidjs/router";
import { useAuth } from "~/lib/auth";

export default function AdminLayout(props: { children: JSX.Element }) {
  const { role, isAuthenticated, user, logout } = useAuth();
  const navigate = useNavigate();
  const [isSidebarOpen, setIsSidebarOpen] = createSignal(false);

  // PROTECCIÓN DE RUTA (Guardia de Seguridad Frontend)
  createEffect(() => {
    // Si no está logueado, o si está logueado pero NO es admin, lo expulsamos
    if (!isAuthenticated() || role() !== "admin") {
      navigate("/login", { replace: true });
    }
  });

  const menuItems =[
    { title: "Dashboard", path: "/admin", icon: "📊" },
    { title: "Psicólogos", path: "/admin/psicologos", icon: "👥" },
    { title: "Especialidades", path: "/admin/especialidades", icon: "🔖" },
    { title: "Noticias", path: "/admin/noticias", icon: "📰" },
    { title: "Staff (SUDO)", path: "/admin/staff", icon: "🛡️" },
  ];

  return (
    <Show when={isAuthenticated() && role() === "admin"}>
      <div class="min-h-screen bg-gray-50 flex font-sans">
        
        {/* --- SIDEBAR (Escritorio) --- */}
        <aside class="hidden md:flex w-64 flex-col bg-colpsi-blue text-white shadow-xl z-20">
          <div class="h-20 flex items-center justify-center border-b border-blue-800">
            <span class="text-2xl font-black tracking-widest flex items-center gap-2">
              <span class="text-colpsi-yellow">Ψ</span> ADMIN
            </span>
          </div>
          
          <nav class="flex-grow p-4 space-y-2">
            {menuItems.map((item) => (
              <A
                href={item.path}
                end={item.path === "/admin"}
                class="flex items-center gap-3 px-4 py-3 rounded-xl text-blue-100 hover:bg-blue-800 hover:text-white transition-colors"
                activeClass="bg-colpsi-yellow text-colpsi-blue font-bold shadow-md hover:bg-yellow-400 hover:text-colpsi-blue"
              >
                <span class="text-lg">{item.icon}</span>
                <span>{item.title}</span>
              </A>
            ))}
          </nav>

          <div class="p-4 border-t border-blue-800">
            <div class="flex items-center gap-3 mb-4 px-2">
              <div class="w-10 h-10 rounded-full bg-blue-800 flex items-center justify-center font-bold">
                {user()?.username.charAt(0).toUpperCase()}
              </div>
              <div class="overflow-hidden">
                <p class="text-sm font-bold truncate">{user()?.username}</p>
                <p class="text-xs text-blue-300 truncate">{user()?.email}</p>
              </div>
            </div>
            <button 
              onClick={logout}
              class="w-full flex items-center justify-center gap-2 bg-red-500 hover:bg-red-600 text-white py-2.5 rounded-xl font-bold transition-colors"
            >
              <span>🚪</span> Cerrar Sesión
            </button>
          </div>
        </aside>

        {/* --- MOBILE HEADER & OVERLAY --- */}
        <div class="md:hidden fixed top-0 left-0 right-0 h-16 bg-colpsi-blue text-white flex items-center justify-between px-4 z-50 shadow-md">
          <span class="font-black tracking-widest">Ψ COLPSI ADMIN</span>
          <button onClick={() => setIsSidebarOpen(!isSidebarOpen())} class="p-2 bg-blue-800 rounded-lg">
            {isSidebarOpen() ? "✕" : "☰"}
          </button>
        </div>

        <Show when={isSidebarOpen()}>
          <div class="md:hidden fixed inset-0 z-40 bg-colpsi-blue flex flex-col pt-16">
            <nav class="p-4 space-y-2 flex-grow">
              {menuItems.map((item) => (
                <A
                  href={item.path}
                  end={item.path === "/admin"}
                  onClick={() => setIsSidebarOpen(false)}
                  class="flex items-center gap-3 px-4 py-4 rounded-xl text-blue-100 hover:bg-blue-800"
                  activeClass="bg-colpsi-yellow text-colpsi-blue font-bold"
                >
                  <span class="text-xl">{item.icon}</span>
                  <span class="text-lg">{item.title}</span>
                </A>
              ))}
            </nav>
            <div class="p-6 border-t border-blue-800">
              <button onClick={logout} class="w-full bg-red-500 text-white py-4 rounded-xl font-bold">
                Cerrar Sesión
              </button>
            </div>
          </div>
        </Show>

        {/* --- MAIN CONTENT AREA --- */}
        <div class="flex-grow flex flex-col min-h-screen pt-16 md:pt-0 overflow-y-auto">
          {/* El 'props.children' inyectará las páginas específicas (index, psicologos, etc.) */}
          <div class="p-6 md:p-10 flex-grow">
            {props.children}
          </div>
          
          {/* Footer interno del Admin */}
          <footer class="py-4 text-center text-xs text-gray-400 border-t border-gray-200 bg-white">
            Colegio de Psicólogos del Estado Carabobo - Panel Administrativo
          </footer>
        </div>

      </div>
    </Show>
  );
}