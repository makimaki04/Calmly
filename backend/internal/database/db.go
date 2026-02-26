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
	logger.Info("DB init started",
		zap.String("component", "repository"),
		zap.String("operation", "init_db"),
	)

	if err := migrations.RunMigrations(dsn, logger); err != nil {
		// Error is logged inside migrations. Avoid duplicates here.
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Error("DB init failed",
			zap.String("component", "repository"),
			zap.String("operation", "init_db"),
			zap.Error(err),
		)
		return nil, fmt.Errorf("open db connection: %w", err)
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(dbCtx); err != nil {
		logger.Error("DB init failed",
			zap.String("component", "repository"),
			zap.String("operation", "init_db"),
			zap.Error(err),
		)
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	db.SetMaxOpenConns(limits.MaxOpenConns)
	db.SetMaxIdleConns(limits.MaxIdleConns)
	db.SetConnMaxLifetime(limits.MaxLifeTime)
	db.SetConnMaxIdleTime(limits.MaxIdleTime)

	logger.Info("DB connected",
		zap.String("component", "repository"),
		zap.String("operation", "init_db"),
	)

	return db, nil
}
