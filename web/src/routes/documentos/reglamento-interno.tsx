// Documento: Reglamento Interno del Colegio de Psicólogos del Estado Carabobo.

import DocumentLayout from "../../components/doc/DocumentLayout";
import * as doc from "../../lib/documentos/reglamento-interno";

export default function ReglamentoInternoRoute() {
  return <DocumentLayout doc={doc as never} />;
}
