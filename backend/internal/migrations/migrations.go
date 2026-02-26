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
	log := logger.With(
		zap.String("component", "repository"),
		zap.String("operation", "run_migrations"),
	)

	log.Info("Migrations started")

	d, err := iofs.New(migrationsDir, "migration_files")
	if err != nil {
		log.Error("Migrations failed", zap.Error(err))
		return fmt.Errorf("open migrations source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		log.Error("Migrations failed", zap.Error(err))
		return fmt.Errorf("create migrator: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Error("Migrator close failed", zap.Error(srcErr))
		}
		if dbErr != nil {
			log.Error("Migrator close failed", zap.Error(dbErr))
		}
	}()

	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Error("Migrations failed", zap.Error(err))
			return fmt.Errorf("apply migrations: %w", err)
		}

		log.Info("Migrations up to date")

		return nil
	}

	log.Info("Migrations applied")

	return nil
}
