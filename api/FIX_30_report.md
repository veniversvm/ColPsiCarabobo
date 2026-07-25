# FIX-30: Documentar contrato Read-before-Write para Save()

## Metadatos

| Campo | Valor |
|-------|-------|
| **FIX** | FIX-30 |
| **Hallazgo** | MED-11 |
| **Severidad** | MEDIO |
| **Archivos** | `user_admin_repo.go`, `psi_repository.go` |
| **Estado** | COMPLETADO |
| **Fecha** | 2026-07-25 |

---

## Problema

7 llamadas a `Save()` en repositories potencialmente sobreescriben zero-values si el caller pasa un modelo parcial. Todos los callers actuales hacen Read-before-Write (son seguros en la práctica), pero el contrato no estaba documentado en `user_admin_repo`, `UpdatePostGrade` ni `UpdateSocialNetwork`.

---

## Decisión de Implementación

**No se convirtió `Save()` a `Updates()`** por las siguientes razones:

1. **Comportamiento correcto actual**: todos los callers ya obtienen el modelo completo vía `GetByID` antes de modificar campos
2. **Los zero-values son correctos**: si un campo booleano es `false`, `Save()` lo persiste correctamente — `Updates()` lo ignoraría
3. **`Updates()` con map sería verboso y propenso a errores**: `UserAdmin` tiene 17+ campos de permisos
4. **`specialty_repo.go` ya documentaba el contrato** con una advertencia técnica clara

---

## Solución

Documentación explícita del contrato Read-before-Write en los 3 repos que faltaban:

| Archivo | Método | Cambio |
|---------|--------|--------|
| `user_admin_repo.go` | `Update` | Agregada advertencia técnica sobre zero-values |
| `psi_repository.go` | `UpdatePostGrade` | Agregada advertencia sobre Read-before-Write |
| `psi_repository.go` | `UpdateSocialNetwork` | Agregada advertencia sobre Read-before-Write |

---

## Verificación

- `go build ./...` → sin errores ✅
- Sin cambios de comportamiento, solo documentación defensiva ✅
