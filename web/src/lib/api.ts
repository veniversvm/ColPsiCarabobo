import Cookies from "js-cookie";
import { isServer } from "solid-js/web";

// La URL de tu backend en Go. En producción esto vendría de process.env o import.meta.env
const API_BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api/v1";

// Estructura de un error estándar que devuelve nuestra API de Go
export class ApiError extends Error {
  constructor(public status: number, message: string, public data: any = null) {
    super(message);
    this.name = "ApiError";
  }
}

/**
 * Función centralizada para realizar peticiones al backend en Go.
 * Maneja la inyección automática del JWT y el parseo de errores.
 */
export async function fetchApi<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  // --- 1. Preparación (Fuera del try/catch de red) ---
  const url = `${API_BASE_URL}${endpoint}`;
  const headers = new Headers(options.headers || {});
  
  if (!headers.has("Content-Type") && !(options.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }

  // Lógica isomórfica de Token
  let token = "";
  if (isServer) {
    const { getRequestEvent } = await import("solid-js/web");
    const event = getRequestEvent();
    const cookieHeader = event?.request.headers.get("cookie") || "";
    const match = cookieHeader.match(new RegExp('(^| )jwt=([^;]+)'));
    if (match) token = match[2];
  } else {
    token = Cookies.get("jwt") || "";
  }

  if (token) headers.set("Authorization", `Bearer ${token}`);

  // --- 2. Ejecución y Captura de Fallos de Conexión ---
  try {
    const response = await fetch(url, { ...options, headers });

    // Manejo de respuestas no exitosas (400, 401, 404, 500)
    if (!response.ok) {
      let errorData: any = null;
      try {
        errorData = await response.json();
      } catch { /* No es JSON */ }

      if (response.status === 401 && !isServer) {
        Cookies.remove("jwt");
      }

      const msg = errorData?.error || errorData?.message || "Error inesperado";
      throw new ApiError(response.status, msg, errorData);
    }

    return await response.json() as T;

  } catch (error: any) {
    // Si ya es un ApiError (ej. un 400), lo relanzamos para el ErrorBoundary
    if (error instanceof ApiError) throw error;

    // Detectamos desconexión del backend
    const isNetworkError = 
      error.message.includes("fetch") || 
      error.message.includes("connection") || 
      error.message.includes("NetworkError");

    if (isNetworkError) {
      // 503 Service Unavailable + Código amigable para el UI
      throw new ApiError(503, "OFFLINE_SERVICE");
    }

    throw error;
  }
}

// Helpers para no tener que escribir 'method: "POST"' en cada componente

export function apiGet<T>(endpoint: string, options?: RequestInit) {
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