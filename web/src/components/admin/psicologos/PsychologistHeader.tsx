// web/src/components/admin/psicologos/PsychologistHeader.tsx
import { A } from "@solidjs/router";
import { PageHeader } from "~/components/admin/ui/PageHeader";
import { Button } from "~/components/admin/ui/Button";
import { Icon } from "~/components/admin/ui/icons";

interface Props {
  title?: string;
  onImportClick: () => void;
}

const btnLinkBase =
  "inline-flex items-center justify-center gap-2 h-9 px-3.5 rounded-md text-sm font-semibold transition-colors outline-none focus:ring-2";

export function PsychologistHeader(props: Props) {
  return (
    <PageHeader
      crumbs={[{ label: "Psicólogos" }]}
      title={props.title ?? "Gestión de Agremiados"}
      description="Base de datos maestra de profesionales colegiados."
      actions={
        <>
          <Button variant="secondary" onClick={props.onImportClick}>
            <Icon name="fileText" />
            Importar CSV
          </Button>
          <A
            href="/admin/psicologos/crear"
            class={`${btnLinkBase} bg-colpsi-blue text-white hover:bg-colpsi-blue-light focus:ring-colpsi-blue/30`}
          >
            <Icon name="plus" />
            Nuevo Registro
          </A>
        </>
      }
    />
  );
}