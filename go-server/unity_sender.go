package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func startAggregation() {
    type Aggregation struct {
        Support  int
        Obstruct int
    }

    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    streamStats := make(map[string]*Aggregation)

    for range ticker.C {
        mu.Lock()
        for streamID, room := range rooms {
            if _, ok := streamStats[streamID]; !ok {
                streamStats[streamID] = &Aggregation{}
            }
            s := streamStats[streamID]

            s.Support += room.Support
            s.Obstruct += room.Obstruct

            log.Printf("[%s] Support: %d, Obstruct: %d (collected)", streamID, room.Support, room.Obstruct)

            room.Support = 0
            room.Obstruct = 0

            for s.Support >= 10 {
                log.Printf("Send to Unity [%s]: Support: 1, Obstruct: %d", streamID, s.Obstruct)
                SendToUnityPerStream(streamID, 1, 0)
                s.Support -= 10
            }
            for s.Obstruct >= 10 {
                log.Printf("Send to Unity [%s]: Support: %d, Obstruct: 1", streamID, s.Support)
                SendToUnityPerStream(streamID, 0, 1)
                s.Obstruct -= 10
            }
        }
        mu.Unlock()
    }
}

func SendToUnityPerStream(streamID string, support int, obstruct int) {
    log.Printf("SendToUnityPerStream: streamID: %s, support: %d, obstruct: %d", streamID, support, obstruct)
    // mu.Lock()
    // defer mu.Unlock()
    room, ok := rooms[streamID]
    if !ok {
        log.Printf("No room for streamID: %s", streamID)
        return
    }
    log.Printf("Unity clients in room [%s]: %d", streamID, len(room.Clients)) // クライアント数を出力

    // Unityクライアント全員に送信
    msg := struct {
        Support  int `json:"support"`
        Obstruct int `json:"obstruct"`
    }{
        Support:  support,
        Obstruct: obstruct,
    }
    data, err := json.Marshal(msg)
    if err != nil {
        log.Printf("JSON marshal error: %v", err)
        return
    }
    for conn := range room.Clients {
        log.Printf("Try send to Unity client in room [%s]", streamID)
        if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
            log.Printf("Send error to Unity (streamID=%s): %v", streamID, err)
            conn.Close()
            delete(room.Clients, conn)
        }
    }
}