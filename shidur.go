package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/recws-org/recws"
)

func connectToShidur(broadcast *string, hub *Hub) {
	log.Printf("Connecting to %s...", *broadcast)
	ws := recws.RecConn{
		KeepAliveTimeout: 10 * time.Second,
	}
	ws.Dial(*broadcast, nil)

	go serveShidur(hub, &ws)
}

func serveShidur(hub *Hub, ws *recws.RecConn) {
	defer ws.Close()
	for {
		if !ws.IsConnected() {
			log.Printf("Websocket disconnected %s", ws.GetURL())
			continue
		}

		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("recv: %s", msg)

		if !isCleanMsg(msg) {
			var message Message
			err = json.Unmarshal(msg, &message)
			if err != nil {
				log.Println("unmarshal err:", err)
				continue
			}
			if message.Approved && message.Type == "question" {
				// Single approved question
				hub.broadcast <- msg
				knownMessages[message.Language] = message
			}
		} else {
			// "New question" message was received
			for k := range knownMessages {
				delete(knownMessages, k)
			}

			clean, err := json.Marshal(map[string]bool{"clear": true})
			if err != nil {
				return
			}
			hub.broadcast <- clean
		}
	}
}

func isCleanMsg(data []byte) bool {
	var qs map[string][]Message
	if err := json.Unmarshal(data, qs); err != nil {
		return false
	}

	for _, q := range qs["questions"] {
		if q.ID != 0 {
			return false
		}
	}
	return true
}
