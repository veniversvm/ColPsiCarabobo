// web/src/routes/admin.tsx
import { JSX, createSignal, Show, createEffect, createResource, For } from "solid-js";
import { A, useNavigate, useLocation } from "@solidjs/router";
import { useAuth } from "~/lib/auth";
import { apiGet } from "~/lib/api";
import type { PendientesResponse } from "~/types/tickets";
import type { PermissionState } from "~/lib/staff-permissions";
import { Icon } from "~/components/admin/ui/icons";

interface AdminMe {
  id: string;
  username: string;
  email: string;
  is_active: boolean;
  sudo: boolean;
  role?: string | null;
  permissions: PermissionState;
}

// Iconos del menú por ruta (SVG inline, hereda el color del texto).
const menuIcons: Record<string, string> = {
  "/admin": "grid",
  "/admin/psicologos": "users",
  "/admin/inscripciones": "fileText",
  "/admin/areas_de_ejercicio_profesional": "tag",
  "/admin/noticias": "newspaper",
  "/admin/notificaciones": "bell",
  "/admin/tickets": "ticket",
  "/admin/proyectos": "kanban",
  "/admin/staff": "shield",
};

export default function AdminLayout(props: { children: JSX.Element }) {
  const { role, isAuthenticated, user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

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
    { title: "Dashboard", path: "/admin", always: true, perms: [] as (keyof PermissionState)[], chip: "bg-colpsi-blue/10 text-colpsi-blue", label: "text-colpsi-blue" },
    { title: "Psicólogos", path: "/admin/psicologos", always: false, perms: ["can_read_psi", "can_create_psi", "can_update_psi", "can_delete_psi"], chip: "bg-indigo-100 text-indigo-600", label: "text-indigo-600" },
    { title: "Inscripciones", path: "/admin/inscripciones", always: false, perms: ["can_read_psi", "can_create_psi", "can_update_psi", "can_delete_psi"], chip: "bg-emerald-100 text-emerald-600", label: "text-emerald-600" },
    { title: "Areas de Ejercicio Psi", path: "/admin/areas_de_ejercicio_profesional", always: false, perms: ["can_create_tags", "can_edit_tags", "can_delete_tags"], chip: "bg-amber-100 text-amber-600", label: "text-amber-600" },
    { title: "Noticias", path: "/admin/noticias", always: false, perms: ["can_publish", "can_update_publish", "can_delete_publish"], chip: "bg-rose-100 text-rose-600", label: "text-rose-600" },
    { title: "Notificaciones", path: "/admin/notificaciones", always: false, perms: ["can_send_notifications", "can_manage_notifications", "can_read_notifications"], chip: "bg-violet-100 text-violet-600", label: "text-violet-600" },
    { title: "Tickets", path: "/admin/tickets", always: false, perms: ["can_manage_tickets"], chip: "bg-orange-100 text-orange-600", label: "text-orange-600" },
    { title: "Proyectos", path: "/admin/proyectos", always: false, perms: ["can_manage_projects"], chip: "bg-cyan-100 text-cyan-600", label: "text-cyan-600" },
    { title: "Staff", path: "/admin/staff", always: false, perms: ["can_create_admin", "can_update_admin", "can_delete_admin"], chip: "bg-teal-100 text-teal-600", label: "text-teal-600" },
  ];

  const visibleMenu = () =>
    menuItems.filter((item) => item.always || hasAny(item.perms));

  // Título de la sección actual para el breadcrumb del topbar.
  const currentTitle = () => {
    const path = location.pathname;
    const item = [...visibleMenu()].reverse().find(
      (m) => path === m.path || path.startsWith(`${m.path}/`)
    );
    return item?.title ?? "";
  };

  // ¿El item coincide con la ruta actual? (para colorear la etiqueta activa)
  const isMenuActive = (path: string) =>
    location.pathname === path || location.pathname.startsWith(`${path}/`);

  return (
    <Show
      when={isAuthenticated() && role() === "admin"}
      fallback={<div class="flex items-center justify-center h-screen font-black text-colpsi-blue">Verificando...</div>}
    >
      <div class="flex h-screen overflow-hidden bg-white font-sans text-colpsi-text">

        {/* ── SIDEBAR (256px, plano, sin sombras) ─────────────────────── */}
        <aside
          class={`${isCollapsed() ? "w-16" : "w-64"} hidden md:flex flex-col shrink-0 border-r border-colpsi-border bg-colpsi-bg transition-all duration-300`}
        >
          <div class="h-14 flex items-center gap-2.5 px-4 border-b border-colpsi-border shrink-0">
            <img src="/emblema.png" alt="Emblema" class="w-8 h-8 rounded-full object-cover" />
            <Show when={!isCollapsed()}>
              <span class="text-[15px] font-bold text-colpsi-blue truncate">Panel de gestión</span>
            </Show>
          </div>

          <nav class="flex-1 overflow-y-auto p-2 space-y-0.5">
            <For each={visibleMenu()}>
              {(item) => (
                <A
                  href={item.path}
                  end={item.path === "/admin"}
                  class={`flex items-center gap-3 h-10 rounded-md border-l-2 border-transparent text-[15px] font-medium text-slate-600 transition-colors hover:bg-white ${isCollapsed() ? "justify-center px-0" : "px-2"}`}
                  activeClass="bg-white border-l-colpsi-yellow font-semibold"
                >
                  <span class={`w-8 h-8 rounded-lg flex items-center justify-center shrink-0 ${item.chip}`}>
                    <Icon name={menuIcons[item.path] ?? "grid"} class="w-5 h-5" />
                  </span>
                  <Show when={!isCollapsed()}>
                    <span class={`truncate flex-1 ${isMenuActive(item.path) ? item.label : ""}`}>{item.title}</span>
                    <Show when={item.path === "/admin/tickets" && ticketsPendientes() > 0}>
                      <span class="min-w-4 h-4 px-1 rounded-full bg-colpsi-red text-white text-[10px] font-semibold flex items-center justify-center">
                        {ticketsPendientes() > 99 ? "99+" : ticketsPendientes()}
                      </span>
                    </Show>
                  </Show>
                </A>
              )}
            </For>
          </nav>

          <div class="shrink-0 border-t border-colpsi-border p-2">
            <button
              onClick={logout}
              class={`flex items-center gap-3 h-10 w-full rounded-md text-[15px] font-medium text-slate-600 transition-colors hover:bg-white hover:text-colpsi-red ${isCollapsed() ? "justify-center px-0" : "px-2"}`}
            >
              <Icon name="logout" class="w-5 h-5" />
              <Show when={!isCollapsed()}>
                <span>Salir de sesión</span>
              </Show>
            </button>
          </div>
        </aside>

        {/* ── COLUMNA DE TRABAJO ──────────────────────────────────────── */}
        <div class="flex-1 flex flex-col min-w-0">

          {/* TOPBAR (56px) */}
          <header class="h-14 shrink-0 border-b border-colpsi-border bg-white flex items-center gap-3 px-4">
            <button
              onClick={() => setIsCollapsed(!isCollapsed())}
              class="h-8 w-8 rounded-md text-slate-500 hover:bg-colpsi-bg hover:text-colpsi-blue flex items-center justify-center"
              title={isCollapsed() ? "Expandir menú" : "Colapsar menú"}
            >
              <Icon name={isCollapsed() ? "menu" : "panelLeft"} />
            </button>

            <nav class="text-xs text-colpsi-muted flex items-center gap-1.5 min-w-0">
              <span class="shrink-0">Admin</span>
              <Show when={currentTitle()}>
                <Icon name="chevronRight" class="w-3.5 h-3.5 text-slate-300 shrink-0" />
                <span class="text-colpsi-text font-medium truncate">{currentTitle()}</span>
              </Show>
            </nav>

            <div class="ml-auto flex items-center gap-2.5 min-w-0">
              <Show when={me()?.username}>
                <span class="hidden sm:block text-sm font-medium text-colpsi-muted truncate max-w-[200px]">{me()?.username}</span>
              </Show>
              <span class="h-8 w-8 rounded-full bg-colpsi-blue text-white flex items-center justify-center text-xs font-bold shrink-0">
                {(me()?.username ?? "A").slice(0, 1).toUpperCase()}
              </span>
            </div>
          </header>

          {/* ÁREA DE TRABAJO */}
          <main class="flex-1 overflow-y-auto bg-colpsi-bg">
            <div class="mx-auto w-full max-w-[1400px] px-4 sm:px-6 py-5 space-y-5">
              {/* LAS PÁGINAS SE CARGAN AQUÍ */}
              {props.children}
            </div>
          </main>
        </div>

      </div>
    </Show>
  );
}