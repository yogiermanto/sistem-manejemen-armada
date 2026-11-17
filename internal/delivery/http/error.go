package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"sistem-manajemen-armada/internal/model"
)

func writeError(ctx *gin.Context, log *slog.Logger, err error) {
	switch {
	case errors.Is(err, model.ErrBadRequest):
		ctx.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
	case errors.Is(err, model.ErrNotFound):
		ctx.JSON(http.StatusNotFound, model.ErrorResponse{Error: err.Error()})
	default:
		log.Error("request failed", "path", ctx.FullPath(), "err", err)
		ctx.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
	}
}
