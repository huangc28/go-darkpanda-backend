package contracts

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type CheckUserInSMSWhiteListParams struct {
	RedisClient *redis.Client
	Username    string
}

type Registerar interface {
	CheckUserInSMSWhiteList(ctx context.Context, p CheckUserInSMSWhiteListParams) (bool, error)
}
