# Manual de Estilo — Colegio de Psicólogos de Carabobo

> Versión 1.0 · Aplica a la identidad visual del sitio web y materiales de comunicación institucional.

## 1. Propósito

Este manual consolida los elementos de identidad visual del Colegio de Psicólogos de Carabobo. Se basa en los recursos de diseño de la carpeta `estilos/` y en los valores ya implementados en el frontend (`web/src/app.css`). Su objetivo es garantizar un uso **consistente** de colores, tipografía y logotipo en toda la organización.

Reglas generales:

- No se recolorean, redibujan ni distorsionan los elementos de marca.
- Se respeta la proporción original de cada elemento (nunca se estiran).
- Preferir siempre los valores de la tabla de colores oficial (Sección 3).

---

## 2. Símbolo y logotipo

### 2.1 Símbolo principal

El símbolo oficial es la letra griega **Psi (Ψ)**. Representa a la Psicología y es el distintivo del Colegio. Ya se utiliza en la interfaz web como símbolo de marca y en las imágenes de error/carga (`psi.png`, `psi404.png`, `psi_loading.png`).

### 2.2 Formatos del emblema/logotipo

La carpeta `estilos/` contiene las versiones oficiales del emblema:

| Archivo | Formato | Uso recomendado |
|---|---|---|
| `Pasted image (3).png` | **Emblema cuadrado — versión light** (fondo blanco, 1254×1254) | Uso principal en fondo claro: sitios web, impresos, documentos. **Versión preferida.** |
| `Pasted image (2).png` | **Emblema cuadrado — versión dark** (fondo negro sólido, 1024×1024) | Sobre fondos oscuros: encabezados oscuros, presentaciones dark, material audiovisual. |
| `image0.png` | **Logo horizontal — versión transparente** (640×213, canal alfa) | Banners, cabeceras web, firmas, material digital donde el emblema debe integrarse al fondo. |
| `wide_high_resolution_logo_graphic_on_a_solid_blac.png` | **Logo horizontal — versión dark** (640×213, fondo negro) | Material horizontal sobre fondo negro: vídeo, portadas, social media. |
| `Pasted image.png` | **Logo horizontal — monocromo (blanco y negro)** (757×387) | Impresiones en escala de grises, fax, documentos formales restringidos a b/n. |

### 2.3 Assets web (autoalojados en `web/public/`)

Versiones optimizadas para la web, derivadas de la carpeta `estilos/`:

| Archivo | Origen | Uso |
|---|---|---|
| `web/public/emblema.png` | `Pasted image (3).png` (light, 512²) | Logo principal en navbar, home, admin, login, loading, favicon. |
| `web/public/emblema-dark.png` | `Pasted image (2).png` (dark, 512²) | Sobre fondos oscuros (según §2.4). |
| `web/public/logo-horizontal.png` | `image0.png` (transparente, 480×160) | Banners y cabeceras horizontales. |

### 2.4 Reglas de uso del logotipo

- **Área de respeto:** mantener un margen mínimo de respeto equivalente a la altura del símbolo Ψ alrededor del emblema. No colocar texto ni elementos visuales dentro de este margen.
- **Tamaño mínimo:** no reducir el emblema por debajo de ~90 px de alto (versión cuadrada) o ~40 px (versión horizontal) en pantalla.
- **Fondos:** usar la versión *light* sobre fondos claros y la versión *dark* sobre fondos oscuros. Nunca invertir la combinación de forma que el logotipo pierda contraste.
- **Prohibido:** recolorear, aplicar degradados, añadir sombras no previstas, rotar, estirar, o combinar el emblema con otros símbolos. No usar el Ψ solo como sustituto del emblema en contextos formales.

### 2.5 Tira venezolana

Elemento identitario de marca: franja horizontal tricolor rojo–verde–azul (bandera de Venezuela). Se usa en el pie del Home, directorio (FlagFooter), 404 y explorar. Colores: `bg-colpsi-red`, `bg-green-700` directo, `bg-colpsi-blue`. La animación `animate-flag-flow` recorre la franja con un brillo diagonal. No se reverte el orden de los colores ni se mezcla con otros patrones.

---

## 3. Paleta de colores

### 3.1 Colores institucionales

Valores oficiales definidos en `web/src/app.css` (`@theme`):

| Token | Nombre | Hex | Uso |
|---|---|---|---|
| `--color-colpsi-blue` | Azul institucional | `#1e3a8a` | Color primario: encabezados, botones, fondos de secciones, enlaces. |
| `--color-colpsi-blue-dark` | Azul oscuro | `#172554` | Variante profunda (diagramas del portal, fondos con más peso). |
| `--color-colpsi-blue-light` | Azul claro (hover) | `#1e40af` | Hover de botones azules primarios. |
| `--color-colpsi-navy` | Azul marino del escudo | `#0a174f` | Fondo de héroes (noticias), profundidad heráldica. |
| `--color-colpsi-navy-800` | Azul marino medio | `#0d2658` | Gradientes heráldicos. |
| `--color-colpsi-gold` | Dorado suave del escudo | `#dfcc87` | Acentos dorados sutiles, bordes finos. |
| `--color-colpsi-surface` | Superficie de cards (gris 50) | `#f9fafb` | Fondos de superficies secundarias dentro de cards (`bg-gray-50`). |
| `--color-colpsi-border` | Borde de cards (gris 100) | `#f3f4f6` | Patrón dominante `border border-colpsi-border` en cards. |
| `--color-colpsi-yellow` | Amarillo institucional | `#facc15` | Acento: CTA, badges, focos de atención, estados activos. |
| `--color-colpsi-yellow-dark` | Amarillo oscuro (hover) | `#eab308` | Hover de CTA amarillos. |
| `--color-colpsi-red` | Rojo institucional | `#991b1b` | Rojo corporativo: estados de alerta, insolvencia, errores. |
| `--color-colpsi-text` | Texto principal | `#1e293b` | Cuerpo de texto. |
| `--color-colpsi-muted` | Texto secundario | `#64748b` | Textos de apoyo, metadatos, descripciones. |
| `--color-colpsi-bg` | Fondo claro | `#f8fafc` | Fondos de página y superficies interiores. |

### 3.2 Paleta extendida (del emblema)

Variantes derivadas del emblema de la carpeta `estilos/`. Se usan para apoyos gráficos, rara vez para texto.

**Azules**

| Hex | Descripción |
|---|---|
| `#092B6F` / `#0A174F` / `#0D2658` | Azules marino del escudo (versión light). |
| `#041862` / `#052B84` / `#002675` / `#00094F` | Azules profundos del escudo (versión dark) y del banner. |
| `#022486` / `#03176A` / `#012386` | Azul vivo del banner horizontal. |

**Rojos**

| Hex | Descripción |
|---|---|
| `#98131A` / `#9B191E` / `#97121A` | Rojo del escudo (versión light). |
| `#AF0717` / `#BE1F1A` / `#650A0F` | Rojo del escudo (versión dark). |
| `#F60124` / `#F40224` / `#D8051C` / `#DE041C` | Rojo vivo del banner horizontal. |

**Dorados / Amarillos**

| Hex | Descripción |
|---|---|
| `#F9B804` / `#F8B505` | Dorado amarillo del banner. |
| `#DFCC87` | Dorado suave del escudo (versión dark). |

**Neutros de apoyo** (del escudo): `#D2D8E7`, `#A7ACC0`, `#57698F`, `#415063`.

### 3.3 Combinación y proporción

- Composición de referencia: **Azul** como color dominante (~60%), **dorado/amarillo** como acento (~30%), **rojo** como color de punto focal (~10%).
- El azul siempre es el color "oficial" del text; el rojo se reserva para alertas y estados de error.
- Sobre fondos oscuros usar preferentemente amarillo/dorado y blanco para texto; sobre fondos claros, azul institucional.

### 3.4 Contraste y accesibilidad

- Texto oficial sobre fondo claro: `--color-colpsi-text` (`#1e293b`) o `--color-colpsi-blue` (`#1e3a8a`).
- Texto sobre amarillo institucional: azul oscuro (`#172554` o `#1e3a8a`) — nunca blanco.
- Mantener al menos WCAG AA (ratio ≥ 4.5:1) para texto normal.

---

## 4. Tipografía

### 4.1 Familia principal

**Inter** (variable) — declarada en `web/src/app.css` y autoalojada en `web/public/fonts/`:

```css
--font-sans: "Inter", ui-sans-serif, system-ui, sans-serif;
```

Archivos de fuente: `inter-latin.woff2` (latin, 48 KB) y `inter-latin-ext.woff2` (latin-ext, 85 KB), ambos en formato variable (`font-weight: 100 900`). Se cargan con `font-display: swap` y preload del subset principal en `entry-server.tsx`. CSP permite `font-src 'self' data:`.

Fuentes del sistema como respaldo: `ui-sans-serif`, `system-ui`, `sans-serif`.

### 4.2 Jerarquía

| Nivel | Estilo | Notas |
|---|---|---|
| `h1` | Azul institucional, `font-bold`, `tracking-tight` | Título principal de página. |
| `h2` | Azul institucional, `font-bold`, `tracking-tight` | Secciones. |
| `h3` | Azul institucional, `font-bold`, `tracking-tight` | Subsecciones. |
| `h4` | Azul institucional, `font-bold`, `tracking-tight` | Bloques menores. |
| `body` | `--color-colpsi-text`, `line-height: 1.6`, `antialiased` | Texto base. |

### 4.3 Reglas

- Títulos con `tracking-tight` (el autoescritura ajustado da presencia tipográfica).
- Peso `bold` o `font-black` para títulos y destacados; evitar peso < 500 para texto de más de 40px.
- Mayúsculas + `tracking-widest` se reservan para etiquetas (`uppercase tracking-widest`), botones de acción y avisos (patrón del portal admin/psi).

---

## 5. Iconografía y bordes

- **Radios de esquina — rango institucional:**
  - `rounded-[2.5rem]` (40 px): tarjetas hero principales (nosotros, inscripción, explorar).
  - `rounded-[1.75rem]` (28 px): tarjetas de perfil (PsychologistCard).
  - `rounded-3xl` (24 px): tarjetas de contenido secundario, headers del portal psi.
  - `rounded-2xl` (16 px): inputs, botones CTAs, cajas de icono.
  - `rounded-xl` (12 px): botones secundarios, inputs de formulario más compactos.
  - `rounded-full`: badges, chips, tags, avatar.
- **Degradado heráldico (`bg-heraldic`)**: gradiente diagonal `#0a174f → #1e3a8a → #172554`. Se usa en los héroes de páginas públicas y del portal psi (explorar, nosotros, inscripción, documentos, noticias con `bg-colpsi-navy` puro, tickets, perfil). Aporta profundidad derivada del escudo del emblema.
- **Iconos de acceso:** cajas de icono cuadradas con `rounded-2xl`, fondo `bg-blue-50` e icono azul (patrón del portal psi). Al hacer hover el fondo pasa a amarillo institucional (`bg-colpsi-yellow`).
- **Alerta de error:** caja `bg-red-50` con borde izquierdo de 4px rojo institucional (`#991b1b`) y etiqueta `uppercase tracking-wide`; animación `animate-shake` (definida en `app.css`).
- **Scrollbar (web):** 6×6 px, track `#f8fafc`, thumb `rgba(30, 58, 138, 0.2)` y `0.4` al hover, radios 8px.

---

## 6. Sombras y profundidad

- **Sombra premium:** `--shadow-premium` (tarjetas hero de páginas públicas):

  ```css
  0 4px 6px -1px rgba(30, 58, 138, 0.05), 0 2px 4px -2px rgba(30, 58, 138, 0.05);
  ```

- **Sombra de contenido:** `shadow-sm border border-gray-100` — tratamiento estándar para tarjetas de contenido secundario (psi portal, admin StatCards, noticias, PsychologistCard en reposo).
- **Sombra de encabezado:** `shadow-inner` sobre fondos `bg-colpsi-blue` — profundidad en headers de sección (nosotros, explorar, inscripción, documentos, notificaciones).
- **Botones primarios:** `shadow-lg` / `shadow-xl` con sombra azulada suave (`shadow-yellow-500/20` sobre botones amarillos, `shadow-yellow-500/30` sobre CTAs destacados).
- **Hover lift:** `hover:shadow-2xl` para elevación al pasar el cursor sobre tarjetas.
- **Los CTA amarillos** usan `bg-colpsi-yellow` con hover `bg-colpsi-yellow-dark` (`#eab308`) y `shadow-yellow-500/20`.

---

## 7. Animaciones

Las animaciones están definidas en `web/src/app.css` y soportan `prefers-reduced-motion: reduce`.

| Clase | Keyframe | Uso |
|---|---|---|
| `animate-shake` | `shake` | Feedback de error en formularios (login, alertas de validación). Duración 0.4s ease-in-out. |
| `route-fade` | `route-fade` | Transiciones de ruta: fade-in + translateY de 4px. Duración 0.25s ease-out. |
| `animate-flag-flow` | `flag-flow` | Animación de la tira venezolana: desplazamiento horizontal continuo. 3s linear infinite. |

Además, `tw-animate-css` proporciona utilidades de animación de entrada/salida: `animate-in`, `fade-in`, `slide-in-from-top-4`, `slide-in-from-bottom-4`, etc. Se usan en resultados del directorio y paneles admin.

---

## 8. Estados y feedback visual

| Estado | Color | Ejemplo |
|---|---|---|
| Activo / solvente | Verde (`bg-green-100 text-green-600`) | Ícono de solvencia. |
| Insolvente / error | Rojo institucional `#991b1b` | Badge "INSOLVENTE", `bg-red-100 text-red-600`. |
| Enfoque / hover | Amarillo institucional `#facc15` o azul vivo | `focus:border-colpsi-yellow`, `hover:bg-colpsi-yellow`. |
| Deshabilitado | `opacity-40` / `disabled:cursor-not-allowed` | Botones bloqueados. |

Micro-interacciones estándar: `active:scale-[0.98]`, `hover:scale-[1.02]`, transición `transition-all` 150–250 ms.

---

## 9. Dos y no-haceres

**Sí**

- Usar la versión *light* del emblema sobre fondos claros y las versiones *dark*/transparente sobre fondos oscuros.
- Mantener azul como color base en encabezados y estructura; amarillo para llamados a la acción; rojo solo para alertas.
- Respetar área de respeto y tamaño mínimo del logotipo.
- Usar Inter para todo el texto.

**No**

- No usar el rojo vivo (`#F60124`, `#F40224`) en texto de una página web; limítalo a material gráfico de alto impacto.
- No mezclar las dos versiones del emblema (light + dark) en un mismo contexto de fondo.
- No aplicar el Ψ solo donde corresponda el emblema institucional completo.
- No estirar, rotar, o añadir efectos al logotipo/emblema.
- No sustituir los hex oficiales por equivalentes "aproximados".

---

## 10. Referencia rápida (web)

Fuente de verdad de los tokens en el código:

| Token CSS | Hex |
|---|---|
| `--color-colpsi-blue` | `#1e3a8a` |
| `--color-colpsi-blue-dark` | `#172554` |
| `--color-colpsi-blue-light` | `#1e40af` |
| `--color-colpsi-navy` | `#0a174f` |
| `--color-colpsi-navy-800` | `#0d2658` |
| `--color-colpsi-gold` | `#dfcc87` |
| `--color-colpsi-surface` | `#f9fafb` |
| `--color-colpsi-border` | `#f3f4f6` |
| `--color-colpsi-yellow` | `#facc15` |
| `--color-colpsi-yellow-dark` | `#eab308` |
| `--color-colpsi-red` | `#991b1b` |
| `--color-colpsi-text` | `#1e293b` |
| `--color-colpsi-muted` | `#64748b` |
| `--color-colpsi-bg` | `#f8fafc` |
| `--font-sans` | Inter, ui-sans-serif, system-ui, sans-serif |

---

*Manual generado a partir de la carpeta `estilos/` y `web/src/app.css`. Gestionado en el repositorio del proyecto.*