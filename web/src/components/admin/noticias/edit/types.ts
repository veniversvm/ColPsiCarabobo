// web/src/components/admin/noticias/edit/types.ts
import { bucketUrl } from "~/lib/bucket";
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

export const STATUS_OPTIONS: { value: PostStatus; label: string }[] = [
  { value: "draft",     label: "Borrador" },
  { value: "published", label: "Publicado" },
  { value: "archived",  label: "Archivado" },
  { value: "scheduled", label: "Programado" },
];

export const STATUS_BADGE: Record<PostStatus, string> = {
  published: "bg-emerald-50 text-emerald-700 border-emerald-200",
  draft:     "bg-amber-50 text-amber-700 border-amber-200",
  archived:  "bg-slate-100 text-slate-600 border-slate-200",
  scheduled: "bg-purple-50 text-purple-700 border-purple-200",
};

export const imgUrl = (key: string) => bucketUrl(key);

export const IC = "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 transition-colors text-colpsi-text";
export const labelClass = "block text-[11px] font-semibold text-colpsi-muted uppercase tracking-wide ml-1 mb-1";