package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

var (
    rooms = make(map[string]*TeamCount)
    mu    sync.Mutex
)

func main() {
    http.HandleFunc("/ws", handleWebSocket)
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "../client/client.html")
    })
    go startAggregation()

    log.Println("Server started at :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}