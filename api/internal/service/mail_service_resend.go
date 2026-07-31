// api/internal/service/mail_service_resend.go

// Transporte alternativo de email basado en la API REST de Resend.
//
// Se selecciona con MAIL_TRANSPORT=resend. Comparte la interfaz IMailService
// con el transporte SMTP (mail_service.go) y el mismo tipo MailJob, por lo que
// los servicios y handlers no necesitan saber cuál está activo.
//
// Diseño del worker (protección de reputación y respeto a límites de Resend):
//   - Cola en memoria (buffer 5000) + un único worker secuencial: nunca hay
//     más de 1 request HTTP en vuelo hacia Resend.
//   - Envío por LOTES (Batch API): cada llamada transporta hasta MAIL_BATCH_SIZE
//     correos (máx 100, límite duro de Resend). Una importación de 3,000 correos
//     se convierte en ~30 llamadas en vez de 3,000.
//   - Ráfaga anti-spam: tras MAIL_BURST_SIZE envíos exitosos, pausa jitter
//     (MAIL_BURST_SLEEP_MIN..MAX segundos) para simular tráfico orgánico.
//   - Tope diario (MAIL_DAILY_CAP): detiene el envío al alcanzar el límite del
//     día y reanuda al día siguiente. Controla la rampa de calentamiento del
//     dominio (warm-up) y evita cuotas del plan (Free = 100/día).
//   - Reintentos con backoff: respeta el header Retry-After en 429 (RateLimit)
//     y aplica backoff exponencial en fallos transitorios del servidor.
package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/resend/resend-go/v2"
	"github.com/rs/zerolog/log"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/templates"
)

const (
	// resendMaxBatchSize es el límite duro de la API de Resend por llamada batch.
	resendMaxBatchSize = 100
	// resendMaxAttempts es el número máximo de intentos por lote.
	resendMaxAttempts = 3
	// resendMaxBackoff limita la espera máxima entre reintentos.
	resendMaxBackoff = 2 * time.Minute
	// resendQueueBuffer es el tamaño del canal de cola en memoria.
	resendQueueBuffer = 5000
)

// ResendMailService es el motor de despacho de correos vía la API de Resend.
//
// Patrón Productor-Consumidor: SendEmail encola (no bloqueante) y el worker
// procesa la cola con throttling/jittering, batch y tope diario.
type ResendMailService struct {
	client *resend.Client
	from   string
	queue  chan MailJob

	cancel    context.CancelFunc
	closeOnce sync.Once
	closed    bool

	// Contadores internos del worker (solo se tocan desde su goroutine).
	sentToday   int    // envíos exitosos del día actual
	dayKey      string // clave del día ("2006-01-02") para reiniciar el tope
	sentInBatch int    // envíos acumulados desde la última pausa anti-spam
}

// NewResendMailService inicializa la integración con Resend y levanta el worker.
// Fallo duro si RESEND_API_KEY no está configurada (evita arrancar con un
// transporte inservible que descartaría correos en silencio).
func NewResendMailService() (*ResendMailService, error) {
	if strings.TrimSpace(config.Envs.ResendAPIKey) == "" {
		return nil, errors.New("RESEND_API_KEY no está configurada")
	}

	client := resend.NewClient(config.Envs.ResendAPIKey)

	// Formateamos el remitente según RFC 5322 (Ej: "Colegio de Psicólogos <no-reply@dominio.com>").
	fromFormatted := fmt.Sprintf("%s <%s>", config.Envs.SMTPFromName, config.Envs.SMTPFrom)

	ms := &ResendMailService{
		client: client,
		from:   fromFormatted,
		queue:  make(chan MailJob, resendQueueBuffer),
	}

	ctx, cancel := context.WithCancel(context.Background())
	ms.cancel = cancel
	go ms.startWorker(ctx)

	log.Info().Str("component", "mail_resend").Msg("Transporte Resend inicializado")
	return ms, nil
}

// Close detiene el worker y cierra el canal de correos.
// Debe invocarse durante el graceful shutdown para evitar goroutine leaks.
// Seguro para múltiples llamadas gracias a sync.Once.
func (s *ResendMailService) Close() {
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

// SendEmail encola un correo (Fire-and-Forget, no bloqueante salvo backpressure).
//
// Filtro de integridad: los destinatarios placeholder ("sincorreo") se omiten
// para evitar hard bounces que degradan la reputación del dominio en Resend.
func (s *ResendMailService) SendEmail(to string, subject string, templateName string, data interface{}) error {
	if s == nil || s.queue == nil || s.closed {
		return errors.New("el servicio de correo no está listo")
	}

	if strings.Contains(to, "sincorreo") {
		log.Info().Str("component", "mail_resend").Str("to", maskEmail(to)).
			Msg("Destinatario placeholder detectado. Envío omitido por integridad")
		return nil
	}

	s.queue <- MailJob{
		To:           to,
		Subject:      subject,
		TemplateName: templateName,
		Data:         data,
	}

	log.Info().Str("component", "mail_resend").Str("to", maskEmail(to)).Str("subject", subject).
		Msg("Correo encolado")
	return nil
}

// startWorker es el consumidor perpetuo de la cola. Agrupa los correos en lotes
// y los despacha con respeto a los límites de Resend y a la reputación del dominio.
func (s *ResendMailService) startWorker(ctx context.Context) {
	log.Info().Str("component", "mail_resend").Msg("Resend Mail Worker iniciado y escuchando cola...")

	for {
		select {
		case <-ctx.Done():
			log.Info().Str("component", "mail_resend").Msg("Resend Mail Worker detenido")
			return
		case job, ok := <-s.queue:
			if !ok {
				log.Info().Str("component", "mail_resend").Msg("Resend Mail Worker detenido: canal cerrado")
				return
			}

			// Tope diario (warm-up de dominio): si se agotó, pausar hasta mañana.
			if !s.dailyCapAvailable() {
				if err := s.sleepUntilDailyReset(ctx); err != nil {
					return
				}
			}

			// Acumular lote: hasta MAIL_BATCH_SIZE correos o la ventana de acumulación.
			batch := []MailJob{job}
			maxSize := config.Envs.MailBatchSize
			if maxSize <= 0 || maxSize > resendMaxBatchSize {
				maxSize = resendMaxBatchSize
			}
			window := time.Duration(config.Envs.MailSendIntervalMS) * time.Millisecond
			if window <= 0 {
				window = 500 * time.Millisecond
			}

			timer := time.NewTimer(window)
		drain:
			for len(batch) < maxSize {
				select {
				case j, ok := <-s.queue:
					if !ok {
						timer.Stop()
						break drain
					}
					batch = append(batch, j)
				case <-timer.C:
					break drain
				}
			}
			timer.Stop()

			if err := s.sendBatchWithRetry(ctx, batch); err != nil {
				log.Error().Err(err).Str("component", "mail_resend").
					Int("emails", len(batch)).Msg("Lote falló tras reintentos")
			}

			// Ráfaga anti-spam: tras MAIL_BURST_SIZE envíos, pausa jitter.
			if config.Envs.MailBurstSize > 0 && s.sentInBatch >= config.Envs.MailBurstSize {
				s.sentInBatch = 0
				if err := s.jitterPause(ctx); err != nil {
					return
				}
			}

			// Pausa de cortesía entre llamadas batch para respetar el rate limit.
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(config.Envs.MailBatchIntervalMS) * time.Millisecond):
			}
		}
	}
}

// sendBatchWithRetry despacha un lote completo con reintentos con backoff.
func (s *ResendMailService) sendBatchWithRetry(ctx context.Context, batch []MailJob) error {
	params := make([]*resend.SendEmailRequest, 0, len(batch))
	for _, job := range batch {
		body, err := s.renderTemplate(job)
		if err != nil {
			// Error permanente de plantilla: se descarta el ítem y se continúa.
			log.Error().Err(err).Str("component", "mail_resend").Str("to", maskEmail(job.To)).
				Msg("Error al renderizar plantilla; correo descartado")
			continue
		}

		req := &resend.SendEmailRequest{
			From:    s.from,
			To:      []string{job.To},
			Subject: job.Subject,
			Html:    body,
		}
		if config.Envs.SMTPReplyTo != "" {
			req.ReplyTo = config.Envs.SMTPReplyTo
		}
		params = append(params, req)
	}

	if len(params) == 0 {
		return nil
	}

	var lastErr error
	for attempt := 1; attempt <= resendMaxAttempts; attempt++ {
		resp, err := s.client.Batch.SendWithOptions(ctx, params, &resend.BatchSendEmailOptions{
			// Permissive: procesa el lote aunque algún destinatario sea inválido,
			// reportando los fallos individuales en resp.Errors.
			BatchValidation: resend.BatchValidationPermissive,
		})
		if err == nil {
			return s.handleBatchResult(resp)
		}
		lastErr = err

		if attempt == resendMaxAttempts {
			break
		}

		wait := s.backoffFor(err, attempt)
		log.Warn().Err(err).Str("component", "mail_resend").Int("attempt", attempt).
			Int("emails", len(params)).Dur("retry_in", wait).
			Msg("Fallo al enviar lote; reintentando")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}

	return fmt.Errorf("lote de %d correos falló tras %d intentos: %w", len(params), resendMaxAttempts, lastErr)
}

// handleBatchResult contabiliza éxitos y registra los rechazos individuales.
func (s *ResendMailService) handleBatchResult(resp *resend.BatchEmailResponse) error {
	successes := 0
	for _, item := range resp.Data {
		if item.Id != "" {
			successes++
		}
	}
	for _, e := range resp.Errors {
		log.Warn().Str("component", "mail_resend").Int("index", e.Index).
			Msgf("Correo del lote rechazado por Resend: %s", e.Message)
	}

	s.sentToday += successes
	s.sentInBatch += successes

	if successes > 0 {
		log.Info().Str("component", "mail_resend").Int("sent", successes).Msg("Lote enviado vía Resend")
	}
	if successes == 0 && len(resp.Errors) > 0 {
		return fmt.Errorf("todos los %d correos del lote fueron rechazados", len(resp.Errors))
	}
	return nil
}

// backoffFor calcula la espera antes de reintentar un lote.
// Para 429 (RateLimitError) respeta el header Retry-After de Resend; en el resto
// aplica backoff exponencial (2s, 4s, ...) acotado por resendMaxBackoff.
func (s *ResendMailService) backoffFor(err error, attempt int) time.Duration {
	var rle *resend.RateLimitError
	if errors.As(err, &rle) {
		if secs, convErr := strconv.Atoi(rle.RetryAfter); convErr == nil && secs > 0 {
			wait := time.Duration(secs) * time.Second
			if wait > resendMaxBackoff {
				return resendMaxBackoff
			}
			return wait
		}
	}

	wait := time.Duration(1<<attempt) * time.Second
	if wait > resendMaxBackoff {
		return resendMaxBackoff
	}
	return wait
}

// renderTemplate compila la plantilla embebida (embed.FS) y le inyecta los datos
// de forma segura (html/template previene XSS).
func (s *ResendMailService) renderTemplate(job MailJob) (string, error) {
	fileName := job.TemplateName + ".html"
	tmpl, err := template.ParseFS(templates.FS, fileName)
	if err != nil {
		return "", fmt.Errorf("error al leer plantilla embebida: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, job.Data); err != nil {
		return "", fmt.Errorf("error al inyectar datos en la plantilla: %w", err)
	}
	return body.String(), nil
}

// dailyCapAvailable verifica el tope diario y reinicia el contador si cambió el día.
func (s *ResendMailService) dailyCapAvailable() bool {
	if config.Envs.MailDailyCap <= 0 {
		return true
	}

	now := time.Now()
	dayKey := now.Format("2006-01-02")
	if s.dayKey != dayKey {
		s.dayKey = dayKey
		s.sentToday = 0
		log.Info().Str("component", "mail_resend").Msg("Nuevo día: contador diario reiniciado")
	}
	return s.sentToday < config.Envs.MailDailyCap
}

// sleepUntilDailyReset pausa el worker hasta la medianoche (warm-up de dominio).
func (s *ResendMailService) sleepUntilDailyReset(ctx context.Context) error {
	nextMidnight := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
	wait := time.Until(nextMidnight) + time.Second

	log.Info().Str("component", "mail_resend").Int("seconds", int(wait.Seconds())).
		Int("daily_cap", config.Envs.MailDailyCap).
		Msg("Tope diario alcanzado. El worker pausa hasta mañana (warm-up de dominio)")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(wait):
	}
	return nil
}

// jitterPause introduce una pausa aleatoria (anti-spam) tras cada ráfaga.
func (s *ResendMailService) jitterPause(ctx context.Context) error {
	min := config.Envs.MailBurstSleepMin
	max := config.Envs.MailBurstSleepMax
	if max <= min {
		max = min + 1
	}
	wait := time.Duration(rand.Intn(max-min)+min) * time.Second

	log.Info().Str("component", "mail_resend").Int("seconds", int(wait.Seconds())).
		Msg("Límite de ráfaga alcanzado. Pausa anti-spam")

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(wait):
	}
	return nil
}
