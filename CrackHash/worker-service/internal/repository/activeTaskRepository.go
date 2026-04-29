package repository

import (
	"context"
	"sync"
)

type ActiveTaskRepository struct {
	activeTasks map[string]context.CancelFunc
	mu          sync.Mutex
}

func NewActiveTaskRepository() *ActiveTaskRepository {
	return &ActiveTaskRepository{
		activeTasks: make(map[string]context.CancelFunc),
	}
}

func (r *ActiveTaskRepository) Add(reqID string, cancel context.CancelFunc) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.activeTasks[reqID]; exists {
		return false
	}

	r.activeTasks[reqID] = cancel
	return true
}

func (r *ActiveTaskRepository) Cancel(reqID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cancel, exists := r.activeTasks[reqID]; exists {
		cancel()
		delete(r.activeTasks, reqID)
	}
}

func (r *ActiveTaskRepository) CancelAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for reqID, cancel := range r.activeTasks {
		cancel()
		delete(r.activeTasks, reqID)
	}
}
