package main

import (
	"context"
	"log"
	"os"

	"github.com/makimaki04/Calmly/internal/database"
	"github.com/makimaki04/Calmly/internal/logger"
	"go.uber.org/zap"
)

func main() {
	cfgPath := os.Getenv("LOGGER_CONFIG")
	if cfgPath == "" {
		cfgPath = "configs/logger.dev.json"
	}

	appLogger, err := logger.InitLogger(cfgPath)
	if err != nil {
		bootstrap, btErr := zap.NewProduction()
		if btErr != nil {
			log.Fatal("Logger init failed")
		}

		bootstrap.Error("Logger init failed",
			zap.String("component", "service"),
			zap.String("operation", "init_logger"),
			zap.Error(err),
		)
		_ = bootstrap.Sync()
		os.Exit(1)
	}
	defer func() {
		_ = appLogger.Sync()
	}()

	appLogger.Info("Logger initialized",
		zap.String("component", "service"),
		zap.String("operation", "init_logger"),
	)

	db, err := database.InitDB(context.Background(), "", appLogger)
	if err != nil {
		// Error is logged inside repository layer (InitDB / migrations). Avoid duplicates here.
		os.Exit(1)
	}
	defer db.Close()
}
