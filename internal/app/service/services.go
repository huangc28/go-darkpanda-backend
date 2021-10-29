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

func (r *RefundService) RefundCustomerIfRefundable(srv *models.Service, canceller *models.User) (bool, error) {
	refunded := false

	p, err := r.paymentDao.GetPaymentByServiceUuid(srv.Uuid.String)

	if err != nil {
		return refunded, err
	}

	if !p.Price.Valid {
		return refunded, fmt.Errorf("corrupted service data, service should have price")
	}

	if !p.PaymentID.Valid {
		return refunded, fmt.Errorf("corrupted service data, service should payment ID")
	}

	if !srv.AppointmentTime.Valid {
		return refunded, fmt.Errorf("corrupted service data, service should have appointment")
	}

	if isInAppointmentTimeBufferRange(srv.AppointmentTime.Time) && canceller.Gender == models.GenderMale {
		return false, nil
	}

	// Add amount back to the user balance and set payment refunded to be true.
	if err := r.refund(
		int(p.PaymentID.Int64),
		int(srv.CustomerID.Int32),
		decimal.NewFromFloat(p.Price.Float64),
	); err != nil {
		return refunded, err
	}

	return true, nil
}

// The below action should be atomic. The better way is to check if
func (r *RefundService) refund(paymentID, userID int, amount decimal.Decimal) error {
	if err := r.paymentDao.SetRefunded(paymentID); err != nil {
		return err
	}

	if err := r.userBalanceDao.AddBalance(userID, amount); err != nil {
		return err
	}

	return nil
}

func GetCancelCause(apt time.Time, gender models.Gender) models.CancelCause {
	if isInAppointmentTimeBufferRange(apt) {
		if gender == models.GenderFemale {
			return models.CancelCauseGirlCancelAfterAppointmentTime
		}

		return models.CancelCauseGuyCancelAfterAppointmentTime
	}

	if gender == models.GenderFemale {
		return models.CancelCauseGirlCancelBeforeAppointmentTime
	}

	return models.CancelCauseGuyCancelBeforeAppointmentTime
}

func isInAppointmentTimeBufferRange(apt time.Time) bool {
	now := time.Now()
	bufferEnd := apt.Add(30 * time.Minute)

	return now.After(apt) && apt.Before(bufferEnd)
}
