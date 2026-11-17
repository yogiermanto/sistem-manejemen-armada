package http

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"sistem-manajemen-armada/internal/model"
	"sistem-manajemen-armada/internal/usecase"
)

type VehicleLocationController struct {
	Log     *slog.Logger
	UseCase *usecase.VehicleLocationUseCase
}

func NewVehicleLocationController(log *slog.Logger, useCase *usecase.VehicleLocationUseCase) *VehicleLocationController {
	return &VehicleLocationController{Log: log, UseCase: useCase}
}

func (c *VehicleLocationController) GetLatest(ctx *gin.Context) {
	response, err := c.UseCase.GetLatest(ctx.Request.Context(), ctx.Param("vehicle_id"))
	if err != nil {
		writeError(ctx, c.Log, err)
		return
	}
	ctx.JSON(http.StatusOK, model.WebResponse[*model.VehicleLocationResponse]{Data: response})
}

func (c *VehicleLocationController) GetHistory(ctx *gin.Context) {
	request := &model.GetLocationHistoryRequest{VehicleID: ctx.Param("vehicle_id")}
	if err := ctx.ShouldBindQuery(request); err != nil {
		writeError(ctx, c.Log, fmt.Errorf("%w: start, end, page and limit must be integers", model.ErrBadRequest))
		return
	}
	responses, paging, err := c.UseCase.GetHistory(ctx.Request.Context(), request)
	if err != nil {
		writeError(ctx, c.Log, err)
		return
	}
	ctx.JSON(http.StatusOK, model.WebResponse[[]model.VehicleLocationResponse]{Data: responses, Meta: paging})
}
