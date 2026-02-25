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

func InitDB(dsn string, logger *zap.Logger) (*sql.DB, error) {
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

	crx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.PingContext(crx); err != nil {
		logger.Error("DB init failed",
			zap.String("component", "repository"),
			zap.String("operation", "init_db"),
			zap.Error(err),
		)
		return nil, fmt.Errorf("db connection error: %w", err)
	}

	return db, nil
}
