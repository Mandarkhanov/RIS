package main

import (
	"os"
	"strconv"
	"time"
)

const (
	DefaultPort                 = "8081"
	DefaultManagerURL           = "http://localhost:8080/internal/api/manager/hash/crack/request"
	DefaultReadTimeout          = 10 * time.Second
	DefaultWriteTimeout         = 10 * time.Second
	DefaultIdleTimeout          = 120 * time.Second
	DefaultClientTimeout        = 10 * time.Second
	DefaultContextCheckInterval = int64(10000)
	DefaultStopOnFirstMatch     = false
)

type Config struct {
	Port                 string
	ManagerURL           string
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	IdleTimeout          time.Duration
	ClientTimeout        time.Duration
	ContextCheckInterval int64
	StopOnFirstMatch     bool
}

func LoadConfig() (*Config, error) {
	return &Config{
		Port:                 parseString("PORT", DefaultPort),
		ManagerURL:           parseString("MANAGER_URL", DefaultManagerURL),
		ReadTimeout:          parseDuration("READ_TIMEOUT", DefaultReadTimeout),
		WriteTimeout:         parseDuration("WRITE_TIMEOUT", DefaultWriteTimeout),
		IdleTimeout:          parseDuration("IDLE_TIMEOUT", DefaultIdleTimeout),
		ClientTimeout:        parseDuration("CLIENT_TIMEOUT", DefaultClientTimeout),
		ContextCheckInterval: parseInt64("CONTEXT_CHECK_INTERVAL", DefaultContextCheckInterval),
		StopOnFirstMatch:     parseBool("STOP_ON_FIRST_MATCH", DefaultStopOnFirstMatch),
	}, nil
}

func parseString(envKey string, defaultVal string) string {
	valStr := os.Getenv(envKey)
	if valStr == "" {
		return defaultVal
	}
	return valStr
}

func parseDuration(envKey string, defaultVal time.Duration) time.Duration {
	valStr := os.Getenv(envKey)
	if valStr == "" {
		return defaultVal
	}
	if valInt, err := strconv.Atoi(valStr); err == nil {
		return time.Duration(valInt) * time.Second
	}
	return defaultVal
}

func parseInt64(envKey string, defaultVal int64) int64 {
	valStr := os.Getenv(envKey)
	if valStr == "" {
		return defaultVal
	}
	if valInt, err := strconv.ParseInt(valStr, 10, 64); err == nil && valInt > 0 {
		return valInt
	}
	return defaultVal
}

func parseBool(envKey string, defaultVal bool) bool {
	valStr := os.Getenv(envKey)
	if valStr == "" {
		return defaultVal
	}
	if valBool, err := strconv.ParseBool(valStr); err == nil {
		return valBool
	}
	return defaultVal
}
