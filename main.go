package main

import "net/http"

func main() {
	c := loadConfig()
	r := newRouter(newCycleStore())
	m := http.NewServeMux()
	m.Handle("/healthz", r)
	m.Handle("/api/", r)
	m.Handle("/", staticHandler())
	if e := serveAddress(":"+c.Port, m); e != nil {
		panic(e)
	}
}
