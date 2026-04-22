package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

type ManagerApp struct {
	cfg     *Config
	storage *MongoStorage
}

func NewManagerApp(cfg *Config, storage *MongoStorage) *ManagerApp {
	return &ManagerApp{
		cfg:     cfg,
		storage: storage,
	}
}

const (
	ManagerCrackPath   = "/api/hash/crack"
	ManagerStatusPath  = "/api/hash/status"
	ManagerRequestPath = "/internal/api/manager/hash/crack/request"

	WorkerTaskPath   = "/internal/api/worker/hash/crack/task"
	WorkerCancelPath = "/internal/api/worker/hash/crack/cancel"
)

func (a *ManagerApp) setupHandlers() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(ManagerCrackPath, a.handleCrackRequest)
	mux.HandleFunc(ManagerStatusPath, a.handleCrackStatusCheck)
	mux.HandleFunc(ManagerRequestPath, a.handleWorkerResponse)
	return mux
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config : %v\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	storage, err := NewMongoStorage(ctx, cfg.MongoURL, cfg.MongoDBName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v\n", err)
	}
	defer storage.Disconnect(context.Background())

	app := NewManagerApp(cfg, storage)
	mux := app.setupHandlers()
	srv := &http.Server{
		Addr:         ":" + app.cfg.Port,
		Handler:      mux,
		ReadTimeout:  app.cfg.ReadTimeout,
		WriteTimeout: app.cfg.WriteTimeout,
		IdleTimeout:  app.cfg.IdleTimeout,
	}

	go app.timeoutWatcher()

	log.Println("Manager started on :" + app.cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	// TODO: Добавить Graceful Shutdown
}
