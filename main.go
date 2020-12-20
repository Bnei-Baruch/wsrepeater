package main

import (
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8080", "http service address")
var broadcast = flag.String("broadcast", "wss://ktuviot.kbb1.com/ws/ws", "broadcast service address")

type Message struct {
	ID       uint   `json:"id"`
	Message  string `json:"message"`
	UserName string `json:"user_name"`
	Type     string `json:"type"`
	Language string `json:"language"`
	Approved bool   `json:"approved"`
}

var knownMessages = map[string]Message{}

func main() {
	flag.Parse()
	hub := newHub()
	go hub.run()

	connectToShidur(broadcast, hub)

	http.HandleFunc("/", serveHome) // Debugging page
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	log.Println("Serving", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "home.html")
}
