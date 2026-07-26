package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/veniversvm/ColPsiCarabobo/api/internal/config"
)

// actualizarEnAudiobookshelf patches an existing Audiobookshelf user profile with the provided fields.
func (s *PsiService) actualizarEnAudiobookshelf(ctx context.Context, absID string, username, password, email *string) error {
	if absID == "" {
		return nil
	}

	url := fmt.Sprintf("http://audiobookshelf:80/api/users/%s", absID)

	payload := map[string]interface{}{}
	if username != nil {
		payload["username"] = *username
	}
	if password != nil {
		payload["password"] = *password
	}
	if email != nil {
		payload["email"] = *email
	}

	if len(payload) == 0 {
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Envs.AbsAdminToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("audiobookshelf respondió con un código inesperado al actualizar el perfil")
	}

	return nil
}

// sincronizarConAudiobookshelf creates a new Audiobookshelf user and returns its ID.
func (s *PsiService) sincronizarConAudiobookshelf(ctx context.Context, username, password, email string) (string, error) {
	url := "http://audiobookshelf:80/api/users"

	payload := map[string]interface{}{
		"username": username,
		"password": password,
		"email":    email,
		"type":     "user",
		"isActive": true,
		"permissions": map[string]bool{
			"download":           true,
			"updateProgress":     true,
			"accessAllLibraries": true,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Envs.AbsAdminToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return "", nil
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", errors.New("audiobookshelf respondió con código de estado inesperado")
	}

	var absData AudiobookshelfUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&absData); err != nil {
		return "", errors.New("error al decodificar la respuesta de audiobookshelf")
	}

	return absData.User.ID, nil
}
