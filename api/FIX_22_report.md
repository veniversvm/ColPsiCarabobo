# FIX-22: Reemplazar println() nativo por log.Printf

## Metadatos

| Campo | Valor |
|-------|-------|
| **FIX** | FIX-22 |
| **Hallazgo** | MED-03 |
| **Severidad** | MEDIO |
| **Archivos** | `psi_service.go`, `psi_user_admin_service.go`, `main.go` |
| **Estado** | COMPLETADO |
| **Fecha** | 2026-07-25 |

---

## Problema

`println()` es una función built-in de Go que escribe a stderr sin timestamps, sin niveles de log, y sin estructura. Los mensajes son invisibles en agregadores de logs de producción. Los comentarios del código decían "logueamos el error" pero usaban `println()` en vez de `log.Printf`.

---

## Cambios Realizados

| # | Archivo | Línea | Antes | Después |
|---|---------|-------|-------|---------|
| 1 | `psi_service.go` | 636 | `println("Error al sincronizar...")` | `log.Printf("WARN: Error al sincronizar actualización con Audiobookshelf: %v", absErr)` |
| 2 | `psi_service.go` | 1101 | `println("Error sincronizando...")` | `log.Printf("WARN: Error sincronizando con Audiobookshelf: %v", absErr)` |
| 3 | `psi_service.go` | 1106 | `println("Usuario creado...")` | `log.Printf("INFO: Usuario creado en Audiobookshelf con ID: %s", absID)` |
| 4 | `psi_service.go` | 1108 | `println("El usuario ya existía...")` | `log.Printf("INFO: El usuario ya existía en Audiobookshelf, no se generó un nuevo ID.")` |
| 5 | `psi_user_admin_service.go` | 690 | `println("Error al sincronizar...")` | `log.Printf("WARN: Error al sincronizar actualización del administrador con Audiobookshelf: %v", absErr)` |
| 6 | `main.go` | 49 | `println("Intentando cargar...")` | `log.Printf("INFO: Intentando cargar configuración...")` |
| 7 | `main.go` | 59 | `println("Configuración cargada...")` | `log.Printf("INFO: Configuración cargada. Intentando conectar a DB...")` |

---

## Verificación

- `rg 'println\('` en los 3 archivos → 0 resultados ✅
- `go build ./...` → sin errores ✅
- `go vet ./...` → sin warnings ✅
- Import `log` ya existente en los 3 archivos → sin cambios de import necesarios ✅
