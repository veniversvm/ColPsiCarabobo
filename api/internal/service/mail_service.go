// api/internal/service/mail_service.go
package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/templates"
	"github.com/wneessen/go-mail"
)

// siteURL es la URL pública del frontend inyectada en las plantillas de correo.
// Temporalmente fija al dominio de despliegue actual para que todos los correos
// (incluso desde dev, que también usa Resend) apunten a producción.
const siteURL = "https://franhsabt-testing-ground.lat"

// maskEmail enmascara un email para logs: "j***@e****.com"
func maskEmail(email string) string {
	at := -1
	for i, c := range email {
		if c == '@' {
			at = i
			break
		}
	}
	if at < 0 || at > len(email)-2 {
		return "***"
	}
	domain := email[at+1:]
	dot := -1
	for i, c := range domain {
		if c == '.' {
			dot = i
			break
		}
	}
	if dot < 0 {
		return string(email[0]) + "***@" + domain
	}
	return string(email[0]) + "***@" + string(domain[0]) + "***" + domain[dot:]
}

// IMailService define el contrato público para el envío de correos electrónicos.
//
// Principio de Inversión de Dependencias (SOLID):
// Al exponer una interfaz en lugar del struct concreto, permitimos que otras capas
// (como los Controladores o los Tests Unitarios) dependan de este contrato.
// Esto hace que el servicio sea "Mockeable" (testeable) sin necesidad de enviar
// correos reales durante la ejecución de las pruebas.
type IMailService interface {
	SendEmail(to string, subject string, templateName string, data interface{}) error
}

// MailJob representa la unidad de trabajo (Payload) que viajará a través del canal.
// Encapsula toda la información necesaria para que el Worker de fondo pueda
// construir y despachar el correo de forma independiente.
type MailJob struct {
	To           string
	Subject      string
	TemplateName string
	Data         interface{}
}

// MailService es el motor de despacho de correos electrónicos.
//
// Patrón Arquitectónico: Productor-Consumidor (Producer-Consumer).
// Mantiene una conexión SMTP configurada y un canal (queue) que actúa como un
// amortiguador (Buffer) entre las peticiones HTTP rápidas y el lento protocolo SMTP.
type MailService struct {
	client    *mail.Client
	from      string
	queue     chan MailJob // Nuestra cola de mensajes en memoria (Thread-safe por naturaleza en Go)
	cancel    context.CancelFunc
	closeOnce sync.Once
	closed    bool
}

// NewMailService inicializa el servicio SMTP y levanta el demonio de procesamiento.
func NewMailService() (*MailService, error) {
	tlsPolicy := mail.TLSMandatory
	if config.Envs.Environment == "development" {
		tlsPolicy = mail.TLSOpportunistic
	}

	opts := []mail.Option{
		mail.WithPort(config.Envs.SMTPPort),
		mail.WithTLSPolicy(tlsPolicy),
	}

	// Auth SMTP solo si hay credenciales (SMTP_USER). Servidores locales de
	// desarrollo (MailHog/Mailpit) no requieren autenticación; forzar AUTH sin
	// credenciales rompe el envío ("SMTP AUTH failed: unencrypted connection").
	if config.Envs.SMTPUser != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(config.Envs.SMTPUser),
			mail.WithPassword(config.Envs.SMTPPass),
		)
	}

	c, err := mail.NewClient(config.Envs.SMTPHost, opts...)
	if err != nil {
		return nil, fmt.Errorf("fallo al crear cliente de correo: %w", err)
	}

	ms := &MailService{
		client: c,
		from:   config.Envs.SMTPFrom,
		// Manejo de Contrapresión (Backpressure):
		// Un buffer de 5000 permite al servidor absorber ráfagas masivas de eventos
		// (ej. un administrador enviando un boletín a todos los psicólogos) sin
		// bloquear la interfaz web. Si la cola se llena, las peticiones HTTP empezarán
		// a bloquearse, protegiendo al servidor de quedarse sin memoria RAM (OOM).
		queue: make(chan MailJob, 5000),
	}

	// Iniciamos el worker en segundo plano (Daemon) al arrancar el servicio.
	// Este hilo vivirá durante toda la ejecución de la aplicación.
	ctx, cancel := context.WithCancel(context.Background())
	ms.cancel = cancel
	go ms.startWorker(ctx)

	return ms, nil
}

// startWorker es el consumidor perpetuo de la cola de correos.
//
// Tácticas Avanzadas de Evasión Anti-Spam (Throttling y Jittering):
// Los proveedores de correo modernos (Gmail, Outlook) penalizan o bloquean IPs que
// envían cientos de correos por segundo (comportamiento de botnet). Este worker implementa:
//  1. Throttling (Estrangulamiento): Límites de ráfagas (Batch de 30 correos).
//  2. Jittering (Aleatoriedad): Introduce pausas matemáticas impredecibles (60 a 180 seg)
//     para simular el patrón de envío de un operador humano y evitar algoritmos heurísticos.
//  3. Rate Limiting por Socket: Una pausa microscópica (500ms) entre correos para evitar
//     el agotamiento de los file descriptors (Socket Exhaustion) del servidor de origen.
func (s *MailService) startWorker(ctx context.Context) {
	log.Info().Str("component", "mail").Msg("Mail Worker iniciado y escuchando cola...")

	sentInBatch := 0

	for {
		select {
		case <-ctx.Done():
			log.Info().Str("component", "mail").Msg("Mail Worker detenido por señal de cierre")
			return
		case job, ok := <-s.queue:
			if !ok {
				log.Info().Str("component", "mail").Msg("Mail Worker detenido: canal cerrado")
				return
			}
			// 1. Intentar enviar el correo
			if err := s.executeSend(job); err != nil {
				log.Error().Err(err).Str("component", "mail").Str("to", maskEmail(job.To)).Msg("ERROR critico en Worker al enviar")
			} else {
				log.Info().Str("component", "mail").Str("to", maskEmail(job.To)).Msg("Correo procesado por Worker")
				sentInBatch++
			}

			// 2. LOGICA DE PAUSA (Evasión de Filtros de Spam y Límites de Cuota)
			// Cada 30 correos, el worker se toma un descanso aleatorio.
			if sentInBatch >= 30 {
				// Generar tiempo aleatorio entre 60 y 180 segundos (1 a 3 min) -> Jitter Pattern
				waitTime := rand.Intn(120) + 60
				log.Info().Str("component", "mail").Int("seconds", waitTime).Msg("Limite de rafaga (30) alcanzado. Worker descansara para evitar spam...")

				select {
				case <-ctx.Done():
					log.Info().Str("component", "mail").Msg("Mail Worker detenido durante pausa de spam")
					return
				case <-time.After(time.Duration(waitTime) * time.Second):
				}

				sentInBatch = 0 // Reiniciar contador de ráfaga para el siguiente ciclo
				log.Info().Str("component", "mail").Msg("Worker reanudado.")
			}

			// Pequeña pausa de cortesía entre correos individuales para no saturar el socket
			// de red ni activar las alarmas de "Conexiones Concurrentes" del proveedor SMTP.
			select {
			case <-ctx.Done():
				log.Info().Str("component", "mail").Msg("Mail Worker detenido")
				return
			case <-time.After(500 * time.Millisecond):
			}
		}
	}
}

// Close detiene el worker y cierra el canal de correos.
// Debe invocarse durante el graceful shutdown para evitar goroutine leaks.
// Seguro para múltiples llamadas gracias a sync.Once.
func (s *MailService) Close() {
	s.closeOnce.Do(func() {
		s.closed = true
		if s.cancel != nil {
			s.cancel()
		}
		if s.queue != nil {
			close(s.queue)
		}
	})
}

// SendEmail es el método público utilizado por los Controladores y otros Servicios.
//
// Patrón "Fire-and-Forget":
// Ahora NO envía el correo físicamente, solo lo encola en la memoria. Al ser una
// operación estrictamente no-bloqueante (inserción en canal), garantiza que los
// endpoints HTTP (como el Registro o Login) respondan en ~10ms en lugar de esperar
// los ~2000ms que normalmente tarda un Handshake SMTP completo.
func (s *MailService) SendEmail(to string, subject string, templateName string, data interface{}) error {
	if s == nil || s.queue == nil || s.closed {
		return errors.New("el servicio de correo no está listo")
	}

	// Ponemos el trabajo en la cola.
	// Nota: Si la cola (buffer 5000) estuviera llena, esta línea se bloquearía
	// temporalmente (Backpressure), lo cual es el comportamiento correcto en Go.
	s.queue <- MailJob{
		To:           to,
		Subject:      subject,
		TemplateName: templateName,
		Data:         data,
	}

	log.Info().Str("component", "mail").Str("to", maskEmail(to)).Str("subject", subject).Msg("Correo encolado")
	return nil
}

// augmentTemplateData inyecta variables globales de configuración en los datos
// de la plantilla (URL de la aplicación y firma institucional). Así, todos los
// correos se personalizan desde .env sin tocar cada punto de envío.
//
// Las variables expuestas a las plantillas son:
//   - SiteURL:   URL pública del frontend (siteURL, fija al dominio de despliegue)
//   - Signature: firma mostrada tras "Atentamente" (config.Envs.MailSignature)
func augmentTemplateData(data interface{}) map[string]interface{} {
	base := map[string]interface{}{
		"SiteURL":   siteURL,
		"Signature": config.Envs.MailSignature,
	}
	if data == nil {
		return base
	}
	if m, ok := data.(map[string]interface{}); ok {
		for k, v := range m {
			base[k] = v
		}
		return base
	}
	// Fallback para datos no-map (no usado por las plantillas actuales):
	// se exponen igualmente las variables globales junto al dato original.
	base["Data"] = data
	return base
}

// executeSend realiza la orquestación, renderización de plantillas y el envío
// físico de forma síncrona. Es invocado exclusivamente por el Worker de fondo.
func (s *MailService) executeSend(job MailJob) error {
	// 1. Preparar la plantilla HTML
	// Se aprovecha `embed.FS` de Go para leer la plantilla directamente de la memoria RAM
	// compilada, evitando lecturas lentas al disco duro (I/O).
	fileName := job.TemplateName + ".html"
	tmpl, err := template.ParseFS(templates.FS, fileName)
	if err != nil {
		return fmt.Errorf("error al leer plantilla embebida: %w", err)
	}

	// 2. Renderizar contenido dinámico
	// Se inyecta la estructura `job.Data` en la plantilla de forma segura (previniendo XSS).
	var body bytes.Buffer
	if err := tmpl.Execute(&body, augmentTemplateData(job.Data)); err != nil {
		return fmt.Errorf("error al inyectar datos en la plantilla: %w", err)
	}

	// 3. Configurar el mensaje
	m := mail.NewMsg()

	// Seteamos el remitente con nombre amigable para la Experiencia de Usuario (UX)
	// (Ej: Colegio de Psicólogos de Carabobo <no-reply@...>)
	// Nota Arquitectónica: NO uses m.From() después de esto o perderás el formato amigable.
	if err := m.FromFormat(config.Envs.SMTPFromName, s.from); err != nil {
		return fmt.Errorf("error en formato del remitente: %w", err)
	}

	// Aplicar Reply-To si está configurado.
	// Mejora operativa: Evita que los usuarios respondan al correo 'no-reply' y
	// redirige sus respuestas al correo real de soporte del colegio.
	if config.Envs.SMTPReplyTo != "" {
		m.ReplyTo(config.Envs.SMTPReplyTo)
	}

	// Destinatario y Asunto
	if err := m.To(job.To); err != nil {
		return fmt.Errorf("error en formato del destinatario: %w", err)
	}
	m.Subject(job.Subject)
	m.SetBodyString(mail.TypeTextHTML, body.String())

	// 4. Envío físico al servidor SMTP
	// DialAndSend se encarga automáticamente de abrir el Socket TCP, realizar el
	// Handshake TLS, autenticar, enviar el payload y cerrar la conexión educadamente (QUIT).
	if err := s.client.DialAndSend(m); err != nil {
		return fmt.Errorf("fallo el envío de email: %w", err)
	}

	return nil
}
