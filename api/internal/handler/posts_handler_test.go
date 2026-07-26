package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/service"
)

func TestListPosts(t *testing.T) {
	t.Run("public_role", func(t *testing.T) {
		posts := []domain.Post{
			*testPost(uuid.New(), "Public Post", domain.PostStatusPublished, "public"),
		}
		repo := &mockPostRepo{
			ListFunc: func(_ context.Context, filter domain.PostFilter, page, limit int) ([]domain.Post, int64, error) {
				require.Equal(t, "public", filter.Type)
				require.Contains(t, filter.Status, domain.PostStatusPublished)
				return posts, 1, nil
			},
		}
		svc := service.NewPostService(repo, nil)
		h := NewPostHandler(svc)

		app := setupHybridRoute(fiber.MethodGet, "/posts", h.ListPosts, &mockAdminRepo{}, &mockPsiRepo{})
		req := httptest.NewRequest(fiber.MethodGet, "/posts/?page=1&limit=10", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		require.NotNil(t, result["data"])
	})

	t.Run("admin_role", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		repo := &mockPostRepo{
			ListFunc: func(_ context.Context, filter domain.PostFilter, page, limit int) ([]domain.Post, int64, error) {
				// Admin sees all statuses (nil filter)
				require.Nil(t, filter.Status)
				return nil, 0, nil
			},
		}
		svc := service.NewPostService(repo, nil)
		h := NewPostHandler(svc)

		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		app := setupHybridRoute(fiber.MethodGet, "/posts", h.ListPosts, adminRepo, &mockPsiRepo{})

		token := generateTestToken(adminID.String(), "admin", futureTime())
		req := httptest.NewRequest(fiber.MethodGet, "/posts/?page=1&limit=10", nil)
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("psi_role", func(t *testing.T) {
		psiID := uuid.New()
		psi := testPsiUser(psiID)

		repo := &mockPostRepo{
			ListFunc: func(_ context.Context, filter domain.PostFilter, page, limit int) ([]domain.Post, int64, error) {
				// Psi sees published only
				require.Contains(t, filter.Status, domain.PostStatusPublished)
				return nil, 0, nil
			},
		}
		svc := service.NewPostService(repo, nil)
		h := NewPostHandler(svc)

		psiRepo := &mockPsiRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.PsiUserModel, error) {
				return psi, nil
			},
		}
		app := setupHybridRoute(fiber.MethodGet, "/posts", h.ListPosts, &mockAdminRepo{}, psiRepo)

		token := generateTestToken(psiID.String(), "psi", futureTime())
		req := httptest.NewRequest(fiber.MethodGet, "/posts/?page=1&limit=10", nil)
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("limit_capped", func(t *testing.T) {
		repo := &mockPostRepo{
			ListFunc: func(_ context.Context, filter domain.PostFilter, page, limit int) ([]domain.Post, int64, error) {
				require.LessOrEqual(t, limit, 100)
				return nil, 0, nil
			},
		}
		svc := service.NewPostService(repo, nil)
		h := NewPostHandler(svc)

		app := setupHybridRoute(fiber.MethodGet, "/posts", h.ListPosts, &mockAdminRepo{}, &mockPsiRepo{})
		req := httptest.NewRequest(fiber.MethodGet, "/posts/?limit=200", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestCreatePost(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)

		repo := &mockPostRepo{
			CreateFunc: func(_ context.Context, post *domain.Post, content *domain.TextModel) error {
				return nil
			},
		}
		svc := service.NewPostService(repo, nil)
		h := NewPostHandler(svc)

		token := generateTestToken(adminID.String(), "admin", futureTime())
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		app := setupAdminRoute(fiber.MethodPost, "/posts", h.CreatePost, adminRepo, &mockPsiRepo{})

		body := `{"title":"New Post","content":"Hello world","type":"public","status":"draft"}`
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/admin/posts/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusCreated, resp.StatusCode)
	})

	t.Run("insufficient_permission", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, false, false)
		admin.CanPublish = false

		repo := &mockPostRepo{}
		svc := service.NewPostService(repo, nil)
		h := NewPostHandler(svc)

		token := generateTestToken(adminID.String(), "admin", futureTime())
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		app := setupAdminRoute(fiber.MethodPost, "/posts", h.CreatePost, adminRepo, &mockPsiRepo{})

		body := `{"title":"Post","content":"Content","type":"public","status":"draft"}`
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/admin/posts/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		// Handler returns 500 for all service errors (no specific ErrPostPermDenied mapping)
		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	})
}

func TestGetPost(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		postID := uuid.New()
		post := testPost(postID, "Public Post", domain.PostStatusPublished, "public")

		repo := &mockPostRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.Post, error) {
				return post, nil
			},
		}
		svc := service.NewPostService(repo, nil)
		h := NewPostHandler(svc)

		app := setupHybridRoute(fiber.MethodGet, "/posts/:id", h.GetPost, &mockAdminRepo{}, &mockPsiRepo{})
		req := httptest.NewRequest(fiber.MethodGet, "/posts/"+postID.String(), nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("not_found", func(t *testing.T) {
		repo := &mockPostRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.Post, error) {
				return nil, domain.ErrPsiNotFound
			},
		}
		svc := service.NewPostService(repo, nil)
		h := NewPostHandler(svc)

		app := setupHybridRoute(fiber.MethodGet, "/posts/:id", h.GetPost, &mockAdminRepo{}, &mockPsiRepo{})
		req := httptest.NewRequest(fiber.MethodGet, "/posts/"+uuid.New().String(), nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

func TestUpdatePost(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		adminID := uuid.New()
		admin := testAdmin(adminID, true, true)
		postID := uuid.New()
		post := testPost(postID, "Old Title", domain.PostStatusDraft, "public")

		repo := &mockPostRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.Post, error) {
				return post, nil
			},
			UpdateFunc: func(_ context.Context, p *domain.Post, content *domain.TextModel) error {
				return nil
			},
		}
		svc := service.NewPostService(repo, nil)
		h := NewPostHandler(svc)

		token := generateTestToken(adminID.String(), "admin", futureTime())
		adminRepo := &mockAdminRepo{
			GetByIDFunc: func(_ context.Context, id uuid.UUID) (*domain.UserAdmin, error) {
				return admin, nil
			},
		}
		app := setupAdminRoute(fiber.MethodPatch, "/posts/:id", h.UpdatePost, adminRepo, &mockPsiRepo{})

		body := `{"title":"Updated Title"}`
		req := httptest.NewRequest(fiber.MethodPatch, "/api/v1/admin/posts/"+postID.String(), strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		authRequest(req, token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestGetSiteMapHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		posts := []domain.Post{
			*testPost(uuid.New(), "Sitemap Post", domain.PostStatusPublished, "public"),
		}
		repo := &mockPostRepo{
			GetSitemapPostsFunc: func(_ context.Context) ([]domain.Post, error) {
				return posts, nil
			},
		}
		svc := service.NewPostService(repo, nil)
		h := NewPostHandler(svc)

		app := setupPublicRoute(fiber.MethodGet, "/posts/public/sitemap-posts", h.GetSiteMapHandler)
		req := httptest.NewRequest(fiber.MethodGet, "/posts/public/sitemap-posts", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}
