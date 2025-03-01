package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	//	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
)

var Confing struct {
	localServer string
	baseAddress string
}

type EnviromentVariables struct {
	server_address string `env:"SERVER_ADDRESS"`
	base_url       string `env:"BASE_URL"`
}

var urlStore = map[string]string{}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const shortIDLength = 6

// Функция для парсинга флагов
func ParseFlags() {
	//	var cfg EnviromentVariables
	//	err := env.Parse(&cfg)
	//	if err == nil {
	//		Confing.localServer = cfg.SERVER_ADDRESS
	//		Confing.baseAddress = cfg.BASE_URL
	//	} else {

	flag.StringVar(&Confing.localServer, "a", "localhost:8080", "start server")
	flag.StringVar(&Confing.baseAddress, "b", "http://localhost:8080/", "shorter URL")
	flag.Parse()

	//	}
	// Убедимся, что baseAddress заканчивается на "/"
	if !strings.HasSuffix(Confing.baseAddress, "/") {
		Confing.baseAddress += "/"
	}

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
	shortURL := fmt.Sprintf("%s%s", Confing.baseAddress, shortID) // Корректный URL

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
	http.ListenAndServe(Confing.localServer, testRouter())
}
