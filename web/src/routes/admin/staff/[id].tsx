// routes/admin/staff/[id].tsx
import { createResource, createSignal, Show } from "solid-js";
import { useNavigate, useParams } from "@solidjs/router";
import { apiGet, apiPatch } from "~/lib/api";
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
import { Badge } from "~/components/admin/ui/Badge";

const IC = "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text";
const labelClass = "block text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1 mb-1";

interface Admin {
  id: string;
  username: string;
  email: string;
  is_active: boolean;
  created_at: string;
  create_by: string;
  role?: string | null;
  can_read_psi: boolean;
  can_create_psi: boolean;
  can_update_psi: boolean;
  can_delete_psi: boolean;
  can_create_admin: boolean;
  can_update_admin: boolean;
  can_delete_admin: boolean;
  can_publish: boolean;
  can_update_publish: boolean;
  can_delete_publish: boolean;
  can_send_notifications: boolean;
  can_manage_notifications: boolean;
  can_read_notifications: boolean;
  can_create_tags: boolean;
  can_edit_tags: boolean;
  can_delete_tags: boolean;
  can_manage_projects: boolean;
  can_manage_tickets: boolean;
  can_view_logs?: boolean;
  can_export_logs?: boolean;
}

export default function AdminEditarStaffPage() {
  const params = useParams<{ id: string }>();
  const navigate = useNavigate();

  const [admin] = createResource(
    () => params.id,
    async (id) => {
      try {
        // El endpoint GetAdmins filtra por search, usamos list y buscamos por id
        const res = await apiGet<{ data: Admin[] }>(`/admin/list?limit=100`);
        return res.data?.find((a) => a.id === id) ?? null;
      } catch (err: any) {
        console.error("[edit staff] error:", err?.status, err?.message);
        return null;
      }
    }
  );

  const [username, setUsername] = createSignal("");
  const [email, setEmail] = createSignal("");
  const [password, setPassword] = createSignal("");
  const [showPassword, setShowPassword] = createSignal(false);
  const [isActive, setIsActive] = createSignal(true);
  const [perms, setPerms] = createSignal<PermissionState>(defaultPerms());
  const [role, setRole] = createSignal<string | null>(null);
  const [initialized, setInitialized] = createSignal(false);

  const [saving, setSaving] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);
  const [success, setSuccess] = createSignal(false);

  const initForm = (a: Admin) => {
    if (initialized()) return;
    setUsername(a.username ?? "");
    setEmail(a.email ?? "");
    setIsActive(a.is_active ?? true);
    setPerms({
      can_read_psi: a.can_read_psi, can_create_psi: a.can_create_psi, can_update_psi: a.can_update_psi, can_delete_psi: a.can_delete_psi,
      can_create_admin: a.can_create_admin, can_update_admin: a.can_update_admin, can_delete_admin: a.can_delete_admin,
      can_publish: a.can_publish, can_update_publish: a.can_update_publish, can_delete_publish: a.can_delete_publish,
      can_send_notifications: a.can_send_notifications, can_manage_notifications: a.can_manage_notifications, can_read_notifications: a.can_read_notifications,
      can_create_tags: a.can_create_tags, can_edit_tags: a.can_edit_tags, can_delete_tags: a.can_delete_tags,
      can_manage_projects: a.can_manage_projects, can_manage_tickets: a.can_manage_tickets,
      can_view_logs: a.can_view_logs ?? false, can_export_logs: a.can_export_logs ?? false,
    });
    setRole(a.role ?? null);
    setInitialized(true);
  };

  const togglePerm = (key: keyof PermissionState) => {
    setPerms((p) => ({ ...p, [key]: !p[key] }));
    setRole("personalizado");
  };

  const toggleGroup = (keys: readonly string[]) => {
    const all = keys.every((k) => perms()[k as keyof PermissionState]);
    setPerms((p) => {
      const next = { ...p };
      keys.forEach((k) => { (next as any)[k] = !all; });
      return next;
    });
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
    setSaving(true);
    setError(null);
    setSuccess(false);

    try {
      const body: any = {
        id: params.id,
        username: username().trim() || undefined,
        email: email().trim() || undefined,
        is_active: isActive(),
        role: role() ?? "personalizado",
        permissions: perms(),
      };
      if (password().trim()) body.password = password();

      await apiPatch("/admin/update", body);

      setSuccess(true);
      window.scrollTo({ top: 0, behavior: "smooth" });
      setTimeout(() => navigate("/admin/staff"), 1200);
    } catch (err: any) {
      setError(getUserFacingError(err));
      window.scrollTo({ top: 0, behavior: "smooth" });
    } finally {
      setSaving(false);
    }
  };

  return (
    <main class="space-y-4 pb-12 max-w-3xl mx-auto">

      {/* ── HEADER ────────────────────────────────────────────────────────── */}
      <div class="flex items-center gap-3 pb-4 border-b border-colpsi-border">
        <button
          onClick={() => navigate(-1)}
          class="inline-flex items-center justify-center h-8 w-8 rounded-md border border-colpsi-border bg-white text-colpsi-muted hover:text-colpsi-blue hover:bg-colpsi-bg transition-colors"
          title="Volver"
        >
          <Icon name="chevronRight" class="w-4 h-4 rotate-180" />
        </button>
        <div class="flex-1 min-w-0">
          <h1 class="text-lg font-semibold text-colpsi-text">Editar Administrador</h1>
          <p class="text-sm text-colpsi-muted mt-0.5 truncate">
            {admin.loading ? "Cargando..." : admin()?.username ?? ""}
          </p>
        </div>
        <Show when={admin()}>
          {(a) => (
            <Badge tone={a().is_active ? "success" : "neutral"} class="shrink-0">
              {a().is_active ? "Activo" : "Inactivo"}
            </Badge>
          )}
        </Show>
      </div>

      {/* ── FEEDBACK ──────────────────────────────────────────────────────── */}
      <Show when={error()}>
        <div class="p-3 rounded-md bg-red-50 text-red-700 border border-red-200 text-sm font-medium">{error()}</div>
      </Show>
      <Show when={success()}>
        <div class="p-3 rounded-md bg-emerald-50 text-emerald-700 border border-emerald-200 text-sm font-medium">
          Administrador actualizado. Redirigiendo...
        </div>
      </Show>

      {/* ── SKELETON ──────────────────────────────────────────────────────── */}
      <Show when={admin.loading}>
        <div class="space-y-4 animate-pulse">
          <div class="bg-white rounded-lg h-48 border border-colpsi-border" />
          <div class="bg-white rounded-lg h-96 border border-colpsi-border" />
        </div>
      </Show>

      {/* ── NO ENCONTRADO ─────────────────────────────────────────────────── */}
      <Show when={!admin.loading && admin() === null}>
        <div class="text-center py-16 bg-white rounded-lg border border-colpsi-border">
          <span class="inline-flex h-12 w-12 items-center justify-center rounded-md bg-colpsi-bg text-colpsi-muted mb-4">
            <Icon name="user" class="w-6 h-6" />
          </span>
          <h2 class="text-base font-semibold text-colpsi-text mb-1">Administrador no encontrado</h2>
          <button onClick={() => navigate("/admin/staff")} class="mt-4 inline-flex items-center gap-1.5 text-colpsi-blue font-semibold text-sm hover:underline">
            <Icon name="chevronRight" class="w-3.5 h-3.5 rotate-180" />
            Volver al listado
          </button>
        </div>
      </Show>

      {/* ── FORMULARIO ────────────────────────────────────────────────────── */}
      <Show when={admin()}>
        {(a) => {
          initForm(a());
          return (
            <form onSubmit={handleSubmit} class="space-y-4">

              {/* ══ DATOS DE ACCESO ════════════════════════════════════════ */}
              <section class="bg-white rounded-lg p-5 border border-colpsi-border space-y-4">
                <h2 class="text-base font-semibold text-colpsi-text border-b border-colpsi-border pb-3">
                  Datos de Acceso
                </h2>

                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label class={labelClass}>Usuario</label>
                    <input
                      type="text" maxLength={25}
                      value={username()}
                      onInput={(e) => setUsername(e.currentTarget.value)}
                      class={IC}
                    />
                    <p class="text-xs text-colpsi-muted mt-1 text-right">{username().length}/25</p>
                  </div>
                  <div>
                    <label class={labelClass}>Email</label>
                    <input
                      type="email"
                      value={email()}
                      onInput={(e) => setEmail(e.currentTarget.value)}
                      class={IC}
                    />
                  </div>
                </div>

                <div>
                  <label class={labelClass}>Nueva Contraseña <span class="text-colpsi-muted font-medium normal-case">(dejar vacío para no cambiar)</span></label>
                  <div class="relative">
                    <input
                      type={showPassword() ? "text" : "password"}
                      placeholder="Nueva contraseña..."
                      value={password()}
                      onInput={(e) => setPassword(e.currentTarget.value)}
                      class={`${IC} pr-14`}
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword((v) => !v)}
                      class="absolute right-2 top-1/2 -translate-y-1/2 text-xs font-semibold text-colpsi-muted hover:text-colpsi-blue px-2 py-1"
                    >
                      {showPassword() ? "Ocultar" : "Ver"}
                    </button>
                  </div>
                </div>

                {/* Estado */}
                <div>
                  <label class={labelClass}>Estado de la Cuenta</label>
                  <div class="flex gap-2 mt-1">
                    {([true, false] as const).map((val) => (
                      <button
                        type="button"
                        onClick={() => setIsActive(val)}
                        class={`flex-1 h-10 rounded-md text-xs font-semibold uppercase tracking-wide border transition-all ${
                          isActive() === val
                            ? val ? "bg-emerald-600 text-white border-emerald-600" : "bg-slate-600 text-white border-slate-600"
                            : "bg-white text-colpsi-muted border-colpsi-border hover:border-slate-300"
                        }`}
                      >
                        {val ? "Activo" : "Inactivo"}
                      </button>
                    ))}
                  </div>
                </div>
              </section>

              {/* ══ PERFIL DE ROL ══════════════════════════════════════ */}
              <section class="bg-white rounded-lg p-5 border border-colpsi-border">
                <div class="flex items-center justify-between border-b border-colpsi-border pb-3 mb-4">
                  <h2 class="text-base font-semibold text-colpsi-text">Perfil de Rol</h2>
                  <span class="text-xs font-medium text-colpsi-muted">Atajo: aplica un conjunto de permisos</span>
                </div>
                <RoleSelector perms={perms()} storedRole={role()} onSelect={applyRole} onClear={clearRole} />
              </section>

              {/* ══ PERMISOS ══════════════════════════════════════════════ */}
              <section class="bg-white rounded-lg p-5 border border-colpsi-border">
                <div class="flex items-center justify-between border-b border-colpsi-border pb-3 mb-4">
                  <h2 class="text-base font-semibold text-colpsi-text">Permisos</h2>
                  <span class="text-xs font-medium text-colpsi-muted">{totalEnabled()}/{TOTAL_PERMS} activos</span>
                </div>

                <div class="space-y-3">
                  {PERM_GROUPS.map((group) => {
                    const allOn = () => group.perms.every((p) => perms()[p.key as keyof PermissionState]);
                    return (
                      <div class="rounded-md border border-colpsi-border overflow-hidden">
                        <button
                          type="button"
                          onClick={() => toggleGroup(group.perms.map((p) => p.key))}
                          class={`w-full flex items-center justify-between px-4 py-2.5 text-left transition-colors ${
                            allOn() ? ACTIVE_MAP[group.color] : `${COLOR_MAP[group.color]} hover:opacity-90`
                          }`}
                        >
                          <span class="font-semibold text-sm">{group.label}</span>
                          <span class="text-[11px] font-medium opacity-80">
                            {allOn() ? "Quitar todos" : "Dar todos"}
                          </span>
                        </button>
                        <div class="grid grid-cols-3 gap-px bg-colpsi-border">
                          {group.perms.map((perm) => {
                            const active = () => perms()[perm.key as keyof PermissionState];
                            return (
                              <button
                                type="button"
                                onClick={() => togglePerm(perm.key as keyof PermissionState)}
                                class={`flex items-center justify-between px-4 py-2.5 text-sm font-medium transition-all ${
                                  active()
                                    ? `${ACTIVE_MAP[group.color]} opacity-90`
                                    : "bg-white text-colpsi-muted hover:bg-colpsi-bg"
                                }`}
                              >
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

              {/* ── BOTONES ─────────────────────────────────────────────── */}
              <div class="sticky bottom-4 z-50 flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => navigate(-1)}
                  class="h-10 px-4 rounded-md border border-colpsi-border bg-white text-sm font-medium text-colpsi-text hover:bg-colpsi-bg transition-colors"
                >
                  Cancelar
                </button>
                <button
                  type="submit"
                  disabled={saving()}
                  class="inline-flex items-center justify-center gap-2 h-10 px-5 rounded-md bg-colpsi-blue text-white text-sm font-semibold transition-colors hover:bg-colpsi-blue-light disabled:opacity-70"
                >
                  <Show when={saving()} fallback={<Icon name="check" />}>
                    <span class="animate-spin h-4 w-4 border-2 border-white/40 border-t-white rounded-full" />
                  </Show>
                  {saving() ? "Guardando..." : "Guardar Cambios"}
                </button>
              </div>

            </form>
          );
        }}
      </Show>

    </main>
  );
}