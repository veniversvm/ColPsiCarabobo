// api/cmd/api/main.go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors" // CORREGIDO: Usar el de Fiber, no el de Gin
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	fiberRecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	_ "github.com/veniversvm/ColPsiCarabobo/api/docs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/router"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/database"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"
)

// @title                       ColPsiCarabobo API
// @version                     1.0
// @description                 Backend para la gestión del Colegio de Psicólogos de Carabobo.
// @host                        localhost:8080
// @BasePath                    /api/v1
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @type                        http
// @scheme                      bearer
// @bearerFormat                JWT
func main() {
	// 1. CONFIGURACIÓN
	println("Intentando cargar configuración...") // println (minúscula) es directo
	config.InitConfig()
	println("Configuración cargada. Intentando conectar a DB...")
	// 2. PERSISTENCIA
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("❌ Error crítico: Falló la conexión a PostgreSQL: %v", err)
	}

	// 3. GENERACIÓN DE ESQUEMA (Atlas)
	log.Println("Syncing database schema...")
	err = db.AutoMigrate(
		&domain.TextModel{},
		&domain.UserAdmin{},
		&domain.PsiUserModel{},
		&domain.PsiUserColData{},
		&domain.PsiUserPostGrade{},
		&domain.Post{},
		&domain.PsiSpecialtyModel{},
		&domain.PsiUserSocialNetwork{},
	)
	if err != nil {
		log.Fatalf("❌ Error: Falló la migración de GORM: %v", err)
	}

	// Opcional: Esto déjalo solo si usas el CLI de Atlas para generar archivos .sql
	// stmts, _ := gormschema.New("postgres").Load(...)
	// io.WriteString(os.Stdout, stmts)

	// SEEDING (Ahora sí, con tablas existentes)
	database.SeedAdmin(db)

	// 4. S3
	s3Client, err := s3.ConnectS3()
	if err != nil {
		log.Printf("⚠️  Advertencia: S3 no disponible: %v", err)
	} else {
		s3Client.VerifyConnection()
	}

	// 5. INICIALIZACIÓN DE FIBER
	app := fiber.New(fiber.Config{
		AppName:           "ColPsiCarabobo API v1.0",
		EnablePrintRoutes: true,
		// ErrorHandler global para asegurar que todos los errores salgan como JSON
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return ctx.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// =========================================================================
	// 6. STACK DE MIDDLEWARES (Orden de Precedencia Senior)
	// =========================================================================

	// A. RECUPERACIÓN (Debe envolver a todos los demás para evitar que la app muera)
	app.Use(fiberRecover.New())

	// B. IDENTIFICACIÓN Y CORS
	app.Use(requestid.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000, http://127.0.0.1:3000, http://127.0.0.1",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Idempotency-Key",
		AllowMethods: "GET, POST, PATCH, DELETE, OPTIONS",
	}))

	// C. MONITOREO DE INFRAESTRUCTURA
	app.Use(healthcheck.New(healthcheck.Config{
		LivenessEndpoint:  "/live",
		ReadinessEndpoint: "/ready",
		ReadinessProbe: func(c *fiber.Ctx) bool {
			sqlDB, err := db.DB()
			return err == nil && sqlDB.Ping() == nil
		},
	}))

	// D. SEGURIDAD DE RED
	app.Use(helmet.New())
	app.Use(idempotency.New(idempotency.Config{
		Lifetime:  30 * time.Minute,
		KeyHeader: "X-Idempotency-Key",
	}))

	// E. RENDIMIENTO Y TRÁFICO
	app.Use(compress.New(compress.Config{
		Level: compress.LevelDefault,
	}))
	app.Use(limiter.New(limiter.Config{
		Max:          60,
		Expiration:   1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
	}))

	// =========================================================================
	// 7. RUTAS Y MANEJADOR CATCH-ALL
	// =========================================================================

	router.SetupRouter(app, db, s3Client)

	// Manejador 404 estandarizado
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": fmt.Sprintf("Cannot %s %s", c.Method(), c.Path()),
		})
	})

	// 8. ARRANQUE
	port := config.Envs.Port
	log.Printf("🚀 ColPsiCarabobo Backend listo en puerto: %s", port)
	log.Fatal(app.Listen(":" + port))
}
