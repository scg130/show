package main

import (
	"fmt"
	"log"
	"net/http"
	"show/broadcast"
	"show/dto"
	"show/handle"
)


func main() {
	go broadcast.MessHandler()
	http.HandleFunc("/websocket", handle.Echo)
	http.Handle("/", http.FileServer(http.Dir("./app/public")))
	http.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
		dto.Anchors.Range(func(key, value interface{}) bool {
			fmt.Println(key)
			return true
		})
	})
	log.Println("Serving at localhost:4000...")
	log.Fatal(http.ListenAndServeTLS("0.0.0.0:4000", "./localhost+3.pem", "./localhost+3-key.pem", nil))
}
