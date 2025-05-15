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
    Support  int `json:"support"`
    Obstruct int `json:"obstruct"`
}

var (
    clients   = make(map[*websocket.Conn]bool)
    teamCount = map[string]int{"support": 0, "obstruct": 0}
    mu        sync.Mutex
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Upgrade error:", err)
        return
    }
    defer conn.Close()

    mu.Lock()
    clients[conn] = true
    mu.Unlock()

    for {
        var msg Message
        if err := conn.ReadJSON(&msg); err != nil {
            log.Println("Read error:", err)
            break
        }

        mu.Lock()
        teamCount["support"] += msg.Support
        teamCount["obstruct"] += msg.Obstruct
        mu.Unlock()
    }

    mu.Lock()
    delete(clients, conn)
    mu.Unlock()
}

func startAggregation() {
    var (
        supportSum  int
        obstructSum int
        count       int
    )
    ticker := time.NewTicker(1 * time.Second)
    for range ticker.C {
        mu.Lock()
        support := teamCount["support"]
        obstruct := teamCount["obstruct"]
        teamCount["support"] = 0
        teamCount["obstruct"] = 0
        mu.Unlock()

        supportSum += support
        obstructSum += obstruct
        count++

        log.Printf("Support: %d, Obstruct: %d (collected)", support, obstruct)

        if count >= 5 {
            log.Printf("Send to Unity: Support: %d, Obstruct: %d", supportSum, obstructSum)
            SendToUnity(supportSum, obstructSum)
            supportSum = 0
            obstructSum = 0
            count = 0
        }
    }
}

func main() {
    http.HandleFunc("/ws", handleWebSocket)
    // 追加: client.htmlをルートで返す
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "client.html")
    })
    go startAggregation()

    log.Println("Server started at :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}