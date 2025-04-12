package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
			log.Printf("Первая итерация HandlePost shortID %s", shortID)
			log.Printf("Первая итерация HandlePost originalURL %s", originalURL)
			rewrite := database.SaveDataBase(db, shortID, originalURL)
			if rewrite != "" {
				log.Printf("Запись есть в HandlePost shortID %s , rewrite: %s ", shortID, rewrite)
				log.Printf("Запись есть в HandlePost originalURL %s", originalURL)
				storage.StoreMux.Unlock()
				shortURL := fmt.Sprintf("%s%s", config.Config.BaseAddress, rewrite)
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusConflict)
				fmt.Fprint(w, shortURL)
				return
			}
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
			rewrite := database.SaveDataBase(db, shortID, apiShorten.URL)
			if rewrite != "" {
				storage.StoreMux.Unlock()
				shortURL := fmt.Sprintf("%s/%s", config.Config.BaseAddress, rewrite)
				response := models.Shorten{
					Result: shortURL,
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				if err := json.NewEncoder(w).Encode(response); err != nil {
					http.Error(w, "Ошибка при формировании ответа", http.StatusInternalServerError)
				}
				return
			}
		}
		storage.StoreMux.Unlock()
		log.Printf("test handleShortenPost shortID %s", shortID)
		log.Printf("test handleShortenPost apiShorten.URL %s", apiShorten.URL)
		log.Printf("test handleShortenPost config.Config.Database  %s", config.Config.Database)
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

func handleBatchPost(db *sql.DB) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var batch []models.BatchRequest
		if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		results := make([]models.BatchResponse, 0, len(batch))
		storage.StoreMux.Lock()
		defer storage.StoreMux.Unlock()

		var tx *sql.Tx
		var err error
		if config.Config.Database != "" {
			tx, err = db.Begin()
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			defer tx.Rollback()
		}

		for _, item := range batch {
			if !isValidURL(item.OriginalURL) {
				http.Error(w, "Invalid URL: "+item.OriginalURL, http.StatusBadRequest)
				return
			}

			shortID := storage.GenerateShortID()

			storage.URLStore[shortID] = item.OriginalURL
			storage.SaveShortURLFile(shortID, item.OriginalURL)

			if config.Config.Database != "" {
				_, err = tx.ExecContext(
					r.Context(),
					"INSERT INTO shorturl (shorturl, originalurl) VALUES ($1, $2)",
					shortID,
					item.OriginalURL,
				)
				if err != nil {
					http.Error(w, "Database insert error", http.StatusInternalServerError)
					return
				}
			}

			results = append(results, models.BatchResponse{
				CorrelationID: item.CorrelationID,
				ShortURL:      fmt.Sprintf("%s/%s", config.Config.BaseAddress, shortID),
			})
		}

		if config.Config.Database != "" {
			if err := tx.Commit(); err != nil {
				http.Error(w, "Database commit error", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(results)
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
		originalURL, exists := storage.URLStore[id]
		if !exists {
			_, exists = storage.ReadFileAndCheckID(id)
		}
		storage.StoreMux.RUnlock()
		if !exists {
			exists = existsDataBase
		}
		// 3. Проверяем в БД
		if !exists && config.Config.Database != "" {
			err := db.QueryRowContext(
				context.Background(),
				"SELECT originalurl FROM shorturl WHERE shorturl = $1",
				id,
			).Scan(&originalURL)

			exists = (err == nil)
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

func handleGetPing(db *sql.DB) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if config.Config.Database == "" {
			http.Error(w, "Database not configured", http.StatusInternalServerError)
			return
		}

		check := database.CheckConnectingDataBase(db)
		if !check {
			http.Error(w, "Database connection failed", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
	return http.HandlerFunc(fn)
}

func getUsersURL(db *sql.DB) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.QueryContext(context.Background(), "SELECT * from shorturl")
		if err != nil {
			log.Printf("Ошибка")
			return
		}
		defer rows.Close()
		for rows.Next() {
			var v models.AllRecord
			err = rows.Scan(&v.Shorturl, &v.Originalurl)
			if err != nil {
				return
			}
			log.Printf("Shorturl %s", v.Shorturl)
			log.Printf("Originalurl %s", v.Originalurl)
		}
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
	r.Post("/api/shorten/batch", customMiddleware.WithLogging(handleBatchPost(db)))
	r.Get("/{id}", customMiddleware.WithLogging(HandleGet(db)))
	r.Get("/ping", customMiddleware.WithLogging(handleGetPing(db)))
	r.Get("/api/user/urls", customMiddleware.WithLogging(getUsersURL(db)))
	return r
}

func isValidURL(rawURL string) bool {
	_, err := url.ParseRequestURI(rawURL)
	return err == nil
}
