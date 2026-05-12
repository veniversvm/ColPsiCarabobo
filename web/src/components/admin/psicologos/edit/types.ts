// web/src/components/admin/psicologos/edit/types.ts

export interface EditFormState {
  // Cuenta
  username: string;
  email: string;

  // Identidad Legal
  first_name: string;
  second_name: string;
  last_name: string;
  second_last_name: string;
  ci: string;
  fpv: string;
  nationality: string;
  genre: string;
  born_date: string;

  // Estatus Institucional
  is_active: boolean;
  solvent: boolean;
  proof_of_life: boolean;

  // Contacto público
  contact_email: string;
  show_contact_email: boolean;
  public_phone: string;
  show_public_phone: boolean;
  service_address: string;
  show_public_service_address: boolean;

  // Carabobo
  municipality_carabobo: string;
  phone_carabobo: string;
  cel_phone_carabobo: string;

  // Fuera de Carabobo
  state_outside: string;
  municipality_outside_carabobo: string;
  phone_outside_carabobo: string;
  cel_phone_outside_carabobo: string;
  service_address_outside_carabobo: string;
  show_phone_outside_carabobo: boolean;
  show_cel_phone_outside_carabobo: boolean;
  show_public_service_address_outside_carabobo: boolean;

  // Exterior
  country: string;
  phone_outside_venezuela: string;
  service_address_outside_venezuela: string;
  show_phone_outside_venezuela: boolean;
  show_cel_phone_outside_venezuela: boolean;
  show_public_service_address_outside_venezuela: boolean;

  // Perfil Profesional
  primary_specialty: string;
  secondary_specialty: string;
  mini_bio: string;
  full_bio: string;

  // Registro Académico
  university_undergraduate: string;
  graduate_date: string;
  mention_undergraduate: string;
  register_number: string;
  register_title_state: string;
  register_title_date: string;
  register_folio: string;
  register_tome: string;

  // Privacidad académica
  show_university_undergraduate: boolean;
  show_graduate_date: boolean;
  show_mention_undergraduate: boolean;

  // Flags Gremiales
  guild_director: boolean;
  sixty_five_or_plus: boolean;
  guild_collaborator: boolean;
  public_employee: boolean;
  university_professor: boolean;
  double_guild: boolean;
  cpsm: boolean;
  date_of_last_solvency: string;
  solvencies: Solvency[];
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
