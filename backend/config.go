package main

import (
	"os"
	"strconv"
)

type Config struct {
	Port          string
	HealthHistory int
}

func loadConfig() Config {
	p := os.Getenv("PORT")
	if p == "" {
		p = "8080"
	}
	limit, _ := strconv.Atoi(os.Getenv("HEALTH_HISTORY"))
	return Config{Port: p, HealthHistory: limit}
}
