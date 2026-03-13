// web/src/components/admin/noticias/NoticiaCard.tsx
import { Show } from "solid-js";
import { Post } from "./types";
import { NoticiaImage } from "./NoticiaImage";
import { NoticiaMetadata } from "./NoticiaMetadata";
import { NoticiaActions } from "./NoticiaActions";

interface Props {
  post: Post;
  isBusy: boolean;
  onToggle: (post: Post) => void;
  onDelete: (id: string) => void;
}

export function NoticiaCard(props: Props) {
  const cardClass = () => {
    const base = "bg-white rounded-2xl border-2 transition-all duration-200 overflow-hidden";
    if (props.post.status === "published") {
      return `${base} border-gray-100 hover:border-blue-100`;
    }
    if (props.post.status === "archived") {
      return `${base} border-dashed border-gray-200 opacity-50`;
    }
    return `${base} border-dashed border-gray-200 opacity-70`;
  };

  return (
    <article class={cardClass()}>
      <div class="flex items-start gap-4 p-4 md:p-5">
        <NoticiaImage imageUrl={props.post.image_url} alt={props.post.title} />
        
        <div class="flex-1 min-w-0">
          <NoticiaMetadata post={props.post} />
        </div>

        <NoticiaActions
          post={props.post}
          isBusy={props.isBusy}
          onToggle={props.onToggle}
          onDelete={props.onDelete}
        />
      </div>
    </article>
  );
}