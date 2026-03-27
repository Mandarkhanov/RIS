package main

import "os"

type Config struct {
	Port       string
	ManagerURL string
}

func LoadConfig() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	managerURL := os.Getenv("MANAGER_URL")
	if managerURL == "" {
		managerURL = "http://localhost:8080/internal/api/manager/hash/crack/request"
	}

	return &Config{
		Port:       port,
		ManagerURL: managerURL,
	}, nil
}
