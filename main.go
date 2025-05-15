package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

type Message struct {
    Team string `json:"team"` // "support" or "obstruct"
}

var (
    clients     = make(map[*websocket.Conn]string)
    teamCount   = map[string]int{"support": 0, "obstruct": 0}
    mu          sync.Mutex
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Upgrade error:", err)
        return
    }
    defer conn.Close()

    mu.Lock()
    clients[conn] = ""
    mu.Unlock()

    for {
        var msg Message
        if err := conn.ReadJSON(&msg); err != nil {
            log.Println("Read error:", err)
            break
        }

        mu.Lock()
        clients[conn] = msg.Team
        teamCount[msg.Team]++
        mu.Unlock()
    }

    mu.Lock()
    delete(clients, conn)
    mu.Unlock()
}

func startAggregation() {
    ticker := time.NewTicker(1 * time.Second)
    for range ticker.C {
        mu.Lock()
        support := teamCount["support"]
        obstruct := teamCount["obstruct"]
        teamCount["support"] = 0
        teamCount["obstruct"] = 0
        mu.Unlock()

        total := support + obstruct
        var ratio float64
        if total > 0 {
            ratio = float64(support) / float64(total)
        }

        log.Printf("Support Ratio: %.2f", ratio)
        SendToUnity(ratio)
    }
}

func main() {
    http.HandleFunc("/ws", handleWebSocket)
    go startAggregation()

    log.Println("Server started at :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
