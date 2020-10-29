package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/internal/app"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/spf13/viper"
)

func main() {
	iniCtx := context.Background()

	manager.
		NewDefaultManager(iniCtx).
		Run(func() {
			// Initialize IoC container.
			if err := deps.Get().Run(); err != nil {
				log.Fatalf("failed to initialize dependency container %s", err.Error())
			}

			fireClient := darkfirestore.Get()
			_, _, err := fireClient.Collection("users").Add(context.Background(), map[string]interface{}{
				"name": "jason huang",
			})

			if err != nil {
				log.Fatal(err)
			}

			r := gin.New()
			app.StartApp(r)

			srv := &http.Server{
				Handler: r,
				Addr:    fmt.Sprintf(":%d", viper.GetInt("port")),

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
		})
}
