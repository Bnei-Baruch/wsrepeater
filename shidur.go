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
		log.Printf("--recv--: %s", msg)

		if isMsg, message := unmarshalMsg(msg); isMsg {
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

			clean, err := json.Marshal(map[string]bool{"question": true})
			if err != nil {
				return
			}
			hub.broadcast <- clean
		}
	}
}

func unmarshalMsg(data []byte) (bool, Message) {
	var q Message
	if err := json.Unmarshal(data, &q); err != nil || q.ID == 0 {
		log.Printf("unmarshalMsg false %s, %s, %s", err, q.Message, q.Type)
		return false, Message{}
	}

	log.Printf("unmarshalMsg true %s, %s", q.Message, q.Type)
	return true, q
}
