package config

import (
	"crackhash/pkg/util"
	"time"
)

const (
	DefaultPort              = "8080"
	DefaultTaskTimeout       = 30 * time.Second
	DefaultTaskWatchInterval = 5 * time.Second
	DefaultReadTimeout       = 10 * time.Second
	DefaultWriteTimeout      = 10 * time.Second
	DefaultIdleTimeout       = 120 * time.Second
	DefaultMongoDBURL        = "mongodb://localhost:27017"
	DefaultMongoDBName       = "crackhash"
	DefaultRabbitMQURL       = "amqp://guest:guest@localhost:5672/"
	DefaultTasksQueue        = "tasks_queue"
	DefaultResultsQueue      = "results_queue"
	DefaultTaskPartsCount    = 3
)

type Config struct {
	Port           string
	TaskTimeout    time.Duration
	WatchInterval  time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MongoURL       string
	MongoDBName    string
	RabbitMQURL    string
	TasksQueue     string
	ResultsQueue   string
	TaskPartsCount int
}

func LoadConfig() (*Config, error) {
	return &Config{
		Port:           util.ParseString("PORT", DefaultPort),
		TaskTimeout:    util.ParseDuration("MANAGER_TASK_TIMEOUT", DefaultTaskTimeout),
		WatchInterval:  util.ParseDuration("MANAGER_WATCH_INTERVAL", DefaultTaskWatchInterval),
		ReadTimeout:    util.ParseDuration("READ_TIMEOUT", DefaultReadTimeout),
		WriteTimeout:   util.ParseDuration("WRITE_TIMEOUT", DefaultWriteTimeout),
		IdleTimeout:    util.ParseDuration("IDLE_TIMEOUT", DefaultIdleTimeout),
		MongoURL:       util.ParseString("MONGODB_URL", DefaultMongoDBURL),
		MongoDBName:    util.ParseString("MONGODB_NAME", DefaultMongoDBName),
		RabbitMQURL:    util.ParseString("RABBITMQ_URL", DefaultRabbitMQURL),
		TasksQueue:     util.ParseString("RABBITMQ_TASKS_QUEUE", DefaultTasksQueue),
		ResultsQueue:   util.ParseString("RABBITMQ_RESULTS_QUEUE", DefaultResultsQueue),
		TaskPartsCount: util.ParseInt("TASK_PARTS_COUNT", DefaultTaskPartsCount),
	}, nil
}
