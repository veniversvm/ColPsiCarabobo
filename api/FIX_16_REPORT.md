# Reporte FIX-16 — Email Templates + MustChangePassword

**Fecha:** 25 de Julio, 2026
**Severidad:** 🟠 ALTO

---

## Hallazgos

### H1: Bug funcional — Password vacío en emails
Los templates usaban `{{.TempPassword}}` pero el código pasaba `"Password"` en el data map.
**Resultado:** El password se renderizaba como string vacío en todos los emails de bienvenida.

### H2: Sin mecanismo de cambio de contraseña obligatorio
No existía `MustChangePassword` en el modelo. Los usuarios podían usar contraseñas temporales indefinidamente.

### H3: Sin validación de fuerza de contraseña en CreatePsiByAdmin
`CreateAdmin` validaba con `IsStrongPassword()`, pero `CreatePsiByAdmin` no.

---

## Solución implementada

### 1. Templates HTML corregidos

**welcome_psi.html** y **welcome_admin.html**:
- `{{.TempPassword}}` → `{{.Password}}` (key consistente con el código)
- Agregado aviso amarillo: "IMPORTANTE: Esta es una contraseña temporal..."

### 2. MustChangePassword en Credentials

```go
// domain/credentials.go
MustChangePassword bool `gorm:"column:must_change_password;default:false" json:"-"`
```

`json:"-"` — nunca se expone en APIs externas.

### 3. Marcado automático al crear usuarios

| Path | MustChangePassword |
|------|-------------------|
| `ImportFromCSV` | `true` |
| `ImportFromXLSX` | `true` |
| `CreatePsiByAdmin` | `true` |
| `CreateAdmin` | `false` (el admin conoce la contraseña) |

### 4. Validación de password en CreatePsiByAdmin

```go
if !utils.IsStrongPassword(req.Password) {
    return errors.New("la contraseña no cumple con los estándares de seguridad")
}
```

### 5. Login retorna must_change_password

**AdminService.Login** — firma cambiada:
```go
// Antes:
func Login(...) (string, error)
// Después:
func Login(...) (string, *domain.UserAdmin, error)
```

**Response del handler:**
```json
{
    "message": "Bienvenido al sistema",
    "token": "eyJ...",
    "must_change_password": true
}
```

### 6. Migración SQL

```sql
ALTER TABLE user_admins ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN DEFAULT false;
ALTER TABLE psi_users ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN DEFAULT false;
```

---

## Archivos modificados

| Archivo | Cambio |
|---------|--------|
| `internal/templates/welcome_psi.html` | `{{.TempPassword}}` → `{{.Password}}` + aviso |
| `internal/templates/welcome_admin.html` | `{{.TempPassword}}` → `{{.Password}}` + aviso |
| `internal/domain/credentials.go` | Nuevo campo `MustChangePassword` |
| `internal/service/psi_user_admin_service.go` | `MustChangePassword: true` + validación password |
| `internal/service/psi_service.go` | `MustChangePassword: true` en CSV |
| `internal/service/psi_service_xlsx.go` | `MustChangePassword: true` en XLSX |
| `internal/service/admin_service.go` | Login retorna `*UserAdmin` |
| `internal/handler/admin_handler.go` | `must_change_password` en response |
| `internal/handler/psi_handler.go` | `must_change_password` en response |
| `internal/service/admin_service_test.go` | Fix firma Login |
| `internal/service/psi_user_admin_service_test.go` | Password fuerte + test rechazo |
| `migrations/20260725020000_fix16_must_change_password.sql` | ALTER TABLE × 2 |

---

## Tests

| Paquete | Resultado |
|---------|-----------|
| `internal/service` | ✅ Todos nuestros tests pasan |
| `internal/domain` | ✅ PASS |
| `internal/middleware` | ✅ PASS |
| `pkg/job` | ✅ PASS |

**2 fallos pre-existentes** (no relacionados):
- `TestAdminService_All/CreateAdmin` — email validation issue
- `TestSpecialtyService_Update` — mock sin `GetByAdminID`
