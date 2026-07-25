# 📧 Plantillas HTML (templates/)

Contiene plantillas HTML embebidas usando el paquete `embed` de Go.

## Archivos

| Archivo | Descripción |
|---------|-------------|
| `templates.go` | Embebe todas las plantillas HTML vía directiva `//go:embed` |

## Uso

Las plantillas se cargan como strings embebidos al inicio de la aplicación. El `MailService` las utiliza para envío asíncrono de correos.

## Plantillas disponibles

- **Password reset** — Correo de restablecimiento de contraseña
- **Welcome** — Correo de bienvenida
- **Notifications** — Correos de notificación
