package config

import (
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)
	return log
}
