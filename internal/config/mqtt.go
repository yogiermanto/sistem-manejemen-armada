package config

import (
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func NewMQTTSubscriber(cfg *Config, log *slog.Logger, handler mqtt.MessageHandler) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBroker).
		SetClientID(cfg.MQTTClientID).
		SetCleanSession(false).
		SetAutoAckDisabled(true).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetMaxReconnectInterval(10 * time.Second).
		SetDefaultPublishHandler(handler).
		SetOnConnectHandler(func(c mqtt.Client) {
			if t := c.Subscribe(cfg.MQTTTopic, 1, handler); t.Wait() && t.Error() != nil {
				log.Error("mqtt subscribe", "topic", cfg.MQTTTopic, "err", t.Error())
				return
			}
			log.Info("mqtt subscribed", "topic", cfg.MQTTTopic)
		}).
		SetConnectionLostHandler(func(_ mqtt.Client, err error) {
			log.Warn("mqtt connection lost, reconnecting", "err", err)
		})

	client := mqtt.NewClient(opts)
	if t := client.Connect(); t.WaitTimeout(30*time.Second) && t.Error() != nil {
		return nil, t.Error()
	}
	return client, nil
}
