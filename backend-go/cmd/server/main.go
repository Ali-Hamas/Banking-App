package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/demobank/atm-auth/internal/config"
	"github.com/demobank/atm-auth/internal/httpapi"
	"github.com/demobank/atm-auth/internal/realtime"
	"github.com/demobank/atm-auth/internal/session"
)

func main() {
	cfg := config.Load()
	store := session.NewStore()
	hub := realtime.NewHub()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go store.RunSweeper(ctx)

	handlers := &httpapi.Handlers{Cfg: cfg, Store: store, Hub: hub}
	router := httpapi.NewRouter(handlers, hub)

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("auth middleware listening on %s", addr)
		log.Printf("ws endpoint  ws://localhost%s/ws?sid=<sessionId>", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs
	log.Println("shutdown signal received, draining…")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("bye")
}
