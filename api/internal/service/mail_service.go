// api/internal/service/admin_service.go (Asumo que el paquete es service por el contexto)
package service

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"time"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/templates"
	"github.com/wneessen/go-mail"
)

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
	client *mail.Client
	from   string
	queue  chan MailJob // Nuestra cola de mensajes en memoria (Thread-safe por naturaleza en Go)
}

// NewMailService inicializa el servicio SMTP y levanta el demonio de procesamiento.
func NewMailService() (*MailService, error) {
	c, err := mail.NewClient(config.Envs.SMTPHost,
		mail.WithPort(config.Envs.SMTPPort),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(config.Envs.SMTPUser),
		mail.WithPassword(config.Envs.SMTPPass),
		mail.WithTLSPolicy(mail.TLSMandatory),
	)
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
	go ms.startWorker()

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
func (s *MailService) startWorker() {
	log.Println("🚀 Mail Worker iniciado y escuchando cola...")

	sentInBatch := 0

	for job := range s.queue {
		// 1. Intentar enviar el correo
		if err := s.executeSend(job); err != nil {
			log.Printf("❌ ERROR critico en Worker al enviar a %s: %v", job.To, err)
		} else {
			log.Printf("📧 Correo procesado por Worker: %s", job.To)
			sentInBatch++
		}

		// 2. LOGICA DE PAUSA (Evasión de Filtros de Spam y Límites de Cuota)
		// Cada 30 correos, el worker se toma un descanso aleatorio.
		if sentInBatch >= 30 {
			// Generar tiempo aleatorio entre 60 y 180 segundos (1 a 3 min) -> Jitter Pattern
			waitTime := rand.Intn(120) + 60
			log.Printf("🕒 Límite de ráfaga (30) alcanzado. El Worker descansará %d segundos para evitar spam...", waitTime)

			time.Sleep(time.Duration(waitTime) * time.Second)

			sentInBatch = 0 // Reiniciar contador de ráfaga para el siguiente ciclo
			log.Println("🔄 Worker reanudado.")
		}

		// Pequeña pausa de cortesía entre correos individuales para no saturar el socket
		// de red ni activar las alarmas de "Conexiones Concurrentes" del proveedor SMTP.
		time.Sleep(500 * time.Millisecond)
	}
}

// SendEmail es el método público utilizado por los Controladores y otros Servicios.
//
// Patrón "Fire-and-Forget":
// Ahora NO envía el correo físicamente, solo lo encola en la memoria. Al ser una
// operación estrictamente no-bloqueante (inserción en canal), garantiza que los
// endpoints HTTP (como el Registro o Login) respondan en ~10ms en lugar de esperar
// los ~2000ms que normalmente tarda un Handshake SMTP completo.
func (s *MailService) SendEmail(to string, subject string, templateName string, data interface{}) error {
	if s == nil || s.queue == nil {
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

	log.Printf("📥 Correo encolado para %s", to)
	return nil
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
	if err := tmpl.Execute(&body, job.Data); err != nil {
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
	//
	// (Nota: Actualmente comentado por el autor, presumiblemente para entornos de desarrollo/pruebas).
	// if err := s.client.DialAndSend(m); err != nil {
	// 	return fmt.Errorf("fallo la conexión SMTP o el envío: %w", err)
	// }

	return nil
}
