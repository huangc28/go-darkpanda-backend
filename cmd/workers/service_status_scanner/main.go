package main

import (
	"context"
	"errors"
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
)

// We need to have a worker ticks every minute and checks routinely on the value of `service_status` for each service record.
// We have to check for the following senarios and change the `service_status` to proper status.
//
// Expired:
//    If current time is greater than the `start_time + buffer time`, we consider the service has expired. Hence,
//    we should set the service status to `expired`.
//
// Completed:
//    If service status is `fulfilling` and current time is greater or equal to `end_time`, we will set the
//    `service_status` of the service to be `completed`.
var (
	errLogger  = log.New()
	infoLogger = log.New()
)

func initErrLogger() {
	errLogPath := config.GetAppConf().ErrorLogPath

	if err := os.MkdirAll(errLogPath, os.ModePerm); err != nil {
		log.Fatalf("failed to create file: %v", err)
	}

	errLogger.SetFormatter(&log.JSONFormatter{})

	file, err := os.OpenFile(
		fmt.Sprintf("%s/%s_service_status_scanner_error.log", errLogPath, time.Now().Format("01-02-2006")),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0755,
	)

	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	errLogger.Out = file
	errLogger.SetLevel(log.ErrorLevel)
}

func initInfoLogger() {
	infoLogPath := config.GetAppConf().InfoLogPath

	if err := os.MkdirAll(infoLogPath, os.ModePerm); err != nil {
		log.Fatalf("failed to create file: %v", err)
	}

	errLogger.SetFormatter(&log.JSONFormatter{})

	file, err := os.OpenFile(
		fmt.Sprintf("%s/%s_service_status_scanner_info.log", infoLogPath, time.Now().Format("01-02-2006")),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0755,
	)

	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	infoLogger.Out = file
	infoLogger.SetLevel(log.InfoLevel)
}

func init() {
	ctx := context.Background()
	manager.NewDefaultManager(ctx).Run(func() {
		initErrLogger()

		initInfoLogger()
	})
}

func ScanCompletedServices(srvDao contracts.ServiceDAOer) error {
	cplSrvs, err := srvDao.ScanCompletedServices()

	if err != nil {
		return errors.New(
			fmt.Sprintf("Failed to scan completed services %s", err.Error()),
		)
	}

	if len(cplSrvs) > 0 {
		ctx := context.Background()
		srvUuids := make([]string, 0)

		for _, srv := range cplSrvs {
			srvUuids = append(srvUuids, srv.Uuid.String())
		}

		df := darkfirestore.Get()
		err := df.UpdateMultipleServiceStatus(
			ctx,
			darkfirestore.UpdateMultipleServiceStatusParams{
				ServiceUuids:  srvUuids,
				ServiceStatus: string(models.ServiceStatusCompleted),
			},
		)

		if err != nil {
			return errors.New(fmt.Sprintf("Failed to scan completed services %s", err.Error()))
		}
	}

	infoLogger.Info("Done update service status to completed")

	return nil
}

func ScanExpiredServices(srvDao contracts.ServiceDAOer) error {
	expSrvs, err := srvDao.ScanExpiredServices()

	// If error occurs, we write error logs into system log.
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to scan expired services %s", err.Error()))
	}

	// Notify those chatroom that the service status has changed, perform batch write.
	ctx := context.Background()
	if len(expSrvs) > 0 {
		srvUuids := make([]string, 0)
		for _, srv := range expSrvs {
			srvUuids = append(srvUuids, srv.Uuid.String())
		}

		df := darkfirestore.Get()
		err := df.UpdateMultipleServiceStatus(
			ctx,
			darkfirestore.UpdateMultipleServiceStatusParams{
				ServiceUuids:  srvUuids,
				ServiceStatus: string(models.ServiceStatusExpired),
			},
		)

		if err != nil {
			return errors.New(fmt.Sprintf("Failed to update service status to expired in firestore %s", err.Error()))
		}

	}

	infoLogger.Info("Done update service status to expired")

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
	if err := deps.Get().Run(); err != nil {
		log.Fatalf("failed to initialize dependency container %s", err.Error())
	}

	quitTicker := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				depCon := deps.Get().Container
				var serviceDao contracts.ServiceDAOer

				depCon.Make(&serviceDao)

				if err := ScanExpiredServices(serviceDao); err != nil {
					errLogger.Error(err)
				}

				if err := ScanCompletedServices(serviceDao); err != nil {
					errLogger.Error(err)
				}

			case <-quitTicker:
				ticker.Stop()

				return
			}
		}
	}()

	quitSig := make(chan os.Signal)
	signal.Notify(quitSig, syscall.SIGINT, syscall.SIGTERM)
	<-quitSig

	log.Info("graceful shutdown worker...")

	close(quitTicker)

	log.Info("worker shutdown complete")
}