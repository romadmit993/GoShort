package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

var urlStore = map[string]string{}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const shortIDLength = 6

func generateShortID() string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	shortID := make([]byte, shortIDLength)
	for i := range shortID {
		shortID[i] = charset[r.Intn(len(charset))]
	}
	return string(shortID)
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения тела запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := string(body)
	fmt.Println("Получен URL:", originalURL)

	shortID := generateShortID()
	urlStore[shortID] = originalURL
	shortURL := fmt.Sprintf("%s%s", Config.baseAddress, shortID) // Корректный URL

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortURL)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id") // Извлекаем ID из URL
	originalURL, exists := urlStore[id]
	if !exists {
		http.Error(w, "Сокращённый URL не найден", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func testRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/", handlePost)   // Обработка POST-запросов
	r.Get("/{id}", handleGet) // Обработка GET-запросов для сокращённых URL
	return r
}

func main() {
	ParseFlags()
	http.ListenAndServe(Config.localServer, testRouter())
}
