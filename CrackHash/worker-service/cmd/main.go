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

	rmqClient := rabbitmq.NewClient(cfg.RabbitMQURL)
	defer rmqClient.Close()

	activeTaskRepo := repository.NewActiveTaskRepository()
	crackWorkerService := service.NewCrackWorkerService(cfg, activeTaskRepo, rmqClient)

	cancelConsumer := amqp.NewCancelConsumer(crackWorkerService)
	rmqClient.ConsumeFanout("cancel_exchange", cancelConsumer.HandleMessage)

	consumer := amqp.NewConsumer(crackWorkerService)
	rmqClient.Consume(cfg.TasksQueue, consumer.HandleMessage)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Worker is shutting down...")
}
