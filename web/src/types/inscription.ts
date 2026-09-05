// web/src/types/inscription.ts

export interface WorkArea {
  id: number;
  name: string;
  description: string;
  active: boolean;
  created_at: string;
  updated_at: string;
  create_by: string;
  update_by: string;
}

export type InscriptionDocumentType = "cedula" | "titulo" | "rif" | "otro";

export interface InscriptionDocument {
  id: string;
  document_type: InscriptionDocumentType;
  url: string;
  title?: string;
  notes?: string;
  original_filename?: string;
}

export interface InscriptionListItem {
  id: string;
  cedula: number;
  nombres: string;
  apellidos: string;
  fpv: number;
  correo: string;
  status: string; // pending | approved | rejected
  control_number: string;
  created_at: string;
}

export interface InscriptionListResponse {
  items: InscriptionListItem[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface InscriptionDetail {
  id: string;
  cedula: number;
  nacionalidad: string;
  nombres: string;
  apellidos: string;
  segundo_nombre: string;
  segundo_apellido: string;
  genero: string;
  fpv: number;
  telefono: string;
  correo: string;
  fecha_nacimiento: string | null;
  titulo_universidad: string;
  titulo_fecha_graduacion: string | null;
  titulo_mencion: string;
  titulo_registro_numero: string;
  titulo_registro_estado: string;
  titulo_registro_tomo: string;
  titulo_registro_folio: string;
  rif: string;
  service_address: string;
  municipality_carabobo: string;
  state_outside: string;
  municipality_outside_carabobo: string;
  country: string;
  service_modality_presencial: boolean;
  service_modality_distance: boolean;
  service_modality_telephone: boolean;
  primary_specialty_id?: number;
  secondary_specialty_id?: number;
  foto_url: string;
  comprobante_url: string;
  documents: InscriptionDocument[];
  status: string;
  control_number: string;
  notes: string;
  psi_user_id: string | null;
  solvency_count: number;
  created_at: string;
  updated_at: string;
}

export interface UpdateInscriptionRequest {
  cedula: number;
  nacionalidad: string;
  nombres: string;
  apellidos: string;
  segundo_nombre: string;
  segundo_apellido: string;
  genero: string;
  fpv: number;
  telefono: string;
  correo: string;
  fecha_nacimiento?: string | null;
  titulo_universidad: string;
  titulo_fecha_graduacion?: string | null;
  titulo_mencion: string;
  titulo_registro_numero: string;
  titulo_registro_estado: string;
  titulo_registro_tomo: string;
  titulo_registro_folio: string;
  rif: string;
  service_address: string;
  municipality_carabobo: string;
  state_outside: string;
  municipality_outside_carabobo: string;
  country: string;
  service_modality_presencial: boolean;
  service_modality_distance: boolean;
  service_modality_telephone: boolean;
  primary_specialty_id?: number | null;
  secondary_specialty_id?: number | null;
}

export interface ApproveInscriptionResponse {
  message: string;
  psi_user_id: string;
  control_number: string;
  email_sent: boolean;
}

export interface UniquenessCheckResponse {
  exists: boolean;
  message?: string;
}

export interface SendEmailToApplicantResponse {
  email_sent: boolean;
}
