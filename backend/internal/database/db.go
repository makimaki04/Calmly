package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/makimaki04/Calmly/internal/config"
	"github.com/makimaki04/Calmly/internal/migrations"
	"go.uber.org/zap"
)

func InitDB(ctx context.Context, dsn string, limits config.DBLimits, logger *zap.Logger) (*sql.DB, error) {
	log := logger.With(
		zap.String("component", "repository"),
		zap.String("operation", "init_db"),
	)

	log.Info("DB init started")

	if err := migrations.RunMigrations(dsn, logger); err != nil {
		// Error is logged inside migrations. Avoid duplicates here.
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Error("DB open failed", zap.Error(err))
		return nil, fmt.Errorf("open db connection: %w", err)
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(dbCtx); err != nil {
		log.Error("DB ping failed", zap.Error(err))
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	db.SetMaxOpenConns(limits.MaxOpenConns)
	db.SetMaxIdleConns(limits.MaxIdleConns)
	db.SetConnMaxLifetime(limits.MaxLifeTime)
	db.SetConnMaxIdleTime(limits.MaxIdleTime)

	log.Info("DB connected")

	return db, nil
}
