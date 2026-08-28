// Documento: Código de Ética Profesional del Psicólogo.

import DocumentLayout from "../../components/doc/DocumentLayout";
import * as doc from "../../lib/documentos/codigo-etica";

export default function CodigoEticaRoute() {
  return <DocumentLayout doc={doc as never} />;
}
