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

const (
	WorkerTaskPath   = "/internal/api/worker/hash/crack/task"
	WorkerCancelPath = "/internal/api/worker/hash/crack/cancel"

	ManagerRequestPath = "/internal/api/manager/hash/crack/request"
)

func (app *WorkerApp) setupServer() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(WorkerTaskPath, app.handleTask)
	mux.HandleFunc(WorkerCancelPath, app.handleCancel)
	return mux
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config : %v", err)
	}

	app := NewWorkerApp(cfg)
	mux := app.setupServer()
	srv := &http.Server{
		Addr:         ":" + app.cfg.Port,
		Handler:      mux,
		ReadTimeout:  app.cfg.ReadTimeout,
		WriteTimeout: app.cfg.WriteTimeout,
		IdleTimeout:  app.cfg.IdleTimeout,
	}

	log.Printf("Worker started on :%s", app.cfg.Port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	// TODO: Добавить Graceful Shutdown
}
