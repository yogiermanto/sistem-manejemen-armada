package route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	deliveryhttp "sistem-manajemen-armada/internal/delivery/http"
)

type RouteConfig struct {
	App                       *gin.Engine
	VehicleLocationController *deliveryhttp.VehicleLocationController
	GeofenceController        *deliveryhttp.GeofenceController
}

func (c *RouteConfig) Setup() {
	c.App.GET("/health", func(ctx *gin.Context) { ctx.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	c.App.GET("/vehicles/:vehicle_id/location", c.VehicleLocationController.GetLatest)
	c.App.GET("/vehicles/:vehicle_id/history", c.VehicleLocationController.GetHistory)

	c.App.POST("/geofences", c.GeofenceController.Create)
	c.App.GET("/geofences", c.GeofenceController.List)
}
