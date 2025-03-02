package main

import (
	"flag"
	"strings"
)

var Confing struct {
	LocalServer string
	BaseAddress string
}

type EnviromentVariables struct {
	Serveraddress string `env:"SERVER_ADDRESS"`
	Baseurl       string `env:"BASE_URL"`
}

func ParseFlags() {
	//	var cfg EnviromentVariables
	//	err := env.Parse(&cfg)
	//	if err == nil {
	//		Confing.localServer = cfg.SERVER_ADDRESS
	//		Confing.baseAddress = cfg.BASE_URL
	//	} else {

	flag.StringVar(&Confing.LocalServer, "a", "localhost:8080", "start server")
	flag.StringVar(&Confing.BaseAddress, "b", "http://localhost:8080/", "shorter URL")
	flag.Parse()

	//	}
	// Убедимся, что baseAddress заканчивается на "/"
	if !strings.HasSuffix(Confing.BaseAddress, "/") {
		Confing.BaseAddress += "/"
	}

}
