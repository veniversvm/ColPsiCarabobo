// web/src/types/auth.ts
export type UserRole = "admin" | "psi";

export interface AuthUser {
  id: string;
  username: string;
  email: string;
  role: UserRole;
  // Campos extra opcionales según el rol
  firstName?: string;
  lastName?: string;
}

export interface AuthState {
  user: AuthUser | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}