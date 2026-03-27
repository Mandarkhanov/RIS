package main

import (
	"context"
	"crackhash/pkg/models"
	"encoding/xml"
	"net/http"
)

func (a *WorkerApp) handleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var task models.WorkerTask
	if err := xml.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid XML body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	a.mu.Lock()
	a.activeTasks[task.RequestID] = cancel
	a.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	go a.processTask(ctx, task)
}

func (a *WorkerApp) handleCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.WorkerCancelRequest
	if err := xml.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid XML body", http.StatusBadRequest)
		return
	}

	a.mu.Lock()
	if cancelFunc, exists := a.activeTasks[req.RequestID]; exists {
		cancelFunc()
		delete(a.activeTasks, req.RequestID)
	}
	a.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}
