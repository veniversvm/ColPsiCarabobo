// web/src/components/admin/noticias/NoticiasHeader.tsx
import { A } from "@solidjs/router";
import { PageHeader } from "~/components/admin/ui/PageHeader";
import { Icon } from "~/components/admin/ui/icons";

export function NoticiasHeader() {
  return (
    <PageHeader
      crumbs={[{ label: "Noticias" }]}
      title="Publicaciones"
      description="Gestión de noticias y comunicados del Colegio"
      actions={
        <A
          href="/admin/noticias/crear"
          class="inline-flex items-center justify-center gap-2 h-9 px-3.5 rounded-md bg-colpsi-blue text-white text-sm font-semibold transition-colors outline-none hover:bg-colpsi-blue-light focus:ring-2 focus:ring-colpsi-blue/30"
        >
          <Icon name="plus" />
          Nueva Publicación
        </A>
      }
    />
  );
}