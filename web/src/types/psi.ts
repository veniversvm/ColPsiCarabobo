// web/src/types/psi.ts

export type PostGrade = {
  type: string; // "diplomado", "especializacion", "maestria", "doctorado"
  title: string;
  university: string;
  year: string;
  description?: string;
  pic_one_url?: string;
  pic_two_url?: string;
  pic_three_url?: string;
};

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
  register_title_date_raw?: string;
  register_title_state?: string;
  birthday_notification?: boolean;
}

// ── Ubicación multi-zona ──────────────────────────────────────────────────────

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
  cell_phone?: string; // Añadido
  address?: string;
};

export type PsiLocation = {
  carabobo?: LocationCarabobo;
  venezuela?: LocationVenezuela;
  exterior?: LocationExterior;
};

// ── Perfil público (directorio) ───────────────────────────────────────────────
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
  location: PsiLocation;
  specialties: string[]; // Aquí el backend mapea las WorkAreas
  mini_bio?: string;
  full_bio_content?: string;
  service_modality?: { presencial: boolean; distance: boolean; telephone: boolean };
  undergraduate: Undergraduate;
  post_grades?: PostGrade[];
  social_networks?: SocialNetwork[];
};

// ── Ajustes del propio perfil (autogestión) ───────────────────────────────────
export type PsiProfileSettings = {
  id: string;
  username: string;
  email: string;
  
  // Contacto principal
  contact_email?: string;
  contact_phone?: string;      // Reemplaza a public_phone
  contact_cell_phone?: string; // Nuevo
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
  cell_phone_outside_venezuela?: string; // Nuevo
  service_address_outside_venezuela?: string;

  // Perfil profesional
  mini_bio?: string;
  full_bio?: string;
  primary_work_area?: string;   // Reemplaza a specialty
  secondary_work_area?: string; // Reemplaza a specialty

  // Modalidad de servicio (auto-gestión)
  service_modality_presencial?: boolean;
  service_modality_distance?: boolean;
  service_modality_telephone?: boolean;
  show_service_modality?: boolean;

  // Aviso de cumpleaños (opt-in)
  birthday_notification: boolean;

  // --- PRIVACIDAD (Privacy Shield) ---
  
  // Privacidad: Contacto principal
  show_contact_email: boolean;
  show_public_service_address: boolean;

  // Privacidad: Carabobo
  show_municipality_carabobo: boolean;
  show_phone_carabobo: boolean;
  show_cel_phone_carabobo: boolean;

  // Privacidad: Fuera de Carabobo
  show_state_outside: boolean;
  show_municipality_outside_carabobo: boolean;
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
  col_data?: Undergraduate; // Datos académicos crudos
};

export type ProfileFormData = PsiProfileSettings & {
  password?: string;
  new_password_1?: string;
  new_password_2?: string;
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
  service_modality?: { presencial: boolean; distance: boolean; telephone: boolean };
};

export interface PostGradeModalProps {
  postGrade: PostGrade | null;
  onClose: () => void;
}

// ── Registro Digital de Documentos del expediente ─────────────────────────────
// La gestión (carga/edición/borrado) es EXCLUSIVA de la administración; el
// psicólogo solo los consulta en /psi/documentos. El backend resuelve
// `document_url` a la URL pública del bucket.
export type PsiUserDocumentType =
  | "cedula"
  | "titulo"
  | "rif"
  | "solvencia"
  | "comprobante"
  | "otro";

export interface PsiUserDocument {
  id: string;
  psi_user_id?: string;
  document_type?: PsiUserDocumentType;
  title: string;
  notes?: string;
  document_date?: string | null;
  document_url?: string;
  filename?: string;
  mime_type?: string;
  size_bytes?: number;
  created_at?: string;
  updated_at?: string;
  create_by?: string;
  update_by?: string;
}