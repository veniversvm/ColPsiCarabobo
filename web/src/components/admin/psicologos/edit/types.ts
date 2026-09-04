// web/src/components/admin/psicologos/edit/types.ts

// web/src/components/admin/psicologos/edit/... (donde esté definido)

export interface EditFormState {
  // 1. Identidad y Acceso
  username: string;
  email: string;
  first_name: string;
  second_name: string;
  last_name: string;
  second_last_name: string;
  ci: string | number;
  fpv: string | number;
  nationality: string;
  control_number: string;
  genre: string;
  born_date: string;

  // 2. Estatus Administrativo
  is_active: boolean;
  solvent: boolean;
  proof_of_life: boolean;
  ministry_registration_confirmed: boolean;

  // 3. Contacto y Privacidad General
  contact_email: string;
  show_contact_email: boolean;
  contact_phone: string;       // Antes public_phone
  contact_cell_phone: string;  // Nuevo
  service_address: string;
  show_public_service_address: boolean;

  // 4. Ubicación: Carabobo
  municipality_carabobo: string;
  show_municipality_carabobo: boolean; // Nuevo
  phone_carabobo: string;
  show_phone_carabobo: boolean;        // Nuevo
  cel_phone_carabobo: string;
  show_cel_phone_carabobo: boolean;    // Nuevo

  // 5. Ubicación: Fuera de Carabobo (Venezuela)
  state_outside: string;
  show_state_outside: boolean;         // Nuevo
  municipality_outside_carabobo: string;
  show_municipality_outside_carabobo: boolean; // Nuevo
  phone_outside_carabobo: string;
  show_phone_outside_carabobo: boolean;
  cel_phone_outside_carabobo: string;
  show_cel_phone_outside_carabobo: boolean;
  service_address_outside_carabobo: string;
  show_public_service_address_outside_carabobo: boolean;

  // 6. Ubicación: Exterior
  country: string;
  phone_outside_venezuela: string;
  show_phone_outside_venezuela: boolean;
  cell_phone_outside_venezuela: string; // Nuevo
  show_cel_phone_outside_venezuela: boolean; // Nuevo
  service_address_outside_venezuela: string;
  show_public_service_address_outside_venezuela: boolean;

  // 7. Profesional (Áreas de Desempeño)
  primary_work_area: string;   // Antes primary_specialty
  secondary_work_area: string; // Antes secondary_specialty
  mini_bio: string;
  full_bio: string;

  // 8. Datos Académicos y Registro
  guild_inscription_date: string; // Nuevo
  university_undergraduate: string;
  graduate_date: string;
  mention_undergraduate: string;
  register_number: string | number;
  register_title_state: string;
  register_title_date: string;
  register_folio: string;
  register_tome: string;

  // 9. Privacidad Académica
  show_university_undergraduate: boolean;
  show_graduate_date: boolean;
  show_mention_undergraduate: boolean;

  // 10. Banderas Institucionales
  guild_director: boolean;
  sixty_five_or_plus: boolean;
  guild_collaborator: boolean;
  public_employee: boolean;
  discapacity: boolean;        // Nuevo
  university_professor: boolean;
  double_guild: boolean;
  double_guild_location: string; // Nuevo
  date_of_last_solvency: string;
  
  solvencies: any[]; // Array de registros de solvencia
}

export interface Solvency {
  Date: string;
  PsiUserModelID: string;
  create_by: string;
  create_by_id: string;
  created_at: string;
  id: string;
  update_by: string;
  update_by_id: string;
  updated_at: string;
}

export interface SocialNetwork {
  id?: string;
  name: string;
  url: string;
}

export interface DeontologiaEntry {
  id?: string;
  content: string;
  created_at?: string;
  create_by?: string;
}

export interface ObservacionesEntry {
  id?: string;
  content: string;
  created_at?: string;
  create_by?: string;
}

export interface PostGrade {
  post_grade_title: string;
  post_grade_graduation_year: string;
}

export interface PsiProfile {
  id: string;
  username: string;
  email: string;
  fpv?: number;
  profile_picture_url?: string;
  is_active: boolean;
  solvent: boolean;
  proof_of_life: boolean;
  social_networks?: SocialNetwork[];
  post_grades?: PostGrade[];
  col_data?: Record<string, any>;
  create_by?: string;
  update_by?: string;
  updated_at?: string;
  [key: string]: any;
}
