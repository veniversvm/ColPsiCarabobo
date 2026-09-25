package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupAuditTestDB conecta (o crea) la base de test y migra la tabla de la bitácora.
func setupAuditTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5433 user=postgres password=postgres dbname=colpsi_test sslmode=disable"
	}

	adminDSN := "host=localhost port=5433 user=postgres password=postgres dbname=postgres sslmode=disable"
	if tmpDb, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{}); err == nil {
		tmpDb.Exec("CREATE DATABASE colpsi_test")
		sqlTmp, _ := tmpDb.DB()
		sqlTmp.Close()
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";")
	err = db.AutoMigrate(&domain.ApiChangeLog{})
	require.NoError(t, err)
	return db
}

func TestAuditRepo_CreateBatchYList(t *testing.T) {
	db := setupAuditTestDB(t)
	db.Exec("DELETE FROM api_change_logs")

	r := NewAuditRepository(db)
	ctx := context.Background()

	psiID := uuid.New()
	actorID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	batch := []domain.ApiChangeLog{
		{
			ID:            uuid.Must(uuid.NewV7()),
			Entity:        domain.AuditEntityPsi,
			EntityID:      psiID.String(),
			EntityLabel:   "Juan Pérez (FPV 12345)",
			Action:        domain.AuditActionLogin,
			ActorID:       actorID,
			ActorRole:     "psi",
			ActorUsername: "juanperez",
			IP:            "127.0.0.1",
			CreatedAt:     now,
		},
		{
			ID:            uuid.Must(uuid.NewV7()),
			Entity:        domain.AuditEntityStaff,
			EntityID:      actorID.String(),
			EntityLabel:   "admin",
			Action:        domain.AuditActionCreate,
			ActorID:       actorID,
			ActorRole:     "sudo",
			ActorUsername: "admin",
			CreatedAt:     now.Add(time.Second),
		},
		{
			ID:            uuid.Must(uuid.NewV7()),
			Entity:        domain.AuditEntityPsi,
			EntityID:      psiID.String(),
			EntityLabel:   "Juan Pérez (FPV 12345)",
			Action:        domain.AuditActionLogout,
			ActorID:       actorID,
			ActorRole:     "psi",
			ActorUsername: "juanperez",
			CreatedAt:     now.Add(2 * time.Second),
		},
	}

	require.NoError(t, r.CreateBatch(ctx, batch))

	t.Run("listado_sin_filtros", func(t *testing.T) {
		logs, total, err := r.List(ctx, domain.AuditLogFilters{Page: 1, Limit: 10})
		require.NoError(t, err)
		require.Equal(t, int64(3), total)
		require.Len(t, logs, 3)
		// Orden descendente por fecha.
		require.Equal(t, now.Add(2*time.Second), logs[0].CreatedAt.UTC())
	})

	t.Run("historial_por_entidad", func(t *testing.T) {
		logs, total, err := r.List(ctx, domain.AuditLogFilters{
			Entity:   domain.AuditEntityPsi,
			EntityID: psiID.String(),
			Page:     1,
			Limit:    10,
		})
		require.NoError(t, err)
		require.Equal(t, int64(2), total)
		require.Len(t, logs, 2)
	})

	t.Run("filtro_por_accion", func(t *testing.T) {
		_, total, err := r.List(ctx, domain.AuditLogFilters{Action: domain.AuditActionLogin, Page: 1, Limit: 10})
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
	})

	t.Run("filtro_por_actor", func(t *testing.T) {
		_, total, err := r.List(ctx, domain.AuditLogFilters{ActorID: &actorID, Page: 1, Limit: 10})
		require.NoError(t, err)
		require.Equal(t, int64(3), total)
	})

	t.Run("rango_de_fechas", func(t *testing.T) {
		from := now
		to := now.Add(time.Second)
		_, total, err := r.List(ctx, domain.AuditLogFilters{From: &from, To: &to, Page: 1, Limit: 10})
		require.NoError(t, err)
		require.Equal(t, int64(2), total)
	})

	t.Run("busqueda_por_texto", func(t *testing.T) {
		_, total, err := r.List(ctx, domain.AuditLogFilters{Q: "juan", Page: 1, Limit: 10})
		require.NoError(t, err)
		require.Equal(t, int64(2), total)
	})

	t.Run("paginacion", func(t *testing.T) {
		logs, total, err := r.List(ctx, domain.AuditLogFilters{Page: 2, Limit: 2})
		require.NoError(t, err)
		require.Equal(t, int64(3), total)
		require.Len(t, logs, 1)
	})
}

func TestAuditRepo_StatsYPurge(t *testing.T) {
	db := setupAuditTestDB(t)
	db.Exec("DELETE FROM api_change_logs")

	r := NewAuditRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()
	for _, entidad := range []string{domain.AuditEntityPsi, domain.AuditEntityStaff} {
		for _, accion := range []string{domain.AuditActionLogin, domain.AuditActionCreate} {
			require.NoError(t, r.CreateBatch(ctx, []domain.ApiChangeLog{{
				ID:        uuid.Must(uuid.NewV7()),
				Entity:    entidad,
				Action:    accion,
				ActorID:   uuid.New(),
				CreatedAt: now,
			}}))
		}
	}

	stats, err := r.Stats(ctx, now.Add(-time.Hour))
	require.NoError(t, err)
	require.Len(t, stats, 4)
	seen := map[string]int64{}
	for _, s := range stats {
		seen[s.Entity+"|"+s.Action] = s.Count
	}
	require.Equal(t, int64(1), seen[domain.AuditEntityPsi+"|"+domain.AuditActionLogin])
	require.Equal(t, int64(1), seen[domain.AuditEntityStaff+"|"+domain.AuditActionCreate])

	// Purge: borrar solo lo anterior a una hora.
	old := domain.ApiChangeLog{
		ID:        uuid.Must(uuid.NewV7()),
		Entity:    domain.AuditEntityAuth,
		Action:    domain.AuditActionLogout,
		ActorID:   uuid.New(),
		CreatedAt: now.Add(-48 * time.Hour),
	}
	require.NoError(t, r.CreateBatch(ctx, []domain.ApiChangeLog{old}))
	n, err := r.PurgeOlderThan(ctx, now.Add(-24*time.Hour))
	require.NoError(t, err)
	require.Equal(t, int64(1), n)
	// Los recientes siguen vivos.
	_, total, err := r.List(ctx, domain.AuditLogFilters{Page: 1, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, int64(4), total)
}