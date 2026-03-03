package rabbitmq

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/oscar/messaging-playgrounds/internal/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

func NewConsumer(url, queueName string) (*Consumer, error) {
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

	err = channel.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	log.Printf("Consumer connected to RabbitMQ, queue: %s", queueName)

	return &Consumer{
		conn:      conn,
		channel:   channel,
		queueName: queueName,
	}, nil
}

func (c *Consumer) Consume(ctx context.Context, handler func(*models.Message) error) error {
	msgs, err := c.channel.Consume(
		c.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("Consumer started, waiting for messages on queue: %s", c.queueName)

	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping consumer")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Message channel closed")
				return fmt.Errorf("message channel closed")
			}

			var message models.Message
			if err := json.Unmarshal(msg.Body, &message); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				msg.Nack(false, false)
				continue
			}

			log.Printf("Received message: ID=%s, Content=%s", message.ID, message.Content)

			if err := handler(&message); err != nil {
				log.Printf("Handler error: %v", err)
				msg.Nack(false, true)
				continue
			}

			if err := msg.Ack(false); err != nil {
				log.Printf("Failed to acknowledge message: %v", err)
			}
		}
	}
}

func (c *Consumer) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}
	log.Println("Consumer closed")
	return nil
}

func (c *Consumer) Reconnect(url string) error {
	log.Println("Attempting to reconnect to RabbitMQ...")

	time.Sleep(5 * time.Second)

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
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel on reconnect: %w", err)
	}

	c.conn = conn
	c.channel = channel

	log.Println("Reconnected to RabbitMQ successfully")
	return nil
}
