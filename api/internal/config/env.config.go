// api/internal/config/env.config.go

// Package config centraliza la gestión de variables de entorno y parámetros de
// configuración global de la aplicación.
package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// Config define la estructura de las variables de entorno necesarias para que
// la aplicación funcione correctamente en distintos entornos (desarrollo, producción).
type Config struct {
	// Server Configuration
	Port string // Puerto donde correrá el servidor Fiber

	// Database Configuration (PostgreSQL)
	DBHost string // Host de la base de datos
	DBPort string // Puerto de la base de datos
	DBUser string // Usuario de acceso
	DBPass string // Contraseña de acceso
	DBName string // Nombre de la base de datos

	// S3 Configuration (AWS S3 o MinIO)
	S3Bucket    string // Nombre del bucket para almacenamiento de archivos
	S3Region    string // Región geográfica del servicio de almacenamiento
	S3Endpoint  string // URL del endpoint interno (MinIO en desarrollo) para el SDK
	S3PublicURL string // URL pública accesible desde el navegador (endpoint/bucket)
	S3AccessKey string // Credencial de acceso (Access Key ID)
	S3SecretKey string // Credencial secreta (Secret Access Key)

	Environment string // Entorno de ejecución (development, production, etc.)

	// Email
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPass     string
	SMTPFrom     string
	SMTPReplyTo  string
	SMTPFromName string

	// URL pública de la aplicación (frontend) y firma institucional que se
	// inyectan en las plantillas de correo. Configurables desde .env.
	AppURL        string
	MailSignature string

	// Transporte de email: "smtp" (Mailpit/MailHog) o "resend" (API Resend)
	MailTransport string

	// Resend (usado solo si MailTransport == "resend")
	ResendAPIKey string

	// Pacing del worker Resend (anti-spam y respeto a cuotas/rate limits)
	MailDailyCap        int // Tope diario de envíos (warm-up de dominio)
	MailBatchSize       int // Máx emails por llamada batch (límite Resend: 100)
	MailBatchIntervalMS int // Pausa entre llamadas batch
	MailBurstSize       int // Ráfaga de envíos antes de la pausa larga
	MailBurstSleepMin   int // Segundos mínimos de la pausa larga
	MailBurstSleepMax   int // Segundos máximos de la pausa larga
	MailSendIntervalMS  int // Ventana de acumulación de la cola antes de despachar un lote

	// Origins
	AllowedOrigins string

	// Valkey (rate limiting store)
	ValkeyAddr string // Valkey/Redis address (e.g. "colpsi_valkey:6379"). Empty = in-memory fallback.

	//
	JwtLibrarySecret string
	AbsAdminToken    string
}

// Envs es una instancia global (Singleton) que contiene la configuración cargada.
// Se utiliza para acceder a los parámetros desde cualquier punto del sistema.
var Envs *Config

// InitConfig carga las variables de entorno desde el archivo .env ubicado en la raíz.
// Si el archivo no existe, intentará leer las variables directamente del sistema operativo.
// Debe ser invocada al inicio de la función main() antes de inicializar cualquier otro servicio.
func InitConfig() {
	// Intentamos cargar el archivo .env
	err := godotenv.Load()
	if err != nil {
		// No lanzamos Fatal aquí porque en entornos como Docker/Heroku/AWS
		// las variables suelen estar ya inyectadas en el sistema.
		log.Warn().Str("component", "config").Msg("No .env file found, using system environment variables")
	}

	// Poblamos el struct Envs con valores del entorno o sus fallbacks (valores por defecto)
	email_port, err := strconv.Atoi(getEnv("SMTP_PORT", "1025"))
	if err != nil {
		log.Error().Err(err).Str("component", "config").Msg("Puerto de email invalido")
	}
	Envs = &Config{
		// Configuración del Servidor
		Port: getEnv("PORT", "8080"),

		// Configuración de Base de Datos
		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: getEnv("DB_PORT", "5432"),
		DBUser: getEnv("DB_USER", "postgres"),
		DBPass: getEnv("DB_PASSWORD", "postgres"),
		DBName: getEnv("DB_NAME", "colpsi_db"),

		// Configuración de S3
		S3Bucket:    getEnv("AWS_S3_BUCKET", "colpsi-bucket"),
		S3Region:    getEnv("AWS_REGION", "us-east-1"),
		S3Endpoint:  getEnv("S3_ENDPOINT", "http://localhost:9000"),
		S3PublicURL: getEnv("S3_PUBLIC_URL", "http://localhost:9000"),
		S3AccessKey: getEnv("AWS_ACCESS_KEY_ID", "minioadmin"),
		S3SecretKey: getEnv("AWS_SECRET_ACCESS_KEY", "minioadmin"),

		// Configuración de Entorno
		Environment: getEnv("APP_ENV", "production"),

		// Configuración de Email
		SMTPHost:     getEnv("SMTP_HOST", "localhost"),
		SMTPPort:     email_port, // Helper parseInt
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPass:     getEnv("SMTP_PASS", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "info@colpsicarabobo.com"),
		SMTPReplyTo:  getEnv("SMTP_REPLY_TO", ""),
		SMTPFromName: getEnv("SMTP_FROM_NAME", "Colegio de Psicólogos de Carabobo"),

		// URL de la app y firma de correo (inyectadas en las plantillas)
		AppURL:        getEnv("APP_URL", "http://localhost:3000"),
		MailSignature: getEnv("MAIL_SIGNATURE", "Administración ColPsiCarabobo"),

		// Transporte de email
		MailTransport: getEnv("MAIL_TRANSPORT", "smtp"),

		// Resend
		ResendAPIKey: getEnv("RESEND_API_KEY", ""),

		// Pacing del worker Resend
		MailDailyCap:        getEnvInt("MAIL_DAILY_CAP", 100),
		MailBatchSize:       getEnvInt("MAIL_BATCH_SIZE", 100),
		MailBatchIntervalMS: getEnvInt("MAIL_BATCH_INTERVAL_MS", 1000),
		MailBurstSize:       getEnvInt("MAIL_BURST_SIZE", 30),
		MailBurstSleepMin:   getEnvInt("MAIL_BURST_SLEEP_MIN", 60),
		MailBurstSleepMax:   getEnvInt("MAIL_BURST_SLEEP_MAX", 180),
		MailSendIntervalMS:  getEnvInt("MAIL_SEND_INTERVAL_MS", 500),

		// Origins
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://127.0.0.1:3000, http://localhost:3000"),

		// Valkey
		ValkeyAddr: getEnv("VALKEY_ADDR", ""),

		// Library
		JwtLibrarySecret: getEnv("JWT_LIBRARY_SECRET", ""),
		AbsAdminToken:    getEnv("ABS_ADMIN_TOKEN", ""),
	}
}

// getEnv es una función auxiliar que intenta obtener una variable de entorno por su clave.
// Retorna el valor encontrado o el valor 'fallback' si la variable no está definida.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// getEnvInt es una función auxiliar que intenta obtener una variable de entorno
// numérica por su clave. Si no está definida o no es válida, retorna el fallback.
func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
		log.Warn().Str("component", "config").Str("key", key).Msg("Valor numérico inválido, usando fallback")
	}
	return fallback
}
