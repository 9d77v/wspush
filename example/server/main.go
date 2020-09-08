package main

import (
	"log"
	"net/http"

	"github.com/9d77v/wspush/redishub"
)

func main() {
	http.HandleFunc("/ws", redishub.Hub.HandlerDynamicChannel())
	// http.HandleFunc("/ws", redishub.Hub.HandlerStaticChannels([]string{"test_channel"}))

	log.Fatal(http.ListenAndServe(":8200", nil))
}
