// routes/admin/staff/crear/index.tsx
import { createSignal } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { apiPost } from "~/lib/api";
import { getUserFacingError } from "~/lib/errors";
import RoleSelector from "~/components/admin/staff/RoleSelector";
import {
  COLOR_MAP,
  ACTIVE_MAP,
  PERM_GROUPS,
  defaultPerms,
  TOTAL_PERMS,
  applyPresetToPerms,
  type PermissionState,
  type RolePreset,
} from "~/lib/staff-permissions";
import { Icon } from "~/components/admin/ui/icons";

const IC = "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text";
const labelClass = "block text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1 mb-1";

export default function AdminCrearStaffPage() {
  const navigate = useNavigate();

  // Una key por montaje — se regenera si el admin navega fuera y vuelve
  const idempotencyKey = crypto.randomUUID();

  const [username, setUsername] = createSignal("");
  const [email, setEmail] = createSignal("");
  const [password, setPassword] = createSignal("");
  const [showPassword, setShowPassword] = createSignal(false);
  const [perms, setPerms] = createSignal<PermissionState>(defaultPerms());
  const [role, setRole] = createSignal<string | null>(null);
  const [saving, setSaving] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);

  const togglePerm = (key: keyof PermissionState) => {
    setPerms((p) => ({ ...p, [key]: !p[key] }));
    setRole("personalizado");
  };

  const toggleGroup = (keys: readonly string[]) => {
    const all = keys.every((k) => perms()[k as keyof PermissionState]);
    setPerms((p) => { const next = { ...p }; keys.forEach((k) => { (next as any)[k] = !all; }); return next; });
    setRole("personalizado");
  };

  // Aplica un preset: rellena todos los permisos y registra la etiqueta del rol.
  const applyRole = (preset: RolePreset) => {
    setPerms((p) => applyPresetToPerms(p, preset));
    setRole(preset.slug);
  };

  const clearRole = () => {
    setRole("personalizado");
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
        role:     role() ?? "personalizado",
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
    <main class="space-y-4 pb-12 max-w-3xl mx-auto">

      <div class="flex items-center gap-3 pb-4 border-b border-colpsi-border">
        <button onClick={() => navigate(-1)} class="inline-flex items-center justify-center h-8 w-8 rounded-md border border-colpsi-border bg-white text-colpsi-muted hover:text-colpsi-blue hover:bg-colpsi-bg transition-colors" title="Volver">
          <Icon name="chevronRight" class="w-4 h-4 rotate-180" />
        </button>
        <div>
          <h1 class="text-lg font-semibold text-colpsi-text">Nuevo Administrador</h1>
          <p class="text-sm text-colpsi-muted mt-0.5">Crea un nuevo miembro del staff con permisos específicos</p>
        </div>
      </div>

      {error() && (
        <div class="p-3 rounded-md bg-red-50 text-red-700 border border-red-200 text-sm font-medium">{error()}</div>
      )}

      <form onSubmit={handleSubmit} class="space-y-4">

        <section class="bg-white rounded-lg p-5 border border-colpsi-border space-y-4">
          <h2 class="text-base font-semibold text-colpsi-text border-b border-colpsi-border pb-3">Datos de Acceso</h2>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class={labelClass}>Usuario <span class="text-red-400">*</span></label>
              <input type="text" required maxLength={25} placeholder="ej. jperez" value={username()} onInput={(e) => setUsername(e.currentTarget.value)} class={IC} />
              <p class="text-xs text-colpsi-muted mt-1 text-right">{username().length}/25</p>
            </div>
            <div>
              <label class={labelClass}>Email <span class="text-red-400">*</span></label>
              <input type="email" required placeholder="ej. jperez@colpsi.org" value={email()} onInput={(e) => setEmail(e.currentTarget.value)} class={IC} />
            </div>
          </div>
          <div>
            <label class={labelClass}>Contraseña <span class="text-red-400">*</span></label>
            <div class="relative">
              <input type={showPassword() ? "text" : "password"} required placeholder="Mínimo 8 caracteres, mayúsculas y números" value={password()} onInput={(e) => setPassword(e.currentTarget.value)} class={`${IC} pr-14`} />
              <button type="button" onClick={() => setShowPassword((v) => !v)} class="absolute right-2 top-1/2 -translate-y-1/2 text-xs font-semibold text-colpsi-muted hover:text-colpsi-blue px-2 py-1">
                {showPassword() ? "Ocultar" : "Ver"}
              </button>
            </div>
            <p class="text-[11px] text-colpsi-muted mt-1 ml-1">El sistema enviará las credenciales al email del administrador.</p>
          </div>
        </section>

        <section class="bg-white rounded-lg p-5 border border-colpsi-border">
          <div class="flex items-center justify-between border-b border-colpsi-border pb-3 mb-4">
            <h2 class="text-base font-semibold text-colpsi-text">Perfil de Rol</h2>
            <span class="text-xs font-medium text-colpsi-muted">Atajo: aplica un conjunto de permisos</span>
          </div>
          <RoleSelector perms={perms()} storedRole={role()} onSelect={applyRole} onClear={clearRole} />
        </section>

        <section class="bg-white rounded-lg p-5 border border-colpsi-border">
          <div class="flex items-center justify-between border-b border-colpsi-border pb-3 mb-4">
            <h2 class="text-base font-semibold text-colpsi-text">Permisos</h2>
            <span class="text-xs font-medium text-colpsi-muted">{totalEnabled()}/{TOTAL_PERMS} activos</span>
          </div>
          <div class="space-y-3">
            {PERM_GROUPS.map((group) => {
              const allOn = () => group.perms.every((p) => perms()[p.key as keyof PermissionState]);
              return (
                <div class={`rounded-md border border-colpsi-border overflow-hidden`}>
                  <button type="button" onClick={() => toggleGroup(group.perms.map((p) => p.key))}
                    class={`w-full flex items-center justify-between px-4 py-2.5 text-left transition-colors ${allOn() ? ACTIVE_MAP[group.color] : `${COLOR_MAP[group.color]} hover:opacity-90`}`}>
                    <span class="font-semibold text-sm">{group.label}</span>
                    <span class="text-[11px] font-medium opacity-80">{allOn() ? "Quitar todos" : "Dar todos"}</span>
                  </button>
                  <div class="grid grid-cols-3 gap-px bg-colpsi-border">
                    {group.perms.map((perm) => {
                      const active = () => perms()[perm.key as keyof PermissionState];
                      return (
                        <button type="button" onClick={() => togglePerm(perm.key as keyof PermissionState)}
                          class={`flex items-center justify-between px-4 py-2.5 text-sm font-medium transition-colors ${active() ? `${ACTIVE_MAP[group.color]} opacity-90` : "bg-white text-colpsi-muted hover:bg-colpsi-bg"}`}>
                          <span>{perm.label}</span>
                          <span class="text-sm">{active() ? "✓" : "○"}</span>
                        </button>
                      );
                    })}
                  </div>
                </div>
              );
            })}
          </div>
        </section>

        <div class="sticky bottom-4 z-50 flex justify-end gap-2">
          <button type="button" onClick={() => navigate(-1)} class="h-10 px-4 rounded-md border border-colpsi-border bg-white text-sm font-medium text-colpsi-text hover:bg-colpsi-bg transition-colors">Cancelar</button>
          <button type="submit" disabled={saving()} class="inline-flex items-center justify-center gap-2 h-10 px-5 rounded-md bg-colpsi-blue text-white text-sm font-semibold transition-colors hover:bg-colpsi-blue-light disabled:opacity-60">
            <Icon name="user" />
            {saving() ? "Creando..." : "Crear Administrador"}
          </button>
        </div>

      </form>
    </main>
  );
}