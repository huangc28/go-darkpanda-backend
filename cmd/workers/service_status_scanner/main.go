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
	dpfcm "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/firebase_messaging"
	"github.com/huangc28/go-darkpanda-backend/manager"
	log "github.com/sirupsen/logrus"

	logger "github.com/huangc28/go-darkpanda-backend/cmd/workers/loggers"
)

// We need to have a worker ticks every minute to check routinely on the value of `service_status` for each service record.
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

func init() {
	ctx := context.Background()
	manager.NewDefaultManager(ctx).Run(func() {

		if err := deps.Get().Run(); err != nil {
			log.Fatalf("failed to initialise dependency container %s", err.Error())
		}

		errLogPath := config.GetAppConf().ErrorLogPath
		infoLogPath := config.GetAppConf().InfoLogPath

		logger.InitErrLogger(errLogPath, "service_status_scanner")
		logger.InitInfoLogger(infoLogPath, "service_status_scanner")
	})
}

func ScanCompletedServices(srvDao contracts.ServiceDAOer) error {
	cplSrvs, err := srvDao.ScanCompletedServices()

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
		err := df.UpdateMultipleServiceStatus(
			ctx,
			darkfirestore.UpdateMultipleServiceStatusParams{
				ServiceUuids:  srvUuids,
				ServiceStatus: string(models.ServiceStatusCompleted),
			},
		)

		if err != nil {
			return fmt.Errorf("failed to scan completed services %s", err.Error())
		}

		// Emit FCM message to notify both parties that the service has completed
		var dpfcmer dpfcm.DPFirebaseMessenger
		depCon := deps.Get().Container
		depCon.Make(&dpfcmer)

		for _, cplSrv := range cplSrvs {
			// Send FCM to service provider
			if err := dpfcmer.PublishServiceCompletedNotification(
				ctx,
				dpfcm.ServiceCompletedMessage{
					Topic:               cplSrv.ServiceProvidersFCMTopic,
					CounterPartUsername: cplSrv.CustomerUsername,
					ServiceUUID:         cplSrv.UUID,
				},
			); err != nil {
				return err
			}

			// Send FCM to customer
			if err := dpfcmer.PublishServiceCompletedNotification(
				ctx,
				dpfcm.ServiceCompletedMessage{
					Topic:               cplSrv.CustomerFCMTopic,
					CounterPartUsername: cplSrv.ServiceProviderUsername,
					ServiceUUID:         cplSrv.UUID,
				},
			); err != nil {
				return err
			}
		}

		log.Printf("completed services %v", cplSrvs)
	}

	return nil
}

func ScanExpiredServices(srvDao contracts.ServiceDAOer) error {
	expSrvs, err := srvDao.ScanExpiredServices()

	// If error occurs, we write error logs into system log.
	if err != nil {
		return fmt.Errorf("failed to scan expired services %s", err.Error())
	}

	// Notify those chatroom that the service status has changed, perform batch write.
	ctx := context.Background()
	if len(expSrvs) > 0 {
		srvUuids := make([]string, 0)
		for _, srv := range expSrvs {
			srvUuids = append(srvUuids, srv.UUID)
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

		var dpfcmer dpfcm.DPFirebaseMessenger
		depCon := deps.Get().Container
		depCon.Make(&dpfcmer)

		for _, expSrv := range expSrvs {
			// Send to service provider
			if err := dpfcmer.PublishServiceExpiredNotification(
				ctx,
				dpfcm.ServiceExpiredMessage{
					Topic:               expSrv.ServiceProvidersFCMTopic,
					CounterPartUsername: expSrv.CustomerUsername,
					ServiceUUID:         expSrv.UUID,
				},
			); err != nil {
				return err
			}

			// Send to customer
			if err := dpfcmer.PublishServiceExpiredNotification(
				ctx,
				dpfcm.ServiceExpiredMessage{
					Topic:               expSrv.CustomerFCMTopic,
					CounterPartUsername: expSrv.ServiceProviderUsername,
					ServiceUUID:         expSrv.UUID,
				},
			); err != nil {
				return err
			}

		}
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

	quitSig := make(chan os.Signal, 1)
	signal.Notify(quitSig, syscall.SIGINT, syscall.SIGTERM)
	<-quitSig

	log.Info("graceful shutdown worker...")

	close(quitTicker)

	log.Info("worker shutdown complete")
}
