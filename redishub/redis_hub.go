package redishub

import (
	"context"
	"log"
	"strings"

	"github.com/9d77v/go-pkg/cache/redis"
	v8 "github.com/go-redis/redis/v8"

	"github.com/gorilla/websocket"

	"github.com/9d77v/wspush/hub"
)

var (
	//Hub ...
	Hub *RedisHub
)

func init() {
	Hub = NewHub(redis.GetClient())
}

//RedisHub ..
type RedisHub struct {
	*hub.Hub
	PubSub *v8.PubSub
}

//NewHub Hub初始化
func NewHub(redis *redis.Client) *RedisHub {
	defer log.Println("server start successed.")
	var ctx = context.Background()
	pubsub := redis.PSubscribe(ctx, "ping")
	newHub := hub.NewHub(ctx, pubsub.PSubscribe, pubsub.PUnsubscribe)
	redisHub := &RedisHub{
		newHub,
		pubsub,
	}
	go redisHub.SendData()
	return redisHub
}

//SendData ..
func (h *RedisHub) SendData() {
	log.Println("SendData Goroutine Ready.")
	ch := h.PubSub.Channel()
	for msg := range ch {
		for k, v := range h.ClientMap {
			if v[msg.Channel] {
				err := k.WriteMessage(websocket.BinaryMessage, []byte(msg.Payload))
				if err != nil {
					log.Println("websocket write error:", err)
					continue
				}
			}
			for sub := range v {
				if strings.Contains(sub, "*") {
					matchStr := strings.Trim(sub, "*")
					if strings.Contains(msg.Channel, matchStr) {
						err := k.WriteMessage(websocket.BinaryMessage, []byte(msg.Payload))
						if err != nil {
							log.Println("websocket write error:", err)
							continue
						}
					}
				}
			}
		}
	}
}
