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

// MailJob representa un correo pendiente en la cola
type MailJob struct {
	To           string
	Subject      string
	TemplateName string
	Data         interface{}
}

type MailService struct {
	client *mail.Client
	from   string
	queue  chan MailJob // Nuestra cola de mensajes
}

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
		queue:  make(chan MailJob, 1000), // Buffer para 1000 correos
	}

	// Iniciamos el worker en segundo plano al arrancar el servicio
	go ms.startWorker()

	return ms, nil
}

// startWorker es el consumidor de la cola
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

		// 2. LOGICA DE PAUSA (Tu requerimiento)
		// Cada 30 correos, el worker se toma un descanso aleatorio
		if sentInBatch >= 30 {
			// Generar tiempo aleatorio entre 60 y 180 segundos (1 a 3 min)
			waitTime := rand.Intn(120) + 60
			log.Printf("🕒 Límite de ráfaga (30) alcanzado. El Worker descansará %d segundos para evitar spam...", waitTime)

			time.Sleep(time.Duration(waitTime) * time.Second)

			sentInBatch = 0 // Reiniciar contador de ráfaga
			log.Println("🔄 Worker reanudado.")
		}

		// Pequeña pausa de cortesía entre correos individuales para no saturar el socket
		time.Sleep(500 * time.Millisecond)
	}
}

// SendEmail ahora NO envía el correo, solo lo pone en la cola (no-bloqueante)
func (s *MailService) SendEmail(to string, subject string, templateName string, data interface{}) error {
	if s == nil || s.queue == nil {
		return errors.New("el servicio de correo no está listo")
	}

	// Ponemos el trabajo en la cola
	s.queue <- MailJob{
		To:           to,
		Subject:      subject,
		TemplateName: templateName,
		Data:         data,
	}

	log.Printf("📥 Correo encolado para %s", to)
	return nil
}

// executeSend realiza la renderización y el envío físico de forma síncrona para el worker.
func (s *MailService) executeSend(job MailJob) error {
	// 1. Preparar la plantilla
	fileName := job.TemplateName + ".html"
	tmpl, err := template.ParseFS(templates.FS, fileName)
	if err != nil {
		return fmt.Errorf("error al leer plantilla embebida: %w", err)
	}

	// 2. Renderizar contenido dinámico
	var body bytes.Buffer
	if err := tmpl.Execute(&body, job.Data); err != nil {
		return fmt.Errorf("error al inyectar datos en la plantilla: %w", err)
	}

	// 3. Configurar el mensaje
	m := mail.NewMsg()

	// Seteamos el remitente con nombre amigable (Ej: Colegio de Psicólogos de Carabobo <no-reply@...>)
	// NO uses m.From() después de esto o perderás el nombre.
	if err := m.FromFormat(config.Envs.SMTPFromName, s.from); err != nil {
		return fmt.Errorf("error en formato del remitente: %w", err)
	}

	// Aplicar Reply-To si está configurado para que las respuestas no lleguen al no-reply
	if config.Envs.SMTPReplyTo != "" {
		m.ReplyTo(config.Envs.SMTPReplyTo)
	}

	// Destinatario y Asunto
	if err := m.To(job.To); err != nil {
		return fmt.Errorf("error en formato del destinatario: %w", err)
	}
	m.Subject(job.Subject)
	m.SetBodyString(mail.TypeTextHTML, body.String())

	// 4. Envío físico
	// DialAndSend se encarga de abrir la conexión, enviar y cerrar.
	if err := s.client.DialAndSend(m); err != nil {
		return fmt.Errorf("fallo la conexión SMTP o el envío: %w", err)
	}

	return nil
}
