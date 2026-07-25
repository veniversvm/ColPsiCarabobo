// api/cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	fiberRecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	_ "github.com/veniversvm/ColPsiCarabobo/api/docs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/repository/postgres"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/router"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
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
	numCPU := runtime.NumCPU()
	if numCPU > 2 {
		runtime.GOMAXPROCS(numCPU / 2)
	} else {
		runtime.GOMAXPROCS(numCPU)
	}

	log.Printf("INFO: Intentando cargar configuración...")
	config.InitConfig()

	if config.Envs.JwtLibrarySecret == "" {
		log.Fatal("[ERROR] JWT_LIBRARY_SECRET no está configurado. Defina la variable de entorno.")
	}
	if config.Envs.AbsAdminToken == "" {
		log.Fatal("[ERROR] ABS_ADMIN_TOKEN no está configurado. Defina la variable de entorno.")
	}

	log.Printf("INFO: Configuración cargada. Intentando conectar a DB...")

	// 2. PERSISTENCIA
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("[ERROR] Error crítico: Falló la conexión a PostgreSQL: %v", err)
	}

	// 3. MIGRACIÓN
	// log.Println("Syncing database schema...")
	// err = db.AutoMigrate(
	// 	&domain.TextModel{},
	// 	&domain.UserAdmin{},
	// 	&domain.PsiUserModel{},
	// 	&domain.PsiUserColData{},
	// 	&domain.PsiUserPostGrade{},
	// 	&domain.PsiUserSolvency{},
	// 	&domain.Post{},
	// 	&domain.PsiSpecialtyModel{},
	// 	&domain.PsiUserSocialNetwork{},
	// 	&domain.PsiODeontologia{},
	// 	&domain.PsiObservations{},
	// 	&domain.LoginEvent{},
	// 	&domain.PageView{},
	// 	&domain.SearchEvent{},
	// 	&domain.ProfileView{},
	// 	&domain.ActiveSession{},
	// )
	// if err != nil {
	// 	log.Fatalf("[ERROR] Error: Falló la migración de GORM: %v", err)
	// }

	// SEEDING
	database.SeedAdmin(db)

	// 4. S3
	s3Client, err := s3.ConnectS3(context.Background())
	if err != nil {
		log.Printf("[WARN] Advertencia: S3 no disponible: %v", err)
	} else {
		s3Client.VerifyConnection()
	}

	// 5. ANALYTICS SERVICE — instancia necesaria para el ticker
	analyticsSvc := service.NewAnalyticsService(postgres.NewAnalyticsRepository(db))
	postSvc := service.NewPostService(
		postgres.NewPostRepository(db),
		s3Client,
	)

	// ── Limpieza periódica de sesiones expiradas ──────────────────────────────
	// Corre cada hora en background — elimina ActiveSession con expires_at < now
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			analyticsSvc.CleanExpiredSessions()
			log.Println("[Analytics] Sesiones expiradas limpiadas")
		}
	}()
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			postSvc.PublishScheduled(context.Background())
		}
	}()
	// ─────────────────────────────────────────────────────────────────────────

	// 6. INICIALIZACIÓN DE FIBER
	app := fiber.New(fiber.Config{
		AppName:           "ColPsiCarabobo API v1.0",
		BodyLimit:         20 * 1024 * 1024,
		EnablePrintRoutes: false,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return ctx.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// =========================================================================
	// 7. STACK DE MIDDLEWARES
	// =========================================================================

	app.Use(fiberRecover.New())

	app.Use(requestid.New())
	var origins string = config.Envs.AllowedOrigins
	log.Printf("Allowed origins: %v", origins)
	app.Use(cors.New(cors.Config{
		AllowOrigins:     origins, // desde .env
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Idempotency-Key",
		AllowMethods:     "GET, POST, PATCH, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	app.Use(healthcheck.New(healthcheck.Config{
		LivenessEndpoint:  "/live",
		ReadinessEndpoint: "/ready",
		ReadinessProbe: func(c *fiber.Ctx) bool {
			sqlDB, err := db.DB()
			return err == nil && sqlDB.Ping() == nil
		},
	}))

	app.Use(helmet.New())

	app.Use(compress.New(compress.Config{
		Level: compress.LevelDefault,
	}))
	app.Use(limiter.New(limiter.Config{
		Max:          60,
		Expiration:   1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
	}))

	// =========================================================================
	// 8. RUTAS
	// =========================================================================

	router.SetupRouter(app, db, s3Client)

	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": fmt.Sprintf("Cannot %s %s", c.Method(), c.Path()),
		})
	})

	// 9. ARRANQUE
	port := config.Envs.Port
	utils.PrintColpsiASCII()
	log.Println("¡Bendito Jesus el rey que viene en el nombre del Señor!")
	log.Printf("Ψ ColPsiCarabobo Backend listo en puerto: %s Ψ", port)
	log.Fatal(app.Listen(":" + port))
}
