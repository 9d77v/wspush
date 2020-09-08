package redishub

import (
	"context"
	"encoding/json"
	"log"

	redis "github.com/go-redis/redis/v8"

	"github.com/9d77v/wspush/hub"
)

//RedisHub ..
type RedisHub struct {
	*hub.Hub
	PubSub *redis.PubSub
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
			writeData := make([]map[string]interface{}, 0)
			json.Unmarshal([]byte(msg.Payload), &writeData)
			if v[msg.Channel] {
				err := k.WriteJSON(writeData)
				if err != nil {
					log.Println("websocket write error:", err)
					continue
				}
			}
		}
	}
}
