// api/cmd/api/abs_sync_worker.go
package main

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/repository/postgres"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"gorm.io/gorm"
)

// runABSSyncLoop reconcilia las cuentas de Audiobookshelf en background:
// una pasada inmediata al arrancar y luego cada AbsSyncIntervalH horas.
// Si el intervalo es <= 0, la sincronización queda desactivada.
func runABSSyncLoop(ctx context.Context, db *gorm.DB) {
	intervalH := config.Envs.AbsSyncIntervalH
	if intervalH <= 0 {
		log.Info().Str("component", "psi-abs-sync").Msg("Sincronización ABS desactivada (ABS_SYNC_INTERVAL_HOURS <= 0)")
		return
	}

	psiRepo := postgres.NewPsiRepository(db)
	absSvc := service.NewAudiobookshelfService(
		config.Envs.AbsBaseURL,
		config.Envs.AbsPublicURL,
		config.Envs.AbsAdminUsername,
		config.Envs.AbsAdminPassword,
		config.Envs.AbsPasswordSecret,
	)
	syncSvc := service.NewPsiService(psiRepo, nil, nil)
	syncSvc.SetAudiobookshelf(absSvc)

	run := func() {
		report, err := syncSvc.SyncAudiobookshelfAccounts(ctx)
		if err != nil {
			log.Error().Err(err).Str("component", "psi-abs-sync").Msg("Error en sincronización ABS")
			return
		}
		log.Info().Str("component", "psi-abs-sync").
			Int("created", report.Created).
			Int("deactivated", report.Deactivated).
			Int("skipped", report.Skipped).
			Int("errors", len(report.Errors)).
			Msg("Sincronización ABS ejecutada")
	}

	// Pasada inicial al levantar el contenedor (absorbe imports XLSX recientes).
	go run()

	interval := time.Duration(intervalH) * time.Hour
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Info().Str("component", "psi-abs-sync").Msg("Sincronización ABS detenida")
			return
		case <-ticker.C:
			run()
		}
	}
}
