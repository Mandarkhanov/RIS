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
	// task := models.WorkerTask{
	// 	RequestID:  "730a04e6-4de9-41f9-9d5b-53b88b17afac",
	// 	Hash:       "e2fc714c4727ee9395f324cd2e7f331f",
	// 	MaxLength:  4,
	// 	PartNumber: 1,
	// 	PartCount:  1,
	// 	Alphabet:   models.Alphabet{Symbols: []string{"a", "b", "c", "d"}},
	// }
	// ProcessTask(task)

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config : %v", err)
	}

	app := NewWorkerApp(cfg)
	mux := app.setupServer()

	log.Printf("Worker started on :%s", app.cfg.Port)
	log.Fatal(http.ListenAndServe(":"+app.cfg.Port, mux))
}
