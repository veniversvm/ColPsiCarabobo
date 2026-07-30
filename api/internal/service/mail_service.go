// api/internal/service/mail_service.go

// Package service implementa la lógica de negocio central de la aplicación.
//
// Este archivo gestiona la infraestructura de mensajería saliente (Outbound Messaging).
// Está diseñado para absorber cargas masivas (ej. 3,000 correos simultáneos)
// delegando el envío a la API de Resend sin bloquear los hilos HTTP del servidor.
package service

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/resend/resend-go/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/templates"
)

// IMailService define el contrato público para el envío de correos electrónicos.
//
// Principio de Inversión de Dependencias (SOLID):
// Al exponer una interfaz en lugar del struct concreto, permitimos que otras capas
// (como los Controladores o los Tests Unitarios) dependan de este contrato.
// Esto hace que el servicio sea "Mockeable" (testeable) sin necesidad de enviar
// correos reales ni gastar cuota de la API durante la ejecución de las pruebas.
type IMailService interface {
	SendEmail(to string, subject string, templateName string, data interface{}) error
}

// MailJob representa la unidad de trabajo (Payload) que viajará a través del canal.
// Encapsula toda la información necesaria para que el Worker de fondo pueda
// compilar la plantilla HTML y despachar el correo de forma independiente.
type MailJob struct {
	To           string
	Subject      string
	TemplateName string
	Data         interface{}
}

// MailService es el motor de despacho de correos electrónicos.
//
// Patrón Arquitectónico: Productor-Consumidor (Producer-Consumer).
// Mantiene un cliente HTTP persistente hacia Resend y un canal (queue) que actúa como un
// amortiguador (Buffer) entre las peticiones rápidas de Go y los límites de la API externa.
type MailService struct {
	client *resend.Client
	from   string
	queue  chan MailJob // Cola de mensajes en memoria (Thread-safe nativo en Go)
}

// NewMailService inicializa la integración con Resend y levanta el demonio de procesamiento.
func NewMailService() (*MailService, error) {
	// Failsafe: Aborta el arranque del servidor si faltan credenciales críticas.
	if config.Envs.ResendAPIKey == "" {
		return nil, errors.New("la clave de API de Resend no está configurada")
	}

	client := resend.NewClient(config.Envs.ResendAPIKey)

	// Formateamos el remitente según el estándar RFC 5322.
	// (Ej: "Colegio de Psicólogos <no-reply@dominio.com>")
	fromFormatted := fmt.Sprintf("%s <%s>", config.Envs.SMTPFromName, config.Envs.SMTPFrom)

	ms := &MailService{
		client: client,
		from:   fromFormatted,
		// Buffer Masivo (Manejo de Contrapresión):
		// Capaz de alojar los 3,000 correos de la importación inicial en memoria RAM,
		// evitando que el Importador CSV se bloquee esperando a la red.
		queue: make(chan MailJob, 5000),
	}

	// Levantamos el demonio (Daemon) de envíos en background
	go ms.startWorker()

	return ms, nil
}

// startWorker procesa la cola protegiendo la reputación del dominio (Deliverability).
//
// Estrategia para la Carga de 3,000 Correos (Warm-up Strategy & Anti-Spam):
// Enviar 3k correos de golpe desde un dominio nuevo casi siempre dispara los filtros
// anti-spam de Gmail y Hotmail. Este worker procesa la cola inyectando Jittering
// (pausas aleatorias) y micro-retrasos. Esto simula tráfico orgánico humano y
// "calienta" tu dominio e IP compartida en Resend de forma segura.
func (s *MailService) startWorker() {
	log.Println("🚀 Resend Mail Worker iniciado y escuchando cola...")

	sentInBatch := 0

	for job := range s.queue {
		if err := s.executeSend(job); err != nil {
			log.Printf("❌ ERROR critico en Resend Worker al enviar a %s: %v", job.To, err)
		} else {
			log.Printf("📧 Correo procesado y enviado via Resend a: %s", job.To)
			sentInBatch++
		}

		// Lógica de "Domain Warm-up" y Evasión de Filtros Heurísticos
		if sentInBatch >= 30 {
			// Jittering: Descanso aleatorio entre 60 y 180 segundos
			waitTime := rand.Intn(120) + 60
			log.Printf("🕒 Límite de ráfaga (30) alcanzado. El Worker descansará %d segundos para evitar filtros de Spam...", waitTime)

			time.Sleep(time.Duration(waitTime) * time.Second)

			sentInBatch = 0
			log.Println("🔄 Worker reanudado.")
		}

		// Rate Limiting Activo: Micro-pausa para respetar los límites de la API
		// de Resend (que suele limitar peticiones concurrentes agresivas por segundo).
		time.Sleep(500 * time.Millisecond)
	}
}

// SendEmail gestiona el ingreso de mensajes a la cola de despacho asíncrono.
//
// Lógica de Omisión (Silent Skip):
// Implementa un filtro de "Higiene de Datos" que detecta direcciones placeholder.
// Si el destinatario contiene la cadena "sincorreo", el sistema aborta el envío
// devolviendo un error 'nil'. Esto es crítico para:
//  1. Evitar "Hard Bounces" en Resend que degradarían la reputación del dominio.
//  2. Optimizar el flujo de importación masiva al no saturar la cola con datos técnicos.
func (s *MailService) SendEmail(to string, subject string, templateName string, data interface{}) error {
	// 1. Verificación de Disponibilidad (Readiness Check)
	if s == nil || s.queue == nil {
		return errors.New("el servicio de correo no está listo")
	}

	// 2. Filtro de Integridad de Destinatario
	// Detecta correos generados sintéticamente para usuarios sin email real.
	if strings.Contains(to, "sincorreo") {
		log.Printf("📥 INFO: %s detectado como placeholder. Envío omitido por integridad.", to)
		return nil // Se retorna nil para no interrumpir procesos masivos (Batch Processes)
	}

	// 3. Encolado Asíncrono (Non-blocking Push)
	// Se coloca el trabajo en el canal para que el Worker lo procese según su ritmo de throttling.
	s.queue <- MailJob{
		To:           to,
		Subject:      subject,
		TemplateName: templateName,
		Data:         data,
	}

	log.Printf("📥 Correo encolado exitosamente para: %s", to)
	return nil
}

// executeSend realiza la orquestación, renderización de plantillas y el envío
// físico a la API REST de Resend. Es invocado EXCLUSIVAMENTE por la Goroutine del Worker.
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
	// Inyección segura (XSS Prevention inherente en html/template) de las variables.
	var body bytes.Buffer
	if err := tmpl.Execute(&body, job.Data); err != nil {
		return fmt.Errorf("error al inyectar datos en la plantilla: %w", err)
	}

	// 3. Ensamblar la petición (Request Payload) para Resend
	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{job.To},
		Subject: job.Subject,
		Html:    body.String(),
	}

	// Aplicar Reply-To si existe en el entorno.
	// Mejora de UX: Redirige las respuestas de los usuarios al correo de soporte real.
	if config.Envs.SMTPReplyTo != "" {
		params.ReplyTo = config.Envs.SMTPReplyTo
	}

	// 4. Disparar API de Resend (Ejecución de Red)
	_, err = s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("fallo la API de Resend: %w", err)
	}

	return nil
}
