# Plan: Paneles de Gestión Administrativa con Visión Profesional

> Objetivo: rediseñar el área `/admin/` para que tenga aspecto de panel
> empresarial denso y sobrio (bordes finos, alta densidad, color funcional),
> eliminando el look de "cartas flotantes con sombras" heredado de landing pages.

**Stack**: SolidStart (SolidJS) + Tailwind v4 + paleta `colpsi-*` de `web/src/app.css`.
**Regla del repo**: los cambios son de presentación; lógica, RBAC, server actions
y endpoints intactos (*"no modifique el funcionamiento"*). Verificar con `npm run build`.

---

## 0. Corrección de ecosistema (libs React descartadas)

Las recomendaciones habituales de UI son **React** y no funcionan en SolidStart:

| Recomendación original | Alternativa real para este proyecto |
|---|---|
| Shadcn UI (copy-in + Radix) | **Kobalte** (`@kobalte/core`): primitivas Radix para Solid, o copiar estructuras/SVGs a mano |
| Tremor (dashboards) | No existe versión Solid → construir bloques KPI/gráficos con Tailwind (ya existe `Sparkline`) |
| Ant Design / Mantine | Incompatibles con SolidJS. Descartados |
| Estética plana y corporativa | Se logra 100% con Tailwind + tokens `colpsi-*` |

---

## 1. Diagnóstico del admin actual (fuentes del look "landing")

Archivos: `web/src/routes/admin.tsx`, `web/src/components/admin/dashboard/StatCard.tsx`,
`web/src/routes/admin/index.tsx`.

1. **`StatCard`** → `rounded-2xl p-5 shadow-sm border-l-4` + `gap-4` entre tarjetas → cajas flotando.
2. **Sidebar** → `shadow-2xl`, items `rounded-xl`, activo `bg-colpsi-yellow ... shadow-lg` → píldora brillante con sombra.
3. **Tipografía de marketing** → `font-black uppercase tracking-widest` por todos lados; `text-2xl font-black` en cada título.
4. **Iconos emoji** (`📊👥📝`) → colores arbitrarios que rompen la paleta funcional.
5. **Títulos de sección** `text-gray-400` (casi invisibles) frente a números `text-3xl font-black`.
6. **Sin topbar**: no hay breadcrumbs, barra de herramientas ni cabecera de tabla → todo flota en `space-y-8`.

---

## 2. Sistema de diseño "panel profesional"

### 2.1 Escala visual

| Elementa | Regla | Ejemplo |
|---|---|---|
| Radio | `rounded-md` máximo (`rounded-lg` solo en paneles externos) | `rounded-md` |
| Sombras | **Cero** en contenido; solo `shadow-sm` en menús/dropdowns | `shadow-sm` |
| Bordes | 1px: `colpsi-border` (paneles), `slate-200`/`gray-200` (tablas) | `border` |
| Gap entre paneles | `gap-3` / `gap-4` (no `gap-6`+) | `grid gap-4` |
| Padding | Paneles `p-4`, celdas `py-2 px-3` | `p-4` |
| Datos | `text-sm` (14px) o `text-[13px]`, siempre `tabular-nums` | |
| Título de página | `text-lg font-semibold` (no `font-black`, no `text-2xl`) | |
| Etiqueta / label KPI | `text-[11px] font-semibold uppercase tracking-wide text-colpsi-muted` | |
| Números KPI | `text-2xl font-semibold tabular-nums` (nada de `font-black`) | |

### 2.2 Color funcional (paleta existente)

- **Estructura**: neutros (`colpsi-bg`, `colpsi-border`, `colpsi-surface`, `colpsi-text`, `colpsi-muted`).
- **Marca/interacción**: `colpsi-blue` (+ `blue-light` hover); `colpsi-yellow` solo como acento/indicador.
- **Semánticos** (solo estados del negocio): verde = activo/completado, ámbar =
  pendiente/procesando, rojo (`colpsi-red`/`red-*`) = error/cancelado, azul = informativo.
- **Nada de degradados ni `bg-heraldic` dentro del admin** (son de páginas públicas).

### 2.3 Do / Don't

```
✅ border border-colpsi-border rounded-md bg-white    ❌ rounded-2xl shadow-xl shadow-sm
✅ text-sm font-semibold text-colpsi-text             ❌ text-xl font-black uppercase tracking-widest
✅ divide-y divide-colpsi-border (listas densas)      ❌ space-y-6 entre tarjetitas
✅ iconos SVG monocromos 16px (currentColor)          ❌ emojis 📊👥 como iconografía
✅ badge semántico de 20px                            ❌ acentos border-l-4 de colores por tarjeta
✅ números tabular-nums alineados a la derecha        ❌ números centrados en cards
```

---

## 3. Arquitectura de layout (3 zonas)

Replantear `web/src/routes/admin.tsx` manteniendo la lógica intacta (RBAC,
`menuItems`, badge de tickets, colapso).

```tsx
<div class="flex h-screen overflow-hidden bg-white font-sans text-colpsi-text">

  {/* ZONA 1 — SIDEBAR fijo 256px, plano */}
  <aside class="w-64 shrink-0 border-r border-colpsi-border bg-colpsi-bg flex flex-col">
    {/* logo: h-14, sin emoji, texto-sm font-semibold */}
    <nav class="flex-1 overflow-y-auto p-2 space-y-0.5">
      {/* item activo: fondo blanco + texto azul + indicador amarillo de 2px (sin sombra) */}
    </nav>
    {/* perfil/salir: border-t, botón ghost */}
  </aside>

  {/* ZONA 2+3 — columna de trabajo */}
  <div class="flex-1 flex flex-col min-w-0">
    {/* TOPBAR fijo 56px */}
    <header class="h-14 shrink-0 border-b border-colpsi-border bg-white flex items-center gap-4 px-4">
      {/* breadcrumbs · buscador · perfil */}
    </header>

    {/* ÁREA DE TRABAJO */}
    <main class="flex-1 overflow-y-auto bg-colpsi-bg">
      <div class="mx-auto w-full max-w-[1400px] px-6 py-5 space-y-5">
        {props.children}
      </div>
    </main>
  </div>
</div>
```

**Item del sidebar (sin píldora amarilla brillante):**

```tsx
<A
  href={item.path}
  end={item.path === "/admin"}
  class="flex items-center gap-3 h-9 px-3 rounded-md border-l-2 border-transparent
         text-sm font-medium text-slate-600 hover:bg-white hover:text-colpsi-blue transition-colors"
  activeClass="bg-white border-l-colpsi-yellow text-colpsi-blue font-semibold"
>
  <span class="w-4 h-4 shrink-0">{/* SVG 16px, stroke currentColor */}</span>
  {!isCollapsed() && <span class="truncate">{item.title}</span>}
  {/* badge de tickets: min-w-4 h-4 text-[10px] sin shadow */}
</A>
```

> Si `activeClass` con `border-transparent`/`border-l-colpsi-yellow` genera conflicto
> de orden en Tailwind v4, usar solo `bg-white text-colpsi-blue font-semibold`.

**Breadcrumbs en el topbar:**

```tsx
<nav class="text-xs text-colpsi-muted">
  Admin <span class="mx-1 text-slate-300">/</span>
  <span class="text-colpsi-text font-medium">Psicólogos</span>
</nav>
```

---

## 4. Componentes clave

### 4.1 Bloque de KPIs integrado (reemplaza a 4× `StatCard`)

Un solo contenedor dividido por líneas, en lugar de tarjetas sueltas:

```tsx
<section class="border border-colpsi-border rounded-lg bg-white overflow-hidden">
  {/* Truco robusto: contenedor con border-t+border-l, celdas con border-b+border-r
      → líneas uniformes aunque el grid cambie de columnas */}
  <div class="grid grid-cols-2 lg:grid-cols-4 border-t border-l border-colpsi-border">
    <div class="border-b border-r border-colpsi-border p-4">
      <p class="text-[11px] font-semibold uppercase tracking-wide text-colpsi-muted">Logins hoy</p>
      <p class="mt-1 text-2xl font-semibold tabular-nums text-colpsi-text">1.205</p>
      <p class="mt-0.5 text-xs text-colpsi-muted">Únicos: 340</p>
    </div>
    <div class="border-b border-r border-colpsi-border p-4">…</div>
    {/* … */}
  </div>
</section>
```

- Iconos/emojis dentro del KPI: **fuera**. Señal opcional: punto
  `w-1.5 h-1.5 rounded-full` del color semántico.
- `accent="border-colpsi-yellow"` por tarjeta: **fuera** (el acento es para acción).

### 4.2 Tabla de datos (patrón único para listas)

```tsx
<div class="border border-colpsi-border rounded-lg bg-white overflow-hidden">
  {/* TOOLBAR */}
  <div class="flex items-center gap-2 p-3 border-b border-colpsi-border">
    <input class="h-9 w-64 rounded-md border border-slate-300 px-3 text-sm
                   focus:border-colpsi-blue focus:ring-2 focus:ring-colpsi-blue/15 outline-none"
           placeholder="Buscar psicólogo…" />
    <div class="ml-auto flex gap-2">
      <button class="h-9 px-3.5 rounded-md border border-colpsi-border bg-white text-sm
                     font-medium hover:bg-colpsi-bg">Filtros</button>
      <button class="h-9 px-3.5 rounded-md bg-colpsi-blue text-white text-sm
                     font-semibold hover:bg-colpsi-blue-light">Nuevo</button>
    </div>
  </div>

  <div class="overflow-x-auto">
    <table class="w-full text-sm border-collapse">
      <thead class="sticky top-0 z-10">
        <tr>
          <th class="px-3 py-2.5 text-[11px] font-semibold uppercase tracking-wide
                     text-colpsi-muted bg-colpsi-bg border-b border-colpsi-border text-left">Nombre</th>
          <th class="px-3 py-2.5 text-[11px] font-semibold uppercase tracking-wide
                     text-colpsi-muted bg-colpsi-bg border-b border-colpsi-border text-right">Inscripción</th>
          <th class="px-3 py-2.5 text-[11px] font-semibold uppercase tracking-wide
                     text-colpsi-muted bg-colpsi-bg border-b border-colpsi-border text-left">Estado</th>
          <th class="w-10 bg-colpsi-bg border-b border-colpsi-border"></th>
        </tr>
      </thead>
      <tbody class="divide-y divide-colpsi-border">
        <tr class="group hover:bg-colpsi-bg/60">
          <td class="py-2 px-3 font-medium">Ana Pérez</td>
          <td class="py-2 px-3 text-right tabular-nums">45.230,00</td>
          <td class="py-2 px-3"><Badge tone="success">Activo</Badge></td>
          <td class="py-2 px-3 text-right">
            <button class="opacity-0 group-hover:opacity-100 focus:opacity-100
                           h-7 w-7 rounded-md hover:bg-white text-colpsi-muted">⋯</button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</div>
```

Reglas de tabla:

- **Sticky header** con fondo contrastado (`bg-colpsi-bg`) y `text-[11px] uppercase`.
- **Texto a la izquierda, números/dinero a la derecha** + `tabular-nums` siempre.
- **Acciones por fila**: menú `⋯` que aparece en `hover`/`focus` (no botones permanentes).
- Cabecera de página **fuera** de la tabla, con `border-b` y acciones a la derecha.

> Evitar `[&>th]:text-left` + `text-right` por celda: en Tailwind v4 el arbitrary
> variant gana por especificidad y la alineación no aplica. Escribir la clase base
> por `th` o definir un utility `@utility th-cell` **sin** alineación.

### 4.3 Badge semántico (micro-etiqueta de estado)

```tsx
const tones = {
  success: "bg-emerald-50 text-emerald-700 border-emerald-200",
  warning: "bg-amber-50 text-amber-700 border-amber-200",
  danger:  "bg-red-50 text-red-700 border-red-200",
  info:    "bg-blue-50 text-blue-700 border-blue-200",
  neutral: "bg-slate-100 text-slate-600 border-slate-200",
};

<span class={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded border text-xs font-medium ${tones[props.tone]}`}>
  <span class="w-1.5 h-1.5 rounded-full bg-current" />
  {props.children}
</span>
```

### 4.4 Botones e inputs (set único en todo el admin)

```tsx
btnPrimary   = "h-9 px-3.5 rounded-md bg-colpsi-blue text-white text-sm font-semibold
                hover:bg-colpsi-blue-light focus:ring-2 focus:ring-colpsi-blue/30 outline-none"
btnSecondary = "h-9 px-3.5 rounded-md border border-colpsi-border bg-white text-sm font-medium
                text-colpsi-text hover:bg-colpsi-bg"
btnGhost     = "h-9 px-2.5 rounded-md text-sm font-medium text-colpsi-muted
                hover:text-colpsi-blue hover:bg-colpsi-bg"
btnDanger    = "h-9 px-3.5 rounded-md bg-colpsi-red text-white text-sm font-semibold hover:opacity-90"

input        = "h-9 w-full rounded-md border border-slate-300 bg-white px-3 text-sm
                placeholder:text-slate-400 focus:border-colpsi-blue focus:ring-2
                focus:ring-colpsi-blue/15 outline-none"
```

Nota: los inputs actuales de login usan `border-2 border-transparent +
focus:border-colpsi-yellow + rounded-xl`. Para el admin, borde visible + focus
azul es más estándar (MANUAL_ESTILO §3.4 permite foco "amarillo **o azul vivo**").

### 4.5 Iconografía

Copiar paths SVG de Lucide (16px, `stroke-width={2}`, `fill="none"`,
`stroke="currentColor"`) a `web/src/components/admin/icons.tsx`. Así el color
hereda del texto (activo = azul, inactivo = gris) y desaparecen los colores
arbitrarios de los emojis. Cero dependencias.

---

## 5. Escala tipográfica y densidad

| Uso | Antes (actual) | Ahora |
|---|---|---|
| Título de página | `text-2xl font-black text-colpsi-blue` | `text-lg font-semibold text-colpsi-text` |
| Etiqueta de sección | `text-xs font-black uppercase tracking-widest text-gray-400` | `text-[11px] font-semibold uppercase tracking-wide text-colpsi-muted` |
| Cuerpo/datos | `text-sm`/`text-base` | `text-sm` (14px), `text-[13px]` en tablas densas |
| Ritmo vertical | `space-y-8` | `space-y-5` (o `space-y-4` en listas) |
| Subtítulo meta | `text-gray-300` (ilegible) | `text-colpsi-muted` |

---

## 6. Fases de migración (orden seguro, sin romper nada)

Cada fase termina con `npm run build` ✅. Son cambios **solo de presentación**:
lógica, RBAC, server actions y endpoints intactos.

1. **Fase 1 — Primitivas nuevas** en `web/src/components/admin/ui/`:
   `Panel`, `KpiStrip`, `Badge`, `Button`, `Input`, `PageHeader`
   (+ `@utility th-cell` / `td-cell` en `web/src/app.css`). Aditivo; nada existente cambia.
2. **Fase 2 — Shell** (`web/src/routes/admin.tsx`): sidebar plano (quitar
   `shadow-2xl`, `shadow-lg`, `rounded-xl`→`rounded-md`, píldora amarilla → indicador
   sutil), nuevo topbar `h-14` con breadcrumbs + zona de perfil. Conservar
   `menuItems`, `visibleMenu`, colapso y badge de tickets.
3. **Fase 3 — Dashboard** (`web/src/routes/admin/index.tsx`): `StatCard` →
   `KpiStrip` integrado, títulos apagados, `Sparkline` dentro del panel de tendencias
   (mismo borde que el resto), banners (`ActiveSessionsBanner`, `BirthdayBanner`,
   `ReceptionSwitchesCard`) pasan de `rounded-2xl` a panel `rounded-md border`.
4. **Fase 4 — Listas**: `psicologos`, `noticias`, `inscripciones`, `staff`,
   `tickets` → patrón tabla con sticky header + toolbar + menú `⋯`. Reestilizar
   `AdminLoadingSkeleton` a la misma densidad (skeleton de filas, no de tarjetas).
5. **Fase 5 — Formularios/CRUD** al set de inputs/botones unificado;
   modales `rounded-lg shadow-sm`.
6. **Fase 6 (opcional)** — buscador `Ctrl+K` en el topbar
   (referencia: `web/src/components/layaout/SearchBar.tsx`).

**Prioridad**: Fases 1 + 2 primero (el shell es el 70% del "look profesional").

---

## 7. Verificación

1. `npm run build` compila (Node) — sin typecheck configurado, esto es la verificación.
2. Docker: `docker compose build web && docker compose up -d`.
3. Smoke test visual: `/admin-access` → login → dashboard, listas y CRUD del admin.
4. Confirmar que el menú filtrado por permisos (RBAC) y el badge de tickets siguen funcionando.

---

## 8. Entregado — Notebook estilo Odoo (rama `feat/admin-paneles-pro`)

Pestañas horizontales tipo Odoo (notebook) implementadas y verificadas con
`npm run build` en tres rutas: `/psi/perfil`, `/admin/psicologos/[id]/detalle`
e `/admin/inscripciones/[id]`.

### Componente compartido

- **`web/src/components/ui/Notebook.tsx`** — barra de pestañas + paneles con
  *lazy-mount* SSR-seguro (`NotebookContext` + `NotebookPage`): la pestaña
  inicial se hidrata desde el SSR; las demás se montan al activarse la primera
  vez y luego permanecen montadas (ocultas) para conservar estado local
  (TipTap, flatpickr, selecciones temporales).
- **Ciclo de 6 colores intercalados por posición**: azul `#1e3a8a`, amarillo
  fuerte (título oscurecido `#a16207`), navy `#0a174f`, vinotinto `#722f37`,
  verde `#166534`, azul-claro `#1e40af`. El título conserva **siempre** su color
  (regla del usuario); la selección se marca con el **degradado azul heráldico**
  (navy → azul → azul-oscuro) + título en píldora clara de su color.
- **`shrink-0`** en cada pestaña: los títulos largos no se cortan
  (`overflow-hidden`); la fila desplaza con `overflow-x-auto`.
- **Subtítulo bajo la barra**: punto del color de la pestaña + nombre de la
  sección activa, en el color del título de la pestaña; y **cuadro explicativo**
  opcional (`description` por página) que describe para qué sirve cada área.

### Rutas migradas

| Ruta | Notebooks / pestañas |
|---|---|
| `/psi/perfil` | 7 pestañas con iconos (user, mail, book, mapPin, fileText, sliders, shield) y descripciones. **Redes Sociales vive aparte** (tarjeta independiente, guardado propio) |
| `/admin/psicologos/[id]/detalle` | 2 notebooks: 8 (expediente, dentro del form) + 5 (gestión, fuera del form) |
| `/admin/inscripciones/[id]` | 4 pestañas |

La lógica, los RBAC y las server actions quedaron **intactos** (cambios solo de
presentación). Notas: `Panel.tsx` (primitiva anterior) se eliminó por ser dead
code; los commits viven en `feat/admin-paneles-pro` y se fusionaron a `main`.

### Refinamientos del perfil psi (tras la fusión 1)

- **Redes Sociales fuera del notebook**: se guarda con lógica propia
  (`apiPost`/`apiDelete` a `/psi/me/social`) y con su propio `<form>`, así que
  salió del notebook y del `<form>` del perfil (evita el form anidado, HTML
  inválido). Tarjeta independiente debajo del notebook, con el mismo lenguaje
  visual (subtítulo con punto + cuadro explicativo, color amarillo de su
  antigua posición en el ciclo).
- **Sección de guardado como tarjeta** (`SaveButton.tsx`): del **mismo ancho
  del notebook** (caja blanca `rounded-lg border`, sin sticky, fija justo
  debajo de las pestañas). Incluye el campo **"Contraseña Actual (obligatoria
  para guardar)"** — antes vivía en la pestaña Cuenta y Seguridad
  (`SecuritySection`, eliminado por dead code) — junto al botón "Guardar
  cambios" (check/spinner) y el feedback de éxito/error debajo. Misma lógica:
  `required` + chequeo de `handleSaveProfile` intactos.

### Redes sociales en el perfil público del directorio (tras la fusión 2)

El perfil público (`/directorio/[slug].tsx` + `ProfileHeader.tsx`) mostraba las
redes como un chip de texto pequeño al final de la tarjeta "Contacto", sin
encabezado, fácil de pasar por alto. Cambios (solo presentación; la API ya
devolvía `social_networks` en `GET /psi/:fpv`):

- **Redes debajo del QR**: `ProfileHeader` recibe `socialNetworks` y las pinta
  en la tarjeta izquierda, justo debajo de la "Ficha Digital" (QR), **sin
  título ni encabezado**: chips `🔗/icono + nombre` centrados con
  `target="_blank" rel="noopener noreferrer"`, visibles solo si hay redes.
- **`ContactCard` sin redes**: se eliminó el mini-bloque del final de la tarjeta
  y la prop `socialNetworks` (única llamada: `/directorio/[slug]`).
- **Iconos de marca** (`web/src/components/psi/profile/SocialBrandIcon.tsx`):
  detecta la red por nombre normalizado (minúsculas, sin acentos) y renderiza
  su SVG oficial de **Simple Icons** para Instagram, Facebook, LinkedIn,
  X/Twitter, WhatsApp, YouTube, TikTok, Telegram, Threads, Pinterest y GitHub.
  Monocromo (`fill="currentColor"` → hereda el azul del chip); aliases cortos
  (`ig`, `x`, `fb`, `wa`, `yt`) solo matchean por igualdad exacta; redes no
  listadas caen al icono genérico de enlace. Componente reutilizable para el
  portal `/psi/perfil` si se desea.

Verificado con `npm run build` y `curl` del SSR (svg del path de Instagram
presente bajo el QR). Commits: `057e61f`, `ee2bf14`, `4f08814`.

### Perfil público del directorio — orden, teléfonos y menú de correo (tras la fusión 3)

Ajustes de UX en `/directorio/[slug]` + `ContactCard` (solo presentación;
datos, API y lógica intactos):

- **Orden de la columna derecha**: el bloque **"Perfil Profesional"** pasa
  antes de la tarjeta **"Contacto"** (carnet | perfil → contacto → formación),
  para presentar primero al profesional.
- **Teléfonos sin enlace**: el teléfono principal y los teléfonos/celulares de
  cada ubicación ya **no** son enlaces `tel:` — quedan como texto plano con el
  mismo estilo de chip (se quitó el `href="tel:..."` de la `ContactCard`).
- **Menú de acciones del correo** (`ContactCard`): el correo conserva
  `mailto:` como acción principal (gestor predeterminado del dispositivo) y
  agrega un botón `▾` con menú desplegable:
  - **"Abrir en Gmail"** → compose web (`mail.google.com/mail/?view=cm&fs=1`)
    en nueva pestaña con `rel="noopener noreferrer"`. Oferta, no imposición:
    Gmail está bloqueado por ISP (CANTV) en Venezuela, por lo que **no** es la
    única opción.
  - **"Copiar correo"** → Clipboard API con fallback `execCommand` y
    confirmación visual "¡Correo copiado!" (~2.5 s).
  - El menú cierra al hacer clic fuera o al seleccionar; `ref` en el contenedor
    completo (botón + menú) para no cerrar en el clic del botón; accesible con
    `role="menu"`/`aria-haspopup`/`focus-visible`. Solo cliente y SSR-seguro.

Verificado con `npm run build` y `curl` del SSR: `mailto:` presente, **0**
`href="tel:"` y orden perfil < contacto < formación. Commits: `f5adfd7`,
`29bedf2`.
