package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/search-trends/internal/api"
	"github.com/yourusername/search-trends/internal/config"
	"github.com/yourusername/search-trends/internal/consumer"
	"github.com/yourusername/search-trends/internal/metrics"
	"github.com/yourusername/search-trends/internal/storage"
)

func main() {
	cfg := config.Load()

	store := storage.NewTrendStorage()
	metricsCollector := metrics.NewMetrics()

	cons, err := consumer.NewConsumer(cfg.NatsURL, cfg.NatsSubject, store, metricsCollector)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer cons.Close()

	if err := cons.Start(); err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	handler := api.NewHandler(store, metricsCollector)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/top", handler.GetTopQueries)
	mux.HandleFunc("/api/v1/stoplist", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetStopList(w, r)
		case http.MethodPost:
			handler.AddToStopList(w, r)
		case http.MethodDelete:
			handler.RemoveFromStopList(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/health", handler.HealthCheck)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		stats := metricsCollector.GetStats()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	go func() {
		log.Printf("HTTP server starting on port %s", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")
}
