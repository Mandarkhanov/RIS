package main

import (
	"context"
	"log"
	"net/http"
	"sync"
)

type WorkerApp struct {
	cfg         *Config
	activeTasks map[string]context.CancelFunc
	mu          sync.Mutex
}

func NewWorkerApp(cfg *Config) *WorkerApp {
	return &WorkerApp{
		cfg:         cfg,
		activeTasks: make(map[string]context.CancelFunc),
	}
}

func (app *WorkerApp) setupServer() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/internal/api/worker/hash/crack/task", app.handleTask)
	mux.HandleFunc("/internal/api/worker/hash/crack/cancel", app.handleCancel)

	return mux
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config : %v", err)
	}

	app := NewWorkerApp(cfg)
	mux := app.setupServer()

	log.Printf("Worker started on :%s", app.cfg.Port)
	log.Fatal(http.ListenAndServe(":"+app.cfg.Port, mux))
}
