package config

import (
	"crackhash/pkg/util"
	"time"
)

const (
	DefaultPort                   = "8081"
	DefaultReadTimeout            = 10 * time.Second
	DefaultWriteTimeout           = 10 * time.Second
	DefaultIdleTimeout            = 120 * time.Second
	DefaultClientTimeout          = 10 * time.Second
	DefaultContextCheckIterations = 10000
	DefaultStopOnFirstMatch       = true
	DefaultRabbitMQURL            = "amqp://guest:guest@localhost:5672/"
	DefaultTasksQueue             = "tasks_queue"
	DefaultResultsQueue           = "results_queue"
)

type Config struct {
	Port                   string
	ReadTimeout            time.Duration
	WriteTimeout           time.Duration
	IdleTimeout            time.Duration
	ContextCheckIterations int
	StopOnFirstMatch       bool
	RabbitMQURL            string
	TasksQueue             string
	ResultsQueue           string
}

func LoadConfig() (*Config, error) {
	return &Config{
		Port:                   util.ParseString("PORT", DefaultPort),
		ReadTimeout:            util.ParseDuration("READ_TIMEOUT", DefaultReadTimeout),
		WriteTimeout:           util.ParseDuration("WRITE_TIMEOUT", DefaultWriteTimeout),
		IdleTimeout:            util.ParseDuration("IDLE_TIMEOUT", DefaultIdleTimeout),
		ContextCheckIterations: util.ParseInt("CONTEXT_CHECK_ITERATIONS", DefaultContextCheckIterations),
		StopOnFirstMatch:       util.ParseBool("STOP_ON_FIRST_MATCH", DefaultStopOnFirstMatch),
		RabbitMQURL:            util.ParseString("RABBITMQ_URL", DefaultRabbitMQURL),
		TasksQueue:             util.ParseString("RABBITMQ_TASKS_QUEUE", DefaultTasksQueue),
		ResultsQueue:           util.ParseString("RABBITMQ_RESULTS_QUEUE", DefaultResultsQueue),
	}, nil
}
