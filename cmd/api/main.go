package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"sistem-manajemen-armada/internal/config"
	"sistem-manajemen-armada/internal/gateway/messaging"
)

func main() {
	log := config.NewLogger()
	if err := run(log); err != nil {
		log.Error("api stopped", "err", err)
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

	producer, err := messaging.NewGeofenceProducer(log, config.RabbitMQExchange, func() (*amqp.Connection, *amqp.Channel, error) {
		return config.NewRabbitMQ(cfg.RabbitMQURL)
	})
	if err != nil {
		return err
	}
	defer producer.Close()

	app := config.NewGin(log)
	mqttClient, err := config.Bootstrap(ctx, &config.BootstrapConfig{
		Config:   cfg,
		DB:       db,
		App:      app,
		Log:      log,
		Validate: config.NewValidator(),
		Producer: producer,
	})
	if err != nil {
		return err
	}
	defer mqttClient.Disconnect(250)

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: app, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Info("api listening", "addr", cfg.HTTPAddr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
