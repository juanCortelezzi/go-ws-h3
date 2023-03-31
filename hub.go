package main

import (
	"errors"
	"fmt"
	"log"
)

type ConnectionHub struct {
	Clients    map[int32]*Client
	FromSocket chan *MsgFromClient
	FromServer chan *MsgFromServer
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *ConnectionHub {
	return &ConnectionHub{
		Clients:    make(map[int32]*Client),
		FromSocket: make(chan *MsgFromClient),
		FromServer: make(chan *MsgFromServer),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
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
				msg.Client.Send <- NewEventMsg("Trigger").toBytes()

			default:
				if client, exists := h.Clients[msg.Client.ID]; exists {
					DeleteClient(h.Clients, client)
				}
			}

		case msg := <-h.FromServer:
			if client, exists := h.Clients[msg.To]; exists {
				client.Send <- msg.toBytes()
				msg.OneShot <- nil
				close(msg.OneShot)
				continue
			}
			msg.OneShot <- errors.New("Client not found")
			close(msg.OneShot)

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
