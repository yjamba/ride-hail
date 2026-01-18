package broker

import (
	"fmt"
	"net/url"
	"os"
)

type BrokerConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	VHost    string
}

func NewBrokerConfig(host, port, user, password, vhost string) *BrokerConfig {
	return &BrokerConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		VHost:    vhost,
	}
}

func (rc *BrokerConfig) GetConnectionString() string {
	vhost := rc.VHost
	if vhost == "" {
		vhost = "/"
	}

	return fmt.Sprintf(
		"amqp://%s:%s@%s:%s/%s",
		rc.User,
		rc.Password,
		rc.Host,
		rc.Port,
		url.PathEscape(vhost),
	)
}

func NewBrokerConfigFromEnv() *BrokerConfig {
	getEnv := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}

	return &BrokerConfig{
		Host:     getEnv("RABBITMQ_HOST", "localhost"),
		Port:     getEnv("RABBITMQ_PORT", "5672"),
		User:     getEnv("RABBITMQ_USER", "guest"),
		Password: getEnv("RABBITMQ_PASSWORD", "guest"),
		VHost:    getEnv("RABBITMQ_VHOST", "/"),
	}
}

// "amqp://wheres-my-pizza:wheresmypizza@localhost:5672/%2F"
