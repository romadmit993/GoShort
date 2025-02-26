package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlePost(t *testing.T) {
	originalURL := "https://example.com"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	req.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	handlePost(w, req)

	resp := w.Result()
	defer resp.Body.Close() // Закрываем тело ответа

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Ожидался статус-код %d, получен %d", http.StatusCreated, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Ошибка при чтении тела ответа: %v", err)
	}

	shortURL := string(body)
	if !strings.HasPrefix(shortURL, "http://localhost:8080/") {
		t.Errorf("Сокращенный URL имеет неверный формат: %s", shortURL)
	}
}

func TestHandleGet(t *testing.T) {
	shortID := "testID"
	originalURL := "https://example.com"
	urlStore[shortID] = originalURL

	req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	w := httptest.NewRecorder()
	handleGet(w, req)

	resp := w.Result()
	defer resp.Body.Close() // Закрываем тело ответа

	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("Ожидался статус-код %d, получен %d", http.StatusTemporaryRedirect, resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location != originalURL {
		t.Errorf("Ожидался Location %s, получен %s", originalURL, location)
	}
}
