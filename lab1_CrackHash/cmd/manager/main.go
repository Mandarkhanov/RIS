package main

import (
	"log"
	"net/http"
)

type ManagerApp struct {
	cfg     *Config
	storage *TasksStorage
}

func NewManagerApp(cfg *Config) *ManagerApp {
	return &ManagerApp{
		cfg:     cfg,
		storage: NewTasksStorage(cfg.TasksStorageSize),
	}
}

func (a *ManagerApp) setupHandlers() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/hash/crack", a.handleCrackRequest)
	mux.HandleFunc("/api/hash/status", a.handleCrackStatusCheck)

	mux.HandleFunc("/internal/api/manager/hash/crack/request", a.handleWorkerResponse)

	return mux
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config : %v\n", err)
	}

	app := NewManagerApp(cfg)
	mux := app.setupHandlers()

	go app.timeoutWatcher()
	log.Println("Manager started on :" + app.cfg.Port)
	log.Fatal(http.ListenAndServe(":"+app.cfg.Port, mux))
}
