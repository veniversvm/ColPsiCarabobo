// api/internal/service/audiobookshelf_service_test.go
package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// fakeAbs emula la API de Audiobookshelf para las pruebas.
type fakeAbs struct {
	mu          sync.Mutex
	users       map[string]string      // username -> password
	adminUser   string                 // username del admin
	adminPass   string                 // password del admin
	loginCalls  int                    // contador de /login
	createCalls int                    // contador de POST /api/users
	usersIDs    map[string]string      // username -> id
}

func newFakeAbs() *fakeAbs {
	return &fakeAbs{
		users:    map[string]string{},
		usersIDs: map[string]string{},
	}
}

func (f *fakeAbs) handler() http.Handler {
	mux := http.NewServeMux()

	login := func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.loginCalls++
		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		json.NewDecoder(r.Body).Decode(&body)

		pass, ok := f.users[body.Username]
		if !ok || pass != body.Password {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"Invalid username or password"}`))
			return
		}
		id := f.usersIDs[body.Username]
		if id == "" {
			id = "user-" + body.Username
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user": map[string]interface{}{
				"id":          id,
				"username":    body.Username,
				"accessToken": "tok-" + body.Username,
			},
		})
	}
	mux.HandleFunc("/login", login)

	create := func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.createCalls++
		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		json.NewDecoder(r.Body).Decode(&body)

		if r.Header.Get("Authorization") != "Bearer tok-"+f.adminUser {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"Unauthorized"}`))
			return
		}
		if _, ok := f.users[body.Username]; ok {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(`{"error":"Username already taken"}`))
			return
		}
		f.users[body.Username] = body.Password
		f.usersIDs[body.Username] = "id-" + body.Username
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user": map[string]interface{}{
				"id":       "id-" + body.Username,
				"username": body.Username,
			},
		})
	}
	mux.HandleFunc("/api/users", create)

	return mux
}

// buildSvc levanta el fake y un servicio apuntando a él.
func buildSvc(t *testing.T, fake *fakeAbs) (*AudiobookshelfService, string) {
	t.Helper()
	if fake.adminUser != "" {
		fake.users[fake.adminUser] = fake.adminPass
	}
	srv := httptest.NewServer(fake.handler())
	t.Cleanup(srv.Close)

	svc := NewAudiobookshelfService(
		srv.URL,              // baseURL (interna)
		"https://abs.public", // publicURL (navegador)
		fake.adminUser,
		fake.adminPass,
		"secret-test",
	)
	return svc, srv.URL
}

func TestGetAccess_ExistingUser(t *testing.T) {
	fake := newFakeAbs()
	fake.adminUser = "colpsi-bot"
	fake.adminPass = "adminpass"

	svc, _ := buildSvc(t, fake)
	// El usuario ya existe en ABS con la clave derivada del secreto.
	fake.users["psi_12345678"] = svc.passwordFor("psi_12345678")
	fake.usersIDs["psi_12345678"] = "id-psi_12345678"

	access, err := svc.GetAccess(context.Background(), "psi_12345678")
	require.NoError(t, err)
	require.False(t, access.Created)
	require.Equal(t, "psi_12345678", access.Username)
	require.Equal(t, "id-psi_12345678", access.UserID)
	require.Equal(t, "https://abs.public/login/?accessToken=tok-psi_12345678", access.URL)
	require.Equal(t, 1, fake.loginCalls)
	require.Equal(t, 0, fake.createCalls)
}

func TestGetAccess_CreatesMissingUser(t *testing.T) {
	fake := newFakeAbs()
	fake.adminUser = "colpsi-bot"
	fake.adminPass = "adminpass"
	// No existe psi_12345678 → debe crearse

	svc, _ := buildSvc(t, fake)
	access, err := svc.GetAccess(context.Background(), "psi_12345678")
	require.NoError(t, err)
	require.True(t, access.Created)
	require.Equal(t, "id-psi_12345678", access.UserID)
	require.Equal(t, "https://abs.public/login/?accessToken=tok-psi_12345678", access.URL)
	require.Equal(t, 3, fake.loginCalls) // psi fallido + admin + psi exitoso
	require.Equal(t, 1, fake.createCalls)
}

func TestGetAccess_ConcurrentSameUser(t *testing.T) {
	fake := newFakeAbs()
	fake.adminUser = "colpsi-bot"
	fake.adminPass = "adminpass"

	svc, _ := buildSvc(t, fake)

	// Dos llamadas simultáneas para el mismo usuario: una crea, la otra
	// recibe "Username already taken" y re-autentica. Ambas deben tener éxito.
	ctx := context.Background()
	var wg sync.WaitGroup
	results := make([]error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, results[idx] = svc.GetAccess(ctx, "psi_1001")
		}(i)
	}
	wg.Wait()
	require.NoError(t, results[0])
	require.NoError(t, results[1])
	require.Equal(t, 2, fake.createCalls) // ambos intentaron crear; solo uno efectivo
}

func TestGetAccess_AdminUnconfigured(t *testing.T) {
	fake := newFakeAbs()
	// admin sin credenciales
	svc, _ := buildSvc(t, fake)
	_, err := svc.GetAccess(context.Background(), "psi_12345678")
	require.Error(t, err)
	require.Contains(t, err.Error(), "falta configurar")
}

func TestGetAccess_ABSUnavailable(t *testing.T) {
	// Servidor que no responde (puerto cerrado)
	svc := NewAudiobookshelfService(
		"http://127.0.0.1:1",
		"https://abs.public",
		"colpsi-bot",
		"adminpass",
		"secret-test",
	)
	_, err := svc.GetAccess(context.Background(), "psi_12345678")
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "no disponible"))
}
