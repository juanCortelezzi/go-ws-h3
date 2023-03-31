package main

import (
	"fmt"
	"log"
)

var Hub *ConnectionHub

func Create() {
	if Hub != nil {
		log.Println("DATABASE: already connected")
		return
	}
	Hub = NewHub()
}

type ConnectionHub struct {
	Clients    map[int32]*Client
	Register   chan *Client
	FromSocket chan *ServerMsg
	Unregister chan *Client
}

func NewHub() *ConnectionHub {
	return &ConnectionHub{
		Clients:    make(map[int32]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		FromSocket: make(chan *ServerMsg),
	}
}

func DeleteClient(clients map[int32]*Client, client *Client) {
	close(client.Send)
	delete(clients, client.ID)
}

func (h *ConnectionHub) Run() {
	for {
		select {
		case msg := <-h.FromSocket:
			switch msg.Type {
			case TypeLocation:
				data, err := parseMessage[LocationData](msg.Data)
				if err != nil {
					log.Printf("parse error: %v", err)
					break
				}
				fmt.Printf("data is: %#v\n", data)

			case TypePhoneInfo:
				data, err := parseMessage[PhoneInfoData](msg.Data)
				if err != nil {
					log.Printf("parse error: %v", err)
					break
				}
				fmt.Printf("data is: %#v\n", data)

			case TypeTrigger:
				// NOTE: this message event is only for testing purposes
				client, exists := h.Clients[msg.ClientID]
				if !exists {
					panic("how does it not exist? this should have never happened")
				}

				client.Send <- NewEventMsg(TypeTakePhoto).toBytes()

			default:
				if client, exists := h.Clients[msg.ClientID]; exists {
					DeleteClient(h.Clients, client)
				}
			}

		case client := <-h.Register:
			h.Clients[client.ID] = client
			log.Printf("registered client `%d` len: %d\n", client.ID, len(h.Clients))

		case client := <-h.Unregister:
			if _, exists := h.Clients[client.ID]; exists {
				DeleteClient(h.Clients, client)
			}
			log.Printf("removed client `%d` len: %d\n", client.ID, len(h.Clients))
		}
	}
}
