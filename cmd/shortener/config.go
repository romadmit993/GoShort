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
}

type EnviromentVariables struct {
	serverAddress string `env:"SERVER_ADDRESS"`
	baseUrl       string `env:"BASE_URL"`
}

var cfg EnviromentVariables

func ParseFlags() {

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Ошибка при парсинге переменных окружения: %s", err)
	}

	flag.StringVar(&Config.localServer, "a", "localhost:8080", "start server")
	flag.StringVar(&Config.baseAddress, "b", "http://localhost:8080/", "shorter URL")
	flag.Parse()

	if !strings.HasSuffix(Config.baseAddress, "/") {
		Config.baseAddress += "/"
	}

}
