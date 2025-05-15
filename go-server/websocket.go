package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Upgrade error:", err)
        return
    }
    defer conn.Close()

    streamID := r.URL.Query().Get("stream")
    if streamID == "" {
        log.Println("Missing stream ID")
        return
    }

    mu.Lock()
    if rooms[streamID] == nil {
        rooms[streamID] = &TeamCount{
            Clients: make(map[*websocket.Conn]bool),
        }
    }
    rooms[streamID].Clients[conn] = true
    mu.Unlock()

    for {
		var msg Message
        if err := conn.ReadJSON(&msg); err != nil {
            log.Println("Read error:", err)
            break
        }

        mu.Lock()
        if room, ok := rooms[msg.StreamID]; ok {
            room.Support += msg.Support
            room.Obstruct += msg.Obstruct
        }
        mu.Unlock()
    }

    mu.Lock()
    delete(rooms[streamID].Clients, conn)
    mu.Unlock()
}