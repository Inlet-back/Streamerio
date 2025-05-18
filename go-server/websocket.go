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
    log.Printf("Add client to room [%s], total: %d", streamID, len(rooms[streamID].Clients))
    mu.Unlock()

    // メッセージ受信ループ
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

    // 切断時にクライアントを削除
    mu.Lock()
    delete(rooms[streamID].Clients, conn)
    log.Printf("Remove client from room [%s], total: %d", streamID, len(rooms[streamID].Clients))
    mu.Unlock()
}