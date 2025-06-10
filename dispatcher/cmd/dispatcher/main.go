package main

import (
	"dispatcher/internal/config"
	"dispatcher/internal/dispatcher" 
	"log"
	"os"
	"context"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.LoadConfig("../../configs/dispatcher.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	topics, err := config.LoadTopics(cfg.Kafka.TopicPath)
	if err != nil {
		log.Fatalf("failed to load topic: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	dispatcher, consumerGroup, err := dispatch.NewDispatcher(
		ctx,
		cfg.Kafka.Brokers,
		cfg.Kafka.GroupID,
		cfg.Dispatcher.NumInBatch,
		cfg.Dispatcher.NumWindow,
		cfg.Dispatcher.RateError,
		topics,
	)
	if err != nil {
		log.Fatalf("Failed to initialize dispatcher: %v", err)
	}

	// Run dispatcher loop in background
	go dispatch.Run(ctx, dispatcher, consumerGroup, topics)

	log.Println("Dispatcher started. Waiting for shutdown signal...")

	<-stop
	log.Println("Shutdown signal received, exiting...")

	if err := consumerGroup.Close(); err != nil {
		log.Printf("Failed to close consumer group: %v", err)
	}
}

