package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"romadmit993/GoShort/internal/config"
	"romadmit993/GoShort/internal/database"
	customMiddleware "romadmit993/GoShort/internal/middleware"
	"romadmit993/GoShort/internal/models"
	"romadmit993/GoShort/internal/storage"

	"database/sql"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func HandlePost(db *sql.DB) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Ошибка чтения тела запроса", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		originalURL := string(body)
		if !isValidURL(originalURL) {
			http.Error(w, "Некорректный URL", http.StatusBadRequest)
			return
		}

		shortID := storage.GenerateShortID()
		storage.StoreMux.Lock()
		storage.URLStore[shortID] = originalURL
		storage.SaveShortURLFile(shortID, originalURL)
		if config.Config.Database != "" {
			database.SaveDataBase(db, shortID, originalURL)
		}
		storage.StoreMux.Unlock()

		shortURL := fmt.Sprintf("%s%s", config.Config.BaseAddress, shortID)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, shortURL)
	}
	return http.HandlerFunc(fn)
}

func handleShortenPost(db *sql.DB) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var apiShorten models.Shorten
		if err := json.NewDecoder(r.Body).Decode(&apiShorten); err != nil {
			http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		if apiShorten.URL == "" || !isValidURL(apiShorten.URL) {
			http.Error(w, "Некорректный URL", http.StatusBadRequest)
			return
		}
		shortID := storage.GenerateShortID()
		storage.StoreMux.Lock()
		storage.URLStore[shortID] = apiShorten.URL
		storage.SaveShortURLFile(shortID, apiShorten.URL)
		if config.Config.Database != "" {
			database.SaveDataBase(db, shortID, apiShorten.URL)
		}
		storage.StoreMux.Unlock()
		shortURL := fmt.Sprintf("%s/%s", config.Config.BaseAddress, shortID)
		response := models.Shorten{
			Result: shortURL,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Ошибка при формировании ответа", http.StatusInternalServerError)
		}
	}
	return http.HandlerFunc(fn)
}

func HandleGet(db *sql.DB) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "ID не может быть пустым", http.StatusBadRequest)
			return
		}
		storage.StoreMux.RLock()
		var existsDataBase bool

		if config.Config.Database != "" {
			existsDataBase = true
			//existsDataBase = database.СheckRecord(db, id)
		}
		originalURL, exists := storage.URLStore[id]
		if !exists {
			_, exists = storage.ReadFileAndCheckID(id)
		}
		storage.StoreMux.RUnlock()
		if !exists {
			exists = existsDataBase
		}
		if !exists {
			http.Error(w, "Сокращённый URL не найден", http.StatusNotFound)
			return
		}
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
	return http.HandlerFunc(fn)
}

func handleGetPing() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if config.Config.Database == "" {
			http.Error(w, "Database not configured", http.StatusInternalServerError)
			return
		}

		check := database.CheckConnectingDataBase()
		if !check {
			http.Error(w, "Database connection failed", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
	return http.HandlerFunc(fn)
}

func TestRouter(db *sql.DB) chi.Router {
	r := chi.NewRouter()
	r.Use(chiMiddleware.CleanPath)
	r.Use(customMiddleware.UngzipMiddleware)
	r.Use(customMiddleware.GzipHandle)
	r.Post("/", customMiddleware.WithLogging(HandlePost(db)))
	r.Post("/api/shorten", customMiddleware.WithLogging(handleShortenPost(db)))
	r.Get("/{id}", customMiddleware.WithLogging(HandleGet(db)))
	r.Get("/ping", customMiddleware.WithLogging(handleGetPing()))
	return r
}

func isValidURL(rawURL string) bool {
	_, err := url.ParseRequestURI(rawURL)
	return err == nil
}
