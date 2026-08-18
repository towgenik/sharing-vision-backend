// Command api runs the Sharing Vision article microservice.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/towgenik/sharing-vision-backend/internal/config"
	"github.com/towgenik/sharing-vision-backend/internal/db"
	"github.com/towgenik/sharing-vision-backend/internal/httpapi"
	"github.com/towgenik/sharing-vision-backend/internal/post"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	pool, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	repo := &post.MySQLRepository{DB: pool}
	router := httpapi.NewRouter(repo)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	// Graceful shutdown on SIGINT/SIGTERM (systemd/container stop).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("api listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
