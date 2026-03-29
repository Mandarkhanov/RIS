package main

import (
	"crackhash/pkg/models"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrStorageFull = errors.New("Not enough memory at TasksStorage")
)

type RequestState struct {
	Hash            string
	MaxLength       int
	Status          models.RequestStatus
	Data            []string
	WorkersFinished int
	TotalWorkers    int
	CreatedAt       time.Time
	mu              sync.Mutex
}

type TasksStorage struct {
	tasks         map[string]*RequestState
	tasksByParams map[string]string
	size          int
	mu            sync.RWMutex
}

func NewTasksStorage(storageSize int) *TasksStorage {
	return &TasksStorage{
		tasks:         make(map[string]*RequestState),
		tasksByParams: make(map[string]string),
		size:          storageSize,
	}
}

func (s *TasksStorage) Get(reqID string) (*RequestState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, exists := s.tasks[reqID]
	return state, exists
}

func (s *TasksStorage) StoreOrGetExists(reqID string, state *RequestState) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cacheKey := fmt.Sprintf("%s_%d", state.Hash, state.MaxLength)

	if existingReqId, exists := s.tasksByParams[cacheKey]; exists {
		return existingReqId, nil
	}

	if len(s.tasks) >= s.size {
		return "", ErrStorageFull
	}

	s.tasks[reqID] = state
	s.tasksByParams[cacheKey] = reqID
	return reqID, nil
}

func (s *TasksStorage) Delete(reqID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if state, exists := s.tasks[reqID]; exists {
		cacheKey := fmt.Sprintf("%s_%d", state.Hash, state.MaxLength)
		delete(s.tasksByParams, cacheKey)
		delete(s.tasks, reqID)
	}
}
