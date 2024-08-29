package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type message struct {
	Message    string `json:"text"`
	Timestamp  uint64 `json:"timestamp"`
	SenderId   uint64 `json:"senderId"`
	ReceiverId uint64 `json:"receiverId"`
	Read       bool   `json:"read"`
}

var (
	wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	connections = make(map[uint64]*websocket.Conn)
	connMutex   sync.Mutex
)

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.URL.Query().Get("userId")

	if userIdStr == "" {
		http.Error(w, "userId parameter is required", http.StatusBadRequest)
		return
	}

	userId, err := strconv.ParseUint(userIdStr, 10, 64)

	if err != nil {
		http.Error(w, "invalid userId format", http.StatusBadRequest)
		return
	}

	wsConn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("could not upgrade: %s \n", err.Error())
		return
	}
	defer wsConn.Close()

	connMutex.Lock()
	connections[userId] = wsConn
	connMutex.Unlock()

	defer func() {
		connMutex.Lock()
		delete(connections, userId)
		connMutex.Unlock()
	}()

	for {
		var msg message
		err := wsConn.ReadJSON(&msg)
		if err != nil {
			fmt.Printf("error reading json: %s \n", err.Error())
			break
		}

		fmt.Printf("message received: %s\n", msg.Message)

		broadcastMessage(msg)
	}
}

func broadcastMessage(msg message) {
	connMutex.Lock()
	defer connMutex.Unlock()

	if senderConn, ok := connections[msg.SenderId]; ok {
		err := senderConn.WriteJSON(msg)
		if err != nil {
			fmt.Printf("error sending message to sender: %s \n", err.Error())
		}
	}

	if receiverConn, ok := connections[msg.ReceiverId]; ok {
		err := receiverConn.WriteJSON(msg)
		if err != nil {
			fmt.Printf("error sending message to receiver: %s \n", err.Error())
		}
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", wsEndpoint)

	log.Fatal(http.ListenAndServe(":9200", router))
}
