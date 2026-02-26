package main

import (
	"context"
	"log"
	"os"

	"github.com/makimaki04/Calmly/internal/config"
	"github.com/makimaki04/Calmly/internal/database"
	"github.com/makimaki04/Calmly/internal/logger"
)

func main() {
	cfg, err := config.LoadAppConfig()
	if err != nil {
		log.Fatal("load config: ", err)
	}

	appLogger, err := logger.InitLogger(cfg.LoggerConfig)
	if err != nil {
		log.Fatal("init logger: ", err)
	}
	defer func() {
		_ = appLogger.Sync()
	}()

	db, err := database.InitDB(context.Background(), cfg.DatabaseURI, cfg.DBConf, appLogger)
	if err != nil {
		// Error is logged inside repository layer (InitDB / migrations). Avoid duplicates here.
		os.Exit(1)
	}
	defer db.Close()
}
