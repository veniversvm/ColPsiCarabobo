// web/src/components/admin/noticias/NoticiaImage.tsx
import { Show } from "solid-js";
import { bucketUrl } from "~/lib/bucket";
import { Icon } from "~/components/admin/ui/icons";

interface Props {
  imageUrl?: string;
  alt: string;
}

export function NoticiaImage(props: Props) {
  return (
    <div class="flex-shrink-0 w-20 h-20 md:w-24 md:h-24 rounded-md overflow-hidden bg-colpsi-bg border border-colpsi-border">
      <Show
        when={props.imageUrl}
        fallback={<div class="w-full h-full flex items-center justify-center text-colpsi-muted"><Icon name="newspaper" class="w-6 h-6" /></div>}
      >
        <img
          src={bucketUrl(props.imageUrl)}
          alt={props.alt}
          class="w-full h-full object-cover"
          loading="lazy"
        />
      </Show>
    </div>
  );
}