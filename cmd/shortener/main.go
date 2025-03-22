package main

import (
	//	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"romadmit993/GoShort/internal/models"

	//	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

var (
	urlStore = make(map[string]string)
	storeMux sync.RWMutex
	sugar    zap.SugaredLogger
	r        = rand.New(rand.NewSource(time.Now().UnixNano()))
)

const (
	charset       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	shortIDLength = 6
)

func generateShortID() string {
	shortID := make([]byte, shortIDLength)
	for i := range shortID {
		shortID[i] = charset[r.Intn(len(charset))]
	}
	return string(shortID)
}

func isValidURL(rawURL string) bool {
	_, err := url.ParseRequestURI(rawURL)
	return err == nil
}

func handlePost() http.HandlerFunc {
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

		shortID := generateShortID()
		storeMux.Lock()
		urlStore[shortID] = originalURL
		storeMux.Unlock()

		shortURL := fmt.Sprintf("%s%s", Config.baseAddress, shortID)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, shortURL)
	}
	return http.HandlerFunc(fn)
}

func handleShortenPost() http.HandlerFunc {
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
		shortID := generateShortID()
		storeMux.Lock()
		urlStore[shortID] = apiShorten.URL
		storeMux.Unlock()
		shortURL := fmt.Sprintf("%s/%s", Config.baseAddress, shortID)
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

func handleGet() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "ID не может быть пустым", http.StatusBadRequest)
			return
		}
		storeMux.RLock()
		originalURL, exists := urlStore[id]
		storeMux.RUnlock()
		if !exists {
			http.Error(w, "Сокращённый URL не найден", http.StatusNotFound)
			return
		}
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
	return http.HandlerFunc(fn)
}

func testRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.CleanPath)
	//	r.Use(ungzipMiddleware) // Добавляем middleware для распаковки
	//	r.Use(gzipMiddleware)   // Добавляем middleware для сжатия
	r.Post("/", withLogging(handlePost()))
	r.Post("/api/shorten", withLogging(handleShortenPost()))
	r.Get("/{id}", withLogging(handleGet()))
	return r
}

func main() {

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()
	sugar = *logger.Sugar()

	ParseFlags()
	sugar.Infow("Сервер запущен", "address", Config.localServer)

	if err := http.ListenAndServe(Config.localServer, testRouter()); err != nil {
		sugar.Fatalf(err.Error(), "Ошибка при запуске сервера")
	}
}

func withLogging(h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
	return http.HandlerFunc(logFn)
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// func ungzipMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
// 			gz, err := gzip.NewReader(r.Body)
// 			if err != nil {
// 				http.Error(w, "Failed to decompress request", http.StatusBadRequest)
// 				return
// 			}
// 			defer gz.Close()
// 			r.Body = gz
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }
// func gzipMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Проверяем поддержку gzip
// 		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
// 			next.ServeHTTP(w, r)
// 			return
// 		}

// 		// Создаем обертку для ResponseWriter
// 		gzw := gzipResponseWriter{
// 			ResponseWriter: w,
// 			gzipWriter:     gzip.NewWriter(w),
// 		}
// 		defer gzw.gzipWriter.Close()

// 		// Устанавливаем заголовки
// 		w.Header().Set("Content-Encoding", "gzip")
// 		next.ServeHTTP(gzw, r)
// 	})
// }

// type gzipResponseWriter struct {
// 	http.ResponseWriter
// 	gzipWriter *gzip.Writer
// }

// func (w gzipResponseWriter) Write(b []byte) (int, error) {
// 	// Сжимаем данные только для допустимых типов контента
// 	contentType := w.Header().Get("Content-Type")
// 	if strings.Contains(contentType, "application/json") ||
// 		strings.Contains(contentType, "text/html") {
// 		return w.gzipWriter.Write(b)
// 	}
// 	return w.ResponseWriter.Write(b)
// }
