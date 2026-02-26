package logger

import (
	"encoding/json"
	"fmt"
	"os"

	"go.uber.org/zap"
)

func InitLogger(cfgPath string) (*zap.Logger, error) {
	cfgJson, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("read logger config: %w", err)
	}

	var cfg zap.Config
	if err := json.Unmarshal(cfgJson, &cfg); err != nil {
		return nil, fmt.Errorf("parse logger config: %w", err)
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}

	return logger, nil
}
