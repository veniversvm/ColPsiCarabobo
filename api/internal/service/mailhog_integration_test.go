package service

import (
	"context"
	"encoding/json"
	"io"
	"mime/quotedprintable"
	"net"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// =========================================================================
// TEST END-TO-END CON MAILHOG/MAILPIT (solo dev)
// =========================================================================
// Verifica que cada psicólogo importado reciba SU PROPIA contraseña a través
// del pipeline real de correo (SMTP local → MailHog), y que esa contraseña
// valide contra el hash persistido.
//
// Se omite (t.Skip) si MailHog no está corriendo, para no romper CI/dev.

const mailhogAPIURL = "http://localhost:28025/api/v2/messages"

func mailhogAvailable() bool {
	conn, err := net.DialTimeout("tcp", "localhost:21025", 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return clearMailhog()
}

// clearMailhog vacía la bandeja de MailHog antes de la corrida, para que el
// poll del test solo vea correos frescos de esta ejecución.
func clearMailhog() bool {
	req, _ := http.NewRequest(http.MethodDelete, "http://localhost:28025/api/v1/messages", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

type mailhogMessage struct {
	Content struct {
		Headers map[string][]string `json:"Headers"`
		Body    string              `json:"Body"`
	} `json:"Content"`
}

func TestImportFromCSV_MailHog_PasswordUnicaReal(t *testing.T) {
	if !mailhogAvailable() {
		t.Skip("MailHog no está disponible (dev compose profile 'dev'); se omite")
	}

	config.InitConfig()
	if config.Envs == nil {
		t.Fatal("config.Envs no inicializado")
	}
	orig := *config.Envs
	defer func() { *config.Envs = orig }()

	// Apuntar SMTP al MailHog local (sin auth, TLS oportunista en dev).
	config.Envs.Environment = "development"
	config.Envs.SMTPHost = "localhost"
	config.Envs.SMTPPort = 21025
	config.Envs.SMTPUser = ""
	config.Envs.SMTPPass = ""
	config.Envs.SMTPFrom = "info@colpsicarabobo.com"
	config.Envs.SMTPFromName = "Colegio de Psicólogos de Carabobo"
	config.Envs.SMTPReplyTo = ""
	config.Envs.MailSignature = "Administración ColPsiCarabobo"

	ms, err := NewMailService()
	if err != nil {
		t.Fatalf("NewMailService() falló: %v", err)
	}
	defer ms.Close()

	// Fuerza el camino de contraseña aleatoria por usuario (no-development).
	config.Envs.Environment = "production"

	rows := [][]string{emptyRow(), emptyRow()}
	addUser := func(fpv, ci, first, last, email string) {
		row := emptyRow()
		row[0] = "9"
		row[3] = fpv
		row[4] = "33154"
		row[5] = "V"
		row[6] = ci
		row[7] = first
		row[8] = "María"
		row[9] = last
		row[10] = "Gómez"
		row[11] = "33000"
		row[13] = "F"
		row[14] = "SI"
		row[15] = email
		row[16] = "valencia"
		row[25] = "UC"
		row[26] = "33500"
		row[30] = "123"
		row[31] = "5"
		row[32] = "2"
		row[44] = "46230"
		rows = append(rows, row)
	}

	emails := []string{"e2e.ana1@test.com", "e2e.luis2@test.com", "e2e.carla3@test.com"}
	addUser("4411", "44111111", "Ana", "Perez", emails[0])
	addUser("4422", "44222222", "Luis", "Gomez", emails[1])
	addUser("4433", "44333333", "Carla", "Ruiz", emails[2])

	hashes := make(map[string]string)
	repo := &mockPsiRepoSvc{
		CreateWithColDataFunc: func(ctx context.Context, psi *domain.PsiUserModel, col *domain.PsiUserColData, sol []domain.PsiUserSolvency, pg []domain.PsiUserPostGrade) error {
			hashes[psi.Email] = psi.Password
			return nil
		},
	}

	svc := NewPsiService(repo, nil, ms)
	buf := createTestXLSX(t, rows)
	success, failed := svc.ImportFromCSV(context.Background(), buf, uuid.Must(uuid.NewV7()))

	if success != 3 {
		t.Fatalf("debe importar 3 filas, obtuve %d (failed: %v)", success, failed)
	}
	if len(hashes) != 3 {
		t.Fatalf("debe guardar 3 hashes, obtuve %d", len(hashes))
	}

	// Independencia: cada hash distinto (no-dev ⇒ contraseñas aleatorias).
	seen := make(map[string]bool)
	for email, h := range hashes {
		if seen[h] {
			t.Fatalf("usuario %s comparte hash bcrypt con otro usuario del batch", email)
		}
		seen[h] = true
	}

	// Esperar a que el worker asíncrono entregue los 3 correos a MailHog.
	delivered := fetchMailhogPasswords(t, emails)

	// Cada correo debe traer la contraseña exacta que valida contra su hash.
	sentPasswords := make(map[string]string)
	for email, pw := range delivered {
		h, ok := hashes[email]
		if !ok {
			t.Errorf("mailhog recibió correo para %s que no fue importado", email)
			continue
		}
		if err := bcrypt.CompareHashAndPassword([]byte(h), []byte(pw)); err != nil {
			t.Errorf("correo a %s llevó la contraseña %q que NO valida contra su hash", email, pw)
		}
		sentPasswords[email] = pw
	}

	// Ningún psicólogo debe compartir contraseña con otro.
	if len(sentPasswords) != 3 && len(delivered) == 3 {
		t.Fatalf("no se extrajeron contraseñas de los 3 correos: %v", delivered)
	}
	for email, pw := range sentPasswords {
		for other, otherPw := range sentPasswords {
			if email != other && pw == otherPw {
				t.Fatalf("los usuarios %s y %s recibieron la misma contraseña %q", email, other, pw)
			}
		}
	}
}

var passwordRe = regexp.MustCompile(`background:#facc15; padding:1px 6px; border-radius:4px;">([^<]+)</strong>`)

// resetURLRe extrae el token de recuperación del enlace de la plantilla
// reset_password_token (botón "Restablecer contraseña").
var resetURLRe = regexp.MustCompile(`/reset-password\?token=([^"<&\s]+)`)

// fetchMailhogResetToken espera en MailHog el correo de recuperación dirigido
// a `wantTo` y devuelve el token en el enlace. Falla con t.Fatalf si no llega.
func fetchMailhogResetToken(t *testing.T, wantTo string) string {
	t.Helper()

	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(mailhogAPIURL)
		if err != nil {
			time.Sleep(250 * time.Millisecond)
			continue
		}
		var envelope struct {
			Items []mailhogMessage `json:"items"`
		}
		decodeErr := json.NewDecoder(resp.Body).Decode(&envelope)
		resp.Body.Close()
		if decodeErr != nil {
			time.Sleep(250 * time.Millisecond)
			continue
		}
		for _, msg := range envelope.Items {
			body := decodeQP(msg.Content.Body)
			toList := msg.Content.Headers["To"]
			if len(toList) > 0 && strings.Contains(strings.ToLower(toList[0]), wantTo) {
				if m := resetURLRe.FindStringSubmatch(body); len(m) == 2 {
					return m[1]
				}
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("correo de recuperación para %s no llegó a MailHog", wantTo)
	return ""
}

// TestPsiRequestReset_MailHog_TokenReal valida el pipeline real del flujo de
// recuperación: solicitud → correo con enlace (token de un solo uso) → reset
// con el token real → rotación de Key y nueva contraseña persistida.
func TestPsiRequestReset_MailHog_TokenReal(t *testing.T) {
	if !mailhogAvailable() {
		t.Skip("MailHog no está disponible (dev compose profile 'dev'); se omite")
	}

	config.InitConfig()
	if config.Envs == nil {
		t.Fatal("config.Envs no inicializado")
	}
	orig := *config.Envs
	defer func() { *config.Envs = orig }()

	config.Envs.Environment = "development"
	config.Envs.SMTPHost = "localhost"
	config.Envs.SMTPPort = 21025
	config.Envs.SMTPUser = ""
	config.Envs.SMTPPass = ""
	config.Envs.SMTPFrom = "info@colpsicarabobo.com"
	config.Envs.SMTPFromName = "Colegio de Psicólogos de Carabobo"
	config.Envs.SMTPReplyTo = ""
	config.Envs.MailSignature = "Administración ColPsiCarabobo"

	ms, err := NewMailService()
	if err != nil {
		t.Fatalf("NewMailService() falló: %v", err)
	}
	defer ms.Close()

	psiEmail := "e2e.reset@test.com"
	psi := activePsi()
	psi.Email = psiEmail

	tokensByHash := make(map[string]*domain.PsiPasswordResetToken)
	var usedTokenID uuid.UUID
	var updatedPSI *domain.PsiUserModel
	var markedUsedCalled bool

	repo := &mockPsiRepoSvc{
		GetByIdentifierFunc: func(ctx context.Context, identifier string) (*domain.PsiUserModel, error) {
			return psi, nil
		},
		CreateResetTokenFunc: func(ctx context.Context, token *domain.PsiPasswordResetToken) error {
			tokensByHash[token.TokenHash] = token
			return nil
		},
		GetResetTokenByHashFunc: func(ctx context.Context, tokenHash string) (*domain.PsiPasswordResetToken, error) {
			token, ok := tokensByHash[tokenHash]
			if !ok {
				return nil, nil
			}
			return token, nil
		},
		ResetPasswordFunc: func(ctx context.Context, p *domain.PsiUserModel) error {
			updatedPSI = p
			return nil
		},
		MarkResetTokenUsedFunc: func(ctx context.Context, id uuid.UUID) error {
			markedUsedCalled = true
			usedTokenID = id
			for _, tok := range tokensByHash {
				if tok.ID == id {
					now := time.Now()
					tok.UsedAt = &now
				}
			}
			return nil
		},
	}

	svc := NewPsiService(repo, nil, ms)

	// 1. Solicitar recuperación → el token viaja EN EL CORREO (única exposición).
	if err := svc.RequestPasswordReset(context.Background(), psiEmail); err != nil {
		t.Fatalf("RequestPasswordReset falló: %v", err)
	}

	rawToken := fetchMailhogResetToken(t, psiEmail)
	if rawToken == "" {
		t.Fatal("no se pudo extraer el token del correo")
	}

	// El hash almacenado jamás coincide con el token en claro.
	for storedHash := range tokensByHash {
		if storedHash == rawToken {
			t.Fatal("el token en claro fue persistido (solo debe guardarse el hash)")
		}
	}

	// 2. Usar el token real desde "el correo" para fijar la nueva contraseña.
	newPassword := "NuevaClaveSegura#2026"
	if err := svc.ResetPasswordWithToken(context.Background(), rawToken, newPassword); err != nil {
		t.Fatalf("ResetPasswordWithToken falló con el token real: %v", err)
	}

	if updatedPSI == nil {
		t.Fatal("no se persistió el psicólogo actualizado")
	}
	if bcrypt.CompareHashAndPassword([]byte(updatedPSI.Password), []byte(newPassword)) != nil {
		t.Error("la nueva contraseña no valida contra el hash persistido")
	}
	if updatedPSI.MustChangePassword {
		t.Error("must_change_password debe apagarse tras el reset")
	}
	if !markedUsedCalled || usedTokenID == uuid.Nil {
		t.Error("el token no se marcó como usado (violación de single-use)")
	}

	// 3. Un segundo intento con el mismo token debe fallar (single-use + hash quemado).
	if err := svc.ResetPasswordWithToken(context.Background(), rawToken, "OtraClave#2026"); err == nil {
		t.Error("reusar el token debe ser rechazado")
	}
}

// decodeQP decodifica cuerpo quoted-printable (el estándar con el que MailHog
// entrega el HTML) para que el regex pueda leerlo.
func decodeQP(raw string) string {
	r := quotedprintable.NewReader(strings.NewReader(raw))
	b, err := io.ReadAll(r)
	if err != nil {
		return raw
	}
	return string(b)
}

// fetchMailhogPasswords consulta la API de MailHog hasta que aparezcan correos
// para cada destinatario pedido y devuelve la contraseña temporal extraída del
// cuerpo HTML de cada uno. Falla con t.Fatalf si algún correo no llega a tiempo.
func fetchMailhogPasswords(t *testing.T, want []string) map[string]string {
	t.Helper()

	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(mailhogAPIURL)
		if err == nil {
			var envelope struct {
				Items []mailhogMessage `json:"items"`
			}
			decodeErr := json.NewDecoder(resp.Body).Decode(&envelope)
			resp.Body.Close()
			if decodeErr == nil {
				got := make(map[string]string)
				for _, msg := range envelope.Items {
					body := decodeQP(msg.Content.Body)
					toList := msg.Content.Headers["To"]
					if len(toList) == 0 {
						continue
					}
					to := strings.ToLower(toList[0])
					for _, w := range want {
						if strings.Contains(to, w) {
							if m := passwordRe.FindStringSubmatch(body); len(m) == 2 {
								got[w] = m[1]
							}
						}
					}
				}
				if len(got) == len(want) {
					return got
				}
			}
		}
		time.Sleep(250 * time.Millisecond)
	}

	var missing []string
	for _, w := range want {
		found := false
		resp, err := http.Get(mailhogAPIURL)
		if err == nil {
			var envelope struct {
				Items []mailhogMessage `json:"items"`
			}
			_ = json.NewDecoder(resp.Body).Decode(&envelope)
			resp.Body.Close()
			for _, msg := range envelope.Items {
				for _, toList := range msg.Content.Headers["To"] {
					if strings.Contains(strings.ToLower(toList), w) {
						found = true
					}
				}
			}
		}
		if !found {
			missing = append(missing, w)
		}
	}
	t.Fatalf("correos no llegaron a MailHog para: %v", missing)
	return nil
}