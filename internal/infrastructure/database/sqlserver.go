package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

func ConnectSQLServer(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("sqlserver", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("opening sqlserver connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.PingTimeout)*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pinging sqlserver: %w", err)
	}

	return db, nil
}
