// web/src/types/inscription.ts

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
  fpv: number;
  telefono: string;
  correo: string;
  fecha_nacimiento: string | null;
  titulo_universidad: string;
  titulo_fecha_graduacion: string | null;
  titulo_mencion: string;
  titulo_registro_numero: string;
  titulo_registro_estado: string;
  rif: string;
  foto_url: string;
  comprobante_url: string;
  status: string;
  control_number: string;
  notes: string;
  created_at: string;
  updated_at: string;
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
