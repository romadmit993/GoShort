package storage

import (
	"bufio"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"romadmit993/GoShort/internal/config"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

type shortenerURLFile struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
type Handler struct {
	cfg *config.Config
}

var (
	URLStore = make(map[string]string)
	StoreMux sync.RWMutex
	Sugar    zap.SugaredLogger
	R        = rand.New(rand.NewSource(time.Now().UnixNano()))
)

const (
	charset       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	shortIDLength = 6
)

func GenerateShortID() string {
	shortID := make([]byte, shortIDLength)
	for i := range shortID {
		shortID[i] = charset[R.Intn(len(charset))]
	}
	return string(shortID)
}

func SaveShortURLFile(shortID string, url string) {
	fileStorage := Handler{}
	if fileStorage.cfg.FileStorage == "" {
		log.Printf("Путь к файлу не задан")
		return
	}

	dir := filepath.Dir(fileStorage.cfg.FileStorage)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Ошибка создания директории: %v", err)
		return
	}
	uuid, _ := ReadFileAndCheckID("")
	record := shortenerURLFile{
		UUID:        strconv.Itoa(uuid),
		ShortURL:    shortID,
		OriginalURL: url,
	}
	jsonData, err := json.Marshal(record)
	if err != nil {
		log.Printf("Ошибка при кодировании в JSON: %v", err)
	}
	jsonData = append(jsonData, '\n')

	file, err := os.OpenFile(fileStorage.cfg.FileStorage, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Ошибка при создании файла: %v", err)
	}
	defer file.Close()
	_, err = file.Write(jsonData)
	if err != nil {
		log.Printf("Ошибка при записи в файл: %v", err)
	}
}

func ReadFileAndCheckID(id string) (int, bool) {
	fileStorage := Handler{}
	file, err := os.Open(fileStorage.cfg.FileStorage)
	if err != nil {
		return 1, false // Если файл не найден, считаем что записей нет
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var record shortenerURLFile
	count := 1
	exists := false

	for scanner.Scan() {
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			continue // Пропускаем битые записи, но продолжаем подсчет
		}
		if record.ShortURL == id {
			exists = true
		}
		count++
	}

	return count, exists
}
