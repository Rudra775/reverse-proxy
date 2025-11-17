package main

import (
	"log"

	"proxyserver/internal/config"
	"proxyserver/internal/proxy"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	router := proxy.NewRouter(cfg)
	handler := proxy.NewProxyHandler(cfg, router)
	server := proxy.NewServer(cfg, handler)

	if err := server.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
