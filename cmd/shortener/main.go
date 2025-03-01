package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

var Confing struct {
	localServer string
	baseAddress string
}

var urlStore = map[string]string{}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const shortIDLength = 6

// Функция для парсинга флагов
func ParseFlags() {
	flag.StringVar(&Confing.localServer, "a", "localhost:8080", "start server")
	flag.StringVar(&Confing.baseAddress, "b", "http://localhost:8080/", "shorter URL")
	flag.Parse()
}

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
	shortURL := fmt.Sprintf("%s%s", Confing.baseAddress, shortID)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortURL)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/")
	originalURL, exists := urlStore[id]
	if !exists {
		http.Error(w, "Сокращённый URL не найден", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handlePost(w, r)
	case http.MethodGet:
		handleGet(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}
func testRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", handle)
	return r
}
func main() {
	ParseFlags()
	http.ListenAndServe(Confing.localServer, testRouter())
}
