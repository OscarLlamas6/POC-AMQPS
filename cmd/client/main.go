package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/oscar/messaging-playgrounds/internal/client"
	"github.com/oscar/messaging-playgrounds/internal/config"
)

func main() {
	cfg := config.LoadClientConfig()

	consumer, err := client.NewConsumer(cfg)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Consumer error: %v", err)
		}
	}()

	<-quit
	log.Println("Received shutdown signal, initiating graceful shutdown...")

	cancel()

	shutdownComplete := make(chan struct{})
	go func() {
		wg.Wait()
		close(shutdownComplete)
	}()

	select {
	case <-shutdownComplete:
		log.Println("Graceful shutdown completed")
	case <-time.After(15 * time.Second):
		log.Println("Shutdown timeout exceeded, forcing exit")
	}

	log.Println("Client exited")
}
