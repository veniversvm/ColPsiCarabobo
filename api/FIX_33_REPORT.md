# FIX-33 Report — Test de imagen con mensaje incorrecto

| Campo | Valor |
|-------|-------|
| **Fix ID** | FIX-33 |
| **Archivo modificado** | `internal/utils/utils_test.go` |
| **Fecha de implementación** | 2026-07-25 |
| **Estado** | Completado |

---

## Problema

El test `TestSanitizeImage_Defensive` esperaba un mensaje de error que ya no coincidía con el código actual:

```go
// Test (línea 28):
require.Contains(t, err.Error(), "archivo no es una imagen válida")

// Código (image_sanitizer.go:86):
return nil, "", "", errors.New("el servidor no reconoce este formato de imagen")
```

Alguien cambió el mensaje de error en `image_sanitizer.go` sin actualizar el test.

---

## Corrección

Actualizar el test para coincidir con el código actual:

```go
// DESPUÉS:
require.Contains(t, err.Error(), "el servidor no reconoce este formato de imagen")
```

---

## Testing

- Utils tests: 4/4 Pass (antes 4/5)

---

## Archivos relacionados

| Archivo | Relación |
|---------|----------|
| `internal/utils/utils_test.go:28` | Test corregido |
| `internal/utils/image_sanitizer.go:86` | Mensaje de error actual |
