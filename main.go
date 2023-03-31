package main

import (
	"log"
	"sync/atomic"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
)

type UserIdentifier struct {
	id atomic.Int32
}

func (u *UserIdentifier) GetId() int32 {
	return u.id.Add(1)
}

var userIdentifier *UserIdentifier

func main() {
	hub := NewHub()
	go hub.Run()

	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.

		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/", websocket.New(func(c *websocket.Conn) {
		client := &Client{
			ID:   userIdentifier.GetId(),
			Hub:  Hub,
			Conn: c,
			Send: make(chan []byte, 256),
		}

		client.Hub.Register <- client

		go client.WritePump()
		client.ReadPump()
	}, websocket.Config{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}))

	log.Fatal(app.Listen(":4000"))
}
