package main

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	TypeLocation  = "location"
	TypePhoneInfo = "phoneInfo"
	TypeTakePhoto = "takePhoto"
	TypeTrigger   = "trigger"
)

type Msg[T any] struct {
	Type string `json:"type"`
	Data T      `json:"data"`
}

func NewEventMsg(t string) *Msg[interface{}] {
	return &Msg[interface{}]{
		Type: t,
		Data: nil,
	}
}

func (m *Msg[T]) toBytes() []byte {
	bytes, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Sprintf("failed to parse message: %s", err))
	}
	return bytes
}

type ServerMsg struct {
	ClientID int32
	Msg[json.RawMessage]
}

type LocationData struct {
	Timestamp time.Time `json:"time"`
	Lat       string    `json:"lat"`
	Lon       string    `json:"lon"`
}

type PhoneInfoData struct {
	Name        string  `json:"name"`
	Battery     float32 `json:"battery"`
	Temperature float32 `json:"temp"`
}

func parseMessage[T any](data []byte) (*T, error) {
	var x T
	if err := json.Unmarshal(data, &x); err != nil {
		return nil, err
	}
	return &x, nil
}
