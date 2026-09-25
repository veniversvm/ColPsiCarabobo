// web/src/components/admin/psicologos/create/FormHeader.tsx
import { A } from "@solidjs/router";
import { PageHeader } from "~/components/admin/ui/PageHeader";

export function FormHeader() {
  return (
    <PageHeader
      crumbs={[{ label: "Psicólogos", href: "/admin/psicologos" }, { label: "Alta de Colegiado" }]}
      title="Alta de Colegiado"
      description="Apertura de nuevo expediente institucional (* Campos obligatorios)"
      actions={
        <A
          href="/admin/psicologos"
          class="inline-flex items-center justify-center h-9 px-3.5 rounded-md border border-colpsi-border bg-white text-sm font-medium text-colpsi-text hover:bg-colpsi-bg transition-colors"
        >
          Cancelar
        </A>
      }
    />
  );
}