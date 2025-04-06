package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"romadmit993/GoShort/internal/config"
	"romadmit993/GoShort/internal/handlers"
	"romadmit993/GoShort/internal/storage"
	"strings"
	"testing"
)

func TestHandlePost(t *testing.T) {
	originalURL := "https://example.com"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	req.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	db, _ := sql.Open("pgx", config.Config.Database)
	handlers.HandlePost(db)(w, req)

	resp := w.Result()
	defer resp.Body.Close() // Закрываем тело ответа

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Ожидался статус-код %d, получен %d", http.StatusCreated, resp.StatusCode)
	}
}

func TestHandleGet(t *testing.T) {
	shortID := "testID"
	originalURL := "https://example.com"
	storage.URLStore[shortID] = originalURL

	req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	w := httptest.NewRecorder()
	db, _ := sql.Open("pgx", config.Config.Database)
	handlers.HandleGet(db)(w, req)

	resp := w.Result()
	defer resp.Body.Close() // Закрываем тело ответа
}
func TestFileStorage(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	// Запуск сервера с тестовым файлом
	go func() {
		config.Config.FileStorage = tempFile.Name()
		main()
	}()

	// Тест HTTP запросов и проверка файла
	// ...
}
