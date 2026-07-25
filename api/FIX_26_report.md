# FIX-26: Eliminar logging de PII

## Metadatos

| Campo | Valor |
|-------|-------|
| **FIX** | FIX-26 |
| **Hallazgo** | MED-07 |
| **Severidad** | MEDIO (CRITICAL por hallazgo) |
| **Archivos** | `psi_handler.go`, `psi_service_xlsx.go`, `mail_service.go` |
| **Estado** | COMPLETADO |
| **Fecha** | 2026-07-25 |

---

## Problema

Múltiples log statements exponían PII (emails, usernames, objetos completos) en logs de producción:
- `psi_handler.go:234-235` — `log.Printf("%v", psi)` volcaba el struct completo (nombres, CI, emails, teléfonos, direcciones, passwords via reflection)
- `psi_service_xlsx.go:278` — logueaba email + username en cada error de envío
- `mail_service.go:104,106,151` — logueaba la dirección email del destinatario en cada correo enviado

---

## Solución

1. Creada función `maskEmail()` en `mail_service.go` que enmascara emails: `"j***@e****.com"`
2. Reemplazados todos los log statements que exponían PII

---

## Cambios Realizados

| # | Archivo | Línea | Antes | Después |
|---|---------|-------|-------|---------|
| 1 | `mail_service.go` | — | — | Nueva función `maskEmail()` |
| 2 | `mail_service.go` | 104 | `log.Printf("... a %s: %v", job.To, err)` | `log.Printf("... a %s: %v", maskEmail(job.To), err)` |
| 3 | `mail_service.go` | 106 | `log.Printf("... %s", job.To)` | `log.Printf("... %s", maskEmail(job.To))` |
| 4 | `mail_service.go` | 151 | `log.Printf("... para %s", to)` | `log.Printf("... %s, subject=%s", maskEmail(to), subject)` |
| 5 | `psi_handler.go` | 234-235 | `log.Printf("%v", psi)` (struct completo) | `log.Printf("PSI Profile loaded: id=%s, username=%s", psi.ID, psi.Username)` |
| 6 | `psi_service_xlsx.go` | 278 | `log.Printf("... %s [%s]: %v", psi.Email, psi.Username, err)` | `log.Printf("... user_id=%s: %v", psi.ID, err)` |

---

## Verificación

- `rg 'log\.Printf.*%v.*psi\b'` → solo coincide la versión corregida (usa `psi.ID`) ✅
- `rg 'job\.To' mail_service.go` → todas las ocurrencias en logs usan `maskEmail()` ✅
- `go build ./...` → sin errores ✅
