// api/internal/service/psi_service_abs_sync_test.go
package service

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

// mockSyncRepo implementa únicamente lo que la sync necesita.
type mockSyncRepo struct {
	domain.PsiUserRepository
	mu        sync.Mutex
	all       []domain.PsiUserModel
	updatedID map[string]string // psi id -> abs id persistido
}

func (m *mockSyncRepo) GetAllForABSSync(ctx context.Context) ([]domain.PsiUserModel, error) {
	return m.all, nil
}

func (m *mockSyncRepo) UpdateAudioBookShellID(ctx context.Context, psi *domain.PsiUserModel) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updatedID[psi.ID.String()] = psi.AudioBookShellId
	return nil
}

func newMockSyncRepo(psis ...domain.PsiUserModel) *mockSyncRepo {
	return &mockSyncRepo{
		all:       psis,
		updatedID: map[string]string{},
	}
}

func TestSyncAudiobookshelfAccounts_CreatesAndDeactivates(t *testing.T) {
	fake := newFakeAbs()
	fake.adminUser = "colpsi-bot"
	fake.adminPass = "adminpass"

	svc, _ := buildSvc(t, fake)

	// Estado deseado: 1111 solvente+activo (se crea), 2222 solvente pero
	// desactivado (no se crea), 3333 insolvente con cuenta previa (se desactiva).
	repo := newMockSyncRepo(
		domain.PsiUserModel{ID: uuid.New(), CI: 1111, Solvent: true, Credentials: domain.Credentials{IsActive: true}},
		domain.PsiUserModel{ID: uuid.New(), CI: 2222, Solvent: true, Credentials: domain.Credentials{IsActive: false}},
		domain.PsiUserModel{ID: uuid.New(), CI: 3333, Solvent: false, Credentials: domain.Credentials{IsActive: true}},
	)

	// 3333 ya tiene cuenta activa en ABS.
	fake.users["psi_3333"] = svc.passwordFor("psi_3333")
	fake.usersActive["psi_3333"] = true
	fake.usersIDs["psi_3333"] = "id-psi_3333"

	psiSvc := &PsiService{repo: repo, absSvc: svc}
	report, err := psiSvc.SyncAudiobookshelfAccounts(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, report.Created)     // psi_1111
	require.Equal(t, 1, report.Deactivated) // psi_3333
	require.Empty(t, report.Errors)

	// psi_1111 existe con la clave derivada y está activa.
	require.True(t, fake.usersActive["psi_1111"])
	require.Equal(t, svc.passwordFor("psi_1111"), fake.users["psi_1111"])
	// psi_2222 no se creó (inactivo).
	_, ok := fake.users["psi_2222"]
	require.False(t, ok)
	// psi_3333 quedó desactivada.
	require.False(t, fake.usersActive["psi_3333"])

	// El id ABS de 1111 se persistió en el expediente.
	require.Equal(t, "id-psi_1111", repo.updatedID[repo.all[0].ID.String()])
}

func TestSyncAudiobookshelfAccounts_AlreadyExists(t *testing.T) {
	fake := newFakeAbs()
	fake.adminUser = "colpsi-bot"
	fake.adminPass = "adminpass"

	svc, _ := buildSvc(t, fake)

	psiID := uuid.New()
	repo := newMockSyncRepo(
		domain.PsiUserModel{ID: psiID, CI: 5555, Solvent: true, Credentials: domain.Credentials{IsActive: true}},
	)

	// 5555 ya tiene cuenta ABS válida (misma clave derivada).
	fake.users["psi_5555"] = svc.passwordFor("psi_5555")
	fake.usersActive["psi_5555"] = true
	fake.usersIDs["psi_5555"] = "id-psi_5555"

	psiSvc := &PsiService{repo: repo, absSvc: svc}
	report, err := psiSvc.SyncAudiobookshelfAccounts(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, report.Created)
	require.Equal(t, 0, report.Deactivated)
	require.Equal(t, 1, report.Skipped)
}

func TestSyncAudiobookshelfAccounts_IgnoresNonPsiUsers(t *testing.T) {
	fake := newFakeAbs()
	fake.adminUser = "colpsi-bot"
	fake.adminPass = "adminpass"

	svc, _ := buildSvc(t, fake)

	repo := newMockSyncRepo()
	// Un usuario que no es psi_* (p.ej. admin de ABS) no debe desactivarse.
	fake.users["root"] = "clave"
	fake.usersActive["root"] = true
	fake.usersIDs["root"] = "id-root"

	psiSvc := &PsiService{repo: repo, absSvc: svc}
	report, err := psiSvc.SyncAudiobookshelfAccounts(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, report.Deactivated)
	require.True(t, fake.usersActive["root"])
}

func TestSyncAudiobookshelfAccounts_NoProbeLogins(t *testing.T) {
	fake := newFakeAbs()
	fake.adminUser = "colpsi-bot"
	fake.adminPass = "adminpass"

	svc, _ := buildSvc(t, fake)

	repo := newMockSyncRepo(
		domain.PsiUserModel{ID: uuid.New(), CI: 1111, Solvent: true, Credentials: domain.Credentials{IsActive: true}},
		domain.PsiUserModel{ID: uuid.New(), CI: 2222, Solvent: true, Credentials: domain.Credentials{IsActive: true}},
		domain.PsiUserModel{ID: uuid.New(), CI: 3333, Solvent: true, Credentials: domain.Credentials{IsActive: true}},
	)

	psiSvc := &PsiService{repo: repo, absSvc: svc}
	report, err := psiSvc.SyncAudiobookshelfAccounts(context.Background())
	require.NoError(t, err)
	require.Equal(t, 3, report.Created)
	require.Empty(t, report.Errors)

	// Solo el login del admin para ListUsers; sin probes fallidos ni re-logins
	// por usuario (un probe por cuenta agotaba el rate limiter de ABS).
	require.Equal(t, 1, fake.loginCalls)
	require.Equal(t, 3, fake.createCalls)
}

func TestSyncAudiobookshelfAccounts_NilAbsSvc(t *testing.T) {
	repo := newMockSyncRepo(
		domain.PsiUserModel{ID: uuid.New(), CI: 9999, Solvent: true, Credentials: domain.Credentials{IsActive: true}},
	)
	psiSvc := &PsiService{repo: repo, absSvc: nil}
	report, err := psiSvc.SyncAudiobookshelfAccounts(context.Background())
	require.NoError(t, err)
	require.NotNil(t, report)
	require.Equal(t, 0, report.Created)
}
