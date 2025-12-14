package main

import (
	"log"

	"github.com/Rudra775/reverse-proxy/pkg/proxy"
	"github.com/Rudra775/reverse-proxy/pkg/proxy/middleware"
)

func main() {
	cfg, err := proxy.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	p := proxy.New(cfg)

	// Middleware chain
	p.Use(middleware.RequestID())
	p.Use(middleware.Logging())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
