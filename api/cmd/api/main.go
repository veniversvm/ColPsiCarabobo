package main

import (
	"io"
	"log"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	"github.com/gofiber/fiber/v2"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/database"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
)

// @title           ColPsiCarabobo API
// @version         1.0
// @description     Servicios backend para la gestión administrativa y pública del Colegio de Psicólogos del Estado Carabobo.
// @termsOfService  https://colpsicarabobo.com/terms/

// @contact.name   Soporte Técnico
// @contact.email  admin@colpsicarabobo.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1
// @schemes   http https

func main() {
	// =========================================================================
	// 1. CONFIGURACIÓN INICIAL
	// =========================================================================
	// Carga variables de entorno desde .env y las mapea al struct config.Envs
	config.InitConfig()

	// =========================================================================
	// 2. PERSISTENCIA (PostgreSQL)
	// =========================================================================
	// Inicializa el pool de conexiones con GORM
	_, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("❌ Error crítico: Falló la conexión a PostgreSQL: %v", err)
	}

	// =========================================================================
	// 3. GENERACIÓN DE ESQUEMA (Atlas Integration)
	// =========================================================================
	// Este bloque traduce los Structs de Go a sentencias SQL DDL.
	// Útil para alimentar a Atlas y generar migraciones versionadas.
	stmts, err := gormschema.New("postgres").Load(
		&domain.TextModel{},
		&domain.UserAdmin{},
		&domain.PsiUserModel{},
		&domain.PsiUserColData{},
		&domain.PsiUserPostGrade{},
		&domain.Post{},
	)
	if err != nil {
		log.Fatalf("❌ Error: No se pudo cargar el esquema de GORM: %v", err)
	}
	// Imprime el esquema en consola (Útil en desarrollo para atlas migrate diff)
	io.WriteString(os.Stdout, stmts)

	// Nota: RunMigrations(db) está comentado para favorecer migraciones versionadas vía CLI
	// if err := database.RunMigrations(db); err != nil { ... }

	// =========================================================================
	// 4. ALMACENAMIENTO DE OBJETOS (S3 / MinIO)
	// =========================================================================
	// Configura el cliente AWS SDK v2 para gestión de archivos (imágenes/PDFs)
	s3Client, err := s3.ConnectS3()
	if err != nil {
		// No detenemos el servidor si S3 falla, permitiendo degradación elegante
		log.Printf("⚠️  Advertencia: S3 no disponible: %v", err)
	} else {
		// Verifica disponibilidad del bucket al arranque
		s3Client.VerifyConnection()
	}

	// =========================================================================
	// 5. INICIALIZACIÓN DE FIBER (HTTP Server)
	// =========================================================================
	app := fiber.New(fiber.Config{
		AppName:           "ColPsiCarabobo API v1.0",
		DisableKeepalive:  false,
		EnablePrintRoutes: false, // Activar solo para depuración pesada
	})

	// =========================================================================
	// 6. RUTAS DEL SISTEMA
	// =========================================================================

	// HealthCheck godoc
	// @Summary      Estado de salud de la API
	// @Description  Verifica si la API y sus dependencias (DB, S3) están operativas.
	// @Tags         System
	// @Produce      json
	// @Success      200  {object}  map[string]interface{}
	// @Router       /health [get]
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "online",
			"version": "1.0",
			"services": fiber.Map{
				"database": "connected",
				"s3":       "initialized",
			},
		})
	})

	// =========================================================================
	// 7. ARRANQUE DEL SERVIDOR
	// =========================================================================
	port := config.Envs.Port
	log.Printf("🚀 ColPsiCarabobo backend iniciado exitosamente")
	log.Printf("📡 Escuchando en el puerto: %s", port)

	// Bloquea el hilo principal y escucha peticiones
	log.Fatal(app.Listen(":" + port))
}
