package contracts

import "github.com/huangc28/go-darkpanda-backend/internal/app/models"

type ReferralDaoer interface {
	GetByRefCode(refCode string, fields []string) (*models.UserRefcode, error)
}
