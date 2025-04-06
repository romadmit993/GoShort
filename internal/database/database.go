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
func SaveDataBase(db *sql.DB, shortURL, originalURL string) error {
	query := `
        INSERT INTO shorturl (shorturl, originalurl)
        VALUES ($1, $2)
        ON CONFLICT (originalurl) DO NOTHING
    `

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, shortURL, originalURL)
	if err != nil {
		log.Printf("Insert error: %v", err)
		return err
	}
	return nil
}

func СheckRecord(db *sql.DB, originalURL string) bool {
	row := db.QueryRowContext(
		context.Background(),
		"SELECT originalurl FROM shorturl WHERE originalurl = $1",
		originalURL,
	)

	var result string
	if err := row.Scan(&result); err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Check record error: %v", err)
		}
		return false
	}
	return true
}
func CheckOriginalURLExists(db *sql.DB, originalURL string) bool {
	row := db.QueryRowContext(
		context.Background(),
		"SELECT originalurl FROM shorturl WHERE originalurl = $1",
		originalURL,
	)

	var result string
	err := row.Scan(&result)
	return err == nil
}
