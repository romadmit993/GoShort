package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"romadmit993/GoShort/internal/models"

	"strings"
	"sync"
	"time"

	"bufio"
	"log"
	"os"
	"strconv"

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
	gzipWriter struct {
		http.ResponseWriter
		Writer io.Writer
	}
	shortenerUrlFile struct {
		Uuid         string `json:"uuid"`
		Short_url    string `json:"short_url"`
		Original_url string `json:"original_url"`
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

func readFile() int {
	var count int = 1
	file, err := os.Open(Config.fileStorage)
	if err != nil {
		return count
	}
	defer file.Close()

	// Читаем файл построчно
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		count += 1
	}
	return count
}
func readCheckFile(id string) bool {
	var check bool
	check = false
	file, err := os.Open(Config.fileStorage)
	if err != nil {
		return check
	}
	defer file.Close()

	// Читаем файл построчно
	scanner := bufio.NewScanner(file)
	var record shortenerUrlFile
	for scanner.Scan() {
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			continue
		}
		if id == record.Short_url {
			check = true
			break
		}
	}
	return check
}
func saveShortUrlFile(shortID string, url string) {
	uuid := readFile()
	record := shortenerUrlFile{
		Uuid:         strconv.Itoa(uuid),
		Short_url:    shortID,
		Original_url: url,
	}
	jsonData, err := json.Marshal(record)
	if err != nil {
		log.Fatalf("Ошибка при кодировании в JSON: %v", err)
	}
	jsonData = append(jsonData, '\n')
	// Открываем файл для записи
	file, err := os.OpenFile("data.json", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Ошибка при создании файла: %v", err)
	}
	defer file.Close()
	_, err = file.Write(jsonData)
	if err != nil {
		log.Fatalf("Ошибка при записи в файл: %v", err)
	}
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
		saveShortUrlFile(shortID, originalURL)
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
		saveShortUrlFile(shortID, apiShorten.URL)
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
		if !exists {
			exists = readCheckFile(id)
		}
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
	r.Use(ungzipMiddleware) // Добавляем middleware для распаковки
	r.Use(gzipHandle)       // Добавляем middleware для сжатия
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

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
func gzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := w.Header().Get("Content-Type")
		if strings.Contains(contentType, "application/json") ||
			strings.Contains(contentType, "text/html") {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				io.WriteString(w, err.Error())
				return
			}
			defer gz.Close()

			w.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func ungzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decompress request", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz
		}
		next.ServeHTTP(w, r)
	})
}
