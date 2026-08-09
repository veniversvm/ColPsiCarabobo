// web/src/lib/api.ts
//
// Cliente HTTP isomórfico hacia la API Go (Fiber).
// - SSR/Server Actions: resuelve la API por `API_URL_INTERNAL` (red Docker).
// - Navegador: usa `VITE_API_URL` (inlineado en build).
// - Token JWT isomórfico: el servidor lee la cookie HttpOnly `jwt`;
//   el cliente usa `sessionStorage.jwt`. Ambos se envían como `Authorization: Bearer`.
// - Convierte errores de red en `ApiError(503, "OFFLINE_SERVICE")` y respuestas
//   no-OK en `ApiError(status, mensaje, data)`.
import { isServer } from "solid-js/web";
import { getRequestEvent } from "solid-js/web";

declare const process: { env?: Record<string, string | undefined> };
const API_BASE_URL =
  process?.env?.API_URL_INTERNAL ||
  import.meta.env.VITE_API_URL ||
  "http://localhost:8080/api/v1";

export class ApiError extends Error {
  constructor(public status: number, message: string, public data: any = null) {
    super(message);
    this.name = "ApiError";
  }
}

export async function fetchApi<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;
  const headers = new Headers(options.headers || {});

  if (!headers.has("Content-Type") && !(options.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }

  // Token isomórfico: servidor lee la cookie HttpOnly, cliente usa sessionStorage
  let token = "";
  if (isServer) {
    const event = getRequestEvent();
    const cookieHeader = event?.request.headers.get("cookie") || "";
    const match = cookieHeader.match(/(^| )jwt=([^;]+)/);
    if (match) token = match[2];
  } else {
    token = sessionStorage.getItem("jwt") || "";
  }

  if (token) headers.set("Authorization", `Bearer ${token}`);

  try {
    const response = await fetch(url, { ...options, headers, credentials: options.credentials ?? "include" });

    if (!response.ok) {
      let errorData: any = null;
      try { errorData = await response.json(); } catch { /* no JSON */ }

      if (response.status === 401 && !isServer) {
        sessionStorage.removeItem("jwt");
      }

      const msg = errorData?.error || errorData?.message || "Error inesperado";
      throw new ApiError(response.status, msg, errorData);
    }

    return await response.json() as T;

  } catch (error: any) {
    if (error instanceof ApiError) throw error;

    const isNetworkError =
      error?.code === "ECONNREFUSED" ||
      error?.cause?.code === "ECONNREFUSED" ||
      error.message?.includes("fetch") ||
      error.message?.includes("ECONNREFUSED") ||
      error.message?.includes("NetworkError");

    if (isNetworkError) {
      throw new ApiError(503, "OFFLINE_SERVICE");
    }

    throw error;
  }
}

export function apiGet<T>(endpoint: string, options?: RequestInit) {
  // console.log("apiGet: ", endpoint, options)
  return fetchApi<T>(endpoint, { ...options, method: "GET" });
}

export function apiPost<T>(endpoint: string, data: any, options?: RequestInit) {
  const isFormData = data instanceof FormData;
  return fetchApi<T>(endpoint, {
    ...options,
    method: "POST",
    body: isFormData ? data : JSON.stringify(data),
  });
}

export function apiPatch<T>(endpoint: string, data: any, options?: RequestInit) {
  const isFormData = data instanceof FormData;
  return fetchApi<T>(endpoint, {
    ...options,
    method: "PATCH",
    body: isFormData ? data : JSON.stringify(data),
  });
}

export function apiDelete<T>(endpoint: string, options?: RequestInit) {
  return fetchApi<T>(endpoint, { ...options, method: "DELETE" });
}