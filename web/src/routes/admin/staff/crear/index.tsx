// routes/admin/staff/crear/index.tsx
import { createSignal } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";

const IC = "w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none transition-all text-gray-800 text-sm";
const labelClass = "block text-[10px] font-black text-gray-500 uppercase tracking-widest ml-1 mb-1";

interface PermissionsState {
  can_create_psi: boolean; can_update_psi: boolean; can_delete_psi: boolean;
  can_create_admin: boolean; can_update_admin: boolean; can_delete_admin: boolean;
  can_publish: boolean; can_update_publish: boolean; can_delete_publish: boolean;
  can_send_notifications: boolean; can_manage_notifications: boolean; can_read_notifications: boolean;
  can_create_tags: boolean; can_edit_tags: boolean; can_delete_tags: boolean;
  can_manage_projects: boolean; can_manage_tickets: boolean;
}

const defaultPerms = (): PermissionsState => ({
  can_create_psi: false, can_update_psi: false, can_delete_psi: false,
  can_create_admin: false, can_update_admin: false, can_delete_admin: false,
  can_publish: false, can_update_publish: false, can_delete_publish: false,
  can_send_notifications: false, can_manage_notifications: false, can_read_notifications: false,
  can_create_tags: false, can_edit_tags: false, can_delete_tags: false,
  can_manage_projects: false, can_manage_tickets: false,
});

const PERM_GROUPS = [
  { label: "Gestión de Colegiados", icon: "👤", color: "blue",
    perms: [{ key: "can_create_psi", label: "Crear" }, { key: "can_update_psi", label: "Editar" }, { key: "can_delete_psi", label: "Eliminar" }] },
  { label: "Gestión de Staff", icon: "🛡️", color: "purple",
    perms: [{ key: "can_create_admin", label: "Crear" }, { key: "can_update_admin", label: "Editar" }, { key: "can_delete_admin", label: "Eliminar" }] },
  { label: "Publicaciones", icon: "📰", color: "emerald",
    perms: [{ key: "can_publish", label: "Publicar" }, { key: "can_update_publish", label: "Editar" }, { key: "can_delete_publish", label: "Eliminar" }] },
  { label: "Notificaciones", icon: "🔔", color: "amber",
    perms: [{ key: "can_send_notifications", label: "Enviar" }, { key: "can_manage_notifications", label: "Gestionar" }, { key: "can_read_notifications", label: "Leer" }] },
  { label: "Especialidades / Tags", icon: "🏷️", color: "rose",
    perms: [{ key: "can_create_tags", label: "Crear" }, { key: "can_edit_tags", label: "Editar" }, { key: "can_delete_tags", label: "Eliminar" }] },
  { label: "Proyectos", icon: "📋", color: "indigo",
    perms: [{ key: "can_manage_projects", label: "Gestionar" }] },
  { label: "Tickets", icon: "🎫", color: "cyan",
    perms: [{ key: "can_manage_tickets", label: "Gestionar" }] },
] as const;

const COLOR_MAP: Record<string, string> = {
  blue: "bg-blue-50 border-blue-200 text-blue-700", purple: "bg-purple-50 border-purple-200 text-purple-700",
  emerald: "bg-emerald-50 border-emerald-200 text-emerald-700", amber: "bg-amber-50 border-amber-200 text-amber-700",
  rose: "bg-rose-50 border-rose-200 text-rose-700", indigo: "bg-indigo-50 border-indigo-200 text-indigo-700",
  cyan: "bg-cyan-50 border-cyan-200 text-cyan-700",
};
const ACTIVE_MAP: Record<string, string> = {
  blue: "bg-blue-700 border-blue-700 text-white", purple: "bg-purple-700 border-purple-700 text-white",
  emerald: "bg-emerald-700 border-emerald-700 text-white", amber: "bg-amber-700 border-amber-700 text-white",
  rose: "bg-rose-700 border-rose-700 text-white", indigo: "bg-indigo-700 border-indigo-700 text-white",
  cyan: "bg-cyan-700 border-cyan-700 text-white",
};

export default function AdminCrearStaffPage() {
  const navigate = useNavigate();

  // Una key por montaje — se regenera si el admin navega fuera y vuelve
  const idempotencyKey = crypto.randomUUID();

  const [username, setUsername] = createSignal("");
  const [email, setEmail] = createSignal("");
  const [password, setPassword] = createSignal("");
  const [showPassword, setShowPassword] = createSignal(false);
  const [perms, setPerms] = createSignal<PermissionsState>(defaultPerms());
  const [saving, setSaving] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);

  const togglePerm = (key: keyof PermissionsState) =>
    setPerms((p) => ({ ...p, [key]: !p[key] }));

  const toggleGroup = (keys: readonly string[]) => {
    const all = keys.every((k) => perms()[k as keyof PermissionsState]);
    setPerms((p) => { const next = { ...p }; keys.forEach((k) => { (next as any)[k] = !all; }); return next; });
  };

  const totalEnabled = () => Object.values(perms()).filter(Boolean).length;

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    if (!username().trim()) { setError("El usuario es obligatorio."); return; }
    if (!email().trim())    { setError("El email es obligatorio."); return; }
    if (!password().trim()) { setError("La contraseña es obligatoria."); return; }

    setSaving(true);
    setError(null);

    try {
      await apiPost("/admin/create", {
        username: username().trim(),
        email:    email().trim(),
        password: password(),
        permissions: perms(),
      }, {
        headers: { "X-Idempotency-Key": idempotencyKey },
      });
      navigate("/admin/staff");
    } catch (err: any) {
      setError(getUserFacingError(err));
      window.scrollTo({ top: 0, behavior: "smooth" });
    } finally {
      setSaving(false);
    }
  };

  return (
    <main class="pb-28 animate-in fade-in duration-500 max-w-3xl mx-auto">

      <div class="flex items-center gap-4 mb-8 bg-white p-6 rounded-3xl shadow-sm border border-gray-100">
        <button onClick={() => navigate(-1)} class="w-10 h-10 bg-gray-50 hover:bg-gray-100 text-gray-600 rounded-full font-bold flex items-center justify-center transition-colors flex-shrink-0">←</button>
        <div>
          <h1 class="text-2xl font-black text-blue-900 uppercase tracking-tight">Nuevo Administrador</h1>
          <p class="text-gray-400 text-sm mt-0.5 font-medium">Crea un nuevo miembro del staff con permisos específicos</p>
        </div>
      </div>

      {error() && (
        <div class="mb-6 p-4 rounded-2xl bg-red-50 text-red-800 font-bold text-sm border-l-4 border-red-500 shadow-sm">{error()}</div>
      )}

      <form onSubmit={handleSubmit} class="space-y-6">

        <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100 space-y-5">
          <h2 class="text-sm font-black text-blue-800 uppercase tracking-widest border-b border-gray-100 pb-3">Datos de Acceso</h2>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
            <div>
              <label class={labelClass}>Usuario <span class="text-red-400">*</span></label>
              <input type="text" required maxLength={25} placeholder="ej. jperez" value={username()} onInput={(e) => setUsername(e.currentTarget.value)} class={IC} />
              <p class="text-[10px] text-gray-400 mt-1 text-right">{username().length}/25</p>
            </div>
            <div>
              <label class={labelClass}>Email <span class="text-red-400">*</span></label>
              <input type="email" required placeholder="ej. jperez@colpsi.org" value={email()} onInput={(e) => setEmail(e.currentTarget.value)} class={IC} />
            </div>
          </div>
          <div>
            <label class={labelClass}>Contraseña <span class="text-red-400">*</span></label>
            <div class="relative">
              <input type={showPassword() ? "text" : "password"} required placeholder="Mínimo 8 caracteres, mayúsculas y números" value={password()} onInput={(e) => setPassword(e.currentTarget.value)} class={`${IC} pr-12`} />
              <button type="button" onClick={() => setShowPassword((v) => !v)} class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 text-xs font-bold px-1">
                {showPassword() ? "Ocultar" : "Ver"}
              </button>
            </div>
            <p class="text-[10px] text-gray-400 mt-1 ml-1">El sistema enviará las credenciales al email del administrador.</p>
          </div>
        </section>

        <section class="bg-white rounded-3xl p-6 md:p-8 shadow-sm border border-gray-100">
          <div class="flex items-center justify-between border-b border-gray-100 pb-3 mb-6">
            <h2 class="text-sm font-black text-blue-800 uppercase tracking-widest">Permisos</h2>
            <span class="text-xs font-black text-gray-500">{totalEnabled()}/17 activos</span>
          </div>
          <div class="space-y-4">
            {PERM_GROUPS.map((group) => {
              const allOn = () => group.perms.every((p) => perms()[p.key as keyof PermissionsState]);
              return (
                <div class={`rounded-2xl border-2 overflow-hidden ${COLOR_MAP[group.color].split(" ").slice(0, 2).join(" ")}`}>
                  <button type="button" onClick={() => toggleGroup(group.perms.map((p) => p.key))}
                    class={`w-full flex items-center justify-between px-4 py-3 text-left transition-colors ${allOn() ? ACTIVE_MAP[group.color] : `${COLOR_MAP[group.color]} hover:opacity-90`}`}>
                    <span class="font-black text-sm flex items-center gap-2"><span>{group.icon}</span>{group.label}</span>
                    <span class="text-[11px] font-black opacity-80">{allOn() ? "Quitar todos" : "Dar todos"}</span>
                  </button>
                  <div class="grid grid-cols-3 gap-px bg-gray-100">
                    {group.perms.map((perm) => {
                      const active = () => perms()[perm.key as keyof PermissionsState];
                      return (
                        <button type="button" onClick={() => togglePerm(perm.key as keyof PermissionsState)}
                          class={`flex items-center justify-between px-4 py-3 text-sm font-bold transition-all ${active() ? `${ACTIVE_MAP[group.color]} opacity-90` : "bg-white text-gray-400 hover:bg-gray-50"}`}>
                          <span>{perm.label}</span>
                          <span class="text-base">{active() ? "✓" : "○"}</span>
                        </button>
                      );
                    })}
                  </div>
                </div>
              );
            })}
          </div>
        </section>

        <div class="sticky bottom-6 z-50 flex justify-end gap-3">
          <button type="button" onClick={() => navigate(-1)} class="bg-white text-gray-600 border-2 border-gray-200 px-6 py-4 rounded-2xl font-black hover:bg-gray-50 transition-all text-sm">Cancelar</button>
          <button type="submit" disabled={saving()} class="bg-blue-800 text-white px-10 py-4 rounded-2xl font-black shadow-2xl hover:scale-105 active:scale-95 transition-all disabled:opacity-70 flex items-center gap-3 border-2 border-white text-sm">
            {saving() ? "CREANDO..." : "👤 CREAR ADMINISTRADOR"}
          </button>
        </div>

      </form>
    </main>
  );
}