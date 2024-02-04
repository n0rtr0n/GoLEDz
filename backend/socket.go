package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// origin := r.Header.Get("Origin")
		// TODO: we do not want to blindly trust any origin.
		// this is only for the sake of simplicity and we'll solve when we add configuration management
		return true
	},
}

type SocketConnection struct {
	conn *websocket.Conn
	ch   chan *PixelMap
	done chan struct{}
}

func (s *SocketConnection) handleMessages() {
	defer s.conn.Close()

	for {
		select {
		case pixelMap := <-s.ch:
			jsonData, err := (*pixelMap).toJSON()

			if err != nil {
				log.Fatal(err)
				return
			}
			err = s.conn.WriteMessage(websocket.TextMessage, jsonData)
			if err != nil {
				log.Fatal(err)
				return
			}
		case <-s.done:
			return
		}
	}
}

func socketHandler(w http.ResponseWriter, r *http.Request, ch chan *PixelMap) {
	fmt.Println("handling websocket connection")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	socketConn := &SocketConnection{
		conn: conn,
		ch:   ch,
		done: make(chan struct{}),
	}

	go socketConn.handleMessages()

	go func() {
		<-socketConn.done
		close(ch)
	}()
}
