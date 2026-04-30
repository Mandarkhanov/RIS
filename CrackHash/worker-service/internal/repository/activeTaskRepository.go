package repository

import (
	"context"
	"sync"
)

type ActiveTaskRepository struct {
	activeTasks    map[string]context.CancelFunc
	cancelledTasks map[string]struct{}
	mu             sync.Mutex
}

func NewActiveTaskRepository() *ActiveTaskRepository {
	return &ActiveTaskRepository{
		activeTasks:    make(map[string]context.CancelFunc),
		cancelledTasks: make(map[string]struct{}),
	}
}

func (r *ActiveTaskRepository) Add(reqID string, cancel context.CancelFunc) (added bool, cancelled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.cancelledTasks[reqID]; exists {
		return false, true
	}

	if _, exists := r.activeTasks[reqID]; exists {
		return false, false
	}

	r.activeTasks[reqID] = cancel
	return true, false
}

func (r *ActiveTaskRepository) Cancel(reqID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cancelledTasks[reqID] = struct{}{}

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
