// api/cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
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
	"github.com/veniversvm/ColPsiCarabobo/api/internal/logger"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/repository/postgres"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/router"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/utils"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/cache"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/database"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/s3"

	"github.com/rs/zerolog/log"
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

	config.InitConfig()
	logger.Init(config.Envs.Environment)

	log.Info().Str("component", "config").Msg("Intentando cargar configuracion")

	if config.Envs.JwtLibrarySecret == "" {
		log.Fatal().Str("component", "config").Msg("JWT_LIBRARY_SECRET no esta configurado. Defina la variable de entorno.")
	}
	if config.Envs.AbsAdminToken == "" {
		log.Fatal().Str("component", "config").Msg("ABS_ADMIN_TOKEN no esta configurado. Defina la variable de entorno.")
	}

	log.Info().Str("component", "config").Msg("Configuracion cargada. Intentando conectar a DB...")

	// 2. PERSISTENCIA
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal().Err(err).Str("component", "database").Msg("Error critico: fallo la conexion a PostgreSQL")
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
		log.Warn().Err(err).Str("component", "s3").Msg("S3 no disponible")
	} else {
		s3Client.VerifyConnection()
	}

	// 5. ANALYTICS SERVICE — instancia necesaria para el ticker
	analyticsSvc := service.NewAnalyticsService(postgres.NewAnalyticsRepository(db))
	postSvc := service.NewPostService(
		postgres.NewPostRepository(db),
		s3Client,
	)

	// ── Contexto raíz para goroutines de background ────────────────────────────
	bgCtx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()

	// ── Limpieza periódica de sesiones expiradas ──────────────────────────────
	// Corre cada hora en background — elimina ActiveSession con expires_at < now
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-bgCtx.Done():
				return
			case <-ticker.C:
				analyticsSvc.CleanExpiredSessions()
				log.Info().Str("component", "analytics").Msg("Sesiones expiradas limpiadas")
			}
		}
	}()
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-bgCtx.Done():
				return
			case <-ticker.C:
				postSvc.PublishScheduled(bgCtx)
			}
		}
	}()
	// ─────────────────────────────────────────────────────────────────────────

	// ── Sincronización de cuentas Audiobookshelf ─────────────────────────────
	// Reconciliación periódica: crea cuentas ABS para solventes y desactiva las
	// de quienes dejaron de serlo. Se dispara UNA vez al arrancar (para absorber
	// imports XLSX recientes) y luego cada AbsSyncIntervalH horas.
	runABSSyncLoop(bgCtx, db)

	// 5.5 CACHÉ COMPARTIDA (Valkey o in-memory) para el directorio público
	appCache := cache.New(config.Envs.ValkeyAddr)

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
	log.Info().Str("component", "cors").Str("origins", origins).Msg("Allowed origins configurados")
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

	router.SetupRouter(app, db, s3Client, appCache)

	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": fmt.Sprintf("Cannot %s %s", c.Method(), c.Path()),
		})
	})

	// 9. ARRANQUE + GRACEFUL SHUTDOWN
	port := config.Envs.Port
	utils.PrintColpsiASCII()
	log.Info().Str("component", "server").Msg("Bendito Jesus el rey que viene en el nombre del Señor!")
	log.Info().Str("component", "server").Str("port", port).Msg("ColPsiCarabobo Backend listo")

	// Canal para capturar señales del sistema operativo
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Arrancar el servidor en una goroutine para no bloquear
	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Listen(":" + port)
	}()

	// Esperar señal de apagado o error del servidor
	select {
	case sig := <-quit:
		log.Warn().Str("component", "shutdown").Str("signal", sig.String()).Msg("Senal recibida, cerrando servidor...")
	case err := <-errCh:
		log.Fatal().Err(err).Str("component", "server").Msg("Servidor fallo al arrancar")
	}

	// Crear contexto con timeout para el cierre ordenado (10 segundos)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Detener goroutines de background
	bgCancel()

	// Cerrar Fiber (espera conexiones activas)
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error().Err(err).Str("component", "shutdown").Msg("Error al cerrar Fiber")
	}

	// Cerrar conexión a base de datos
	sqlDB, err := db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Error().Err(err).Str("component", "shutdown").Msg("Error al cerrar DB")
		} else {
			log.Info().Str("component", "shutdown").Msg("Conexion a DB cerrada")
		}
	}

	log.Info().Str("component", "shutdown").Msg("Servidor cerrado correctamente")
}
