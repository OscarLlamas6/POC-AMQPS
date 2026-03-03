package config

import (
	"os"
)

type ServerConfig struct {
	Port        string
	RabbitMQURL string
	QueueName   string
}

type ClientConfig struct {
	RabbitMQURL string
	QueueName   string
}

func LoadServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:        getEnv("SERVER_PORT", "8080"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		QueueName:   getEnv("QUEUE_NAME", "messages"),
	}
}

func LoadClientConfig() *ClientConfig {
	return &ClientConfig{
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		QueueName:   getEnv("QUEUE_NAME", "messages"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
