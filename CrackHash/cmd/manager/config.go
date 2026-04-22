package main

import (
	"crackhash/util"
	"strings"
	"time"
)

const (
	DefaultPort              = "8080"
	DefaultWorkerURL         = "http://localhost:8081"
	DefaultTaskTimeout       = 30 * time.Second
	DefaultTaskWatchInterval = 5 * time.Second
	DefaultReadTimeout       = 10 * time.Second
	DefaultWriteTimeout      = 10 * time.Second
	DefaultIdleTimeout       = 120 * time.Second
	DefaultClientTimeout     = 10 * time.Second
	DefaultMongoDBURL        = "mongodb://localhost:27017"
	DefaultMongoDBName       = "crackhash"
)

type Config struct {
	Port             string
	WorkerURLs       []string
	TasksStorageSize int
	TaskTimeout      time.Duration
	WatchInterval    time.Duration
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	IdleTimeout      time.Duration
	ClientTimeout    time.Duration
	MongoURL         string
	MongoDBName      string
}

func LoadConfig() (*Config, error) {
	urlsEnv := util.ParseString("WORKER_URLS", DefaultWorkerURL)
	var urls []string
	for _, n := range strings.Split(urlsEnv, ",") {
		urls = append(urls, strings.TrimSpace(n))
	}

	return &Config{
		Port:          util.ParseString("PORT", DefaultPort),
		WorkerURLs:    urls,
		TaskTimeout:   util.ParseDuration("MANAGER_TASK_TIMEOUT", DefaultTaskTimeout),
		WatchInterval: util.ParseDuration("MANAGER_WATCH_INTERVAL", DefaultTaskWatchInterval),
		ReadTimeout:   util.ParseDuration("READ_TIMEOUT", DefaultReadTimeout),
		WriteTimeout:  util.ParseDuration("WRITE_TIMEOUT", DefaultWriteTimeout),
		IdleTimeout:   util.ParseDuration("IDLE_TIMEOUT", DefaultIdleTimeout),
		ClientTimeout: util.ParseDuration("CLIENT_TIMEOUT", DefaultClientTimeout),
		MongoURL:      util.ParseString("MONGODB_URL", DefaultMongoDBURL),
		MongoDBName:   util.ParseString("MONGODB_NAME", DefaultMongoDBName),
	}, nil
}
