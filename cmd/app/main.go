package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huangc28/go-darkpanda-backend/internal/app"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/manager"
	"github.com/prometheus/common/log"
	"github.com/spf13/viper"
)

func main() {
	manager.
		NewDefaultManager().
		Run(func() {
			// Initialize IoC container.
			deps.Get().Run()

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
