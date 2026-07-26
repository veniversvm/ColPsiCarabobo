// api/pkg/job/key_cleanup.go
// Package job contiene tareas periódicas de mantenimiento (background jobs).
package job

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// KeyCleanupResult contiene el resultado de una ejecución del job de limpieza.
type KeyCleanupResult struct {
	AdminsCleaned int64
	PsiCleaned    int64
}

// CleanExpiredKeys elimina las claves de sesión (key) de usuarios cuyo UUID v7
// tenga un timestamp superior a la edad máxima permitida.
func CleanExpiredKeys(ctx context.Context, db *gorm.DB, maxAge time.Duration) (KeyCleanupResult, error) {
	var result KeyCleanupResult
	cutoff := time.Now().Add(-maxAge)

	adminKeys, err := fetchKeys(ctx, db, "user_admins")
	if err != nil {
		return result, err
	}
	for _, k := range adminKeys {
		if isKeyExpired(k.Key, cutoff) {
			if err := clearKey(ctx, db, "user_admins", k.ID); err != nil {
				log.Printf("[CLEANUP][WARN] Error borrando key de admin %s: %v", k.ID, err)
				continue
			}
			result.AdminsCleaned++
		}
	}

	psiKeys, err := fetchKeys(ctx, db, "psi_users")
	if err != nil {
		return result, err
	}
	for _, k := range psiKeys {
		if isKeyExpired(k.Key, cutoff) {
			if err := clearKey(ctx, db, "psi_users", k.ID); err != nil {
				log.Printf("[CLEANUP][WARN] Error borrando key de psi %s: %v", k.ID, err)
				continue
			}
			result.PsiCleaned++
		}
	}

	return result, nil
}

type keyRecord struct {
	ID  string
	Key string
}

func fetchKeys(ctx context.Context, db *gorm.DB, table string) ([]keyRecord, error) {
	var records []keyRecord
	err := db.WithContext(ctx).Raw(
		"SELECT id::text, key FROM " + table + " WHERE key != '' AND deleted_at IS NULL",
	).Scan(&records).Error
	return records, err
}

func clearKey(ctx context.Context, db *gorm.DB, table, id string) error {
	return db.WithContext(ctx).Exec(
		"UPDATE "+table+" SET key = '', updated_at = ? WHERE id = ?::uuid",
		time.Now(), id,
	).Error
}

// isKeyExpired parsea un UUID v7 y verifica si su timestamp es anterior al cutoff.
func isKeyExpired(key string, cutoff time.Time) bool {
	parsed, err := uuid.Parse(key)
	if err != nil {
		return true
	}
	sec, nsec := parsed.Time().UnixTime()
	ts := time.Unix(sec, nsec)
	return ts.Before(cutoff)
}
