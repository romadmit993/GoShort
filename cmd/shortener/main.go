package main

import (
	"net/http"
	"romadmit993/GoShort/internal/config"
	"romadmit993/GoShort/internal/database"
	"romadmit993/GoShort/internal/handlers"
	"romadmit993/GoShort/internal/storage"

	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	db, err := sql.Open("pgx", config.Config.Database)
	if err != nil {
		log.Printf("Connection error: %v", err)
	}
	// Добавьте эту проверку!
	if config.Config.Database != "" {
		if !database.CheckConnectingDataBase() {
			log.Printf("Connection error: %v", err)
		}
	}
	//defer db.Close()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	storage.Sugar = *logger.Sugar()
	config.ParseFlags()
	storage.Sugar.Infow("Сервер запущен", "address", config.Config.LocalServer)
	if err := http.ListenAndServe(config.Config.LocalServer, handlers.TestRouter(db)); err != nil {
		storage.Sugar.Fatalf(err.Error(), "Ошибка при запуске сервера")
	}
}
