package rabbitmq

import (
	"fmt"
	"os"
	"strings"

	"github.com/rabbitmq/amqp091-go"
)

func ConnectionRabbitMQ() (*amqp091.Connection, *amqp091.Channel, error) {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		host := getEnv("RABBITMQ_HOST", "localhost")
		port := getEnv("RABBITMQ_PORT", "5672")
		user := getEnv("RABBITMQ_USER", "guest")
		password := getEnv("RABBITMQ_PASSWORD", "guest")
		vhost := strings.TrimPrefix(getEnv("RABBITMQ_VHOST", "/"), "/")

		url = fmt.Sprintf("amqp://%s:%s@%s:%s/%s", user, password, host, port, vhost)
	}

	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	return conn, ch, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
