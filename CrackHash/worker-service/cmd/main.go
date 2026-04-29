package main

import (
	"crackhash/pkg/rabbitmq"
	"crackhash/worker/internal/config"
	"crackhash/worker/internal/repository"
	"crackhash/worker/internal/service"
	"crackhash/worker/internal/transport/amqp"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config : %v", err)
	}

	rmqClient, err := rabbitmq.NewClient(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rmqClient.Close()

	if err := rmqClient.DeclareQueue(cfg.TasksQueue); err != nil {
		log.Fatalf("Failed to declare tasks queue: %v", err)
	}
	if err := rmqClient.DeclareQueue(cfg.ResultsQueue); err != nil {
		log.Fatalf("Failed to declare results queue: %v", err)
	}

	activeTaskRepo := repository.NewActiveTaskRepository()
	crackWorkerService := service.NewCrackWorkerService(cfg, activeTaskRepo, rmqClient)

	msgs, err := rmqClient.Consume(cfg.TasksQueue)
	if err != nil {
		log.Fatalf("Failed to consume from tasks queue: %v", err)
	}

	consumer := amqp.NewConsumer(crackWorkerService)
	go consumer.Start(msgs)

	// 6. Graceful Shutdown (Ожидание сигнала завершения)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Worker is shutting down...")
}
