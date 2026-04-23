package main

import (
	"context"
	"crackhash/manager/internal/config"
	"crackhash/manager/internal/database"
	"crackhash/manager/internal/repository"
	"crackhash/manager/internal/service"
	"crackhash/manager/internal/transport/rest"
	"log"
	"net/http"
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

	crackManagerService := service.NewCrackManagerService(cfg, taskRepo)

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

	log.Println("Manager started on :" + cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	// TODO: Добавить Graceful Shutdown
}
