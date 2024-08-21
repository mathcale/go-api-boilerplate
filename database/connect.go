package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Database interface {
	Connect() (*sql.DB, error)
}

func Connect(host, user, password, name, sslMode string, port int64) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=%s", host, user, password, name, sslMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
