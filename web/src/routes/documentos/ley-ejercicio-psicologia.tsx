// Documento: Ley de Ejercicio de la Psicología.

import DocumentLayout from "../../components/doc/DocumentLayout";
import * as doc from "../../lib/documentos/ley-ejercicio-psicologia";

export default function LeyEjercicioRoute() {
  return <DocumentLayout doc={doc as never} />;
}
