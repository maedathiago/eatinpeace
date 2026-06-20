package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/maedathiago/eatinpeace/internal/application"
	"github.com/maedathiago/eatinpeace/internal/httpapi"
	"github.com/maedathiago/eatinpeace/internal/storage/memory"
)

func main() {
	store := memory.NewStore()
	service := application.NewService(store)
	if err := service.SeedPilotFixtures(context.Background()); err != nil {
		log.Fatalf("seed fixtures: %v", err)
	}

	addr := os.Getenv("EATINPEACE_HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           httpapi.NewHandler(service),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Eat in Peace API listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
