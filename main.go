package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type message struct {
    Message string `json:"message"`
    Time    int64  `json:"time"`
}

var (
	wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	wsConn *websocket.Conn
)

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	wsUpgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	wsConn, err := wsUpgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Printf("could not upgrade: %s \n", err.Error())
	}
	defer wsConn.Close()

	for {

		currentTime := time.Now()
        epoch := currentTime.Unix()
		var msg message
		err := wsConn.ReadJSON(&msg)
		msg.Time = epoch


		if err != nil {
			fmt.Printf("error reading json: %s \n", err.Error())			
		}

		fmt.Printf("message received: %s\n", msg.Message, )

		// Echo the received message back to the client
		err = wsConn.WriteJSON(msg)
		if err != nil {
			fmt.Printf("error writing json: %s \n", err.Error())
			break
		}

	}
}

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/", wsEndpoint)

	log.Fatal(http.ListenAndServe(":9200", router))

}
