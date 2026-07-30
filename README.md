# ColPsiCarabobo

Sitio web del Colegio de Psicólogos de Carabobo.

## Estructura

```
.
├── api/    → Backend REST API en Go (Clean Architecture)
└── web/   → Frontend SolidStart con TypeScript
```

| Directorio | Descripción |
|------------|-------------|
| [`api/`](./api/) | Backend REST API — Go, Fiber, PostgreSQL, S3 |
| [`web/`](./web/) | Frontend — SolidStart, Tailwind, Vinxi/Nitro |

## Inicio rápido

```bash
# Frontend
cd web && npm install && npm run dev

# Backend
cd api && docker compose up -d    # Ver api/README.md para detalles
```

Ver [`web/README.md`](./web/README.md) para comandos de build, typecheck y preview.
Ver [`api/README.md`](./api/README.md) para documentación de la API.

---
**[⬆ Arriba](#colpsicarabobo)**
