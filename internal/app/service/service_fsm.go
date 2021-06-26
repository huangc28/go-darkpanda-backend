package service

import (
	cinternal "github.com/golobby/container/pkg/container"
	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/looplab/fsm"
)

type ServiceActions string

var (
	Paid         ServiceActions = "paid"
	PayFailed    ServiceActions = "pay_failed"
	Cancel       ServiceActions = "cancel"
	StartService ServiceActions = "start_service"
	Complete     ServiceActions = "complete"
	TooEarly     ServiceActions = "too_early"
	Expired      ServiceActions = "expired"
)

func (a *ServiceActions) ToString() string {
	return string(*a)
}

func NewServiceFSM(initial models.ServiceStatus) *fsm.FSM {
	f := fsm.NewFSM(
		initial.ToString(),
		fsm.Events{
			{
				Name: Paid.ToString(),
				Src: []string{
					string(models.ServiceStatusUnpaid),
				},
				Dst: string(models.ServiceStatusToBeFulfilled),
			},
			{
				Name: PayFailed.ToString(),
				Src: []string{
					string(models.ServiceStatusUnpaid),
				},
				Dst: string(models.ServiceStatusPaymentFailed),
			},
			{
				Name: Cancel.ToString(),
				Src: []string{
					string(models.ServiceStatusToBeFulfilled),
				},
				Dst: string(models.ServiceStatusCanceled),
			},
			{
				Name: StartService.ToString(),
				Src: []string{
					string(models.ServiceStatusToBeFulfilled),
				},
				Dst: string(models.ServiceStatusFulfilling),
			},
			{
				Name: Complete.ToString(),
				Src: []string{
					string(models.ServiceStatusFulfilling),
				},
				Dst: string(models.ServiceStatusCompleted),
			},
			{
				Name: Expired.ToString(),
				Src: []string{
					string(models.ServiceStatusToBeFulfilled),
				},
				Dst: string(models.ServiceStatusExpired),
			},
		},
		fsm.Callbacks{},
	)

	return f
}

func ServiceServiceProvider(c cinternal.Container) func() error {
	return func() error {
		c.Transient(func() contracts.ServiceFSMer {
			return NewServiceFSM(models.ServiceStatusUnpaid)
		})

		return nil
	}
}
