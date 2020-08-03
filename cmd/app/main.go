package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/huangc28/go-darkpanda-backend/internal/auth"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func main() {
	r := mux.NewRouter()
	verRouter := r.PathPrefix("/v1").Subrouter()

	auth.Routes(verRouter)

	http.Handle("/", r)

	srv := &http.Server{
		Handler: r,
		Addr:    ":3000",

		// Good practice: enforce timeouts for servers created.
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		log.Infof("listen on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("graceful shutdown...")

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown with error: %s", err.Error())
	}

	log.Info("shutdown complete..")
}
