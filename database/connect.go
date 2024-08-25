package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Database interface {
	Connect() (*sql.DB, error)
}

type database struct {
	host                string
	port                int
	user                string
	password            string
	name                string
	sslMode             string
	maxOpenConns        int
	maxIdleConns        int
	connMaxLifetimeSecs int
	connMaxIdleTimeSecs int
}

func NewDatabase(
	host, user, password, name, sslMode string,
	port, maxOpenConns, maxIdleConns, connMaxLifetimeSecs, connMaxIdleTimeSecs int,
) Database {
	return &database{
		host:                host,
		port:                port,
		user:                user,
		password:            password,
		name:                name,
		sslMode:             sslMode,
		maxOpenConns:        maxOpenConns,
		maxIdleConns:        maxIdleConns,
		connMaxLifetimeSecs: connMaxLifetimeSecs,
		connMaxIdleTimeSecs: connMaxIdleTimeSecs,
	}
}

func (d *database) Connect() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.host, d.port, d.user, d.password, d.name, d.sslMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(d.maxOpenConns)
	db.SetMaxIdleConns(d.maxIdleConns)
	db.SetConnMaxLifetime(time.Duration(d.connMaxLifetimeSecs) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(d.connMaxIdleTimeSecs) * time.Second)

	return db, nil
}
