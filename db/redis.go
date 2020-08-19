package db

import (
	"context"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

var rds *redis.Client

type RedisConf struct {
	Addr     string
	Password string
	DB       int
}

func InitRedis(conf RedisConf) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password, // no password set
		DB:       conf.DB,       // use default DB
	})

	ctx := context.Background()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.
			WithFields(log.Fields{
				"addr": conf.Addr,
			}).
			Debugf("error occur when initializing redis client %s", err.Error())
		return err
	}

	rds = rdb

	log.Infof("redis connected on: %s", conf.Addr)

	return nil
}

func GetRedis() *redis.Client {
	return rds
}
