package main

import (
	"context"
	"net/http"
	"romadmit993/GoShort/internal/config"
	"romadmit993/GoShort/internal/handlers"
	"romadmit993/GoShort/internal/storage"

	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	config.ParseFlags()
	var db *sql.DB
	var err error
	if config.Config.Database != "" {
		db, err = sql.Open("pgx", config.Config.Database)
		if err != nil {
			log.Fatal("Connection error:", err)
		}
		defer db.Close()

		if !initializeDatabase(db) {
			log.Fatal("Database initialization failed")
		}
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	storage.Sugar = *logger.Sugar()
	storage.Sugar.Infow("Сервер запущен", "address", config.Config.LocalServer)
	if err := http.ListenAndServe(config.Config.LocalServer, handlers.TestRouter(db)); err != nil {
		storage.Sugar.Fatalf(err.Error(), "Ошибка при запуске сервера")
	}
}
func initializeDatabase(db *sql.DB) bool {
	createTableSQL := `
        CREATE TABLE IF NOT EXISTS shorturl (
            uuid SERIAL PRIMARY KEY,
            shorturl TEXT UNIQUE NOT NULL,
            originalurl TEXT UNIQUE NOT NULL
        );
		CREATE INDEX IF NOT EXISTS idx_originalurl ON shorturl (originalurl);
		`

	_, err := db.ExecContext(context.Background(), createTableSQL)
	if err != nil {
		log.Printf("Table creation error: %v", err)
		return false
	}
	return true
}
