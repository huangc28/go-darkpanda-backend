package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	"github.com/huangc28/go-darkpanda-backend/manager"
	log "github.com/sirupsen/logrus"

	logger "github.com/huangc28/go-darkpanda-backend/cmd/workers/loggers"
)

// We need to have a worker ticks every minute to check routinely on the value of `inquiry_status` for each service_inquiries record.
// We have to check for the following senarios and change the `inquiry_status` to proper status.
//
// Inquiring:
//    If created_at exceed 5 hour, we consider the inqury has canceled. Hence,
//    we should set the inquiry status to `canceled`.
//
var (
	errLogger  = log.New()
	infoLogger = log.New()
)

func init() {
	ctx := context.Background()
	manager.NewDefaultManager(ctx).Run(func() {

		if err := deps.Get().Run(); err != nil {
			log.Fatalf("failed to initialise dependency container %s", err.Error())
		}

		errLogPath := config.GetAppConf().ErrorLogPath
		infoLogPath := config.GetAppConf().InfoLogPath

		logger.InitErrLogger(errLogPath, "service_inquiries_status_scanner")
		logger.InitInfoLogger(infoLogPath, "service_inquiries_status_scanner")
	})
}

func ScanInquiringServicesInquiries(srvDao contracts.ServiceDAOer) error {
	cplSrvs, err := srvDao.ScanInquiringServiceInquiries()

	if err != nil {
		return fmt.Errorf("failed to scan completed services %s", err.Error())
	}

	if len(cplSrvs) > 0 {
		ctx := context.Background()
		srvUuids := make([]string, 0)

		for _, srv := range cplSrvs {
			srvUuids = append(srvUuids, srv.UUID)
		}

		df := darkfirestore.Get()
		err := df.UpdateMultipleInquiryStatus(
			ctx,
			darkfirestore.UpdateMultipleInquiryStatusParams{
				InquiryUuids: srvUuids,
				Status:       string(models.InquiryStatusCanceled),
			},
		)

		if err != nil {
			return fmt.Errorf("failed to scan completed services %s", err.Error())
		}

		log.Printf("completed services %v", cplSrvs)
	}

	return nil
}

func main() {
	tickSec := 60
	tickSecEnv := os.Getenv("TICK_INTERVAL_IN_SECOND")

	if len(tickSecEnv) > 0 {
		tickSecEnvInt, err := strconv.Atoi(tickSecEnv)

		if err != nil {
			tickSec = tickSecEnvInt
		}
	}

	ticker := time.NewTicker(time.Duration(tickSec) * time.Second)

	quitTicker := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:

				depCon := deps.Get().Container
				var serviceDao contracts.ServiceDAOer

				depCon.Make(&serviceDao)

				if err := ScanInquiringServicesInquiries(serviceDao); err != nil {
					errLogger.Error(err)
				}

			case <-quitTicker:
				ticker.Stop()

				return
			}
		}
	}()

	quitSig := make(chan os.Signal, 1)
	signal.Notify(quitSig, syscall.SIGINT, syscall.SIGTERM)
	<-quitSig

	log.Info("graceful shutdown worker...")

	close(quitTicker)

	log.Info("worker shutdown complete")
}
