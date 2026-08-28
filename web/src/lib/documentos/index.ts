// Marco Legal y Normativo — catálogo central de documentos estáticos.
// Los datos viven en módulos tipados por documento (convertidos desde Docs/analisis).

import * as estatutosFpv from "./estatutos-fpv";
import * as leyEjercicioPsicologia from "./ley-ejercicio-psicologia";
import * as codigoEtica from "./codigo-etica";
import * as reglamentoInterno from "./reglamento-interno";

export interface DocumentoBloque {
  tipo: "capitulo" | "articulo" | "texto";
  texto?: string;
  numero?: string;
}

export interface DocumentoSeccion {
  titulo: string;
  bloques: DocumentoBloque[];
}

export type DocModulo = {
  slug: string;
  titulo: string;
  fuente: string;
  categoria: string;
  descripcion: string;
  archivo: string;
  tipoArchivo: string;
  secciones: DocumentoSeccion[];
};

export type ControlSecciones = {
  accion: "abrir" | "colapsar";
  tick: number;
};

export const documentos: DocModulo[] = [
  estatutosFpv,
  leyEjercicioPsicologia,
  codigoEtica,
  reglamentoInterno,
] as unknown as DocModulo[];

export function getDocumento(slug?: string): DocModulo | undefined {
  if (!slug) return undefined;
  return documentos.find((d) => d.slug === slug);
}
