package main

import (
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
	handlePost()(w, req)

	resp := w.Result()
	defer resp.Body.Close() // Закрываем тело ответа

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Ожидался статус-код %d, получен %d", http.StatusCreated, resp.StatusCode)
	}
}

func TestHandleGet(t *testing.T) {
	shortID := "testID"
	originalURL := "https://example.com"
	urlStore[shortID] = originalURL

	req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	w := httptest.NewRecorder()
	handleGet()(w, req)

	resp := w.Result()
	defer resp.Body.Close() // Закрываем тело ответа
}
