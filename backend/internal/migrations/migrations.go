package migrations

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"
)

//go:embed migration_files/*.sql
var migrationsDir embed.FS

func RunMigrations(dsn string, logger *zap.Logger) error {
	logger.Info("Migrations started",
		zap.String("component", "repository"),
		zap.String("operation", "run_migrations"),
	)

	d, err := iofs.New(migrationsDir, "migration_files")
	if err != nil {
		logger.Error("Migrations failed",
			zap.String("component", "repository"),
			zap.String("operation", "run_migrations"),
			zap.Error(err),
		)
		return fmt.Errorf("open migrations source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		logger.Error("Migrations failed",
			zap.String("component", "repository"),
			zap.String("operation", "run_migrations"),
			zap.Error(err),
		)
		return fmt.Errorf("create migrator: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			logger.Error("Migrations close failed",
				zap.String("component", "repository"),
				zap.String("operation", "close_migrations"),
				zap.Error(srcErr),
			)
		}

		if dbErr != nil {
			logger.Error("Migrations close failed",
				zap.String("component", "repository"),
				zap.String("operation", "close_migrations"),
				zap.Error(dbErr),
			)
		}
	}()

	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			logger.Error("Migrations failed",
				zap.String("component", "repository"),
				zap.String("operation", "run_migrations"),
				zap.Error(err),
			)
			return fmt.Errorf("apply migrations: %w", err)
		}

		logger.Info("Migrations up to date",
			zap.String("component", "repository"),
			zap.String("operation", "run_migrations"),
		)

		return nil
	}

	logger.Info("Migrations applied",
		zap.String("component", "repository"),
		zap.String("operation", "run_migrations"),
	)

	return nil
}
