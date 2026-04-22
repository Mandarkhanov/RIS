package util

import (
	"log"
	"os"
	"strconv"
	"time"
)

func ParseString(envKey string, defaultVal string) string {
	valStr, exists := os.LookupEnv(envKey)
	if !exists {
		log.Printf("Environment variable \"%s\" not set. Using default value: %s\n", envKey, defaultVal)
		return defaultVal
	}
	return valStr
}

func ParseDuration(envKey string, defaultVal time.Duration) time.Duration {
	valStr, exists := os.LookupEnv(envKey)
	if !exists {
		log.Printf("Environment variable \"%s\" not set. Using default value: %s\n", envKey, defaultVal)
		return defaultVal
	}

	duration, err := time.ParseDuration(valStr)
	if err != nil {
		log.Printf("Invalid format for environment variable \"%s\": \"%s\". Using default value: %s. Error: %v\n", envKey, valStr, defaultVal, err)
		return defaultVal
	}
	return duration
}

func ParseInt(envKey string, defaultVal int) int {
	valStr, exists := os.LookupEnv(envKey)
	if !exists {
		log.Printf("Environment variable \"%s\" not set. Using default value: %s\n", envKey, int(defaultVal))
		return defaultVal
	}
	if valInt, err := strconv.ParseInt(valStr, 10, 64); err == nil && valInt > 0 {
		return int(valInt)
	}
	return defaultVal
}

func ParseBool(envKey string, defaultVal bool) bool {
	valStr, exists := os.LookupEnv(envKey)
	if !exists {
		log.Printf("Environment variable \"%s\" not set. Using default value: %s\n", envKey, bool(defaultVal))
		return defaultVal
	}
	if valBool, err := strconv.ParseBool(valStr); err == nil {
		return valBool
	}
	return defaultVal
}
