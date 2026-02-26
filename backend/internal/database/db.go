package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/makimaki04/Calmly/internal/migrations"
	"go.uber.org/zap"
)

func InitDB(ctx context.Context, dsn string, logger *zap.Logger) (*sql.DB, error) {
	logger.Info("DB init started",
		zap.String("component", "repository"),
		zap.String("operation", "init_db"),
	)

	if err := migrations.RunMigrations(dsn, logger); err != nil {
		return nil, fmt.Errorf("run migration error: %w", err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Error("DB init failed",
			zap.String("component", "repository"),
			zap.String("operation", "init_db"),
			zap.Error(err),
		)
		return nil, fmt.Errorf("db open error: %w", err)
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
		return nil, fmt.Errorf("db connection error: %w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(12*time.Minute)
	db.SetConnMaxIdleTime(5*time.Minute)

	logger.Info("DB init succeeded",
		zap.String("component", "repository"),
		zap.String("operation", "init_db"),
	)

	return db, nil
}
