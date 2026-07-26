package integration

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// =========================================================================
// 7.1 — JWT TOKEN ATTACKS
// =========================================================================

func TestSecurity_JWT_NoneAlgorithm(t *testing.T) {
	truncateAll(t)
	admin := seedSudo(t)
	app := buildTestApp(testDB)

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	claims := fmt.Sprintf(`{"user_id":"%s","role":"admin","exp":%d,"iat":%d}`,
		admin.ID.String(), time.Now().Add(24*time.Hour).Unix(), time.Now().Unix())
	payload := base64.RawURLEncoding.EncodeToString([]byte(claims))
	forgedToken := header + "." + payload + "."

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	req.Header.Set("Authorization", authHeader(forgedToken))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSecurity_JWT_WrongKey(t *testing.T) {
	truncateAll(t)
	admin := seedSudo(t)
	app := buildTestApp(testDB)

	wrongKeyToken := generateToken(t, admin.ID, "admin", "completely-different-key")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	req.Header.Set("Authorization", authHeader(wrongKeyToken))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSecurity_JWT_ExpiredToken(t *testing.T) {
	truncateAll(t)
	admin := seedSudo(t)
	app := buildTestApp(testDB)

	claims := jwt.MapClaims{
		"user_id": admin.ID.String(),
		"role":    "admin",
		"exp":     time.Now().Add(-1 * time.Hour).Unix(),
		"iat":     time.Now().Add(-2 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(admin.Key))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	req.Header.Set("Authorization", authHeader(signed))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSecurity_JWT_MalformedHeader(t *testing.T) {
	truncateAll(t)
	app := buildTestApp(testDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiJ9.broken.signature")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// =========================================================================
// 7.2 — KEY ROTATION & SESSION INVALIDATION
// =========================================================================

func TestSecurity_KeyRotation_LogoutInvalidates(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/logout", nil)
	logoutReq.Header.Set("Authorization", authHeader(token))
	logoutResp, err := app.Test(logoutReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, logoutResp.StatusCode)

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	req2.Header.Set("Authorization", authHeader(token))
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

func TestSecurity_KeyRotation_DoubleLogin(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token1 := loginAdmin(t, app, "sudo", "Sudo123!@#")

	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	req1.Header.Set("Authorization", authHeader(token1))
	resp1, err := app.Test(req1)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp1.StatusCode)

	token2 := loginAdmin(t, app, "sudo", "Sudo123!@#")

	req2old := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	req2old.Header.Set("Authorization", authHeader(token1))
	resp2old, err := app.Test(req2old)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp2old.StatusCode)

	req2new := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	req2new.Header.Set("Authorization", authHeader(token2))
	resp2new, err := app.Test(req2new)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp2new.StatusCode)
}

func TestSecurity_KeyRotation_PasswordChange(t *testing.T) {
	truncateAll(t)
	admin := seedAdmin(t, "pwdchange@test.com", fullAdminPerms())
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "pwdchange@test.com", "Admin123!@#")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	newPass := "NewPass9!@#"
	updateBody := fmt.Sprintf(`{"id":"%s","password":"%s"}`, admin.ID.String(), newPass)
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/update", strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", authHeader(token))
	updateResp, err := app.Test(updateReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, updateResp.StatusCode)

	reqOld := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	reqOld.Header.Set("Authorization", authHeader(token))
	respOld, err := app.Test(reqOld)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, respOld.StatusCode)
}

// =========================================================================
// 7.3 — RBAC & PRIVILEGE ESCALATION
// =========================================================================

func TestSecurity_RBAC_CannotGrantUnpossessedPerms(t *testing.T) {
	truncateAll(t)
	seedAdmin(t, "limited@test.com", map[string]bool{
		"can_create_admin": true,
	})
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "limited@test.com", "Admin123!@#")

	body := `{"username":"victim@test.com","email":"victim@test.com","password":"Strong123!@#","permissions":{"can_delete_admin":true}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	data := decodeBody(t, resp)
	require.Contains(t, data["error"], "no puedes otorgar el permiso")
}

func TestSecurity_RBAC_SudoCannotBeEdited(t *testing.T) {
	truncateAll(t)
	sudo := seedSudo(t)
	seedAdmin(t, "editor@test.com", map[string]bool{
		"can_update_admin": true,
	})
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "editor@test.com", "Admin123!@#")

	body := fmt.Sprintf(`{"id":"%s","email":"hacked@test.com"}`, sudo.ID.String())
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/update", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	data := decodeBody(t, resp)
	require.Contains(t, data["error"], "Super Usuario")
}

func TestSecurity_RBAC_SudoCannotBeDeletedViaAPI(t *testing.T) {
	truncateAll(t)
	sudo := seedSudo(t)
	seedAdmin(t, "deleter@test.com", map[string]bool{
		"can_delete_admin": true,
	})
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "deleter@test.com", "Admin123!@#")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/delete/"+sudo.ID.String(), nil)
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	data := decodeBody(t, resp)
	require.Contains(t, data["error"], "Super Usuario")
}

func TestSecurity_RBAC_NonAdminCannotCreateAdmin(t *testing.T) {
	truncateAll(t)
	seedAdmin(t, "noperm@test.com", map[string]bool{
		"can_publish": true,
	})
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "noperm@test.com", "Admin123!@#")

	body := `{"username":"victim@test.com","email":"victim@test.com","password":"Strong123!@#"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	data := decodeBody(t, resp)
	require.Contains(t, data["error"], "permisos insuficientes")
}

// =========================================================================
// 7.4 — PASSWORD POLICY
// =========================================================================

func TestSecurity_Password_WeakRejected(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	body := `{"username":"weak@test.com","email":"weak@test.com","password":"abcdefg1!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	data := decodeBody(t, resp)
	require.Contains(t, data["error"], "contraseña")
}

func TestSecurity_Password_NoSpecialChar(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	body := `{"username":"nosp@test.com","email":"nosp@test.com","password":"Abcdefg1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	data := decodeBody(t, resp)
	require.Contains(t, data["error"], "contraseña")
}

func TestSecurity_Password_ContainsSpace(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	body := `{"username":"space@test.com","email":"space@test.com","password":"Abcdef 1!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	data := decodeBody(t, resp)
	require.Contains(t, data["error"], "contraseña")
}

// =========================================================================
// 7.5 — INFORMATION LEAKAGE PREVENTION
// =========================================================================

func TestSecurity_CredentialsNeverInResponse(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	body := `{"identifier":"sudo","password":"Sudo123!@#"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	raw, err := json.Marshal(decodeBody(t, resp))
	require.NoError(t, err)
	responseStr := string(raw)
	require.NotContains(t, responseStr, `"password"`)
	require.NotContains(t, responseStr, `"key"`)
}

func TestSecurity_Admin404_Obscurity(t *testing.T) {
	truncateAll(t)
	app := buildTestApp(testDB)

	forgedToken := generateToken(t, uuid.New(), "admin", "some-key")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	req.Header.Set("Authorization", authHeader(forgedToken))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// =========================================================================
// 7.6 — EMAIL VALIDATION
// =========================================================================

func TestSecurity_Email_InvalidFormat(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	body := `{"username":"not-an-email","email":"not-an-email","password":"Strong123!@#"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	data := decodeBody(t, resp)
	require.Contains(t, data["error"], "email")
}

func TestSecurity_Email_EmptyString(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	body := `{"username":"","email":"","password":"Strong123!@#"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// =========================================================================
// 7.7 — INPUT INJECTION & SANITIZATION
// =========================================================================

func TestSecurity_SQLInjection_Login(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	body := `{"identifier":"admin' OR '1'='1","password":"anything"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSecurity_SQLInjection_Search(t *testing.T) {
	truncateAll(t)
	app := buildTestApp(testDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/psi/directory", nil)
	q := req.URL.Query()
	q.Set("search", "'; DROP TABLE psi_users; --")
	req.URL.RawQuery = q.Encode()
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	data := decodeBody(t, resp)
	if total, ok := data["total"].(float64); ok {
		require.Equal(t, float64(0), total)
	}

	var tableExists int
	testDB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'psi_users'").Scan(&tableExists)
	require.Equal(t, 1, tableExists, "psi_users table should still exist after SQL injection attempt")
}

func TestSecurity_XSS_InSpecialtyName(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	body := `{"name":"<script>alert(1)</script>","description":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/specialties/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/specialties/", nil)
	listResp, err := app.Test(listReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, listResp.StatusCode)
}

// =========================================================================
// 7.8 — HTTP SECURITY HEADERS
// =========================================================================

func TestSecurity_Headers_Present(t *testing.T) {
	truncateAll(t)
	app := buildTestApp(testDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/specialties/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	require.NotEmpty(t, resp.Header.Get("X-Content-Type-Options"))
}

// =========================================================================
// 7.9 — METHOD & CONTENT-TYPE ABUSE
// =========================================================================

func TestSecurity_WrongContentType(t *testing.T) {
	truncateAll(t)
	app := buildTestApp(testDB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader("identifier=admin&password=123"))
	req.Header.Set("Content-Type", "text/plain")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.NotEqual(t, http.StatusOK, resp.StatusCode)
}

func TestSecurity_WrongHTTPMethod(t *testing.T) {
	truncateAll(t)
	app := buildTestApp(testDB)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/auth/login", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

// =========================================================================
// 7.10 — EMPTY & MALFORMED BODY
// =========================================================================

func TestSecurity_EmptyBody_Login(t *testing.T) {
	truncateAll(t)
	app := buildTestApp(testDB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSecurity_NullFields_Login(t *testing.T) {
	truncateAll(t)
	app := buildTestApp(testDB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"identifier":null,"password":null}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// =========================================================================
// 7.11 — IDEMPOTENCY
// =========================================================================

func TestSecurity_Idempotency_Replay(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	body := `{"username":"idemp@test.com","email":"idemp@test.com","password":"Strong123!@#","first_name":"A","last_name":"B","ci":12345678,"fpv":12345,"nationality":"V","born_date":"1990-01-01","genre":"M"}`

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/admin/psi/create", strings.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", authHeader(token))
	req1.Header.Set("X-Idempotency-Key", "test-idemp-key-123")
	resp1, err := app.Test(req1)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp1.StatusCode)

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/admin/psi/create", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", authHeader(token))
	req2.Header.Set("X-Idempotency-Key", "test-idemp-key-123")
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp2.StatusCode)

	data1 := decodeBody(t, resp1)
	data2 := decodeBody(t, resp2)
	require.Equal(t, data1["message"], data2["message"])
}

func TestSecurity_Idempotency_DifferentKeys(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	body1 := `{"title":"Post Alpha","content":"<p>Content A</p>","type":"public","status":"published"}`
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/admin/posts/", strings.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", authHeader(token))
	req1.Header.Set("X-Idempotency-Key", "post-key-alpha")
	resp1, err := app.Test(req1)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp1.StatusCode)

	body2 := `{"title":"Post Beta","content":"<p>Content B</p>","type":"public","status":"published"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/admin/posts/", strings.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", authHeader(token))
	req2.Header.Set("X-Idempotency-Key", "post-key-beta")
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp2.StatusCode)

	var count int64
	testDB.Raw("SELECT COUNT(*) FROM posts WHERE deleted_at IS NULL").Scan(&count)
	require.Equal(t, int64(2), count)
}

// =========================================================================
// 7.12 — CSRF & AUTHENTICATION HEADERS
// =========================================================================

func TestSecurity_CSRF_NoAuthHeader(t *testing.T) {
	truncateAll(t)
	app := buildTestApp(testDB)

	body := `{"username":"test@test.com","email":"test@test.com","password":"Strong123!@#"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSecurity_Auth_MissingBearerPrefix(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	req.Header.Set("Authorization", token)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSecurity_Auth_EmptyBearer(t *testing.T) {
	truncateAll(t)
	app := buildTestApp(testDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	req.Header.Set("Authorization", "Bearer ")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// =========================================================================
// 7.13 — TOKEN IN WRONG LOCATION
// =========================================================================

func TestSecurity_Token_InQueryParams(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list?token="+token, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSecurity_Token_InWrongHeader(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	req.Header.Set("X-Auth-Token", token)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// =========================================================================
// 7.14 — PRIVILEGE ESCALATION ADVANCED
// =========================================================================

func TestSecurity_RBAC_AdminCannotSelfElevate(t *testing.T) {
	truncateAll(t)
	admin := seedAdmin(t, "nopriv@test.com", map[string]bool{})
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "nopriv@test.com", "Admin123!@#")

	body := fmt.Sprintf(`{"id":"%s","permissions":{"can_create_admin":true}}`, admin.ID.String())
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/update", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	data := decodeBody(t, resp)
	require.Contains(t, data["error"], "Crear Admin")
}

func TestSecurity_RBAC_NonSudoCannotEditSudo(t *testing.T) {
	truncateAll(t)
	sudo := seedSudo(t)
	seedAdmin(t, "attempter@test.com", map[string]bool{
		"can_update_admin": true,
	})
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "attempter@test.com", "Admin123!@#")

	body := fmt.Sprintf(`{"id":"%s","email":"hacked@test.com"}`, sudo.ID.String())
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/update", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	data := decodeBody(t, resp)
	require.Contains(t, data["error"], "Super Usuario")
}

func TestSecurity_RBAC_CannotCreateSudo(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	body := `{"username":"newsudo@test.com","email":"newsudo@test.com","password":"Strong123!@#","permissions":{"sudo":true}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created domain.UserAdmin
	testDB.Where("username = ?", "newsudo@test.com").First(&created)
	require.False(t, created.Sudo, "new admin should not be sudo")
}

// =========================================================================
// 7.15 — SESSION SECURITY
// =========================================================================

func TestSecurity_Session_DualLoginIndependent(t *testing.T) {
	truncateAll(t)
	seedAdmin(t, "admin_a@test.com", fullAdminPerms())
	seedAdmin(t, "admin_b@test.com", fullAdminPerms())
	app := buildTestApp(testDB)

	tokenA := loginAdmin(t, app, "admin_a@test.com", "Admin123!@#")
	tokenB := loginAdmin(t, app, "admin_b@test.com", "Admin123!@#")

	reqA := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	reqA.Header.Set("Authorization", authHeader(tokenA))
	respA, err := app.Test(reqA)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, respA.StatusCode)

	reqB := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	reqB.Header.Set("Authorization", authHeader(tokenB))
	respB, err := app.Test(reqB)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, respB.StatusCode)
}

func TestSecurity_Session_PasswordChangeKillsToken(t *testing.T) {
	truncateAll(t)
	admin := seedAdmin(t, "pwdkill@test.com", fullAdminPerms())
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "pwdkill@test.com", "Admin123!@#")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	newPass := "KilledPass9!@#"
	body := fmt.Sprintf(`{"id":"%s","password":"%s"}`, admin.ID.String(), newPass)
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/update", strings.NewReader(body))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", authHeader(token))
	updateResp, err := app.Test(updateReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, updateResp.StatusCode)

	reqOld := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	reqOld.Header.Set("Authorization", authHeader(token))
	respOld, err := app.Test(reqOld)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, respOld.StatusCode)
}

func TestSecurity_Session_LogoutClearsKey(t *testing.T) {
	truncateAll(t)
	seedAdmin(t, "logout@test.com", fullAdminPerms())
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "logout@test.com", "Admin123!@#")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/logout", nil)
	logoutReq.Header.Set("Authorization", authHeader(token))
	logoutResp, err := app.Test(logoutReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, logoutResp.StatusCode)

	reqAfter := httptest.NewRequest(http.MethodGet, "/api/v1/admin/list", nil)
	reqAfter.Header.Set("Authorization", authHeader(token))
	respAfter, err := app.Test(reqAfter)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, respAfter.StatusCode)

	var dbAdmin domain.UserAdmin
	testDB.Where("username = ?", "logout@test.com").First(&dbAdmin)
	require.Empty(t, dbAdmin.Key, "Key should be empty after logout")
}

// =========================================================================
// 7.16 — CONTENT & ENCODING ATTACKS
// =========================================================================

func TestSecurity_Unicode_InCredentials(t *testing.T) {
	truncateAll(t)
	app := buildTestApp(testDB)

	body := `{"identifier":"administradorñ","password":"Ñoño123!@"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.NotEqual(t, http.StatusInternalServerError, resp.StatusCode)
	require.NotEqual(t, http.StatusBadGateway, resp.StatusCode)
}

func TestSecurity_Emoji_InSearch(t *testing.T) {
	truncateAll(t)
	app := buildTestApp(testDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/psi/directory?search=%F0%9F%94%AC%F0%9F%A7%AA", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSecurity_VeryLongInput(t *testing.T) {
	truncateAll(t)
	app := buildTestApp(testDB)

	longString := strings.Repeat("A", 10000)
	body := fmt.Sprintf(`{"identifier":"%s","password":"test"}`, longString)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.NotEqual(t, http.StatusInternalServerError, resp.StatusCode)
}

// =========================================================================
// 7.17 — DATA ISOLATION & AUDIT TRAIL
// =========================================================================

func TestSecurity_DataIsolation_AuditTrailOnCreate(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	body := `{"username":"audited@test.com","email":"audited@test.com","password":"Strong123!@#"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created domain.UserAdmin
	testDB.Where("username = ?", "audited@test.com").First(&created)
	require.Equal(t, "sudo", created.CreateBy, "CreateBy should be set to creator username")
	require.NotNil(t, created.CreateById, "CreateById should be set")
	require.Equal(t, "sudo", created.UpdateBy, "UpdateBy should be set to creator username")
}

// =========================================================================
// IMPORTS USED BY REFERENCE
// =========================================================================

var _ = domain.UserAdmin{}
