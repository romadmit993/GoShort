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

// func Test_handleShortenPost(t *testing.T) {
// 	requestBody := models.Shorten{URL: "http://practicum.yandex.ru"}
// 	jsonBody, _ := json.Marshal(requestBody)
// 	req, err := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(jsonBody))
// 	if err != nil {
// 		t.Fatalf("Ошибка при создании запроса: %v", err)
// 	}
// 	req.Header.Set("Content-Type", "application/json")
// 	rr := httptest.NewRecorder()
// 	handler := handleShortenPost()
// 	handler.ServeHTTP(rr, req)
// 	if status := rr.Code; status != http.StatusCreated {
// 		t.Errorf("Ожидался статус код %v, получен %v", http.StatusCreated, status)
// 	}
// 	expectedContentType := "application/json"
// 	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
// 		t.Errorf("Ожидался Content-Type %v, получен %v", expectedContentType, contentType)
// 	}
// 	var response models.Shorten
// 	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
// 		t.Fatalf("Ошибка при декодировании ответа: %v", err)
// 	}
// 	if response.Result == "" {
// 		t.Error("Ожидался непустой результат с коротким URL")
// 	}
// }
