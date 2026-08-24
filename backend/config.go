package main

import "os"

type Config struct{ Port string }

func loadConfig() Config {
	p := os.Getenv("PORT")
	if p == "" {
		p = "8080"
	}
	return Config{Port: p}
}
