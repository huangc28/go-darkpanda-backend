package contracts

import "context"

type AuthDaoer interface {
	IsTokenInvalid(ctx context.Context, jwtToken string) (bool, error)
}
