package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Тестирование handlePost
func TestHandlePost(t *testing.T) {
	// Создаем тестовый запрос
	originalURL := "https://example.com"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	req.Header.Set("Content-Type", "text/plain")

	// Создаем ResponseRecorder для записи ответа
	w := httptest.NewRecorder()

	// Вызываем handlePost
	handlePost(w, req)

	// Проверяем статус-код
	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Ожидался статус-код %d, получен %d", http.StatusCreated, resp.StatusCode)
	}

	// Читаем тело ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Ошибка при чтении тела ответа: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем, что сокращенный URL имеет правильный формат
	shortURL := string(body)
	if !strings.HasPrefix(shortURL, "http://localhost:8080/") {
		t.Errorf("Сокращенный URL имеет неверный формат: %s", shortURL)
	}

	// Извлекаем shortID из сокращенного URL
	shortID := strings.TrimPrefix(shortURL, "http://localhost:8080/")

	// Проверяем, что оригинальный URL сохранен в urlStore
	if urlStore[shortID] != originalURL {
		t.Errorf("Ожидался оригинальный URL %s, получен %s", originalURL, urlStore[shortID])
	}
}

// Тестирование handleGet
func TestHandleGet(t *testing.T) {
	// Добавляем тестовый URL в urlStore
	shortID := "testID"
	originalURL := "https://example.com"
	urlStore[shortID] = originalURL

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)

	// Создаем ResponseRecorder для записи ответа
	w := httptest.NewRecorder()

	// Вызываем handleGet
	handleGet(w, req)

	// Проверяем статус-код
	resp := w.Result()
	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("Ожидался статус-код %d, получен %d", http.StatusTemporaryRedirect, resp.StatusCode)
	}

	// Проверяем заголовок Location
	location := resp.Header.Get("Location")
	if location != originalURL {
		t.Errorf("Ожидался Location %s, получен %s", originalURL, location)
	}

	// Проверяем обработку несуществующего shortID
	req = httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w = httptest.NewRecorder()

	handleGet(w, req)

	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Ожидался статус-код %d, получен %d", http.StatusBadRequest, resp.StatusCode)
	}
}
