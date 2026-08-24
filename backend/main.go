package main

import (
	"net/http"
	"time"
)

func main() {
	c := loadConfig()
	if e := startAndServe(c); e != nil {
		panic(e)
	}
}

func startAndServe(c Config) error {
	r := newRouter(newCycleStore())
	m := http.NewServeMux()
	m.Handle("/healthz", r)
	m.Handle("/api/", r)
	m.Handle("/", staticHandler())
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			// 后台自检
		}
	}()
	go func() {
		for {
			time.Sleep(10 * time.Minute)
		}
	}()
	return serveHTTP(newEnterpriseServer(":"+c.Port, m))
}
