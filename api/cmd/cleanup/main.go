// api/cmd/cleanup/main.go
// Binario independiente para limpieza periódica de keys de sesión expiradas.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/database"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/job"
	"gorm.io/gorm"
)

func main() {
	log.Info().Str("component", "cleanup").Msg("Iniciando servicio de limpieza de keys...")

	config.InitConfig()

	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal().Err(err).Str("component", "cleanup").Msg("No se pudo conectar a la DB")
	}
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	maxAge := 24 * time.Hour
	interval := 30 * time.Minute

	runCleanup(db, maxAge)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Info().Str("component", "cleanup").Dur("interval", interval).Dur("max_age", maxAge).Msg("Tick configurado, keys expiradas serán borradas")

	for {
		select {
		case <-ticker.C:
			runCleanup(db, maxAge)
		case sig := <-sigCh:
			log.Info().Str("component", "cleanup").Str("signal", sig.String()).Msg("Señal recibida, cerrando...")
			return
		}
	}
}

func runCleanup(db *gorm.DB, maxAge time.Duration) {
	result, err := job.CleanExpiredKeys(context.Background(), db, maxAge)
	if err != nil {
		log.Error().Err(err).Str("component", "cleanup").Msg("Error durante limpieza de keys")
		return
	}
	total := result.AdminsCleaned + result.PsiCleaned
	if total > 0 {
		log.Info().Str("component", "cleanup").Int64("total", total).Int64("admins", result.AdminsCleaned).Int64("psi", result.PsiCleaned).Msg("Keys expiradas borradas")
	}
}
