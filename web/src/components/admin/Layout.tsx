// web/src/components/admin/Layout.tsx
import { JSX, createSignal, Show, createEffect } from "solid-js";
import { A, useNavigate } from "@solidjs/router";
import { useAuth } from "~/lib/auth";
import { Animate, Presence } from "~/components/ui/Motion";
export default function AdminLayout(props: { children: JSX.Element }) {
  const { role, isAuthenticated, user, logout } = useAuth();
  const navigate = useNavigate();
  
  const [isMobileOpen, setIsMobileOpen] = createSignal(false);
  const [isCollapsed, setIsCollapsed] = createSignal(false);

  // PROTECCIÓN DE RUTA
  createEffect(() => {
    if (!isAuthenticated() || role() !== "admin") {
      navigate("/", { replace: true });
    }
  });

  const menuItems =[
    { title: "Dashboard", path: "/admin", icon: "📊" },
    { title: "Psicólogos", path: "/admin/psicologos", icon: "👥" },
    { title: "Areas de Ejercicio", path: "/admin/areas_de_ejercicio_profesional", icon: "🔖" },
    { title: "Noticias", path: "/admin/noticias", icon: "📰" },
    { title: "Notificaciones", path: "/admin/notificaciones", icon: "🔔" },
    { title: "Staff (SUDO)", path: "/admin/staff", icon: "🛡️" },
  ];

  return (
    <Show when={isAuthenticated() && role() === "admin"}>
      <div class="min-h-screen bg-[#f8fafc] flex font-sans overflow-hidden">
        
        {/* SIDEBAR DE ESCRITORIO */}
        <aside 
          class={`hidden md:flex flex-col bg-colpsi-blue text-white shadow-2xl z-20 transition-all duration-300 ease-in-out relative ${
            isCollapsed() ? "w-20" : "w-72"
          }`}
        >
          {/* Botón para colapsar/expandir */}
          <button 
            onClick={() => setIsCollapsed(!isCollapsed())}
            class="absolute -right-4 top-8 bg-colpsi-yellow text-colpsi-blue w-8 h-8 rounded-full flex items-center justify-center shadow-md hover:scale-110 transition-transform z-30 font-black border-4 border-[#f8fafc]"
            title={isCollapsed() ? "Expandir menú" : "Colapsar menú"}
          >
            {isCollapsed() ? "▶" : "◀"}
          </button>

          {/* Header del Sidebar */}
          <div class="h-20 flex items-center justify-center border-b border-blue-800/50 shrink-0">
            <A href="/admin" class="flex items-center gap-3 overflow-hidden px-4">
              <div class="w-10 h-10 bg-white rounded-xl flex items-center justify-center shadow-inner shrink-0">
                <span class="text-colpsi-blue text-2xl font-black">Ψ</span>
              </div>
              <Show when={!isCollapsed()}>
                <span class="text-xl font-black tracking-widest uppercase whitespace-nowrap">
                  Admin
                </span>
              </Show>
            </A>
          </div>
          
          {/* Enlaces de Navegación */}
          <nav class="flex-grow p-4 space-y-2 overflow-y-auto overflow-x-hidden scrollbar-thin scrollbar-thumb-blue-800">
            {menuItems.map((item) => (
              <A
                href={item.path}
                end={item.path === "/admin"}
                title={isCollapsed() ? item.title : ""}
                class="flex items-center px-3 py-3.5 rounded-xl text-blue-100 hover:bg-blue-800 hover:text-white transition-all group overflow-hidden"
                activeClass="bg-colpsi-yellow !text-colpsi-blue font-black shadow-lg"
              >
                <span class={`text-xl shrink-0 ${isCollapsed() ? "mx-auto" : "mr-4"}`}>
                  {item.icon}
                </span>
                <Show when={!isCollapsed()}>
                  <span class="truncate">
                    {item.title}
                  </span>
                </Show>
              </A>
            ))}
          </nav>

          {/* Perfil y Logout */}
          <div class="p-4 border-t border-blue-800/50 shrink-0">
            <Show 
              when={!isCollapsed()} 
              fallback={
                <button onClick={logout} title="Cerrar Sesión" class="w-full flex justify-center py-3 bg-red-500/20 text-red-400 hover:bg-red-500 hover:text-white rounded-xl transition-colors">
                  <span class="text-xl">🚪</span>
                </button>
              }
            >
              <div class="flex items-center justify-between bg-blue-800/30 p-3 rounded-2xl mb-3">                <div class="flex items-center gap-3 overflow-hidden">
                  <div class="w-10 h-10 rounded-full bg-colpsi-yellow text-colpsi-blue flex items-center justify-center font-black shrink-0">
                    {user()?.username.charAt(0).toUpperCase()}
                  </div>
                  <div class="overflow-hidden">
                    <p class="text-sm font-bold text-white truncate">{user()?.username}</p>
                    <p class="text-[10px] text-blue-300 truncate uppercase tracking-widest">Superuser</p>
                  </div>
                </div>
              </div>
              <button 
                onClick={logout}
                class="w-full flex items-center justify-center gap-2 bg-red-500 hover:bg-red-600 text-white py-3 rounded-xl font-bold transition-all shadow-md active:scale-95"
              >
                <span>🚪</span> Cerrar Sesión
              </button>
            </Show>
          </div>
        </aside>

        {/* HEADER MÓVIL */}
        <div class="md:hidden fixed top-0 left-0 right-0 h-16 bg-colpsi-blue text-white flex items-center justify-between px-4 z-50 shadow-lg">
          <div class="flex items-center gap-2">
             <span class="text-colpsi-yellow text-2xl font-black">Ψ</span>
             <span class="font-black tracking-widest uppercase text-sm">Panel Admin</span>
          </div>
          <button onClick={() => setIsMobileOpen(!isMobileOpen())} class="p-2 bg-blue-800 rounded-xl hover:bg-blue-700 transition-colors">
            {isMobileOpen() ? "✕" : "☰"}
          </button>
        </div>

        {/* MENÚ MÓVIL */}
        <Presence>
          <Show when={isMobileOpen()}>
            <Animate variant="slide-top" class="md:hidden fixed inset-0 z-40 bg-colpsi-blue flex flex-col pt-16" exit={{ opacity: 0, y: -20 }}>
              <nav class="p-6 space-y-3 flex-grow overflow-y-auto">
                {menuItems.map((item) => (
                  <A
                    href={item.path}
                    end={item.path === "/admin"}
                    onClick={() => setIsMobileOpen(false)}
                    class="flex items-center gap-4 px-5 py-4 rounded-2xl text-blue-100 hover:bg-blue-800 bg-blue-900/20"
                    activeClass="bg-colpsi-yellow !text-colpsi-blue font-black shadow-lg"
                  >
                    <span class="text-2xl">{item.icon}</span>
                    <span class="text-lg tracking-wide">{item.title}</span>
                  </A>
                ))}
              </nav>
              <div class="p-6 border-t border-blue-800 bg-blue-900/40">
                <div class="flex items-center gap-4 mb-6">
                  <div class="w-12 h-12 rounded-full bg-colpsi-yellow text-colpsi-blue flex items-center justify-center font-black text-xl">
                    {user()?.username.charAt(0).toUpperCase()}
                  </div>
                  <div>
                    <p class="text-lg font-bold text-white">{user()?.username}</p>
                    <p class="text-sm text-blue-300">{user()?.email}</p>
                  </div>
                </div>
                <button onClick={logout} class="w-full flex items-center justify-center gap-2 bg-red-500 text-white py-4 rounded-2xl font-black text-lg active:scale-95 transition-transform shadow-lg">
                  <span>🚪</span> CERRAR SESIÓN
                </button>
              </div>
            </Animate>
          </Show>
        </Presence>

        {/* ÁREA DE CONTENIDO */}
        <div class="flex-grow flex flex-col h-screen pt-16 md:pt-0 overflow-y-auto relative">
          <header class="hidden md:flex h-20 items-center justify-end px-10 bg-white border-b border-gray-100 shrink-0 sticky top-0 z-10 shadow-sm">
             <div class="flex items-center gap-4">
                <span class="relative flex h-3 w-3">
                  <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                  <span class="relative inline-flex rounded-full h-3 w-3 bg-green-500"></span>
                </span>
                <span class="text-xs font-bold text-gray-400 uppercase tracking-widest">Sistema en línea</span>
             </div>
          </header>

          <main class="p-4 md:p-8 lg:p-10 flex-grow max-w-7xl mx-auto w-full">
            {props.children}
          </main>
          
          <footer class="py-6 text-center text-xs text-gray-400 border-t border-gray-200 bg-white shrink-0 mt-auto">
            © {new Date().getFullYear()} Colegio de Psicólogos del Estado Carabobo <br/> 
            <span class="opacity-50">Plataforma Administrativa Segura</span>
          </footer>
        </div>

      </div>
    </Show>
  );
}