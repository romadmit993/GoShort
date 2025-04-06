package database

import (
	"context"
	"database/sql"
	"log"
	"romadmit993/GoShort/internal/config"
	"time"
)

type ShortURL struct {
	ShortURL    string
	OriginalURL string
}

func CheckConnectingDataBase() bool {
	db, err := sql.Open("pgx", config.Config.Database)
	if err != nil {
		log.Printf("Connection error: %v", err)
		return false
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Printf("Ping error: %v", err)
		return false
	}

	createTableSQL := `
        CREATE TABLE IF NOT EXISTS shorturl (
			uuid SERIAL PRIMARY KEY,
            shorturl TEXT UNIQUE NOT NULL,
            originalurl TEXT UNIQUE NOT NULL
        )`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, createTableSQL); err != nil {
		log.Printf("Table creation error: %v", err)
		return false
	}

	return true
}
func SaveDataBase(db *sql.DB, shortURL, originalURL string) {
	createTableSQL := `
        CREATE TABLE IF NOT EXISTS shorturl (
			uuid SERIAL PRIMARY KEY,
            shorturl TEXT UNIQUE NOT NULL,
            originalurl TEXT UNIQUE NOT NULL
        )`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, createTableSQL); err != nil {
		log.Printf("Table creation error: %v", err)
	}

	query := `
        INSERT INTO shorturl (shorturl, originalurl)
        VALUES ($1, $2)
    `

	db.QueryRowContext(ctx, query, shortURL, originalURL)
}

func СheckRecord(db *sql.DB, shortURLL string) bool {
	row := db.QueryRowContext(context.Background(),
		"SELECT shortURL FROM shorturl WHERE shortURL = $1, shortURLL")

	var (
		shortURL string
	)
	err := row.Scan(&shortURL)
	if err != nil {
		log.Printf("Ошибка в выборке: %v", err)
		return false
	}
	if shortURL == "" {
		log.Printf("Нет значений")
		return false
	}
	return true
}
