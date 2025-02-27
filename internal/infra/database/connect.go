package database

import (
	"fmt"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/mathcale/go-api-boilerplate/internal/pkg/apperror"
	"github.com/mathcale/go-api-boilerplate/internal/pkg/logger"
)

type Database interface {
	Connect() (*sqlx.DB, error)
}

type database struct {
	logger              logger.Logger
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
	logger logger.Logger,
	host, user, password, name, sslMode string,
	port, maxOpenConns, maxIdleConns, connMaxLifetimeSecs, connMaxIdleTimeSecs int,
) Database {
	return &database{
		logger:              logger,
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

func (d *database) Connect() (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		d.host, d.user, d.password, d.name, d.port, d.sslMode,
	)

	dbx, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, apperror.New(
			err, "database connection error", apperror.DatabaseKind,
			apperror.RepositoryOrigin, "database/connect", nil, nil,
		)
	}

	dbx.SetMaxOpenConns(d.maxOpenConns)
	dbx.SetMaxIdleConns(d.maxIdleConns)
	dbx.SetConnMaxLifetime(time.Duration(d.connMaxLifetimeSecs) * time.Second)
	dbx.SetConnMaxIdleTime(time.Duration(d.connMaxIdleTimeSecs) * time.Second)

	return dbx, nil
}
