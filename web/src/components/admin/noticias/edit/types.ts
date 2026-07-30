// web/src/components/admin/noticias/edit/types.ts
export type PostStatus = "draft" | "published" | "archived" | "scheduled";

export interface PostDetail {
  id: string;
  title: string;
  short_description: string;
  type: "public" | "psi";
  status: PostStatus;
  publish_at?: string;
  image_url: string;
  text: { id: string; content: string };
}

export const STATUS_OPTIONS: { value: PostStatus; label: string; icon: string }[] = [
  { value: "draft",     label: "Borrador",   icon: "📝" },
  { value: "published", label: "Publicado",  icon: "✅" },
  { value: "archived",  label: "Archivado",  icon: "📦" },
  { value: "scheduled", label: "Programado", icon: "⏰" },
];

export const STATUS_BADGE: Record<PostStatus, string> = {
  published: "bg-emerald-100 text-emerald-700",
  draft:     "bg-amber-100 text-amber-700",
  archived:  "bg-gray-100 text-gray-500",
  scheduled: "bg-purple-100 text-purple-700",
};

export const BUCKET_URL = import.meta.env.VITE_BUCKET_URL || "";
export const imgUrl = (key: string) => (key ? `${BUCKET_URL}/${key}` : "");

export const IC = "w-full bg-white border-2 border-gray-200 focus:border-blue-500 rounded-xl px-4 py-2.5 outline-none transition-all text-gray-800 text-sm";
export const labelClass = "block text-[10px] font-black text-gray-500 uppercase tracking-widest ml-1 mb-1";