package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func NewRequestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		if ctx.Request.URL.Path == "/health" {
			return
		}
		log.Info("http request", "method", ctx.Request.Method, "path", ctx.Request.URL.Path,
			"status", ctx.Writer.Status(), "duration_ms", float64(time.Since(start).Microseconds())/1000)
	}
}
