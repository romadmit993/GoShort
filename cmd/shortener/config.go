package main

import (
	"flag"
	"log"

	"strings"

	"github.com/caarlos0/env/v6"
)

// Config хранит конфигурацию приложения.
var Config struct {
	localServer string
	baseAddress string
}

// EnviromentVariables представляет переменные окружения.
type EnviromentVariables struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

var cfg EnviromentVariables

// ParseFlags загружает конфигурацию из переменных окружения и флагов.
func ParseFlags() {
	// Загружаем переменные окружения.
	if err := env.Parse(&cfg); err != nil {
		log.Printf("Ошибка при парсинге переменных окружения: %s", err)
	}

	// Устанавливаем значения по умолчанию для флагов.
	flag.StringVar(&Config.localServer, "a", "localhost:8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&Config.baseAddress, "b", "http://localhost:8080/", "базовый адрес сокращённого URL")
	flag.Parse()

	// Приоритет: переменные окружения > флаги.
	if cfg.ServerAddress != "" {
		Config.localServer = cfg.ServerAddress
	}
	if cfg.BaseURL != "" {
		Config.baseAddress = cfg.BaseURL
	}

	// Убедимся, что BaseAddress заканчивается на "/".
	if !strings.HasSuffix(Config.baseAddress, "/") {
		Config.baseAddress += "/"
	}
}
