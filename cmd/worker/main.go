package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"sistem-manajemen-armada/internal/config"
	"sistem-manajemen-armada/internal/delivery/messaging"
	"sistem-manajemen-armada/internal/repository"
	"sistem-manajemen-armada/internal/usecase"
)

func main() {
	log := config.NewLogger()
	if err := run(log); err != nil {
		log.Error("worker stopped", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.NewConfig()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := config.NewDatabase(ctx, cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	conn, ch, err := config.NewRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}
	deliveries, err := ch.Consume(config.RabbitMQQueue, "geofence-worker", false, false, false, false, nil)
	if err != nil {
		return err
	}

	geofenceEventUseCase := usecase.NewGeofenceEventUseCase(log, config.NewValidator(), repository.NewGeofenceEventRepository(db))
	consumer := messaging.NewGeofenceEventConsumer(log, geofenceEventUseCase)

	log.Info("worker consuming", "queue", config.RabbitMQQueue)
	return messaging.ConsumeQueue(ctx, conn, deliveries, consumer.Consume)
}
