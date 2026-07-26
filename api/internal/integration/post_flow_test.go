package integration

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
)

func TestPostFlow_CreateAndView(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	var postID string
	t.Run("AdminCreatePost", func(t *testing.T) {
		body := `{"title":"Test Post","type":"public","status":"published","content":"<p>Hello World</p>"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/posts/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("PublicListPosts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetPostByID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/", nil)
		resp, _ := app.Test(req)
		data := decodeBody(t, resp)
		posts := data["data"].([]interface{})
		require.GreaterOrEqual(t, len(posts), 1)
		postMap := posts[0].(map[string]interface{})
		postID = postMap["id"].(string)

		req2 := httptest.NewRequest(http.MethodGet, "/api/v1/posts/"+postID, nil)
		resp2, err := app.Test(req2)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp2.StatusCode)
	})

	_ = postID
}

func TestPostFlow_UpdateAndArchive(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	post := seedPost(t, "ToUpdate", domain.PostStatusPublished, "public")
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	t.Run("UpdatePost", func(t *testing.T) {
		body := `{"title":"Updated Title","status":"archived"}`
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/posts/"+post.ID.String(), strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestPostFlow_ScheduledPublish(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	body := `{"title":"Scheduled Post","type":"public","status":"scheduled","publish_at":"2020-01-01T00:00:00Z","content":"<p>Scheduled content</p>"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/posts/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(token))
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestPostFlow_DraftNotVisible(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	seedPost(t, "MyDraft", domain.PostStatusDraft, "public")
	app := buildTestApp(testDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	data := decodeBody(t, resp)
	total := int64(data["total"].(float64))
	require.Equal(t, int64(0), total)
}

func TestPostFlow_ListWithAuth(t *testing.T) {
	truncateAll(t)
	seedSudo(t)
	seedPost(t, "DraftOnly", domain.PostStatusDraft, "public")
	seedPost(t, "PublishedOnly", domain.PostStatusPublished, "public")
	app := buildTestApp(testDB)

	token := loginAdmin(t, app, "sudo", "Sudo123!@#")

	t.Run("PublicSeesOnlyPublished", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		data := decodeBody(t, resp)
		total := int64(data["total"].(float64))
		require.Equal(t, int64(1), total)
	})

	t.Run("AdminSeesAll", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/", nil)
		req.Header.Set("Authorization", authHeader(token))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
