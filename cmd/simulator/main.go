package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"sistem-manajemen-armada/internal/config"
	"sistem-manajemen-armada/internal/model"
)

const (
	metersPerDegLat = 111320.0
	halfRouteM      = 300.0
	stepM           = 20.0
)

func main() {
	config.NewLogger()
	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBroker).
		SetClientID(fmt.Sprintf("armada-simulator-%d", os.Getpid())).
		SetAutoReconnect(true).
		SetConnectRetry(true)
	client := mqtt.NewClient(opts)
	if t := client.Connect(); t.Wait() && t.Error() != nil {
		slog.Error("mqtt connect", "err", t.Error())
		os.Exit(1)
	}
	defer client.Disconnect(250)

	slog.Info("simulator started", "vehicles", cfg.SimVehicles, "interval", cfg.SimInterval.String())
	for i, id := range cfg.SimVehicles {
		go drive(ctx, client, cfg, id, i)
	}
	<-ctx.Done()
}

func drive(ctx context.Context, client mqtt.Client, cfg *config.Config, vehicleID string, i int) {
	n := float64(len(cfg.SimVehicles))
	bearing := float64(i) * math.Pi / n
	offset := -halfRouteM + math.Mod(float64(i)*90, 2*halfRouteM)
	dir := 1.0
	topic := fmt.Sprintf("/fleet/vehicle/%s/location", vehicleID)
	metersPerDegLon := metersPerDegLat * math.Cos(cfg.SimCenterLat*math.Pi/180)

	ticker := time.NewTicker(cfg.SimInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			loc := model.LocationMessage{
				VehicleID: vehicleID,
				Latitude:  cfg.SimCenterLat + offset*math.Cos(bearing)/metersPerDegLat,
				Longitude: cfg.SimCenterLon + offset*math.Sin(bearing)/metersPerDegLon,
				Timestamp: now.Unix(),
			}
			body, _ := json.Marshal(loc)
			if t := client.Publish(topic, 1, false, body); t.Wait() && t.Error() != nil {
				slog.Error("publish", "vehicle_id", vehicleID, "err", t.Error())
			} else {
				slog.Info("published", "topic", topic, "latitude", loc.Latitude, "longitude", loc.Longitude, "distance_from_center_m", math.Abs(offset))
			}

			offset += dir * stepM
			if math.Abs(offset) >= halfRouteM {
				dir = -dir
			}
		}
	}
}
