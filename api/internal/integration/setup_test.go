package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	repopostgres "github.com/veniversvm/ColPsiCarabobo/api/internal/repository/postgres"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/router"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
	"golang.org/x/crypto/bcrypt"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	config.Envs = &config.Config{
		Environment:      "development",
		JwtLibrarySecret: "test-library-secret",
		AbsAdminToken:    "test-abs-token",
	}

	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5433 user=postgres password=postgres dbname=colpsi_test sslmode=disable"
	}

	adminDSN := "host=localhost port=5433 user=postgres password=postgres dbname=postgres sslmode=disable"
	tmpDb, _ := gorm.Open(gormpostgres.Open(adminDSN), &gorm.Config{})
	tmpDb.Exec("CREATE DATABASE colpsi_test")
	sqlTmp, _ := tmpDb.DB()
	sqlTmp.Close()

	var err error
	testDB, err = gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect test database: " + err.Error())
	}

	testDB.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"")
	testDB.Exec("CREATE EXTENSION IF NOT EXISTS unaccent")

	err = testDB.AutoMigrate(
		&domain.UserAdmin{},
		&domain.PsiUserModel{},
		&domain.PsiUserColData{},
		&domain.PsiUserPostGrade{},
		&domain.PsiUserSocialNetwork{},
		&domain.PsiUserSolvency{},
		&domain.PsiSpecialtyModel{},
		&domain.TextModel{},
		&domain.Post{},
		&domain.LoginEvent{},
		&domain.PageView{},
		&domain.SearchEvent{},
		&domain.ProfileView{},
		&domain.ActiveSession{},
	)
	if err != nil {
		panic("failed to auto-migrate: " + err.Error())
	}

	exitCode := m.Run()

	tables := []string{
		"psi_user_social_networks", "psi_user_post_grades", "psi_user_col_data",
		"psi_user_solvency", "psi_users", "user_admins", "psi_specialty_models",
		"posts", "text_models", "login_events", "active_sessions",
		"page_views", "search_events", "profile_views",
	}
	for _, tbl := range tables {
		testDB.Exec("TRUNCATE TABLE " + tbl + " RESTART IDENTITY CASCADE")
	}

	os.Exit(exitCode)
}

// =========================================================================
// DB CLEANUP
// =========================================================================

func truncateAll(t *testing.T) {
	t.Helper()
	tables := []string{
		"psi_user_social_networks",
		"psi_user_post_grades",
		"psi_user_col_data",
		"psi_user_solvency",
		"psi_users",
		"user_admins",
		"psi_specialty_models",
		"posts",
		"text_models",
		"login_events",
		"active_sessions",
		"page_views",
		"search_events",
		"profile_views",
	}
	for _, table := range tables {
		testDB.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE")
	}
}

// =========================================================================
// PASSWORD HELPER
// =========================================================================

func hashPassword(t *testing.T, plain string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	require.NoError(t, err)
	return string(hash)
}

// =========================================================================
// SEED FUNCTIONS
// =========================================================================

func seedSudo(t *testing.T) *domain.UserAdmin {
	t.Helper()
	admin := &domain.UserAdmin{
		Credentials: domain.Credentials{
			Username: "sudo",
			Email:    "sudo@test.com",
			Password: hashPassword(t, "Sudo123!@#"),
			Key:      uuid.Must(uuid.NewV7()).String(),
			IsActive: true,
		},
		Sudo:                 true,
		CanCreateAdmin:       true,
		CanUpdateAdmin:       true,
		CanDeleteAdmin:       true,
		CanCreatePsi:         true,
		CanUpdatePsi:         true,
		CanDeletePsi:         true,
		CanPublish:           true,
		CanUpdatePublish:     true,
		CanDeletePublish:     true,
		CanCreateTags:        true,
		CanEditTags:          true,
		CanDeleteTags:        true,
		CanReadNotifications: true,
	}
	require.NoError(t, testDB.Create(admin).Error)
	return admin
}

func seedAdmin(t *testing.T, email string, perms map[string]bool) *domain.UserAdmin {
	t.Helper()
	admin := &domain.UserAdmin{
		Credentials: domain.Credentials{
			Username: email,
			Email:    email,
			Password: hashPassword(t, "Admin123!@#"),
			Key:      uuid.Must(uuid.NewV7()).String(),
			IsActive: true,
		},
		CanCreatePsi:     perms["can_create_psi"],
		CanUpdatePsi:     perms["can_update_psi"],
		CanDeletePsi:     perms["can_delete_psi"],
		CanCreateAdmin:   perms["can_create_admin"],
		CanUpdateAdmin:   perms["can_update_admin"],
		CanDeleteAdmin:   perms["can_delete_admin"],
		CanPublish:       perms["can_publish"],
		CanUpdatePublish: perms["can_update_publish"],
		CanDeletePublish: perms["can_delete_publish"],
		CanCreateTags:    perms["can_create_tags"],
		CanEditTags:      perms["can_edit_tags"],
		CanDeleteTags:    perms["can_delete_tags"],
	}
	require.NoError(t, testDB.Create(admin).Error)
	return admin
}

func seedPsi(t *testing.T, fpv int, ci int, username string) *domain.PsiUserModel {
	t.Helper()
	bioText := &domain.TextModel{Content: "<p>Bio for " + username + "</p>"}
	require.NoError(t, testDB.Create(bioText).Error)

	psi := &domain.PsiUserModel{
		Credentials: domain.Credentials{
			Username: username,
			Email:    username + "@test.com",
			Password: hashPassword(t, "Psi123!@#"),
			Key:      uuid.Must(uuid.NewV7()).String(),
			IsActive: true,
		},
		FirstName:            "Maria",
		LastName:             "Garcia",
		FPV:                  fpv,
		CI:                   ci,
		Nationality:          "V",
		Genre:                "F",
		BornDate:             time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		ContactPhone:         "+58-241-1234567",
		ContactCellPhone:     "+58-412-1234567",
		ContactEmail:         username + "@contact.com",
		MunicipalityCarabobo: "Valencia",
		PrimaryWorkArea:      "Psicologia Clinica",
		Solvent:              true,
		ProofOfLife:          true,
		ShowContactEmail:     true,
		BioTextID:            bioText.ID,
	}
	require.NoError(t, testDB.Create(psi).Error)
	return psi
}

func seedSpecialty(t *testing.T, name string) *domain.PsiSpecialtyModel {
	t.Helper()
	spec := &domain.PsiSpecialtyModel{
		Name:        name,
		Description: "Description for " + name,
		Active:      true,
	}
	require.NoError(t, testDB.Create(spec).Error)
	return spec
}

func seedPost(t *testing.T, title string, status domain.PostStatus, postType string) *domain.Post {
	t.Helper()
	textContent := &domain.TextModel{
		Content: "<p>Contenido de prueba para " + title + "</p>",
	}
	require.NoError(t, testDB.Create(textContent).Error)

	post := &domain.Post{
		Title:      title,
		Type:       postType,
		Status:     status,
		TextID:     textContent.ID,
		ImageS3Key: "images/test.jpg",
	}
	require.NoError(t, testDB.Create(post).Error)
	return post
}

// =========================================================================
// JWT TOKEN GENERATOR
// =========================================================================

func generateToken(t *testing.T, userID uuid.UUID, role, key string) string {
	t.Helper()
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(key))
	require.NoError(t, err)
	return signed
}

// =========================================================================
// APP BUILDER
// =========================================================================

func buildTestApp(db *gorm.DB) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	analyticsRepo := repopostgres.NewAnalyticsRepository(db)
	analyticsSvc := service.NewAnalyticsService(analyticsRepo)

	app.Use(middleware.AnalyticsMiddleware(analyticsSvc))
	app.Use(helmet.New())

	adminRepo := repopostgres.NewAdminRepository(db)
	psiRepo := repopostgres.NewPsiRepository(db)
	postRepo := repopostgres.NewPostRepository(db)
	specialtyRepo := repopostgres.NewSpecialtyRepository(db)

	api := app.Group("/api/v1")

	router.SetupAdminRoutes(api, adminRepo, psiRepo, analyticsSvc, nil)
	router.SetupPsiRoutes(api, psiRepo, adminRepo, nil, analyticsSvc, nil, nil)
	router.SetupSpecialtyRoutes(api, psiRepo, adminRepo, specialtyRepo, analyticsSvc)
	router.SetupPostRoutes(api, adminRepo, psiRepo, postRepo, nil, analyticsSvc)

	return app
}

// =========================================================================
// HTTP HELPERS
// =========================================================================

func authHeader(token string) string {
	return "Bearer " + token
}

func decodeBody(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	return result
}

func decodeBodyArray(t *testing.T, resp *http.Response) []interface{} {
	t.Helper()
	var result []interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	return result
}

func ptrString(s string) *string { return &s }
func ptrBool(b bool) *bool       { return &b }

// fullAdminPerms returns a permissions map with all flags set to true.
func fullAdminPerms() map[string]bool {
	return map[string]bool{
		"can_create_psi":     true,
		"can_update_psi":     true,
		"can_delete_psi":     true,
		"can_create_admin":   true,
		"can_update_admin":   true,
		"can_delete_admin":   true,
		"can_publish":        true,
		"can_update_publish": true,
		"can_delete_publish": true,
		"can_create_tags":    true,
		"can_edit_tags":      true,
		"can_delete_tags":    true,
	}
}

// loginAdmin performs a POST /api/v1/auth/login and returns the token.
func loginAdmin(t *testing.T, app *fiber.App, identifier, password string) string {
	t.Helper()
	body := `{"identifier":"` + identifier + `","password":"` + password + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	data := decodeBody(t, resp)
	return data["token"].(string)
}

// loginPsi performs a POST /api/v1/psi/login and returns the token.
func loginPsi(t *testing.T, app *fiber.App, identifier, password string) string {
	t.Helper()
	body := `{"identifier":"` + identifier + `","password":"` + password + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/psi/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	data := decodeBody(t, resp)
	return data["token"].(string)
}
