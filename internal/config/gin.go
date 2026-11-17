package config

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"sistem-manajemen-armada/internal/delivery/http/middleware"
)

func NewGin(log *slog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()
	app.Use(gin.Recovery(), middleware.NewRequestLogger(log))
	return app
}
