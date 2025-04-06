package config

import (
	"flag"
	"log"
	"strings"

	"github.com/caarlos0/env/v6"
)

var Config struct {
	LocalServer string
	BaseAddress string
	FileStorage string
	Database    string
}

type EnviromentVariables struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	FileStorage   string `env:"FILE_STORAGE_PATH"`
	DataBase      string `env:"DATABASE_DSN"`
}

var cfg EnviromentVariables

func ParseFlags() {
	if err := env.Parse(&cfg); err != nil {
		log.Printf("Ошибка при парсинге переменных окружения: %s", err)
	}
	flag.StringVar(&Config.LocalServer, "a", "localhost:8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&Config.BaseAddress, "b", "http://localhost:8080/", "базовый адрес сокращённого URL")
	flag.StringVar(&Config.FileStorage, "f", "/tmp/short-url-db.json", "путь к файлу для хранения данных")
	flag.StringVar(&Config.Database, "d", "", "подключение к базе данных")
	flag.Parse()

	if cfg.ServerAddress != "" {
		Config.LocalServer = cfg.ServerAddress
	}
	if cfg.BaseURL != "" {
		Config.BaseAddress = cfg.BaseURL
	}
	if cfg.FileStorage != "" {
		Config.FileStorage = cfg.FileStorage
	}
	if cfg.DataBase != "" {
		Config.Database = cfg.DataBase
	}
	if !strings.HasSuffix(Config.BaseAddress, "/") {
		Config.BaseAddress += "/"
	}
}
