// web/src/lib/auth.tsx

import { createContext, useContext, createSignal, JSX, onMount } from "solid-js";
import Cookies from "js-cookie";
import { isServer } from "solid-js/web";
import { AuthUser, UserRole } from "~/types/auth";

interface AuthContextValue {
  user: () => AuthUser | null;
  isAuthenticated: () => boolean;
  role: () => UserRole | null;
  login: (token: string, user: AuthUser) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue>();

export function AuthProvider(props: { children: JSX.Element }) {
  const [user, setUser] = createSignal<AuthUser | null>(null);

  onMount(() => {
    const savedUser = Cookies.get("user_data");
    if (savedUser) {
      try {
        setUser(JSON.parse(savedUser));
      } catch (e) {
        clearLocalSession();
      }
    }
  });

  const login = (token: string, userData: AuthUser) => {
    Cookies.set("jwt", token, { expires: 1, secure: true, sameSite: "strict" });
    Cookies.set("user_data", JSON.stringify(userData), { expires: 1 });
    setUser(userData);
  };

  // Limpia el estado local sin redirigir — usado internamente
  const clearLocalSession = () => {
    Cookies.remove("jwt");
    Cookies.remove("user_data");
    setUser(null);
  };

  const logout = async () => {
    const token = Cookies.get("jwt");
    const currentRole = user()?.role;

    // ── Notificar al backend ANTES de borrar el token ────────────────────────
    // Solo psicólogos tienen endpoint de logout (admins no tienen sesión activa trackeada)
    // Fire-and-forget: si falla la red, igual limpiamos localmente
    if (token && currentRole === "psi") {
      console.log(`${import.meta.env.VITE_API_URL}/psi/me/logout`)
      try {
        await fetch(`${import.meta.env.VITE_API_URL}/psi/me/logout`, {
          method: "POST",
          headers: { Authorization: `Bearer ${token}` },
        });
      } catch {
        // Silencioso — la rotación de key es best-effort desde el cliente
        // El token expirará solo a las 24h de todas formas
      }
    }
    // ────────────────────────────────────────────────────────────────────────

    clearLocalSession();

    if (!isServer) window.location.href = "/";
  };

  const value: AuthContextValue = {
    user,
    isAuthenticated: () => !!user(),
    role: () => user()?.role || null,
    login,
    logout,
  };

  return (
    <AuthContext.Provider value={value}>
      {props.children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth debe ser usado dentro de un AuthProvider");
  return context;
}