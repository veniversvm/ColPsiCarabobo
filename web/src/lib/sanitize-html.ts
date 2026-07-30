import DOMPurify from "dompurify";

const ALLOWED_TAGS = [
  "p", "br", "b", "i", "u", "strong", "em", "s", "sub", "sup",
  "h1", "h2", "h3", "h4", "h5", "h6",
  "ul", "ol", "li",
  "a", "img",
  "blockquote", "pre", "code", "hr",
  "table", "thead", "tbody", "tr", "th", "td",
  "div", "span", "figure", "figcaption",
];

const ALLOWED_ATTR = ["href", "src", "alt", "class", "target", "rel", "width", "height"];

/**
 * Sanitiza HTML peligroso usando DOMPurify.
 * Solo permite un subconjunto seguro de etiquetas (p, a, img, tablas, etc.)
 * y atributos (href, src, class, etc.). No permite data-attributes.
 * Retorna string vacío si la entrada es nula/vacía.
 */
export function sanitizeHtml(html: string): string {
  if (!html) return "";
  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS,
    ALLOWED_ATTR,
    ALLOW_DATA_ATTR: false,
  });
}
