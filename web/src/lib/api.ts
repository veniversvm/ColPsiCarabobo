// web/src/lib/api.ts
import Cookies from "js-cookie";
import { isServer } from "solid-js/web";
import { getRequestEvent } from "solid-js/web";

const API_BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api/v1";

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

  // Token isomórfico: servidor lee la cookie del request, cliente usa js-cookie
  let token = "";
  if (isServer) {
    const event = getRequestEvent();
    const cookieHeader = event?.request.headers.get("cookie") || "";
    const match = cookieHeader.match(/(^| )jwt=([^;]+)/);
    if (match) token = match[2];
  } else {
    token = Cookies.get("jwt") || "";
  }

  if (token) headers.set("Authorization", `Bearer ${token}`);

  try {
    const response = await fetch(url, { ...options, headers });

    if (!response.ok) {
      let errorData: any = null;
      try { errorData = await response.json(); } catch { /* no JSON */ }

      if (response.status === 401 && !isServer) {
        Cookies.remove("jwt");
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