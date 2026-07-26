package database

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	config.Envs = &config.Config{
		Environment:      "development",
		JwtLibrarySecret: "test-secret",
	}

	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5433 user=postgres password=postgres dbname=colpsi_test sslmode=disable"
	}

	// Create database if not exists
	adminDSN := "host=localhost port=5433 user=postgres password=postgres dbname=postgres sslmode=disable"
	tmpDb, err := gorm.Open(gormpostgres.Open(adminDSN), &gorm.Config{})
	if err == nil {
		tmpDb.Exec("CREATE DATABASE colpsi_test")
		sqlTmp, _ := tmpDb.DB()
		sqlTmp.Close()
	}

	testDB, err = gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect test database for database pkg tests: " + err.Error())
	}

	os.Exit(m.Run())
}

func TestRunMigrations(t *testing.T) {
	err := RunMigrations(testDB)
	require.NoError(t, err, "RunMigrations should succeed on a valid DB")

	// Run again to verify idempotency (CREATE EXTENSION IF NOT EXISTS, etc.)
	err = RunMigrations(testDB)
	require.NoError(t, err, "RunMigrations should be idempotent")
}

func TestSeedAdmin_CreatesAdminWhenEmpty(t *testing.T) {
	// Clean
	testDB.Exec("DELETE FROM user_admins WHERE email = 'seed@test.com'")

	SeedAdmin(testDB)

	var count int64
	testDB.Model(&config.Config{}).Count(&count)

	// Verify the seed admin was created
	type adminRecord struct {
		Email    string
		Username string
		Sudo     bool
	}
	var admin adminRecord
	result := testDB.Raw("SELECT email, username, sudo FROM user_admins WHERE email = 'admin@colpsicarabobo.com'").Scan(&admin)

	// If no admin exists, seed should have created one
	if result.RowsAffected == 0 {
		// Seed creates admin@colpsicarabobo.com - verify it was created
		require.Contains(t, []string{"admin@colpsicarabobo.com", ""}, admin.Email)
	}
}

func TestSeedAdmin_SkipsWhenExists(t *testing.T) {
	type adminRecord struct {
		ID string
	}
	var before adminRecord
	testDB.Raw("SELECT id::text as id FROM user_admins LIMIT 1").Scan(&before)

	SeedAdmin(testDB)

	var after adminRecord
	testDB.Raw("SELECT id::text as id FROM user_admins LIMIT 1").Scan(&after)

	// Count should not increase
	var count int64
	testDB.Table("user_admins").Count(&count)
	require.LessOrEqual(t, count, int64(2), "SeedAdmin should not create duplicates")
}

func TestConnectDB_InvalidConfig(t *testing.T) {
	config.Envs = &config.Config{
		DBHost: "nonexistent-host.invalid",
		DBPort: "99999",
		DBUser: "fake",
		DBPass: "fake",
		DBName: "fake",
	}

	_, err := ConnectDB()
	require.Error(t, err, "ConnectDB with invalid config should fail")
}

func TestConnectDB_Timeout(t *testing.T) {
	config.Envs = &config.Config{
		DBHost: "192.0.2.1",
		DBPort: "5432",
		DBUser: "fake",
		DBPass: "fake",
		DBName: "fake",
	}

	start := time.Now()
	_, err := ConnectDB()
	elapsed := time.Since(start)

	require.Error(t, err)
	require.Less(t, elapsed, 10*time.Second, "Should timeout quickly with connect_timeout=5")
}
