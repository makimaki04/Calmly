package config

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	Address      string `env:"RUN_ADDRESS"`
	DatabaseURI  string `env:"DATABASE_URI"`
	JWTSecret    string `env:"JWT_SECRET"`
	LoggerConfig string `env:"LOGGER_CONFIG"`
	DBConf       DBLimits
	DumpExpTime  time.Duration `env:"DUMP_EXPTIME"`
}

type DBLimits struct {
	MaxOpenConns int
	MaxIdleConns int
	MaxLifeTime  time.Duration
	MaxIdleTime  time.Duration
}

var dbLims = DBLimits{
	MaxOpenConns: 20,
	MaxIdleConns: 20,
	MaxLifeTime:  12 * time.Minute,
	MaxIdleTime:  5 * time.Minute,
}

func LoadAppConfig() (Config, error) {
	var cfg Config

	// Optional: local dev convenience. Missing .env is not an error.
	_ = godotenv.Load(".env")

	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse env config: %w", err)
	}

	flag.StringVar(&cfg.Address, "a", cfg.Address, "server address, example :8080")
	flag.StringVar(&cfg.DatabaseURI, "db", cfg.DatabaseURI, "database uri")
	flag.StringVar(&cfg.JWTSecret, "jwt-secret", cfg.JWTSecret, "jwt secret")
	flag.StringVar(&cfg.LoggerConfig, "logger", cfg.LoggerConfig, "logger config path")
	flag.DurationVar(&cfg.DumpExpTime, "dump-exptime", cfg.DumpExpTime, "dump expires time value in days")

	flag.Parse()

	cfg.Address = strings.TrimSpace(cfg.Address)
	cfg.DatabaseURI = strings.TrimSpace(cfg.DatabaseURI)
	cfg.JWTSecret = strings.TrimSpace(cfg.JWTSecret)
	cfg.LoggerConfig = strings.TrimSpace(cfg.LoggerConfig)

	if cfg.Address == "" {
		return Config{}, fmt.Errorf("server address is required")
	}
	if cfg.DatabaseURI == "" {
		return Config{}, fmt.Errorf("database uri is required")
	}
	if cfg.LoggerConfig == "" {
		return Config{}, fmt.Errorf("logger config path is required")
	}

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("jwt secret is required")
	}

	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("jwt secret must be at least 32 chars")
	}

	// Apply defaults if limits were not provided.
	if cfg.DBConf == (DBLimits{}) {
		cfg.DBConf = dbLims
	}

	if cfg.DumpExpTime <= 0 {
		cfg.DumpExpTime = 20 * 24 * time.Hour
	}

	return cfg, nil
}
