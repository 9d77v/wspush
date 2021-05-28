package redishub

import (
	"strings"

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
	var r redis.UniversalClient
	addresses := strings.Split(redisAddress, ",")
	if len(addresses) == 1 {
		r = redis.NewClient(&redis.Options{
			Addr:     addresses[0],
			Password: redisPassword,
		})
	} else {
		r = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    addresses,
			Password: redisPassword,
		})
	}
	Hub = NewHub(r)
}
