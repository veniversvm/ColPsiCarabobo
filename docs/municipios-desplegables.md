# Municipios y estados como desplegables en cascada (formularios)

> **Estado: implementado** (commit `f706e3d` en `main`).
> Solo frontend: **sin cambios en API ni BD**. Los selects envían al API los
> mismos strings que antes enviaban los inputs de texto (p. ej. `state_outside`,
> `municipality_outside_carabobo`), así que no se modifica ningún contrato.

## Objetivo

Sustituir los campos de texto de ubicación por desplegables en los formularios
del frontend:

- **Municipio (Carabobo)**: catálogo `MUNICIPIOS_CARABOBO`.
- **Otro estado (fuera de Carabobo)**: catálogo `ESTADOS_VENEZUELA`.
- **Municipio / ciudad (fuera de Carabobo)**: cascada — los municipios se
  filtran según el estado elegido.

## Catálogo — `web/src/lib/geo.ts`

| Export | Contenido |
|---|---|
| `MUNICIPIOS_CARABOBO` | Los 14 municipios de Carabobo (ya existía). |
| `ESTADOS_VENEZUELA` | Los 23 estados + Dependencias Federales + Distrito Capital (ya existía). |
| `ESTADOS_VENEZUELA_INCL_CARABOBO` | `["Carabobo", ...ESTADOS_VENEZUELA]` — para filtros donde la base también aplica (Notificaciones). |
| `MUNICIPIOS_POR_ESTADO` | Municipios oficiales de las 24 entidades (321 + los 14 de Carabobo = **335**, cuadrando con el censo oficial INE/CONARE). Casos especiales: `Distrito Capital → [Libertador]`, `La Guaira → [Vargas]`, `Dependencias Federales → []`. |
| `municipiosDe(estado)` | `MUNICIPIOS_POR_ESTADO[estado] ?? []`. |

> El backend (`api/internal/utils/geo_venezuela.go`) valida solo presencia del
> municipio "fuera de Carabobo"; cualquier valor del catálogo pasa, por eso el
> frontend conserva valores legacy como opción "(no estándar)".

## Comportamiento

- **Cascada**: el select de municipio se deshabilita hasta elegir un estado; al
  cambiar de estado, si el municipio elegido ya no pertenece al nuevo estado se
  limpia automáticamente.
- **Legacy "(no estándar)"**: si hay un valor persistido (borrador de ficha,
  perfil psi o psicólogo en BBDD) que no está en el catálogo, aparece al final
  del select como `{valor} (no estándar)` para conservarlo sin pérdida de datos
  ni selects en blanco.
- **Estados sin municipios**: `Dependencias Federales` muestra el placeholder
  "Este estado no tiene municipios".
- **Notificaciones**: `Municipio` → desplegable de los 14 de Carabobo y
  `Estado` → desplegable con Carabobo incluido; ambos con la opción "Todos"
  (`value=""`) que preserva el filtro vacío original.

## Archivos

| Archivo | Cambio |
|---|---|
| `web/src/lib/geo.ts` | Catálogo nacional + helpers. |
| `web/src/components/inscripcion/InscriptionForm.tsx` | Ficha de inscripción pública: 3 selects con legacy. |
| `web/src/routes/admin/inscripciones/[id].tsx` | Ficha admin de inscripción: 3 selects con legacy. |
| `web/src/components/psi/profile/LocationSection.tsx` | Perfil psi: Municipio (Carabobo), Estado y Ciudad/Municipio → selects (`SelectField` local con estilo idéntico al `InputField`). |
| `web/src/components/admin/psicologos/edit/sections/LocationSection.tsx` | Edición psi admin: Ciudad/Municipio → select cascada (el patrón legacy de Carabobo/estado ya existía). |
| `web/src/routes/admin/notificaciones/crear/index.tsx` | Municipio/Estado → selects con "Todos". |

## Verificación

```bash
cd web && deno task build   # ✓ compila (no hay typecheck configurado)
```

El dev server sirve los cambios al instante (`vinxi dev`); la ruta
`/inscripcion` responde 200 con el módulo nuevo.