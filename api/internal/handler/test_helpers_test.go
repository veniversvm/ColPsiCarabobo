package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/middleware"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/request_structs"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

const testSecret = "test-secret-key-for-handler-tests"

// =========================================================================
// TEST INIT
// =========================================================================

func init() {
	config.Envs = &config.Config{Environment: "test", JwtLibrarySecret: "test-library-secret"}
}

// =========================================================================
// MOCK ADMIN REPOSITORY
// =========================================================================

type mockAdminRepo struct {
	domain.UserAdminRepository
	GetByIDFunc         func(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error)
	GetByIdentifierFunc func(ctx context.Context, identifier string) (*domain.UserAdmin, error)
	CreateFunc          func(ctx context.Context, user *domain.UserAdmin) error
	UpdateFunc          func(ctx context.Context, user *domain.UserAdmin) error
	UpdateKeyFunc       func(ctx context.Context, user *domain.UserAdmin) error
	DeleteFunc          func(ctx context.Context, id uuid.UUID) error
	ListFunc            func(ctx context.Context, active *bool, search string, page, limit int) ([]domain.UserAdmin, int64, error)
	CountSudosFunc      func(ctx context.Context) (int64, error)
}

func (m *mockAdminRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, domain.ErrPsiNotFound
}
func (m *mockAdminRepo) GetByIdentifier(ctx context.Context, identifier string) (*domain.UserAdmin, error) {
	if m.GetByIdentifierFunc != nil {
		return m.GetByIdentifierFunc(ctx, identifier)
	}
	return nil, domain.ErrPsiNotFound
}
func (m *mockAdminRepo) Create(ctx context.Context, user *domain.UserAdmin) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil
}
func (m *mockAdminRepo) Update(ctx context.Context, user *domain.UserAdmin) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	return nil
}
func (m *mockAdminRepo) UpdateKey(ctx context.Context, user *domain.UserAdmin) error {
	if m.UpdateKeyFunc != nil {
		return m.UpdateKeyFunc(ctx, user)
	}
	return nil
}
func (m *mockAdminRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}
func (m *mockAdminRepo) List(ctx context.Context, active *bool, search string, page, limit int) ([]domain.UserAdmin, int64, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, active, search, page, limit)
	}
	return nil, 0, nil
}
func (m *mockAdminRepo) CountSudos(ctx context.Context) (int64, error) {
	if m.CountSudosFunc != nil {
		return m.CountSudosFunc(ctx)
	}
	return 0, nil
}

// =========================================================================
// MOCK PSI REPOSITORY
// =========================================================================

type mockPsiRepo struct {
	domain.PsiUserRepository
	GetByIDFunc                    func(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error)
	GetByFPVFunc                   func(ctx context.Context, id int) (domain.PsiUserModel, error)
	GetByIdentifierFunc            func(ctx context.Context, identifier string) (*domain.PsiUserModel, error)
	UpdateFunc                     func(ctx context.Context, psi *domain.PsiUserModel, colData *domain.PsiUserColData, bioText *domain.TextModel, solvencies []domain.PsiUserSolvency) error
	UpdateKeyFunc                  func(ctx context.Context, psi *domain.PsiUserModel) error
	UpdateAudioBookShellIDFunc     func(ctx context.Context, psi *domain.PsiUserModel) error
	SearchDirectoryFunc            func(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error)
	GetSitemapDataFunc             func(ctx context.Context) ([]domain.PsiUserModel, error)
	GetPsiUserColDataFunc          func(ctx context.Context, psiID uuid.UUID) (*domain.PsiUserColData, error)
	CreatePostGradeFunc            func(ctx context.Context, pg *domain.PsiUserPostGrade) error
	CreateSocialNetworkFunc        func(ctx context.Context, sn *domain.PsiUserSocialNetwork) error
	UpdateSocialNetworkFunc        func(ctx context.Context, sn *domain.PsiUserSocialNetwork) error
	DeleteSocialNetworkFunc        func(ctx context.Context, id uuid.UUID) error
	GetSocialNetworkByIDFunc       func(ctx context.Context, id uuid.UUID) (*domain.PsiUserSocialNetwork, error)
	CountSocialNetworksByPsiIDFunc func(ctx context.Context, psiID uuid.UUID) (int64, error)
	CreateDeontologiaFunc            func(ctx context.Context, entry *domain.PsiODeontologia) error
	ListDeontologiaByPsiIDFunc       func(ctx context.Context, psiID uuid.UUID) ([]domain.PsiODeontologia, error)
	GetDeontologiaByIDFunc           func(ctx context.Context, id uuid.UUID) (*domain.PsiODeontologia, error)
	DeleteDeontologiaFunc            func(ctx context.Context, id uuid.UUID) error
}

func (m *mockPsiRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, domain.ErrPsiNotFound
}
func (m *mockPsiRepo) GetByFPV(ctx context.Context, id int) (domain.PsiUserModel, error) {
	if m.GetByFPVFunc != nil {
		return m.GetByFPVFunc(ctx, id)
	}
	return domain.PsiUserModel{}, domain.ErrPsiNotFound
}
func (m *mockPsiRepo) GetByIdentifier(ctx context.Context, identifier string) (*domain.PsiUserModel, error) {
	if m.GetByIdentifierFunc != nil {
		return m.GetByIdentifierFunc(ctx, identifier)
	}
	return nil, domain.ErrPsiNotFound
}
func (m *mockPsiRepo) Update(ctx context.Context, psi *domain.PsiUserModel, colData *domain.PsiUserColData, bioText *domain.TextModel, solvencies []domain.PsiUserSolvency) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, psi, colData, bioText, solvencies)
	}
	return nil
}
func (m *mockPsiRepo) UpdateKey(ctx context.Context, psi *domain.PsiUserModel) error {
	if m.UpdateKeyFunc != nil {
		return m.UpdateKeyFunc(ctx, psi)
	}
	return nil
}
func (m *mockPsiRepo) UpdateAudioBookShellID(ctx context.Context, psi *domain.PsiUserModel) error {
	if m.UpdateAudioBookShellIDFunc != nil {
		return m.UpdateAudioBookShellIDFunc(ctx, psi)
	}
	return nil
}
func (m *mockPsiRepo) SearchDirectory(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
	if m.SearchDirectoryFunc != nil {
		return m.SearchDirectoryFunc(ctx, filter)
	}
	return nil, 0, nil
}
func (m *mockPsiRepo) GetSitemapData(ctx context.Context) ([]domain.PsiUserModel, error) {
	if m.GetSitemapDataFunc != nil {
		return m.GetSitemapDataFunc(ctx)
	}
	return nil, nil
}
func (m *mockPsiRepo) GetPsiUserColData(ctx context.Context, psiID uuid.UUID) (*domain.PsiUserColData, error) {
	if m.GetPsiUserColDataFunc != nil {
		return m.GetPsiUserColDataFunc(ctx, psiID)
	}
	return nil, domain.ErrPsiNotFound
}
func (m *mockPsiRepo) CreatePostGrade(ctx context.Context, pg *domain.PsiUserPostGrade) error {
	if m.CreatePostGradeFunc != nil {
		return m.CreatePostGradeFunc(ctx, pg)
	}
	return nil
}
func (m *mockPsiRepo) CreateSocialNetwork(ctx context.Context, sn *domain.PsiUserSocialNetwork) error {
	if m.CreateSocialNetworkFunc != nil {
		return m.CreateSocialNetworkFunc(ctx, sn)
	}
	return nil
}
func (m *mockPsiRepo) UpdateSocialNetwork(ctx context.Context, sn *domain.PsiUserSocialNetwork) error {
	if m.UpdateSocialNetworkFunc != nil {
		return m.UpdateSocialNetworkFunc(ctx, sn)
	}
	return nil
}
func (m *mockPsiRepo) DeleteSocialNetwork(ctx context.Context, id uuid.UUID) error {
	if m.DeleteSocialNetworkFunc != nil {
		return m.DeleteSocialNetworkFunc(ctx, id)
	}
	return nil
}
func (m *mockPsiRepo) GetSocialNetworkByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserSocialNetwork, error) {
	if m.GetSocialNetworkByIDFunc != nil {
		return m.GetSocialNetworkByIDFunc(ctx, id)
	}
	return nil, domain.ErrPsiNotFound
}
func (m *mockPsiRepo) CountSocialNetworksByPsiID(ctx context.Context, psiID uuid.UUID) (int64, error) {
	if m.CountSocialNetworksByPsiIDFunc != nil {
		return m.CountSocialNetworksByPsiIDFunc(ctx, psiID)
	}
	return 0, nil
}
func (m *mockPsiRepo) CreateDeontologia(ctx context.Context, entry *domain.PsiODeontologia) error {
	if m.CreateDeontologiaFunc != nil {
		return m.CreateDeontologiaFunc(ctx, entry)
	}
	return nil
}
func (m *mockPsiRepo) ListDeontologiaByPsiID(ctx context.Context, psiID uuid.UUID) ([]domain.PsiODeontologia, error) {
	if m.ListDeontologiaByPsiIDFunc != nil {
		return m.ListDeontologiaByPsiIDFunc(ctx, psiID)
	}
	return nil, nil
}
func (m *mockPsiRepo) GetDeontologiaByID(ctx context.Context, id uuid.UUID) (*domain.PsiODeontologia, error) {
	if m.GetDeontologiaByIDFunc != nil {
		return m.GetDeontologiaByIDFunc(ctx, id)
	}
	return nil, domain.ErrDeontologiaNotFound
}
func (m *mockPsiRepo) DeleteDeontologia(ctx context.Context, id uuid.UUID) error {
	if m.DeleteDeontologiaFunc != nil {
		return m.DeleteDeontologiaFunc(ctx, id)
	}
	return nil
}
func (m *mockPsiRepo) CreateWithColData(ctx context.Context, psi *domain.PsiUserModel, colData *domain.PsiUserColData, solvencies []domain.PsiUserSolvency, postgrades []domain.PsiUserPostGrade) error {
	return nil
}
func (m *mockPsiRepo) ValidateUniqueCredentials(ctx context.Context, username, email string, excludeID uuid.UUID) error {
	return nil
}
func (m *mockPsiRepo) UpdatePublicProfile(ctx context.Context, psi *domain.PsiUserModel, colData *domain.PsiUserColData, bioText *domain.TextModel) error {
	return nil
}
func (m *mockPsiRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockPsiRepo) Search(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]domain.PsiUserModel, int64, error) {
	return nil, 0, nil
}
func (m *mockPsiRepo) SearchAdmin(ctx context.Context, filter request_structs.PsiDirectoryFilterDTO) ([]domain.PsiUserModel, int64, error) {
	return nil, 0, nil
}
func (m *mockPsiRepo) GetPostGradeByID(ctx context.Context, id uuid.UUID) (*domain.PsiUserPostGrade, error) {
	return nil, nil
}
func (m *mockPsiRepo) UpdatePostGrade(ctx context.Context, pg *domain.PsiUserPostGrade) error {
	return nil
}
func (m *mockPsiRepo) CreateSolvency(ctx context.Context, pg *domain.PsiUserSolvency) error {
	return nil
}
func (m *mockPsiRepo) GetSolvencies(ctx context.Context, id uuid.UUID) ([]domain.PsiUserSolvency, error) {
	return nil, nil
}
func (m *mockPsiRepo) CreateOrUpdateSolvencies(ctx context.Context, solvencies []domain.PsiUserSolvency) error {
	return nil
}
func (m *mockPsiRepo) GetTextContentByID(ctx context.Context, id uuid.UUID) (string, error) {
	return "", nil
}

// =========================================================================
// MOCK SPECIALTY REPOSITORY
// =========================================================================

type mockSpecialtyRepo struct {
	domain.SpecialtyRepository
	CreateFunc       func(ctx context.Context, s *domain.PsiSpecialtyModel) error
	GetAllFunc       func(ctx context.Context, status string) ([]domain.PsiSpecialtyModel, error)
	GetByIDFunc      func(ctx context.Context, id uint32, active bool) (*domain.PsiSpecialtyModel, error)
	GetByAdminIDFunc func(ctx context.Context, id uint32) (*domain.PsiSpecialtyModel, error)
	UpdateFunc       func(ctx context.Context, s *domain.PsiSpecialtyModel) error
	DeleteFunc       func(ctx context.Context, id uint32) error
	CountFunc        func(ctx context.Context, actives *bool) (int64, error)
	GetAllAdminFunc  func(ctx context.Context) ([]domain.PsiSpecialtyModel, error)
}

func (m *mockSpecialtyRepo) Create(ctx context.Context, s *domain.PsiSpecialtyModel) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, s)
	}
	return nil
}
func (m *mockSpecialtyRepo) GetAll(ctx context.Context, status string) ([]domain.PsiSpecialtyModel, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx, status)
	}
	return nil, nil
}
func (m *mockSpecialtyRepo) GetByID(ctx context.Context, id uint32, active bool) (*domain.PsiSpecialtyModel, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id, active)
	}
	return nil, domain.ErrPsiNotFound
}
func (m *mockSpecialtyRepo) GetByAdminID(ctx context.Context, id uint32) (*domain.PsiSpecialtyModel, error) {
	if m.GetByAdminIDFunc != nil {
		return m.GetByAdminIDFunc(ctx, id)
	}
	return nil, domain.ErrPsiNotFound
}
func (m *mockSpecialtyRepo) Update(ctx context.Context, s *domain.PsiSpecialtyModel) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, s)
	}
	return nil
}
func (m *mockSpecialtyRepo) Delete(ctx context.Context, id uint32) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}
func (m *mockSpecialtyRepo) Count(ctx context.Context, actives *bool) (int64, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx, actives)
	}
	return 0, nil
}
func (m *mockSpecialtyRepo) GetAllAdmin(ctx context.Context) ([]domain.PsiSpecialtyModel, error) {
	if m.GetAllAdminFunc != nil {
		return m.GetAllAdminFunc(ctx)
	}
	return nil, nil
}

// =========================================================================
// MOCK POST REPOSITORY
// =========================================================================

type mockPostRepo struct {
	domain.PostRepository
	CreateFunc          func(ctx context.Context, post *domain.Post, content *domain.TextModel) error
	UpdateFunc          func(ctx context.Context, post *domain.Post, content *domain.TextModel) error
	DeleteFunc          func(ctx context.Context, id uuid.UUID) error
	GetByIDFunc         func(ctx context.Context, id uuid.UUID) (*domain.Post, error)
	ListFunc            func(ctx context.Context, filter domain.PostFilter, page, limit int) ([]domain.Post, int64, error)
	GetSitemapPostsFunc func(ctx context.Context) ([]domain.Post, error)
}

func (m *mockPostRepo) Create(ctx context.Context, post *domain.Post, content *domain.TextModel) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, post, content)
	}
	return nil
}
func (m *mockPostRepo) Update(ctx context.Context, post *domain.Post, content *domain.TextModel) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, post, content)
	}
	return nil
}
func (m *mockPostRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}
func (m *mockPostRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, domain.ErrPsiNotFound
}
func (m *mockPostRepo) List(ctx context.Context, filter domain.PostFilter, page, limit int) ([]domain.Post, int64, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filter, page, limit)
	}
	return nil, 0, nil
}
func (m *mockPostRepo) GetSitemapPosts(ctx context.Context) ([]domain.Post, error) {
	if m.GetSitemapPostsFunc != nil {
		return m.GetSitemapPostsFunc(ctx)
	}
	return nil, nil
}

// =========================================================================
// MOCK MAIL SERVICE (implements IMailService interface)
// =========================================================================

type mockMailService struct {
	SendEmailFunc func(to, subject, templateName string, data interface{}) error
}

func (m *mockMailService) SendEmail(to, subject, templateName string, data interface{}) error {
	if m.SendEmailFunc != nil {
		return m.SendEmailFunc(to, subject, templateName, data)
	}
	return nil
}

// =========================================================================
// MOCK ANALYTICS REPOSITORY
// =========================================================================

type mockAnalyticsRepo struct {
	domain.AnalyticsRepository
	GetDashboardStatsFunc func(ctx context.Context) (*domain.DashboardStats, error)
}

func (m *mockAnalyticsRepo) CreateLoginEvent(context.Context, domain.LoginEvent) error { return nil }
func (m *mockAnalyticsRepo) UpsertActiveSession(context.Context, domain.ActiveSession) error {
	return nil
}
func (m *mockAnalyticsRepo) DeleteActiveSession(context.Context, uuid.UUID) error { return nil }
func (m *mockAnalyticsRepo) UpdateSessionHeartbeat(context.Context, uuid.UUID, time.Time, time.Time) error {
	return nil
}
func (m *mockAnalyticsRepo) CreateSearchEvent(context.Context, domain.SearchEvent) error { return nil }
func (m *mockAnalyticsRepo) CreateProfileView(context.Context, domain.ProfileView) error { return nil }
func (m *mockAnalyticsRepo) CreatePageView(context.Context, domain.PageView) error       { return nil }
func (m *mockAnalyticsRepo) CountRecentPageViews(context.Context, string, time.Time) (int64, error) {
	return 0, nil
}
func (m *mockAnalyticsRepo) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	if m.GetDashboardStatsFunc != nil {
		return m.GetDashboardStatsFunc(ctx)
	}
	return &domain.DashboardStats{}, nil
}
func (m *mockAnalyticsRepo) DeletePageViewsOlderThan(context.Context, time.Time) error { return nil }
func (m *mockAnalyticsRepo) DeleteSearchEventsOlderThan(context.Context, time.Time) error {
	return nil
}
func (m *mockAnalyticsRepo) DeleteProfileViewsOlderThan(context.Context, time.Time) error { return nil }
func (m *mockAnalyticsRepo) DeleteExpiredSessions(context.Context, time.Time) error       { return nil }

// =========================================================================
// JWT GENERATOR
// =========================================================================

func generateTestToken(userID string, role string, expiresAt time.Time) string {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     expiresAt.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testSecret))
	return signed
}

// =========================================================================
// DOMAIN FIXTURES
// =========================================================================

func testAdmin(id uuid.UUID, canCreateAdmin, sudo bool) *domain.UserAdmin {
	return &domain.UserAdmin{
		ID: id,
		Credentials: domain.Credentials{
			Username:           "admin_" + id.String()[:8],
			Email:              "admin_" + id.String()[:8] + "@test.com",
			Password:           "$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789012345678", // dummy hash
			Key:                testSecret,
			IsActive:           true,
			MustChangePassword: false,
		},
		Sudo:                 sudo,
		CanCreateAdmin:       canCreateAdmin,
		CanPublish:           true,
		CanCreatePsi:         true,
		CanUpdatePsi:         true,
		CanDeletePsi:         true,
		CanUpdateAdmin:       true,
		CanDeleteAdmin:       true,
		CanCreateTags:        true,
		CanEditTags:          true,
		CanDeleteTags:        true,
		CanReadNotifications: true,
	}
}

func testPsiUser(id uuid.UUID) *domain.PsiUserModel {
	return &domain.PsiUserModel{
		ID: id,
		Credentials: domain.Credentials{
			Username: "psi_" + id.String()[:8],
			Email:    "psi_" + id.String()[:8] + "@test.com",
			Key:      testSecret,
			IsActive: true,
		},
		FirstName:            "Juan",
		LastName:             "Perez",
		FPV:                  12345,
		CI:                   12345678,
		Nationality:          "V",
		Genre:                "M",
		ContactPhone:         "+58-241-1234567",
		ContactCellPhone:     "+58-412-1234567",
		ContactEmail:         "juan@test.com",
		MunicipalityCarabobo: "Valencia",
		PrimaryWorkArea:      "Psicologia Clinica",
	}
}

func testSpecialty(id uint32, name string) *domain.PsiSpecialtyModel {
	return &domain.PsiSpecialtyModel{
		ID:          id,
		Name:        name,
		Description: "Description for " + name,
		Active:      true,
	}
}

func testPost(id uuid.UUID, title string, status domain.PostStatus, postType string) *domain.Post {
	return &domain.Post{
		ID:         id,
		Title:      title,
		Type:       postType,
		Status:     status,
		TextID:     uuid.New(),
		ImageS3Key: "images/test.jpg",
	}
}

// =========================================================================
// FIBER APP BUILDERS
// =========================================================================

// newTestApp creates a minimal Fiber app for handler tests.
func newTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
}

// buildAuthMiddleware creates a real AuthMiddleware with mock repos for JWT validation.
func buildAuthMiddleware(adminRepo *mockAdminRepo, psiRepo *mockPsiRepo) *middleware.AuthMiddleware {
	analytics := service.NewAnalyticsService(&mockAnalyticsRepo{})
	return middleware.NewAuthMiddleware(adminRepo, psiRepo, analytics)
}

// setupAdminRoute builds a Fiber app with ProtectedAdmin404 middleware + a single route.
// The admin in c.Locals is injected by the real JWT middleware.
func setupAdminRoute(method, path string, handler fiber.Handler, adminRepo *mockAdminRepo, psiRepo *mockPsiRepo) *fiber.App {
	app := newTestApp()
	mw := buildAuthMiddleware(adminRepo, psiRepo)
	group := app.Group("/api/v1/admin", mw.ProtectedAdmin404())
	group.Add(method, path, handler)
	return app
}

// setupAdminRouteNoMW builds a Fiber app WITHOUT auth middleware for testing handler error paths.
func setupAdminRouteNoMW(method, path string, handler fiber.Handler) *fiber.App {
	app := newTestApp()
	app.Add(method, path, handler)
	return app
}

// setupPsiRoute builds a Fiber app with ProtectedPsiUser middleware + a single route.
func setupPsiRoute(method, path string, handler fiber.Handler, adminRepo *mockAdminRepo, psiRepo *mockPsiRepo) *fiber.App {
	app := newTestApp()
	mw := buildAuthMiddleware(adminRepo, psiRepo)
	group := app.Group("/api/v1/psi/me", mw.ProtectedPsiUser())
	group.Add(method, path, handler)
	return app
}

// setupHybridRoute builds a Fiber app with OptionalHybridAuth middleware + a single route.
func setupHybridRoute(method, path string, handler fiber.Handler, adminRepo *mockAdminRepo, psiRepo *mockPsiRepo) *fiber.App {
	app := newTestApp()
	mw := buildAuthMiddleware(adminRepo, psiRepo)
	app.Add(method, path, mw.OptionalHybridAuth(), handler)
	return app
}

// setupPublicRoute builds a Fiber app with NO auth middleware for public endpoints.
func setupPublicRoute(method, path string, handler fiber.Handler) *fiber.App {
	app := newTestApp()
	app.Add(method, path, handler)
	return app
}

// authRequest adds the Bearer token header to an http.Request.
func authRequest(req *http.Request, token string) *http.Request {
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

// decodeBody is a helper to read and parse a JSON response body.
func decodeBody(resp *http.Response) map[string]interface{} {
	var result map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result
}
