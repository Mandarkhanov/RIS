package main

import (
	"context"
	"crackhash/manager/internal/config"
	"crackhash/manager/internal/database"
	"crackhash/manager/internal/repository"
	"crackhash/manager/internal/service"
	"crackhash/manager/internal/transport/amqp"
	"crackhash/manager/internal/transport/rest"
	"crackhash/pkg/rabbitmq"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config : %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoConn, err := database.NewMongoConnection(ctx, cfg.MongoURL, cfg.MongoDBName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoConn.Disconnect(context.Background())

	taskRepo, err := repository.NewTaskRepository(context.Background(), mongoConn.DB)
	if err != nil {
		log.Fatalf("Failed to init TaskRepository: %v", err)
	}

	rmqClient := rabbitmq.NewClient(cfg.RabbitMQURL)
	defer rmqClient.Close()

	crackManagerService := service.NewCrackManagerService(cfg, taskRepo, rmqClient)

	amqpConsumer := amqp.NewConsumer(crackManagerService)
	rmqClient.Consume(cfg.ResultsQueue, amqpConsumer.HandleMessage)

	handler := rest.NewHandler(crackManagerService)
	mux := handler.InitRoutes()

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go crackManagerService.TimeoutWatcher()
	go crackManagerService.OutboxRelay()
	go func() {
		log.Println("Manager REST API started on :" + cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Manager is shutting down...")
}
