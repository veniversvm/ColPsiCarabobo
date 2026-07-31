// web/src/lib/actions/auth.ts
// Server actions de autenticación (se ejecutan SOLO en el servidor).
//
// Modelo de sesión:
// - Go autentica y devuelve el JWT; el server action lo setea como cookie
//   HttpOnly (`jwt`) → el browser nunca lo ve.
// - El SSR / las server actions de admin leen esa cookie para llamar a la API
//   con `Authorization: Bearer`.
// - El cliente guarda una copia en `sessionStorage` para las peticiones fetch
//   del navegador (ver `src/lib/api.ts`).
// - `logoutAction` llama a `/psi/me/logout` de Go y borra las cookies.
"use server";

import { action, redirect } from "@solidjs/router";
import { setCookie, deleteCookie, getCookie } from "vinxi/http";

declare const process: { env?: Record<string, string | undefined> };
const API_BASE =
  process?.env?.API_URL_INTERNAL ||
  import.meta.env.VITE_API_URL ||
  "http://localhost:8080/api/v1";

// ── Constantes de cookie ──────────────────────────────────────────────────────
const COOKIE_JWT      = "jwt";
const COOKIE_USERDATA = "user_data";

const secureCookieBase = {
  httpOnly: true,
  secure:   import.meta.env.PROD === true,
  sameSite: "strict" as const,
  path:     "/",
  maxAge:   60 * 60 * 24, // 1 día
};

// ── Helpers ───────────────────────────────────────────────────────────────────
function decodeJwtPayload(token: string): Record<string, any> {
  try {
    return JSON.parse(atob(token.split(".")[1]));
  } catch {
    throw new Error("Token malformado");
  }
}

// ─────────────────────────────────────────────────────────────────────────────
// LOGIN PSI
// Llama a Go, setea JWT como HttpOnly desde el servidor, devuelve perfil al cliente.
// ─────────────────────────────────────────────────────────────────────────────
export const psiLoginAction = action(async (formData: FormData) => {
  "use server";

  const identifier = formData.get("identifier") as string;
  const password   = formData.get("password")   as string;

  if (!identifier || !password) {
    return { error: "Identificador y contraseña son requeridos." };
  }

  // 1. Autenticar contra Go
  let authData: { token: string };
  try {
    const res = await fetch(`${API_BASE}/psi/login`, {
      method:  "POST",
      headers: { "Content-Type": "application/json" },
      body:    JSON.stringify({ identifier, password }),
    });

    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      return { error: err.error || err.message || "Credenciales inválidas." };
    }

    authData = await res.json();
  } catch {
    return { error: "No se pudo conectar con el servidor." };
  }

  // 2. Obtener perfil completo usando el token recién obtenido
  let profile: any;
  try {
    const res = await fetch(`${API_BASE}/psi/me`, {
      headers: { Authorization: `Bearer ${authData.token}` },
    });

    if (!res.ok) return { error: "Error al recuperar el perfil." };
    profile = await res.json();
  } catch {
    return { error: "No se pudo recuperar el perfil." };
  }

  // 3. Setear JWT como HttpOnly — el browser nunca lo ve
  setCookie(COOKIE_JWT, authData.token, secureCookieBase);

  // 4. Devolver solo los datos públicos del usuario al cliente
  // (sin token — el cliente nunca necesita el JWT)
  return {
    user: {
      id:        profile.id,
      username:  profile.username,
      email:     profile.email,
      role:      "psi" as const,
      firstName: profile.first_name,
      lastName:  profile.last_name,
    },
  };
});

// ─────────────────────────────────────────────────────────────────────────────
// LOGIN ADMIN
// ─────────────────────────────────────────────────────────────────────────────
export const adminLoginAction = action(async (formData: FormData) => {
  "use server";

  const identifier = formData.get("identifier") as string;
  const password   = formData.get("password")   as string;

  if (!identifier || !password) {
    return { error: "Identificador y contraseña son requeridos." };
  }

  let authData: { token: string; message: string };
  try {
    const res = await fetch(`${API_BASE}/auth/login`, {
      method:  "POST",
      headers: { "Content-Type": "application/json" },
      body:    JSON.stringify({ identifier, password }),
    });

    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      return { error: err.error || err.message || "Credenciales inválidas." };
    }

    authData = await res.json();
  } catch {
    return { error: "No se pudo conectar con el servidor." };
  }

  // Decodificar payload para obtener user_id
  const payload = decodeJwtPayload(authData.token);

  // Setear JWT como HttpOnly
  setCookie(COOKIE_JWT, authData.token, secureCookieBase);

  return {
    user: {
      id:        payload.user_id,
      username:  identifier,
      email:     identifier,
      role:      "admin" as const,
      firstName: "Administrador",
    },
  };
});

// ─────────────────────────────────────────────────────────────────────────────
// LOGOUT
// ─────────────────────────────────────────────────────────────────────────────
export const logoutAction = action(async () => {
  "use server";

  // Leer el JWT para llamar al endpoint de logout de Go (solo psi lo tiene)
  const token = getCookie(COOKIE_JWT);

  if (token) {
    // Fire-and-forget — si falla Go igual limpiamos la cookie
    fetch(`${API_BASE}/psi/me/logout`, {
      method:  "POST",
      headers: { Authorization: `Bearer ${token}` },
    }).catch(() => {});
  }

  deleteCookie(COOKIE_JWT,      { path: "/" });
  deleteCookie(COOKIE_USERDATA, { path: "/" });

  return redirect("/");
});