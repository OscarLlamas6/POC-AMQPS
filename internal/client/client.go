package client

import (
	"context"
	"log"
	"time"

	"github.com/oscar/messaging-playgrounds/internal/config"
	"github.com/oscar/messaging-playgrounds/internal/models"
	"github.com/oscar/messaging-playgrounds/internal/rabbitmq"
)

type Consumer struct {
	consumer *rabbitmq.Consumer
	config   *config.ClientConfig
}

func NewConsumer(cfg *config.ClientConfig) (*Consumer, error) {
	consumer, err := rabbitmq.NewConsumer(cfg.RabbitMQURL, cfg.QueueName)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer: consumer,
		config:   cfg,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer stopping due to context cancellation")
			return nil
		default:
			err := c.consumer.Consume(ctx, c.handleMessage)
			if err != nil {
				log.Printf("Consumer error: %v", err)
				
				if ctx.Err() != nil {
					return nil
				}

				log.Println("Attempting to reconnect...")
				if err := c.consumer.Reconnect(c.config.RabbitMQURL); err != nil {
					log.Printf("Reconnection failed: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}
			}
		}
	}
}

func (c *Consumer) handleMessage(msg *models.Message) error {
	log.Printf("Processing message: ID=%s, Content=%s, Source=%s, Type=%s, Timestamp=%s",
		msg.ID,
		msg.Content,
		msg.Metadata.Source,
		msg.Metadata.Type,
		msg.Timestamp.Format(time.RFC3339),
	)

	time.Sleep(100 * time.Millisecond)

	log.Printf("Message processed successfully: ID=%s", msg.ID)
	return nil
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}
