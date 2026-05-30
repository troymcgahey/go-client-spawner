package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/troymcgahey/go-client-spawner/internal/poller"
)

type Config struct {
	DownstreamURL      string
	PollIntervalValSec int
}

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := LoadConfig()

	log.Println(cfg.DownstreamURL)
	log.Println(cfg.PollIntervalValSec)

	p := poller.NewPoller(
		cfg.DownstreamURL,
		time.Duration(cfg.PollIntervalValSec)*time.Second,
	)

	go p.Start(ctx)

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	go func() {
		log.Println("server listening on :8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("shutting down server")
	server.Shutdown(shutdownCtx)
}

func LoadConfig() Config {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	DownstreamURL := os.Getenv("DOWNSTREAM_URL")
	PollIntervalValSec, err := strconv.Atoi(os.Getenv("POLL_INTERVAL_SECONDS"))

	return Config{
		DownstreamURL:      DownstreamURL,
		PollIntervalValSec: PollIntervalValSec,
	}

}
