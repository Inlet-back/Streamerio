// / interactive-game/go-server/main.go
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

var (
	rooms = make(map[string]*TeamCount)
	mu    sync.Mutex
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
            // 初期化されていなければ作る
            if _, ok := streamStats[streamID]; !ok {
                streamStats[streamID] = &Aggregation{}
            }
            s := streamStats[streamID]

            // 現在の集計値を加算
            s.Support += room.Support
            s.Obstruct += room.Obstruct

            log.Printf("[%s] Support: %d, Obstruct: %d (collected)", streamID, room.Support, room.Obstruct)

            // 配信者の現在の集計値をリセット
            room.Support = 0
            room.Obstruct = 0

            // SupportまたはObstructが10以上になったらUnityへ送信し、リセット
            if s.Support >= 10{
                log.Printf("Send to Unity [%s]: Support: %d, Obstruct: %d", streamID, s.Support, s.Obstruct)
                SendToUnityPerStream(streamID, s.Support, s.Obstruct)
                s.Support = 0
                
            }
            if s.Obstruct >= 10{
                log.Printf("Send to Unity [%s]: Support: %d, Obstruct: %d", streamID, s.Support, s.Obstruct)
                SendToUnityPerStream(streamID, s.Support, s.Obstruct)
                s.Obstruct = 0
            }
        }
        mu.Unlock()
    }
}
func SendToUnityPerStream(streamID string, support int, obstruct int) {
	log.Printf("[UnityMock] [%s] support=%d, obstruct=%d", streamID, support, obstruct)
	// WebSocket送信やUnity連携部分はここで実装
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