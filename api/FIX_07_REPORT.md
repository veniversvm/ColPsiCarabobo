# FIX-07 Report — SMTP habilitado con MailHog para dev

| Campo | Valor |
|-------|-------|
| **Fix ID** | FIX-07 |
| **Hallazgo original** | CRIT-07 |
| **Archivos modificados** | `internal/service/mail_service.go`, `docker-compose.yml` |
| **Fecha de implementación** | 2026-07-25 |
| **Estado** | Completado |

---

## Problema

El envío SMTP estaba comentado en `executeSend()`. Los usuarios nunca recibían credenciales, notificaciones de login, etc. El servicio "funcionaba" pero nunca enviaba nada.

```go
// ANTES (comentado):
// if err := s.client.DialAndSend(m); err != nil {
//     return fmt.Errorf("fallo la conexión SMTP o el envío: %w", err)
// }
return nil
```

---

## Corrección

### 1. Descomentar DialAndSend

```go
// DESPUÉS:
if err := s.client.DialAndSend(m); err != nil {
    return fmt.Errorf("fallo el envío de email: %w", err)
}
return nil
```

### 2. TLS condicional (MailHog no soporta TLS obligatorio)

```go
tlsPolicy := mail.TLSMandatory
if config.Envs.Environment == "development" {
    tlsPolicy = mail.TLSOpportunistic
}
```

### 3. MailHog en docker-compose.yml (profile dev)

```yaml
mailhog:
  image: mailhog/mailhog:latest
  container_name: colpsi_mailhog
  ports:
    - "1025:1025"   # SMTP server
    - "8025:8025"   # Web UI
  profiles:
    - dev
```

### 4. .env para desarrollo

```
SMTP_HOST=mailhog
SMTP_PORT=1025
```

---

## Comportamiento por entorno

| Entorno | SMTP_HOST | MailHog | TLS | Resultado |
|---------|-----------|---------|-----|-----------|
| Producción | `smtp.tu-servidor.com` | No | Mandatory | Envía correo real con TLS |
| Desarrollo con MailHog | `mailhog` | Sí | Opportunistic | Envía a MailHog → `localhost:8025` |
| Desarrollo sin MailHog | `` (vacío) | No | Opportunistic | `NewMailService` falla → no envía |

---

## Uso en desarrollo

```bash
docker compose --profile dev up -d
# Correos se ven en http://localhost:8025
```

---

## Testing

- `go build ./...` — Pass
- `go vet ./...` — Pass
- Middleware tests: 15/15 Pass
- Utils tests: 4/4 Pass

---

## Archivos relacionados

| Archivo | Relación |
|---------|----------|
| `internal/service/mail_service.go:52-58` | TLS condicional |
| `internal/service/mail_service.go:197-200` | DialAndSend descomentado |
| `docker-compose.yml` | Servicio MailHog |
| `internal/router/psi_router.go:20-22` | FIX-10: no panica si SMTP falla |
