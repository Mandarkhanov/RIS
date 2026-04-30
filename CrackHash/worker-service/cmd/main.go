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

	if err := rmqClient.DeclareExchange("cancel_exchange", "fanout"); err != nil {
		log.Fatalf("Failed to declare cancel exchange: %v", err)
	}
	cancelQueueName, err := rmqClient.DeclareEphemeralQueue()
	if err != nil {
		log.Fatalf("Failed to declare ephemeral queue: %v", err)
	}
	if err := rmqClient.BindQueue(cancelQueueName, "", "cancel_exchange"); err != nil {
		log.Fatalf("Failed to bind cancel queue: %v", err)
	}
	cancelMsgs, err := rmqClient.Consume(cancelQueueName)
	if err != nil {
		log.Fatalf("Failed to consume from cancel queue: %v", err)
	}

	activeTaskRepo := repository.NewActiveTaskRepository()
	crackWorkerService := service.NewCrackWorkerService(cfg, activeTaskRepo, rmqClient)

	cancelConsumer := amqp.NewCancelConsumer(crackWorkerService)
	go cancelConsumer.Start(cancelMsgs)

	msgs, err := rmqClient.Consume(cfg.TasksQueue)
	if err != nil {
		log.Fatalf("Failed to consume from tasks queue: %v", err)
	}

	consumer := amqp.NewConsumer(crackWorkerService)
	go consumer.Start(msgs)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Worker is shutting down...")
}
