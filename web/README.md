# Frontend — SolidStart

> **[⬆ Raíz](../)** — `web/`

Sitio web del Colegio de Psicólogos de Carabobo.

## Stack

- **Framework**: SolidStart (SolidJS) con TypeScript
- **Renderizado**: SSR + CSR
- **Ruteo**: Basado en archivos (`src/routes/`)
- **Build**: Vinxi / Nitro, preset `deno-server-legacy`
- **CSS**: Tailwind
- **Auth**: JWT via sessionStorage (cliente) + HttpOnly cookie (server)
- **Bucket**: MinIO/S3, URLs centralizadas vía `src/lib/bucket.ts`

## Scripts

```bash
npm run dev         # Servidor de desarrollo
npm run build       # Build y verifica errores
npm run typecheck   # Solo typecheck (tsc --noEmit)

# Preview del build:
deno task --config .output/deno.json start
```

## Convenciones

- Archivos en snake_case para español
- Un solo commit por fix
- Preservar funcionamiento existente ("no modifique el funcionamiento")

## Módulos de seguridad

| Archivo | Propósito |
|---------|-----------|
| `src/lib/sanitize-html.ts` | Sanitiza HTML con DOMPurify (innerHTML seguro) |
| `src/lib/bucket.ts` | Helper `bucketUrl()` / `siteUrl()` |
| `src/lib/errors.ts` | Traduce errores internos a mensajes amigables |
| `src/routes/robots.txt.ts` | robots.txt dinámico con URL del sitio |

Ver `AGENTS.md` en la raíz para el detalle completo de los cambios de seguridad.

**[⬆ Volver a raíz](../)**
