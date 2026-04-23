package main

import (
	"crackhash/worker/internal/config"
	"crackhash/worker/internal/service"
	"crackhash/worker/internal/transport/rest"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config : %v", err)
	}

	crackWorkerService := service.NewCrackWorkerService(cfg)

	handler := rest.NewHandler(crackWorkerService)
	mux := handler.InitRoutes()

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	log.Printf("Worker started on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	// TODO: Добавить Graceful Shutdown
}
