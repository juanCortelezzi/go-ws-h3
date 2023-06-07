package main

import (
	"encoding/json"
	"fmt"
)

const (
	TypeLocation  = "location"
	TypeTakePhoto = "takePhoto"
)

type Msg[T any] struct {
	Type string `json:"type"`
	Data T      `json:"data"`
}

type MsgFromClient struct {
	Client *Client
	Msg[json.RawMessage]
}

type MsgFromServer struct {
	To      int32
	OneShot chan<- error
	Msg[interface{}]
}

type LocationData struct {
	// Timestamp time.Time `json:"time"`
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	ID  uint    `json:"id"`
}

type IDDdata struct {
	Id uint `json:"id"`
}

type PhoneInfoData struct {
	Name        string  `json:"name"`
	Battery     float32 `json:"battery"`
	Temperature float32 `json:"temp"`
}

func NewEventMsg(t string) Msg[interface{}] {
	return Msg[interface{}]{
		Type: t,
		Data: nil,
	}
}

func (m Msg[T]) toBytes() []byte {
	bytes, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Sprintf("failed to parse message: %s", err))
	}
	return bytes
}

func parseMessage[T any](data []byte) (*T, error) {
	var x T
	if err := json.Unmarshal(data, &x); err != nil {
		return nil, err
	}
	return &x, nil
}
