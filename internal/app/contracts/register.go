package contracts

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type CheckUserInSMSWhiteListParams struct {
	RedisClient *redis.Client
	UserUuid    string
}

type Registerar interface {
	CheckUserInSMSWhiteList(ctx context.Context, p CheckUserInSMSWhiteListParams) (bool, error)
}
