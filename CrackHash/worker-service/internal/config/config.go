package config

import (
	"crackhash/pkg/util"
	"time"
)

const (
	DefaultPort                   = "8081"
	DefaultManagerURL             = "http://localhost:8080"
	DefaultReadTimeout            = 10 * time.Second
	DefaultWriteTimeout           = 10 * time.Second
	DefaultIdleTimeout            = 120 * time.Second
	DefaultClientTimeout          = 10 * time.Second
	DefaultContextCheckIterations = 10000
	DefaultStopOnFirstMatch       = true
)

type Config struct {
	Port                    string
	ManagerURL              string
	ReadTimeout             time.Duration
	WriteTimeout            time.Duration
	IdleTimeout             time.Duration
	ClientTimeout           time.Duration
	ContextCheckInterations int
	StopOnFirstMatch        bool
}

func LoadConfig() (*Config, error) {
	return &Config{
		Port:                    util.ParseString("PORT", DefaultPort),
		ManagerURL:              util.ParseString("MANAGER_URL", DefaultManagerURL),
		ReadTimeout:             util.ParseDuration("READ_TIMEOUT", DefaultReadTimeout),
		WriteTimeout:            util.ParseDuration("WRITE_TIMEOUT", DefaultWriteTimeout),
		IdleTimeout:             util.ParseDuration("IDLE_TIMEOUT", DefaultIdleTimeout),
		ClientTimeout:           util.ParseDuration("CLIENT_TIMEOUT", DefaultClientTimeout),
		ContextCheckInterations: util.ParseInt("CONTEXT_CHECK_ITERATIONS", DefaultContextCheckIterations),
		StopOnFirstMatch:        util.ParseBool("STOP_ON_FIRST_MATCH", DefaultStopOnFirstMatch),
	}, nil
}
