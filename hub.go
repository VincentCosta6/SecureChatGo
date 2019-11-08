package main

import (
	"encoding/json"
	"fmt"
)

type Hub struct {
	clients map[*Client]bool
	clientMap map[string]*Client // user [user id] to Client
	receive chan []byte
	createMessage chan CreatedMessageStruct
	register chan *Client
	unregister chan *Client
}

type CreatedMessageStruct struct {
	message *WebsocketMessageType
	clients *[]string // user ids
}

func newHub() *Hub {
	return &Hub{
		createMessage:  make(chan CreatedMessageStruct),
		receive: make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		clientMap: make(map[string]*Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.clientMap[client.user.User.ID.Hex()] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.clientMap, client.user.User.ID.Hex())
				close(client.send)
			}
		case message := <-h.createMessage:
			bytes, err := json.Marshal(message.message)

			if err != nil {
				fmt.Println("Bad", err)
				continue
			}

			for _, client := range *message.clients {
				var connection = h.clientMap[client]
				if connection!= nil {
					select {
					case connection.send <- bytes:
					default:
						close(connection.send)
						delete(h.clients, connection)
						delete(h.clientMap, client)
					}
				}
			}
		}
	}
}

