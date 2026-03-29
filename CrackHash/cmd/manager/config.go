package main

import (
	"crackhash/util"
	"strings"
	"time"
)

const (
	DefaultPort              = "8080"
	DefaultWorkerNodes       = "http://localhost:8081"
	DefaultTasksStorageSize  = 1000
	DefaultTaskTimeout       = 30 * time.Second
	DefaultTaskWatchInterval = 5 * time.Second
	DefaultReadTimeout       = 10 * time.Second
	DefaultWriteTimeout      = 10 * time.Second
	DefaultIdleTimeout       = 120 * time.Second
	DefaultClientTimeout     = 10 * time.Second
)

type Config struct {
	Port             string
	WorkerNodes      []string
	TasksStorageSize int
	TaskTimeout      time.Duration
	WatchInterval    time.Duration
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	IdleTimeout      time.Duration
	ClientTimeout    time.Duration
}

func LoadConfig() (*Config, error) {
	nodesEnv := util.ParseString("WORKER_NODES", DefaultWorkerNodes)
	var nodes []string
	for _, n := range strings.Split(nodesEnv, ",") {
		nodes = append(nodes, strings.TrimSpace(n))
	}

	return &Config{
		Port:             util.ParseString("PORT", DefaultPort),
		WorkerNodes:      nodes,
		TasksStorageSize: util.ParseInt("MANAGER_TASKS_STORAGE_SIZE", DefaultTasksStorageSize),
		TaskTimeout:      util.ParseDuration("MANAGER_TASK_TIMEOUT", DefaultTaskTimeout),
		WatchInterval:    util.ParseDuration("MANAGER_WATCH_INTERVAL", DefaultTaskWatchInterval),
		ReadTimeout:      util.ParseDuration("READ_TIMEOUT", DefaultReadTimeout),
		WriteTimeout:     util.ParseDuration("WRITE_TIMEOUT", DefaultWriteTimeout),
		IdleTimeout:      util.ParseDuration("IDLE_TIMEOUT", DefaultIdleTimeout),
		ClientTimeout:    util.ParseDuration("CLIENT_TIMEOUT", DefaultClientTimeout),
	}, nil
}
