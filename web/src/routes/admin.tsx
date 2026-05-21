// web/src/routes/admin.tsx
import { JSX, createSignal, Show, createEffect } from "solid-js";
import { A, useNavigate } from "@solidjs/router";
import { useAuth } from "~/lib/auth";

export default function AdminLayout(props: { children: JSX.Element }) {
  const { role, isAuthenticated, user, logout } = useAuth();
  const navigate = useNavigate();
  
  const [isCollapsed, setIsCollapsed] = createSignal(false);

  createEffect(() => {
    if (!isAuthenticated() || role() !== "admin") {
      navigate("/admin-access", { replace: true });
    }
  });

  const menuItems = [
    { title: "Dashboard", path: "/admin", icon: "📊" },
    { title: "Psicólogos", path: "/admin/psicologos", icon: "👥" },
    { title: "Areas de Ejercicio Psi", path: "/admin/areas_de_ejercicio_profesional", icon: "🔖" },
    { title: "Noticias", path: "/admin/noticias", icon: "📰" },
    { title: "Staff", path: "/admin/staff", icon: "🛡️" },
  ];

  return (
    <Show 
      when={isAuthenticated() && role() === "admin"} 
      fallback={<div class="flex items-center justify-center h-screen font-black text-colpsi-blue">Verificando...</div>}
    >
      <div class="min-h-screen bg-[#f8fafc] flex font-sans overflow-hidden">
        
        {/* SIDEBAR */}
        <aside class={`hidden md:flex flex-col bg-colpsi-blue text-white shadow-2xl z-20 transition-all duration-300 relative ${isCollapsed() ? "w-20" : "w-72"}`}>
          <button onClick={() => setIsCollapsed(!isCollapsed())} class="absolute -right-4 top-8 bg-colpsi-yellow text-colpsi-blue w-8 h-8 rounded-full flex items-center justify-center shadow-md z-30 border-4 border-[#f8fafc]">
            {isCollapsed() ? "▶" : "◀"}
          </button>

          <div class="h-20 flex items-center justify-center border-b border-blue-800/50">
            <A href="/admin" class="flex items-center gap-3 px-4">
              <div class="w-10 h-10 bg-white rounded-xl flex items-center justify-center shrink-0">
                <span class="text-colpsi-blue text-2xl font-black">Ψ</span>
              </div>
              {!isCollapsed() && <span class="text-xl font-black uppercase tracking-widest">Admin</span>}
            </A>
          </div>
          
          <nav class="flex-grow p-4 space-y-2 overflow-y-auto">
            {menuItems.map((item) => (
              <A href={item.path} end={item.path === "/admin"} class="flex items-center px-3 py-3.5 rounded-xl text-blue-100 hover:bg-blue-800 transition-all" activeClass="bg-colpsi-yellow !text-colpsi-blue font-black shadow-lg">
                <span class={`text-xl ${isCollapsed() ? "mx-auto" : "mr-4"}`}>{item.icon}</span>
                {!isCollapsed() && <span>{item.title}</span>}
              </A>
            ))}
          </nav>

          <div class="p-4 border-t border-blue-800/50">
             <button onClick={logout} class="w-full flex items-center justify-center gap-2 bg-red-500 py-3 rounded-xl font-bold">
               <span>🚪</span> {!isCollapsed() && "Cerrar Sesión"}
             </button>
          </div>
        </aside>

        {/* CONTENIDO */}
        <div class="flex-grow flex flex-col h-screen overflow-y-auto relative">
          <header class="hidden md:flex h-20 items-center justify-end px-10 bg-white border-b sticky top-0 z-10">
             <span class="text-xs font-bold text-gray-400 uppercase tracking-widest">Sistema en línea</span>
          </header>

          <main class="p-4 md:p-8 lg:p-10 flex-grow max-w-7xl mx-auto w-full">
            {/* LAS PÁGINAS SE CARGAN AQUÍ */}
            {props.children}
          </main>
        </div>

      </div>
    </Show>
  );
}