package main

import (
	"log"
	"net/http"

	"github.com/9d77v/wspush/redishub"
)

func checkToken(token string) bool {
	return false
}

func main() {
	http.HandleFunc("/ws", redishub.Hub.HandlerDynamicChannel("bbb.", checkToken))
	// http.HandleFunc("/ws", redishub.Hub.HandlerDynamicChannel("bbb.", nil))
	// http.HandleFunc("/ws", redishub.Hub.HandlerStaticChannels([]string{"test_channel"}))

	log.Fatal(http.ListenAndServe(":8200", nil))
}
