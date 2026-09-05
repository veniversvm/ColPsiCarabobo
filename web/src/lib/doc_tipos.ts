// web/src/lib/doc_tipos.ts

// Constantes compartidas del módulo de Registro Digital de Documentos del
// expediente del psicólogo (mismas categorías que el backend Go).
export const DOCUMENT_TYPE_LABELS: Record<string, string> = {
  cedula: "Cédula",
  titulo: "Título",
  rif: "RIF",
  solvencia: "Solvencia",
  comprobante: "Comprobante",
  otro: "Otro",
};

export const DOCUMENT_TYPE_EMOJI: Record<string, string> = {
  cedula: "🪪",
  titulo: "🎓",
  rif: "🧾",
  solvencia: "✅",
  comprobante: "📄",
  otro: "📁",
};

export const DOCUMENT_TYPE_ORDER = [
  "cedula",
  "titulo",
  "rif",
  "solvencia",
  "comprobante",
  "otro",
] as const;

// Un documento es PDF si el backend lo marcó como application/pdf o su nombre
// original termina en .pdf (los archivos se almacenan con su extensión real).
export const isPdf = (doc: {
  mime_type?: string;
  document_url?: string;
  filename?: string;
}): boolean =>
  doc?.mime_type === "application/pdf" ||
  /\.pdf$/i.test(doc?.filename ?? "");