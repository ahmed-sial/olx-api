package main

import (
	"log"
	"os"

	"github.com/ahmed-sial/olx-api/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	log.Println("initializing database migrations")
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <up | down>")
	}

	cfg := config.MustLoad()

	log.Println("running database migrations")
	m, err := migrate.New("file://migrations", cfg.DbUrl)
	if err != nil {
		log.Fatalf("migration.new: %v", err)
	}

	switch os.Args[1] {
	case "up":
		if err := m.Up(); err != nil {
			log.Fatalf("migration.up: %v", err)
		}
	case "down":
		if err := m.Steps(-1); err != nil {
			log.Fatalf("migration.down: %v", err)
		}
	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}
	log.Println("database migrations run successful")
}
