# Términos y Condiciones (`feat/terminos-y-condiciones`)

Dos documentos legales, **independientes entre sí**: uno público que ve
cualquiera y otro privado que solo ve un agremiado con sesión. La regla que
gobierna el diseño de este archivo es una sola, y está en la sección 1.

---

## 1. La regla que gobierna todo: el anónimo no sabe que hay un segundo texto

> **Quien no ha iniciado sesión no debe poder inferir que existe otro
> documento. Ni por un enlace, ni por un rótulo, ni por una frase, ni por la
> URL.**

El requisito seesaurus así, y se refinó: no basta con quitar el enlace. Si el
público ve "Parte I" en un badge, o una frase que diga "estas condiciones se
suman a las anteriores", ya sabe que hay un conjunto anterior. Cualquier
superficie que lo insinúe es un defecto, no un detalle de redacción.

| Superficie | Decisión | Por qué |
|---|---|---|
| Ruta del público | `/terminos` | El nombre más plano posible |
| Ruta del agremiado | `/psi/terminos` | Ya lives dentro del portal: la URL delata la existencia del portal, **no** la de un documento |
| Índice de partes | **no existe** | Un índice obliga a nombrar las partes. `/terminos` ES el documento |
| Rótulos | "Condiciones de uso del sitio" / "Condiciones del portal" | Describen, no numeran. Nada de "Parte I" |
| Enlace cruzado | solo con `sessionReady() && isAuthenticated()` | Para el anónimo no se renderiza **nada**: ni el botón, ni una versión apagada |
| `titulo` y `resumen` del público | sin "plataforma", sin "cuenta" | Un título que reparte el alcance entre el sitio y la plataforma dice que la otra mitad está en otro sitio |
| Mención a las otras condiciones | "las condiciones de uso del sitio" | Reemplaza a "la Parte I" en todas las remisiones |

**Ocultar es seguro; redirigir no lo es.** Por eso el enlace se condiciona a
`isAuthenticated()` y el guard de `/psi/terminos` exige además `sessionReady()`
(ver sección 5).

## 2. Sin CRUD en el panel, y por qué

El texto es estático: vive en módulos tipados bajo `web/src/lib/terminos/`, con
el mismo patrón que `web/src/lib/documentos/`. **No hay pantallas de admin para
editarlo**, y es deliberado: un instrumento jurídico versionado y editable desde
un formulario invita a publicarlo a medio escribir, y un cambio de texto tiene
que pasar por revisión legal, no por un campo de texto.

```
web/src/lib/terminos/
├── colegio.ts              datos institucionales (RIF, dirección, correos, URL)
├── index.ts                tipos, agrupado por nivel, catálogo
├── parte-i-publica.ts      15 secciones  ← el documento público
└── parte-ii-agremiado.ts   14 secciones  ← el del portal
```

### 2.1 `colegio.ts` es la única fuente de los datos del Colegio

Todo dato institucional se interpola desde ahí, nunca se escribe a mano:

| Constante | Valor | Nota |
|---|---|---|
| `COLEGIO_NOMBRE` | Colegio de Psicólogos del Estado Carabobo | |
| `COLEGIO_RIF` | `J-508172418` | |
| `COLEGIO_DIRECCION` | Edificio Profesional y Comercial, Oficentro 108, p4, of. D, Av. 101 Díaz Moreno, Valencia | Sin abreviaturas: en un documento legal se escriben completas |
| `COLEGIO_CORREO` | `admon.colpsicarabobo@gmail.com` | También privacidad y reportes, declarado aparte |
| `COLEGIO_TELEFONO` | **`null`** | El Colegio no publica ninguno: el dato no existe y los documentos no lo inventan |
| `COLEGIO_URL` | `https://colegio-psicologos-carabobo.com` | `http://` rompería el login (cookie `secure`) |
| `FPV_URL` | `https://fpv.org.ve/` | |

`COLEGIO_TELEFONO` es `string | null` y no cadena vacía a propósito: un
documento que lo interpole debe mostrar una **omisión consciente**, no un hueco
invisible. Si algún día se publica un número, se cambia a `null` → `"+58…"` y
los documentos que hoy lo omiten lo empiezan a mostrar solos.

### 2.2 La versión la declara el backend

`TERMINOS_VERSION` **no** es la fuente de verdad: lo es `domain.TermsVersion`
en Go, y el frontend lo recibe por `GET /psi/me/terms`. La copia de `colegio.ts`
solo se usa como arranque si el endpoint falla y para pintar la fecha de la
cabecera. Si quedan desalineadas, el síntoma es visible: el texto muestra una
fecha y el modal sigue pidiendo aceptar otra.

## 3. Aceptación: el backend registra, el frontend bloquea

| Capa | Qué hace |
|---|---|
| `psi_terms_acceptance` | Tabla **append-only**. Índice único en `(psi_user_id, version)` cierra la carrera del doble clic |
| `POST /psi/me/terms` | **Idempotente**. Ya aceptó esa versión → 200, no otra fila |
| Versión equivocada | **409** con `current_version` en el cuerpo; el modal recarga y explica |
| `AceptarTerminosModal.tsx` | Bloqueo **en cliente**, nunca en servidor |
| Auditoría | Evento `terminos` / `accept` |

**El backend nunca bloquea.** Si `GET /psi/me/terms` falla, el portal funciona
completo y no se muestra ningún modal: un fallo de red no puede dejar a un
colegiado sin poder trabajar ni presentar una solicitud.

## 4. El §10.1 obliga al código, no al revés

El borrador del Colegio afirmaba que la plataforma «puede registrar visitas al
sitio, visitantes únicos, búsquedas y vistas de fichas, y estas cifras se
utilizan de forma agregada». Al verificarlo contra el código, **afirmaba cosas
que no eran ciertas y callaba las que sí**:

| Afirmación del borrador | Realidad antes del cambio |
|---|---|
| «registrar visitas al sitio» | Cierto, pero por un camino que no se buscaba: `RecordPageView` no tiene llamadores (es un señuelo) y el que escribe es `TrackPageView`, invocado por `AnalyticsMiddleware` desde `router.go:29` |
| «cantidad de visitantes únicos» | Cierto: el mismo middleware pone la cookie `_sid` |
| «de forma agregada» | **Falso por omisión**: se guardaba la IP **en claro**, sin hashear, y **no existía retención** |

Lo que de verdad se guardaba, y no se decía: el **texto de la búsqueda tal
cual** (255 caracteres, el panel de métricas lo muestra como ranking), la **IP
en claro** en `page_views`, `search_events` y `profile_views`, y la **cabecera
`Referer` completa** — que al llegar desde un buscador trae esa URL entera con
sus parámetros, o sea datos de terceros. Y `PurgeOldData` existía, tenía test
unitario, y **nadie lo invocaba**: las IPs se acumulaban para siempre.

Decisión del Colegio: **corregir el código primero**. El texto legal no puede
describir un estado que el código no tiene.

### 4.1 Los cuatro cambios (`api/internal/service/analytics_privacy.go`)

| # | Cambio | Dónde |
|---|---|---|
| 1 | La IP se sustituye por una **huella** SHA-256 con sal del servidor | En los **métodos del servicio** (`TrackPageView`, `RecordSearch`, `RecordProfileView`, `RecordPageView`), no en los puntos de llamada: todo llamador futuro hereda la garantía |
| 2 | El `Referer` se reduce al **origen** (`esquema://host`) | `refererOrigin()`, en el mismo archivo |
| 3 | La cookie `_sid` baja de **365 a 30 días** | `middleware/analytics.go` |
| 4 | `PurgeOldData` **se programa**: purga al arrancar y cada 24 h | `cmd/api/main.go`, patrón de la bitácora |

**El texto de la búsqueda se conserva** y el §10.1 lo dice: es el dato que el
Colegio consulta en el panel, y omitirlo sería la misma falta de transparencia
que se iba a corregir.

**Huella no es anonimización.** Una IPv4 tiene 2³² valores, así que la huella
es reversible por fuerza bruta si alguien obtiene la base. Por eso:

- El texto legal dice **«huella criptográfica»**, nunca «anónimo».
- La retención de 90 días es la que acota de verdad el daño.
- `login_events` y `active_sessions` **no se tocan**: son bitácora de seguridad
  de cuentas con sesión y el panel los usa. Se declaran aparte en el §10.3.

`ANALYTICS_IP_SALT` debe ser **estable entre reinicios**: si cambia, la misma
IP produce otra huella y los conteos de visitantes únicos se rompen. Vacío usa
una constante por defecto, con el mismo fin.

## 5. `sessionReady()`: por qué el guard necesita un segundo paso

Ningún `/psi/*` tiene guard de autenticación en cliente — tampoco
`psi/manual.tsx`; el portal se apoya en los 401 de la API. El guard de
`/psi/terminos` es la excepción deliberada (el requisito del punto 1 lo
obliga), y por eso necesita `sessionReady()`:

`user()` es `null` durante el SSR y el primer pintado del cliente aunque haya
sesión, porque `await restoreSession()` es asíncrono (`lib/actions/session.ts`
devuelve el token de la cookie HttpOnly). Un guard que mirara solo
`isAuthenticated()` expulsaría al agremiado legítimo en el caso de pestaña
nueva o tras reiniciar el navegador.

| Situación | Se usa |
|---|---|
| Ocultar el enlace cruzado | `isAuthenticated()` |
| **Redirigir** fuera de `/psi/terminos` | `sessionReady() && isAuthenticated()` |

`sessionReady()` se pone en un `try/finally` que envuelve el cuerpo de
`onMount`: tiene cuatro `return`Tempranos y sin el `finally` el signal nunca
se levantaría en esos caminos.

## 6. Sin `innerHTML`, y el agrupado por nivel

`TerminosLayout.tsx` construye nodos, no HTML: aunque mañana alguien edite una
coma y el texto llegue con `<script>`, sería texto y nada más. Es el mismo
criterio que `lib/sanitize-html.ts` aplica en el resto del proyecto. El
subconjunto de Markdown en línea es deliberadamente corto: `**negrita**`,
`*cursiva*`, `` `código` `` y `[enlace](https://…)` — los enlaces solo con
`http(s)`, que bloquea la única vía de XSS que tendría sentido aquí.

Dos arreglos que solo se hicieron necesarios **al pegar el texto de verdad**:

- **El articulado era plano.** Con contenido real son 24 secciones, todas como
  tarjetas iguales: "5.1 Origen de la información" se veía igual que "5. El
  directorio". En un documento legal la jerarquía *es* el sentido. Ahora las
  subsecciones cuelgan de su padre con guía lateral y título menor, y el
  sumario lista solo los 15 niveles principales. El nivel se deduce del propio
  número (`agruparSecciones()` en `index.ts`), así que no hay que marcar nada.
- **`ordenada` se ignoraba.** El tipo `TerminosBloque` declaraba
  `ordenada?: boolean` y el renderizador no lo leía: las listas numeradas del §7
  salían con viñetas. En un articulado numerado eso rompe las referencias
  cruzadas del tipo "lo previsto en el numeral 2".

## 7. Lo que se corrigió del texto, y lo que quedó por confirmar

### Correcciones de fondo (el borrador afirmaba cosas distintas al código)

| Sección | El borrador decía | reality |
|---|---|---|
| §6.3 (portal) | «restringir la visibilidad de su ficha **o de parte de sus datos** (por ejemplo, sus áreas)» | El perfil del insolvente **colapsa a identidad**: quedan nombres, FPV, cédula, género, foto y universidad. Se ocultan contacto, ubicación, modalidad, presentación, áreas, posgrados y redes |
| §10.1 (público) | «de forma agregada» | Ver sección 4 de este documento |
| §2.2 (portal) | Política de contraseña completa | Se publica **como norma exigible** (decisión del Colegio). Ver la nota de abajo |
| §12 (portal) | Nota técnica «hoy el portal no permite vaciar un campo de texto» | Se quita la nota, se conserva la frase verdadera: para dejar en blanco un campo ya guardado hay que pedirlo al Colegio |

### Se quitó el andamiaje del borrador

El bloque «Borrador para revisión legal» (que además decía «este documento
tiene dos partes», la peor fuga posible), la línea final en cursiva «debe ser
revisado por asesoría jurídica», todas las barras `---` y cada marcador
`[CONFIRMAR]`, `[REVISIÓN LEGAL]`, `[NOTA TÉCNICA]` y `[RECOMENDACIÓN TÉCNICA]`.

### Decisiones sobre los huecos que quedaban

- **El plazo se dejó ambiguo a propósito** (§8 y §10.4): el Colegio responde
  por el correo o el teléfono que la persona facilitó, «dentro de un plazo
  razonable», sin cifra. La decisión fue del Colegio: el gremio no se
  compromete a un número.
- **§8 decía «[CONFIRMAR: cómo y en qué plazo se comunica el resultado]»**, y
  la verdad era que **no se avisa a quien solicita**: solo el personal ve la
  ficha. Ahora se dice lo que sí pasa — el Colegio comunica el resultado por el
  correo o el teléfono del formulario, dentro de un plazo razonable.
- **Se añadió una aclaración en §12** que el borrador no tenía: qué significa
  «retirar» datos en la práctica. Sin ella, el Colegio promete algo que su
  propia normativa gremial le impide cumplir (retirar = **dejar de aparecer en
  el directorio**, no borrar el expediente).

### Lo que sigue pendiente de la asesoría legal

- **§8.1.4 del portal** (la notificación se tiene por *puesta en su
  conocimiento*): se publicó tal como lo redactó el Colegio, sin verificación
  técnica de su alcance. Era el único `[REVISIÓN LEGAL]` sin resolver.
- **La política de contraseña (§2.2) se publica como norma, pero el backend no
  la exige al agremiado**: `isStrongPassword` solo está conectada al alta que
  hace el personal (`admin/psicologos/crear.tsx`), y el cambio de contraseña
  del portal y el reset solo comprueban longitud ≥ 8
  (`psi_service_password_reset.go:90`). Es un agujero del backend, no del
  documento.
- **`[PLAZO]`** sigue sin cifra por decisión del Colegio.

## 8. Gotchas de implementación

| # | Gotcha |
|---|---|
| 1 | **`routes/terminos.tsx` junto a la carpeta `terminos/` es un LAYOUT de ella**, igual que `psi.tsx` envuelve `psi/`. Eso hizo que `/terminos` diera 404. El índice va **dentro** de la carpeta. No lo muevas de vuelta |
| 2 | La URL vieja **`/terminos/publica` es una redirección en cliente** (`navigate(..., { replace: true })`), no un 301: nunca estuvo desplegada, así que no hay equity ni SEO que preservar. Mismo patrón que `reset-password.tsx` |
| 3 | Cada documento es **canónico de su ruta**. El mapeo va explícito en `TerminosLayout.tsx` porque invertirlo en silencio haría que Google indexara cada página bajo la URL de la otra |
| 4 | El botón de volver es **contextual**: el público sale del sitio («Volver al inicio» → `/explorar`); el del portal vuelve al otro texto |
| 5 | El modal se **suprime en `/psi/terminos`**: el agremiado ya está leyendo el texto, no puede aceptar desde dentro de él |

## 9. Verificación

25 comprobaciones E2E con Chromium headless sobre CDP, con sesión forjada (JWT
con `exp` futuro + cookie `user_data`) y sin ella. La que sostiene el
requisito:

- Anónimo en `/terminos`: 13.767 caracteres, **cero** enlaces a `/psi/terminos`
  y **cero** apariciones de «parte», «segundo documento» o «condiciones del
  portal».
- Anónimo en `/psi/terminos`: aterriza en `/terminos` sin haber pintado nunca
  el título del portal.
- Agremiado: sí ve el enlace cruzado; el documento del portal carga en su ruta
  con su canonical correcto.
- La banda roja «Texto legal pendiente» no aparece en ninguno de los dos.

Backend: `go build`, `go vet`, `make test-unit` y `atlas migrate validate` en
verde, con 11 tests nuevos en `analytics_privacy_test.go` — incluido uno de
contrato: aunque el middleware maneje la IP en claro, lo que llega al
repositorio ya está reducido.

⚠️ **Operativo**: tras aplicar `20260929210000_psi_terms_acceptance.sql`,
**reiniciar `colpsi_pgbouncer`** (gotcha 14 de `api/AGENTS.md`).
