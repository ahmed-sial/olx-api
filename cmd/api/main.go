package main

import (
	"log"
	"net"
	"net/http"

	"github.com/ahmed-sial/olx-api/internal/config"
)

func main() {
	cfg := config.MustLoad()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	server := &http.Server{
		Addr:    net.JoinHostPort(cfg.HttpServer.Address, cfg.HttpServer.Port),
		Handler: mux,
	}

	log.Printf("starting HTTP server on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
