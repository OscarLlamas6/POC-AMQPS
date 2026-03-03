package rabbitmq

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

func NewPublisher(url, queueName string) (*Publisher, error) {
	var conn *amqp.Connection
	var err error

	if strings.HasPrefix(url, "amqps://") {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		conn, err = amqp.DialTLS(url, tlsConfig)
	} else {
		conn, err = amqp.Dial(url)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	_, err = channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	log.Printf("Publisher connected to RabbitMQ, queue: %s", queueName)

	return &Publisher{
		conn:      conn,
		channel:   channel,
		queueName: queueName,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = p.channel.PublishWithContext(
		ctx,
		"",
		p.queueName,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Message published to queue %s", p.queueName)
	return nil
}

func (p *Publisher) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}
	log.Println("Publisher closed")
	return nil
}
