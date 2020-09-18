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
		var message Message
		err = json.Unmarshal(msg, &message)
		if err != nil {
			log.Println("unmarshal err:", err)
			continue
		}
		if message.Type != "" {
			if message.Approved && message.Type == "question" {
				// Single approved question
				hub.broadcast <- msg
				updateMessage(message)
			}
		} else {
			// "New question" message was received
			err = json.Unmarshal(msg, &knownMessages)
			if err != nil {
				log.Println("unmarshal err:", err)
				continue
			}
			sendKnownMessages(hub, &knownMessages)
		}
	}
}

func sendKnownMessages(hub *Hub, knownMessages *map[string][]Message) {
	for _, message := range (*knownMessages)["questions"] {
		msg, err := json.Marshal(message)
		if err != nil {
			log.Println(err)
			continue
		}
		hub.broadcast <- msg
	}
}

func updateMessage(message Message) {
	for idx, msg := range knownMessages["questions"] {
		if message.Language == msg.Language {
			knownMessages["questions"][idx] = message
			break
		}
	}
	// not found
	knownMessages["questions"] = append(knownMessages["questions"], message)
}
