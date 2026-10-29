// Package config читает конфигурацию сервиса из переменных окружения.
package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config содержит параметры запуска сервиса.
type Config struct {
	HTTPAddr   string
	BaseURL    string
	CodeLength int
	LogLevel   slog.Level
}

// Load читает и валидирует переменные окружения.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:   ":8080",
		BaseURL:    "http://localhost:8080",
		CodeLength: 7,
		LogLevel:   slog.LevelInfo,
	}

	if val, ok := os.LookupEnv("HTTP_ADDR"); ok {
		val = strings.TrimSpace(val)
		if val == "" {
			return Config{}, fmt.Errorf("HTTP_ADDR: cannot be empty")
		}
		cfg.HTTPAddr = val
	}

	if val, ok := os.LookupEnv("BASE_URL"); ok {
		val = strings.TrimSpace(val)
		parsedURL, err := url.ParseRequestURI(val)
		if err != nil || !parsedURL.IsAbs() || parsedURL.Host == "" {
			return Config{}, fmt.Errorf("BASE_URL: expected absolute URL with host, got %q", val)
		}
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return Config{}, fmt.Errorf("BASE_URL: scheme must be http or https, got %q", parsedURL.Scheme)
		}
		cfg.BaseURL = strings.TrimRight(val, "/")
	}

	if val, ok := os.LookupEnv("CODE_LENGTH"); ok {
		n, err := strconv.Atoi(val)
		if err != nil || n < 1 {
			return Config{}, fmt.Errorf("CODE_LENGTH: expected integer >= 1, got %q", val)
		}
		cfg.CodeLength = n
	}

	if val, ok := os.LookupEnv("LOG_LEVEL"); ok {
		switch strings.ToLower(strings.TrimSpace(val)) {
		case "debug":
			cfg.LogLevel = slog.LevelDebug
		case "info":
			cfg.LogLevel = slog.LevelInfo
		case "warn":
			cfg.LogLevel = slog.LevelWarn
		case "error":
			cfg.LogLevel = slog.LevelError
		default:
			return Config{}, fmt.Errorf("LOG_LEVEL: unknown level %q", val)
		}
	}

	return cfg, nil
}