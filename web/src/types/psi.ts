// export interface PostGrade {
//   id: string;
//   post_grade_title: string;
//   post_grade_university: string;
//   post_grade_graduation_year: string;
//   post_grade_description?: string;
//   pic_one_url?: string;
//   pic_two_url?: string;
//   pic_three_url?: string;
//   is_active: boolean;
// }

// web/src/types/psi.ts
// Tipos unificados para psicólogos

export type PostGrade = {
  title: string;
  university: string;
  year: string;
  description?: string;
  pic_one_url?: string;
  pic_two_url?: string;
  pic_three_url?: string;
};

export type Undergraduate = {
  university?: string;
  date?: string;
  mention?: string;
};

export type Location = {
  state: string;
  municipality: string;
};

export type PsiProfile = {
  id: string;
  first_name: string;
  last_name: string;
  fpv: number;
  gender: string;
  profile_picture: string;
  solvent: boolean;
  email?: string;
  phone?: string;
  address?: string;
  location: Location;
  specialties: string[];
  mini_bio: string;
  undergraduate: Undergraduate;
  post_grades?: PostGrade[];
  social_networks?: SocialNetwork[];
};

// Props para componentes
export interface PostGradeModalProps {
  postGrade: PostGrade | null;
  onClose: () => void;
}

export type PsiProfileSettings = {
  username: string;
  email: string;
  contact_email?: string;
  public_phone?: string;
  service_address?: string;
  
  // Ubicación Carabobo
  municipality_carabobo?: string;
  phone_carabobo?: string;
  cel_phone_carabobo?: string;
  
  // Ubicación Exterior
  state_outside?: string;
  municipality_outside_carabobo?: string;
  phone_outside_carabobo?: string;
  cel_phone_outside_carabobo?: string;
  
  // Perfil profesional
  mini_bio?: string;
  primary_specialty?: string;
  secondary_specialty?: string;
  
  // Privacidad
  show_contact_email: boolean;
  show_public_phone: boolean;
  show_public_service_address: boolean;
  show_university_undergraduate: boolean;
  show_graduate_date: boolean;
  show_mention_undergraduate: boolean;
  
  // Redes sociales
  social_networks?: SocialNetwork[];
};

export type SocialNetwork = {
  id?: string;
  name: string;
  url: string;
};

export type ProfileFormData = PsiProfileSettings & {
  password: string;
  new_password_1: string;
  new_password_2: string;
};


// web/src/types/psi.ts (agregamos este tipo)
export type DirectoryPsychologist = {
  id: string;
  first_name: string;
  last_name: string;
  fpv: number;
  profile_picture: string;
  specialties: string[];
  mini_bio: string;
};