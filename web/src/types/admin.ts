// web/src/types/admin.ts
export interface PsiAdminListItem {
  id: string;
  first_name: string;
  last_name: string;
  ci: number;
  fpv: number;
  email: string;
  control_number: string;
  age?: number;
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

export interface PsicologoForm {
  username: string;
  email: string;
  password?: string;

  first_name: string;
  second_name?: string;
  last_name: string;
  second_last_name?: string;
  ci: string | number;
  fpv: string | number;
  nationality: string;
  genre: string;
  born_date: string;

  is_active: boolean;
  solvent: boolean;
  proof_of_life: boolean;

  contact_email: string;
  contact_phone: string;
  contact_cell_phone: string;
  service_address: string;

  municipality_carabobo: string;
  phone_carabobo: string;
  cel_phone_carabobo: string;

  state_outside: string;
  municipality_outside_carabobo: string;
  phone_outside_carabobo: string;
  cel_phone_outside_carabobo: string;
  service_address_outside_carabobo: string;

  country: string;
  phone_outside_venezuela: string;
  cell_phone_outside_venezuela: string;
  service_address_outside_venezuela: string;

  primary_work_area: string;   // Antes specialty
  secondary_work_area: string; // Antes specialty

  guild_inscription_date: string;
  university_undergraduate: string;
  graduate_date: string;
  mention_undergraduate: string;

  register_number: string | number;
  register_title_state: string;
  register_title_date: string;
  register_folio: string;
  register_tome: string;

  guild_director: boolean;
  sixty_five_or_plus: boolean;
  guild_collaborator: boolean;
  public_employee: boolean;
  discapacity: boolean;
  university_professor: boolean;
  double_guild: boolean;
  double_guild_location: string;
  cpsm: boolean;
  date_of_last_solvency: string;
}