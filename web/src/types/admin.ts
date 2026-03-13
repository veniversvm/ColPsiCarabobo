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

export type PsicologoForm = {
  username: string; email: string; password: string;
  first_name: string; second_name: string; last_name: string; second_last_name: string;
  ci: string; fpv: string; nationality: string; genre: string; born_date: string;
  is_active: boolean; solvent: boolean; proof_of_life: boolean;
  public_phone: string; service_address: string;
  municipality_carabobo: string; phone_carabobo: string; cel_phone_carabobo: string;
  state_outside: string; municipality_outside_carabobo: string; 
  phone_outside_carabobo: string; cel_phone_outside_carabobo: string;
  primary_specialty: string; secondary_specialty: string;
  university_undergraduate: string; graduate_date: string; mention_undergraduate: string;
  register_number: string; register_title_state: string; register_title_date: string; 
  register_folio: string; register_tome: string;
  guild_director: boolean; sixty_five_or_plus: boolean; guild_collaborator: boolean; 
  public_employee: boolean; university_professor: boolean; double_guild: boolean; 
  cpsm: boolean; date_of_last_solvency: string;
};