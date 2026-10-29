// Команда shortener — сервис сокращения ссылок.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"urlshortener/internal/config"
	"urlshortener/internal/httpapi"
	"urlshortener/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	store := storage.New()
	handler := httpapi.NewHandler(cfg, store, logger)

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: handler,
	}

	logger.Info("starting server", slog.String("addr", cfg.HTTPAddr))
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}