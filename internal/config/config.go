package config

import (
	"flag"
	"os"

	//	"log"

	"strings"
	//	"github.com/caarlos0/env/v6"
)

type Config struct {
	LocalServer string
	BaseAddress string
	FileStorage string
	Database    string
}

func New() *Config {

	a := flag.String("a", "localhost:8080", "адрес запуска HTTP-сервера")
	b := flag.String("b", "http://localhost:8080/", "базовый адрес сокращённого URL")
	f := flag.String("f", "/tmp/short-url-db.json", "путь к файлу для хранения данных")
	d := flag.String("d", "", "подключение к базе данных")
	flag.Parse()

	config := &Config{
		LocalServer: *a,
		BaseAddress: *b,
		FileStorage: *f,
		Database:    *d,
	}
	config.envVar()
	config.ensureTrailingSlash()
	return config
}

func (c *Config) envVar() {
	if serverAddress := getEnv("SERVER_ADDRESS", ""); serverAddress != "" {
		c.LocalServer = serverAddress
	}
	if baseURL := getEnv("BASE_URL", ""); baseURL != "" {
		c.BaseAddress = baseURL
	}
	if fileStorage := getEnv("FILE_STORAGE_PATH", ""); fileStorage != "" {
		c.FileStorage = fileStorage
	}
	if database := getEnv("DATABASE_DSN", ""); database != "" {
		c.Database = database
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func (c *Config) ensureTrailingSlash() {
	if !strings.HasSuffix(c.BaseAddress, "/") {
		c.BaseAddress += "/"
	}
}

// var Config struct {
// 	LocalServer string
// 	BaseAddress string
// 	FileStorage string
// 	Database    string
// }

// type EnviromentVariables struct {
// 	ServerAddress string `env:"SERVER_ADDRESS"`
// 	BaseURL       string `env:"BASE_URL"`
// 	FileStorage   string `env:"FILE_STORAGE_PATH"`
// 	DataBase      string `env:"DATABASE_DSN"`
// }

// var cfg EnviromentVariables

// func ParseFlags() {
// 	if err := env.Parse(&cfg); err != nil {
// 		log.Printf("Ошибка при парсинге переменных окружения: %s", err)
// 	}
// 	flag.StringVar(&Config.LocalServer, "a", "localhost:8080", "адрес запуска HTTP-сервера")
// 	flag.StringVar(&Config.BaseAddress, "b", "http://localhost:8080/", "базовый адрес сокращённого URL")
// 	flag.StringVar(&Config.FileStorage, "f", "/tmp/short-url-db.json", "путь к файлу для хранения данных")
// 	flag.StringVar(&Config.Database, "d", "", "подключение к базе данных")
// 	flag.Parse()

// 	if cfg.ServerAddress != "" {
// 		Config.LocalServer = cfg.ServerAddress
// 	}
// 	if cfg.BaseURL != "" {
// 		Config.BaseAddress = cfg.BaseURL
// 	}
// 	if cfg.FileStorage != "" {
// 		Config.FileStorage = cfg.FileStorage
// 	}
// 	if cfg.DataBase != "" {
// 		Config.Database = cfg.DataBase
// 	}
// 	if !strings.HasSuffix(Config.BaseAddress, "/") {
// 		Config.BaseAddress += "/"
// 	}
// }
