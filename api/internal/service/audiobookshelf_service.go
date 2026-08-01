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

	created, err := s.createUser(ctx, adminToken, username, password)
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

// createUser crea una cuenta web de usuario en ABS usando el token admin.
// Retorna true si la cuenta se creó (false si ya existía).
func (s *AudiobookshelfService) createUser(ctx context.Context, adminToken, username, password string) (bool, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"username": username,
		"password": password,
		"type":     "user",
		"isActive": true,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/api/users", bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrAbsUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict || strings.Contains(strings.ToLower(readBody(resp)), "already taken") {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("%w: POST /api/users %d", ErrAbsUnavailable, resp.StatusCode)
	}
	return true, nil
}

// readBody lee el cuerpo (hasta 4KB) para poder inspeccionar mensajes de error.
func readBody(resp *http.Response) string {
	buf := make([]byte, 4096)
	n, _ := resp.Body.Read(buf)
	return string(buf[:n])
}
