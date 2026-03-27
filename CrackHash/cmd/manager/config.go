package main

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultTasksStorageSize  = 1000
	DefaultTaskTimeout       = 30
	DefaultTaskWatchInterval = 5
)

type Config struct {
	Port        string
	WorkersURLs []string

	TasksStorageSize int
	TaskTimeout      time.Duration
	WatchInterval    time.Duration
}

func LoadConfig() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	urlsEnv := os.Getenv("WORKERS_URLS")
	if urlsEnv == "" {
		urlsEnv = "http://localhost:8081/internal/api/worker/hash/crack/task"
	}
	var urls []string
	for _, u := range strings.Split(urlsEnv, ",") {
		urls = append(urls, strings.TrimSpace(u))
	}

	storageSize := DefaultTasksStorageSize
	if storageSizeStr := os.Getenv("MANAGER_MAX_TASKS_STORAGE"); storageSizeStr != "" {
		if val, err := strconv.Atoi(storageSizeStr); err == nil {
			storageSize = val
		}
	}

	taskTimeout := DefaultTaskTimeout
	if taskTimeoutStr := os.Getenv("MANAGER_TASK_TIMEOUT"); taskTimeoutStr != "" {
		if val, err := strconv.Atoi(taskTimeoutStr); err == nil {
			taskTimeout = val
		}
	}

	watchInterval := DefaultTaskWatchInterval
	if watchIntervalStr := os.Getenv("MANAGER_WATCH_INTERVAL"); watchIntervalStr != "" {
		if val, err := strconv.Atoi(watchIntervalStr); err == nil {
			watchInterval = val
		}
	}

	return &Config{
		Port:             port,
		WorkersURLs:      urls,
		TasksStorageSize: storageSize,
		TaskTimeout:      time.Duration(taskTimeout) * time.Second,
		WatchInterval:    time.Duration(watchInterval) * time.Second,
	}, nil
}
