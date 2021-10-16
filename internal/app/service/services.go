package service

import (
	"fmt"
	"time"

	"github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	"github.com/huangc28/go-darkpanda-backend/internal/app/models"
	"github.com/shopspring/decimal"
)

type RefundService struct {
	paymentDao     contracts.PaymentDAOer
	userBalanceDao contracts.UserBalancer
}

func NewRefundService(pd contracts.PaymentDAOer, ub contracts.UserBalancer) *RefundService {
	return &RefundService{
		paymentDao:     pd,
		userBalanceDao: ub,
	}
}

func (r *RefundService) RefundCustomerIfRefundable(srv *models.Service, canceller *models.User) (models.Cause, error) {
	cause := models.CauseNone

	if !srv.AppointmentTime.Valid {
		return models.CauseNone, fmt.Errorf("corrupted service data, service should have appointment")
	}

	p, err := r.paymentDao.GetPaymentByServiceUuid(srv.Uuid.String)

	if err != nil {
		return models.CauseNone, err
	}

	if !p.Price.Valid {
		return models.CauseNone, fmt.Errorf("corrupted service data, service should have price")
	}

	cause = getCause(srv.AppointmentTime.Time, canceller.Gender)

	if err := r.paymentDao.SetRefundCause(p.PaymentID, cause); err != nil {
		return models.CauseNone, err
	}

	if isInAppointmentTimeBufferRange(srv.AppointmentTime.Time) {
		if canceller.Gender == models.GenderFemale {
			// Add amount back to the user balance and set payment refunded to be true.
			if err := r.refund(
				int(srv.CustomerID.Int32),
				decimal.NewFromFloat(p.Price.Float64),
			); err != nil {
				return cause, err
			}

			return cause, nil
		}

		return cause, nil
	}

	return cause, nil
}

// The below action should be atomic. The better way is to check if
func (r *RefundService) refund(userID int, amount decimal.Decimal) error {
	if err := r.userBalanceDao.AddBalance(userID, amount); err != nil {
		return err
	}

	return nil
}

func getCause(apt time.Time, gender models.Gender) models.Cause {
	if isInAppointmentTimeBufferRange(apt) {
		if gender == models.GenderFemale {
			return models.CauseGirlCancelAfterAppointmentTime
		}

		return models.CauseGirlCancelAfterAppointmentTime
	}

	if gender == models.GenderFemale {
		return models.CauseGirlCancelBeforeAppointmentTime
	}

	return models.CauseGuyCancelBeforeAppointmentTime
}

func isInAppointmentTimeBufferRange(apt time.Time) bool {
	now := time.Now()
	bufferEnd := apt.Add(30 * time.Minute)

	return now.After(apt) && apt.Before(bufferEnd)
}
