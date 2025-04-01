package main

import (
	"flag"
	"log"

	"strings"

	"github.com/caarlos0/env/v6"
)

var Config struct {
	localServer string
	baseAddress string
	fileStorage string
	database    string
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
	flag.StringVar(&Config.localServer, "a", "localhost:8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&Config.baseAddress, "b", "http://localhost:8080/", "базовый адрес сокращённого URL")
	flag.StringVar(&Config.fileStorage, "f", "/tmp/short-url-db.json", "путь к файлу для хранения данных")
	flag.StringVar(&Config.database, "d", "", "подключение к базе данных")
	flag.Parse()

	if cfg.ServerAddress != "" {
		Config.localServer = cfg.ServerAddress
	}
	if cfg.BaseURL != "" {
		Config.baseAddress = cfg.BaseURL
	}
	if cfg.FileStorage != "" {
		Config.fileStorage = cfg.FileStorage
	}
	if cfg.DataBase != "" {
		Config.database = cfg.DataBase
	}
	if !strings.HasSuffix(Config.baseAddress, "/") {
		Config.baseAddress += "/"
	}
}
