package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	logger "github.com/huangc28/go-darkpanda-backend/cmd/workers/loggers"
	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/deps"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	darkfirestore "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/dark_firestore"
	dpfcm "github.com/huangc28/go-darkpanda-backend/internal/app/pkg/firebase_messaging"
	"github.com/huangc28/go-darkpanda-backend/manager"
)

// This worker checks if service is paid within 30 minutes from the time being booked.
func init() {
	ctx := context.Background()
	manager.NewDefaultManager(ctx).Run(func() {
		if err := deps.Get().Run(); err != nil {
			log.Fatalf("failed to initialise dependency container %s", err.Error())
		}

		errLogPath := config.GetAppConf().ErrorLogPath
		infoLogPath := config.GetAppConf().InfoLogPath

		logger.InitErrLogger(errLogPath, "service_payment_checker")
		logger.InitInfoLogger(infoLogPath, "service_payment_checker")
	})
}

const TickInterval = 2 * time.Second

func main() {
	// Retrieve all services with status "unpaid"
	ticker := time.NewTicker(TickInterval)

	quitTicker := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:

				var srvDao contracts.ServiceDAOer
				deps.Get().Container.Make(&srvDao)

				if err := ScanAndUpdateUnpaidServiceExceed30Minutes(srvDao); err != nil {
					logger.GetErrorLogger().Error(err)
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

// If male user did not pay for the booked service within 30 minutes,
// we mark the service as `payment failed`. Moreover, we will notify girl
// via firestore message that the service has been canceled.
func ScanAndUpdateUnpaidServiceExceed30Minutes(srvDao contracts.ServiceDAOer) error {
	ms, err := srvDao.CancelUnpaidServicesIfExceed30Minuties()

	if err != nil {
		return err
	}

	if len(ms) == 0 {
		return nil
	}

	logger.GetInfoLogger().Infof("update expired unpaid services %v", ms)

	ctx := context.Background()
	var fcm dpfcm.DPFirebaseMessenger
	deps.Get().Container.Make(&fcm)

	uuids := make([]string, 0)

	for _, m := range ms {

		// Send FCM message to notify both service provider and customer
		// that the service has been canceled due to expired unpaid service.
		// Send to multiple services related personale.
		if m.CustomerFCMTopic != nil {
			if err := fcm.PublishUnpaidServiceExpiredNotification(ctx, dpfcm.PublishUnpaidServiceExpiredMessage{
				Topic:               *m.CustomerFCMTopic,
				ServiceUUID:         m.Uuid.String,
				CustomerName:        m.CustomerName,
				ServiceProviderName: m.ServiceProviderName,
			}); err != nil {
				return fmt.Errorf("failed to send to topic %s %v", *m.CustomerFCMTopic, err.Error())
			}
		}

		if m.ServiceProviderFCMTopic != nil {
			if err := fcm.PublishUnpaidServiceExpiredNotification(ctx, dpfcm.PublishUnpaidServiceExpiredMessage{
				Topic:               *m.ServiceProviderFCMTopic,
				ServiceUUID:         m.Uuid.String,
				CustomerName:        m.CustomerName,
				ServiceProviderName: m.ServiceProviderName,
			}); err != nil {
				return fmt.Errorf("failed to send to topic %s %v", *m.CustomerFCMTopic, err.Error())
			}
		}

		uuids = append(uuids, m.Uuid.String)
	}

	df := darkfirestore.Get()

	if err := df.UpdateMultipleServiceStatus(
		ctx,
		darkfirestore.UpdateMultipleServiceStatusParams{
			ServiceUuids:  uuids,
			ServiceStatus: string(models.ServiceStatusPaymentFailed),
		},
	); err != nil {
		return err
	}

	return nil
}
