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
  // Estado inicial: Intentamos recuperar de la cookie si estamos en el cliente
  const [user, setUser] = createSignal<AuthUser | null>(null);

  // Inicialización (Solo corre en el navegador al cargar)
  onMount(() => {
    const savedUser = Cookies.get("user_data");
    if (savedUser) {
      try {
        setUser(JSON.parse(savedUser));
      } catch (e) {
        logout();
      }
    }
  });

  const login = (token: string, userData: AuthUser) => {
    // 1. Guardar JWT para la API (Seguridad)
    Cookies.set("jwt", token, { expires: 1, secure: true, sameSite: 'strict' });
    
    // 2. Guardar perfil básico para la UI (No sensible)
    Cookies.set("user_data", JSON.stringify(userData), { expires: 1 });
    
    // 3. Actualizar estado reactivo
    setUser(userData);
  };

  const logout = () => {
    Cookies.remove("jwt");
    Cookies.remove("user_data");
    setUser(null);
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

// Hook personalizado para usar la autenticación en cualquier parte
export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth debe ser usado dentro de un AuthProvider");
  return context;
}