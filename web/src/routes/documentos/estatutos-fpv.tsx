// Documento: Estatutos de la Federación de Psicólogos de Venezuela (FPV).

import DocumentLayout from "../../components/doc/DocumentLayout";
import * as doc from "../../lib/documentos/estatutos-fpv";

export default function EstatutosFpvRoute() {
  return <DocumentLayout doc={doc as never} />;
}
