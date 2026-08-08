// api/internal/service/audiobookshelf_service.go

// Integración con Audiobookshelf: provee el acceso automático de los
// agremiados solventes a la biblioteca digital. Si la cuenta del psicólogo
// aún no existe en ABS, se crea al vuelo con la API de administración y se
// devuelve una URL de auto-login (login?accessToken=...) lista para el
// navegador.
package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// AudiobookshelfAccess es la respuesta de GetAccess: la URL de auto-login y
// datos del usuario ABS involucrado.
type AudiobookshelfAccess struct {
	URL      string // URL pública de auto-login para abrir en el navegador
	Username string // Usuario en Audiobookshelf (psi_<ci>)
	UserID   string // Id interno de la cuenta en Audiobookshelf
	Created  bool   // true si la cuenta se creó en esta llamada
}

// Errores tipificados para que el handler decida el status HTTP.
var (
	// ErrAbsUnauthorized: credenciales del admin de ABS inválidas (no debe
	// filtrarse al cliente, solo registrarse en logs).
	ErrAbsUnauthorized = errors.New("credenciales inválidas en Audiobookshelf")
	// ErrAbsUnavailable: ABS no responde o falló un request de la API.
	ErrAbsUnavailable = errors.New("Audiobookshelf no disponible")
)

// AudiobookshelfService orquesta la API de Audiobookshelf.
type AudiobookshelfService struct {
	baseURL        string // URL interna que usa la API (SDK)
	publicURL      string // URL pública que abre el navegador
	adminUsername  string
	adminPassword  string
	passwordSecret string // Secreto para derivar la contraseña de cada agremiado
	client         *http.Client
}

// NewAudiobookshelfService crea el servicio con la config de ABS. Si alguna
// de las credenciales de admin está vacía, las operaciones que requieran
// aprovisionar fallarán (la autenticación del agremiado aún puede funcionar).
func NewAudiobookshelfService(baseURL, publicURL, adminUsername, adminPassword, passwordSecret string) *AudiobookshelfService {
	return &AudiobookshelfService{
		baseURL:        strings.TrimSuffix(baseURL, "/"),
		publicURL:      strings.TrimSuffix(publicURL, "/"),
		adminUsername:  adminUsername,
		adminPassword:  adminPassword,
		passwordSecret: passwordSecret,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// absUser es el subconjunto del usuario ABS que nos interesa.
type absUser struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	AccessToken string `json:"accessToken"`
}

// AbsUser es la vista resumida de un usuario ABS para la sincronización masiva.
type AbsUser struct {
	ID       string
	Username string
	IsActive bool
}

type absLoginResponse struct {
	User absUser `json:"user"`
}

// GetAccess garantiza la cuenta del agremiado en ABS y devuelve la URL de
// auto-login. username debe ser el nombre único del usuario (psi_<ci>).
func (s *AudiobookshelfService) GetAccess(ctx context.Context, username string) (*AudiobookshelfAccess, error) {
	password := s.passwordFor(username)

	// 1) Intento directo: la cuenta ya existe y la clave derivada es válida.
	user, err := s.login(ctx, username, password)
	if err == nil {
		return s.buildAccess(user, false)
	}

	// 2) La cuenta no existe (o la clave no coincide): se aprovisiona.
	adminToken, err := s.adminLogin(ctx)
	if err != nil {
		return nil, err
	}

	_, created, err := s.createUser(ctx, adminToken, username, password)
	if err != nil {
		// Carrera entre dos peticiones simultáneas: otra ya la creó.
		if !strings.Contains(strings.ToLower(err.Error()), "already taken") {
			return nil, err
		}
	}

	user, err = s.login(ctx, username, password)
	if err != nil {
		return nil, fmt.Errorf("%w: no se pudo autenticar la cuenta recién creada", ErrAbsUnavailable)
	}
	return s.buildAccess(user, created)
}

// buildAccess arma la respuesta con la URL de auto-login de ABS.
func (s *AudiobookshelfService) buildAccess(user *absUser, created bool) (*AudiobookshelfAccess, error) {
	if user == nil || user.AccessToken == "" {
		return nil, fmt.Errorf("%w: respuesta sin accessToken", ErrAbsUnavailable)
	}
	return &AudiobookshelfAccess{
		URL:      fmt.Sprintf("%s/login/?accessToken=%s", s.publicURL, user.AccessToken),
		Username: user.Username,
		UserID:   user.ID,
		Created:  created,
	}, nil
}

// passwordFor deriva (determinista) la contraseña de cada agremiado a partir
// del secreto global. Nunca se devuelve al cliente; solo la conoce ABS.
func (s *AudiobookshelfService) passwordFor(username string) string {
	if s.passwordSecret == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(s.passwordSecret))
	mac.Write([]byte("abs:" + username))
	return hex.EncodeToString(mac.Sum(nil))[:32]
}

// login autentica contra ABS y devuelve el usuario con su accessToken.
func (s *AudiobookshelfService) login(ctx context.Context, username, password string) (*absUser, error) {
	body, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/login", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAbsUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrAbsUnauthorized
	}

	var parsed absLoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("%w: cuerpo inválido", ErrAbsUnavailable)
	}
	return &parsed.User, nil
}

// adminLogin obtiene el token del admin para operaciones de aprovisionamiento.
func (s *AudiobookshelfService) adminLogin(ctx context.Context) (string, error) {
	if s.adminUsername == "" || s.adminPassword == "" {
		return "", fmt.Errorf("%w: falta configurar el admin de ABS", ErrAbsUnauthorized)
	}
	user, err := s.login(ctx, s.adminUsername, s.adminPassword)
	if err != nil {
		return "", err
	}
	return user.AccessToken, nil
}

// ListUsers lista los usuarios de ABS usando el token del admin. Se usa para
// comparar contra la base de datos en la sincronización masiva de cuentas.
func (s *AudiobookshelfService) ListUsers(ctx context.Context) ([]AbsUser, error) {
	_, users, err := s.ListUsersWithToken(ctx)
	return users, err
}

// ListUsersWithToken lista los usuarios de ABS y devuelve el token admin usado
// para autenticarse. Reutilizar ese token en el resto de la pasada evita hacer
// un login por operación: el endpoint /login de ABS está limitado por IP y un
// bucle de aprovisionamiento masivo lo agota (deja cuentas sin crear).
func (s *AudiobookshelfService) ListUsersWithToken(ctx context.Context) (string, []AbsUser, error) {
	adminToken, err := s.adminLogin(ctx)
	if err != nil {
		return "", nil, err
	}
	users, err := s.listUsers(ctx, adminToken)
	if err != nil {
		return "", nil, err
	}
	return adminToken, users, nil
}

// listUsers ejecuta GET /api/users con un token admin ya obtenido.
func (s *AudiobookshelfService) listUsers(ctx context.Context, adminToken string) ([]AbsUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/api/users", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAbsUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: GET /api/users %d", ErrAbsUnavailable, resp.StatusCode)
	}

	var parsed struct {
		Users []struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			IsActive bool   `json:"isActive"`
		} `json:"users"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("%w: GET /api/users cuerpo inválido", ErrAbsUnavailable)
	}

	users := make([]AbsUser, 0, len(parsed.Users))
	for _, u := range parsed.Users {
		users = append(users, AbsUser{ID: u.ID, Username: u.Username, IsActive: u.IsActive})
	}
	return users, nil
}

// DeactivateUser desactiva la cuenta ABS de un agremiado que dejó de ser
// solvente. PATCH /api/users/{id} con isActive=false.
func (s *AudiobookshelfService) DeactivateUser(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("%w: id de usuario vacío", ErrAbsUnavailable)
	}
	adminToken, err := s.adminLogin(ctx)
	if err != nil {
		return err
	}

	body, _ := json.Marshal(map[string]bool{"isActive": false})
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, s.baseURL+"/api/users/"+userID, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAbsUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: PATCH /api/users/%s %d", ErrAbsUnavailable, userID, resp.StatusCode)
	}
	return nil
}

// createUser crea una cuenta web de usuario en ABS usando el token admin.
// Retorna el id de la cuenta creada y true si se creó (false si ya existía).
func (s *AudiobookshelfService) createUser(ctx context.Context, adminToken, username, password string) (string, bool, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"username": username,
		"password": password,
		"type":     "user",
		"isActive": true,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/api/users", bytes.NewReader(body))
	if err != nil {
		return "", false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", false, fmt.Errorf("%w: %v", ErrAbsUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return "", false, nil
	}
	if resp.StatusCode != http.StatusOK {
		if strings.Contains(strings.ToLower(readBody(resp)), "already taken") {
			return "", false, nil
		}
		return "", false, fmt.Errorf("%w: POST /api/users %d", ErrAbsUnavailable, resp.StatusCode)
	}

	// ABS responde { "user": { "id": "...", "username": "..." } }
	var parsed struct {
		User absUser `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", false, fmt.Errorf("%w: POST /api/users sin id", ErrAbsUnavailable)
	}
	return parsed.User.ID, true, nil
}

// createUserIfMissing crea la cuenta ABS sin hacer un login previo de sondeo,
// usando un token admin ya obtenido. Pensada para la sincronización masiva:
// el listado previo ya indica qué cuentas faltan y un probe por usuario
// duplicaría los logins (agotando el rate limiter de /login de ABS). Si la
// cuenta ya existe (carrera con otra pasada o un GetAccess concurrente), se
// autentica para recuperar el id real.
func (s *AudiobookshelfService) createUserIfMissing(ctx context.Context, adminToken, username string) (string, bool, error) {
	password := s.passwordFor(username)

	userID, created, err := s.createUser(ctx, adminToken, username, password)
	if err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "already taken") {
			return "", false, err
		}
	} else if created {
		return userID, true, nil
	}

	// Ya existía (conflicto "already taken"): se autentica para obtener su id.
	user, loginErr := s.login(ctx, username, password)
	if loginErr != nil {
		return "", false, loginErr
	}
	return user.ID, false, nil
}

// readBody lee el cuerpo (hasta 4KB) para poder inspeccionar mensajes de error.
func readBody(resp *http.Response) string {
	buf := make([]byte, 4096)
	n, _ := resp.Body.Read(buf)
	return string(buf[:n])
}
