// web/src/types/admin.ts
export interface PsiAdminListItem {
  id: string;
  first_name: string;
  last_name: string;
  ci: number;
  fpv: number;
  email: string;
  solvent: boolean;
  is_active: boolean;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}