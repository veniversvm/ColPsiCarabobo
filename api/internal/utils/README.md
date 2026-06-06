[← Volver al inicio](../../README.md)

# 🧰 `utils` — La caja de herramientas explicada para que cualquier programador la entienda

> **¿Qué es esto?** Una colección de funciones **sueltas y reutilizables** que usan todos los demás archivos del proyecto. Como un cajón de herramientas 🛠️: cada función hace una cosa específica y la hace bien. No importa si eres un junior que acaba de terminar un bootcamp, aquí te explico **cada línea** en español claro.
>
> **📍 Dónde está:** `api/internal/utils/`
>
> **🧠 Lema del paquete:** *"Una vez, y solo una vez"* — Si tienes que hacer algo más de una vez (validar un email, limpiar un texto, revisar una contraseña), esa lógica debe vivir aquí para que todos los demás archivos la compartan.

---

## 📋 Índice (lo que necesites, acá está)

| # | Archivo | ¿Qué hace? | ¿Para qué me sirve? |
|---|---|---|---|
| 1 | 🎨 `christ.go` | Muestra el logo del colegio en la terminal | Darle identidad visual al sistema |
| 2 | 🧹 `clean_alpha_numeric.go` | Borra símbolos raros, deja solo letras y números | Limpiar textos que vienen de usuarios |
| 3 | 🌎 `geo_venezuela.go` | Corrige nombres de municipios y estados | No guardar "Valencia" y "Balencia" como distintos |
| 4 | 🖼️ `image_sanitizer.go` | Comprime fotos y las vuelve seguras | Ahorrar espacio y evitar virus |
| 5 | 🕵️ `no_empty_req.go` | Dice si un formulario llegó vacío | No hacer QUERIES SQL al pedo |
| 6 | 🔤 `normalize_platform_name.go` | "ig" lo vuelve "Instagram" | Que el frontend muestre los iconos correctos |
| 7 | 🔐 `radom_string.go` | Genera textos aleatorios | Tokens de sesión, reseteo de passwords |
| 8 | 🔒 `secure_password.go` | Revisa si una clave es segura | Proteger cuentas de psicólogos |
| 9 | 📧 `validate_email.go` | Revisa si un email es válido | Que no te guarden "correo mal escrito" |
| 10 | 🧪 `utils_test.go` | Pruebas de todo lo anterior | Asegurar que ningún cambio rompa nada |

---

## ❓ Antes de empezar: conceptos que tenés que saber

> 💡 **Si ya sabés Go, skipeá esta sección.** Si venís de un bootcamp y esto es nuevo, leela bien.

### 🔹 ¿Qué es `package utils`?

Cada archivo `.go` arranca con `package utils`. Esto significa que **todos estos archivos son un solo grupo**: pueden verse las funciones entre sí, aunque estén en archivos separados. Es como si tuvieras 10 hojas de papel diferentes pero todas pertenecen al mismo capítulo de un libro.

### 🔹 ¿Qué es una función "pura"?

Una función **pura** es aquella que:
1. ✅ Dado el mismo input, SIEMPRE devuelve el mismo output.
2. ✅ No modifica nada fuera de sí misma (no toca la base de datos, no escribe archivos, no manda emails).

Ejemplo:
```go
// ✅ PURA: siempre que le pases 5, devuelve 25
func Cuadrado(x int) int { return x * x }

// ❌ NO PURA: depende de la hora actual
func HolaSegunHora() string {
    if time.Now().Hour() < 12 { return "Buen día" }
    return "Buenas tardes"
}
```

Todas las funciones de `utils` son **puras** (excepto `PrintColpsiASCII` que imprime en pantalla, pero eso es decorativo 🎨).

### 🔹 ¿Qué es `strings.Builder`?

Es una herramienta de Go para **armar textos largos sin ser lento**. Imaginate que tenés que escribir una carta de 1000 palabras:

- **MAL:** `carta = carta + palabra` → cada vez que agregás una palabra, tenés que copiar toda la carta anterior en una hoja nueva. 1000 copias = **muy lento** 🐢.
- **BIEN:** `strings.Builder` → agarrás una hoja gigante de una vez y escribís todo. 1 sola copia = **rápido** 🚀.

### 🔹 ¿Qué es una "runa" (rune) en Go?

Una **runa** (rune) es un carácter Unicode. En Go:
- `byte` = 1 letra inglesa (a, b, c, 1, 2, 3, $, %)
- `rune` = **cualquier carácter del mundo** (a, ñ, á, é, 中, 😀, 🎉)

Si usás bytes para recorrer un texto con `ñ`, se te rompe. Si usás runas, funciona perfecto. Por eso todas las funciones de este paquete usan **runas**.

### 🔹 ¿Qué es un "puntero" (`*bool`, `*string`)?

Un puntero es una **dirección** en la memoria de la computadora. En vez de guardar el valor directamente, guarda **dónde está** el valor.

```go
var x int = 10   // x = 10 (guardás el 10)
var p *int = &x  // p = dirección de x (guardás "está en la celda #123")
```

¿Por qué usarlos? Porque a veces necesitás decir **"esto no tiene valor"** (nil), no solo "esto es false". Un `bool` solo puede ser `true` o `false`. Un `*bool` puede ser `true`, `false`... **o nada** (`nil`). Eso es súper útil.

### 🔹 ¿Qué es `reflect` (reflexión)?

Es la capacidad de un programa de **examinarse a sí mismo**. Como si un espejo 🤳 le permitiera a tu código ver su propia estructura. `IsEmptyReq` usa reflexión para mirar dentro de un struct y ver si todos sus campos están vacíos, sin importar qué tipo de struct sea.

---

## 🎨 `christ.go` — Así se ve la identidad del colegio

### 📖 ¿Qué hace EXACTAMENTE?

```go
package utils      // Todos los archivos de esta carpeta son del mismo grupo

import "fmt"       // fmt = formato. Sirve para imprimir en pantalla

// PrintColpsiASCII dibuja el logo del Colegio en la terminal
func PrintColpsiASCII() {
    asciiArt := `   // ← Esto son BACKTICKS, no comillas
    ... (dibujito enorme) ...
    `
    fmt.Print(asciiArt)  // Manda el dibujito a la pantalla
    // log.Print(asciiArt)  // Alternativa comentada: manda el dibujito a los logs
}
```

### 🔍 Vamos línea por línea

| Línea | ¿Qué significa? | Traducción |
|---|---|---|
| `package utils` | Este archivo pertenece al grupo `utils` | "Soy parte de la caja de herramientas" |
| `import "fmt"` | Trae el módulo `fmt` para imprimir | "Necesito el martillo para clavar" |
| `func PrintColpsiASCII()` | Define una función que cualquiera puede llamar | "Acá hay una función lista para usar" |
| `asciiArt := \`...\`` | Guarda un texto gigante exactamente como está escrito | "El dibujo se guarda tal cual, con sus saltos de línea y todo" |
| `fmt.Print(asciiArt)` | Imprime el dibujo | "Sacá el dibujo a la pantalla" |

### 🧪 Cómo probarla manualmente

Creá un archivo `main.go` en cualquier lado y pegá esto:

```go
package main

import (
    "fmt"
    "tu-proyecto/api/internal/utils"
)

func main() {
    fmt.Println("¡Prendiendo el sistema de ColPsiCarabobo!")
    utils.PrintColpsiASCII()
    fmt.Println("Sistema listo ✅")
}
```

Luego corré `go run main.go`. Te va a aparecer el logo del colegio en la terminal.

### 🤔 Preguntas frecuentes de juniors

**P: ¿Por qué usa `fmt.Print` y no `fmt.Println`?**
R: Porque el dibujo ya termina con un salto de línea adentro. Si usara `Println`, agregaría OTRO salto de línea al final y quedaría un espacio feo.

**P: ¿Y la línea comentada de `log.Print`?**
R: Dejaron los dos preparados. `log.Print` escribiría en los archivos de registro del servidor (en vez de en la pantalla). Si algún día quieren que el logo aparezca en los logs en vez de en la terminal, solo cambian esa línea.

**P: ¿Ese dibujito es código o es arte?**
R: Es arte 🎨. Se llama *arte ASCII* y se hace solo con caracteres de texto. Es el escudo del colegio.

### ⚠️ Errores comunes

- ❌ **Error de importación:** Si no importás `fmt`, te da error. Go no te deja tener imports sin usar.
- ❌ **Backticks vs comillas:** Si usás comillas dobles `"` en vez de backticks `` ` ``, el string no puede tener saltos de línea y explota 💥.

---

## 🧹 `clean_alpha_numeric.go` — El filtro que solo deja pasar letras y números

### 📖 ¿Qué hace EXACTAMENTE?

```go
package utils

import (
    "strings"   // Tiene herramientas para manipular texto
    "unicode"   // Sabe reconocer TODO tipo de letras del mundo
)

// CleanAlphaNumeric recibe cualquier texto y devuelve SOLO letras y números
func CleanAlphaNumeric(s string) string {
    var builder strings.Builder          // Prepara una hoja en blanco para escribir
    builder.Grow(len(s))                  // La hoja es del tamaño exacto que necesito

    for _, r := range s {                 // Recorre el texto letra por letra
        if unicode.IsLetter(r) || unicode.IsDigit(r) {  // ¿Es letra o número?
            builder.WriteRune(r)          // Sí → la guarda en la hoja
        }
        // No → no hace nada, la descarta
    }

    return builder.String()               // Devuelve la hoja con lo que quedó
}
```

### 🔍 Vamos línea por línea

#### `import "strings"`

El paquete `strings` viene con Go (no hay que instalarlo). Tiene funciones como `ToLower`, `TrimSpace`, `ReplaceAll`. En esta función **no usamos directamente** `strings`, pero el paquete está importado porque otros archivos del grupo lo necesitan (y Go obliga a que el package sea coherente).

#### `import "unicode"`

El paquete `unicode` sabe reconocer **cualquier letra de cualquier idioma** del mundo:
- `unicode.IsLetter('a')` → `true` ✅ (inglés)
- `unicode.IsLetter('ñ')` → `true` ✅ (español)
- `unicode.IsLetter('é')` → `true` ✅ (francés)
- `unicode.IsLetter('中')` → `true` ✅ (chino)
- `unicode.IsLetter('3')` → `false` ❌ (es número)
- `unicode.IsLetter('!')` → `false` ❌ (es símbolo)

#### `builder.Grow(len(s))` — El truco de velocidad

Imaginate que tenés que filtrar arena con un colador:

- **Sin `Grow`:** Agarrás una taza de arena, la colás, ponés lo que queda en un plato. Agarrás otra taza, la colás, ponés lo que queda en OTRO plato, y al final tenés que juntar todos los platos. **Muchos platos = lento**.
- **Con `Grow`:** Agarrás una bandeja gigante desde el principio. Colás toda la arena de una y va cayendo directamente a la bandeja. **Una bandeja = rápido**.

El `len(s)` le dice "este texto tiene 20 caracteres, prepará espacio para 20".

#### `for _, r := range s`

En la mayoría de los lenguajes, si hacés un loop sobre un string, obtenés **letras individuales**. En Go, si el string tiene `ñ`, obtener bytes te daría `195` y `177` (dos números raros), no la letra `ñ`.

Usando `for _, r := range s`, Go automáticamente:
1. Reconoce caracteres de múltiples bytes (como `ñ`, `á`, `😀`).
2. Los trata como una sola unidad (runa).
3. Te devuelve la runa completa.

#### `unicode.IsLetter(r) || unicode.IsDigit(r)`

| `r` (carácter) | `IsLetter(r)` | `IsDigit(r)` | ¿Pasa el filtro? |
|---|---|---|---|
| `'a'` | ✅ `true` | ❌ `false` | ✅ Sí (es letra) |
| `'Z'` | ✅ `true` | ❌ `false` | ✅ Sí |
| `'5'` | ❌ `false` | ✅ `true` | ✅ Sí (es número) |
| `'ñ'` | ✅ `true` | ❌ `false` | ✅ Sí |
| `'á'` | ✅ `true` | ❌ `false` | ✅ Sí |
| `' '` (espacio) | ❌ `false` | ❌ `false` | ❌ No |
| `'.'` | ❌ `false` | ❌ `false` | ❌ No |
| `'<'` | ❌ `false` | ❌ `false` | ❌ No |

#### `builder.WriteRune(r)`

Agrega la runa (carácter) al final del texto que estamos construyendo. Es más rápido que hacer `texto = texto + string(r)` porque no crea un string nuevo cada vez.

### 🧪 Qué devuelve con diferentes entradas

```go
CleanAlphaNumeric("Hola Mundo!")          // → "HolaMundo"     (borra espacio y !)
CleanAlphaNumeric("<script>alert('xss')</script>") // → "scriptalertxssscript"  (borra todo lo malo)
CleanAlphaNumeric("  María José 2024  ")  // → "MaríaJosé2024"  (borra espacios, preserva acentos)
CleanAlphaNumeric("")                      // → ""              (vacío → vacío)
CleanAlphaNumeric("!!!@@@###")             // → ""              (solo símbolos → vacío)
CleanAlphaNumeric("user@dominio.com")      // → "userdominiocom" (borra @ y .)
```

### 🧪 Cómo probarla manualmente

Creá un archivo `test_clean.go`:

```go
package main

import (
    "fmt"
    "tu-proyecto/api/internal/utils"
)

func main() {
    tests := []string{
        "Hola Mundo!",
        "<script>alert('xss')</script>",
        "  María José 2024  ",
        "",
        "user@dominio.com",
    }
    for _, t := range tests {
        resultado := utils.CleanAlphaNumeric(t)
        fmt.Printf("Entrada: %q → Salida: %q\n", t, resultado)
    }
}
```

Corré con `go run test_clean.go`.

### 📋 ¿Cuándo usar esta función?

| Situación | ¿Usar? | Explicación |
|---|---|---|
| Limpiar un nombre de usuario | ✅ Sí | "  JOSÉ   " → "José" |
| Validar un email | ❌ No | Borra el @ y el . |
| Crear un slug para URL | ❌ No | Borra los guiones |
| Sanitizar un campo de búsqueda | ✅ Sí | Evita inyección SQL |
| Limpiar una dirección | ❌ No | Pierde números y espacios |

### ⚠️ Errores comunes

1. **Pensar que conserva espacios:** No, los espacios se borran.
2. **Usarlo para emails:** Borra el `@`, así que `user@mail.com` → `usermailcom`, que no sirve.
3. **No entender las runas:** Si modificás el código para usar bytes en vez de runas, la `ñ` se rompe.
4. **Olvidar el `Grow`:** Sin eso funciona igual, pero es más lento. En una app chica no se nota, pero con 10,000 peticiones por segundo, sí.

---

## 🌎 `geo_venezuela.go` — Geografía venezolana, sin errores de tipeo

### 📖 ¿Qué hace EXACTAMENTE?

Tres funciones en un mismo archivo:

1. **`NormalizeMunicipioCarabobo`** — Corrige el nombre de un municipio de Carabobo.
2. **`NormalizeEstadoVenezuela`** — Corrige el nombre de un estado de Venezuela.
3. **`BoolFromForm`** — Traduce strings de formularios a verdadero/falso/nada.

### 🌆 1. `NormalizeMunicipioCarabobo` — Para direcciones en Carabobo

```go
func NormalizeMunicipioCarabobo(input string) (string, bool) {
    normalized := strings.TrimSpace(input)  // Saca espacios de los bordes

    for _, m := range municipiosCarabobo {  // Recorre la lista de municipios
        if foldCompare(normalized, m) {     // ¿Se parecen?
            return m, true                   // Sí → devuelve el nombre oficial y "ok"
        }
    }

    return "", false                         // No → devuelve vacío y "falló"
}
```

#### 🔍 La lista de municipios (en el código)

```go
var municipiosCarabobo = []string{
    "Bejuma",
    "Carlos Arvelo",
    "Diego Ibarra",
    "Guacara",
    "Juan José Mora",
    "Libertador",
    "Los Guayos",
    "Miranda",
    "Montalbán",
    "Naguanagua",
    "Puerto Cabello",
    "San Diego",
    "San Joaquín",
    "Valencia",
}
```

Son **14 municipios**. ¿No sabés cuáles son los municipios de Carabobo? No importa: esta lista ya los tiene. Es la versión digital de una **lista oficial**.

#### 🔍 La función `foldCompare` — el corazón de la lógica

```go
func foldCompare(a, b string) bool {
    return strings.EqualFold(removeDiacritics(a), removeDiacritics(b))
}
```

Esta función se pregunta: **"¿Son estos dos textos equivalentes, aunque uno tenga tildes y otro no, uno esté en mayúsculas y otro no?"**

**Paso 1: `removeDiacritics(a)`** — Le saca los acentos a `a`.

**Paso 2: `removeDiacritics(b)`** — Le saca los acentos a `b`.

**Paso 3: `strings.EqualFold(...)`** — Pregunta "¿son iguales sin importar mayúsculas/minúsculas?"

Ejemplo:
```go
foldCompare("San Joaquín", "SAN JOAQUIN")
// Paso 1: removeDiacritics("San Joaquín") → "San Joaquin"
// Paso 2: removeDiacritics("SAN JOAQUIN") → "SAN JOAQUIN"
// Paso 3: EqualFold("San Joaquin", "SAN JOAQUIN") → true ✅
```

#### 🔍 `removeDiacritics` — la magia de sacar tildes

```go
func removeDiacritics(s string) string {
    t := transform.Chain(
        norm.NFD,                    // Parte la 'á' en 'a' + '´'
        transform.RemoveFunc(func(r rune) bool {
            return unicode.Is(unicode.Mn, r)  // Saca los acentos
        }),
        norm.NFC,                    // Junta lo que quedó
    )
    result, _, _ := transform.String(t, s)
    return result
}
```

No te asustes con el nombre. Es más simple de lo que parece:

1. **NFD:** Toma cada letra con acento y la separa en dos: la letra + el acentito. `"á"` se convierte en `"a" + "´"`.
2. **RemoveFunc:** Saca los acentitos (todo lo que sea `unicode.Mn`, que significa "Marca, sin ancho" — justo lo que son las tildes).
3. **NFC:** Junta las letras que quedaron en un texto normal.

Resultado: `"Montalbán"` → `"Montalban"`, `"San Joaquín"` → `"San Joaquin"`, etc.

### 🇻🇪 2. `NormalizeEstadoVenezuela` — Para direcciones fuera de Carabobo

```go
func NormalizeEstadoVenezuela(input string) (string, bool) {
    normalized := strings.TrimSpace(input)
    for _, e := range estadosVenezuela {
        if foldCompare(normalized, e) {
            return e, true
        }
    }
    return "", false
}
```

> [!WARNING]
> **Carabobo NO está en la lista.** ¿Por qué? Porque el sistema ya sabe que la ubicación principal es Carabobo. Esta lista es para psicólogos que viven en **otros estados**. Si alguien pone "Carabobo" aquí, no lo va a encontrar y la función va a decir "no válido". Es intencional.

### ✅ 3. `BoolFromForm` — Cuando un formulario se vuelve código

```go
func BoolFromForm(s string) *bool {
    switch strings.ToLower(strings.TrimSpace(s)) {
    case "1", "true", "yes":
        v := true
        return &v
    case "0", "false", "no":
        v := false
        return &v
    default:
        return nil
    }
}
```

#### 🎯 ¿Cuál es el problema que resuelve esto?

Imaginate que tenés un formulario web donde un psicólogo puede activar o desactivar su disponibilidad:

```html
<input type="checkbox" name="disponible" value="1">
```

**Caso 1:** El usuario marca el checkbox → el navegador envía `"disponible=1"`.
**Caso 2:** El usuario desmarca → el navegador NO envía nada.
**Caso 3:** El usuario ni tocó el campo → el navegador NO envía nada.

Los casos 2 y 3 son distintos, pero el navegador envía lo mismo: **nada**. Sin `BoolFromForm`, no podrías diferenciar:
- "El usuario dijo que NO disponible" → `UPDATE disponible = false`
- "El usuario no tocó el campo" → NO hacer nada, dejar el valor anterior.

#### 💡 La solución del puntero

| Lo que viene del formulario | `BoolFromForm` devuelve | `nil`? | ¿Qué hace la base de datos? |
|---|---|---|---|
| `"1"`, `"true"`, `"yes"` | ✅ `&true` (puntero a verdadero) | No | `UPDATE campo = true` |
| `"0"`, `"false"`, `"no"` | ❌ `&false` (puntero a falso) | No | `UPDATE campo = false` |
| Cualquier otra cosa (vacío, `"quizas"`, etc.) | ⬜ `nil` (nada) | **Sí** | **No toca el campo** |

#### 🧪 Cómo probarla manualmente

```go
package main

import (
    "fmt"
    "tu-proyecto/api/internal/utils"
)

func main() {
    tests := []string{"1", "true", "yes", "0", "false", "no", "", "tal vez"}
    for _, t := range tests {
        resultado := utils.BoolFromForm(t)
        if resultado == nil {
            fmt.Printf("Entrada: %q → nil (no tocar)\n", t)
        } else {
            fmt.Printf("Entrada: %q → %v (actualizar)\n", t, *resultado)
        }
    }
}
```

### 🧪 Tabla COMPLETA de lo que acepta `NormalizeMunicipioCarabobo`

| Lo que escribe el usuario | Lo que devuelve | `ok` | Explicación |
|---|---|---|---|
| `"Naguanagua"` | `"Naguanagua"` | ✅ `true` | Exacto |
| `"naguanagua"` | `"Naguanagua"` | ✅ `true` | Minúsculas |
| `"  NAGUANAGUA  "` | `"Naguanagua"` | ✅ `true` | Mayúsculas + espacios |
| `"Naganagua"` | `""` | ❌ `false` | Error de tipeo (falta la u) |
| `"San Joaquín"` | `"San Joaquín"` | ✅ `true` | Con tilde |
| `"san joaquin"` | `"San Joaquín"` | ✅ `true` | Sin tilde, minúsculas |
| `"San Joaquin"` | `"San Joaquín"` | ✅ `true` | Sin tilde, mayúscula |
| `"Montalbán"` | `"Montalbán"` | ✅ `true` | Con tilde |
| `"montalban"` | `"Montalbán"` | ✅ `true` | Sin tilde |
| `"Caracas"` | `""` | ❌ `false` | No es un municipio de Carabobo |
| `""` (vacío) | `""` | ❌ `false` | No escribió nada |

### ⚠️ Errores comunes

1. **Pensar que `BoolFromForm` recibe un `bool`:** Recibe un `string`. Los formularios web siempre envían texto.
2. **No revisar el segundo valor (`ok`):** Si llamás a `NormalizeMunicipioCarabobo` y no revisás el `bool`, podés estar usando un string vacío como si fuera un municipio válido.
3. **Olvidar que Carabobo no está en `estadosVenezuela`:** Si un usuario pone "Carabobo" ahí, la función dice "no válido", pero es correcto: Carabobo va en otro campo.

---

## 🖼️ `image_sanitizer.go` — El filtro más importante del sistema

### 📖 ¿Qué hace EXACTAMENTE?

Cada vez que alguien sube una foto o un documento:
1. 🛡️ **Revisa que sea una imagen real** (no un virus).
2. ✂️ **La achica** si es muy grande.
3. 🎨 **Le pone fondo blanco** donde haya transparencia.
4. 📦 **La comprime al máximo** para ahorrar espacio.
5. ✅ **Devuelve los bytes listos** para guardar en S3.

### 🔍 Desglose de las constantes

```go
const (
    maxAvatarSizeBytes   = 150 * 1024  // = 153,600 bytes ≈ 150 KB
    maxDocumentSizeBytes = 400 * 1024  // = 409,600 bytes ≈ 400 KB
    compressionScaleFactor = 0.8       // = 80% (reduce 20% cada vez)
    minDimensionPx         = 100       // No reducir a menos de 100 px
    webpQuality            = 80.0      // Calidad 80/100
    maxAvatarDimension     = 800       // 800 píxeles de ancho/alto máximo
    maxDocumentDimension   = 1600      // 1600 píxeles de ancho/alto máximo
)
```

#### ¿Por qué estos números y no otros?

| Constante | Valor | ¿Por qué exactamente este número? |
|---|---|---|
| `maxAvatarSizeBytes` | 150 KB | Una foto de perfil debe cargar en menos de 1 segundo incluso en 3G. 150 KB es el límite recomendado por Google para que una imagen se vea "instantánea". |
| `maxDocumentSizeBytes` | 400 KB | Los títulos y certificados necesitan calidad para que se lean los sellos y firmas. 400 KB es suficiente para un documento escaneado legible. |
| `compressionScaleFactor` | 0.8 (80%) | Reducir 20% cada vez: si la imagen pesa 500 KB y debe pesar 150 KB, después de 1 ciclo → 320 KB, 2 ciclos → 205 KB, 3 ciclos → 131 KB ✅. Es un balance entre calidad y velocidad. |
| `webpQuality` | 80 | En WebP, 80 es el punto óptimo: se ve casi igual que 100 pero ocupa la mitad. Google lo recomienda. |
| `minDimensionPx` | 100 | Una imagen de menos de 100 px se ve como un puntito y no sirve. Es un seguro para no hacer bucles infinitos. |
| `maxAvatarDimension` | 800 | En una pantalla Retina (2x), 400 px se ven nítidos como 800 px reales. No necesitás más. |
| `maxDocumentDimension` | 1600 | Suficiente para hacer zoom y leer texto, sin saturar la memoria. |

### 🔄 El proceso completo paso a paso (con dibujo)

```
📸 El usuario sube un archivo desde su celular/computadora
│
│   Ejemplo: "foto_perfil.jpg" (4000 x 3000 px, 5.2 MB)
│
▼
┌──────────────────────────────────────────────────────────────────────┐
│ FASE 1: ¿Es esto una imagen de verdad?                             │
│                                                                    │
│ image.Decode(file) lee los primeros bytes del archivo.             │
│                                                                     │
│ Los archivos tienen "magic numbers":                                │
│   - JPG empieza con FF D8 FF                                       │
│   - PNG empieza con 89 50 4E 47                                    │
│   - GIF empieza con 47 49 46                                       │
│                                                                     │
│ Si el archivo dice "foto.jpg" pero empieza con "<?php"             │
│ → ❌ error, no pasa                                                  │
│                                                                     │
│ Si el archivo es un JPG real → extrae solo los píxeles             │
│ (ignora metadatos EXIF: dónde se tomó la foto, con qué cámara,     │
│  GPS, etc.)                                                         │
└──────────────────────────────────────────────────────────────────────┘
│
▼
┌──────────────────────────────────────────────────────────────────────┐
│ FASE 2: ¿Es muy grande? Achícala                                  │
│                                                                     │
│ La foto es 4000x3000. El límite de avatar es 800px.                │
│ → La reduce a 800x600 (manteniendo la proporción).                 │
│                                                                     │
│ ¿Qué significa "mantener la proporción"?                           │
│   - Original: 4000 ancho, 3000 alto → proporción 4:3              │
│   - Reducida: 800 ancho, 600 alto → proporción 4:3                │
│   - Si no se mantuviera, la imagen se vería estirada/achatada     │
│                                                                     │
│ La cuenta matemática:                                              │
│   El lado más grande es 4000 (el ancho)                           │
│   new_ancho = 800 (el máximo)                                      │
│   new_alto = 3000 * (800 / 4000) = 3000 * 0.2 = 600              │
└──────────────────────────────────────────────────────────────────────┘
│
▼
┌──────────────────────────────────────────────────────────────────────┐
│ FASE 3: ¿Tiene transparencia? Ponele fondo blanco                 │
│                                                                     │
│ Los PNG pueden tener partes transparentes (como un PNG de logo     │
│ con fondo de cuadritos).                                           │
│                                                                     │
│ ¿Por qué sacar la transparencia?                                   │
│   - Si el usuario usa modo oscuro 🌙 y la imagen tiene             │
│     transparencia, se ve un halo negro feo.                        │
│   - Con fondo blanco siempre se ve bien.                           │
│                                                                     │
│ Cómo se hace:                                                      │
│   draw.Draw(..., &image.Uniform{color.White}, ..., draw.Src)      │
│   → Pinta TODO el fondo de blanco.                                 │
│   draw.Draw(..., src, ..., draw.Over)                             │
│   → Pinta la imagen encima, donde no es transparente.             │
└──────────────────────────────────────────────────────────────────────┘
│
▼
┌──────────────────────────────────────────────────────────────────────┐
│ FASE 4: Comprimir a WebP (y si pesa mucho, repetir)               │
│                                                                     │
│ Convierte la imagen al formato WebP (moderno, de Google).          │
│                                                                     │
│ Pesa la imagen comprimida: ¿150 KB o menos?                        │
│   ✅ Sí → Listo, devolver los bytes.                               │
│   ❌ No → Reduce la imagen un 20% y vuelve a comprimir.            │
│                                                                     │
│ Ciclo 1: 5.2 MB → WebP → 800 KB ❌ (pesa mucho)                    │
│ Ciclo 2: Reduce 20% → 640x480 → WebP → 500 KB ❌                  │
│ Ciclo 3: Reduce 20% → 512x384 → WebP → 320 KB ❌                  │
│ Ciclo 4: Reduce 20% → 410x307 → WebP → 200 KB ❌                  │
│ Ciclo 5: Reduce 20% → 328x246 → WebP → 128 KB ✅                  │
│                                                                     │
│ Pero si en algún momento la imagen queda más chica que 100 px:     │
│   → Para el bucle aunque pese más del límite.                      │
│   → Es mejor una imagen chica que un servidor trabado.             │
└──────────────────────────────────────────────────────────────────────┘
│
▼
✅ Resultado final:
   Formato: .webp
   MIME: image/webp
   Tamaño: ~128 KB (antes era 5.2 MB → 97% más chica)
   Segura: sin virus, sin GPS, sin metadatos
```

### 🔍 Explicación línea por línea de las funciones clave

#### `SanitizeImage` y `SanitizeDocument` — las dos caras de la misma moneda

```go
func SanitizeImage(file io.Reader) ([]byte, string, string, error) {
    return processImage(file, maxAvatarDimension, maxAvatarSizeBytes)
}

func SanitizeDocument(file io.Reader) ([]byte, string, string, error) {
    return processImage(file, maxDocumentDimension, maxDocumentSizeBytes)
}
```

Son **idénticas** excepto por los límites:
- `SanitizeImage` → 800 px, 150 KB
- `SanitizeDocument` → 1600 px, 400 KB

La lógica real está en `processImage`. Estas dos funciones son "accesos directos" con configuraciones pre-hechas.

#### `processImage` — el director de orquesta

```go
func processImage(file io.Reader, maxDimension int, maxSizeBytes int) ([]byte, string, string, error) {
    // 1. Decodificar
    img, _, err := image.Decode(file)
    if err != nil {
        log.Printf("Error decodificando imagen: %v", err)
        return nil, "", "", errors.New("el servidor no reconoce este formato de imagen")
    }

    // 2. Redimensionar
    img = capDimensions(img, maxDimension)

    // 3. Quitar transparencia
    img = FlattenAlpha(img)

    // 4. Comprimir a WebP
    compressed, err := compressToWebP(img, maxSizeBytes)
    if err != nil {
        return nil, "", "", errors.New("error al codificar la imagen")
    }

    return compressed, ".webp", "image/webp", nil
}
```

#### `capDimensions` — la que no deja que las imágenes se pasen de grandes

```go
func capDimensions(src image.Image, maxDimension int) image.Image {
    bounds := src.Bounds()     // ¿Cuánto mide?
    w := bounds.Dx()           // Ancho actual
    h := bounds.Dy()           // Alto actual

    // Si ya es más chica que el máximo, ni la toques
    if w <= maxDimension && h <= maxDimension {
        return src
    }

    // Calculá las nuevas dimensiones
    var newW, newH int
    if w >= h {                // ¿Es más ancha que alta? (apaisada)
        newW = maxDimension    // El ancho se vuelve el máximo
        newH = int(float64(h) * float64(maxDimension) / float64(w))  // Alto proporcional
    } else {                   // ¿Es más alta que ancha? (retrato)
        newH = maxDimension    // El alto se vuelve el máximo
        newW = int(float64(w) * float64(maxDimension) / float64(h))  // Ancho proporcional
    }

    return resizeImage(src, newW, newH)
}
```

**Analogía:** Tenés una foto enorme y querés que entre en un marco de 800x800.
- Si la foto es 4000x3000 (apaisada), el marco se pone de costado: ancho 800, alto 600.
- Si la foto es 3000x4000 (retrato), el marco se pone vertical: alto 800, ancho 600.

#### `compressToWebP` — la que aprieta hasta que entre en el presupuesto

```go
func compressToWebP(img image.Image, maxSizeBytes int) ([]byte, error) {
    options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, webpQuality)
    if err != nil {
        return nil, err
    }

    current := img   // Arranca con la imagen actual
    for {
        buf := new(bytes.Buffer)
        webp.Encode(buf, current, options)  // Comprime a WebP

        // ¿Entra en el límite?
        if len(buf.Bytes()) <= maxSizeBytes {
            return buf.Bytes(), nil          // ✅ Sí, devolver
        }

        // ❌ No entra. Achicar 20%
        bounds := current.Bounds()
        newW := int(float64(bounds.Dx()) * 0.8)
        newH := int(float64(bounds.Dy()) * 0.8)

        // Si es más chica que 100 px, rendirse
        if newW < 100 || newH < 100 {
            return buf.Bytes(), nil           // ⛔ Salir, no se puede comprimir más
        }

        current = resizeImage(current, newW, newH)  // Redimensionar y repetir
    }
}
```

### 🧪 Cómo probarla manualmente

Necesitás una imagen real. Poné cualquier JPG o PNG en `test.jpg`:

```go
package main

import (
    "fmt"
    "os"
    "tu-proyecto/api/internal/utils"
)

func main() {
    // Abrí una imagen
    file, err := os.Open("test.jpg")
    if err != nil {
        fmt.Println("Error abriendo archivo:", err)
        return
    }
    defer file.Close()

    // Procesala
    data, ext, mime, err := utils.SanitizeImage(file)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Printf("✅ Imagen procesada!\n")
    fmt.Printf("   Tamaño: %d bytes (%.2f KB)\n", len(data), float64(len(data))/1024)
    fmt.Printf("   Extensión: %s\n", ext)
    fmt.Printf("   Tipo MIME: %s\n", mime)

    // Guardala para ver el resultado
    os.WriteFile("test_output.webp", data, 0644)
    fmt.Println("   Guardado como test_output.webp")
}
```

### ⚠️ Errores comunes

1. **Pasar un archivo que no es imagen:** `SanitizeImage` devuelve error, pero el texto del error es genérico por seguridad ("no reconozco este formato"). No dice "es un virus" porque los hackers usan esa info.
2. **Olvidar cerrar el archivo:** Notá el `defer file.Close()` en el ejemplo. Si no cerrás el archivo, se queda abierto y gastás memoria.
3. **Asumir que el formato de salida siempre es JPG:** Siempre es WebP. Si el frontend no soporta WebP, hay que convertirlo.
4. **No revisar el error:** La función devuelve 4 valores. Si ignorás el error y usás los bytes igual, podés tener datos corruptos.

---

## 🕵️ `no_empty_req.go` — El detector que evita desastres

### 📖 ¿Qué hace EXACTAMENTE?

```go
func IsEmptyReq(s interface{}) bool {
    v := reflect.ValueOf(s)     // "Mirá" qué hay adentro de lo que te pasaron

    // Si es un puntero, abrirlo para ver qué contiene
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
    }

    // Revisá campo por campo
    for i := 0; i < v.NumField(); i++ {
        if !v.Field(i).IsZero() {  // ¿Este campo tiene algún valor?
            return false            // → Hay datos, no está vacío
        }
    }

    return true  // Todos los campos están vacíos
}
```

### 🔍 ¿Qué significa cada línea?

#### `func IsEmptyReq(s interface{}) bool`

`interface{}` significa "**cualquier tipo** de dato". Podés pasarle un struct, un puntero, un string, lo que sea. Es la forma que tiene Go de decir "no sé qué tipo es, pero procesalo igual".

#### `v := reflect.ValueOf(s)`

Reflexión. Es como ponerle un **rayo X** 👁️ al struct para ver qué tiene adentro: cuántos campos tiene, cómo se llaman, qué valores tienen, de qué tipo son.

#### `if v.Kind() == reflect.Ptr`

¿Es un puntero? O sea, ¿el usuario pasó `&MiStruct{}` en vez de `MiStruct{}`? Si es así, `v.Elem()` "abre" el puntero y mira qué hay adentro (el struct real).

#### `for i := 0; i < v.NumField(); i++`

`v.NumField()` dice "este struct tiene 4 campos". El for los recorre del 0 al 3.

#### `v.Field(i).IsZero()`

`IsZero()` pregunta: "¿Este campo tiene el valor **por defecto** de Go?"

| Tipo de campo | Zero value (valor por defecto) |
|---|---|
| `string` | `""` (vacío) |
| `int` | `0` |
| `bool` | `false` |
| `*bool` (puntero) | `nil` (no apunta a nada) |
| `[]string` (slice) | `nil` (no tiene elementos) |
| `float64` | `0.0` |

### 🎯 Ejemplo de la vida real

Imaginá que el frontend hace un PATCH a `api/psicologo/123` para actualizar datos:

```go
type UpdateRequest struct {
    Name    string  `json:"name"`
    Email   string  `json:"email"`
    Phone   string  `json:"phone"`
}

// El frontend envía: {}
// (quería actualizar pero no mandó nada)
```

Sin `IsEmptyReq`:
```go
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
    var req UpdateRequest
    json.NewDecoder(r.Body).Decode(&req)

    // req = {Name: "", Email: "", Phone: ""}
    db.Model(&psicologo).Updates(req)
    // UPDATE psicologos SET name = '', email = '', phone = '' WHERE id = 123
    // 💥 BORRÓ todos los datos del psicólogo!
}
```

Con `IsEmptyReq`:
```go
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
    var req UpdateRequest
    json.NewDecoder(r.Body).Decode(&req)

    if IsEmptyReq(req) {
        http.Error(w, "No enviaste datos para actualizar", http.StatusBadRequest)
        return  // 🛑 Se detiene antes de hacer daño
    }

    db.Model(&psicologo).Updates(req)
    // Solo se ejecuta si hay al menos un campo con datos
}
```

### 🔍 ¿Qué pasa con diferentes tipos de struct?

#### Struct vacío → `true`
```go
type Persona struct {
    Name string
    Age  int
}
IsEmptyReq(Persona{})  // → true
```
- Name = `""` → IsZero = true
- Age = `0` → IsZero = true
- Todos true → devuelve true (vacío)

#### Struct con datos → `false`
```go
IsEmptyReq(Persona{Name: "Fran"})  // → false
```
- Name = `"Fran"` → IsZero = false
- Ya encontró uno no vacío → devuelve false

#### Puntero a struct vacío → `true`
```go
IsEmptyReq(&Persona{})  // → true
```
- Detecta que es puntero, lo abre, ve `Persona{}`, todo vacío → true

### 🧪 Cómo probarla manualmente

```go
package main

import (
    "fmt"
    "tu-proyecto/api/internal/utils"
)

type TestStruct struct {
    Name string
    Age  int
}

func main() {
    fmt.Println(utils.IsEmptyReq(TestStruct{}))           // → true
    fmt.Println(utils.IsEmptyReq(TestStruct{Name: "A"}))  // → false
    fmt.Println(utils.IsEmptyReq(TestStruct{Age: 25}))    // → false
    fmt.Println(utils.IsEmptyReq(&TestStruct{}))           // → true
    fmt.Println(utils.IsEmptyReq(&TestStruct{Name: "B"})) // → false
}
```

### ⚠️ Errores comunes

1. **Pasar un tipo que no es struct:** Si pasás un `int` o un `string`, `reflect.ValueOf` funciona pero `NumField()` no existe para esos tipos y el programa **entra en pánico** (panic). La función está diseñada solo para structs.
2. **Olvidar que los bools por defecto son `false`:** Si el struct tiene `Activo bool`, por defecto es `false`. `IsEmptyReq` dice que ese campo está vacío. Pero tal vez el valor por defecto de `Activo` es justamente `false`. Para esos casos hay que usar punteros (`*bool`).
3. **Pensar que es mágica:** Usa reflexión, que es más lenta que el código normal. No la llames 10,000 veces por segundo sin medir el rendimiento.

---

## 🔤 `normalize_platform_name.go` — El traductor de redes sociales

### 📖 ¿Qué hace EXACTAMENTE?

```go
func NormalizePlatformName(name string) string
```

Recibe el nombre de una red social **como lo escribió el usuario** y devuelve el nombre **oficial**.

### 💡 ¿Por qué es necesaria?

Imaginate que un psicólogo pone su Instagram en el perfil:

> **Psicólogo 1:** "Sígueme en ig: @psicologo1"
> **Psicólogo 2:** "Mi Insta: @psicologo2"
> **Psicólogo 3:** "instagram.com/psicologo3"

Sin `NormalizePlatformName`:
- En la base de datos se guardaría: `"ig"`, `"Insta"`, `"instagram.com/psicologo3"` (3 strings distintos).
- El frontend no sabe qué ícono mostrar. Tiene que tener 3 condicionales distintos.
- Las búsquedas no funcionan: "mostrame todos los que tienen Instagram" no encuentra a los que pusieron "ig".

Con `NormalizePlatformName`:
- Se guarda: `"Instagram"`, `"Instagram"`, `"Instagram"` (siempre igual).
- El frontend solo necesita: `if plataforma == "Instagram" { mostrarIcono("instagram.svg") }`.

### 🔍 El diccionario completo (las 18 redes que entiende)

```go
var platformVariants = map[string]string{
    "instagram": "Instagram",     "ig": "Instagram",       "insta": "Instagram",
    "instagran": "Instagram",     "instgram": "Instagram",

    "facebook": "Facebook",       "fb": "Facebook",        "face": "Facebook",
    "facbook": "Facebook",        "facebok": "Facebook",   "fbk": "Facebook",

    "twitter": "X (Twitter)",     "x": "X (Twitter)",      "tw": "X (Twitter)",
    "twiter": "X (Twitter)",      "twiiter": "X (Twitter)","twttr": "X (Twitter)",

    "youtube": "YouTube",         "yt": "YouTube",         "yutube": "YouTube",
    "ytb": "YouTube",             "youtub": "YouTube",     "tube": "YouTube",
    "youtu.be": "YouTube",        "youtube.com": "YouTube",

    "linkedin": "LinkedIn",       "in": "LinkedIn",        "linkdin": "LinkedIn",
    "lnkd": "LinkedIn",

    "tiktok": "TikTok",           "tk": "TikTok",          "ticktok": "TikTok",
    "tictok": "TikTok",

    "whatsapp": "WhatsApp",       "wa": "WhatsApp",        "wsp": "WhatsApp",
    "watsapp": "WhatsApp",        "whatsap": "WhatsApp",   "wa.me": "WhatsApp",

    "snapchat": "Snapchat",       "snap": "Snapchat",      "sc": "Snapchat",
    "snapc": "Snapchat",

    "pinterest": "Pinterest",     "pin": "Pinterest",      "pint": "Pinterest",
    "pinterst": "Pinterest",

    "reddit": "Reddit",           "rd": "Reddit",          "redit": "Reddit",

    "telegram": "Telegram",       "tg": "Telegram",        "t.me": "Telegram",
    "tele": "Telegram",

    "discord": "Discord",         "dc": "Discord",

    "twitch": "Twitch",           "twitchtv": "Twitch",

    "signal": "Signal",
    "wechat": "WeChat",           "wc": "WeChat",
    "line": "Line",
    "viber": "Viber",
}
```

### 🔍 Las 4 fases del proceso

```
📥 Entrada: cualquier cosa que escriba el usuario
│
│   Ejemplo: "  INSTA  " , "https://youtu.be/dQw4w9WgXcQ" , "bluesky"
│
▼
┌──────────────────────────────────────────────────────────────────────┐
│ FASE 1: LIMPIEZA                                                    │
│                                                                     │
│ clean = strings.ToLower(name)         → "  insta  "                 │
│ clean = strings.TrimSpace(clean)      → "insta"                     │
│ lookup = strings.ReplaceAll(clean, " ", "") → "insta"               │
│                                                                     │
│ ¿Por qué dos pasos de limpieza?                                     │
│ - ToLower: "Insta" e "insta" deben ser lo mismo.                   │
│ - TrimSpace: "  insta  " tiene espacios al inicio/fin.             │
│ - ReplaceAll: "insta gram" (con espacio adentro) → "instagram".    │
│   (Algún usuario puede escribir "insta gram" separado).            │
└──────────────────────────────────────────────────────────────────────┘
│
▼
┌──────────────────────────────────────────────────────────────────────┐
│ FASE 2: BUSCAR EN EL DICCIONARIO (LOOKUP)                          │
│                                                                     │
│ El mapa es como un diccionario de papel. Le decís una palabra      │
│ y te dice su traducción.                                           │
│                                                                     │
│ ¿"insta" está en el mapa? → Sí → devuelve "Instagram" ✅            │
│                                                                     │
│ Esto funciona para el 95% de los casos.                            │
│                                                                     │
│ ¿Por qué es rápido?                                                │
│ Los mapas en Go son "hash tables". Buscar una palabra en un mapa   │
│ con 50 entradas es tan rápido como buscar en uno con 50,000.      │
│ Se llama "tiempo constante O(1)".                                  │
└──────────────────────────────────────────────────────────────────────┘
│
│   Si NO encontró en el mapa...
│
▼
┌──────────────────────────────────────────────────────────────────────┐
│ FASE 3: BUSCAR POR PARECIDO (PATTERN MATCHING)                     │
│                                                                     │
│ Acá no busca una palabra exacta. Pregunta:                         │
│ "¿El texto contiene alguna palabra clave?"                         │
│                                                                     │
│ ¿Contiene "youtu"? → "YouTube"     (cubre youtu.be, youtube.com)  │
│ ¿Contiene "instagr"? → "Instagram" (cubre instagram.com, ...)     │
│ ¿Contiene "facebo"? → "Facebook"                                   │
│ ¿Contiene "fb.com"? → "Facebook"                                    │
│                                                                     │
│ Ejemplo: "https://youtu.be/dQw4w9WgXcQ"                            │
│ → No está en el mapa (es una URL larga).                           │
│ → Contiene "youtu"? → ✅ Sí → "YouTube"                             │
└──────────────────────────────────────────────────────────────────────┘
│
│   Si tampoco encontró...
│
▼
┌──────────────────────────────────────────────────────────────────────┐
│ FASE 4: PLAN B (FALLBACK)                                           │
│                                                                     │
│ Si la red social es NUEVA (Bluesky, Mastodon, Threads, etc.):      │
│ → Aplica formato de título (primera letra de cada palabra en may.) │
│                                                                     │
│ "bluesky" → "Bluesky"                                              │
│ "mi blog personal" → "Mi Blog Personal"                             │
│ "" (vacío) → "" (vacío)                                            │
│ "    " (solo espacios) → "" (vacío)                                 │
│                                                                     │
│ ¿Por qué no devolver error?                                         │
│ Porque no queremos que el sistema se caiga si aparece una red      │
│ nueva. Mejor guardamos algo que se vea bonito aunque no tenga      │
│ un icono asociado.                                                 │
└──────────────────────────────────────────────────────────────────────┘
│
▼
📤 Salida: el nombre oficial de la red social
```

### 🧪 Cómo probarla manualmente

```go
package main

import (
    "fmt"
    "tu-proyecto/api/internal/utils"
)

func main() {
    tests := []string{
        "ig",                        // → Instagram
        "   INSTA   ",               // → Instagram
        "facbook",                   // → Facebook (error ortográfico)
        "x",                         // → X (Twitter)
        "https://youtu.be/abc123",   // → YouTube
        "instagram.com/psico",       // → Instagram
        "fb.com/grupo",              // → Facebook
        "threads",                   // → Threads (nueva)
        "mi blog personal",          // → Mi Blog Personal
        "",                          // → ""
        "    ",                      // → ""
    }

    for _, t := range tests {
        resultado := utils.NormalizePlatformName(t)
        fmt.Printf("Entrada: %-30q → Salida: %q\n", t, resultado)
    }
}
```

### 🤔 ¿Por qué hay tantas variantes de cada red?

Fijate en los errores ortográficos que están contemplados:

| Red | Variantes | ¿Qué error tienen? |
|---|---|---|
| Instagram | `instagran`, `instgram` | Falta letras |
| Facebook | `facbook`, `facebok`, `fbk` | Mala ortografía, siglas |
| X (Twitter) | `twiter`, `twiiter`, `twttr` | Le falta la "t", le sobra "i" |
| YouTube | `yutube`, `ytb` | Fonético, siglas |
| LinkedIn | `linkdin` | Cómo suena |

Estas variantes no están puestas al azar. **Están basadas en errores reales que cometen los usuarios**. Alguien las vio miles de veces y las agregó una por una.

### ⚠️ Errores comunes

1. **Agregar una red al mapa pero sin minúsculas:** Todas las claves del mapa están en **minúsculas**. Si agregás `"Discord"` con mayúscula, nunca va a coincidir porque el input pasa por `ToLower`.
2. **No actualizar el frontend:** Si agregás una red nueva, el frontend necesita un icono SVG para mostrarla. Si no lo tiene, se va a ver un texto feo.
3. **El fallback con Title Case:** Si el usuario escribe `"MI_RED_NUEVA"`, el Title Case lo convierte en `"Mi_Red_Nueva"` (con guiones bajos). Depende del caso puede quedar feo.

---

## 🔐 `radom_string.go` — El generador de contraseñas imposibles de adivinar

### 📖 ¿Qué hace EXACTAMENTE?

```go
func GenerateSecureRandomString(n int) string {
    var sb strings.Builder
    sb.Grow(n)  // Prepara espacio para n caracteres

    charsetLen := len(key_charset)  // 63 caracteres disponibles

    for i := 0; i < n; i++ {
        randomIndex := rand.IntN(charsetLen)  // Elegí un número al azar del 0 al 62
        sb.WriteByte(key_charset[randomIndex]) // Tomá el carácter en esa posición
    }

    return sb.String()
}
```

### 🔤 El alfabeto (los caracteres que puede generar)

```go
const key_charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
```

**63 caracteres en total:**
- 26 minúsculas: `a` a `z`
- 26 mayúsculas: `A` a `Z`
- 10 números: `0` a `9`
- 1 guión bajo: `_`

### ¿Por qué estos 63 y no otros?

| Carácter | ¿Está? | ¿Por qué? |
|---|---|---|
| `a-z` | ✅ | Letras básicas |
| `A-Z` | ✅ | Letras mayúsculas |
| `0-9` | ✅ | Números |
| `_` | ✅ | Útil para separar, no se confunde |
| `-` (guión) | ❌ | En URLs puede confundirse con signo negativo |
| `.` (punto) | ❌ | En SQL, las tablas se llaman `tabla.campo` |
| `!@#$%^&*` | ❌ | En URLs se convierten a `%21`, `%40`, etc. |
| `'\"<>` | ❌ | Peligrosos para inyección SQL/HTML |

### 🎰 La magia de `math/rand/v2` (ChaCha8)

```go
randomIndex := rand.IntN(charsetLen)
```

**Antes de Go 1.22:** Para generar números aleatorios había que:
```go
rand.Seed(time.Now().UnixNano())  // ← Si olvidabas esto, siempre generaba los mismos números
randomIndex := rand.Intn(charsetLen)
```

Si dos servidores arrancaban en el mismo segundo, generaban los **mismos tokens**. ¡Imaginate que los tokens de sesión de todos los usuarios sean iguales! 😱

**Con Go 1.22+ (`math/rand/v2`):**
```go
randomIndex := rand.IntN(charsetLen)  // ← No necesita semilla, es seguro desde el inicio
```

Usa **ChaCha8**:
1. No necesita "semilla" inicial.
2. Es rápido como un rayo.
3. Es seguro: no se puede predecir el próximo número aunque sepas los anteriores.
4. Funciona en paralelo: 1000 usuarios pidiendo tokens al mismo tiempo no se bloquean entre sí.

### ⚡ ¿Por qué `sb.Grow(n)` es importante?

Sin `Grow(n)`:

```
Pedís un token de 32 caracteres.

strings.Builder empieza con espacio para... 0 caracteres.
Primer carácter: necesita espacio → pide 8 bytes.
Segundo carácter: entra en los 8, ok.
...
Noveno carácter: se acabaron los 8, pide 16 más (24 total).
...
Vigésimo quinto carácter: se acabaron los 24, pide 32 más (56 total).
→ 3 re-asignaciones de memoria
```

Con `Grow(32)`:

```
strings.Builder empieza con espacio para 32 caracteres exactos.

Todos los 32 caracteres entran sin problema.
→ 1 sola asignación de memoria
```

En una app chica no se nota. En una app que genera 10,000 tokens por minuto, la diferencia es **enorme**.

### 🧪 Cómo probarla manualmente

```go
package main

import (
    "fmt"
    "tu-proyecto/api/internal/utils"
)

func main() {
    // Generar tokens de diferentes longitudes
    fmt.Println("Token de 8 caracteres:  ", utils.GenerateSecureRandomString(8))
    fmt.Println("Token de 16 caracteres: ", utils.GenerateSecureRandomString(16))
    fmt.Println("Token de 32 caracteres: ", utils.GenerateSecureRandomString(32))
    fmt.Println("Token de 64 caracteres: ", utils.GenerateSecureRandomString(64))

    // Verificar que dos tokens seguidos son distintos
    t1 := utils.GenerateSecureRandomString(16)
    t2 := utils.GenerateSecureRandomString(16)
    fmt.Println("\n¿Son distintos?", t1 != t2)  // → true (casi siempre)
    fmt.Println("Token 1:", t1)
    fmt.Println("Token 2:", t2)
}
```

### 🤔 ¿Qué tan probable es que dos tokens sean iguales?

Con tokens de 32 caracteres y un alfabeto de 63 caracteres:

- Posibles combinaciones: 63³² ≈ 7.6 × 10⁵⁷
- Probabilidad de colisión: esencialmente **cero**.

Para que te des una idea:
- Hay más combinaciones posibles que **átomos en el universo**.
- Si generaras mil millones de tokens por segundo, tardarías más que la edad del universo en tener una colisión.

### ⚠️ Errores comunes

1. **Pasar un número negativo:** Si `n` es negativo, `sb.Grow(n)` no funciona bien. Siempre pasá números positivos.
2. **Pasar 0:** Genera un string vacío. No es un error, pero no sirve para nada.
3. **Creer que es criptográficamente seguro:** `math/rand/v2` es "criptográficamente seguro" (*CSPRNG*). Está bien para tokens de sesión. Si necesitás algo **ultra seguro** (como claves de encriptación bancaria), usá `crypto/rand`.
4. **Guardar tokens sin hash:** Si guardás estos tokens directamente en la base de datos (sin aplicar bcrypt primero), si alguien roba la base de datos tiene todos los tokens de sesión. ¡Siempre hashealos!

---

## 🔒 `secure_password.go` — El portero que revisa 5 cosas antes de dejar pasar

### 📖 ¿Qué hace EXACTAMENTE?

```go
func IsStrongPassword(password string) bool {
    // Estos son los 5 "chequeos" que vamos a hacer
    var (
        hasMinLen  = false   // 1. ¿Tiene al menos 8 caracteres?
        hasUpper   = false   // 2. ¿Tiene al menos 1 mayúscula?
        hasLower   = false   // 3. ¿Tiene al menos 1 minúscula?
        hasNumber  = false   // 4. ¿Tiene al menos 1 número?
        hasSpecial = false   // 5. ¿Tiene al menos 1 símbolo?
    )

    // ───────────────────────────────────────────────────────
    // PRIMERO: revisar el largo
    // ───────────────────────────────────────────────────────
    if len(password) >= 8 {
        hasMinLen = true
    }

    // ───────────────────────────────────────────────────────
    // SEGUNDO: buscar espacios (y si hay, rechazar AL TOQUE)
    // ───────────────────────────────────────────────────────
    for _, char := range password {
        if unicode.IsSpace(char) {  // ¿Es espacio, tab, enter?
            return false             // ❌ FUERA. No gastes más tiempo.
        }
    }

    // ───────────────────────────────────────────────────────
    // TERCERO: analizar cada carácter
    // ───────────────────────────────────────────────────────
    for _, char := range password {
        switch {
        case unicode.IsUpper(char):                    // ¿Mayúscula?
            hasUpper = true
        case unicode.IsLower(char):                    // ¿Minúscula?
            hasLower = true
        case unicode.IsNumber(char):                   // ¿Número?
            hasNumber = true
        case unicode.IsPunct(char) || unicode.IsSymbol(char):  // ¿Símbolo?
            hasSpecial = true
        }
        // Si no es nada de eso (ej. espacio), no hacemos nada (lo ignoramos)
    }

    // ───────────────────────────────────────────────────────
    // CUARTO: devolver true SOLO si pasa TODOS los filtros
    // ───────────────────────────────────────────────────────
    return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}
```

### 🔍 Los 5 filtros, uno por uno

#### Filtro 1: Longitud mínima (`len(password) >= 8`)

| Largo de contraseña | Tiempo para romperla (fuerza bruta) |
|---|---|
| 4 caracteres | Menos de 1 segundo ⚡ |
| 6 caracteres | ~ 2 minutos |
| 8 caracteres | ~ 4 horas |
| 10 caracteres | ~ 3 meses |
| 12 caracteres | ~ 200 años |
| 16 caracteres | ~ 5 millones de años |

8 es el mínimo que **OWASP** (la organización mundial de seguridad de software) recomienda.

#### Filtro 2: Sin espacios (`unicode.IsSpace(char)`)

```go
for _, char := range password {
    if unicode.IsSpace(char) {
        return false  // ¡Corchazo! Ni miramos el resto
    }
}
```

Este filtro está **antes** del análisis de composición. Es un **Fail-Fast** (fallo rápido): si la contraseña tiene espacios, no tiene sentido seguir analizando.

**¿Por qué están prohibidos los espacios?**

1. **HTTP Basic Auth:** La cabecera de autenticación usa el formato `usuario:contraseña`. Si la contraseña tiene un espacio, se rompe.
2. **Auto-correción de móviles:** Los teléfonos a veces agregan un espacio después de autocompletar. El usuario cree que su clave es `"Pass123!"` pero guardó `"Pass123! "`.
3. **Soporte técnico:** "No puedo iniciar sesión 😭" → "¿Su contraseña tiene espacios?" → "Aaah, sí, tenía uno atrás". Horas de debug ahorradas.

#### Filtro 3: Mayúscula (`unicode.IsUpper`)

```go
case unicode.IsUpper(char):
    hasUpper = true
```

¿Reconoce `Ñ`? Sí. `unicode.IsUpper('Ñ')` → `true`. ¿Reconoce `Ç`? Sí. Cualquier letra mayúscula de cualquier idioma.

#### Filtro 4: Minúscula (`unicode.IsLower`)

```go
case unicode.IsLower(char):
    hasLower = true
```

Ídem.

#### Filtro 5: Número (`unicode.IsNumber`)

```go
case unicode.IsNumber(char):
    hasNumber = true
```

Cubre `0-9` y dígitos de otros alfabetos (ej. dígitos árabes `٠١٢٣٤٥٦٧٨٩`).

#### Filtro 6: Especial (`unicode.IsPunct || unicode.IsSymbol`)

```go
case unicode.IsPunct(char) || unicode.IsSymbol(char):
    hasSpecial = true
```

Cubre:
- Puntuación (`! @ # $ % ^ & * ( ) - + = { } [ ] : ; " ' < > , . ? / \ | ~`)
- Símbolos (`© ® ™ € ¥ £ ¢ ∞ § ¶ •`)

### 🧪 Todos los casos de prueba, explicados uno por uno

```go
// 1. ✅ Válida
{"Pass1234!", true}
//    Largo: 9 ✅, May: P ✅, Min: a ✅, Núm: 1 ✅, Esp: ! ✅

// 2. ❌ Corta
{"P1s@2", false}
//    Largo: 5 ❌ (falla acá)

// 3. ❌ Sin mayúscula
{"password123!", false}
//    Largo: 12 ✅, May: ❌ (no hay ninguna)

// 4. ❌ Sin minúscula
{"PASSWORD123!", false}
//    Largo: 12 ✅, May: ✅, Min: ❌

// 5. ❌ Sin número
{"Password!", false}
//    Largo: 9 ✅, May: ✅, Min: ✅, Núm: ❌

// 6. ❌ Sin especial
{"Password123", false}
//    Largo: 11 ✅, May: ✅, Min: ✅, Núm: ✅, Esp: ❌

// 7. ✅ Con acentos (español)
{"Contraseña9$", true}
//    Largo: 12 ✅, May: C ✅, Min: o ✅, Núm: 9 ✅, Esp: $ ✅
//    La 'ñ' y la 'a' con acento son letras normales para Go.

// 8. ❌ Con espacios
{"Strong Password 1", false}
//    Tiene espacio → Fail-Fast → return false

// 9. ❌ Vacía
{"", false}
//    Largo: 0 ❌
```

### 🧪 Cómo probarla manualmente

```go
package main

import (
    "fmt"
    "tu-proyecto/api/internal/utils"
)

func main() {
    tests := []string{
        "Pass1234!",
        "P1s@2",
        "password123!",
        "PASSWORD123!",
        "Password!",
        "Password123",
        "Contraseña9$",
        "Strong Password 1",
        "",
    }

    for _, t := range tests {
        resultado := utils.IsStrongPassword(t)
        marca := "✅"
        if !resultado { marca = "❌" }
        fmt.Printf("%s %q → %v\n", marca, t, resultado)
    }
}
```

### ⚠️ Errores comunes

1. **Usar `len(password)` con caracteres UTF-8:** `len("ñ")` devuelve 2 (porque son 2 bytes). Pero en Go, `len("ñ")` sobre un string cuenta **bytes**, no caracteres. Si la contraseña tiene muchos acentos, `len` puede dar un número mayor a la cantidad real de caracteres. Para este uso no es problema porque 8 bytes de seguridad son suficientes, pero tenelo en cuenta.
2. **Confundir `unicode.IsLetter` con `unicode.IsUpper`:** `IsLetter` te dice si es una letra (cualquier letra). `IsUpper` te dice si es una letra **mayúscula**.
3. **El espacio como carácter especial:** El espacio NO es carácter especial para `unicode.IsPunct`. El espacio es `unicode.IsSpace`.

---

## 📧 `validate_email.go` — El cartero digital

### 📖 ¿Qué hace EXACTAMENTE?

```go
func ParseAndValidateEmail(email string) (string, error) {
    // PASO 1: Sacar espacios de los bordes
    email = strings.TrimSpace(email)

    // PASO 2: Si queda vacío, error
    if email == "" {
        return "", errors.New("el correo electrónico no puede estar vacío")
    }

    // PASO 3: Validar formato con el parser oficial de Go
    addr, err := mail.ParseAddress(email)
    if err != nil {
        return "", errors.New("formato de correo electrónico inválido")
    }

    // PASO 4: Pasar todo a minúsculas y devolver
    cleanEmail := strings.ToLower(addr.Address)
    return cleanEmail, nil
}
```

### 🔍 Paso a paso, con ejemplos reales

#### Paso 1: `strings.TrimSpace(email)` — Sacar espacios invisibles

```
"  fran@mail.com  "  →  "fran@mail.com"
"fran@mail.com"      →  "fran@mail.com"  (no tenía)
```

¿Por qué? Los teclados de los celulares **autocompletan direcciones** y a veces agregan un espacio al final sin que el usuario se dé cuenta. Sin `TrimSpace`, el usuario escribiría `"fran@mail.com "` (con espacio al final) y el sistema diría "correo inválido". El usuario se frustraría 😤.

#### Paso 2: Verificar que no esté vacío

`""` → ❌ Error: "el correo no puede estar vacío".

#### Paso 3: `mail.ParseAddress(email)` — La validación profesional

```go
addr, err := mail.ParseAddress(email)
```

`mail.ParseAddress` es parte de la biblioteca estándar de Go. Sigue el estándar **RFC 5322** (el documento oficial que define cómo son los correos electrónicos).

**¿Qué puede hacer que un simple regex no puede?**

| Característica | Regex básica | `mail.ParseAddress` |
|---|---|---|
| `user+tag@example.com` | ❌ Lo rechaza | ✅ El + es válido |
| `"user name"@example.com` | ❌ Espacio en la parte local | ✅ Con comillas es válido |
| `user@[192.168.1.1]` | ❌ Dominio como IP | ✅ Válido |
| `user@sub.dominio.co.ve` | ❌ Muchos subdominios | ✅ Válido |
| `user@example` (sin .com) | ❌ Lo rechaza | ✅ Válido (correos internos) |

#### ¿Por qué `mail.ParseAddress` y no una regex?

**Problema de seguridad: ReDoS**

Un atacante puede enviar:
```
aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa!@a.com
```
y una regex mal hecha se **traba** consumiendo toda la CPU del servidor. Es como si le pidieras a alguien que lea una palabra de 100,000 letras: se va a fatigar y no va a poder hacer nada más.

`mail.ParseAddress` usa un **parser determinista**: siempre procesa en tiempo proporcional al largo del texto (O(n)). No importa si el texto es corto o larguísimo, no se traba.

#### Paso 4: `strings.ToLower(addr.Address)` — Todo a minúsculas

```go
cleanEmail := strings.ToLower(addr.Address)
```

**¿Por qué?**

En la base de datos:
- `"Fran@Mail.com"` y `"fran@mail.com"` son **strings diferentes**.
- Si hay un índice `UNIQUE` en el email, se pueden crear **dos cuentas** para el mismo correo.
- El usuario pone su correo como `"Fran@Mail.com"` al registrarse, pero al iniciar sesión escribe `"fran@mail.com"` y el sistema dice "usuario no encontrado".

Pasando todo a minúsculas, estos problemas desaparecen.

### 🧪 ¿Qué correos acepta y cuáles no?

| Correo | Resultado | Explicación |
|---|---|---|
| `"user@example.com"` | ✅ `"user@example.com"` | Normal |
| `"Fran@Mail.com"` | ✅ `"fran@mail.com"` | Pasa a minúsculas |
| `"  user@example.com  "` | ✅ `"user@example.com"` | Saca espacios |
| `"user+tag@example.com"` | ✅ `"user+tag@example.com"` | El + está permitido |
| `""` | ❌ Error "no puede estar vacío" | No hay nada que validar |
| `"not-an-email"` | ❌ Error "formato inválido" | No tiene @ |
| `"@example.com"` | ❌ Error "formato inválido" | No tiene usuario |
| `"user@"` | ❌ Error "formato inválido" | No tiene dominio |
| `"a@b"` | ✅ `"a@b"` | Técnicamente válido (correo interno) |
| `"user@.com"` | ❌ Error "formato inválido" | Dominio incompleto |

### 🧪 Cómo probarla manualmente

```go
package main

import (
    "fmt"
    "tu-proyecto/api/internal/utils"
)

func main() {
    tests := []string{
        "user@example.com",
        "Fran@Mail.com",
        "  user@example.com  ",
        "user+tag@example.com",
        "",
        "not-an-email",
        "@example.com",
        "user@",
    }

    for _, t := range tests {
        resultado, err := utils.ParseAndValidateEmail(t)
        if err != nil {
            fmt.Printf("❌ %q → ERROR: %v\n", t, err)
        } else {
            fmt.Printf("✅ %q → %q\n", t, resultado)
        }
    }
}
```

### ⚠️ Errores comunes

1. **Pensar que `mail.ParseAddress` acepta cualquier cosa:** No acepta correos sin `@`, ni con caracteres absolutamente prohibidos. Solo sigue el estándar RFC 5322.
2. **Olvidar que el resultado está en minúsculas:** Si después comparás el email con `==` contra una versión que no está en minúsculas, va a fallar.
3. **No revisar el error:** La función devuelve `(string, error)`. Si ignorás el error y usás el string igual, podés tener un string vacío o un correo inválido.

---

## 🧪 `utils_test.go` — Las pruebas que cuidan tu espalda

### 📖 ¿Qué es esto?

Un archivo que **no corre en producción**. Solo se ejecuta cuando hacés `go test`. Su trabajo es verificar que todas las funciones de arriba funcionen correctamente, **siempre**.

Cada vez que alguien modifica una función, corre `go test` y si algo se rompe, los tests lo detectan.

### 🔍 Cada prueba explicada

#### `TestSanitizeImage_Defensive`

```go
func TestSanitizeImage_Defensive(t *testing.T) {
    fakeImage := strings.NewReader("Esto no es un JPEG, es texto para engañar al sistema")
    bytes, ext, contentType, err := SanitizeImage(fakeImage)

    require.Error(t, err, "Debe rechazar archivos que no sean imágenes reales")
    require.Nil(t, bytes)
    require.Empty(t, ext)
    require.Empty(t, contentType)
}
```

**¿Qué prueba?** Que si le pasás **texto cualquiera** (no una imagen real), la función rechace el archivo.

**¿Cómo lo hace?**
1. Crea un "archivo falso" que en realidad es texto: `"Esto no es un JPEG..."`.
2. Lo pasa por `SanitizeImage`.
3. Verifica que devuelva error ✅.
4. Verifica que los bytes, extensión y MIME type estén vacíos ✅.

**¿Por qué es importante?** Porque los hackers pueden tomar un archivo que se llame `foto.jpg` pero que por dentro contenga código PHP. Si `SanitizeImage` no detecta el engaño, el código malicioso se guarda en el servidor y pueden hackear el sitio.

> [!WARNING]
> **⚠️ Problema detectado:** El test en la línea 28 dice:
> ```go
> require.Contains(t, err.Error(), "archivo no es una imagen válida")
> ```
> Pero en el código real (`image_sanitizer.go:86`), el mensaje de error es:
> ```go
> errors.New("el servidor no reconoce este formato de imagen")
> ```
> **Son mensajes diferentes.** Si corrés `go test` hoy, **puede fallar** porque el test espera una cosa y el código devuelve otra.
>
> Posibles soluciones:
> - ✏️ Actualizar el test para que busque el mensaje real.
> - ✏️ Actualizar el código para que use el mensaje del test.
> - ✏️ O ambos están bien y alguien olvidó actualizar uno de los dos.

#### `TestIsEmptyReq`

```go
func TestIsEmptyReq(t *testing.T) {
    cool := true
    tests := []struct {
        name     string
        input    interface{}
        expected bool
    }{
        {"Struct totalmente vacío", DummyStruct{}, true},
        {"Struct con un string", DummyStruct{Name: "Fran"}, false},
        {"Struct con un int válido", DummyStruct{Age: 25}, false},
        {"Struct con puntero inicializado", DummyStruct{IsCool: &cool}, false},
        {"Puntero a struct vacío", &DummyStruct{}, true},
        {"Puntero a struct con datos", &DummyStruct{Name: "Admin"}, false},
    }
    // ... ejecuta cada caso
}
```

**¿Qué prueba?** 6 escenarios diferentes de `IsEmptyReq`:

| # | Escenario | ¿Qué se prueba? |
|---|---|---|
| 1 | Struct vacío → `true` | Sin datos = vacío ✅ |
| 2 | Struct con string → `false` | String con texto = no vacío ✅ |
| 3 | Struct con int → `false` | Número diferente de 0 = no vacío ✅ |
| 4 | Struct con puntero → `false` | Puntero no nil = no vacío ✅ |
| 5 | Puntero a struct vacío → `true` | Puntero a struct vacío = vacío ✅ |
| 6 | Puntero con datos → `false` | Puntero con datos = no vacío ✅ |

#### `TestNormalizePlatformName`

```go
// 9 sub-pruebas:
{"Instagram diminutivo", "ig", "Instagram"},                   // Mapa directo
{"Instagram mayúsculas con espacios", "   INSTA   ", "Instagram"}, // Limpieza
{"Twitter moderno", "x", "X (Twitter)"},                        // Twitter renombrado
{"Facebook error ortográfico", "facbook", "Facebook"},          // Error común
{"URL de Youtube", "https://youtu.be/mivideo", "YouTube"},      // URL completa
{"URL de Instagram", "instagram.com/psico", "Instagram"},       // URL con subcadena
{"URL de Facebook", "fb.com/grupo", "Facebook"},                // fb.com
{"Plataforma nueva", "threads", "Threads"},                     // Fallback
{"Compuesto no mapeado", "mi blog personal", "Mi Blog Personal"}, // Title Case
{"Vacío", "", ""},                                              // Vacío
{"Solo espacios", "    ", ""},                                  // Solo espacios
```

#### `TestGenerateSecureRandomString`

Tres sub-pruebas:

1. **Verificar longitud:** Pide un token de 32 caracteres y verifica que tenga exactamente 32. Si tiene 31 o 33, algo está mal.
2. **Verificar aleatoriedad (anti-colisión):** Genera dos tokens de 16 caracteres y verifica que sean diferentes. Si son iguales, el generador no es aleatorio.
3. **Verificar caracteres permitidos:** Genera un token de 100 caracteres y revisa que cada carácter esté dentro del charset `key_charset`. Si aparece un `%` o un `&`, el código tiene un error.

#### `TestIsStrongPassword`

```go
// 9 sub-pruebas (explicadas en detalle arriba en la sección 🔒):
{"Password válida con todos los criterios", "Pass1234!", true},           // ✅
{"Inválida por ser muy corta (< 8)", "P1s@2", false},                    // ❌ Corta
{"Inválida por falta de mayúscula", "password123!", false},              // ❌ Sin mayúscula
{"Inválida por falta de minúscula", "PASSWORD123!", false},              // ❌ Sin minúscula
{"Inválida por falta de número", "Password!", false},                    // ❌ Sin número
{"Inválida por falta de carácter especial", "Password123", false},       // ❌ Sin especial
{"Válida con caracteres UTF-8 (acentos/símbolos)", "Contraseña9$", true}, // ✅ Con ñ
{"Válida con espacios (el espacio cuenta como símbolo/puntuación)", "Strong Password 1", false}, // ❌ Espacio
{"String vacío", "", false},                                              // ❌ Vacía
```

### 🧪 Cómo correr las pruebas

```bash
# Ir a la carpeta donde están los tests
cd api/internal/utils

# Correr todos los tests
go test -v

# Correr un test específico
go test -v -run TestIsStrongPassword

# Ver cuánto cubren los tests (cobertura)
go test -cover
```

### 📊 ¿Qué funciones TIENEN tests y cuáles NO?

| Función | ¿Tiene test? | Archivo del test |
|---|---|---|
| `SanitizeImage` | ✅ Sí | `utils_test.go` |
| `SanitizeDocument` | ❌ No | — |
| `FlattenAlpha` | ❌ No | — |
| `CleanAlphaNumeric` | ❌ No | — |
| `NormalizeMunicipioCarabobo` | ❌ No | — |
| `NormalizeEstadoVenezuela` | ❌ No | — |
| `BoolFromForm` | ❌ No | — |
| `IsEmptyReq` | ✅ Sí | `utils_test.go` |
| `NormalizePlatformName` | ✅ Sí | `utils_test.go` |
| `GenerateSecureRandomString` | ✅ Sí | `utils_test.go` |
| `IsStrongPassword` | ✅ Sí | `utils_test.go` |
| `ParseAndValidateEmail` | ❌ No | — |

> **💡 Si querés contribuir:** Agregar tests para las funciones que no tienen es una excelente manera de empezar. Buscá las que están marcadas con ❌ y escribí pruebas similares a las existentes.

---

## 🚀 Resumen ejecutivo (si no leíste nada más, leé esto)

### Las 11 funciones, en 11 líneas

```go
utils.PrintColpsiASCII()                      // 🎨 Muestra el logo del colegio
utils.CleanAlphaNumeric("texto")              // 🧹 Deja solo letras y números
utils.NormalizeMunicipioCarabobo("texto")     // 🌆 Normaliza un municipio de Carabobo
utils.NormalizeEstadoVenezuela("texto")       // 🇻🇪 Normaliza un estado de Venezuela
utils.BoolFromForm("1")                       // ✅ Convierte "true"/"1" a *bool
utils.SanitizeImage(archivo)                  // 🖼️ Comprime y limpia imágenes
utils.SanitizeDocument(archivo)               // 📄 Comprime y limpia documentos
utils.IsEmptyReq(struct{})                    // 🕵️ ¿El struct está vacío?
utils.NormalizePlatformName("ig")             // 🔤 "ig" → "Instagram"
utils.GenerateSecureRandomString(32)          // 🔐 Genera un token aleatorio
utils.IsStrongPassword("Pass1234!")           // 🔒 ¿La contraseña es segura?
utils.ParseAndValidateEmail("a@b.com")        // 📧 ¿El email es válido?
```

### Las funciones que necesitan tests (🚧)

| Función | Estado | ¿Querés ayudar? |
|---|---|---|
| `SanitizeDocument` | ❌ Sin test | Hacé uno similar a `TestSanitizeImage` |
| `FlattenAlpha` | ❌ Sin test | Probá con imágenes PNG transparentes |
| `CleanAlphaNumeric` | ❌ Sin test | Probá con caracteres especiales y ñ |
| `NormalizeMunicipioCarabobo` | ❌ Sin test | Probá con mayúsculas, minúsculas, tildes |
| `NormalizeEstadoVenezuela` | ❌ Sin test | Ídem, con todos los estados |
| `BoolFromForm` | ❌ Sin test | Probá "1", "true", "yes", "0", etc. |
| `ParseAndValidateEmail` | ❌ Sin test | Probá correos válidos e inválidos |

### ⚠️ Cosas para tener en cuenta

1. **Posible bug en `TestSanitizeImage`:** El test espera el error `"archivo no es una imagen válida"` pero el código devuelve `"el servidor no reconoce este formato de imagen"`. Revisá cuál es el correcto.

2. **Carabobo no está en `estadosVenezuela`:** Es intencional. Carabobo se maneja aparte porque es la sede del colegio.

3. **Los espacios están prohibidos en contraseñas:** Por seguridad y UX. No es un error, es una decisión de diseño.

4. **Todas las funciones son "puras":** No modifican nada externo, no tocan la base de datos, no hacen llamadas de red. Son predecibles y seguras.

5. **Usá siempre `go test` antes de hacer un commit:** Te ahorra que te rompan el código en producción.
