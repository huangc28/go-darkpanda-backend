package register

import (
	"context"
)

type RegisterService struct {
	dao *RegisterDAO
}

func NewRegisterService(dao *RegisterDAO) *RegisterService {
	return &RegisterService{
		dao: dao,
	}
}

type ErrorCode string

var (
	ReferralCodeNotExists ErrorCode = "ReferralCodeNotExists"
	ReferralCodeOccupied  ErrorCode = "ReferralCodeOccupied"
	ReferralCodeExpired   ErrorCode = "ReferralCodeExpired"
)

type ValidateReferralCodeError struct {
	ErrCode    ErrorCode
	ErrMessage string
}

func (e *ValidateReferralCodeError) Error() string {
	switch e.ErrCode {
	case ReferralCodeNotExists:
		return "referral code does not exist"
	case ReferralCodeOccupied:
		return "referral code is occupied"
	case ReferralCodeExpired:
		return "referral code has expired"
	default:
		return e.ErrMessage
	}

}

// ValidateReferralCode checks if the given referral code
// passes the following conditions:
//   - referral code exists
//   - it's not occupied by other user
//   - referral code is not expired
func (s *RegisterService) ValidateReferralCode(ctx context.Context, refCode string) error {
	exists, err := s.dao.CheckReferCodeExists(ctx, refCode)

	if err != nil {
		return &ValidateReferralCodeError{
			ErrMessage: err.Error(),
		}
	}

	if !exists {
		return &ValidateReferralCodeError{
			ErrCode: ReferralCodeNotExists,
		}
	}

	m, err := s.dao.GetReferralCodeByReferralCode(refCode)

	if err != nil {
		return &ValidateReferralCodeError{
			ErrMessage: err.Error(),
		}
	}

	if m.InviteeID.Valid {
		return &ValidateReferralCodeError{
			ErrCode: ReferralCodeOccupied,
		}
	}

	return nil
}
