package main

import (
	"log"
	"net"
	"net/http"

	"github.com/ahmed-sial/olx-api/internal/config"
	"github.com/ahmed-sial/olx-api/internal/db"
	"github.com/ahmed-sial/olx-api/internal/handlers"
)

func main() {
	cfg := config.MustLoad()

	log.Printf("database connection in progress")
	db, err := db.Connect(cfg.DbUrl)
	if err != nil {
		log.Fatalf("main.db.Connect: %v", err)
	}

	log.Printf("database connected successfully")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handlers.Health())
	mux.HandleFunc("GET /listings", handlers.List(db))

	server := &http.Server{
		Addr:    net.JoinHostPort(cfg.HttpServer.Address, cfg.HttpServer.Port),
		Handler: mux,
	}

	log.Printf("started HTTP server on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
