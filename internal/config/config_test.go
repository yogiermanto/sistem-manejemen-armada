package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaultsWithoutEnvFile(t *testing.T) {
	t.Chdir(t.TempDir())
	c, err := NewConfig()
	if err != nil {
		t.Fatal(err)
	}
	if c.HTTPAddr != ":8080" || c.SimInterval != 2*time.Second || len(c.SimVehicles) != 3 {
		t.Fatalf("unexpected defaults: %+v", c)
	}
}

func TestLoadEnvFileThenEnvOverride(t *testing.T) {
	t.Chdir(t.TempDir())
	os.WriteFile(".env", []byte("HTTP_ADDR=:9090\nMQTT_CLIENT_ID=from-file\nDATABASE_HOST=localhost\nDATABASE_USER=armada\nDATABASE_PASSWORD=p@ss/word\n"), 0o600)
	t.Setenv("MQTT_CLIENT_ID", "from-env")
	t.Setenv("DATABASE_HOST", "postgres")

	c, err := NewConfig()
	if err != nil {
		t.Fatal(err)
	}
	if c.HTTPAddr != ":9090" {
		t.Fatalf("want .env value :9090, got %q", c.HTTPAddr)
	}
	if c.DatabaseURL != "postgres://armada:p%40ss%2Fword@postgres:5432/armada?sslmode=disable" {
		t.Fatalf("want env-overridden host and escaped password, got %q", c.DatabaseURL)
	}
	if c.MQTTBroker != "tcp://localhost:1883" || c.RabbitMQURL != "amqp://armada:armada@localhost:5672/" {
		t.Fatalf("unexpected broker URLs: %q %q", c.MQTTBroker, c.RabbitMQURL)
	}
	if c.MQTTClientID != "from-env" {
		t.Fatalf("want env var to override .env (from-env), got %q", c.MQTTClientID)
	}
}

func TestLoadRejectsInvalid(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("SIM_INTERVAL", "abc")
	if _, err := NewConfig(); err == nil {
		t.Fatal("want error for non-duration SIM_INTERVAL")
	}
}
