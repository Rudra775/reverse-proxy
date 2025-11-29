package main

import (
	"log"

	"github.com/Rudra775/reverse-proxy/pkg/proxy"
)

func main() {
	cfg, err := proxy.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	p := proxy.New(cfg)

	if err := p.Start(); err != nil {
		log.Fatalf("proxy error: %v", err)
	}
}
