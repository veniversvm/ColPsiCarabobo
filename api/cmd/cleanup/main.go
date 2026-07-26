// api/cmd/cleanup/main.go
// Binario independiente para limpieza periódica de keys de sesión expiradas.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/database"
	"github.com/veniversvm/ColPsiCarabobo/api/pkg/job"
	"gorm.io/gorm"
)

func main() {
	log.Println("[CLEANUP] Iniciando servicio de limpieza de keys...")

	config.InitConfig()

	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("[CLEANUP][ERROR] No se pudo conectar a la DB: %v", err)
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

	log.Printf("[CLEANUP] Tick cada %v | Keys > %v serán borradas", interval, maxAge)

	for {
		select {
		case <-ticker.C:
			runCleanup(db, maxAge)
		case sig := <-sigCh:
			log.Printf("[CLEANUP] Señal %v recibida, cerrando...", sig)
			return
		}
	}
}

func runCleanup(db *gorm.DB, maxAge time.Duration) {
	result, err := job.CleanExpiredKeys(context.Background(), db, maxAge)
	if err != nil {
		log.Printf("[CLEANUP][ERROR] %v", err)
		return
	}
	total := result.AdminsCleaned + result.PsiCleaned
	if total > 0 {
		log.Printf("[CLEANUP] %d keys expiradas borradas (admins: %d, psi: %d)", total, result.AdminsCleaned, result.PsiCleaned)
	}
}
