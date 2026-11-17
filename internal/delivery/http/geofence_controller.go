package http

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"sistem-manajemen-armada/internal/model"
	"sistem-manajemen-armada/internal/usecase"
)

type GeofenceController struct {
	Log     *slog.Logger
	UseCase *usecase.GeofenceUseCase
}

func NewGeofenceController(log *slog.Logger, useCase *usecase.GeofenceUseCase) *GeofenceController {
	return &GeofenceController{Log: log, UseCase: useCase}
}

func (c *GeofenceController) Create(ctx *gin.Context) {
	request := new(model.CreateGeofenceRequest)
	if err := ctx.ShouldBindJSON(request); err != nil {
		writeError(ctx, c.Log, fmt.Errorf("%w: body must be JSON {name, latitude, longitude, radius_m}", model.ErrBadRequest))
		return
	}
	response, err := c.UseCase.Create(ctx.Request.Context(), request)
	if err != nil {
		writeError(ctx, c.Log, err)
		return
	}
	ctx.JSON(http.StatusCreated, model.WebResponse[*model.GeofenceResponse]{Data: response})
}

func (c *GeofenceController) List(ctx *gin.Context) {
	responses, err := c.UseCase.List(ctx.Request.Context())
	if err != nil {
		writeError(ctx, c.Log, err)
		return
	}
	ctx.JSON(http.StatusOK, model.WebResponse[[]model.GeofenceResponse]{Data: responses})
}
