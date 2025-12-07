package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

// simple handler returning identifier and latency
func createHandler(port int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		delay := rand.Intn(500) // up to 500ms random latency
		time.Sleep(time.Duration(delay) * time.Millisecond)

		msg := fmt.Sprintf(
			"[Backend %d] path=%s delay=%dms time=%s\n",
			port, r.URL.Path, delay, time.Now().Format(time.RFC3339Nano),
		)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(msg))
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run server.go <port>")
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid port: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", createHandler(port))

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Backend server running on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
