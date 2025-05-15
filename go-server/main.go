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
	Support  int    `json:"support"`
	Obstruct int    `json:"obstruct"`
	StreamID string `json:"stream"`
}

type TeamCount struct {
	Support  int
	Obstruct int
	Clients  map[*websocket.Conn]bool
}

type StreamStat struct {
	Support  int
	Obstruct int
	Count    int
}

var (
	rooms       = make(map[string]*TeamCount)
	streamStats = make(map[string]*StreamStat)
	mu          sync.Mutex
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

func startAggregation() {
	ticker := time.NewTicker(1 * time.Second)

	for range ticker.C {
		mu.Lock()
		for streamID, room := range rooms {
			stat, exists := streamStats[streamID]
			if !exists {
				stat = &StreamStat{}
				streamStats[streamID] = stat
			}

			stat.Support += room.Support
			stat.Obstruct += room.Obstruct
			stat.Count++

			log.Printf("[%s] Support: %d, Obstruct: %d (collected)", streamID, room.Support, room.Obstruct)

			room.Support = 0
			room.Obstruct = 0

			if stat.Count >= 5 {
				log.Printf("Send to Unity [%s]: Support: %d, Obstruct: %d", streamID, stat.Support, stat.Obstruct)
				SendToUnityPerStream(streamID, stat.Support, stat.Obstruct)
				stat.Support = 0
				stat.Obstruct = 0
				stat.Count = 0
			}
		}
		mu.Unlock()
	}
}

func SendToUnityPerStream(streamID string, support int, obstruct int) {
	log.Printf("[UnityMock] [%s] support=%d, obstruct=%d", streamID, support, obstruct)
	// ここでUnity側のエンドポイントにHTTP POSTするなどに差し替え可能
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../client/client.html")
	})
	go startAggregation()

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}