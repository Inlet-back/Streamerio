package main

import "github.com/gorilla/websocket"

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