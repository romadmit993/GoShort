package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-chi/chi/v5"
)

var (
	urlStore = make(map[string]string)
	storeMux sync.RWMutex
)

const (
	charset       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	shortIDLength = 6
)

func generateShortID() string {
	b := make([]byte, shortIDLength)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

// isValidURL проверяет, является ли строка корректным URL.
func isValidURL(rawURL string) bool {
	_, err := url.ParseRequestURI(rawURL)
	return err == nil
}

func handlePost(w http.ResponseWriter, r *http.Request) {
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

func handleGet(w http.ResponseWriter, r *http.Request) {
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

func testRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/", handlePost)
	r.Get("/{id}", handleGet)
	return r
}

func main() {
	ParseFlags()
	log.Printf("Сервер запущен на %s", Config.localServer)
	if err := http.ListenAndServe(Config.localServer, testRouter()); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %s", err)
	}
}
