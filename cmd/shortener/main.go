package main

import (
	"net/http"
	"romadmit993/GoShort/internal/config"
	"romadmit993/GoShort/internal/handlers"
	"romadmit993/GoShort/internal/storage"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	storage.Sugar = *logger.Sugar()

	conf := config.New()
	//	config.ParseFlags()
	storage.Sugar.Infow("Сервер запущен", "address", conf.LocalServer)
	if err := http.ListenAndServe(conf.LocalServer, handlers.TestRouter()); err != nil {
		storage.Sugar.Fatalf(err.Error(), "Ошибка при запуске сервера")
	}
}
