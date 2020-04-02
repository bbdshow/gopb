package cmd

import (
	"github.com/gorilla/websocket"
	"time"
)

type DoMsg struct {
	Duration   string `json:"duration"`
	Concurrent int    `json:"concurrent"`
	//TotalCalls        int               `json:"total_calls"`
	//Method            string            `json:"method"`
	//URL               string            `json:"url"`
	//Headers           map[string]string `json:"headers"`
	//DisableKeepAlives bool              `json:"disable_keep_alives"`
	//Params            map[string]string `json:"params"`
	//Body              string            `json:"body"`
	//Contains          string            `json:"contains"`
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadDeadline(time.Now().Add(time.Minute))
}
