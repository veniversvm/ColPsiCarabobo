# FIX-24: Eliminar fmt.Printf/Println de debug en repositorio

## Metadatos

| Campo | Valor |
|-------|-------|
| **FIX** | FIX-24 |
| **Hallazgo** | MED-05 |
| **Severidad** | MEDIO |
| **Archivo** | `internal/repository/postgres/psi_repository.go` |
| **Estado** | COMPLETADO |
| **Fecha** | 2026-07-25 |

---

## Problema

Dos statements `fmt.Printf`/`fmt.Println` con prefijo `### REPO DEBUG:` imprimían información de debug a stdout unconditionalmente. Estos mensajes son invisibles en log aggregators y ensucian la salida del servidor.

---

## Cambios Realizados

| Línea | Antes | Después |
|-------|-------|---------|
| 329 | `fmt.Printf("### REPO DEBUG: Recibidas %d solvencias para procesar\n", len(solvencies))` | Eliminada |
| 344 | `fmt.Println("### REPO DEBUG: Solvencias insertadas/actualizadas con éxito")` | Eliminada |

---

## Verificación

- `rg 'REPO DEBUG' psi_repository.go` → 0 resultados ✅
- `go build ./...` → sin errores ✅
- Import `fmt` se mantiene (12 usos restantes en el archivo) ✅
