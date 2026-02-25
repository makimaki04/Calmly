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
	d, err := iofs.New(migrationsDir, "migration_files")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	defer func ()  {
		srcErr, dbErr := m.Close()
		if srcErr != nil {

		}

		if dbErr != nil {

		}
	}()

	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("load migrations error: %w", err)
		}
	}

	return nil
}