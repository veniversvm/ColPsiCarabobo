// web/src/routes/admin.tsx
import { JSX, createSignal, Show, createEffect, createResource } from "solid-js";
import { A, useNavigate } from "@solidjs/router";
import { useAuth } from "~/lib/auth";
import { apiGet } from "~/lib/api";
import type { PendientesResponse } from "~/types/tickets";
import type { PermissionState } from "~/lib/staff-permissions";

interface AdminMe {
  id: string;
  username: string;
  email: string;
  is_active: boolean;
  sudo: boolean;
  role?: string | null;
  permissions: PermissionState;
}

export default function AdminLayout(props: { children: JSX.Element }) {
  const { role, isAuthenticated, user, logout } = useAuth();
  const navigate = useNavigate();
  
  const [isCollapsed, setIsCollapsed] = createSignal(false);

  createEffect(() => {
    if (!isAuthenticated() || role() !== "admin") {
      navigate("/admin-access", { replace: true });
    }
  });

  // Estado y permisos del admin autenticado (para filtrar el menú).
  // El backend sigue validando cada operación: esto es solo cosmético/UX.
  const [me] = createResource<AdminMe | null>(
    async () => {
      try {
        return await apiGet<AdminMe>("/admin/me");
      } catch {
        return null;
      }
    },
    { initialValue: null }
  );

  // Badge de tickets pendientes (polling cada 30s). Silencioso si falla.
  const [pendientes] = createResource(
    () => "",
    async (_k) => {
      try {
        return await apiGet<PendientesResponse>("/admin/tickets/pendientes-count");
      } catch {
        return { pendientes: 0 };
      }
    },
    { initialValue: { pendientes: 0 }, refetchInterval: 30000 }
  );
  const ticketsPendientes = () => pendientes()?.pendientes ?? 0;

  const hasAny = (keys: (keyof PermissionState)[]): boolean => {
    const m = me();
    if (!m) return false;
    if (m.sudo) return true;
    return keys.some((k) => m.permissions[k]);
  };

  // Visibilidad del menú por permisos (el backend sigue siendo la barrera real).
  const menuItems = [
    { title: "Dashboard", path: "/admin", icon: "📊", always: true, perms: [] as (keyof PermissionState)[] },
    { title: "Psicólogos", path: "/admin/psicologos", icon: "👥", always: false, perms: ["can_read_psi", "can_create_psi", "can_update_psi", "can_delete_psi"] },
    { title: "Inscripciones", path: "/admin/inscripciones", icon: "📝", always: false, perms: ["can_read_psi", "can_create_psi", "can_update_psi", "can_delete_psi"] },
    { title: "Areas de Ejercicio Psi", path: "/admin/areas_de_ejercicio_profesional", icon: "🔖", always: false, perms: ["can_create_tags", "can_edit_tags", "can_delete_tags"] },
    { title: "Noticias", path: "/admin/noticias", icon: "📰", always: false, perms: ["can_publish", "can_update_publish", "can_delete_publish"] },
    { title: "Notificaciones", path: "/admin/notificaciones", icon: "🔔", always: false, perms: ["can_send_notifications", "can_manage_notifications", "can_read_notifications"] },
    { title: "Tickets", path: "/admin/tickets", icon: "🎫", always: false, perms: ["can_manage_tickets"] },
    { title: "Proyectos", path: "/admin/proyectos", icon: "📋", always: false, perms: ["can_manage_projects"] },
    { title: "Staff", path: "/admin/staff", icon: "🛡️", always: false, perms: ["can_create_admin", "can_update_admin", "can_delete_admin"] },
  ];

  const visibleMenu = () =>
    menuItems.filter((item) => item.always || hasAny(item.perms));

  return (
    <Show 
      when={isAuthenticated() && role() === "admin"} 
      fallback={<div class="flex items-center justify-center h-screen font-black text-colpsi-blue">Verificando...</div>}
    >
      <div class="min-h-screen bg-colpsi-bg flex font-sans overflow-hidden">
        
        {/* SIDEBAR */}
        <aside class={`hidden md:flex flex-col bg-colpsi-blue text-white shadow-2xl z-20 transition-all duration-300 relative ${isCollapsed() ? "w-20" : "w-72"}`}>
          <button onClick={() => setIsCollapsed(!isCollapsed())} class="absolute -right-4 top-8 bg-colpsi-yellow text-colpsi-blue w-8 h-8 rounded-full flex items-center justify-center shadow-md z-30 border-4 border-colpsi-bg">
            {isCollapsed() ? "▶" : "◀"}
          </button>

          <div class="h-20 flex items-center justify-center border-b border-blue-800/50 shrink-0">
            <A href="/admin" class="flex items-center px-4">
              {!isCollapsed() && <span class="text-xl font-black uppercase tracking-widest">Admin</span>}
            </A>
          </div>
          
          <nav class="p-4 space-y-2 pb-2">
            {visibleMenu().map((item) => (
              <A href={item.path} end={item.path === "/admin"} class="flex items-center px-3 py-3.5 rounded-xl text-blue-100 hover:bg-blue-800 transition-all" activeClass="bg-colpsi-yellow !text-colpsi-blue font-black shadow-lg">
                <span class={`text-xl ${isCollapsed() ? "mx-auto" : "mr-4"}`}>{item.icon}</span>
                {!isCollapsed() && <span>{item.title}</span>}
                <Show when={item.path === "/admin/tickets" && ticketsPendientes() > 0}>
                  <span class="ml-auto bg-red-500 text-white text-[10px] font-black min-w-6 h-6 px-1.5 rounded-full flex items-center justify-center shadow-lg">
                    {ticketsPendientes() > 99 ? "99+" : ticketsPendientes()}
                  </span>
                </Show>
              </A>
            ))}
          </nav>

          <div class="shrink-0 px-4 pt-3 pb-4 border-t border-blue-800/50">
            <button onClick={logout} class="w-full flex items-center justify-center gap-2 bg-red-500 py-3 rounded-xl font-bold">
              <span>🚪</span> {!isCollapsed() && "Cerrar Sesión"}
            </button>
          </div>
        </aside>

        {/* CONTENIDO */}
        <div class="flex-grow flex flex-col h-screen overflow-y-auto relative">
          <main class="p-4 md:p-8 lg:p-10 flex-grow max-w-7xl mx-auto w-full">
            {/* LAS PÁGINAS SE CARGAN AQUÍ */}
            {props.children}
          </main>
        </div>

      </div>
    </Show>
  );
}