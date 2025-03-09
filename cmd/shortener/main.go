package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"

	"math/rand"
	"time"

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

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func generateShortID() string {
	shortID := make([]byte, shortIDLength)
	for i := range shortID {
		shortID[i] = charset[r.Intn(len(charset))]
	}
	return string(shortID)
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
