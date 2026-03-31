package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"localaihub/localaihub_go/internal/config"
)

func NewMySQL(ctx context.Context, cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	db.SetMaxOpenConns(int(cfg.MaxOpenConns))
	db.SetMaxIdleConns(int(cfg.MinIdleConns))
	db.SetConnMaxLifetime(time.Duration(cfg.MaxConnLifetime) * time.Minute)
	db.SetConnMaxIdleTime(time.Duration(cfg.MaxConnIdleTime) * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return db, nil
}
