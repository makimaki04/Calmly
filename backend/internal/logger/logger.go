package logger

import (
	"encoding/json"
	"os"

	"go.uber.org/zap"
)

func InitLogger(cfgPath string) (*zap.Logger, error) {
	cfgJson, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var cfg zap.Config
	if err := json.Unmarshal(cfgJson, &cfg);err != nil {
		return nil, err
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	
	return logger, nil
}