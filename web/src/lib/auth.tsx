// web/src/lib/auth.tsx

import { createContext, useContext, createSignal, JSX, onMount, onCleanup } from "solid-js";
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

// ── Helpers JWT ───────────────────────────────────────────────────────────────
function getTokenExpiry(token: string): number | null {
  try {
    const payload = JSON.parse(atob(token.split(".")[1]));
    return typeof payload.exp === "number" ? payload.exp : null;
  } catch {
    return null;
  }
}

function isTokenExpired(token: string): boolean {
  const exp = getTokenExpiry(token);
  if (exp === null) return true;
  return Date.now() / 1000 > exp;
}

function msUntilExpiry(token: string): number {
  const exp = getTokenExpiry(token);
  if (exp === null) return -1;
  return exp * 1000 - Date.now();
}
// ─────────────────────────────────────────────────────────────────────────────

export function AuthProvider(props: { children: JSX.Element }) {
  const [user, setUser] = createSignal<AuthUser | null>(null);

  const clearLocalSession = () => {
    sessionStorage.removeItem("jwt");
    Cookies.remove("user_data");
    setUser(null);
  };

  const logout = async () => {
    const token = sessionStorage.getItem("jwt");
    const currentRole = user()?.role;

    if (token && currentRole === "psi") {
      try {
        await fetch(`${import.meta.env.VITE_API_URL}/psi/me/logout`, {
          method: "POST",
          headers: { Authorization: `Bearer ${token}` },
        });
      } catch {
        // Silencioso — limpiamos localmente de todas formas
      }
    }

    clearLocalSession();
    if (!isServer) window.location.href = "/";
  };

  // Cierre silencioso por expiración — sin llamar al backend
  const expireSession = () => {
    clearLocalSession();
    if (!isServer) window.location.href = "/";
  };

  onMount(() => {
    const savedUser = Cookies.get("user_data");
    const token     = sessionStorage.getItem("jwt");

    if (!savedUser || !token) {
      clearLocalSession();
      return;
    }

    if (isTokenExpired(token)) {
      expireSession();
      return;
    }

    try {
      setUser(JSON.parse(savedUser));
    } catch {
      clearLocalSession();
      return;
    }

    // Timer preciso: se dispara exactamente cuando vence el JWT
    const ms = msUntilExpiry(token);
    if (ms > 0) {
      const timer = setTimeout(expireSession, ms);
      onCleanup(() => clearTimeout(timer));
    }
  });

  const login = (token: string, userData: AuthUser) => {
    sessionStorage.setItem("jwt", token);
    Cookies.set("user_data", JSON.stringify(userData), { expires: 1 });
    setUser(userData);

    // Arrancar timer también en login
    const ms = msUntilExpiry(token);
    if (ms > 0) {
      const timer = setTimeout(expireSession, ms);
      onCleanup(() => clearTimeout(timer));
    }
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