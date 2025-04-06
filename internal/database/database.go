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

	// if err := db.Ping(); err != nil {
	// 	log.Printf("Ping error: %v", err)
	// 	return false
	// }

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
        ON CONFLICT (originalurl) DO UPDATE 
        SET shorturl = EXCLUDED.shorturl
        RETURNING shorturl
    `

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result string
	err := db.QueryRowContext(ctx, query, shortURL, originalURL).Scan(&result)
	if err != nil {
		log.Printf("Insert error: %v", err)
		return err
	}

	if result != shortURL {
		//return fmt.Errorf("short URL mismatch")
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
