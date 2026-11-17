package config

import (
	"context"
	"log/slog"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"

	deliveryhttp "sistem-manajemen-armada/internal/delivery/http"
	"sistem-manajemen-armada/internal/delivery/http/route"
	"sistem-manajemen-armada/internal/delivery/messaging"
	"sistem-manajemen-armada/internal/repository"
	"sistem-manajemen-armada/internal/usecase"
)

type BootstrapConfig struct {
	Config   *Config
	DB       *pgxpool.Pool
	App      *gin.Engine
	Log      *slog.Logger
	Validate *validator.Validate
	Producer usecase.GeofenceEventProducer
}

func Bootstrap(ctx context.Context, c *BootstrapConfig) (mqtt.Client, error) {
	vehicleLocationRepository := repository.NewVehicleLocationRepository(c.DB)
	geofenceRepository := repository.NewGeofenceRepository(c.DB)

	geofences, err := geofenceRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	detector := usecase.NewGeofenceDetector(geofences)
	c.Log.Info("geofences loaded", "count", len(geofences))

	vehicleLocationUseCase := usecase.NewVehicleLocationUseCase(c.Log, c.Validate, vehicleLocationRepository, detector, c.Producer)
	geofenceUseCase := usecase.NewGeofenceUseCase(c.Log, c.Validate, geofenceRepository, detector)

	routeConfig := route.RouteConfig{
		App:                       c.App,
		VehicleLocationController: deliveryhttp.NewVehicleLocationController(c.Log, vehicleLocationUseCase),
		GeofenceController:        deliveryhttp.NewGeofenceController(c.Log, geofenceUseCase),
	}
	routeConfig.Setup()

	locationConsumer := messaging.NewLocationConsumer(ctx, c.Log, vehicleLocationUseCase)
	return NewMQTTSubscriber(c.Config, c.Log, locationConsumer.Handle)
}
