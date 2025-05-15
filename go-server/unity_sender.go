package main

import (
	"log"
	"time"
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
    log.Printf("[UnityMock] [%s] support=%d, obstruct=%d", streamID, support, obstruct)
    // WebSocket送信やUnity連携部分はここで実装
}