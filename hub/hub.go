package hub

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

//Hub  receive mq data and user websocket to push to browser
// 支持channel不固定，订阅多个channel
type Hub struct {
	channelMap map[string][]*websocket.Conn
	ClientMap  map[*websocket.Conn]map[string]bool
	channels   []string
	lock       *sync.Mutex
	SubFunc    func(context.Context, ...string) error
	UnSubFunc  func(context.Context, ...string) error
	ctx        context.Context
}

//NewHub ..
func NewHub(ctx context.Context, subfunc func(context.Context, ...string) error, unsubfunc func(context.Context, ...string) error) *Hub {
	return &Hub{
		channelMap: make(map[string][]*websocket.Conn),
		ClientMap:  make(map[*websocket.Conn]map[string]bool),
		lock:       new(sync.Mutex),
		SubFunc:    subfunc,
		UnSubFunc:  unsubfunc,
		ctx:        ctx,
	}
}

//AddClient 新增客户端
func (h *Hub) AddClient(channels []string, conn *websocket.Conn) {
	if len(channels) == 0 || conn == nil {
		return
	}
	log.Println("新增客户端", conn.RemoteAddr(), " channels:", channels)
	h.lock.Lock()
	defer h.lock.Unlock()

	oldChannelMap := h.ClientMap[conn]
	removeChannelMap, addChannelMap := h.DiffChannelMap(oldChannelMap, channels)
	for k := range removeChannelMap {
		if h.channelMap[k] != nil {
			index := -1
			for i, c := range h.channelMap[k] {
				if c == conn {
					index = i
					break
				}
			}
			if index > -1 {
				h.channelMap[k] = append(h.channelMap[k][:index], h.channelMap[k][index+1:]...)
			}
		}
	}
	for k := range addChannelMap {
		if h.channelMap[k] == nil {
			h.channelMap[k] = []*websocket.Conn{conn}
		} else {
			index := -1
			for i, c := range h.channelMap[k] {
				if c == conn {
					index = i
					break
				}
			}
			if index == -1 {
				h.channelMap[k] = append(h.channelMap[k], conn)
			}
		}
	}
	channelMap := make(map[string]bool)
	for _, v := range channels {
		channelMap[v] = true
	}
	h.ClientMap[conn] = channelMap
	h.UpdateSubChannels()
}

//RemoveClient 移除客户端
func (h *Hub) RemoveClient(conn *websocket.Conn) {
	if conn == nil {
		return
	}
	log.Println("remove client", conn.RemoteAddr())
	h.lock.Lock()
	defer h.lock.Unlock()
	channels := h.ClientMap[conn]
	if channels == nil {
		return
	}
	for k := range channels {
		if h.channelMap[k] != nil {
			for i, vv := range h.channelMap[k] {
				if vv == conn {
					if len(h.channelMap[k]) == 1 {
						delete(h.channelMap, k)
					} else {
						h.channelMap[k] = append(h.channelMap[k][:i], h.channelMap[k][i+1:]...)
					}
					break
				}
			}
		}
	}
	delete(h.ClientMap, conn)
	h.UpdateSubChannels()
}

//UpdateSubChannels 更新Sub的channels
func (h *Hub) UpdateSubChannels() {
	newChannels := h.GetChannels()
	oldChannelMap := make(map[string]bool, 0)
	for _, v := range h.channels {
		oldChannelMap[v] = true
	}
	removeChannelMap, addChannelMap := h.DiffChannelMap(oldChannelMap, newChannels)
	removeChannels := make([]string, 0)
	for k := range removeChannelMap {
		if k != "ping" {
			removeChannels = append(removeChannels, k)
		}
	}
	if len(removeChannels) > 0 {
		h.UnSubFunc(h.ctx, removeChannels...)
	}
	addChannels := make([]string, 0)
	for k := range addChannelMap {
		addChannels = append(addChannels, k)
	}
	if len(addChannels) > 0 {
		h.SubFunc(h.ctx, addChannels...)
	}

	h.channels = newChannels
}

//DiffChannelMap channelMap比较
func (h *Hub) DiffChannelMap(oldChannelMap map[string]bool, newChannels []string) (removeChannelMap map[string]bool, addChannelMap map[string]bool) {
	removeChannelMap = make(map[string]bool)
	addChannelMap = make(map[string]bool)

	newChannelMap := make(map[string]bool)
	for _, v := range newChannels {
		newChannelMap[v] = true
	}
	for k, v := range oldChannelMap {
		if !newChannelMap[k] && v {
			removeChannelMap[k] = v
		}
	}
	for _, v := range newChannels {
		if !oldChannelMap[v] {
			addChannelMap[v] = true
		}
	}
	return removeChannelMap, addChannelMap
}

//GetChannels 获取所有channel
func (h *Hub) GetChannels() []string {
	channels := make([]string, 0, len(h.channelMap))
	for k := range h.channelMap {
		channels = append(channels, k)
	}
	return channels
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//HandlerDynamicChannel 向websocket推送mq数据,可以变更channel
func (h *Hub) HandlerDynamicChannel(prefix string, checkToken func(token string) bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.URL.Query().Get("token")
		if checkToken != nil && !checkToken(accessToken) {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade:", err)
			return
		}
		defer c.Close()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			channels := strings.Split(string(message), ";")
			newChannels := make([]string, 0, len(channels))
			for _, v := range channels {
				if v != "" {
					newChannels = append(newChannels, prefix+v)
				}
			}
			h.AddClient(newChannels, c)
			log.Println(h.channelMap)
			log.Println(h.ClientMap)
		}
		h.RemoveClient(c)
		log.Println(h.channelMap)
		log.Println(h.ClientMap)
	}
}

//HandlerStaticChannels 向websocket推送mq数据,不可变更channel
func (h *Hub) HandlerStaticChannels(channels []string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()
		h.AddClient(channels, c)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err, message)
				break
			}
		}
		h.RemoveClient(c)
	}
}
