package redishub

import (
	"github.com/9d77v/wspush/utils"
	redis "github.com/go-redis/redis/v8"
)

var (
	//Hub ...
	Hub           *RedisHub
	redisAddress  = utils.GetEnvStr("REDIS_ADDRESS", "domain.local:6379")
	redisPassword = utils.GetEnvStr("REDIS_PASSWORD", "")
)

func init() {
	r := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: redisPassword,
		DB:       0, // use default DB
	})
	Hub = NewHub(r)
}
