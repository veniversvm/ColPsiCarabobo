import DOMPurify from "dompurify";

export function sanitizeHtml(html: string): string {
  if (!html) return "";
  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS: [
      "p", "br", "b", "i", "u", "strong", "em", "s", "sub", "sup",
      "h1", "h2", "h3", "h4", "h5", "h6",
      "ul", "ol", "li",
      "a", "img",
      "blockquote", "pre", "code", "hr",
      "table", "thead", "tbody", "tr", "th", "td",
      "div", "span", "figure", "figcaption",
    ],
    ALLOWED_ATTR: ["href", "src", "alt", "class", "target", "rel", "width", "height"],
    ALLOW_DATA_ATTR: false,
  });
}
