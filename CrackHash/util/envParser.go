package util

import (
	"os"
	"strconv"
	"time"
)

func ParseString(envKey string, defaultVal string) string {
	valStr := os.Getenv(envKey)
	if valStr == "" {
		return defaultVal
	}
	return valStr
}

func ParseDuration(envKey string, defaultVal time.Duration) time.Duration {
	valStr := os.Getenv(envKey)
	if valStr == "" {
		return defaultVal
	}
	if valInt, err := strconv.Atoi(valStr); err == nil {
		return time.Duration(valInt) * time.Second
	}
	return defaultVal
}

func ParseInt(envKey string, defaultVal int) int {
	valStr := os.Getenv(envKey)
	if valStr == "" {
		return defaultVal
	}
	if valInt, err := strconv.ParseInt(valStr, 10, 64); err == nil && valInt > 0 {
		return int(valInt)
	}
	return defaultVal
}

func ParseBool(envKey string, defaultVal bool) bool {
	valStr := os.Getenv(envKey)
	if valStr == "" {
		return defaultVal
	}
	if valBool, err := strconv.ParseBool(valStr); err == nil {
		return valBool
	}
	return defaultVal
}
