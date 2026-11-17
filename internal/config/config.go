package config

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	HTTPAddr     string
	DatabaseURL  string
	MQTTBroker   string
	MQTTClientID string
	MQTTTopic    string
	RabbitMQURL  string

	SimVehicles  []string
	SimInterval  time.Duration
	SimCenterLat float64
	SimCenterLon float64
}

func NewConfig() (*Config, error) {
	v := viper.New()
	v.SetDefault("HTTP_ADDR", ":8080")
	v.SetDefault("DATABASE_HOST", "localhost")
	v.SetDefault("DATABASE_PORT", "5432")
	v.SetDefault("DATABASE_USER", "armada")
	v.SetDefault("DATABASE_PASSWORD", "armada")
	v.SetDefault("DATABASE_NAME", "armada")
	v.SetDefault("DATABASE_SSLMODE", "disable")
	v.SetDefault("MQTT_HOST", "localhost")
	v.SetDefault("MQTT_PORT", "1883")
	v.SetDefault("MQTT_CLIENT_ID", "armada-api")
	v.SetDefault("MQTT_TOPIC", "/fleet/vehicle/+/location")
	v.SetDefault("RABBITMQ_HOST", "localhost")
	v.SetDefault("RABBITMQ_PORT", "5672")
	v.SetDefault("RABBITMQ_USER", "armada")
	v.SetDefault("RABBITMQ_PASSWORD", "armada")
	v.SetDefault("SIM_VEHICLES", "B1234XYZ,B5678ABC,B9012DEF")
	v.SetDefault("SIM_INTERVAL", "2s")
	v.SetDefault("SIM_CENTER_LAT", -6.2088)
	v.SetDefault("SIM_CENTER_LON", 106.8456)

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("read .env: %w", err)
	}

	c := &Config{
		HTTPAddr: v.GetString("HTTP_ADDR"),
		DatabaseURL: (&url.URL{
			Scheme:   "postgres",
			User:     url.UserPassword(v.GetString("DATABASE_USER"), v.GetString("DATABASE_PASSWORD")),
			Host:     net.JoinHostPort(v.GetString("DATABASE_HOST"), v.GetString("DATABASE_PORT")),
			Path:     v.GetString("DATABASE_NAME"),
			RawQuery: "sslmode=" + url.QueryEscape(v.GetString("DATABASE_SSLMODE")),
		}).String(),
		MQTTBroker:   "tcp://" + net.JoinHostPort(v.GetString("MQTT_HOST"), v.GetString("MQTT_PORT")),
		MQTTClientID: v.GetString("MQTT_CLIENT_ID"),
		MQTTTopic:    v.GetString("MQTT_TOPIC"),
		RabbitMQURL: (&url.URL{
			Scheme: "amqp",
			User:   url.UserPassword(v.GetString("RABBITMQ_USER"), v.GetString("RABBITMQ_PASSWORD")),
			Host:   net.JoinHostPort(v.GetString("RABBITMQ_HOST"), v.GetString("RABBITMQ_PORT")),
			Path:   "/",
		}).String(),
		SimVehicles:  strings.Split(v.GetString("SIM_VEHICLES"), ","),
		SimInterval:  v.GetDuration("SIM_INTERVAL"),
		SimCenterLat: v.GetFloat64("SIM_CENTER_LAT"),
		SimCenterLon: v.GetFloat64("SIM_CENTER_LON"),
	}

	if c.SimInterval < time.Second {
		return nil, errors.New("SIM_INTERVAL must be a duration >= 1s (timestamps are unix seconds), e.g. 2s")
	}
	return c, nil
}
