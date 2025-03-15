package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
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

func handleGet() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
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
	r.Post("/", withLogging(handlePost()))
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
	sugar.Infow(
		"Сервер запущен на", Config.localServer,
	)

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
