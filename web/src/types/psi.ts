// web/src/types/psi.ts

export type PostGrade = {
  title: string;
  university: string;
  year: string;
  description?: string;
  pic_one_url?: string;
  pic_two_url?: string;
  pic_three_url?: string;
};

// web/src/types/psi.ts
export interface Undergraduate {
  university?: string;
  date?: string;
  mention?: string;
  title_image_one_url?: string;
  title_image_two_url?: string;
  title_image_three_url?: string;
  register_number?: number;
  register_folio?: string;
  register_tome?: string;
  register_title_date?: string;
  register_title_state?: string;
}

// ── Ubicación multi-zona ──────────────────────────────────────────────────────
// Cada bloque es opcional — solo viene si el psicólogo tiene presencia en esa zona.

export type LocationCarabobo = {
  municipality: string;
  phone?: string;
  cell_phone?: string;
  address?: string;
};

export type LocationVenezuela = {
  state: string;
  municipality?: string;
  phone?: string;
  cell_phone?: string;
  address?: string;
};

export type LocationExterior = {
  country: string;
  phone?: string;
  address?: string;
};

export type PsiLocation = {
  carabobo?: LocationCarabobo;
  venezuela?: LocationVenezuela;
  exterior?: LocationExterior;
};

// ── Perfil público (directorio) ───────────────────────────────────────────────
// solvent eliminado — nunca debe exponerse al público.
export type PsiProfile = {
  id: string;
  first_name: string;
  second_name?: string;
  last_name: string;
  second_last_name?: string;
  fpv: number;
  ci: number;
  gender: string;
  profile_picture: string;
  email?: string;
  phone?: string;
  address?: string;
  location: PsiLocation;
  specialties: string[];
  mini_bio?: string;
  full_bio_content?: string;
  undergraduate: Undergraduate;
  post_grades?: PostGrade[];
  social_networks?: SocialNetwork[];
};

// ── Ajustes del propio perfil (autogestión) ───────────────────────────────────
export type PsiProfileSettings = {
  id: string;
  username: string;
  email: string;
  contact_email?: string;
  public_phone?: string;
  service_address?: string;

  first_name: string;
  second_name?: string;
  last_name: string;
  second_last_name?: string;

  // Ubicación: Carabobo
  municipality_carabobo?: string;
  phone_carabobo?: string;
  cel_phone_carabobo?: string;

  // Ubicación: Fuera de Carabobo (Venezuela)
  state_outside?: string;
  municipality_outside_carabobo?: string;
  phone_outside_carabobo?: string;
  cel_phone_outside_carabobo?: string;
  service_address_outside_carabobo?: string;

  // Ubicación: Fuera de Venezuela
  country?: string;
  phone_outside_venezuela?: string;
  service_address_outside_venezuela?: string;

  // Perfil profesional
  mini_bio?: string;
  full_bio?: string;
  primary_specialty?: string;
  secondary_specialty?: string;

  // Privacidad: Contacto principal
  show_contact_email: boolean;
  show_public_phone: boolean;
  show_public_service_address: boolean;

  // Privacidad: Fuera de Carabobo
  show_phone_outside_carabobo: boolean;
  show_cel_phone_outside_carabobo: boolean;
  show_public_service_address_outside_carabobo: boolean;

  // Privacidad: Exterior
  show_phone_outside_venezuela: boolean;
  show_cel_phone_outside_venezuela: boolean;
  show_public_service_address_outside_venezuela: boolean;

  // Privacidad: Datos colegiales
  show_university_undergraduate: boolean;
  show_graduate_date: boolean;
  show_mention_undergraduate: boolean;

  social_networks?: SocialNetwork[];
};

export type ProfileFormData = PsiProfileSettings & {
  password: string;
  new_password_1: string;
  new_password_2: string;
};

export type SocialNetwork = {
  id?: string;
  name: string;
  url: string;
};

export type DirectoryPsychologist = {
  id: string;
  first_name: string;
  last_name: string;
  second_name?: string;
  second_last_name?: string;
  fpv: number;
  ci: number;
  profile_picture: string;
  specialties: string[];
  mini_bio: string;
};

export interface PostGradeModalProps {
  postGrade: PostGrade | null;
  onClose: () => void;
}

