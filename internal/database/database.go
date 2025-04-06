package database

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type ShortURL struct {
	ShortURL    string
	OriginalURL string
}

func CheckConnectingDataBase(db *sql.DB) bool {
	if err := db.Ping(); err != nil {
		log.Printf("Ping error: %v", err)
		return false
	}
	return true
}

func SaveDataBase(db *sql.DB, shortURL, originalURL string) error {
	query := `
        INSERT INTO shorturl (shorturl, originalurl)
        VALUES ($1, $2)
        ON CONFLICT (originalurl) DO UPDATE 
        SET shorturl = EXCLUDED.shorturl
        RETURNING shorturl
    `

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result string
	err := db.QueryRowContext(ctx, query, shortURL, originalURL).Scan(&result)
	if err != nil {
		// Обрабатываем случай, когда запись уже существует
		if err == sql.ErrNoRows {
			return nil
		}
		log.Printf("Insert error: %v", err)
		return err
	}
	return nil
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
