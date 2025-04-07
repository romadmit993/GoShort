package database

import (
	"context"
	"database/sql"
	"log"
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

func SaveDataBase(db *sql.DB, shortURL, originalURL string) string {
	checkWrite, record := CheckOriginalURLExists(db, originalURL)
	if checkWrite {
		return record
	}
	query := `INSERT INTO shorturl (shorturl, originalurl) VALUES ($1, $2)`
	db.QueryRowContext(context.Background(), query, shortURL, originalURL)
	return ""
}

func CheckOriginalURLExists(db *sql.DB, originalURL string) (check bool, record string) {
	row := db.QueryRowContext(
		context.Background(),
		"SELECT originalurl FROM shorturl WHERE originalurl = $1",
		originalURL,
	)

	var result string
	err := row.Scan(&result)
	if result != "" {
		return true, result
	}
	return err == nil, ""
}
