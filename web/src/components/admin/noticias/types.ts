// web/src/components/admin/noticias/types.ts
export type PostStatus = "draft" | "published" | "archived" | "scheduled";

export interface Post {
  id: string;
  title: string;
  short_description: string;
  type: "public" | "psi";
  image_url: string;
  status: PostStatus;
  publish_at?: string;
  created_at: string;
  updated_at: string;
  create_by: string;
}

export const TYPE_LABELS: Record<string, { label: string; color: string }> = {
  public: { label: "Público",        color: "bg-emerald-100 text-emerald-700" },
  psi:    { label: "Solo Colegiados", color: "bg-blue-100 text-blue-700" },
};

export const STATUS_LABELS: Record<PostStatus, { label: string; color: string }> = {
  published: { label: "Publicado",  color: "bg-emerald-100 text-emerald-700" },
  draft:     { label: "Borrador",   color: "bg-amber-100 text-amber-700" },
  archived:  { label: "Archivado",  color: "bg-gray-100 text-gray-500" },
  scheduled: { label: "Programado", color: "bg-purple-100 text-purple-700" },
};

export const formatDate = (iso: string) => {
  if (!iso) return "";
  return new Date(iso).toLocaleDateString("es-VE", {
    day: "2-digit", month: "short", year: "numeric",
  });
};