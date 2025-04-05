package database

import (
	"database/sql"
	"romadmit993/GoShort/internal/config"
)

func CheckConnectingDataBase() bool {
	db, err := sql.Open("pgx", config.Config.Database)
	if err != nil {
		return false
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return false
	}
	return true
}
