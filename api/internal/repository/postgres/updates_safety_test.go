// Package postgres contiene pruebas unitarias que validan la corrección de los
// métodos de actualización (Updates) tras el FIX-30.
//
// Estos tests garantizan que los campos booleanos e int con zero-values
// se persisten correctamente usando gorm.Expr, evitando el bug clásico
// de Save() que sobreescribe booleanos con false cuando el valor real es true.
//
// Utiliza SQLite en memoria para no depender de una instancia PostgreSQL.
// Crea las tablas manualmente para evitar incompatibilidades con funciones
// específicas de PostgreSQL (como uuidv7).
package postgres

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veniversvm/ColPsiCarabobo/api/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupDryRunDB crea una base de datos SQLite en memoria con tablas manuales
// que replican la estructura mínima necesaria para probar Updates().
func setupDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	stmts := []string{
		`CREATE TABLE user_admins (
			id TEXT PRIMARY KEY,
			created_at datetime, updated_at datetime, deleted_at datetime,
			create_by text, update_by text, create_by_id TEXT, update_by_id TEXT,
			username text NOT NULL UNIQUE, email text NOT NULL UNIQUE,
			password text NOT NULL, key text,
			is_active numeric DEFAULT 1, must_change_password numeric DEFAULT 0,
			sudo numeric DEFAULT 0,
			can_create_psi numeric DEFAULT 0, can_update_psi numeric DEFAULT 0,
			can_delete_psi numeric DEFAULT 0,
			can_create_admin numeric DEFAULT 0, can_update_admin numeric DEFAULT 0,
			can_delete_admin numeric DEFAULT 0,
			can_publish numeric DEFAULT 0, can_update_publish numeric DEFAULT 0,
			can_delete_publish numeric DEFAULT 0,
			can_send_notifications numeric DEFAULT 0, can_manage_notifications numeric DEFAULT 0,
			can_read_notifications numeric DEFAULT 0,
			can_create_tags numeric DEFAULT 0, can_edit_tags numeric DEFAULT 0,
			can_delete_tags numeric DEFAULT 0
		)`,
		`CREATE TABLE psi_specialty_models (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name text NOT NULL UNIQUE, description text,
			active numeric DEFAULT 1,
			created_at datetime, updated_at datetime, deleted_at datetime,
			create_by text, update_by text, create_by_id TEXT, update_by_id TEXT
		)`,
		`CREATE TABLE text_models (
			id TEXT PRIMARY KEY,
			created_at datetime, updated_at datetime, deleted_at datetime,
			create_by text, update_by text, create_by_id TEXT, update_by_id TEXT,
			content text
		)`,
		`CREATE TABLE psi_user_post_grades (
			id TEXT PRIMARY KEY,
			created_at datetime, updated_at datetime, deleted_at datetime,
			create_by text, update_by text, create_by_id TEXT, update_by_id TEXT,
			psi_user_id TEXT, type text NOT NULL,
			title text NOT NULL, university text, graduation_year integer,
			description text, active numeric DEFAULT 1,
			pic_one_s3_key text, pic_two_s3_key text, pic_three_s3_key text
		)`,
		`CREATE TABLE psi_user_social_networks (
			id TEXT PRIMARY KEY,
			created_at datetime, updated_at datetime, deleted_at datetime,
			create_by text, update_by text, create_by_id TEXT, update_by_id TEXT,
			psi_user_id TEXT NOT NULL,
			name text NOT NULL, url text NOT NULL, is_active numeric DEFAULT 1
		)`,
		`CREATE TABLE posts (
			id TEXT PRIMARY KEY,
			created_at datetime, updated_at datetime, deleted_at datetime,
			create_by text, update_by text, create_by_id TEXT, update_by_id TEXT,
			title text NOT NULL, short_description text, type text DEFAULT 'public',
			text_id TEXT, image_s3_key text, status text NOT NULL DEFAULT 'draft',
			publish_at datetime
		)`,
	}

	for _, s := range stmts {
		require.NoError(t, db.Exec(s).Error)
	}
	return db
}

// TestAdminRepo_Update_PreservesBooleanFalse valida que Update() de admin
// persista correctamente booleanos en false (zero-value) usando gorm.Expr.
func TestAdminRepo_Update_PreservesBooleanFalse(t *testing.T) {
	db := setupDryRunDB(t)
	r := NewAdminRepository(db)

	adminID := uuid.New()
	admin := &domain.UserAdmin{
		ID: adminID,
		Credentials: domain.Credentials{
			Username: "test_admin",
			Email:    "test@admin.com",
			Password: "hashed",
			IsActive: true,
		},
		Sudo:                 true,
		CanCreatePsi:         true,
		CanUpdatePsi:         true,
		CanDeletePsi:         true,
		CanCreateAdmin:       true,
		CanUpdateAdmin:       true,
		CanDeleteAdmin:       true,
		CanPublish:           true,
		CanUpdatePublish:     true,
		CanDeletePublish:     true,
		CanSendNotifications: true,
		CanManageNotifications: true,
		CanReadNotifications: true,
		CanCreateTags:        true,
		CanEditTags:          true,
		CanDeleteTags:        true,
	}
	require.NoError(t, r.Create(t.Context(), admin))

	admin.Sudo = false
	admin.CanCreatePsi = false
	admin.CanUpdatePsi = false
	admin.CanDeletePsi = false
	admin.CanCreateAdmin = false
	admin.CanUpdateAdmin = false
	admin.CanDeleteAdmin = false
	admin.CanPublish = false
	admin.CanUpdatePublish = false
	admin.CanDeletePublish = false
	admin.CanSendNotifications = false
	admin.CanManageNotifications = false
	admin.CanReadNotifications = false
	admin.CanCreateTags = false
	admin.CanEditTags = false
	admin.CanDeleteTags = false
	admin.IsActive = false

	err := r.Update(t.Context(), admin)
	require.NoError(t, err)

	fetched, err := r.GetByID(t.Context(), adminID)
	require.NoError(t, err)
	assert.False(t, fetched.Sudo, "Sudo should be false")
	assert.False(t, fetched.IsActive, "IsActive should be false")
	assert.False(t, fetched.CanCreatePsi, "CanCreatePsi should be false")
	assert.False(t, fetched.CanUpdatePsi, "CanUpdatePsi should be false")
	assert.False(t, fetched.CanDeletePsi, "CanDeletePsi should be false")
	assert.False(t, fetched.CanCreateAdmin, "CanCreateAdmin should be false")
	assert.False(t, fetched.CanUpdateAdmin, "CanUpdateAdmin should be false")
	assert.False(t, fetched.CanDeleteAdmin, "CanDeleteAdmin should be false")
	assert.False(t, fetched.CanPublish, "CanPublish should be false")
	assert.False(t, fetched.CanUpdatePublish, "CanUpdatePublish should be false")
	assert.False(t, fetched.CanDeletePublish, "CanDeletePublish should be false")
	assert.False(t, fetched.CanSendNotifications, "CanSendNotifications should be false")
	assert.False(t, fetched.CanManageNotifications, "CanManageNotifications should be false")
	assert.False(t, fetched.CanReadNotifications, "CanReadNotifications should be false")
	assert.False(t, fetched.CanCreateTags, "CanCreateTags should be false")
	assert.False(t, fetched.CanEditTags, "CanEditTags should be false")
	assert.False(t, fetched.CanDeleteTags, "CanDeleteTags should be false")
}

// TestAdminRepo_Update_PreservesBooleanTrue valida round-trip true→false→true.
func TestAdminRepo_Update_PreservesBooleanTrue(t *testing.T) {
	db := setupDryRunDB(t)
	r := NewAdminRepository(db)

	adminID := uuid.New()
	admin := &domain.UserAdmin{
		ID: adminID,
		Credentials: domain.Credentials{
			Username: "test_admin_2",
			Email:    "test2@admin.com",
			Password: "hashed",
			IsActive: false,
		},
	}
	require.NoError(t, r.Create(t.Context(), admin))

	admin.IsActive = true
	admin.Sudo = true
	admin.CanPublish = true
	require.NoError(t, r.Update(t.Context(), admin))

	fetched, err := r.GetByID(t.Context(), adminID)
	require.NoError(t, err)
	assert.True(t, fetched.IsActive)
	assert.True(t, fetched.Sudo)
	assert.True(t, fetched.CanPublish)
	assert.False(t, fetched.CanCreatePsi, "Unchecked booleans must remain false")
}

// TestSpecialtyRepo_Update_PreservesActiveFalse valida que Update() de
// especialidad persista Active=false correctamente.
func TestSpecialtyRepo_Update_PreservesActiveFalse(t *testing.T) {
	db := setupDryRunDB(t)
	r := NewSpecialtyRepository(db)

	sp := &domain.PsiSpecialtyModel{
		Name:   "Clínica",
		Active: true,
	}
	require.NoError(t, r.Create(t.Context(), sp))

	sp.Active = false
	sp.Name = "Clínica (Inactiva)"
	require.NoError(t, r.Update(t.Context(), sp))

	fetched, err := r.GetByID(t.Context(), sp.ID, false)
	require.NoError(t, err)
	assert.False(t, fetched.Active, "Active should be false")
	assert.Equal(t, "Clínica (Inactiva)", fetched.Name)
}

// TestPostRepo_Update_PreservesStatusAndType valida que Update() de post
// persista status y tipo correctamente.
func TestPostRepo_Update_PreservesStatusAndType(t *testing.T) {
	db := setupDryRunDB(t)
	r := NewPostRepository(db)

	text := &domain.TextModel{
		ID:      uuid.New(),
		Content: "Original content",
	}
	post := &domain.Post{
		ID:     uuid.New(),
		Title:  "Test Post",
		Type:   "public",
		Status: domain.PostStatusDraft,
	}
	require.NoError(t, r.Create(t.Context(), post, text))

	post.Status = domain.PostStatusPublished
	post.Type = "psi"
	require.NoError(t, r.Update(t.Context(), post, nil))

	fetched, err := r.GetByID(t.Context(), post.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.PostStatusPublished, fetched.Status)
	assert.Equal(t, "psi", fetched.Type)
}

// TestPostRepo_Update_TextContent valida que Update() con texto actualice
// el contenido correctamente.
func TestPostRepo_Update_TextContent(t *testing.T) {
	db := setupDryRunDB(t)
	r := NewPostRepository(db)

	text := &domain.TextModel{
		ID:      uuid.New(),
		Content: "Original",
	}
	post := &domain.Post{
		ID:     uuid.New(),
		Title:  "Post with text",
		TextID: text.ID,
	}
	require.NoError(t, r.Create(t.Context(), post, text))

	updatedText := &domain.TextModel{
		ID:      text.ID,
		Content: "Updated content",
	}
	require.NoError(t, r.Update(t.Context(), post, updatedText))

	fetched, err := r.GetByID(t.Context(), post.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated content", fetched.Text.Content)
}

// TestPsiRepo_UpdatePostGrade_PreservesActiveFalse valida que UpdatePostGrade()
// persista Active=false (zero-value) correctamente.
func TestPsiRepo_UpdatePostGrade_PreservesActiveFalse(t *testing.T) {
	db := setupDryRunDB(t)
	r := NewPsiRepository(db)

	pg := &domain.PsiUserPostGrade{
		ID:        uuid.New(),
		PsiUserID: uuid.New(),
		Type:      domain.Diplomado,
		Title:     "Psicología Clínica",
		Active:    true,
	}
	require.NoError(t, r.CreatePostGrade(t.Context(), pg))

	pg.Active = false
	pg.Title = "Psicología Clínica (Inactiva)"
	require.NoError(t, r.UpdatePostGrade(t.Context(), pg))

	var fetched domain.PsiUserPostGrade
	err := db.Where("id = ?", pg.ID).First(&fetched).Error
	require.NoError(t, err)
	assert.False(t, fetched.Active, "Active should be false")
	assert.Equal(t, "Psicología Clínica (Inactiva)", fetched.Title)
}

// TestPsiRepo_UpdateSocialNetwork_PreservesIsActiveFalse valida que
// UpdateSocialNetwork() persista IsActive=false correctamente.
func TestPsiRepo_UpdateSocialNetwork_PreservesIsActiveFalse(t *testing.T) {
	db := setupDryRunDB(t)
	r := NewPsiRepository(db)

	sn := &domain.PsiUserSocialNetwork{
		ID:        uuid.New(),
		PsiUserID: uuid.New(),
		Name:      "Instagram",
		URL:       "https://instagram.com/test",
		IsActive:  true,
	}
	require.NoError(t, r.CreateSocialNetwork(t.Context(), sn))

	sn.IsActive = false
	sn.URL = "https://instagram.com/test_updated"
	require.NoError(t, r.UpdateSocialNetwork(t.Context(), sn))

	var fetched domain.PsiUserSocialNetwork
	err := db.Where("id = ?", sn.ID).First(&fetched).Error
	require.NoError(t, err)
	assert.False(t, fetched.IsActive, "IsActive should be false")
	assert.Equal(t, "https://instagram.com/test_updated", fetched.URL)
}
