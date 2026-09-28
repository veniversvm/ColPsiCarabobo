// web/src/lib/actions/session.ts
// Server action de recuperación de sesión (sin dependencias de `vinxi/http`:
// este módulo se importa desde auth.tsx, que corre en el navegador).
//
// El navegador pierde `sessionStorage.jwt` en pestañas nuevas / reinicios, pero
// la cookie HttpOnly `jwt` sigue viva → el SSR sigue autenticando mientras las
// llamadas fetch del navegador quedan sin token. Esta action devuelve el MISMO
// token de la cookie para que el cliente reestablezca su copia per-tab SIN
// re-login (un re-login rota la key del admin — gotcha 13 de api/AGENTS.md — y
// en ≤60s checkSession mataba la sesión de las demás pestañas). No agrega
// exposición: el login ya deposita una copia del token en sessionStorage por
// diseño (ver `src/lib/auth.tsx`).
import { action } from "@solidjs/router";
import { getRequestEvent } from "solid-js/web";

export const restoreSessionAction = action(async () => {
  "use server";

  // En SSR la petición entrante trae la cookie HttpOnly `jwt` (mismo patrón de
  // lectura que `src/lib/api.ts`).
  const event = getRequestEvent();
  const cookieHeader = event?.request?.headers?.get("cookie") || "";
  const match = cookieHeader.match(/(^| )jwt=([^;]+)/);
  return match ? match[2] : "";
});